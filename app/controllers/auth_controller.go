package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct{}

type RegisterRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Credit   int    `json:"credit"`
}

// Session duration: 30 days
const sessionDuration = 30 * 24 * time.Hour

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// createSession creates a new session in the database
func createSession(userID string) (string, error) {
	token := generateToken()
	session := models.Session{
		ID:        generateID(),
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := initializers.Db.Create(&session).Error; err != nil {
		return "", err
	}
	return token, nil
}

// getSessionUser retrieves user from session token
func getSessionUser(token string) *models.User {
	var session models.Session
	if err := initializers.Db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error; err != nil {
		return nil
	}

	var user models.User
	if err := initializers.Db.Where("id = ?", session.UserID).First(&user).Error; err != nil {
		return nil
	}

	// Extend session expiry on activity (rolling session)
	initializers.Db.Model(&session).Update("expires_at", time.Now().Add(sessionDuration))

	return &user
}

// deleteSession removes a session from the database
func deleteSession(token string) {
	initializers.Db.Where("token = ?", token).Delete(&models.Session{})
}

// cleanupExpiredSessions removes expired sessions (can be called periodically)
func cleanupExpiredSessions() {
	initializers.Db.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
}

func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var req RegisterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" || req.Username == "" || req.Email == "" || req.Password == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "All fields are required"})
	}

	// Check if user exists
	var existingUser models.User
	if err := initializers.Db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		return ctx.Status(409).JSON(fiber.Map{"error": "User with this email or username already exists"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	user := models.User{
		ID:       generateID(),
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Credit:   0,
	}

	if err := initializers.Db.Create(&user).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	token, err := createSession(user.ID)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create session"})
	}

	return ctx.JSON(AuthResponse{
		Token: token,
		User: UserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Email:    user.Email,
			Credit:   user.Credit,
		},
	})
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Email == "" || req.Password == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
	}

	var user models.User
	if err := initializers.Db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := createSession(user.ID)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create session"})
	}

	return ctx.JSON(AuthResponse{
		Token: token,
		User: UserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Email:    user.Email,
			Credit:   user.Credit,
		},
	})
}

func (c *AuthController) Me(ctx *fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	user := getSessionUser(token)
	if user == nil {
		return ctx.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	return ctx.JSON(UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Credit:   user.Credit,
	})
}

func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	deleteSession(token)
	return ctx.JSON(fiber.Map{"message": "Logged out successfully"})
}

// GetUserFromToken extracts user from authorization header
func GetUserFromToken(ctx *fiber.Ctx) *models.User {
	token := ctx.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	return getSessionUser(token)
}
