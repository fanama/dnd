package server

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter is a simple token bucket used to throttle WebSocket actions
// (per connection) and the login endpoint (per client IP).
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64 // tokens refilled per second
	burst   float64 // maximum burst
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewRateLimiter(rate, burst float64) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   burst,
	}
}

// Allow consumes one token; it reports whether the caller may proceed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	now := time.Now()
	if !ok {
		b = &bucket{tokens: rl.burst, last: now}
		rl.buckets[key] = b
	}
	elapsed := now.Sub(b.last).Seconds()
	b.last = now
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Prune drops buckets that have been idle for a while to bound memory use.
func (rl *RateLimiter) Prune() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-time.Hour)
	for key, b := range rl.buckets {
		if b.last.Before(cutoff) {
			delete(rl.buckets, key)
		}
	}
}

// CheckOrigin builds a webSocket origin checker that accepts requests with no
// Origin header (non-browser clients), same-origin requests, and an explicit
// allowlist configured via ALLOWED_ORIGINS.
func CheckOrigin(allowed []string, logger *slog.Logger) func(*http.Request) bool {
	allowedSet := make(map[string]bool, len(allowed))
	for _, o := range allowed {
		if o = strings.TrimSpace(o); o != "" {
			allowedSet[o] = true
		}
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		if allowedSet[origin] {
			return true
		}
		if strings.EqualFold(origin, "http://"+r.Host) || strings.EqualFold(origin, "https://"+r.Host) {
			return true
		}
		logger.Warn("origin rejected", "origin", origin, "host", r.Host)
		return false
	}
}