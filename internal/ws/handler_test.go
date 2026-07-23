package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"
)

const testSecret = "test-secret-key"

func generateTestToken(userID string, expired bool) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": 0,
		"iat":  now.Unix(),
	}
	if expired {
		claims["exp"] = now.Add(-1 * time.Hour).Unix()
	} else {
		claims["exp"] = now.Add(1 * time.Hour).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))
	return signed
}

func TestHandleUpgrade_MissingTokenReturns401(t *testing.T) {
	hub := NewHub()
	handler := hub.HandleUpgrade(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/ws", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing token")
}

func TestHandleUpgrade_InvalidTokenReturns401(t *testing.T) {
	hub := NewHub()
	handler := hub.HandleUpgrade(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/ws?token=invalid.token.here", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid token")
}

func TestHandleUpgrade_ExpiredTokenReturns401(t *testing.T) {
	hub := NewHub()
	handler := hub.HandleUpgrade(testSecret)

	token := generateTestToken("user-1", true)
	req := httptest.NewRequest(http.MethodGet, "/api/ws?token="+token, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid token")
}

func TestHandleUpgrade_WrongSecretReturns401(t *testing.T) {
	hub := NewHub()
	handler := hub.HandleUpgrade(testSecret)

	// Generate token with a different secret
	claims := jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("wrong-secret"))

	req := httptest.NewRequest(http.MethodGet, "/api/ws?token="+signed, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleUpgrade_ValidTokenUpgradesConnection(t *testing.T) {
	hub := NewHub()
	handler := hub.HandleUpgrade(testSecret)

	server := httptest.NewServer(handler)
	defer server.Close()

	token := generateTestToken("user-ws-test", false)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + token

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ws, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer ws.Close(websocket.StatusNormalClosure, "done")

	// Give the hub time to register the connection
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, hub.UserConnCount("user-ws-test"))

	// Send an event and read it from the WebSocket
	hub.SendToUser("user-ws-test", Event{
		Type:    "test_event",
		Payload: map[string]string{"key": "value"},
	})

	readCtx, readCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer readCancel()

	_, msg, err := ws.Read(readCtx)
	require.NoError(t, err)
	assert.Contains(t, string(msg), "test_event")
	assert.Contains(t, string(msg), "value")
}

func TestHandleUpgrade_ClientDisconnectCleansUp(t *testing.T) {
	hub := NewHub()
	handler := hub.HandleUpgrade(testSecret)

	server := httptest.NewServer(handler)
	defer server.Close()

	token := generateTestToken("user-disconnect", false)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + token

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ws, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)

	// Wait for registration
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.UserConnCount("user-disconnect"))

	// Close from client side
	ws.Close(websocket.StatusNormalClosure, "bye")

	// Give the server time to process the disconnect
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, hub.UserConnCount("user-disconnect"))
}

func TestValidateToken_NoUserIDReturnsError(t *testing.T) {
	// Token with no sub or userId claim
	claims := jwt.MapClaims{
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
		"role": 0,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))

	_, err := validateToken(signed, testSecret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no user ID")
}

func TestValidateToken_UserIdClaimFallback(t *testing.T) {
	// Token with userId instead of sub
	claims := jwt.MapClaims{
		"userId": "user-fallback",
		"exp":    time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))

	userID, err := validateToken(signed, testSecret)
	assert.NoError(t, err)
	assert.Equal(t, "user-fallback", userID)
}
