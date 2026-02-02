package routes

import (
	"GoFiberMVC/app/controllers"
	ws "GoFiberMVC/app/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// RegisterWebRoutes registers web routes
func RegisterWebRoutes(app *fiber.App) {
	userController := &controllers.UserController{}

	app.Get("", userController.Index)

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
