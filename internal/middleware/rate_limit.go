package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
	"obsidian-automation/pkg/utils"
)

type RateLimitMiddleware struct {
	rateLimiter *utils.RateLimiter
	logger      *zap.Logger
}

type RateLimitResponse struct {
	Error      string    `json:"error"`
	Code       int       `json:"code"`
	ResetAt    int64     `json:"reset_at,omitempty"`
	RetryAfter int       `json:"retry_after,omitempty"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	Time       time.Time `json:"time"`
}

func NewRateLimitMiddleware(rateLimiter *utils.RateLimiter, logger *zap.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (m *RateLimitMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context or request
			userID := m.extractUserID(r)
			if userID == 0 {
				// For anonymous requests, use a default rate limit or IP-based
				ip := r.RemoteAddr
				if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
					ip = forwarded
				}
				// Convert IP to int64 for rate limiting (simple hash)
				userID = m.ipToInt64(ip)
			}

			// Check rate limit
			if !m.rateLimiter.Allow(userID) {
				m.handleRateLimitExceeded(w, r, userID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *RateLimitMiddleware) handleRateLimitExceeded(w http.ResponseWriter, r *http.Request, userID int64) {
	resetTime := m.rateLimiter.GetResetTime(userID, "hour")

	response := RateLimitResponse{
		Error:      "Rate limit exceeded",
		Code:       http.StatusTooManyRequests,
		ResetAt:    resetTime,
		RetryAfter: int(time.Unix(resetTime, 0).Sub(time.Now()).Seconds()),
		Path:       r.URL.Path,
		Method:     r.Method,
		Time:       time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
	w.Header().Set("Retry-After", strconv.Itoa(response.RetryAfter))
	w.WriteHeader(http.StatusTooManyRequests)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		m.logger.Error("Failed to encode rate limit response", zap.Error(err))
	}

	m.logger.Warn("Rate limit exceeded",
		zap.Int64("user_id", userID),
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.Int64("reset_time", resetTime),
	)
}

func (m *RateLimitMiddleware) extractUserID(r *http.Request) int64 {
	// Try different methods to extract user ID

	// 1. From JWT token in Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		if userID := m.extractFromJWT(auth); userID != 0 {
			return userID
		}
	}

	// 2. From session cookie
	if cookie, err := r.Cookie("session"); err == nil {
		if userID := m.extractFromSession(cookie.Value); userID != 0 {
			return userID
		}
	}

	// 3. From query parameter (for API keys)
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			return userID
		}
	}

	// 4. From header
	if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			return userID
		}
	}

	return 0 // Anonymous
}

func (m *RateLimitMiddleware) extractFromJWT(authHeader string) int64 {
	// This would parse JWT token and extract user ID
	// Implementation depends on your JWT structure
	// For now, return 0
	return 0
}

func (m *RateLimitMiddleware) extractFromSession(sessionToken string) int64 {
	// This would validate session token and extract user ID
	// Implementation depends on your session store
	// For now, return 0
	return 0
}

func (m *RateLimitMiddleware) ipToInt64(ip string) int64 {
	// Simple hash function for IP addresses
	// This is not cryptographically secure but sufficient for rate limiting
	var hash int64
	for _, char := range ip {
		hash = hash*31 + int64(char)
	}
	return hash
}

// GetStats returns rate limiting statistics
func (m *RateLimitMiddleware) GetStats() map[string]interface{} {
	return m.rateLimiter.GetStats()
}

// ResetUserLimits resets rate limits for a specific user
func (m *RateLimitMiddleware) ResetUserLimits(userID int64) {
	m.rateLimiter.Reset(userID)
}
