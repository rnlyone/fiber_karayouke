package artisan

import (
	"fmt"
	"os"
	"path/filepath"
)

func runMakeModel(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("model name is required, e.g. artisan make model User")
	}

	structName := toPascalCase(args[0])
	if structName == "" {
		return fmt.Errorf("invalid model name: %s", args[0])
	}

	snake := toSnakeCase(structName)
	fileName := fmt.Sprintf("%s_model.go", snake)
	filePath := filepath.Join("app", "models", fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("model file already exists: %s", filePath)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to ensure models directory: %w", err)
	}

	tableName := fmt.Sprintf("tbl_%s", snake)
	template := fmt.Sprintf(`package models

import "gorm.io/gorm"

type %s struct {
	gorm.Model
}

func (%s) TableName() string {
	return "%s"
}
`, structName, structName, tableName)

	if err := os.WriteFile(filePath, []byte(template), 0o644); err != nil {
		return fmt.Errorf("failed to write model file: %w", err)
	}

	fmt.Printf("Model created: %s\n", filePath)
	return nil
}
