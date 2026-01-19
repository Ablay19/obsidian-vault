package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	Password     string        `json:"password" yaml:"password"`
	DB           int           `json:"db" yaml:"db"`
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxConnAge   time.Duration `json:"max_conn_age" yaml:"max_conn_age"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
}

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client redis.UniversalClient
	logger *zap.Logger
	config RedisConfig
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config RedisConfig, logger *zap.Logger) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxConnAge:   config.MaxConnAge,
		IdleTimeout:  config.IdleTimeout,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &RedisCache{
		client: rdb,
		logger: logger,
		config: config,
	}

	logger.Info("Redis cache initialized",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.Int("db", config.DB),
		zap.Int("pool_size", config.PoolSize),
	)

	return cache, nil
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			rc.logger.Debug("Cache miss", zap.String("key", key))
			return nil, nil
		}
		rc.logger.Error("Failed to get from cache", zap.String("key", key), zap.Error(err))
		return nil, err
	}

	rc.logger.Debug("Cache hit", zap.String("key", key))
	return val, nil
}

// Set stores a value in cache with expiration
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			rc.logger.Error("Failed to marshal value for cache", zap.String("key", key), zap.Error(err))
			return err
		}
		val = string(jsonBytes)
	}

	err := rc.client.Set(ctx, key, val, expiration).Err()
	if err != nil {
		rc.logger.Error("Failed to set cache", zap.String("key", key), zap.Error(err))
		return err
	}

	rc.logger.Debug("Cache set", zap.String("key", key), zap.Duration("expiration", expiration))
	return nil
}

// Delete removes a key from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	err := rc.client.Del(ctx, key).Err()
	if err != nil {
		rc.logger.Error("Failed to delete from cache", zap.String("key", key), zap.Error(err))
		return err
	}

	rc.logger.Debug("Cache deleted", zap.String("key", key))
	return nil
}

// Exists checks if a key exists in cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		rc.logger.Error("Failed to check cache existence", zap.String("key", key), zap.Error(err))
		return false, err
	}

	exists := count > 0
	rc.logger.Debug("Cache existence check", zap.String("key", key), zap.Bool("exists", exists))
	return exists, nil
}

// Expire sets expiration on a key
func (rc *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := rc.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		rc.logger.Error("Failed to set expiration", zap.String("key", key), zap.Error(err))
		return err
	}

	rc.logger.Debug("Cache expiration set", zap.String("key", key), zap.Duration("expiration", expiration))
	return nil
}

// TTL returns the remaining time to live of a key
func (rc *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := rc.client.TTL(ctx, key).Result()
	if err != nil {
		rc.logger.Error("Failed to get TTL", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	rc.logger.Debug("Cache TTL retrieved", zap.String("key", key), zap.Duration("ttl", ttl))
	return ttl, nil
}

// Increment increments the number stored at key by one
func (rc *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	val, err := rc.client.Incr(ctx, key).Result()
	if err != nil {
		rc.logger.Error("Failed to increment cache", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	rc.logger.Debug("Cache incremented", zap.String("key", key), zap.Int64("value", val))
	return val, nil
}

// Decrement decrements the number stored at key by one
func (rc *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	val, err := rc.client.Decr(ctx, key).Result()
	if err != nil {
		rc.logger.Error("Failed to decrement cache", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	rc.logger.Debug("Cache decremented", zap.String("key", key), zap.Int64("value", val))
	return val, nil
}

// GetJSON retrieves and unmarshals a JSON value from cache
func (rc *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == nil {
		return nil
	}

	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("cache value is not a string")
	}

	err = json.Unmarshal([]byte(strVal), dest)
	if err != nil {
		rc.logger.Error("Failed to unmarshal JSON from cache", zap.String("key", key), zap.Error(err))
		return err
	}

	rc.logger.Debug("JSON retrieved from cache", zap.String("key", key))
	return nil
}

// SetJSON marshals and stores a value as JSON in cache
func (rc *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.Set(ctx, key, value, expiration)
}

// MGet retrieves multiple keys at once
func (rc *RedisCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	vals, err := rc.client.MGet(ctx, keys...).Result()
	if err != nil {
		rc.logger.Error("Failed to get multiple keys from cache", zap.Strings("keys", keys), zap.Error(err))
		return nil, err
	}

	rc.logger.Debug("Multiple keys retrieved from cache", zap.Int("count", len(keys)))
	return vals, nil
}

// MSet sets multiple key-value pairs
func (rc *RedisCache) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	if len(pairs) == 0 {
		return nil
	}

	// Convert to Redis format
	redisPairs := make([]interface{}, 0, len(pairs)*2)
	for key, value := range pairs {
		var val string
		switch v := value.(type) {
		case string:
			val = v
		case []byte:
			val = string(v)
		default:
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				rc.logger.Error("Failed to marshal value for MSet", zap.String("key", key), zap.Error(err))
				return err
			}
			val = string(jsonBytes)
		}
		redisPairs = append(redisPairs, key, val)
	}

	// Set values
	err := rc.client.MSet(ctx, redisPairs...).Err()
	if err != nil {
		rc.logger.Error("Failed to set multiple keys in cache", zap.Error(err))
		return err
	}

	// Set expiration for each key if specified
	if expiration > 0 {
		for key := range pairs {
			if err := rc.Expire(ctx, key, expiration); err != nil {
				rc.logger.Warn("Failed to set expiration for key in MSet", zap.String("key", key), zap.Error(err))
			}
		}
	}

	rc.logger.Debug("Multiple keys set in cache", zap.Int("count", len(pairs)))
	return nil
}

// FlushDB clears all keys in the current database
func (rc *RedisCache) FlushDB(ctx context.Context) error {
	err := rc.client.FlushDB(ctx).Err()
	if err != nil {
		rc.logger.Error("Failed to flush database", zap.Error(err))
		return err
	}

	rc.logger.Info("Database flushed")
	return nil
}

// Ping tests the connection to Redis
func (rc *RedisCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	rc.logger.Info("Closing Redis cache connection")
	return rc.client.Close()
}

// GetStats returns cache statistics
func (rc *RedisCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := rc.client.Info(ctx, "stats").Result()
	if err != nil {
		rc.logger.Error("Failed to get Redis stats", zap.Error(err))
		return nil, err
	}

	stats := map[string]interface{}{
		"redis_info": info,
		"pool_size":  rc.config.PoolSize,
		"host":       rc.config.Host,
		"port":       rc.config.Port,
		"db":         rc.config.DB,
	}

	return stats, nil
}

// Publish publishes a message to a Redis channel
func (rc *RedisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	var msg string
	switch v := message.(type) {
	case string:
		msg = v
	default:
		jsonBytes, err := json.Marshal(message)
		if err != nil {
			rc.logger.Error("Failed to marshal message for publish", zap.String("channel", channel), zap.Error(err))
			return err
		}
		msg = string(jsonBytes)
	}

	err := rc.client.Publish(ctx, channel, msg).Err()
	if err != nil {
		rc.logger.Error("Failed to publish message", zap.String("channel", channel), zap.Error(err))
		return err
	}

	rc.logger.Debug("Message published", zap.String("channel", channel))
	return nil
}

// Subscribe subscribes to Redis channels
func (rc *RedisCache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	pubsub := rc.client.Subscribe(ctx, channels...)
	rc.logger.Info("Subscribed to channels", zap.Strings("channels", channels))
	return pubsub
}

// Pipeline creates a Redis pipeline
func (rc *RedisCache) Pipeline() redis.Pipeliner {
	return rc.client.Pipeline()
}

// TxPipeline creates a Redis transaction pipeline
func (rc *RedisCache) TxPipeline() redis.Pipeliner {
	return rc.client.TxPipeline()
}

// DefaultRedisConfig returns a default Redis configuration
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxConnAge:   30 * time.Minute,
		IdleTimeout:  5 * time.Minute,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// NewRedisCluster creates a Redis cluster client
func NewRedisCluster(addrs []string, password string, logger *zap.Logger) (*RedisCache, error) {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,
		Password: password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	cache := &RedisCache{
		client: rdb,
		logger: logger,
	}

	logger.Info("Redis cluster cache initialized", zap.Strings("addrs", addrs))
	return cache, nil
}

// CacheKey generates a standardized cache key
func CacheKey(prefix, identifier string) string {
	return fmt.Sprintf("%s:%s", prefix, identifier)
}

// Common cache key prefixes
const (
	UserPrefix      = "user"
	SessionPrefix   = "session"
	APIKeyPrefix    = "apikey"
	ResponsePrefix  = "response"
	RateLimitPrefix = "ratelimit"
	ConfigPrefix    = "config"
)
