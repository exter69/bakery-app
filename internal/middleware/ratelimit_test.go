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

func TestRateLimiter_EvictsStaleEntries(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 5,
		Window:      10 * time.Second,
		UserIDExtractor: func(r *http.Request) string {
			return r.Header.Get("X-User-ID")
		},
	}
	rl := NewRateLimiter(config)

	// Use a controllable time function
	currentTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	rl.now = func() time.Time { return currentTime }
	rl.lastCleanup = currentTime

	// Generate requests from 100 different "users" (simulating stale keys)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.Header.Set("X-User-ID", "ip-"+string(rune('0'+i%10))+string(rune('0'+i/10)))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// All 100 entries should exist
	rl.mu.Lock()
	assert.Equal(t, 100, len(rl.windows))
	rl.mu.Unlock()

	// Advance time past the window + cleanup interval (5x window = 50 seconds)
	currentTime = currentTime.Add(60 * time.Second)

	// Trigger a new request to trigger the cleanup
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.Header.Set("X-User-ID", "new-user")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// All stale entries should be evicted, only "new-user" should remain
	rl.mu.Lock()
	assert.Equal(t, 1, len(rl.windows))
	rl.mu.Unlock()
}
