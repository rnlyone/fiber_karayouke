package main

import (
	"log"
	"os"
	"strings"

	"GoFiberMVC/app/artisan"
	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"
	"GoFiberMVC/app/providers"
	"GoFiberMVC/app/routes"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "artisan" {
		if err := artisan.Run(os.Args[2:]); err != nil {
			log.Fatalf("artisan error: %v", err)
		}
		return
	}

	app := providers.AppProvider()

	// Database connection for auth and rooms
	if err := initializers.DbConnection(); err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Running in WebSocket-only mode without persistence")
	} else {
		// Auto-migrate to add any new columns (e.g., Room.MaxDuration)
		initializers.Db.AutoMigrate(&models.Room{})
	}

	routes.RegisterWebRoutes(app)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3000"
	}
	addr := ":" + port
	log.Printf("Karayouke server starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
