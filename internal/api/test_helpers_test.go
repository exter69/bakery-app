package api

import (
	"net/http"

	"github.com/lucatorrekens/bakery-app/internal/middleware"
)

// testAuthMiddleware simulates the JWT auth middleware in tests.
// It reads the X-User-ID header and injects it into the request context,
// replicating what the real JWT middleware does after token validation.
func testAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.Header.Get("X-User-ID")
		if uid != "" {
			ctx := middleware.WithUserID(r.Context(), uid)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
