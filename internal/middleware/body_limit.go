package middleware

import (
	"net/http"
)

const (
	// DefaultBodyLimit is 1MB — applied to regular API endpoints.
	DefaultBodyLimit int64 = 1 << 20 // 1 MB

	// UploadBodyLimit is 10MB — applied to file upload endpoints.
	UploadBodyLimit int64 = 10 << 20 // 10 MB
)

// BodyLimit returns a middleware that restricts the request body to maxBytes.
// If the client sends more data than allowed, the read will fail with an error,
// and the server responds with 413 Request Entity Too Large.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
