package controllers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/gofiber/fiber/v2"
)

type IPaymuController struct{}

// ========================================
// iPaymu Configuration
// ========================================

const (
	ipaymuSandboxURL    = "https://sandbox.ipaymu.com"
	ipaymuProductionURL = "https://my.ipaymu.com"
)

func getIPaymuVA() string {
	return GetConfigValue(models.ConfigIPaymuVA, "0000001354000500")
}

func getIPaymuAPIKey() string {
	return GetConfigValue(models.ConfigIPaymuAPIKey, "SANDBOXE7008C33-C912-41B7-9F5C-9B79D4DE2D48")
}

func isIPaymuSandbox() bool {
	return GetConfigValue(models.ConfigIPaymuSandbox, "true") == "true"
}

func getIPaymuBaseURL() string {
	if isIPaymuSandbox() {
		return ipaymuSandboxURL
	}
	return ipaymuProductionURL
}

// ========================================
// iPaymu Signature Generation
// ========================================

func generateIPaymuSignature(method string, body []byte) string {
	va := getIPaymuVA()
	apiKey := getIPaymuAPIKey()

	// SHA-256 hash of request body
	bodyHash := sha256.Sum256(body)
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))

	// String to sign: METHOD:VA:BODY_HASH:API_KEY
	stringToSign := method + ":" + va + ":" + bodyHashHex + ":" + apiKey

	// HMAC-SHA256 with API key
	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

// ========================================
// iPaymu API Call Helper
// ========================================

func callIPaymuAPI(endpoint string, body []byte) (map[string]interface{}, error) {
	baseURL := getIPaymuBaseURL()
	url := baseURL + endpoint

	signature := generateIPaymuSignature("POST", body)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("va", getIPaymuVA())
	req.Header.Set("signature", signature)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("iPaymu API error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v (body: %s)", err, string(respBody))
	}

	fmt.Printf("[iPaymu] %s response: %s\n", endpoint, string(respBody))

	return result, nil
}

// ========================================
// Create Payment (for both subscriptions and extra credits)
// ========================================

// CreatePayment creates an iPaymu payment for a subscription plan or extra credit package
func (c *IPaymuController) CreatePayment(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Reset free credits if needed
	ResetFreeCreditIfNeeded(user)

	var req struct {
		PackageID string `json:"package_id"` // For extra credit packages
		PlanID    string `json:"plan_id"`    // For subscription plans
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.PackageID == "" && req.PlanID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Either package_id or plan_id is required"})
	}

	var productName string
	var price int64
	var txType string
	var packageID *string
	var planID *string

	if req.PlanID != "" {
		// Subscription plan purchase
		var plan models.SubscriptionPlan
		if err := initializers.Db.Where("id = ? AND visibility = ?", req.PlanID, true).First(&plan).Error; err != nil {
			return ctx.Status(404).JSON(fiber.Map{"error": "Subscription plan not found"})
		}
		productName = "Subscription: " + plan.PlanName
		price = plan.Price
		txType = models.TxTypeSubscription
		planID = &plan.ID
	} else {
		// Extra credit package purchase
		var pkg models.Package
		if err := initializers.Db.Where("id = ? AND visibility = ?", req.PackageID, true).First(&pkg).Error; err != nil {
			return ctx.Status(404).JSON(fiber.Map{"error": "Package not found"})
		}
		productName = "Extra Credits: " + pkg.PackageName
		price = pkg.Price
		txType = models.TxTypeExtraCredit
		packageID = &pkg.ID
	}

	// Handle free items (price = 0)
	if price == 0 {
		return c.handleFreeItem(ctx, user, packageID, planID, txType)
	}

	// Create transaction record
	txID := generateTransactionID()
	transaction := models.Transaction{
		ID:            txID,
		UserID:        user.ID,
		PackageID:     packageID,
		PlanID:        planID,
		Amount:        price,
		Status:        models.TransactionStatusPending,
		PaymentMethod: "ipaymu",
		TxType:        txType,
	}

	if err := initializers.Db.Create(&transaction).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	// Build return/callback URLs
	baseURL := ctx.Protocol() + "://" + ctx.Hostname()
	// Use frontend origin if available
	origin := ctx.Get("Origin")
	if origin != "" {
		baseURL = origin
	}

	returnURL := baseURL + "/payment/status/" + txID
	cancelURL := baseURL + "/packages"

	// Determine notify URL (callback from iPaymu server-to-server)
	notifyURL := ctx.Protocol() + "://" + ctx.Hostname() + "/api/ipaymu/callback"

	// Build iPaymu payment request
	paymentBody, _ := json.Marshal(map[string]interface{}{
		"product":     []string{productName},
		"qty":         []int{1},
		"price":       []int64{price},
		"returnUrl":   returnURL,
		"cancelUrl":   cancelURL,
		"notifyUrl":   notifyURL,
		"referenceId": txID,
		"buyerName":   user.Name,
		"buyerEmail":  user.Email,
	})

	result, err := callIPaymuAPI("/api/v2/payment", paymentBody)
	if err != nil {
		// Rollback transaction
		initializers.Db.Delete(&transaction)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create payment: " + err.Error()})
	}

	// Check iPaymu response status
	status, _ := result["Status"].(float64)
	if int(status) != 200 {
		initializers.Db.Delete(&transaction)
		msg := "Payment creation failed"
		if m, ok := result["Message"].(string); ok {
			msg = m
		}
		return ctx.Status(500).JSON(fiber.Map{"error": msg})
	}

	// Extract payment URL and session ID from response
	data, _ := result["Data"].(map[string]interface{})
	paymentURL := ""
	sessionID := ""
	if data != nil {
		if u, ok := data["Url"].(string); ok {
			paymentURL = u
		}
		if s, ok := data["SessionID"].(string); ok {
			sessionID = s
		}
	}

	// Update transaction with iPaymu reference
	if sessionID != "" {
		transaction.ExternalID = sessionID
	}
	initializers.Db.Save(&transaction)

	return ctx.JSON(fiber.Map{
		"transaction_id": txID,
		"payment_url":    paymentURL,
		"amount":         price,
		"product":        productName,
	})
}

