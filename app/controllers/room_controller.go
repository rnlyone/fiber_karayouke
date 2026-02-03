package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"time"

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

	// Check credit balance
	creationCost := GetRoomCreationCost()
	if user.Credit < creationCost {
		return ctx.Status(400).JSON(fiber.Map{
			"error":          "Insufficient credits",
			"required":       creationCost,
			"current_credit": user.Credit,
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

	// Calculate expiration time
	maxDuration := GetRoomMaxDuration()
	expiredAt := time.Now().Add(time.Duration(maxDuration) * time.Minute)

	room := models.Room{
		ID:          generateRoomID(),
		RoomKey:     roomKey,
		RoomName:    req.Name,
		RoomCreator: user.ID,
		RoomMaster:  user.ID,
		ExpiredAt:   &expiredAt,
	}

	// Deduct credits
	user.Credit -= creationCost
	if err := initializers.Db.Save(user).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to deduct credits"})
	}

	// Log credit deduction
	creditLog := models.CreditLog{
		ID:          generateRoomID(), // reuse ID generator
		UserID:      user.ID,
		Amount:      -creationCost,
		Balance:     user.Credit,
		Type:        models.CreditTypeRoomCreation,
		ReferenceID: room.ID,
		Description: "Room creation: " + room.RoomName,
	}
	initializers.Db.Create(&creditLog)

	if err := initializers.Db.Create(&room).Error; err != nil {
		// Refund credits on failure
		user.Credit += creationCost
		initializers.Db.Save(user)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create room"})
	}

	var expiredAtStr *string
	if room.ExpiredAt != nil {
		formatted := room.ExpiredAt.Format("2006-01-02T15:04:05Z07:00")
		expiredAtStr = &formatted
	}

	return ctx.JSON(RoomResponse{
		ID:        room.ID,
		RoomKey:   room.RoomKey,
		RoomName:  room.RoomName,
		CreatorID: room.RoomCreator,
		MasterID:  room.RoomMaster,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiredAt: expiredAtStr,
		IsExpired: room.IsExpired(),
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

	response := make([]RoomResponse, len(rooms))
	for i, room := range rooms {
		var expiredAtStr *string
		if room.ExpiredAt != nil {
			formatted := room.ExpiredAt.Format("2006-01-02T15:04:05Z07:00")
			expiredAtStr = &formatted
		}
		response[i] = RoomResponse{
			ID:        room.ID,
			RoomKey:   room.RoomKey,
			RoomName:  room.RoomName,
			CreatorID: room.RoomCreator,
			MasterID:  room.RoomMaster,
			CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiredAt: expiredAtStr,
			IsExpired: room.IsExpired(),
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

	var expiredAtStr *string
	if room.ExpiredAt != nil {
		formatted := room.ExpiredAt.Format("2006-01-02T15:04:05Z07:00")
		expiredAtStr = &formatted
	}

	return ctx.JSON(RoomResponse{
		ID:        room.ID,
		RoomKey:   room.RoomKey,
		RoomName:  room.RoomName,
		CreatorID: room.RoomCreator,
		MasterID:  room.RoomMaster,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiredAt: expiredAtStr,
		IsExpired: room.IsExpired(),
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

	// Check if room is expired
	if room.IsExpired() {
		return ctx.Status(410).JSON(fiber.Map{
			"error":      "Room has expired",
			"expired_at": room.ExpiredAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	user := GetUserFromToken(ctx)
	isMaster := user != nil && (user.ID == room.RoomCreator || user.ID == room.RoomMaster)

	var expiredAtStr *string
	if room.ExpiredAt != nil {
		formatted := room.ExpiredAt.Format("2006-01-02T15:04:05Z07:00")
		expiredAtStr = &formatted
	}

	return ctx.JSON(fiber.Map{
		"room_key":   room.RoomKey,
		"room_name":  room.RoomName,
		"is_master":  isMaster,
		"user_name":  getUserName(user),
		"user_id":    getUserID(user),
		"expired_at": expiredAtStr,
		"is_expired": room.IsExpired(),
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
