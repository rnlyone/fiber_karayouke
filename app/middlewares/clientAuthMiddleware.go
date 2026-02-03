package middlewares

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

type ClientAuthMiddleware struct {
}

const PublicKeyPath = "app/public/assets/oauth-public.key"

func (clientMiddleware *ClientAuthMiddleware) Auth(ctx *fiber.Ctx) error {
	//Import Middleware

	// Get the Authorization header value
	authorization := ctx.Get("Authorization")

	// Check if the Authorization header is present and starts with "Bearer "
	if authorization == "" || !strings.HasPrefix(authorization, "Bearer ") {
		// Return a 401 Unauthorized response if the token is missing or invalid
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Extract the token from the Authorization header
	tokenString := strings.TrimPrefix(authorization, "Bearer ")
	// Read the public key file
	publicKeyBytes, err := ioutil.ReadFile(PublicKeyPath)
	if err != nil {
		// Return a 500 Internal Server Error response if unable to read the public key
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Internal Server Error",
		})
	}

	// Parse the public key
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		// Return a 500 Internal Server Error response if unable to parse the public key
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Internal Server Error",
		})
	}

	// Verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the token signing method and return the public key
		if token.Method.Alg() == jwt.SigningMethodRS256.Name {
			return publicKey, nil
		}
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	})
	if err != nil || !token.Valid {
		// Return a 401 Unauthorized response if the token is invalid
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Extract the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// Return a 500 Internal Server Error response if unable to extract the claims
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Internal Server Error",
		})
	}

	subJti := claims["jti"].(string)
	// Perform your authentication logic here
	// You can validate the token, decode claims, check permissions, etc.
	user := &OauthAccessTokens{}
	user.Id = subJti
	if user.Validate() == false {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Invalid Token",
		})
	}

	// Example: Check if the token is valid and belongs to an authenticated user
	// Add the oauth object to the context
	// NOTE: models.Client is not defined in this codebase - this middleware is unused legacy code
	// TODO: Remove this middleware or implement models.Client if needed
	ctx.Locals("auth", user)
	// Print all properties of the client object
	// Continue to the next middleware or route handler
	return ctx.Next()
}
