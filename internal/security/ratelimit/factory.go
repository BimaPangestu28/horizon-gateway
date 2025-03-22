package ratelimit

import (
	"fmt"
)

// NewRateLimiter creates a rate limiter based on configuration
func NewRateLimiter(config *RateLimitConfig) (RateLimiter, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Create appropriate rate limiter based on type
	switch config.Type {
	case RateLimitTypeFixedWindow:
		return NewFixedWindowRateLimiter(config)

	case RateLimitTypeSlidingWindow:
		return NewSlidingWindowRateLimiter(config)

	case RateLimitTypeTokenBucket:
		return NewTokenBucketRateLimiter(config)

	default:
		return nil, fmt.Errorf("unsupported rate limit type: %s", config.Type)
	}
}
