package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimitConfig())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Send 10 requests (the limit) — all should pass
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
		req.Header.Set("X-User-ID", "user-1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "request %d should pass", i+1)
	}
}

func TestRateLimiter_ExceedsLimit(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimitConfig())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Send 10 requests that should pass
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
		req.Header.Set("X-User-ID", "user-1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// 11th request should be rate-limited
	req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set("X-User-ID", "user-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	var errResp rateLimitError
	err := json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "RATE_LIMITED", errResp.Code)
	assert.Equal(t, "too many requests, please try again later", errResp.Message)
}

func TestRateLimiter_WindowReset(t *testing.T) {
	config := DefaultRateLimitConfig()
	rl := NewRateLimiter(config)

	// Use a controllable time function
	currentTime := time.Now()
	rl.now = func() time.Time { return currentTime }

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
		req.Header.Set("X-User-ID", "user-1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Verify request is rejected
	req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set("X-User-ID", "user-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// Advance time past the window
	currentTime = currentTime.Add(61 * time.Second)

	// Request should be allowed again
	req = httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set("X-User-ID", "user-1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimiter_DifferentUsersIndependent(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimitConfig())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit for user-1
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
		req.Header.Set("X-User-ID", "user-1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// user-1 is rate limited
	req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set("X-User-ID", "user-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// user-2 should still be allowed
	req = httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set("X-User-ID", "user-2")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimiter_NoUserID_PassesThrough(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimitConfig())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Without user ID header, request should pass through (no rate limiting)
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "request %d without user ID should pass", i+1)
	}
}

func TestRateLimiter_ContentTypeHeader(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimitConfig())

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust limit
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
		req.Header.Set("X-User-ID", "user-1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Rate-limited response should have JSON content type
	req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set("X-User-ID", "user-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