// handleFreeItem processes free packages/plans without payment
func (c *IPaymuController) handleFreeItem(ctx *fiber.Ctx, user *models.User, packageID *string, planID *string, txType string) error {
	txID := generateTransactionID()
	now := time.Now()

	transaction := models.Transaction{
		ID:            txID,
		UserID:        user.ID,
		PackageID:     packageID,
		PlanID:        planID,
		Amount:        0,
		Status:        models.TransactionStatusSettlement,
		PaymentMethod: "free",
		TxType:        txType,
		PaidAt:        &now,
	}

	if err := initializers.Db.Create(&transaction).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	// Process the free item
	settleTransaction(&transaction, user)

	return ctx.JSON(fiber.Map{
		"free":           true,
		"transaction_id": txID,
		"message":        "Free item claimed successfully",
	})
}

// ========================================
// iPaymu Callback Handler
// ========================================

// HandleCallback processes iPaymu payment notifications (notifyUrl)
func (c *IPaymuController) HandleCallback(ctx *fiber.Ctx) error {
	// iPaymu sends form data or JSON callback
	// Common fields: trx_id, reference_id, status, status_code, sid

	var callback struct {
		TrxID       string `json:"trx_id" form:"trx_id"`
		ReferenceID string `json:"reference_id" form:"reference_id"`
		Status      string `json:"status" form:"status"`
		StatusCode  string `json:"status_code" form:"status_code"`
		Sid         string `json:"sid" form:"sid"`
		Via         string `json:"via" form:"via"`
	}

	// Try JSON first, then form data
	if err := ctx.BodyParser(&callback); err != nil {
		fmt.Printf("[iPaymu Callback] Failed to parse body: %v\n", err)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid callback data"})
	}

	fmt.Printf("[iPaymu Callback] trx_id=%s reference_id=%s status=%s status_code=%s via=%s\n",
		callback.TrxID, callback.ReferenceID, callback.Status, callback.StatusCode, callback.Via)

	// Find transaction by reference_id (our transaction ID)
	var transaction models.Transaction
	if err := initializers.Db.Where("id = ?", callback.ReferenceID).First(&transaction).Error; err != nil {
		fmt.Printf("[iPaymu Callback] Transaction not found: %s\n", callback.ReferenceID)
		return ctx.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	// Store iPaymu trx_id
	if callback.TrxID != "" {
		transaction.ExternalID = callback.TrxID
	}
	if callback.Via != "" {
		transaction.PaymentMethod = "ipaymu:" + callback.Via
	}

	// Only process if still pending
	if transaction.Status != models.TransactionStatusPending {
		fmt.Printf("[iPaymu Callback] Transaction already processed: %s (status: %s)\n", transaction.ID, transaction.Status)
		return ctx.JSON(fiber.Map{"status": "already_processed"})
	}

	// iPaymu status_code: 1 = success, others = failed
	if callback.StatusCode == "1" || strings.ToLower(callback.Status) == "berhasil" {
		var user models.User
		if err := initializers.Db.Where("id = ?", transaction.UserID).First(&user).Error; err != nil {
			fmt.Printf("[iPaymu Callback] User not found: %s\n", transaction.UserID)
			return ctx.Status(500).JSON(fiber.Map{"error": "User not found"})
		}

		now := time.Now()
		transaction.Status = models.TransactionStatusSettlement
		transaction.PaidAt = &now
		initializers.Db.Save(&transaction)

		settleTransaction(&transaction, &user)

		fmt.Printf("[iPaymu Callback] Payment successful: tx=%s user=%s\n", transaction.ID, user.ID)
	} else {
		transaction.Status = models.TransactionStatusFailed
		initializers.Db.Save(&transaction)
		fmt.Printf("[iPaymu Callback] Payment failed: tx=%s status=%s\n", transaction.ID, callback.Status)
	}

	return ctx.JSON(fiber.Map{"status": "ok"})
}

// ========================================
// Check Transaction Status
// ========================================

// CheckTransaction checks iPaymu transaction status
func (c *IPaymuController) CheckTransaction(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	txID := ctx.Params("id")
	if txID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Transaction ID is required"})
	}

	var transaction models.Transaction
	if err := initializers.Db.Where("id = ? AND user_id = ?", txID, user.ID).First(&transaction).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	// If still pending and has external ID, check with iPaymu
	if transaction.Status == models.TransactionStatusPending && transaction.ExternalID != "" {
		checkBody, _ := json.Marshal(map[string]interface{}{
			"transactionId": transaction.ExternalID,
		})

		result, err := callIPaymuAPI("/api/v2/transaction", checkBody)
		if err == nil {
			status, _ := result["Status"].(float64)
			if int(status) == 200 {
				if data, ok := result["Data"].(map[string]interface{}); ok {
					if ipaymuStatus, ok := data["Status"].(float64); ok {
						if int(ipaymuStatus) == 1 {
							// Payment success
							now := time.Now()
							transaction.Status = models.TransactionStatusSettlement
							transaction.PaidAt = &now
							initializers.Db.Save(&transaction)
							settleTransaction(&transaction, user)
						} else if int(ipaymuStatus) == -1 {
							// Payment expired
							transaction.Status = models.TransactionStatusExpired
							initializers.Db.Save(&transaction)
						}
					}
				}
			}
		}
	}

	return ctx.JSON(fiber.Map{
		"id":     transaction.ID,
		"status": transaction.Status,
		"amount": transaction.Amount,
		"type":   transaction.TxType,
	})
}

