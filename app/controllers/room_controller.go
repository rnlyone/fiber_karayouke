package controllers

import (
	"crypto/rand"
	"encoding/hex"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/gofiber/fiber/v2"
)

type RoomController struct{}

type CreateRoomRequest struct {
	Name string `json:"name"`
}

type RoomResponse struct {
	ID        string  `json:"id"`
	RoomKey   string  `json:"room_key"`
	RoomName  string  `json:"room_name"`
	CreatorID string  `json:"creator_id"`
	MasterID  string  `json:"master_id"`
	CreatedAt string  `json:"created_at"`
	ExpiredAt *string `json:"expired_at"`
	IsExpired bool    `json:"is_expired"`
}

func generateRoomKey() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	bytes := make([]byte, 6)
	rand.Read(bytes)
	for i := range bytes {
		bytes[i] = chars[int(bytes[i])%len(chars)]
	}
	return string(bytes)
}

func generateRoomID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (c *RoomController) Create(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req CreateRoomRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Room name is required"})
	}

	// Reset free credits if needed
	ResetFreeCreditIfNeeded(user)

	// Check credit balance (free + extra)
	creationCost := GetRoomCreationCost()
	if user.TotalCredits() < creationCost {
		return ctx.Status(400).JSON(fiber.Map{
			"error":        "Insufficient credits",
			"required":     creationCost,
			"free_credit":  user.FreeCredit,
			"extra_credit": user.Credit,
			"total_credit": user.TotalCredits(),
		})
	}

	// Generate unique room key
	var roomKey string
	for {
		roomKey = generateRoomKey()
		var existing models.Room
		if err := initializers.Db.Where("room_key = ?", roomKey).First(&existing).Error; err != nil {
			break
		}
	}

	// Calculate room duration based on user's subscription
	maxDuration := GetUserRoomDuration(user)

	room := models.Room{
		ID:          generateRoomID(),
		RoomKey:     roomKey,
		RoomName:    req.Name,
		RoomCreator: user.ID,
		RoomMaster:  user.ID,
		MaxDuration: maxDuration,
	}

	// Deduct credits (free first, then extra)
	if !user.DeductCredits(creationCost) {
		return ctx.Status(400).JSON(fiber.Map{"error": "Insufficient credits"})
	}
	if err := initializers.Db.Save(user).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to deduct credits"})
	}

	// Log credit deduction
	creditLog := models.CreditLog{
		ID:          generateRoomID(),
		UserID:      user.ID,
		Amount:      -creationCost,
		Balance:     user.TotalCredits(),
		Type:        models.CreditTypeRoomCreation,
		ReferenceID: room.ID,
		Description: "Room creation: " + room.RoomName,
	}
	initializers.Db.Create(&creditLog)

	if err := initializers.Db.Create(&room).Error; err != nil {
		// Refund credits on failure (add back to free first if was deducted from extra)
		user.FreeCredit += creationCost
		initializers.Db.Save(user)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create room"})
	}

	expiredAt := room.GetExpiredAt(maxDuration)
	formattedExpiredAt := expiredAt.Format("2006-01-02T15:04:05Z07:00")

	return ctx.JSON(RoomResponse{
		ID:        room.ID,
		RoomKey:   room.RoomKey,
		RoomName:  room.RoomName,
		CreatorID: room.RoomCreator,
		MasterID:  room.RoomMaster,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiredAt: &formattedExpiredAt,
		IsExpired: room.IsExpired(maxDuration),
	})
}

func (c *RoomController) List(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var rooms []models.Room
	if err := initializers.Db.Where("room_creator = ? OR room_master = ?", user.ID, user.ID).Order("created_at DESC").Find(&rooms).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch rooms"})
	}

	defaultDuration := GetRoomMaxDuration() // fallback for old rooms without stored duration
	response := make([]RoomResponse, len(rooms))
	for i, room := range rooms {
		expiredAt := room.GetExpiredAt(defaultDuration)
		formattedExpiredAt := expiredAt.Format("2006-01-02T15:04:05Z07:00")
		response[i] = RoomResponse{
			ID:        room.ID,
			RoomKey:   room.RoomKey,
			RoomName:  room.RoomName,
			CreatorID: room.RoomCreator,
			MasterID:  room.RoomMaster,
			CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiredAt: &formattedExpiredAt,
			IsExpired: room.IsExpired(defaultDuration),
		}
	}

	return ctx.JSON(response)
}

func (c *RoomController) Get(ctx *fiber.Ctx) error {
	roomKey := ctx.Params("roomKey")
	if roomKey == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Room key is required"})
	}

	var room models.Room
	if err := initializers.Db.Where("room_key = ?", roomKey).First(&room).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Room not found"})
	}

	defaultDuration := GetRoomMaxDuration()
	expiredAt := room.GetExpiredAt(defaultDuration)
	formattedExpiredAt := expiredAt.Format("2006-01-02T15:04:05Z07:00")

	return ctx.JSON(RoomResponse{
		ID:        room.ID,
		RoomKey:   room.RoomKey,
		RoomName:  room.RoomName,
		CreatorID: room.RoomCreator,
		MasterID:  room.RoomMaster,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiredAt: &formattedExpiredAt,
		IsExpired: room.IsExpired(defaultDuration),
	})
}

func (c *RoomController) CheckAccess(ctx *fiber.Ctx) error {
	roomKey := ctx.Params("roomKey")
	if roomKey == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Room key is required"})
	}

	var room models.Room
	if err := initializers.Db.Where("room_key = ?", roomKey).First(&room).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Room not found"})
	}

	defaultDuration := GetRoomMaxDuration()
	expiredAt := room.GetExpiredAt(defaultDuration)

	// Check if room is expired
	if room.IsExpired(defaultDuration) {
		return ctx.Status(410).JSON(fiber.Map{
			"error":      "Room has expired",
			"expired_at": expiredAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	user := GetUserFromToken(ctx)
	isMaster := user != nil && (user.ID == room.RoomCreator || user.ID == room.RoomMaster)

	formattedExpiredAt := expiredAt.Format("2006-01-02T15:04:05Z07:00")

	return ctx.JSON(fiber.Map{
		"room_key":   room.RoomKey,
		"room_name":  room.RoomName,
		"is_master":  isMaster,
		"user_name":  getUserName(user),
		"user_id":    getUserID(user),
		"expired_at": formattedExpiredAt,
		"is_expired": room.IsExpired(defaultDuration),
	})
}

func getUserName(user *models.User) string {
	if user == nil {
		return ""
	}
	return user.Name
}

func getUserID(user *models.User) string {
	if user == nil {
		return ""
	}
	return user.ID
}
