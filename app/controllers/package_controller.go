package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/gofiber/fiber/v2"
)

type PackageController struct{}

// ========================================
// Public: List Subscription Plans
// ========================================

// ListSubscriptionPlans returns visible subscription plans for users
func (c *PackageController) ListSubscriptionPlans(ctx *fiber.Ctx) error {
	var plans []models.SubscriptionPlan
	if err := initializers.Db.Where("visibility = ?", true).Order("sort_order ASC, price ASC").Find(&plans).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch subscription plans"})
	}

	type PlanResponse struct {
		ID                  string `json:"id"`
		PlanName            string `json:"plan_name"`
		PlanDetail          string `json:"plan_detail"`
		Price               int64  `json:"price"` // IDR
		BillingPeriodDays   int    `json:"billing_period_days"`
		DailyFreeCredits    int    `json:"daily_free_credits"`
		RoomDurationMinutes int    `json:"room_duration_minutes"`
	}

	response := make([]PlanResponse, len(plans))
	for i, plan := range plans {
		response[i] = PlanResponse{
			ID:                  plan.ID,
			PlanName:            plan.PlanName,
			PlanDetail:          string(plan.PlanDetail),
			Price:               plan.Price,
			BillingPeriodDays:   plan.BillingPeriodDays,
			DailyFreeCredits:    plan.DailyFreeCredits,
			RoomDurationMinutes: plan.RoomDurationMinutes,
		}
	}

	return ctx.JSON(response)
}

// ========================================
// Public: List Extra Credit Packages
// ========================================

// ListPublic returns all visible extra credit packages for users
func (c *PackageController) ListPublic(ctx *fiber.Ctx) error {
	var packages []models.Package
	if err := initializers.Db.Where("visibility = ?", true).Order("credit_amount ASC").Find(&packages).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch packages"})
	}

	type PackageResponse struct {
		ID            string `json:"id"`
		PackageName   string `json:"package_name"`
		PackageDetail string `json:"package_detail"`
		Price         int64  `json:"price"` // IDR
		CreditAmount  int    `json:"credit_amount"`
	}

	response := make([]PackageResponse, len(packages))
	for i, pkg := range packages {
		response[i] = PackageResponse{
			ID:            pkg.ID,
			PackageName:   pkg.PackageName,
			PackageDetail: string(pkg.PackageDetail),
			Price:         pkg.Price,
			CreditAmount:  pkg.CreditAmount,
		}
	}

	return ctx.JSON(response)
}

// ========================================
// My Transactions
// ========================================

func (c *PackageController) MyTransactions(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	status := ctx.Query("status", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := initializers.Db.Where("user_id = ?", user.ID).
		Preload("Package").Preload("Plan")
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	var transactions []models.Transaction
	var total int64

	query.Model(&models.Transaction{}).Count(&total)
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch transactions"})
	}

	type TransactionResponse struct {
		ID            string `json:"id"`
		ItemName      string `json:"item_name"`
		CreditAmount  int    `json:"credit_amount"`
		Amount        int64  `json:"amount"` // IDR
		Status        string `json:"status"`
		PaymentMethod string `json:"payment_method"`
		TxType        string `json:"tx_type"`
		CreatedAt     string `json:"created_at"`
	}

	response := make([]TransactionResponse, len(transactions))
	for i, tx := range transactions {
		itemName := ""
		creditAmt := 0
		if tx.TxType == models.TxTypeExtraCredit && tx.Package != nil {
			itemName = tx.Package.PackageName
			creditAmt = tx.Package.CreditAmount
		} else if tx.TxType == models.TxTypeSubscription && tx.Plan != nil {
			itemName = tx.Plan.PlanName + " Subscription"
		}
		response[i] = TransactionResponse{
			ID:            tx.ID,
			ItemName:      itemName,
			CreditAmount:  creditAmt,
			Amount:        tx.Amount,
			Status:        tx.Status,
			PaymentMethod: tx.PaymentMethod,
			TxType:        tx.TxType,
			CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	hasMore := int64(offset+limit) < total

	return ctx.JSON(fiber.Map{
		"transactions": response,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"has_more":     hasMore,
	})
}

// ========================================
// Get Transaction
// ========================================

func (c *PackageController) GetTransaction(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	txID := ctx.Params("id")
	if txID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Transaction ID is required"})
	}

	var transaction models.Transaction
	if err := initializers.Db.Where("id = ? AND user_id = ?", txID, user.ID).
		Preload("Package").Preload("Plan").First(&transaction).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	itemName := ""
	creditAmt := 0
	if transaction.TxType == models.TxTypeExtraCredit && transaction.Package != nil {
		itemName = transaction.Package.PackageName
		creditAmt = transaction.Package.CreditAmount
	} else if transaction.TxType == models.TxTypeSubscription && transaction.Plan != nil {
		itemName = transaction.Plan.PlanName + " Subscription"
	}

	return ctx.JSON(fiber.Map{
		"id":             transaction.ID,
		"item_name":      itemName,
		"credit_amount":  creditAmt,
		"amount":         transaction.Amount,
		"status":         transaction.Status,
		"payment_method": transaction.PaymentMethod,
		"tx_type":        transaction.TxType,
		"external_id":    transaction.ExternalID,
		"created_at":     transaction.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ========================================
// Get My Credits (with subscription info)
// ========================================

func (c *PackageController) GetMyCredits(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Reset free credits if needed
	ResetFreeCreditIfNeeded(user)

	// Get subscription info
	var subscriptionInfo fiber.Map
	if user.HasActiveSubscription() {
		var plan models.SubscriptionPlan
		if err := initializers.Db.Where("id = ?", *user.SubscriptionPlanID).First(&plan).Error; err == nil {
			subscriptionInfo = fiber.Map{
				"plan_name":            plan.PlanName,
				"daily_free_credits":   plan.DailyFreeCredits,
				"room_duration_minutes": plan.RoomDurationMinutes,
				"expires_at":           user.SubscriptionExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
	}

	var creditLogs []models.CreditLog
	initializers.Db.Where("user_id = ?", user.ID).Order("created_at DESC").Limit(50).Find(&creditLogs)

	return ctx.JSON(fiber.Map{
		"extra_credit":    user.Credit,
		"free_credit":     user.FreeCredit,
		"total_credit":    user.TotalCredits(),
		"subscription":    subscriptionInfo,
		"room_duration":   GetUserRoomDuration(user),
		"history":         creditLogs,
	})
}

// ========================================
// Helpers
// ========================================

func generateCreditLogID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateTransactionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
