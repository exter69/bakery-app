package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig holds configuration for the rate limiter.
type RateLimitConfig struct {
	// MaxRequests is the maximum number of requests allowed per window.
	MaxRequests int
	// Window is the duration of the sliding window.
	Window time.Duration
	// UserIDExtractor extracts the user ID from the request.
	// If nil, defaults to reading the X-User-ID header.
	UserIDExtractor func(r *http.Request) string
}

// DefaultRateLimitConfig returns the default rate limit configuration:
// 10 requests per user per 60-second window.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests: 10,
		Window:      60 * time.Second,
		UserIDExtractor: func(r *http.Request) string {
			if uid := r.Header.Get("X-User-ID"); uid != "" {
				return uid
			}
			return ""
		},
	}
}

// userWindow tracks request timestamps within a window for a single user.
type userWindow struct {
	timestamps []time.Time
}

// RateLimiter is an in-memory per-user rate limiter using a sliding window.
// Stale entries are evicted periodically to bound memory growth.
type RateLimiter struct {
	mu          sync.Mutex
	config      RateLimitConfig
	windows     map[string]*userWindow
	now         func() time.Time // for testing
	lastCleanup time.Time
}

// NewRateLimiter creates a new RateLimiter with the given configuration.
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config:      config,
		windows:     make(map[string]*userWindow),
		now:         time.Now,
		lastCleanup: time.Now(),
	}
}

// cleanupInterval determines how often stale entries are evicted.
// Set to 5x the window duration (e.g. 5 minutes for a 1-minute window).
const cleanupMultiplier = 5

// evictStale removes entries with no timestamps within the current window.
// Called under the existing mutex — no additional locking needed.
func (rl *RateLimiter) evictStale(now time.Time) {
	interval := rl.config.Window * cleanupMultiplier
	if now.Sub(rl.lastCleanup) < interval {
		return
	}
	windowStart := now.Add(-rl.config.Window)
	for key, w := range rl.windows {
		if len(w.timestamps) == 0 {
			delete(rl.windows, key)
			continue
		}
		// If the newest timestamp is older than the window, the entry is stale.
		if w.timestamps[len(w.timestamps)-1].Before(windowStart) {
			delete(rl.windows, key)
		}
	}
	rl.lastCleanup = now
}

// allow checks if a user is within the rate limit. If allowed, it records the request.
func (rl *RateLimiter) allow(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()

	// Periodically evict stale entries to bound memory.
	rl.evictStale(now)

	windowStart := now.Add(-rl.config.Window)

	w, exists := rl.windows[userID]
	if !exists {
		w = &userWindow{}
		rl.windows[userID] = w
	}

	// Remove expired timestamps outside the window.
	valid := w.timestamps[:0]
	for _, ts := range w.timestamps {
		if ts.After(windowStart) {
			valid = append(valid, ts)
		}
	}
	w.timestamps = valid

	if len(w.timestamps) >= rl.config.MaxRequests {
		return false
	}

	w.timestamps = append(w.timestamps, now)
	return true
}

// rateLimitError is the JSON error response for rate-limited requests.
type rateLimitError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Middleware returns an HTTP middleware that enforces rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := rl.config.UserIDExtractor(r)
		if userID == "" {
			// If we can't identify the user, let the request through
			// (auth middleware should handle unauthenticated requests).
			next.ServeHTTP(w, r)
			return
		}

		if !rl.allow(userID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(rateLimitError{
				Code:    "RATE_LIMITED",
				Message: "too many requests, please try again later",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
