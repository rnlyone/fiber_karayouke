package initializers

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseConnection struct {
}

var (
	Db      *gorm.DB
	OauthDb *gorm.DB
	err     error
	once    sync.Once
)

func loadEnv() {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			if os.IsNotExist(err) {
				log.Println(".env file not found; relying on environment variables")
				return
			}
			log.Printf("Warning: could not load .env file: %v", err)
		}
	})
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		if !isProduction() && (key == "DB_PASSWORD" || key == "OAUTH_DB_PASSWORD") {
			return value
		}
		log.Fatalf("environment variable %s is required", key)
	}
	return value
}

func optionalEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func isProduction() bool {
	flag := strings.ToLower(strings.TrimSpace(os.Getenv("IS_PRODUCTION")))
	if flag == "false" || flag == "0" || flag == "no" {
		return false
	}
	return true
}

func DbConnection() error {
	//once.Do(func() {
	loadEnv()
	// Read the environment variables
	dbHost := mustGetEnv("DB_HOST")
	dbPort := mustGetEnv("DB_PORT")
	dbUser := mustGetEnv("DB_USER")
	dbPassword := mustGetEnv("DB_PASSWORD")
	dbName := mustGetEnv("DB_NAME")
	sslMode := optionalEnv("DB_SSLMODE", "disable")
	timezone := optionalEnv("DB_TIMEZONE", "UTC")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", dbHost, dbUser, dbPassword, dbName, dbPort, sslMode, timezone)
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	sqlDB, err := Db.DB()
	if err != nil {
		log.Fatalf("failed to access the underlying database: %v", err)
	}

	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Minute * 10)
	sqlDB.SetConnMaxIdleTime(0)
	//})

	return nil
}

func OauthDatabaseConnection() error {
	//once.Do(func() {
	loadEnv()
	dbHost := mustGetEnv("OAUTH_DB_HOST")
	dbPort := mustGetEnv("OAUTH_DB_PORT")
	dbUser := mustGetEnv("OAUTH_DB_USER")
	dbPassword := mustGetEnv("OAUTH_DB_PASSWORD")
	dbName := mustGetEnv("OAUTH_DB_NAME")
	sslMode := optionalEnv("OAUTH_DB_SSLMODE", optionalEnv("DB_SSLMODE", "disable"))
	timezone := optionalEnv("OAUTH_DB_TIMEZONE", optionalEnv("DB_TIMEZONE", "UTC"))

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", dbHost, dbUser, dbPassword, dbName, dbPort, sslMode, timezone)
	OauthDb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	sqlDB, err := OauthDb.DB()
	if err != nil {
		log.Fatalf("failed to access the underlying database: %v", err)
	}

	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Minute * 10)
	sqlDB.SetConnMaxIdleTime(0)
	//})
	return nil
}
