package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"time"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/gofiber/fiber/v2"
)

type TVController struct{}

// Token expires after 2 minutes (gives buffer beyond the 1-minute refresh)
const tvTokenDuration = 2 * time.Minute

// Characters for short code (uppercase letters and numbers, excluding confusing ones like O/0, I/1/L)
const shortCodeChars = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// generateShortCode generates a 5-character unique code
func generateShortCode() string {
	code := make([]byte, 5)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(shortCodeChars))))
		code[i] = shortCodeChars[n.Int64()]
	}
	return string(code)
}

// generateTVToken generates a random token for QR code
func generateTVToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// cleanupExpiredTVTokens removes expired TV tokens
func cleanupExpiredTVTokens() {
	initializers.Db.Where("expires_at < ?", time.Now()).Delete(&models.TVToken{})
}

// GenerateToken creates a new TV token with QR token and short code
// This endpoint does NOT require authentication - TV is a display-only device
func (c *TVController) GenerateToken(ctx *fiber.Ctx) error {
	// Clean up expired tokens first
	cleanupExpiredTVTokens()

	// Generate unique short code (retry if collision)
	var shortCode string
	for i := 0; i < 10; i++ {
		shortCode = generateShortCode()
		var existing models.TVToken
		if err := initializers.Db.Where("short_code = ? AND expires_at > ?", shortCode, time.Now()).First(&existing).Error; err != nil {
			// No collision, use this code
			break
		}
		// Collision, try again
		if i == 9 {
			return ctx.Status(500).JSON(fiber.Map{"error": "Failed to generate unique code"})
		}
	}

	token := generateTVToken()
	tvToken := models.TVToken{
		ID:        generateID(),
		Token:     token,
		ShortCode: shortCode,
		RoomKey:   "", // Not connected yet
		ExpiresAt: time.Now().Add(tvTokenDuration),
	}

	if err := initializers.Db.Create(&tvToken).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create TV token"})
	}

	return ctx.JSON(fiber.Map{
		"token":      token,
		"short_code": shortCode,
		"expires_at": tvToken.ExpiresAt,
	})
}

// GetStatus checks if a TV token is connected to a room
// This endpoint does NOT require authentication - TV polls this
func (c *TVController) GetStatus(ctx *fiber.Ctx) error {
	token := ctx.Params("token")
	if token == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Token is required"})
	}

	var tvToken models.TVToken
	if err := initializers.Db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&tvToken).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"error":   "Token not found or expired",
			"expired": true,
		})
	}

	if tvToken.RoomKey == "" {
		return ctx.JSON(fiber.Map{
			"connected": false,
			"room_key":  nil,
		})
	}

	// Verify room still exists
	var room models.Room
	if err := initializers.Db.Where("room_key = ?", tvToken.RoomKey).First(&room).Error; err != nil {
		// Room was deleted, disconnect TV
		tvToken.RoomKey = ""
		initializers.Db.Save(&tvToken)
		return ctx.JSON(fiber.Map{
			"connected": false,
			"room_key":  nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"connected": true,
		"room_key":  tvToken.RoomKey,
		"room_name": room.RoomName,
	})
}

// Connect links a TV token to a room
// This endpoint REQUIRES authentication - only room master can connect TV
func (c *TVController) Connect(ctx *fiber.Ctx) error {
	user := GetUserFromToken(ctx)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req struct {
		Code    string `json:"code"` // Either QR token or short code
		RoomKey string `json:"room_key"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Code == "" || req.RoomKey == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Code and room_key are required"})
	}

	// Verify user has access to this room (is room creator or master)
	var room models.Room
	if err := initializers.Db.Where("room_key = ?", req.RoomKey).First(&room).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "Room not found"})
	}

	if room.RoomCreator != user.ID && room.RoomMaster != user.ID {
		return ctx.Status(403).JSON(fiber.Map{"error": "You don't have permission to connect TV to this room"})
	}

	// Check if room is expired
	if room.IsExpired() {
		return ctx.Status(400).JSON(fiber.Map{"error": "Room has expired"})
	}

	// Find TV token by either full token or short code
	var tvToken models.TVToken
	code := strings.TrimSpace(req.Code)

	// Try as short code first (5 chars, uppercase)
	if len(code) == 5 {
		code = strings.ToUpper(code)
		if err := initializers.Db.Where("short_code = ? AND expires_at > ?", code, time.Now()).First(&tvToken).Error; err != nil {
			// Try as token
			if err := initializers.Db.Where("token = ? AND expires_at > ?", req.Code, time.Now()).First(&tvToken).Error; err != nil {
				return ctx.Status(404).JSON(fiber.Map{"error": "Invalid or expired code"})
			}
		}
	} else {
		// Try as full token
		if err := initializers.Db.Where("token = ? AND expires_at > ?", code, time.Now()).First(&tvToken).Error; err != nil {
			return ctx.Status(404).JSON(fiber.Map{"error": "Invalid or expired code"})
		}
	}

	// Check if already connected to a different room
	if tvToken.RoomKey != "" && tvToken.RoomKey != req.RoomKey {
		// Disconnect from old room
		tvToken.RoomKey = ""
	}

	// Connect TV to room
	tvToken.RoomKey = req.RoomKey
	// Extend expiry when connected (TV stays connected for longer)
	tvToken.ExpiresAt = time.Now().Add(24 * time.Hour)

	if err := initializers.Db.Save(&tvToken).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to connect TV"})
	}

	return ctx.JSON(fiber.Map{
		"success":   true,
		"room_key":  req.RoomKey,
		"room_name": room.RoomName,
	})
}

// Disconnect removes a TV token from a room
func (c *TVController) Disconnect(ctx *fiber.Ctx) error {
	token := ctx.Params("token")
	if token == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Token is required"})
	}

	var tvToken models.TVToken
	// Try as token first, then as short code
	if err := initializers.Db.Where("token = ?", token).First(&tvToken).Error; err != nil {
		if err := initializers.Db.Where("short_code = ?", strings.ToUpper(token)).First(&tvToken).Error; err != nil {
			return ctx.Status(404).JSON(fiber.Map{"error": "Token not found"})
		}
	}

	// Clear room connection
	tvToken.RoomKey = ""
	initializers.Db.Save(&tvToken)

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
