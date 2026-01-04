package security

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRedisRateLimiter(redisAddr string, limit int, window time.Duration) *RateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &RateLimiter{
		client: rdb,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) Allow(key string) (bool, error) {
	ctx := context.Background()
	pipe := rl.client.Pipeline()
	now := time.Now().UnixNano()
	windowStart := now - rl.window.Nanoseconds()

	// Remove old entries from the sorted set.
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	// Add the current request's timestamp.
	pipe.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: now})
	// Get the count of requests in the current window.
	count := pipe.ZCard(ctx, key)
	// Set the key to expire after the window duration to clean up.
	pipe.Expire(ctx, key, rl.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return count.Val() <= int64(rl.limit), nil
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		allowed, err := rl.Allow(ip)
		if err != nil {
			// If Redis is down, we can choose to fail open or closed.
			// For now, we'll fail open.
			next.ServeHTTP(w, r)
			return
		}
		if !allowed {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		allowed, err := rl.Allow(ip)
		if err != nil {
			// If Redis is down, we can choose to fail open or closed.
			// For now, we'll fail open.
			c.Next()
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
