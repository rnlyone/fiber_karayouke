package controllers

import (
	"encoding/base64"
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

type FlipController struct{}

// ========================================
// Flip Configuration
// ========================================

const (
	flipProductionURL = "https://bigflip.id/api"
	flipSandboxURL    = "https://bigflip.id/big_sandbox_api"
)

func getFlipSecretKey() string {
	return GetConfigValue(models.ConfigFlipSecretKey, "")
}

func getFlipValidationToken() string {
	return GetConfigValue(models.ConfigFlipValidationToken, "")
}

func getFlipEnvironment() string {
	return GetConfigValue(models.ConfigFlipEnvironment, "sandbox")
}

func getFlipBaseURL() string {
	if getFlipEnvironment() == "production" {
		return flipProductionURL
	}
	return flipSandboxURL
}

func getFlipAuthHeader() string {
	secretKey := getFlipSecretKey()
	// Basic Auth: Base64Encode("secret_key" + ":")
	encoded := base64.StdEncoding.EncodeToString([]byte(secretKey + ":"))
	return "Basic " + encoded
}

// ========================================
// Flip API Call Helper
// ========================================

func callFlipAPI(endpoint string, body string, contentType string) (map[string]interface{}, error) {
	baseURL := getFlipBaseURL()
	url := baseURL + endpoint

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", getFlipAuthHeader())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Flip API error: %v", err)
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

	fmt.Printf("[Flip] %s response (status %d): %s\n", endpoint, resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		msg := "Flip API error"
		if m, ok := result["message"].(string); ok {
			msg = m
		}
		return result, fmt.Errorf("%s (HTTP %d)", msg, resp.StatusCode)
	}

	return result, nil
}

// callFlipAPIGet makes a GET request to Flip API
func callFlipAPIGet(endpoint string) (map[string]interface{}, error) {
	baseURL := getFlipBaseURL()
	url := baseURL + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", getFlipAuthHeader())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Flip API error: %v", err)
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

	fmt.Printf("[Flip] GET %s response (status %d): %s\n", endpoint, resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		msg := "Flip API error"
		if m, ok := result["message"].(string); ok {
			msg = m
		}
		return result, fmt.Errorf("%s (HTTP %d)", msg, resp.StatusCode)
	}

	return result, nil
}

// ========================================
// Create Bill (Flip Checkout PopUp)
// ========================================

// CreateBill creates a Flip bill for the popup checkout flow
func (c *FlipController) CreateBill(ctx *fiber.Ctx) error {
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
		PaymentMethod: "flip",
		TxType:        txType,
	}

	if err := initializers.Db.Create(&transaction).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	// Build redirect URL
	baseURL := ctx.Protocol() + "://" + ctx.Hostname()
	origin := ctx.Get("Origin")
	if origin != "" {
		baseURL = origin
	}
	redirectURL := baseURL + "/payment/status/" + txID

	// Build Flip Create Bill V3 request (JSON body, supports popup checkout)
	billBody, _ := json.Marshal(map[string]interface{}{
		"title":        productName,
		"type":         "SINGLE",
		"amount":       price,
		"step":         "checkout_seamless",
		"redirect_url": redirectURL,
		"reference_id": txID,
		"sender_name":  user.Name,
		"sender_email": user.Email,
	})

	result, flipErr := callFlipAPI("/v3/pwf/bill", string(billBody), "application/json")
	if flipErr != nil {
		// Rollback transaction
		initializers.Db.Delete(&transaction)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create bill: " + flipErr.Error()})
	}

	// Extract company_code, product_code, link_id, and link_url from response
	companyCode := ""
	productCode := ""
	linkID := ""
	linkURL := ""

	if cc, ok := result["company_code"].(string); ok {
		companyCode = cc
	}
	if pc, ok := result["product_code"].(string); ok {
		productCode = pc
	}
	if lu, ok := result["link_url"].(string); ok {
		linkURL = lu
	}

	// link_id can be float64 or string depending on API version
	switch v := result["link_id"].(type) {
	case float64:
		linkID = fmt.Sprintf("%.0f", v)
	case string:
		linkID = v
	}

	// Update transaction with Flip reference
	if linkID != "" {
		transaction.ExternalID = linkID
	}
	transaction.FlipURL = linkURL
	transaction.FlipCompanyCode = companyCode
	transaction.FlipProductCode = productCode
	initializers.Db.Save(&transaction)

	return ctx.JSON(fiber.Map{
		"transaction_id": txID,
		"company_code":   companyCode,
		"product_code":   productCode,
		"link_url":       linkURL,
		"amount":         price,
		"product":        productName,
	})
}

