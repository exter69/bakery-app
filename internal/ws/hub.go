package ws

import (
	"encoding/json"
	"log"
	"sync"
)

// Event is a message sent to connected WebSocket clients.
type Event struct {
	Type    string      `json:"type"`    // e.g. "order_status", "new_order", "reservation_status"
	Payload interface{} `json:"payload"` // event-specific data
}

// Hub manages WebSocket connections grouped by user ID.
// A single user may have multiple connections (multiple tabs/devices).
type Hub struct {
	mu    sync.RWMutex
	conns map[string][]*Conn // userID → active connections
}

// NewHub creates a new WebSocket connection hub.
func NewHub() *Hub {
	return &Hub{
		conns: make(map[string][]*Conn),
	}
}

// Register adds a connection for the given user.
func (h *Hub) Register(userID string, conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = append(h.conns[userID], conn)
	log.Printf("[WS] user %s connected (total conns: %d)", userID, len(h.conns[userID]))
}

// Unregister removes a specific connection for a user.
func (h *Hub) Unregister(userID string, conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.conns[userID]
	for i, c := range conns {
		if c == conn {
			h.conns[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	// Clean up empty slices
	if len(h.conns[userID]) == 0 {
		delete(h.conns, userID)
	}
	log.Printf("[WS] user %s disconnected", userID)
}

// SendToUser sends an event to all connections for a specific user.
func (h *Hub) SendToUser(userID string, event Event) {
	h.mu.RLock()
	conns := h.conns[userID]
	h.mu.RUnlock()

	if len(conns) == 0 {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WS] failed to marshal event for user %s: %v", userID, err)
		return
	}

	for _, conn := range conns {
		conn.Send(data)
	}
}

// Broadcast sends an event to all connected users.
func (h *Hub) Broadcast(event Event) {
	h.mu.RLock()
	allConns := make([]*Conn, 0)
	for _, conns := range h.conns {
		allConns = append(allConns, conns...)
	}
	h.mu.RUnlock()

	if len(allConns) == 0 {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WS] failed to marshal broadcast event: %v", err)
		return
	}

	for _, conn := range allConns {
		conn.Send(data)
	}
}

// ConnCount returns the total number of active connections (useful for monitoring).
func (h *Hub) ConnCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, conns := range h.conns {
		count += len(conns)
	}
	return count
}

// UserConnCount returns the number of active connections for a specific user.
func (h *Hub) UserConnCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns[userID])
}
