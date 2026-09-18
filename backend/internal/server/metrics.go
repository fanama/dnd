package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks basic operational counters exposed on /api/metrics.
type Metrics struct {
	activeConnections atomic.Int64
	totalActions      atomic.Int64
	authFailures      atomic.Int64
	rateLimited       atomic.Int64
	wsErrors          atomic.Int64

	mu          sync.Mutex
	actionTimes []time.Time
}

var metrics Metrics

// RecordAction records an accepted action and timestamps it for the
// actions-per-minute figure.
func RecordAction() {
	metrics.totalActions.Add(1)
	now := time.Now()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	metrics.actionTimes = append(metrics.actionTimes, now)
	cutoff := now.Add(-time.Minute)
	i := 0
	for i < len(metrics.actionTimes) && metrics.actionTimes[i].Before(cutoff) {
		i++
	}
	metrics.actionTimes = metrics.actionTimes[i:]
}

func actionsPerMinute() int {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	cutoff := time.Now().Add(-time.Minute)
	n := 0
	for _, t := range metrics.actionTimes {
		if !t.Before(cutoff) {
			n++
		}
	}
	return n
}

// ActiveConnectionsDelta adjusts the live connection counter.
func ActiveConnectionsDelta(delta int) {
	metrics.activeConnections.Add(int64(delta))
}

// MarkAuthFailure increments the auth-failure counter.
func MarkAuthFailure() {
	metrics.authFailures.Add(1)
}

// MarkRateLimited increments the rate-limit counter.
func MarkRateLimited() {
	metrics.rateLimited.Add(1)
}

// MarkWSError increments the WebSocket error counter.
func MarkWSError() {
	metrics.wsErrors.Add(1)
}

// MetricsHandler returns a JSON snapshot of the counters.
func MetricsHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]any{
			"active_connections": metrics.activeConnections.Load(),
			"total_actions":      metrics.totalActions.Load(),
			"actions_per_minute": actionsPerMinute(),
			"auth_failures":      metrics.authFailures.Load(),
			"rate_limited":       metrics.rateLimited.Load(),
			"ws_errors":          metrics.wsErrors.Load(),
			"uptime_seconds":     int(time.Since(startTime).Seconds()),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			logger.Error("metrics encode failed", "err", err)
		}
	}
}

var startTime = time.Now()