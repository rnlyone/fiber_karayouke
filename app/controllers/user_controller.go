package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// UserController handles user-related requests
type UserController struct {
	// Controller dependencies or services can be injected here
}

// Index handles the GET request for listing users
func (c *UserController) Index(ctx *fiber.Ctx) error {
	// Render index template
	return ctx.Render("index", fiber.Map{
		"Title": "Karayouke",
	})
}

// Show handles the GET request for retrieving a single user
func (c *UserController) Show(ctx *fiber.Ctx) error {
	// Logic to fetch and return a single user
	userID := ctx.Params("id")
	return ctx.SendString("User ID: " + userID)
}

// Store handles the POST request for creating a new user
func (c *UserController) Store(ctx *fiber.Ctx) error {
	// Logic to create a new user based on the request data
	return ctx.SendString("User created")
}

// Update handles the PUT request for updating an existing user
func (c *UserController) Update(ctx *fiber.Ctx) error {
	// Logic to update an existing user based on the request data
	userID := ctx.Params("id")
	return ctx.SendString("User updated: " + userID)
}

// Delete handles the DELETE request for deleting a user
func (c *UserController) Delete(ctx *fiber.Ctx) error {
	// Logic to delete a user
	userID := ctx.Params("id")
	return ctx.SendString("User deleted: " + userID)
}
