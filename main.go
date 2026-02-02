package main

import (
	"log"
	"os"

	"GoFiberMVC/app/artisan"
	"GoFiberMVC/app/initializers"
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
	}

	routes.RegisterWebRoutes(app)

	log.Println("Karayouke server starting on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
