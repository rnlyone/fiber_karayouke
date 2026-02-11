package controllers

import (
	"strconv"
	"time"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/gofiber/fiber/v2"
)

// AdminEmail is the email address with admin access
const AdminEmail = "me@ruzman.my.id"

type AdminController struct{}

// IsAdmin checks if the current user is an admin
func IsAdmin(ctx *fiber.Ctx) bool {
	user := GetUserFromToken(ctx)
	if user == nil {
		return false
	}
	return user.Email == AdminEmail
}

// AdminMiddleware restricts access to admin users only
func AdminMiddleware(ctx *fiber.Ctx) error {
	if !IsAdmin(ctx) {
		return ctx.Status(403).JSON(fiber.Map{"error": "Admin access required"})
	}
	return ctx.Next()
}

// CheckAdmin returns whether the current user is an admin
func (c *AdminController) CheckAdmin(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{"is_admin": IsAdmin(ctx)})
}

// ========================================
// SYSTEM CONFIG
// ========================================

type UpdateConfigRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (c *AdminController) GetConfigs(ctx *fiber.Ctx) error {
	var configs []models.SystemConfig
	if err := initializers.Db.Find(&configs).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch configs"})
	}

	// Convert to map for easier frontend consumption
	configMap := make(map[string]string)
	for _, config := range configs {
		configMap[config.Key] = config.Value
	}

	// Set defaults if not present
	defaults := map[string]string{
		models.ConfigRoomMaxDuration:  "120", // 2 hours default
		models.ConfigRoomCreationCost: "1",   // 1 credit default
		models.ConfigDefaultCredits:   "0",   // 0 credits for new users
		models.ConfigDailyFreeCredits: "5",   // 5 daily free credits for free plan
		models.ConfigFreeRoomDuration: "40",  // 40 min room for free plan
		models.ConfigIPaymuSandbox:    "true",
	}
	for key, defaultValue := range defaults {
		if _, exists := configMap[key]; !exists {
			configMap[key] = defaultValue
		}
	}

	return ctx.JSON(configMap)
}

func (c *AdminController) CreateConfig(ctx *fiber.Ctx) error {
	var req UpdateConfigRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Key == "" || req.Value == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Key and value are required"})
	}

	// Check if config already exists
	var existing models.SystemConfig
	if err := initializers.Db.Where("key = ?", req.Key).First(&existing).Error; err == nil {
		return ctx.Status(409).JSON(fiber.Map{"error": "Config with this key already exists"})
	}

	config := models.SystemConfig{
		ID:    generateID(),
		Key:   req.Key,
		Value: req.Value,
	}

	if err := initializers.Db.Create(&config).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create config"})
	}

	return ctx.Status(201).JSON(config)
}

