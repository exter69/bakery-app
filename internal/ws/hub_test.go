package ws

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConn creates a Conn with a nil websocket for testing hub operations.
func mockConn(userID string) *Conn {
	return &Conn{
		ws:     nil,
		userID: userID,
		send:   make(chan []byte, sendBufferSize),
		done:   make(chan struct{}),
	}
}

func TestHub_RegisterAddsConnection(t *testing.T) {
	hub := NewHub()
	conn := mockConn("user-1")

	hub.Register("user-1", conn)

	assert.Equal(t, 1, hub.ConnCount())
	assert.Equal(t, 1, hub.UserConnCount("user-1"))
}

func TestHub_RegisterMultipleConnectionsForSameUser(t *testing.T) {
	hub := NewHub()
	conn1 := mockConn("user-1")
	conn2 := mockConn("user-1")

	hub.Register("user-1", conn1)
	hub.Register("user-1", conn2)

	assert.Equal(t, 2, hub.ConnCount())
	assert.Equal(t, 2, hub.UserConnCount("user-1"))
}

func TestHub_UnregisterRemovesSpecificConnection(t *testing.T) {
	hub := NewHub()
	conn1 := mockConn("user-1")
	conn2 := mockConn("user-1")

	hub.Register("user-1", conn1)
	hub.Register("user-1", conn2)
	hub.Unregister("user-1", conn1)

	assert.Equal(t, 1, hub.ConnCount())
	assert.Equal(t, 1, hub.UserConnCount("user-1"))
}

func TestHub_UnregisterLastConnectionCleansUpMap(t *testing.T) {
	hub := NewHub()
	conn := mockConn("user-1")

	hub.Register("user-1", conn)
	hub.Unregister("user-1", conn)

	assert.Equal(t, 0, hub.ConnCount())
	assert.Equal(t, 0, hub.UserConnCount("user-1"))
}

func TestHub_SendToUserDeliversToAllUserConnections(t *testing.T) {
	hub := NewHub()
	conn1 := mockConn("user-1")
	conn2 := mockConn("user-1")

	hub.Register("user-1", conn1)
	hub.Register("user-1", conn2)

	event := Event{
		Type:    "order_status",
		Payload: map[string]string{"orderID": "ord-123", "status": "preparing"},
	}
	hub.SendToUser("user-1", event)

	// Both connections should receive the message
	select {
	case msg := <-conn1.send:
		var got Event
		require.NoError(t, json.Unmarshal(msg, &got))
		assert.Equal(t, "order_status", got.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("conn1 did not receive message")
	}

	select {
	case msg := <-conn2.send:
		var got Event
		require.NoError(t, json.Unmarshal(msg, &got))
		assert.Equal(t, "order_status", got.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("conn2 did not receive message")
	}
}

func TestHub_SendToUserDoesNothingForUnknownUser(t *testing.T) {
	hub := NewHub()

	// Should not panic
	hub.SendToUser("unknown-user", Event{Type: "test", Payload: nil})
}

func TestHub_BroadcastDeliversToAllConnectedUsers(t *testing.T) {
	hub := NewHub()
	conn1 := mockConn("user-1")
	conn2 := mockConn("user-2")

	hub.Register("user-1", conn1)
	hub.Register("user-2", conn2)

	event := Event{
		Type:    "announcement",
		Payload: map[string]string{"message": "hello"},
	}
	hub.Broadcast(event)

	select {
	case msg := <-conn1.send:
		var got Event
		require.NoError(t, json.Unmarshal(msg, &got))
		assert.Equal(t, "announcement", got.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("conn1 did not receive broadcast")
	}

	select {
	case msg := <-conn2.send:
		var got Event
		require.NoError(t, json.Unmarshal(msg, &got))
		assert.Equal(t, "announcement", got.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("conn2 did not receive broadcast")
	}
}

func TestHub_ConcurrentRegisterUnregister(t *testing.T) {
	hub := NewHub()
	var wg sync.WaitGroup

	// Simulate concurrent register/unregister from multiple goroutines
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			userID := "user-concurrent"
			conn := mockConn(userID)
			hub.Register(userID, conn)
			time.Sleep(time.Millisecond)
			hub.Unregister(userID, conn)
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 0, hub.ConnCount())
}

func TestHub_ConcurrentSendToUser(t *testing.T) {
	hub := NewHub()
	conn := mockConn("user-1")
	hub.Register("user-1", conn)

	var wg sync.WaitGroup
	// Send from multiple goroutines concurrently
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			hub.SendToUser("user-1", Event{Type: "test", Payload: n})
		}(i)
	}
	wg.Wait()

	// Drain the channel — we should have 20 messages
	received := 0
	for {
		select {
		case <-conn.send:
			received++
		default:
			assert.Equal(t, 20, received)
			return
		}
	}
}
