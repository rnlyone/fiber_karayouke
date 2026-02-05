package websocket

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"GoFiberMVC/app/controllers"
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

	// Check if room is expired
	maxDuration := controllers.GetRoomMaxDuration()
	if dbRoom.IsExpired(maxDuration) {
		log.Printf("WebSocket: room %s is expired", roomKey)
		c.WriteMessage(websocket.TextMessage, []byte(`{"type":"room_expired","message":"This room has expired"}`))
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

	// Start expiration checker goroutine
	stopChecker := make(chan struct{})
	go func() {
		checkExpirationAndKick(roomKey, room, stopChecker)
	}()

	// Handle incoming messages
	defer func() {
		close(stopChecker)
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

// checkExpirationAndKick checks if the room has expired and kicks all users
func checkExpirationAndKick(roomKey string, room *Room, stop chan struct{}) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			// Re-fetch room to check expiration
			var dbRoom models.Room
			if err := initializers.Db.Where("room_key = ?", roomKey).First(&dbRoom).Error; err != nil {
				continue
			}

			maxDuration := controllers.GetRoomMaxDuration()
			if dbRoom.IsExpired(maxDuration) {
				log.Printf("WebSocket: room %s has expired, kicking all users", roomKey)
				// Broadcast expiration to all clients
				msg := map[string]interface{}{
					"type":    "room_expired",
					"message": "This room has expired",
				}
				data, _ := json.Marshal(msg)
				room.Broadcast(data)

				// Close all connections
				room.mu.Lock()
				for conn := range room.Connections {
					conn.Close()
				}
				room.mu.Unlock()
				return
			}
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
