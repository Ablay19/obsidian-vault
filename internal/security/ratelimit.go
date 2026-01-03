package security

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	tokens map[string]int
	mu     sync.Mutex
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(map[string]int),
		limit:  limit,
		window: window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	count := rl.tokens[key]
	if count >= rl.limit {
		return false
	}

	rl.tokens[key] = count + 1
	return true
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		rl.tokens = make(map[string]int)
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Identify by IP or session
		ip := r.RemoteAddr
		if !rl.Allow(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
