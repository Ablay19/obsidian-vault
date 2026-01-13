package utils

import (
	"strconv"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	limits map[string]*UserLimit
	mutex  sync.RWMutex
	logger *Logger
}

// UserLimit tracks rate limit for a user
type UserLimit struct {
	hourlyCount int
	dailyCount  int
	hourlyReset time.Time
	dailyReset  time.Time
	mutex       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(logger *Logger) *RateLimiter {
	limiter := &RateLimiter{
		limits: make(map[string]*UserLimit),
		logger: logger,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a user is allowed to make a request
func (rl *RateLimiter) Allow(userID int64) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userKey := strconv.FormatInt(userID, 10)
	limit, exists := rl.limits[userKey]
	if !exists {
		limit = &UserLimit{
			hourlyReset: time.Now().Truncate(time.Hour).Add(time.Hour),
			dailyReset:  time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour),
		}
		rl.limits[userKey] = limit
	}

	limit.mutex.Lock()
	defer limit.mutex.Unlock()

	now := time.Now()

	// Reset counters if needed
	if now.After(limit.hourlyReset) {
		limit.hourlyCount = 0
		limit.hourlyReset = now.Truncate(time.Hour).Add(time.Hour)
	}

	if now.After(limit.dailyReset) {
		limit.dailyCount = 0
		limit.dailyReset = now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	}

	// Check limits
	const maxHourly = 100
	const maxDaily = 1000

	if limit.hourlyCount >= maxHourly {
		rl.logger.RateLimit(userID, "hour", limit.hourlyReset.Unix())
		return false
	}

	if limit.dailyCount >= maxDaily {
		rl.logger.RateLimit(userID, "day", limit.dailyReset.Unix())
		return false
	}

	// Increment counters
	limit.hourlyCount++
	limit.dailyCount++

	return true
}

// GetResetTime returns when rate limits will reset for a user
func (rl *RateLimiter) GetResetTime(userID int64, limitType string) int64 {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	userKey := strconv.FormatInt(userID, 10)
	limit, exists := rl.limits[userKey]
	if !exists {
		return time.Now().Unix()
	}

	limit.mutex.Lock()
	defer limit.mutex.Unlock()

	switch limitType {
	case "hour":
		return limit.hourlyReset.Unix()
	case "day":
		return limit.dailyReset.Unix()
	default:
		return time.Now().Unix()
	}
}

// Reset resets rate limits for a user
func (rl *RateLimiter) Reset(userID int64) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userKey := strconv.FormatInt(userID, 10)
	if limit, exists := rl.limits[userKey]; exists {
		limit.mutex.Lock()
		limit.hourlyCount = 0
		limit.dailyCount = 0
		limit.hourlyReset = time.Now().Truncate(time.Hour).Add(time.Hour)
		limit.dailyReset = time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
		limit.mutex.Unlock()
	}
}

// GetStats returns rate limiting statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_users": len(rl.limits),
	}

	return stats
}

// cleanup removes expired user limits
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()

		for userKey, limit := range rl.limits {
			limit.mutex.Lock()

			// Remove if both hourly and daily limits have expired
			hourlyExpired := now.After(limit.hourlyReset.Add(2 * time.Hour))
			dailyExpired := now.After(limit.dailyReset.Add(24 * time.Hour))

			if hourlyExpired && dailyExpired {
				delete(rl.limits, userKey)
			}

			limit.mutex.Unlock()
		}

		rl.mutex.Unlock()
	}
}
