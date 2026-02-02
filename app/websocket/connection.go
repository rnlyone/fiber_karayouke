package websocket

import (
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Connection wraps a WebSocket connection
type Connection struct {
	Conn   *websocket.Conn
	Room   *Room
	mu     sync.Mutex
	closed bool
}

// NewConnection creates a new connection wrapper
func NewConnection(conn *websocket.Conn, room *Room) *Connection {
	return &Connection{
		Conn:   conn,
		Room:   room,
		closed: false,
	}
}

// Send sends a message to the connection
func (c *Connection) Send(message []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
		c.closed = true
	}
}

// Close marks the connection as closed
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}
