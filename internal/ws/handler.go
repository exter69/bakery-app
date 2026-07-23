package ws

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"nhooyr.io/websocket"
)

const (
	// pongWait is how long we wait for a pong response before considering the connection dead.
	pongWait = 60 * time.Second

	// pingInterval is how often we send ping frames to keep the connection alive.
	pingInterval = 30 * time.Second
)

// HandleUpgrade returns an HTTP handler that upgrades connections to WebSocket.
// Authentication is done via a JWT token passed as the `token` query parameter.
func (h *Hub) HandleUpgrade(jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract and validate JWT from query param
		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		userID, err := validateToken(tokenStr, jwtSecret)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 2. Upgrade to WebSocket
		ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			// Allow all origins in dev; in production, configure InsecureSkipVerify or set specific origins.
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Printf("[WS] upgrade failed for user %s: %v", userID, err)
			return
		}

		// 3. Create connection and register with hub
		conn := NewConn(ws, userID)
		h.Register(userID, conn)

		// 4. Start write pump in a goroutine
		go conn.WritePump()

		// 5. Read pump (blocks until disconnect) — also handles ping/pong
		h.readPump(conn)

		// 6. Clean up on disconnect
		conn.Close()
		h.Unregister(userID, conn)
	}
}

// readPump blocks reading from the WebSocket. We don't expect client messages,
// but we need to read to detect disconnection and handle control frames.
func (h *Hub) readPump(conn *Conn) {
	// Set up a ping ticker to keep connections alive through proxies/LBs
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	ctx := context.Background()

	for {
		// Use a context with timeout to detect dead connections
		readCtx, cancel := context.WithTimeout(ctx, pongWait)
		_, _, err := conn.ws.Read(readCtx)
		cancel()

		if err != nil {
			// Normal close or network error — exit
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				return
			}
			// Timeout or other error
			if errors.Is(err, context.DeadlineExceeded) {
				// Try a ping
				pingCtx, pingCancel := context.WithTimeout(ctx, 10*time.Second)
				pingErr := conn.ws.Ping(pingCtx)
				pingCancel()
				if pingErr != nil {
					return
				}
				continue
			}
			return
		}
		// We received a message from the client — we discard it (no client→server protocol)
	}
}

// validateToken parses and validates the JWT token, returning the user ID from claims.
func validateToken(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	// Try "sub" first, then "userId" (matching middleware/auth.go logic)
	userID := ""
	if sub, exists := claims["sub"]; exists {
		userID, _ = sub.(string)
	}
	if userID == "" {
		if uid, exists := claims["userId"]; exists {
			userID, _ = uid.(string)
		}
	}

	if userID == "" {
		return "", errors.New("no user ID in token")
	}

	return userID, nil
}
