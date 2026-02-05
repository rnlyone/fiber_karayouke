package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

type StripeController struct{}

// Stripe configuration - loaded from environment variables
func getStripeSecretKey() string {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		// Fallback to test key for development
		key = "sk_test_51SwiWs04TpTRkqrXNMnYSC9ZPw9JN4Ie1NzikMIgfbW99mIq57fOxp5RR0ImV10tWFmcZmqPtwkeL2hmN2lDUrsY00Qf2DGt45"
	}
	return key
}

func getStripeWebhookSecret() string {
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if secret == "" {
		// Fallback to test webhook secret for development
		secret = "whsec_3e82ac60fe885aadcdfb14d0433b5be05e952556cb591ddc8bda747937f7a449"
	}
	return secret
}

func getStripePublishableKey() string {
	key := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	if key == "" {
		// Fallback to test key for development
		key = "pk_test_51SwiWs04TpTRkqrXNnrFjFmx0mYYNnVbPwYjzNsgbuhxDQ87EHdGFi3K3ek1CcWup4bmcLufKUTVi7KyU4TPNNsR00HKlP0Klh"
	}
	return key
}

// GetPublishableKey returns the Stripe publishable key for the frontend
func (c *StripeController) GetPublishableKey(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"publishableKey": getStripePublishableKey(),
	})
}

// GetCurrencyFromIP detects user's currency based on their IP address using ipapi.co
func GetCurrencyFromIP(ip string) string {
	// Skip for localhost/private IPs
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "usd" // Default for local development
	}

	// Call ipapi.co free API
	resp, err := http.Get(fmt.Sprintf("https://ipapi.co/%s/json/", ip))
	if err != nil {
		fmt.Printf("Error getting IP info: %v\n", err)
		return "usd" // Default on error
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("IP API returned status: %d\n", resp.StatusCode)
		return "usd"
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading IP API response: %v\n", err)
		return "usd"
	}

	var result struct {
		CountryCode string `json:"country_code"`
		Currency    string `json:"currency"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parsing IP API response: %v\n", err)
		return "usd"
	}

	// Map currency codes to supported currencies
	currency := strings.ToLower(result.Currency)
	switch currency {
	case "jpy":
		return "jpy"
	case "idr":
		return "idr"
	case "usd":
		return "usd"
	default:
		// For unsupported currencies, map by country code
		switch result.CountryCode {
		case "JP":
			return "jpy"
		case "ID":
			return "idr"
		case "US":
			return "usd"
		default:
			return "usd" // Default to USD for other countries
		}
	}
}

// CreatePaymentIntent creates a Stripe PaymentIntent for a package purchase
func (c *StripeController) CreatePaymentIntent(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req struct {
		PackageID string `json:"package_id"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.PackageID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Package ID is required"})
	}

	// Get the package
	var pkg models.Package
	if err := initializers.Db.Where("id = ? AND visibility = ?", req.PackageID, true).First(&pkg).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Package not found"})
	}

	// Detect currency from user's IP address
	clientIP := ctx.IP()
	currency := GetCurrencyFromIP(clientIP)
	fmt.Printf("Detected currency '%s' for IP: %s\n", currency, clientIP)

	// Parse price (stored in cents)
	price, _ := strconv.ParseInt(pkg.Price, 10, 64)
	if price < 0 {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid package price"})
	}

	// Handle free packages - skip payment gateway
	if price == 0 {
		return c.handleFreePackage(ctx, user, &pkg)
	}

	// Convert price based on currency (price is stored as USD cents)
	// Use proper rounding to avoid truncation issues
	amount := price
	switch currency {
	case "jpy":
		// JPY doesn't use cents, convert from USD cents to JPY
		// Using approximate rate: 1 USD = 149 JPY
		// Round to nearest integer to avoid truncation issues
		amount = (price*149 + 50) / 100 // Add 50 for proper rounding
		// Ensure minimum amount for JPY (50 JPY)
		if amount < 50 {
			amount = 50
		}
	case "idr":
		// IDR: 1 USD = ~15800 IDR
		// Round to nearest integer
		amount = (price*15800 + 50) / 100
		// Ensure minimum amount for IDR (1000 IDR)
		if amount < 1000 {
			amount = 1000
		}
	default:
		currency = "usd"
		// Already in cents, ensure minimum (50 cents)
		if amount < 50 {
			amount = 50
		}
	}

	// Set Stripe API key
	stripe.Key = getStripeSecretKey()

	// Create a transaction record first
	txID := generateTransactionID()
	transaction := models.Transaction{
		ID:            txID,
		UserID:        user.ID,
		PackageID:     pkg.ID,
		Amount:        price, // Store original USD cents
		Status:        models.TransactionStatusPending,
		PaymentMethod: "stripe",
	}

	if err := initializers.Db.Create(&transaction).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	// Create PaymentIntent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: map[string]string{
			"transaction_id": txID,
			"package_id":     pkg.ID,
			"user_id":        user.ID,
			"package_name":   pkg.PackageName,
			"credit_amount":  strconv.Itoa(pkg.CreditAmount),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		// Rollback transaction
		initializers.Db.Delete(&transaction)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create payment intent: " + err.Error()})
	}

	// Update transaction with Stripe payment intent ID
	transaction.ExternalID = pi.ID
	initializers.Db.Save(&transaction)

	return ctx.JSON(fiber.Map{
		"clientSecret":    pi.ClientSecret,
		"transactionId":   txID,
		"amount":          amount,
		"currency":        currency,
		"packageName":     pkg.PackageName,
		"creditAmount":    pkg.CreditAmount,
		"paymentIntentId": pi.ID,
	})
}

