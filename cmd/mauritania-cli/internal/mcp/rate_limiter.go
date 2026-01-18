package mcp

import (
	"sync"
	"time"
)

// RateLimiter implements rate limiting for MCP tool calls
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerHour int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requestsPerHour,
		window:   time.Hour,
	}
}

// Allow checks if a request from the given identifier is allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean old requests
	if times, exists := rl.requests[identifier]; exists {
		var valid []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		rl.requests[identifier] = valid

		if len(valid) >= rl.limit {
			return false
		}
	}

	// Add current request
	rl.requests[identifier] = append(rl.requests[identifier], now)
	return true
}