// ========================================
// Settlement Logic (shared)
// ========================================

// settleTransaction processes a settled transaction (awards credits or activates subscription)
func settleTransaction(transaction *models.Transaction, user *models.User) {
	switch transaction.TxType {
	case models.TxTypeExtraCredit:
		if transaction.PackageID == nil {
			return
		}
		var pkg models.Package
		if err := initializers.Db.Where("id = ?", *transaction.PackageID).First(&pkg).Error; err != nil {
			fmt.Printf("[settle] Package not found: %v\n", err)
			return
		}
		// Award extra credits
		user.Credit += pkg.CreditAmount
		initializers.Db.Save(user)

		// Log credit
		creditLog := models.CreditLog{
			ID:          generateCreditLogID(),
			UserID:      user.ID,
			Amount:      pkg.CreditAmount,
			Balance:     user.TotalCredits(),
			Type:        models.CreditTypePurchase,
			ReferenceID: transaction.ID,
			Description: "Extra Credit Purchase: " + pkg.PackageName,
		}
		initializers.Db.Create(&creditLog)
		fmt.Printf("[settle] Awarded %d extra credits to user %s\n", pkg.CreditAmount, user.ID)

	case models.TxTypeSubscription:
		if transaction.PlanID == nil {
			return
		}
		var plan models.SubscriptionPlan
		if err := initializers.Db.Where("id = ?", *transaction.PlanID).First(&plan).Error; err != nil {
			fmt.Printf("[settle] Plan not found: %v\n", err)
			return
		}
		// Activate subscription
		now := time.Now()
		expiresAt := now.AddDate(0, 0, plan.BillingPeriodDays)

		// If already has active subscription, extend from current expiry
		if user.HasActiveSubscription() && user.SubscriptionExpiresAt.After(now) {
			expiresAt = user.SubscriptionExpiresAt.AddDate(0, 0, plan.BillingPeriodDays)
		}

		user.SubscriptionPlanID = &plan.ID
		user.SubscriptionExpiresAt = &expiresAt
		// Reset free credits to new plan's daily amount
		user.FreeCredit = plan.DailyFreeCredits
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		user.FreeCreditResetAt = &today
		initializers.Db.Save(user)

		// Log
		creditLog := models.CreditLog{
			ID:          generateCreditLogID(),
			UserID:      user.ID,
			Amount:      0,
			Balance:     user.TotalCredits(),
			Type:        models.CreditTypeSubscription,
			ReferenceID: transaction.ID,
			Description: fmt.Sprintf("Subscription: %s (%d days, %d daily credits, %d min rooms)",
				plan.PlanName, plan.BillingPeriodDays, plan.DailyFreeCredits, plan.RoomDurationMinutes),
		}
		initializers.Db.Create(&creditLog)
		fmt.Printf("[settle] Activated subscription '%s' for user %s (expires %s)\n",
			plan.PlanName, user.ID, expiresAt.Format("2006-01-02"))
	}
}