// handleFreePackage processes free packages (price = 0) without payment gateway
func (c *StripeController) handleFreePackage(ctx *fiber.Ctx, user *models.User, pkg *models.Package) error {
	// Create a completed transaction directly
	txID := generateTransactionID()
	transaction := models.Transaction{
		ID:            txID,
		UserID:        user.ID,
		PackageID:     pkg.ID,
		Amount:        0,
		Status:        models.TransactionStatusSettlement,
		PaymentMethod: "free",
	}

	if err := initializers.Db.Create(&transaction).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	// Award credits immediately
	user.Credit += pkg.CreditAmount
	if err := initializers.Db.Save(user).Error; err != nil {
		// Rollback transaction
		initializers.Db.Delete(&transaction)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to award credits"})
	}

	return ctx.JSON(fiber.Map{
		"free":          true,
		"transactionId": txID,
		"packageName":   pkg.PackageName,
		"creditAmount":  pkg.CreditAmount,
		"message":       "Free package claimed successfully",
	})
}

// HandleWebhook processes Stripe webhook events
func (c *StripeController) HandleWebhook(ctx *fiber.Ctx) error {
	payload := ctx.Body()
	sigHeader := ctx.Get("Stripe-Signature")

	event, err := webhook.ConstructEvent(payload, sigHeader, getStripeWebhookSecret())
	if err != nil {
		fmt.Printf("Webhook signature verification failed: %v\n", err)
		return ctx.Status(400).JSON(fiber.Map{"error": "Webhook signature verification failed"})
	}

	// Handle the event
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			fmt.Printf("Error parsing webhook JSON: %v\n", err)
			return ctx.Status(400).JSON(fiber.Map{"error": "Error parsing webhook JSON"})
		}
		c.handlePaymentSuccess(&paymentIntent)

	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			fmt.Printf("Error parsing webhook JSON: %v\n", err)
			return ctx.Status(400).JSON(fiber.Map{"error": "Error parsing webhook JSON"})
		}
		c.handlePaymentFailure(&paymentIntent)

	case "payment_intent.canceled":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			fmt.Printf("Error parsing webhook JSON: %v\n", err)
			return ctx.Status(400).JSON(fiber.Map{"error": "Error parsing webhook JSON"})
		}
		c.handlePaymentFailure(&paymentIntent)

	case "payment_intent.requires_action":
		// Payment requires additional action (e.g., 3D Secure)
		// Log but don't change status - user needs to complete action
		fmt.Printf("Payment intent requires action: %s\n", event.Type)

	default:
		fmt.Printf("Unhandled event type: %s\n", event.Type)
	}

	return ctx.JSON(fiber.Map{"received": true})
}

