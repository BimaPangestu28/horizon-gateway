package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// SlidingWindowRateLimiter implements sliding window rate limiting
type SlidingWindowRateLimiter struct {
	config     *RateLimitConfig
	prevWindow map[string]int
	currWindow map[string]int
	lastRotate time.Time
	mu         sync.RWMutex
	window     time.Duration
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter
func NewSlidingWindowRateLimiter(config *RateLimitConfig) (*SlidingWindowRateLimiter, error) {
	window, err := time.ParseDuration(config.Window)
	if err != nil {
		return nil, err
	}

	limiter := &SlidingWindowRateLimiter{
		config:     config,
		prevWindow: make(map[string]int),
		currWindow: make(map[string]int),
		lastRotate: time.Now(),
		window:     window,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter, nil
}

// Allow checks if a request is allowed based on rate limiting rules
func (l *SlidingWindowRateLimiter) Allow(key string) (bool, int, time.Time) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Rotate windows if needed
	if now.Sub(l.lastRotate) >= l.window {
		l.prevWindow = l.currWindow
		l.currWindow = make(map[string]int)
		l.lastRotate = now
	}

	// Calculate the position in the current window (0.0 to 1.0)
	position := float64(now.Sub(l.lastRotate)) / float64(l.window)

	// Calculate the weighted count using both windows
	prevCount := float64(l.prevWindow[key]) * (1 - position)
	currCount := float64(l.currWindow[key])
	weightedCount := int(prevCount + currCount)

	// Check if limit has been reached
	if weightedCount >= l.config.Limit {
		// Calculate time until reset
		timeToReset := l.window - now.Sub(l.lastRotate)
		resetTime := now.Add(timeToReset)
		return false, 0, resetTime
	}

	// Increment counter for current window
	l.currWindow[key]++

	// Recalculate the weighted count
	currCount = float64(l.currWindow[key])
	weightedCount = int(prevCount + currCount)

	// Calculate remaining tokens
	remaining := l.config.Limit - weightedCount

	// Calculate time until reset
	timeToReset := l.window - now.Sub(l.lastRotate)
	resetTime := now.Add(timeToReset)

	return true, remaining, resetTime
}

// GetConfig returns the rate limiter configuration
func (l *SlidingWindowRateLimiter) GetConfig() *RateLimitConfig {
	return l.config
}

// ExtractKey extracts the rate limiting key from the request
func (l *SlidingWindowRateLimiter) ExtractKey(req *http.Request, authData map[string]interface{}) string {
	// Use the same key extraction logic as fixed window
	fixedWindow := &FixedWindowRateLimiter{config: l.config}
	return fixedWindow.ExtractKey(req, authData)
}

// cleanup periodically rotates windows
func (l *SlidingWindowRateLimiter) cleanup() {
	ticker := time.NewTicker(l.window)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		l.prevWindow = l.currWindow
		l.currWindow = make(map[string]int)
		l.lastRotate = time.Now()
		l.mu.Unlock()
	}
}
