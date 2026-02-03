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

// ListPublic returns all visible packages for users
func (c *PackageController) ListPublic(ctx *fiber.Ctx) error {
	var packages []models.Package
	if err := initializers.Db.Where("visibility = ?", true).Order("credit_amount ASC").Find(&packages).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch packages"})
	}

	// Convert price to int for frontend
	type PackageResponse struct {
		ID            string `json:"id"`
		PackageName   string `json:"package_name"`
		PackageDetail string `json:"package_detail"`
		Price         int64  `json:"price"`
		CreditAmount  int    `json:"credit_amount"`
	}

	response := make([]PackageResponse, len(packages))
	for i, pkg := range packages {
		price, _ := strconv.ParseInt(pkg.Price, 10, 64)
		response[i] = PackageResponse{
			ID:            pkg.ID,
			PackageName:   pkg.PackageName,
			PackageDetail: string(pkg.PackageDetail),
			Price:         price,
			CreditAmount:  pkg.CreditAmount,
		}
	}

	return ctx.JSON(response)
}

// Purchase initiates a package purchase
func (c *PackageController) Purchase(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	packageID := ctx.Params("id")
	if packageID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Package ID is required"})
	}

	var pkg models.Package
	if err := initializers.Db.Where("id = ? AND visibility = ?", packageID, true).First(&pkg).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Package not found"})
	}

	price, _ := strconv.ParseInt(pkg.Price, 10, 64)

	// Generate transaction ID
	txBytes := make([]byte, 16)
	rand.Read(txBytes)
	txID := hex.EncodeToString(txBytes)

	// Generate external ID for payment gateway
	extBytes := make([]byte, 8)
	rand.Read(extBytes)
	externalID := "KRY-" + hex.EncodeToString(extBytes)

	transaction := models.Transaction{
		ID:            txID,
		UserID:        user.ID,
		PackageID:     pkg.ID,
		Amount:        price,
		Status:        models.TransactionStatusPending,
		PaymentMethod: "manual", // Will be updated by payment gateway
		ExternalID:    externalID,
	}

	if err := initializers.Db.Create(&transaction).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	// If price is 0, auto-settle
	if price == 0 {
		return c.settleTransaction(&transaction, user)
	}

	// Return transaction for payment processing
	// In real implementation, this would return payment gateway URL
	return ctx.JSON(fiber.Map{
		"transaction_id": transaction.ID,
		"external_id":    transaction.ExternalID,
		"amount":         price,
		"status":         transaction.Status,
		"package":        pkg.PackageName,
		"credits":        pkg.CreditAmount,
		// payment_url would be added here when integrating payment gateway
	})
}

// settleTransaction completes a transaction and awards credits
func (c *PackageController) settleTransaction(transaction *models.Transaction, user *models.User) error {
	// Load package
	var pkg models.Package
	if err := initializers.Db.Where("id = ?", transaction.PackageID).First(&pkg).Error; err != nil {
		return err
	}

	// Update transaction status
	transaction.Status = models.TransactionStatusSettlement
	if err := initializers.Db.Save(transaction).Error; err != nil {
		return err
	}

	// Award credits
	user.Credit += pkg.CreditAmount
	if err := initializers.Db.Save(user).Error; err != nil {
		return err
	}

	// Log credit
	creditLog := models.CreditLog{
		ID:          generateCreditLogID(),
		UserID:      user.ID,
		Amount:      pkg.CreditAmount,
		Balance:     user.Credit,
		Type:        models.CreditTypePurchase,
		ReferenceID: transaction.ID,
		Description: "Purchase: " + pkg.PackageName,
	}
	initializers.Db.Create(&creditLog)

	return nil
}

// MyTransactions returns the current user's transactions
func (c *PackageController) MyTransactions(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var transactions []models.Transaction
	if err := initializers.Db.Where("user_id = ?", user.ID).Preload("Package").Order("created_at DESC").Find(&transactions).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch transactions"})
	}

	return ctx.JSON(transactions)
}

// GetMyCredits returns the current user's credit balance and history
func (c *PackageController) GetMyCredits(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var creditLogs []models.CreditLog
	initializers.Db.Where("user_id = ?", user.ID).Order("created_at DESC").Limit(50).Find(&creditLogs)

	return ctx.JSON(fiber.Map{
		"balance": user.Credit,
		"history": creditLogs,
	})
}

// PaymentCallback handles payment gateway callbacks
// This is a template for integrating external payment gateways
func (c *PackageController) PaymentCallback(ctx *fiber.Ctx) error {
	// Parse callback data from payment gateway
	var callback struct {
		ExternalID string `json:"external_id"`
		Status     string `json:"status"` // success, failed, pending
	}

	if err := ctx.BodyParser(&callback); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid callback data"})
	}

	var transaction models.Transaction
	if err := initializers.Db.Where("external_id = ?", callback.ExternalID).First(&transaction).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	// Only process if still pending
	if transaction.Status != models.TransactionStatusPending {
		return ctx.JSON(fiber.Map{"status": "already_processed"})
	}

	switch callback.Status {
	case "success":
		var user models.User
		if err := initializers.Db.Where("id = ?", transaction.UserID).First(&user).Error; err != nil {
			return ctx.Status(500).JSON(fiber.Map{"error": "User not found"})
		}
		if err := c.settleTransaction(&transaction, &user); err != nil {
			return ctx.Status(500).JSON(fiber.Map{"error": "Failed to settle transaction"})
		}
	case "failed":
		transaction.Status = models.TransactionStatusFailed
		initializers.Db.Save(&transaction)
	case "expired":
		transaction.Status = models.TransactionStatusExpired
		initializers.Db.Save(&transaction)
	}

	return ctx.JSON(fiber.Map{"status": "ok"})
}

func generateCreditLogID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
