package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// bucketState represents the state of a token bucket
type bucketState struct {
	tokens    float64
	timestamp time.Time
}

// TokenBucketRateLimiter implements token bucket rate limiting
type TokenBucketRateLimiter struct {
	config     *RateLimitConfig
	buckets    map[string]bucketState
	mu         sync.RWMutex
	burstSize  int
	refillRate float64
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter
func NewTokenBucketRateLimiter(config *RateLimitConfig) (*TokenBucketRateLimiter, error) {
	// Parse window to get refill rate if not explicitly set
	window, err := time.ParseDuration(config.Window)
	if err != nil {
		return nil, err
	}

	// Use configured burst size, or limit if not specified
	burstSize := config.BurstSize
	if burstSize <= 0 {
		burstSize = config.Limit
	}

	// Use configured refill rate, or calculate from limit and window
	refillRate := config.RefillRate
	if refillRate <= 0 {
		refillRate = float64(config.Limit) / window.Seconds()
	}

	limiter := &TokenBucketRateLimiter{
		config:     config,
		buckets:    make(map[string]bucketState),
		burstSize:  burstSize,
		refillRate: refillRate,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter, nil
}

// Allow checks if a request is allowed based on rate limiting rules
func (l *TokenBucketRateLimiter) Allow(key string) (bool, int, time.Time) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get current bucket state
	bucket, exists := l.buckets[key]
	if !exists {
		// Initialize bucket with maximum tokens
		bucket = bucketState{
			tokens:    float64(l.burstSize),
			timestamp: now,
		}
	}

	// Calculate tokens to add based on time elapsed
	elapsed := now.Sub(bucket.timestamp).Seconds()
	bucket.tokens += elapsed * l.refillRate

	// Cap tokens at burst size
	if bucket.tokens > float64(l.burstSize) {
		bucket.tokens = float64(l.burstSize)
	}

	// Check if we have enough tokens
	if bucket.tokens < 1 {
		// Calculate time until next token is available
		timeToNextToken := time.Duration((1 - bucket.tokens) / l.refillRate * float64(time.Second))
		resetTime := now.Add(timeToNextToken)

		// Update timestamp but not tokens
		bucket.timestamp = now
		l.buckets[key] = bucket

		return false, 0, resetTime
	}

	// Consume one token
	bucket.tokens -= 1
	bucket.timestamp = now
	l.buckets[key] = bucket

	// Calculate remaining tokens
	remaining := int(bucket.tokens)

	// Calculate time until full
	timeToFull := time.Duration((float64(l.burstSize) - bucket.tokens) / l.refillRate * float64(time.Second))
	resetTime := now.Add(timeToFull)

	return true, remaining, resetTime
}

// GetConfig returns the rate limiter configuration
func (l *TokenBucketRateLimiter) GetConfig() *RateLimitConfig {
	return l.config
}

// ExtractKey extracts the rate limiting key from the request
func (l *TokenBucketRateLimiter) ExtractKey(req *http.Request, authData map[string]interface{}) string {
	// Use the same key extraction logic as fixed window
	fixedWindow := &FixedWindowRateLimiter{config: l.config}
	return fixedWindow.ExtractKey(req, authData)
}

// cleanup periodically removes expired buckets
func (l *TokenBucketRateLimiter) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		inactiveThreshold := 24 * time.Hour

		for key, bucket := range l.buckets {
			if now.Sub(bucket.timestamp) > inactiveThreshold {
				delete(l.buckets, key)
			}
		}
		l.mu.Unlock()
	}
}
