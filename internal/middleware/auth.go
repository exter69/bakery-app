package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a private type for context keys in this package.
type contextKey string

const (
	// userIDKey is the context key for the authenticated user ID.
	userIDKey contextKey = "userID"

	// userRoleKey is the context key for the authenticated user's role.
	userRoleKey contextKey = "userRole"
)

// authErrorResponse is the JSON structure returned for authentication errors.
type authErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JWTAuth returns an HTTP middleware that validates JWT tokens from the Authorization header.
// It extracts the user ID from the "sub" claim and stores it in the request context.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}

			// Expect "Bearer <token>" format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}

			tokenString := parts[1]
			if tokenString == "" {
				writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}

			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Ensure the signing method is HMAC
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(secret), nil
			})

			if err != nil {
				// Determine error type
				if errors.Is(err, jwt.ErrTokenExpired) {
					writeAuthError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "token has expired")
					return
				}
				writeAuthError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid token")
				return
			}

			if !token.Valid {
				writeAuthError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid token")
				return
			}

			// Extract user ID from claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid token")
				return
			}

			// Try "sub" claim first, then "userId"
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
				writeAuthError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid token")
				return
			}

			// Extract role from claims (optional)
			userRole := -1
			if r, exists := claims["role"]; exists {
				switch v := r.(type) {
				case float64:
					userRole = int(v)
				case int:
					userRole = v
				}
			}

			// Store user ID and role in context and proceed
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, userRoleKey, userRole)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extracts the authenticated user ID from the request context.
// Returns an empty string if no user ID is present.
func GetUserIDFromContext(ctx context.Context) string {
	if uid, ok := ctx.Value(userIDKey).(string); ok {
		return uid
	}
	return ""
}

// GetUserRoleFromContext extracts the authenticated user's role from the request context.
// Returns -1 if no role is present.
func GetUserRoleFromContext(ctx context.Context) int {
	if role, ok := ctx.Value(userRoleKey).(int); ok {
		return role
	}
	return -1
}

// writeAuthError writes a JSON error response for authentication failures.
func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(authErrorResponse{
		Code:    code,
		Message: message,
	})
}

// WithUserID returns a new context with the given user ID set.
// Intended for use in tests to inject auth context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithUserRole returns a new context with the given user role set.
// Intended for use in tests to inject auth context.
func WithUserRole(ctx context.Context, role int) context.Context {
	return context.WithValue(ctx, userRoleKey, role)
}
