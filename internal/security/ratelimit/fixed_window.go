package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// windowKey combines a key with a time window
type windowKey struct {
	key       string
	timestamp int64
}

// FixedWindowRateLimiter implements fixed window rate limiting
type FixedWindowRateLimiter struct {
	config   *RateLimitConfig
	counters map[windowKey]int
	mu       sync.RWMutex
	window   time.Duration
}

// NewFixedWindowRateLimiter creates a new fixed window rate limiter
func NewFixedWindowRateLimiter(config *RateLimitConfig) (*FixedWindowRateLimiter, error) {
	window, err := time.ParseDuration(config.Window)
	if err != nil {
		return nil, err
	}

	limiter := &FixedWindowRateLimiter{
		config:   config,
		counters: make(map[windowKey]int),
		window:   window,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter, nil
}

// Allow checks if a request is allowed based on rate limiting rules
func (l *FixedWindowRateLimiter) Allow(key string) (bool, int, time.Time) {
	now := time.Now()

	// Calculate the start of the current time window
	windowStart := now.Truncate(l.window)
	windowKey := windowKey{
		key:       key,
		timestamp: windowStart.Unix(),
	}

	// Calculate when the current window resets
	resetTime := windowStart.Add(l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get current count for this key and window
	count := l.counters[windowKey]

	// Check if limit has been reached
	if count >= l.config.Limit {
		return false, 0, resetTime
	}

	// Increment counter
	l.counters[windowKey]++

	// Return remaining tokens
	remaining := l.config.Limit - l.counters[windowKey]
	return true, remaining, resetTime
}

// GetConfig returns the rate limiter configuration
func (l *FixedWindowRateLimiter) GetConfig() *RateLimitConfig {
	return l.config
}

// ExtractKey extracts the rate limiting key from the request
func (l *FixedWindowRateLimiter) ExtractKey(req *http.Request, authData map[string]interface{}) string {
	switch l.config.Scope {
	case RateLimitScopeGlobal:
		return "global"

	case RateLimitScopeIP:
		return req.RemoteAddr

	case RateLimitScopeHeader:
		return req.Header.Get(l.config.Header)

	case RateLimitScopeAuth:
		if authData == nil {
			return "anonymous"
		}

		// Try to get user ID or key name
		if userID, ok := authData["sub"].(string); ok {
			return userID
		}
		if keyName, ok := authData["name"].(string); ok {
			return keyName
		}

		return "authenticated"

	default:
		return req.RemoteAddr
	}
}

// cleanup periodically removes expired window counters
func (l *FixedWindowRateLimiter) cleanup() {
	ticker := time.NewTicker(l.window)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		expired := now.Add(-l.window).Unix()

		l.mu.Lock()
		for wk := range l.counters {
			if wk.timestamp <= expired {
				delete(l.counters, wk)
			}
		}
		l.mu.Unlock()
	}
}
