package cache

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements Redis-based caching
type RedisCache struct {
	config    *CacheConfig
	client    *redis.Client
	ttl       time.Duration
	keyPrefix string
	memCache  *MemoryCache // For key generation and ShouldCache logic
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(config *CacheConfig) (*RedisCache, error) {
	// Create memory cache for key generation and ShouldCache logic
	memCache, err := NewMemoryCache(config)
	if err != nil {
		return nil, err
	}

	// Parse TTL
	ttl, err := time.ParseDuration(config.TTL)
	if err != nil {
		return nil, err
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	cache := &RedisCache{
		config:    config,
		client:    client,
		ttl:       ttl,
		keyPrefix: config.Redis.KeyPrefix,
		memCache:  memCache,
	}

	return cache, nil
}

// Get retrieves a cached response
func (c *RedisCache) Get(key string) (*CachedResponse, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Get from Redis
	data, err := c.client.Get(ctx, c.keyPrefix+key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		return nil, false
	}

	// Deserialize
	var resp CachedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, false
	}

	// Check if response has expired
	if time.Now().After(resp.Expires) {
		// Delete expired item
		c.Delete(key)
		return nil, false
	}

	return &resp, true
}

// Set stores a response in the cache
func (c *RedisCache) Set(key string, response *CachedResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Serialize
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	// Calculate TTL
	ttl := response.Expires.Sub(time.Now())
	if ttl <= 0 {
		return nil // Don't cache expired responses
	}

	// Store in Redis
	return c.client.Set(ctx, c.keyPrefix+key, data, ttl).Err()
}

// Delete removes a cached response
func (c *RedisCache) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return c.client.Del(ctx, c.keyPrefix+key).Err()
}

// Clear clears all cached responses
func (c *RedisCache) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find all keys with our prefix
	pattern := c.keyPrefix + "*"
	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error

		batch, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
	}

	// Delete keys
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// GenerateKey generates a cache key from a request
func (c *RedisCache) GenerateKey(req *http.Request) string {
	return c.memCache.GenerateKey(req)
}

// ShouldCache determines if a request/response should be cached
func (c *RedisCache) ShouldCache(req *http.Request, resp *http.Response) bool {
	return c.memCache.ShouldCache(req, resp)
}

// Close closes the cache
func (c *RedisCache) Close() error {
	return c.client.Close()
}