// handlePaymentSuccess processes successful payments
func (c *StripeController) handlePaymentSuccess(pi *stripe.PaymentIntent) {
	transactionID := pi.Metadata["transaction_id"]
	if transactionID == "" {
		fmt.Println("No transaction_id in payment intent metadata")
		return
	}

	var transaction models.Transaction
	if err := initializers.Db.Where("id = ?", transactionID).First(&transaction).Error; err != nil {
		fmt.Printf("Transaction not found: %s\n", transactionID)
		return
	}

	// Already processed
	if transaction.Status == models.TransactionStatusSettlement {
		return
	}

	// Get user
	var user models.User
	if err := initializers.Db.Where("id = ?", transaction.UserID).First(&user).Error; err != nil {
		fmt.Printf("User not found: %s\n", transaction.UserID)
		return
	}

	// Get package
	var pkg models.Package
	if err := initializers.Db.Where("id = ?", transaction.PackageID).First(&pkg).Error; err != nil {
		fmt.Printf("Package not found: %s\n", transaction.PackageID)
		return
	}

	// Update transaction status
	transaction.Status = models.TransactionStatusSettlement
	if err := initializers.Db.Save(&transaction).Error; err != nil {
		fmt.Printf("Failed to update transaction: %v\n", err)
		return
	}

	// Award credits
	user.Credit += pkg.CreditAmount
	if err := initializers.Db.Save(&user).Error; err != nil {
		fmt.Printf("Failed to update user credits: %v\n", err)
		return
	}

	// Log credit
	creditLog := models.CreditLog{
		ID:          generateCreditLogID(),
		UserID:      user.ID,
		Amount:      pkg.CreditAmount,
		Balance:     user.Credit,
		Type:        models.CreditTypePurchase,
		ReferenceID: transaction.ID,
		Description: "Purchase: " + pkg.PackageName + " (Stripe)",
	}
	initializers.Db.Create(&creditLog)

	fmt.Printf("Payment successful: Transaction %s, User %s, Credits +%d\n", transactionID, user.ID, pkg.CreditAmount)
}

// handlePaymentFailure processes failed payments
func (c *StripeController) handlePaymentFailure(pi *stripe.PaymentIntent) {
	transactionID := pi.Metadata["transaction_id"]
	if transactionID == "" {
		fmt.Println("No transaction_id in payment intent metadata")
		return
	}

	var transaction models.Transaction
	if err := initializers.Db.Where("id = ?", transactionID).First(&transaction).Error; err != nil {
		fmt.Printf("Transaction not found: %s\n", transactionID)
		return
	}

	// Update transaction status
	transaction.Status = models.TransactionStatusFailed
	initializers.Db.Save(&transaction)

	fmt.Printf("Payment failed: Transaction %s\n", transactionID)
}

// ConfirmPayment is called by frontend after successful Stripe payment to verify status
func (c *StripeController) ConfirmPayment(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req struct {
		PaymentIntentID string `json:"payment_intent_id"`
		TransactionID   string `json:"transaction_id"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Set Stripe API key
	stripe.Key = getStripeSecretKey()

	// Retrieve the payment intent from Stripe
	pi, err := paymentintent.Get(req.PaymentIntentID, nil)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Failed to retrieve payment intent"})
	}

	// Verify the payment intent belongs to this transaction
	if pi.Metadata["transaction_id"] != req.TransactionID {
		return ctx.Status(400).JSON(fiber.Map{"error": "Payment intent mismatch"})
	}

	// Check payment status
	var transaction models.Transaction
	if err := initializers.Db.Where("id = ?", req.TransactionID).First(&transaction).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	// Handle different payment intent statuses
	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded:
		// If payment succeeded but webhook hasn't processed yet, process it now
		if transaction.Status == models.TransactionStatusPending {
			c.handlePaymentSuccess(pi)
			// Reload transaction
			initializers.Db.Where("id = ?", req.TransactionID).First(&transaction)
		}
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		// Payment method failed (card declined, insufficient funds, etc.)
		// This is a failed state - the user needs to try again with a different method
		if transaction.Status == models.TransactionStatusPending {
			c.handlePaymentFailure(pi)
			// Reload transaction
			initializers.Db.Where("id = ?", req.TransactionID).First(&transaction)
		}
	case stripe.PaymentIntentStatusCanceled:
		// Payment was canceled
		if transaction.Status == models.TransactionStatusPending {
			c.handlePaymentFailure(pi)
			// Reload transaction
			initializers.Db.Where("id = ?", req.TransactionID).First(&transaction)
		}
	case stripe.PaymentIntentStatusRequiresAction:
		// Still waiting for customer action (e.g., 3D Secure)
		// Keep as pending
	case stripe.PaymentIntentStatusProcessing:
		// Payment is being processed (e.g., bank transfer)
		// Keep as pending
	}

	return ctx.JSON(fiber.Map{
		"status":        string(transaction.Status),
		"paymentStatus": string(pi.Status),
		"transactionId": transaction.ID,
	})
}

func generateTransactionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
