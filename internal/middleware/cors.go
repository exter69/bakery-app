package middleware

import (
	"net/http"
	"os"
	"strings"
)

// devAllowedOrigins is the fixed set of origins permitted in development mode.
// Only local dev servers are allowed — never reflect arbitrary origins.
var devAllowedOrigins = map[string]bool{
	"http://localhost:5173":  true,
	"http://localhost:3000":  true,
	"http://127.0.0.1:5173": true,
	"http://127.0.0.1:3000": true,
}

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigin is the single origin allowed to make cross-origin requests.
	// Should be set from FRONTEND_ORIGIN env var. Never use "*" in production.
	AllowedOrigin string
}

// CORS returns a middleware that applies restrictive CORS headers.
// In development mode (APP_ENV=development), it allows only origins in the
// fixed devAllowedOrigins allowlist.
// In production, only the configured AllowedOrigin is permitted.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	isDev := strings.EqualFold(os.Getenv("APP_ENV"), "development")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			if isDev {
				allowed = devAllowedOrigins[origin]
			} else {
				allowed = origin != "" && origin == cfg.AllowedOrigin
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "Link")
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
