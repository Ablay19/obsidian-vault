package utils

import (
	"log"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
)

// RateLimiter implements rate limiting for transport clients
type RateLimiter struct {
	maxRequests int
	window      time.Duration
	requests    []time.Time
	logger      *log.Logger
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, window time.Duration, logger *log.Logger) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		requests:    make([]time.Time, 0),
		logger:      logger,
	}
}

// IsRateLimited checks if the current request would exceed the rate limit
func (rl *RateLimiter) IsRateLimited() bool {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Remove old requests outside the window
	validRequests := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(windowStart) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	return len(rl.requests) >= rl.maxRequests
}

// RecordUsage records a request usage
func (rl *RateLimiter) RecordUsage() {
	now := time.Now()
	rl.requests = append(rl.requests, now)

	// Clean up old requests
	windowStart := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(windowStart) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	rl.logger.Printf("Rate limiter: %d requests in window", len(rl.requests))
}

// GetStatus returns the current rate limit status
func (rl *RateLimiter) GetStatus() (*models.RateLimit, error) {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Count valid requests
	validCount := 0
	for _, req := range rl.requests {
		if req.After(windowStart) {
			validCount++
		}
	}

	resetTime := now.Add(rl.window)

	return &models.RateLimit{
		RequestsPerHour:   rl.maxRequests,
		RequestsRemaining: rl.maxRequests - validCount,
		ResetTime:         resetTime,
		IsThrottled:       validCount >= rl.maxRequests,
	}, nil
}
