package websocket

import (
	"log"
	"strings"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var roomManager = NewRoomManager()

// WebSocketUpgrade middleware to check if request is a WebSocket upgrade
func WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// HandleWebSocket handles WebSocket connections for karaoke rooms
func HandleWebSocket(c *websocket.Conn) {
	// Get room key from path parameter
	roomKey := c.Params("roomKey")
	if roomKey == "" {
		// Try to extract from query or path
		path := c.Locals("path")
		if pathStr, ok := path.(string); ok {
			parts := strings.Split(pathStr, "/")
			if len(parts) > 0 {
				roomKey = parts[len(parts)-1]
			}
		}
	}

	if roomKey == "" {
		log.Println("WebSocket: no room key provided")
		return
	}

	// Validate room exists in database
	var dbRoom models.Room
	if err := initializers.Db.Where("room_key = ?", roomKey).First(&dbRoom).Error; err != nil {
		log.Printf("WebSocket: room %s not found in database", roomKey)
		c.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","error":"Room not found"}`))
		return
	}

	log.Printf("WebSocket: connection to room %s (%s)", roomKey, dbRoom.RoomName)

	// Get or create room
	room := roomManager.GetOrCreateRoom(roomKey)

	// Create connection wrapper
	conn := NewConnection(c, room)

	// Add connection to room
	room.AddConnection(conn)

	// Send initial state
	room.SendState(conn)

	// Handle incoming messages
	defer func() {
		conn.Close()
		room.RemoveConnection(conn)
		log.Printf("WebSocket: disconnected from room %s", roomKey)
	}()

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			room.HandleMessage(conn, message)
		}
	}
}

// GetRoomState returns the current state of a room via HTTP
func GetRoomState(c *fiber.Ctx) error {
	roomKey := c.Params("roomKey")
	if roomKey == "" {
		return c.Status(400).JSON(fiber.Map{"error": "room key required"})
	}

	// Validate room exists in database
	var dbRoom models.Room
	if err := initializers.Db.Where("room_key = ?", roomKey).First(&dbRoom).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Room not found"})
	}

	room := roomManager.GetOrCreateRoom(roomKey)
	room.mu.RLock()
	defer room.mu.RUnlock()

	return c.JSON(room.State)
}
