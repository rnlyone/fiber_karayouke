package main

import (
	"log"
	"os"

	"GoFiberMVC/app/artisan"
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
	routes.RegisterWebRoutes(app)

	// Database connections are optional for WebSocket-only mode
	// Uncomment if you need database functionality:
	// initializers.DbConnection()
	// initializers.OauthDatabaseConnection()

	log.Println("Karayouke WebSocket server starting on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
