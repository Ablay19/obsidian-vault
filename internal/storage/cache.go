package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"obsidian-automation/pkg/utils"
)

// Cache represents Redis cache client
type Cache struct {
	client *redis.Client
	logger *utils.Logger
}

// NewCache creates a new Redis cache client
func NewCache(redisURL, password string, db int, logger *utils.Logger) (*Cache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// Set password if provided
	if password != "" {
		opt.Password = password
	}

	// Set database
	opt.DB = db

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &Cache{
		client: client,
		logger: logger,
	}

	logger.CacheOperation("connect", "redis", true)
	return cache, nil
}

// Close closes the Redis connection
func (c *Cache) Close() error {
	return c.client.Close()
}

// Set stores a value with expiration
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal cache value", "key", key, "error", err)
		return err
	}

	if err := c.client.Set(ctx, key, jsonValue, expiration).Err(); err != nil {
		c.logger.CacheOperation("set", key, false)
		return err
	}

	c.logger.CacheOperation("set", key, true)
	return nil
}

// Get retrieves a value
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			c.logger.CacheOperation("get", key, false) // Cache miss
			return ErrCacheNotFound
		}
		c.logger.CacheOperation("get", key, false)
		return err
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		c.logger.Error("Failed to unmarshal cache value", "key", key, "error", err)
		return err
	}

	c.logger.CacheOperation("get", key, true) // Cache hit
	return nil
}

// Delete removes a value
func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.CacheOperation("delete", key, false)
		return err
	}

	c.logger.CacheOperation("delete", key, true)
	return nil
}

// Exists checks if a key exists
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		c.logger.CacheOperation("exists", key, false)
		return false, err
	}

	exists := result > 0
	c.logger.CacheOperation("exists", key, true)
	return exists, nil
}

// Increment increments a numeric value
func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	result, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		c.logger.CacheOperation("increment", key, false)
		return 0, err
	}

	c.logger.CacheOperation("increment", key, true)
	return result, nil
}

// SetWithTTL stores a value with time-to-live
func (c *Cache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl)
}

// GetWithTTL gets a value and its TTL
func (c *Cache) GetWithTTL(ctx context.Context, key string, dest interface{}) (time.Duration, error) {
	// Get the value
	if err := c.Get(ctx, key, dest); err != nil {
		return 0, err
	}

	// Get TTL
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return ttl, nil
}

// ClearPattern removes all keys matching a pattern
func (c *Cache) ClearPattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.logger.CacheOperation("clear_pattern", pattern, false)
		return err
	}

	c.logger.CacheOperation("clear_pattern", pattern, true)
	return nil
}

// Health checks Redis connection health
func (c *Cache) Health(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}
	return nil
}

// Stats returns cache statistics
func (c *Cache) Stats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"info": info,
	}

	return stats, nil
}

// Error definitions
var (
	ErrCacheNotFound = fmt.Errorf("cache key not found")
)
