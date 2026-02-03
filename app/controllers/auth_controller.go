package controllers

import (
	"crypto/rand"
	"encoding/hex"

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

// Session storage (in production, use Redis or database)
var sessions = make(map[string]*models.User)

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

	token := generateToken()
	sessions[token] = &user

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

	token := generateToken()
	sessions[token] = &user

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

	user, exists := sessions[token]
	if !exists {
		return ctx.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	// Refresh user from database to ensure latest credit balance
	var freshUser models.User
	if err := initializers.Db.Where("id = ?", user.ID).First(&freshUser).Error; err == nil {
		user = &freshUser
		sessions[token] = &freshUser
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
	delete(sessions, token)
	return ctx.JSON(fiber.Map{"message": "Logged out successfully"})
}

// GetUserFromToken extracts user from authorization header
func GetUserFromToken(ctx *fiber.Ctx) *models.User {
	token := ctx.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	if user, exists := sessions[token]; exists {
		// Refresh user from database to ensure latest credit balance
		var freshUser models.User
		if err := initializers.Db.Where("id = ?", user.ID).First(&freshUser).Error; err == nil {
			sessions[token] = &freshUser
			return &freshUser
		}
		return user
	}
	return nil
}
