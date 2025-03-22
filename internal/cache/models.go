package cache

import (
	"net/http"
	"time"
)

// CacheType represents the type of caching to use
type CacheType string

const (
	// CacheTypeMemory uses in-memory caching
	CacheTypeMemory CacheType = "memory"

	// CacheTypeRedis uses Redis as the cache backend
	CacheTypeRedis CacheType = "redis"
)

// CacheConfig defines caching configuration
type CacheConfig struct {
	// Enabled indicates whether caching is enabled
	Enabled bool `yaml:"enabled"`

	// Type is the cache storage type to use
	Type CacheType `yaml:"type" default:"memory"`

	// TTL is the default time-to-live for cached responses
	TTL string `yaml:"ttl" default:"1m"`

	// MaxSize is the maximum size in bytes of the response to cache (0 = no limit)
	MaxSize int64 `yaml:"max_size" default:"1048576"` // 1MB

	// Methods defines which HTTP methods to cache (default: GET only)
	Methods []string `yaml:"methods" default:"[GET]"`

	// CacheKeyTemplate defines a template for cache key generation
	CacheKeyTemplate string `yaml:"cache_key_template" default:"{method}:{path}:{query}"`

	// IgnoreQueryParams lists query parameters to exclude from the cache key
	IgnoreQueryParams []string `yaml:"ignore_query_params,omitempty"`

	// VaryHeaders lists headers to include in the cache key
	VaryHeaders []string `yaml:"vary_headers,omitempty"`

	// NeverCache lists paths that should never be cached
	NeverCache []string `yaml:"never_cache,omitempty"`

	// AlwaysCache lists paths that should always be cached regardless of Cache-Control
	AlwaysCache []string `yaml:"always_cache,omitempty"`

	// RespectCacheControl indicates whether to respect Cache-Control headers
	RespectCacheControl bool `yaml:"respect_cache_control" default:"true"`

	// IncludeHost indicates whether to include host in the cache key
	IncludeHost bool `yaml:"include_host" default:"false"`

	// Redis configuration (when Type is "redis")
	Redis *RedisCacheConfig `yaml:"redis,omitempty"`
}

// RedisCacheConfig defines Redis connection settings
type RedisCacheConfig struct {
	// Address is the Redis server address
	Address string `yaml:"address" default:"localhost:6379"`

	// Password is the Redis server password
	Password string `yaml:"password,omitempty"`

	// DB is the Redis database number
	DB int `yaml:"db" default:"0"`

	// KeyPrefix is the prefix for cache keys
	KeyPrefix string `yaml:"key_prefix" default:"horizon:cache:"`
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	// Status is the HTTP status code
	Status int

	// Headers contains the response headers
	Headers http.Header

	// Body is the response body
	Body []byte

	// Created is when the response was cached
	Created time.Time

	// Expires is when the response expires
	Expires time.Time

	// ETag is the entity tag for the response (if any)
	ETag string

	// LastModified is the last modified time (if any)
	LastModified string
}

// Cache interface defines caching functionality
type Cache interface {
	// Get retrieves a cached response
	Get(key string) (*CachedResponse, bool)

	// Set stores a response in the cache
	Set(key string, response *CachedResponse) error

	// Delete removes a cached response
	Delete(key string) error

	// Clear clears all cached responses
	Clear() error

	// GenerateKey generates a cache key from a request
	GenerateKey(req *http.Request) string

	// ShouldCache determines if a request/response should be cached
	ShouldCache(req *http.Request, resp *http.Response) bool

	// Close closes the cache
	Close() error
}