func (c *AdminController) UpdateConfig(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	if key == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Key is required"})
	}

	var req struct {
		Value string `json:"value"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Value == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Value is required"})
	}

	var config models.SystemConfig
	if err := initializers.Db.Where("key = ?", key).First(&config).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Config not found"})
	}

	config.Value = req.Value
	if err := initializers.Db.Save(&config).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to update config"})
	}

	return ctx.JSON(config)
}

func (c *AdminController) DeleteConfig(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	if key == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Key is required"})
	}

	// Prevent deletion of critical configs
	criticalConfigs := []string{
		models.ConfigRoomMaxDuration,
		models.ConfigRoomCreationCost,
		models.ConfigDefaultCredits,
		models.ConfigDailyFreeCredits,
		models.ConfigFreeRoomDuration,
		models.ConfigIPaymuVA,
		models.ConfigIPaymuAPIKey,
		models.ConfigIPaymuSandbox,
	}
	for _, critical := range criticalConfigs {
		if key == critical {
			return ctx.Status(400).JSON(fiber.Map{"error": "Cannot delete critical system configuration"})
		}
	}

	var config models.SystemConfig
	if err := initializers.Db.Where("key = ?", key).First(&config).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Config not found"})
	}

	if err := initializers.Db.Delete(&config).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to delete config"})
	}

	return ctx.JSON(fiber.Map{"message": "Config deleted successfully"})
}

// ========================================
// PACKAGE MANAGEMENT
// ========================================

type CreatePackageRequest struct {
	PackageName   string `json:"package_name"`
	PackageDetail string `json:"package_detail"`
	Price         int64  `json:"price"`
	CreditAmount  int    `json:"credit_amount"`
	Visibility    bool   `json:"visibility"`
}

type UpdatePackageRequest struct {
	PackageName   string `json:"package_name"`
	PackageDetail string `json:"package_detail"`
	Price         int64  `json:"price"`
	CreditAmount  int    `json:"credit_amount"`
	Visibility    bool   `json:"visibility"`
}

func (c *AdminController) ListPackages(ctx *fiber.Ctx) error {
	var packages []models.Package
	if err := initializers.Db.Order("credit_amount ASC").Find(&packages).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch packages"})
	}
	return ctx.JSON(packages)
}

func (c *AdminController) CreatePackage(ctx *fiber.Ctx) error {
	var req CreatePackageRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.PackageName == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Package name is required"})
	}

	pkg := models.Package{
		ID:            generateID(),
		PackageName:   req.PackageName,
		PackageDetail: []byte(req.PackageDetail),
		Price:         req.Price,
		CreditAmount:  req.CreditAmount,
		Visibility:    req.Visibility,
	}

	if err := initializers.Db.Create(&pkg).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create package"})
	}

	return ctx.JSON(pkg)
}

func (c *AdminController) UpdatePackage(ctx *fiber.Ctx) error {
	packageID := ctx.Params("id")
	if packageID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Package ID is required"})
	}

	var pkg models.Package
	if err := initializers.Db.Where("id = ?", packageID).First(&pkg).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Package not found"})
	}

	var req UpdatePackageRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	pkg.PackageName = req.PackageName
	pkg.PackageDetail = []byte(req.PackageDetail)
	pkg.Price = req.Price
	pkg.CreditAmount = req.CreditAmount
	pkg.Visibility = req.Visibility

	if err := initializers.Db.Save(&pkg).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to update package"})
	}

	return ctx.JSON(pkg)
}

func (c *AdminController) DeletePackage(ctx *fiber.Ctx) error {
	packageID := ctx.Params("id")
	if packageID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Package ID is required"})
	}

	if err := initializers.Db.Where("id = ?", packageID).Delete(&models.Package{}).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to delete package"})
	}

	return ctx.JSON(fiber.Map{"success": true})
}

// ========================================
// USER MANAGEMENT
// ========================================

type UserListResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	ExtraCredit int     `json:"extra_credit"`
	FreeCredit  int     `json:"free_credit"`
	TotalCredit int     `json:"total_credit"`
	PlanName    *string `json:"plan_name"`
}

func (c *AdminController) ListUsers(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	search := ctx.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var users []models.User
	var total int64

	query := initializers.Db.Model(&models.User{})
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR username ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("name ASC").Find(&users).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
	}

	response := make([]UserListResponse, len(users))
	for i, user := range users {
		var planName *string
		if user.SubscriptionPlanID != nil {
			var plan models.SubscriptionPlan
			if err := initializers.Db.Where("id = ?", *user.SubscriptionPlanID).First(&plan).Error; err == nil {
				planName = &plan.PlanName
			}
		}
		response[i] = UserListResponse{
			ID:          user.ID,
			Name:        user.Name,
			Username:    user.Username,
			Email:       user.Email,
			ExtraCredit: user.Credit,
			FreeCredit:  user.FreeCredit,
			TotalCredit: user.TotalCredits(),
			PlanName:    planName,
		}
	}

	return ctx.JSON(fiber.Map{
		"users": response,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (total + int64(limit) - 1) / int64(limit),
	})
}

func (c *AdminController) GetUser(ctx *fiber.Ctx) error {
	userID := ctx.Params("id")
	if userID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "User ID is required"})
	}

	var user models.User
	if err := initializers.Db.Where("id = ?", userID).First(&user).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Get credit logs
	var creditLogs []models.CreditLog
	initializers.Db.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&creditLogs)

	var planName *string
	if user.SubscriptionPlanID != nil {
		var plan models.SubscriptionPlan
		if err := initializers.Db.Where("id = ?", *user.SubscriptionPlanID).First(&plan).Error; err == nil {
			planName = &plan.PlanName
		}
	}

	return ctx.JSON(fiber.Map{
		"user": UserListResponse{
			ID:          user.ID,
			Name:        user.Name,
			Username:    user.Username,
			Email:       user.Email,
			ExtraCredit: user.Credit,
			FreeCredit:  user.FreeCredit,
			TotalCredit: user.TotalCredits(),
			PlanName:    planName,
		},
		"credit_logs": creditLogs,
	})
}

// ========================================
// CREDITS AWARDING
// ========================================

type AwardCreditsRequest struct {
	UserID      string `json:"user_id"`
	Amount      int    `json:"amount"`
	CreditType  string `json:"credit_type"` // "extra" or "free"
	Description string `json:"description"`
}

func (c *AdminController) AwardCredits(ctx *fiber.Ctx) error {
	var req AwardCreditsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.UserID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "User ID is required"})
	}

	if req.Amount == 0 {
		return ctx.Status(400).JSON(fiber.Map{"error": "Amount cannot be zero"})
	}

	// Default to extra credit
	if req.CreditType == "" {
		req.CreditType = "extra"
	}
	if req.CreditType != "extra" && req.CreditType != "free" {
		return ctx.Status(400).JSON(fiber.Map{"error": "credit_type must be 'extra' or 'free'"})
	}

	var user models.User
	if err := initializers.Db.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	var newBalance int
	if req.CreditType == "free" {
		newBalance = user.FreeCredit + req.Amount
		if newBalance < 0 {
			return ctx.Status(400).JSON(fiber.Map{"error": "Cannot reduce free credits below zero"})
		}
		user.FreeCredit = newBalance
	} else {
		newBalance = user.Credit + req.Amount
		if newBalance < 0 {
			return ctx.Status(400).JSON(fiber.Map{"error": "Cannot reduce extra credits below zero"})
		}
		user.Credit = newBalance
	}

	if err := initializers.Db.Save(&user).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to update user credits"})
	}

	// Log the credit change
	description := req.Description
	if description == "" {
		if req.Amount > 0 {
			description = "Admin " + req.CreditType + " credit award"
		} else {
			description = "Admin " + req.CreditType + " credit deduction"
		}
	}

	creditLog := models.CreditLog{
		ID:          generateID(),
		UserID:      user.ID,
		Amount:      req.Amount,
		Balance:     user.TotalCredits(),
		Type:        models.CreditTypeAdminAward,
		ReferenceID: "",
		Description: description,
	}

	initializers.Db.Create(&creditLog)

	return ctx.JSON(fiber.Map{
		"success":      true,
		"credit_type":  req.CreditType,
		"new_balance":  newBalance,
		"total_credit": user.TotalCredits(),
		"credit_log":   creditLog,
	})
}

// ========================================
// TRANSACTIONS
// ========================================

func (c *AdminController) ListTransactions(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	status := ctx.Query("status", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var transactions []models.Transaction
	var total int64

	query := initializers.Db.Model(&models.Transaction{}).Preload("User").Preload("Package").Preload("Plan")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch transactions"})
	}

	return ctx.JSON(fiber.Map{
		"transactions": transactions,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"pages":        (total + int64(limit) - 1) / int64(limit),
	})
}

// UpdateTransactionStatus allows admin to manually update transaction status
func (c *AdminController) UpdateTransactionStatus(ctx *fiber.Ctx) error {
	transactionID := ctx.Params("id")
	if transactionID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Transaction ID is required"})
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	validStatuses := []string{
		models.TransactionStatusPending,
		models.TransactionStatusSettlement,
		models.TransactionStatusFailed,
		models.TransactionStatusExpired,
		models.TransactionStatusRefunded,
	}

	isValid := false
	for _, s := range validStatuses {
		if req.Status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid status"})
	}

	var transaction models.Transaction
	if err := initializers.Db.Where("id = ?", transactionID).Preload("Package").Preload("Plan").First(&transaction).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	oldStatus := transaction.Status
	transaction.Status = req.Status

	// If transitioning to settlement, use settleTransaction logic
	if oldStatus != models.TransactionStatusSettlement && req.Status == models.TransactionStatusSettlement {
		now := time.Now()
		transaction.PaidAt = &now

		var user models.User
		if err := initializers.Db.Where("id = ?", transaction.UserID).First(&user).Error; err == nil {
			if transaction.TxType == models.TxTypeSubscription && transaction.Plan != nil {
				// Activate subscription
				user.SubscriptionPlanID = &transaction.Plan.ID
				expiresAt := now.AddDate(0, 0, transaction.Plan.BillingPeriodDays)
				user.SubscriptionExpiresAt = &expiresAt
				initializers.Db.Save(&user)

				creditLog := models.CreditLog{
					ID:          generateID(),
					UserID:      user.ID,
					Amount:      0,
					Balance:     user.TotalCredits(),
					Type:        models.CreditTypeSubscription,
					ReferenceID: transaction.ID,
					Description: "Subscription activated: " + transaction.Plan.PlanName,
				}
				initializers.Db.Create(&creditLog)
			} else if transaction.Package != nil {
				// Award extra credits
				user.Credit += transaction.Package.CreditAmount
				initializers.Db.Save(&user)

				creditLog := models.CreditLog{
					ID:          generateID(),
					UserID:      user.ID,
					Amount:      transaction.Package.CreditAmount,
					Balance:     user.TotalCredits(),
					Type:        models.CreditTypePurchase,
					ReferenceID: transaction.ID,
					Description: "Purchase: " + transaction.Package.PackageName,
				}
				initializers.Db.Create(&creditLog)
			}
		}
	}

	if err := initializers.Db.Save(&transaction).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to update transaction"})
	}

	return ctx.JSON(transaction)
}

// ========================================
// SUBSCRIPTION PLAN MANAGEMENT
// ========================================

type CreateSubscriptionPlanRequest struct {
	PlanName            string `json:"plan_name"`
	PlanDetail          string `json:"plan_detail"`
	Price               int64  `json:"price"` // IDR
	BillingPeriodDays   int    `json:"billing_period_days"`
	DailyFreeCredits    int    `json:"daily_free_credits"`
	RoomDurationMinutes int    `json:"room_duration_minutes"`
	SortOrder           int    `json:"sort_order"`
	Visibility          bool   `json:"visibility"`
}

func (c *AdminController) ListSubscriptionPlans(ctx *fiber.Ctx) error {
	var plans []models.SubscriptionPlan
	if err := initializers.Db.Order("sort_order ASC, price ASC").Find(&plans).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch subscription plans"})
	}
	return ctx.JSON(plans)
}

func (c *AdminController) CreateSubscriptionPlan(ctx *fiber.Ctx) error {
	var req CreateSubscriptionPlanRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.PlanName == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Plan name is required"})
	}
	if req.BillingPeriodDays <= 0 {
		return ctx.Status(400).JSON(fiber.Map{"error": "Billing period must be positive"})
	}

	plan := models.SubscriptionPlan{
		ID:                  generateID(),
		PlanName:            req.PlanName,
		PlanDetail:          []byte(req.PlanDetail),
		Price:               req.Price,
		BillingPeriodDays:   req.BillingPeriodDays,
		DailyFreeCredits:    req.DailyFreeCredits,
		RoomDurationMinutes: req.RoomDurationMinutes,
		SortOrder:           req.SortOrder,
		Visibility:          req.Visibility,
	}

	if err := initializers.Db.Create(&plan).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create subscription plan"})
	}

	return ctx.Status(201).JSON(plan)
}

func (c *AdminController) UpdateSubscriptionPlan(ctx *fiber.Ctx) error {
	planID := ctx.Params("id")
	if planID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Plan ID is required"})
	}

	var plan models.SubscriptionPlan
	if err := initializers.Db.Where("id = ?", planID).First(&plan).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Subscription plan not found"})
	}

	var req CreateSubscriptionPlanRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	plan.PlanName = req.PlanName
	plan.PlanDetail = []byte(req.PlanDetail)
	plan.Price = req.Price
	plan.BillingPeriodDays = req.BillingPeriodDays
	plan.DailyFreeCredits = req.DailyFreeCredits
	plan.RoomDurationMinutes = req.RoomDurationMinutes
	plan.SortOrder = req.SortOrder
	plan.Visibility = req.Visibility

	if err := initializers.Db.Save(&plan).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to update subscription plan"})
	}

	return ctx.JSON(plan)
}

func (c *AdminController) DeleteSubscriptionPlan(ctx *fiber.Ctx) error {
	planID := ctx.Params("id")
	if planID == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Plan ID is required"})
	}

	// Check if any users are on this plan
	var count int64
	initializers.Db.Model(&models.User{}).Where("subscription_plan_id = ?", planID).Count(&count)
	if count > 0 {
		return ctx.Status(400).JSON(fiber.Map{
			"error":       "Cannot delete plan with active subscribers",
			"subscribers": count,
		})
	}

	if err := initializers.Db.Where("id = ?", planID).Delete(&models.SubscriptionPlan{}).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to delete subscription plan"})
	}

	return ctx.JSON(fiber.Map{"success": true})
}

// ========================================
// ROOMS MANAGEMENT
// ========================================

func (c *AdminController) ListRooms(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	status := ctx.Query("status", "") // active, expired, all

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var rooms []models.Room
	var total int64

	query := initializers.Db.Model(&models.Room{}).Preload("Creator")

	maxDuration := GetRoomMaxDuration()

	// Note: We can't filter by expiration in SQL anymore since it's calculated dynamically
	// We'll fetch all rooms and filter in memory if needed
	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&rooms).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch rooms"})
	}

	// Filter by status if requested (in memory since expiration is calculated)
	var filteredRooms []models.Room
	if status == "active" || status == "expired" {
		for _, room := range rooms {
			isExpired := room.IsExpired(maxDuration)
			if (status == "active" && !isExpired) || (status == "expired" && isExpired) {
				filteredRooms = append(filteredRooms, room)
			}
		}
	} else {
		filteredRooms = rooms
	}

	return ctx.JSON(fiber.Map{
		"rooms": filteredRooms,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// ========================================
// DASHBOARD STATS
// ========================================

func (c *AdminController) GetDashboardStats(ctx *fiber.Ctx) error {
	var totalUsers int64
	var totalRooms int64
	var activeRooms int64
	var expiredRooms int64
	var totalTransactions int64
	var pendingTransactions int64
	var settledTransactions int64
	var failedTransactions int64
	var totalPackages int64
	var totalRevenue int64
	var totalCreditsAwarded int64
	var totalSubscriptionPlans int64
	var activeSubscribers int64

	now := time.Now()

	initializers.Db.Model(&models.User{}).Count(&totalUsers)
	initializers.Db.Model(&models.Room{}).Count(&totalRooms)

	// Calculate active/expired rooms dynamically
	var allRooms []models.Room
	initializers.Db.Model(&models.Room{}).Find(&allRooms)
	maxDuration := GetRoomMaxDuration()
	for _, room := range allRooms {
		if room.IsExpired(maxDuration) {
			expiredRooms++
		} else {
			activeRooms++
		}
	}

	initializers.Db.Model(&models.Transaction{}).Count(&totalTransactions)
	initializers.Db.Model(&models.Transaction{}).Where("status = ?", models.TransactionStatusPending).Count(&pendingTransactions)
	initializers.Db.Model(&models.Transaction{}).Where("status = ?", models.TransactionStatusSettlement).Count(&settledTransactions)
	initializers.Db.Model(&models.Transaction{}).Where("status = ?", models.TransactionStatusFailed).Count(&failedTransactions)
	initializers.Db.Model(&models.Package{}).Count(&totalPackages)
	initializers.Db.Model(&models.Transaction{}).Where("status = ?", models.TransactionStatusSettlement).
		Select("COALESCE(SUM(amount),0)").Scan(&totalRevenue)
	initializers.Db.Model(&models.CreditLog{}).Where("amount > 0").
		Select("COALESCE(SUM(amount),0)").Scan(&totalCreditsAwarded)
	initializers.Db.Model(&models.SubscriptionPlan{}).Count(&totalSubscriptionPlans)
	initializers.Db.Model(&models.User{}).
		Where("subscription_plan_id IS NOT NULL AND subscription_expires_at > ?", now).
		Count(&activeSubscribers)

	// Recap periods
	last24h := now.Add(-24 * time.Hour)
	last7d := now.Add(-7 * 24 * time.Hour)
	last30d := now.Add(-30 * 24 * time.Hour)

	getRecap := func(since time.Time) fiber.Map {
		var users int64
		var rooms int64
		var transactions int64
		var revenue int64
		var credits int64

		initializers.Db.Model(&models.User{}).Where("created_at >= ?", since).Count(&users)
		initializers.Db.Model(&models.Room{}).Where("created_at >= ?", since).Count(&rooms)
		initializers.Db.Model(&models.Transaction{}).Where("created_at >= ?", since).Count(&transactions)
		initializers.Db.Model(&models.Transaction{}).
			Where("status = ? AND created_at >= ?", models.TransactionStatusSettlement, since).
			Select("COALESCE(SUM(amount),0)").Scan(&revenue)
		initializers.Db.Model(&models.CreditLog{}).Where("amount > 0 AND created_at >= ?", since).
			Select("COALESCE(SUM(amount),0)").Scan(&credits)

		return fiber.Map{
			"users":        users,
			"rooms":        rooms,
			"transactions": transactions,
			"revenue":      revenue,
			"credits":      credits,
		}
	}

	// Time series (last 7 days)
	labels := make([]string, 0, 7)
	usersSeries := make([]int64, 0, 7)
	roomsSeries := make([]int64, 0, 7)
	transactionsSeries := make([]int64, 0, 7)
	revenueSeries := make([]int64, 0, 7)

	startOfDay := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}

	for i := 6; i >= 0; i-- {
		dayStart := startOfDay(now.AddDate(0, 0, -i))
		dayEnd := dayStart.Add(24 * time.Hour)

		var users int64
		var rooms int64
		var transactions int64
		var revenue int64

		initializers.Db.Model(&models.User{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&users)
		initializers.Db.Model(&models.Room{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&rooms)
		initializers.Db.Model(&models.Transaction{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&transactions)
		initializers.Db.Model(&models.Transaction{}).
			Where("status = ? AND created_at >= ? AND created_at < ?", models.TransactionStatusSettlement, dayStart, dayEnd).
			Select("COALESCE(SUM(amount),0)").Scan(&revenue)

		labels = append(labels, dayStart.Format("Jan 2"))
		usersSeries = append(usersSeries, users)
		roomsSeries = append(roomsSeries, rooms)
		transactionsSeries = append(transactionsSeries, transactions)
		revenueSeries = append(revenueSeries, revenue)
	}

	return ctx.JSON(fiber.Map{
		"totalUsers":             totalUsers,
		"totalRooms":             totalRooms,
		"activeRooms":            activeRooms,
		"expiredRooms":           expiredRooms,
		"totalTransactions":      totalTransactions,
		"pendingTransactions":    pendingTransactions,
		"settledTransactions":    settledTransactions,
		"failedTransactions":     failedTransactions,
		"totalPackages":          totalPackages,
		"totalRevenue":           totalRevenue,
		"totalCreditsAwarded":    totalCreditsAwarded,
		"totalSubscriptionPlans": totalSubscriptionPlans,
		"activeSubscribers":      activeSubscribers,
		"recaps": fiber.Map{
			"last24h": getRecap(last24h),
			"last7d":  getRecap(last7d),
			"last30d": getRecap(last30d),
		},
		"series": fiber.Map{
			"labels":       labels,
			"users":        usersSeries,
			"rooms":        roomsSeries,
			"transactions": transactionsSeries,
			"revenue":      revenueSeries,
		},
		"system": fiber.Map{
			"roomMaxDuration":  GetRoomMaxDuration(),
			"roomCreationCost": GetRoomCreationCost(),
			"defaultCredits": func() int {
				value, err := strconv.Atoi(GetConfigValue(models.ConfigDefaultCredits, "5"))
				if err != nil {
					return 5
				}
				return value
			}(),
		},
		"serverTime": now.Format(time.RFC3339),
	})
}

// ========================================
// HELPERS
// ========================================

// GetConfigValue retrieves a system config value with a default fallback
func GetConfigValue(key string, defaultValue string) string {
	var config models.SystemConfig
	if err := initializers.Db.Where("key = ?", key).First(&config).Error; err != nil {
		return defaultValue
	}
	return config.Value
}

// GetRoomMaxDuration returns the room max duration in minutes
func GetRoomMaxDuration() int {
	value := GetConfigValue(models.ConfigRoomMaxDuration, "120")
	duration, err := strconv.Atoi(value)
	if err != nil {
		return 120
	}
	return duration
}

// GetRoomCreationCost returns the credit cost to create a room
func GetRoomCreationCost() int {
	value := GetConfigValue(models.ConfigRoomCreationCost, "1")
	cost, err := strconv.Atoi(value)
	if err != nil {
		return 1
	}
	return cost
}
