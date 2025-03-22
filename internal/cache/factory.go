package cache

import (
	"fmt"
)

// NewCache creates a cache instance based on configuration
func NewCache(config *CacheConfig) (Cache, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Create appropriate cache based on type
	switch config.Type {
	case CacheTypeMemory:
		return NewMemoryCache(config)

	case CacheTypeRedis:
		return NewRedisCache(config)

	default:
		return nil, fmt.Errorf("unsupported cache type: %s", config.Type)
	}
}
