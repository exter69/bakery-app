package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigin is the single origin allowed to make cross-origin requests.
	// Should be set from FRONTEND_ORIGIN env var. Never use "*" in production.
	AllowedOrigin string
}

// CORS returns a middleware that applies restrictive CORS headers.
// In development mode (APP_ENV=development), it allows any origin for convenience.
// In production, only the configured AllowedOrigin is permitted.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	isDev := strings.EqualFold(os.Getenv("APP_ENV"), "development")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if isDev && origin != "" {
				// Permissive CORS in development
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if origin == cfg.AllowedOrigin && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", cfg.AllowedOrigin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "Link")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
