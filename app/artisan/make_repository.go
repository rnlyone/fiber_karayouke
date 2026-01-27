package artisan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runMakeRepository(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("repository name is required, e.g. artisan make repository Client")
	}

	structName := toPascalCase(args[0])
	if structName == "" {
		return fmt.Errorf("invalid repository name: %s", args[0])
	}

	if !strings.HasSuffix(strings.ToLower(structName), "repository") {
		structName += "Repository"
	}

	snake := toSnakeCase(structName)
	fileName := fmt.Sprintf("%s.go", snake)
	filePath := filepath.Join("app", "repositories", fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("repository already exists: %s", filePath)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to ensure repositories directory: %w", err)
	}

	template := fmt.Sprintf(`package repositories

import "gorm.io/gorm"

// %s handles data persistence.
type %s struct {
    db *gorm.DB
}

func New%s(db *gorm.DB) *%s {
    return &%s{db: db}
}
`, structName, structName, structName, structName, structName)

	if err := os.WriteFile(filePath, []byte(template), 0o644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	fmt.Printf("Repository created: %s\n", filePath)
	return nil
}
