package artisan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runMakeController(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("controller name is required, e.g. artisan make controller User")
	}

	structName := toPascalCase(args[0])
	if structName == "" {
		return fmt.Errorf("invalid controller name: %s", args[0])
	}

	if !strings.HasSuffix(strings.ToLower(structName), "controller") {
		structName += "Controller"
	}

	snake := toSnakeCase(structName)
	fileName := fmt.Sprintf("%s.go", snake)
	filePath := filepath.Join("app", "controllers", fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("controller already exists: %s", filePath)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to ensure controllers directory: %w", err)
	}

	template := fmt.Sprintf(`package controllers

import "github.com/gofiber/fiber/v2"

// %s provides HTTP handlers.
type %s struct{}

func (ctl *%s) Index(ctx *fiber.Ctx) error {
    return ctx.JSON(fiber.Map{"message": "%s index"})
}
`, structName, structName, structName, structName)

	if err := os.WriteFile(filePath, []byte(template), 0o644); err != nil {
		return fmt.Errorf("failed to write controller file: %w", err)
	}

	fmt.Printf("Controller created: %s\n", filePath)
	return nil
}
