package initializers

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
	OauthDb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
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