// ========================================
// Daily Free Credit Reset
// ========================================

// ResetFreeCreditIfNeeded checks and resets the user's daily free credits if needed
func ResetFreeCreditIfNeeded(user *models.User) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Check if already reset today
	if user.FreeCreditResetAt != nil && !user.FreeCreditResetAt.Before(today) {
		return // Already reset today
	}

	// Determine daily free credits based on subscription
	dailyCredits := GetDefaultFreeCredits()

	if user.SubscriptionPlanID != nil {
		if user.SubscriptionExpiresAt != nil && user.SubscriptionExpiresAt.After(now) {
			// Active subscription
			var plan models.SubscriptionPlan
			if err := initializers.Db.Where("id = ?", *user.SubscriptionPlanID).First(&plan).Error; err == nil {
				dailyCredits = plan.DailyFreeCredits
			}
		} else {
			// Subscription expired - revert to free plan
			user.SubscriptionPlanID = nil
			user.SubscriptionExpiresAt = nil
		}
	}

	user.FreeCredit = dailyCredits
	user.FreeCreditResetAt = &today
	initializers.Db.Save(user)
}

// GetDefaultFreeCredits returns the default daily free credits for free plan users
func GetDefaultFreeCredits() int {
	value := GetConfigValue(models.ConfigDailyFreeCredits, "5")
	credits, err := strconv.Atoi(value)
	if err != nil {
		return 5
	}
	return credits
}

// GetUserRoomDuration returns the room duration in minutes for a given user
func GetUserRoomDuration(user *models.User) int {
	if user.HasActiveSubscription() {
		var plan models.SubscriptionPlan
		if err := initializers.Db.Where("id = ?", *user.SubscriptionPlanID).First(&plan).Error; err == nil {
			return plan.RoomDurationMinutes
		}
	}
	return GetRoomMaxDuration()
}
