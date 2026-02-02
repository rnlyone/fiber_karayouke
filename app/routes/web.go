package routes

import (
	"GoFiberMVC/app/controllers"
	ws "GoFiberMVC/app/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

// RegisterWebRoutes registers web routes
func RegisterWebRoutes(app *fiber.App) {
	// Enable CORS for frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://127.0.0.1:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	userController := &controllers.UserController{}
	authController := &controllers.AuthController{}
	roomController := &controllers.RoomController{}

	app.Get("", userController.Index)

	// Auth routes
	app.Post("/api/auth/register", authController.Register)
	app.Post("/api/auth/login", authController.Login)
	app.Get("/api/auth/me", authController.Me)
	app.Post("/api/auth/logout", authController.Logout)

	// Room routes
	app.Post("/api/rooms", roomController.Create)
	app.Get("/api/rooms", roomController.List)
	app.Get("/api/rooms/:roomKey", roomController.Get)
	app.Get("/api/rooms/:roomKey/access", roomController.CheckAccess)

	// WebSocket routes for karaoke rooms
	// Support both /ws/:roomKey and /parties/main/:roomKey for compatibility
	app.Use("/ws", ws.WebSocketUpgrade)
	app.Get("/ws/:roomKey", websocket.New(ws.HandleWebSocket))

	// PartyKit-compatible route
	app.Use("/parties/main", ws.WebSocketUpgrade)
	app.Get("/parties/main/:roomKey", websocket.New(ws.HandleWebSocket))

	// HTTP API for room state
	app.Get("/api/room/:roomKey", ws.GetRoomState)
}
