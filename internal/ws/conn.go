package ws

import (
	"context"
	"log"
	"time"

	"nhooyr.io/websocket"
)

const (
	// writeTimeout is the max time to wait for a write operation.
	writeTimeout = 10 * time.Second

	// sendBufferSize is the channel buffer for outgoing messages.
	sendBufferSize = 64
)

// Conn wraps a single WebSocket connection and manages its write lifecycle.
type Conn struct {
	ws     *websocket.Conn
	userID string
	send   chan []byte
	done   chan struct{}
}

// NewConn creates a new connection wrapper.
func NewConn(ws *websocket.Conn, userID string) *Conn {
	return &Conn{
		ws:     ws,
		userID: userID,
		send:   make(chan []byte, sendBufferSize),
		done:   make(chan struct{}),
	}
}

// UserID returns the user ID associated with this connection.
func (c *Conn) UserID() string {
	return c.userID
}

// Send queues a message for delivery. Non-blocking: drops the message if the buffer is full.
func (c *Conn) Send(data []byte) {
	select {
	case c.send <- data:
	case <-c.done:
		// Connection already closed
	default:
		// Buffer full — drop the message to avoid blocking the sender
		log.Printf("[WS] dropping message for user %s (buffer full)", c.userID)
	}
}

// WritePump reads from the send channel and writes to the WebSocket.
// It should be run as a goroutine. It exits when the done channel is closed
// or the send channel is drained after close.
func (c *Conn) WritePump() {
	defer c.ws.Close(websocket.StatusNormalClosure, "closing")

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// Channel closed
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
			err := c.ws.Write(ctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				log.Printf("[WS] write error for user %s: %v", c.userID, err)
				return
			}
		case <-c.done:
			return
		}
	}
}

// Close signals the connection to shut down.
func (c *Conn) Close() {
	select {
	case <-c.done:
		// Already closed
	default:
		close(c.done)
	}
}