// handleFreeItem processes free packages/plans without payment
func (c *FlipController) handleFreeItem(ctx *fiber.Ctx, user *models.User, packageID *string, planID *string, txType string) error {
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
// Flip Accept Payment Callback
// ========================================

// HandleCallback processes Flip accept payment callbacks
// IMPORTANT: Flip requires HTTP 200 response. Non-200 causes retries (5x, 2min interval).
func (c *FlipController) HandleCallback(ctx *fiber.Ctx) error {
	// Log raw request for debugging
	rawBody := string(ctx.Body())
	contentType := ctx.Get("Content-Type")
	fmt.Printf("[Flip Callback] Received callback, Content-Type=%s, body length=%d\n", contentType, len(rawBody))
	fmt.Printf("[Flip Callback] Raw body: %s\n", rawBody)

	// Flip sends application/x-www-form-urlencoded with data=JSON&token=VALIDATION_TOKEN
	dataStr := ctx.FormValue("data")
	token := ctx.FormValue("token")

	// Fallback: if form parsing returned empty data, try parsing raw body as JSON
	// (in case Flip V3 sends JSON body instead of form-urlencoded)
	if dataStr == "" && len(rawBody) > 0 {
		fmt.Printf("[Flip Callback] FormValue 'data' is empty, trying raw body as JSON\n")
		// Check if the raw body is JSON (starts with '{')
		trimmed := strings.TrimSpace(rawBody)
		if strings.HasPrefix(trimmed, "{") {
			dataStr = trimmed
		}
	}

	if dataStr == "" {
		fmt.Printf("[Flip Callback] No callback data found in request\n")
		// Always return 200 to Flip
		return ctx.Status(200).JSON(fiber.Map{"status": "error", "message": "no data"})
	}

	// Validate token (only if we have a validation token configured)
	validationToken := getFlipValidationToken()
	if validationToken != "" && token != "" && token != validationToken {
		fmt.Printf("[Flip Callback] Invalid validation token: got=%s expected=%s\n", token, validationToken)
		// Still return 200 to avoid Flip retries
		return ctx.Status(200).JSON(fiber.Map{"status": "error", "message": "invalid token"})
	}

	// Parse callback data
	var callbackData struct {
		ID             string      `json:"id"`
		BillLinkID     interface{} `json:"bill_link_id"` // Can be int or string
		BillLink       string      `json:"bill_link"`
		BillTitle      string      `json:"bill_title"`
		ReferenceID    string      `json:"reference_id"`
		SenderName     string      `json:"sender_name"`
		SenderEmail    string      `json:"sender_email"`
		SenderBank     string      `json:"sender_bank"`
		SenderBankType string      `json:"sender_bank_type"`
		Amount         float64     `json:"amount"`
		Status         string      `json:"status"`
		CreatedAt      string      `json:"created_at"`
	}

	if err := json.Unmarshal([]byte(dataStr), &callbackData); err != nil {
		fmt.Printf("[Flip Callback] Failed to parse data JSON: %v, dataStr=%s\n", err, dataStr)
		return ctx.Status(200).JSON(fiber.Map{"status": "error", "message": "invalid json"})
	}

	fmt.Printf("[Flip Callback] Parsed: id=%s bill_link_id=%v reference_id=%s status=%s amount=%.0f sender_bank=%s\n",
		callbackData.ID, callbackData.BillLinkID, callbackData.ReferenceID, callbackData.Status, callbackData.Amount, callbackData.SenderBank)

	// Find transaction by reference_id (our transaction ID)
	var transaction models.Transaction
	txFound := false

	// Try finding by reference_id first (our txID)
	if callbackData.ReferenceID != "" {
		if err := initializers.Db.Where("id = ?", callbackData.ReferenceID).First(&transaction).Error; err == nil {
			txFound = true
		}
	}

	// Fallback: try finding by external_id (Flip's link_id)
	if !txFound && callbackData.BillLinkID != nil {
		billLinkIDStr := ""
		switch v := callbackData.BillLinkID.(type) {
		case float64:
			billLinkIDStr = fmt.Sprintf("%.0f", v)
		case string:
			billLinkIDStr = v
		}
		if billLinkIDStr != "" {
			if err := initializers.Db.Where("external_id = ?", billLinkIDStr).First(&transaction).Error; err == nil {
				txFound = true
			}
		}
	}

	if !txFound {
		fmt.Printf("[Flip Callback] Transaction not found: reference_id=%s bill_link_id=%v\n",
			callbackData.ReferenceID, callbackData.BillLinkID)
		return ctx.Status(200).JSON(fiber.Map{"status": "error", "message": "transaction not found"})
	}

	fmt.Printf("[Flip Callback] Found transaction: tx=%s current_status=%s\n", transaction.ID, transaction.Status)

	// Store Flip payment ID
	if callbackData.ID != "" {
		transaction.ExternalID = callbackData.ID
	}
	if callbackData.SenderBank != "" {
		bankType := callbackData.SenderBankType
		if bankType == "" {
			bankType = "bank"
		}
		transaction.PaymentMethod = "flip:" + callbackData.SenderBank + ":" + bankType
	}

	// Only process if still pending
	if transaction.Status != models.TransactionStatusPending {
		fmt.Printf("[Flip Callback] Transaction already processed: %s (status: %s)\n", transaction.ID, transaction.Status)
		return ctx.Status(200).JSON(fiber.Map{"status": "already_processed"})
	}

	// Flip status: SUCCESSFUL, CANCELLED, FAILED
	status := strings.ToUpper(callbackData.Status)
	if status == "SUCCESSFUL" {
		var user models.User
		if err := initializers.Db.Where("id = ?", transaction.UserID).First(&user).Error; err != nil {
			fmt.Printf("[Flip Callback] User not found: %s\n", transaction.UserID)
			return ctx.Status(200).JSON(fiber.Map{"status": "error", "message": "user not found"})
		}

		now := time.Now()
		transaction.Status = models.TransactionStatusSettlement
		transaction.PaidAt = &now
		initializers.Db.Save(&transaction)

		settleTransaction(&transaction, &user)

		fmt.Printf("[Flip Callback] ✅ Payment successful: tx=%s user=%s amount=%.0f\n", transaction.ID, user.ID, callbackData.Amount)
	} else if status == "CANCELLED" {
		transaction.Status = models.TransactionStatusExpired
		initializers.Db.Save(&transaction)
		fmt.Printf("[Flip Callback] ❌ Payment cancelled/expired: tx=%s\n", transaction.ID)
	} else if status == "FAILED" {
		transaction.Status = models.TransactionStatusFailed
		initializers.Db.Save(&transaction)
		fmt.Printf("[Flip Callback] ❌ Payment failed: tx=%s\n", transaction.ID)
	} else {
		fmt.Printf("[Flip Callback] Unknown status '%s' for tx=%s\n", callbackData.Status, transaction.ID)
	}

	return ctx.Status(200).JSON(fiber.Map{"status": "ok"})
}

// ========================================
// Check Transaction Status
// ========================================

// CheckTransaction checks Flip transaction status
// If the transaction is still pending and has a Flip link_id, actively checks Flip's API
func (c *FlipController) CheckTransaction(ctx *fiber.Ctx) error {
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

	// If still pending and we have a Flip link_id, actively check Flip's API
	if transaction.Status == models.TransactionStatusPending && transaction.ExternalID != "" {
		c.checkAndUpdateFlipPayment(&transaction, user)
	}

	return ctx.JSON(fiber.Map{
		"id":     transaction.ID,
		"status": transaction.Status,
		"amount": transaction.Amount,
		"type":   transaction.TxType,
	})
}

// checkAndUpdateFlipPayment calls Flip's Get Payment V3 API to verify the actual payment status
func (c *FlipController) checkAndUpdateFlipPayment(transaction *models.Transaction, user *models.User) {
	// Flip API: GET /v3/pwf/:bill_id/payment
	endpoint := fmt.Sprintf("/v3/pwf/%s/payment", transaction.ExternalID)
	result, err := callFlipAPIGet(endpoint)
	if err != nil {
		fmt.Printf("[Flip Check] Error checking payment for tx=%s link_id=%s: %v\n",
			transaction.ID, transaction.ExternalID, err)
		return
	}

	// Parse data array from response
	dataArr, ok := result["data"].([]interface{})
	if !ok || len(dataArr) == 0 {
		fmt.Printf("[Flip Check] No payments found for tx=%s link_id=%s\n",
			transaction.ID, transaction.ExternalID)
		return
	}

	// Check each payment (SINGLE type bill should have at most 1)
	for _, item := range dataArr {
		payment, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		status, _ := payment["status"].(string)
		status = strings.ToUpper(status)

		fmt.Printf("[Flip Check] Payment found for tx=%s: status=%s\n", transaction.ID, status)

		if status == "SUCCESSFUL" {
			now := time.Now()
			transaction.Status = models.TransactionStatusSettlement
			transaction.PaidAt = &now

			if senderBank, ok := payment["sender_bank"].(string); ok && senderBank != "" {
				bankType, _ := payment["sender_bank_type"].(string)
				if bankType == "" {
					bankType = "bank"
				}
				transaction.PaymentMethod = "flip:" + senderBank + ":" + bankType
			}

			initializers.Db.Save(transaction)
			settleTransaction(transaction, user)
			fmt.Printf("[Flip Check] ✅ Payment verified as successful via API: tx=%s user=%s\n",
				transaction.ID, user.ID)
			return
		} else if status == "CANCELLED" {
			transaction.Status = models.TransactionStatusExpired
			initializers.Db.Save(transaction)
			fmt.Printf("[Flip Check] ❌ Payment cancelled via API: tx=%s\n", transaction.ID)
			return
		} else if status == "FAILED" {
			transaction.Status = models.TransactionStatusFailed
			initializers.Db.Save(transaction)
			fmt.Printf("[Flip Check] ❌ Payment failed via API: tx=%s\n", transaction.ID)
			return
		}
		// PENDING or other status - do nothing, keep polling
	}
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
