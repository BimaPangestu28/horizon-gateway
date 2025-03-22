package circuitbreaker

import (
	"sync"
	"time"
)

// Request represents a recorded request with timestamp
type request struct {
	timestamp time.Time
	success   bool
}

// ErrorRateBreaker implements circuit breaking based on error rate
type ErrorRateBreaker struct {
	config               *CircuitBreakerConfig
	state                CircuitState
	requests             []request
	mu                   sync.RWMutex
	lastStateChange      time.Time
	halfOpenSuccessCount int
	halfOpenFailureCount int
	samplingWindow       time.Duration
	openStateDuration    time.Duration
}

// NewErrorRateBreaker creates a new error rate circuit breaker
func NewErrorRateBreaker(config *CircuitBreakerConfig) (*ErrorRateBreaker, error) {
	// Parse durations
	samplingWindow, err := time.ParseDuration(config.SamplingWindow)
	if err != nil {
		return nil, err
	}

	openStateDuration, err := time.ParseDuration(config.OpenStateDuration)
	if err != nil {
		return nil, err
	}

	// Initialize circuit breaker
	breaker := &ErrorRateBreaker{
		config:            config,
		state:             CircuitStateClosed,
		requests:          make([]request, 0, config.MinimumRequests*2),
		lastStateChange:   time.Now(),
		samplingWindow:    samplingWindow,
		openStateDuration: openStateDuration,
	}

	// Start cleanup goroutine
	go breaker.cleanup()

	return breaker, nil
}

// IsAllowed checks if a request should be allowed through
func (b *ErrorRateBreaker) IsAllowed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	switch b.state {
	case CircuitStateClosed:
		return true

	case CircuitStateOpen:
		// Check if open state duration has elapsed
		if time.Since(b.lastStateChange) > b.openStateDuration {
			go b.transitionToHalfOpen()
		}
		return false

	case CircuitStateHalfOpen:
		// Only allow a limited number of requests through
		totalRequests := b.halfOpenSuccessCount + b.halfOpenFailureCount
		return totalRequests < b.config.HalfOpenMaxRequests

	default:
		return true
	}
}

// RecordSuccess records a successful request
func (b *ErrorRateBreaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Record the request
	b.requests = append(b.requests, request{
		timestamp: time.Now(),
		success:   true,
	})

	// Special handling for half-open state
	if b.state == CircuitStateHalfOpen {
		b.halfOpenSuccessCount++

		// Check if we should close the circuit
		if b.halfOpenSuccessCount >= b.config.HalfOpenMaxRequests {
			b.transitionToClosed()
		}

		return
	}

	// Process metrics for closed state
	b.processMetrics()
}

// RecordFailure records a failed request
func (b *ErrorRateBreaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Record the request
	b.requests = append(b.requests, request{
		timestamp: time.Now(),
		success:   false,
	})

	// Special handling for half-open state
	if b.state == CircuitStateHalfOpen {
		b.halfOpenFailureCount++

		// A single failure in half-open should reopen the circuit
		b.transitionToOpen()

		return
	}

	// Process metrics for closed state
	b.processMetrics()
}

// RecordTimeout records a request timeout (same as failure for error rate breaker)
func (b *ErrorRateBreaker) RecordTimeout() {
	b.RecordFailure()
}

// RecordRequestStarted is a no-op for error rate breaker
func (b *ErrorRateBreaker) RecordRequestStarted() {
	// No-op for error rate breaker
}

// RecordRequestCompleted is a no-op for error rate breaker
func (b *ErrorRateBreaker) RecordRequestCompleted() {
	// No-op for error rate breaker
}

// GetState returns the current state of the circuit breaker
func (b *ErrorRateBreaker) GetState() CircuitState {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.state
}

// GetConfig returns the circuit breaker configuration
func (b *ErrorRateBreaker) GetConfig() *CircuitBreakerConfig {
	return b.config
}

// processMetrics calculates error rate and potentially opens the circuit
func (b *ErrorRateBreaker) processMetrics() {
	// Only process in closed state with enough requests
	if b.state != CircuitStateClosed || len(b.requests) < b.config.MinimumRequests {
		return
	}

	// Get relevant requests within sampling window
	now := time.Now()
	windowStart := now.Add(-b.samplingWindow)

	var windowRequests []request
	for _, req := range b.requests {
		if req.timestamp.After(windowStart) {
			windowRequests = append(windowRequests, req)
		}
	}

	// Don't open circuit if we don't have enough samples
	if len(windowRequests) < b.config.MinimumRequests {
		return
	}

	// Calculate error rate
	var failureCount int
	for _, req := range windowRequests {
		if !req.success {
			failureCount++
		}
	}

	errorRate := (failureCount * 100) / len(windowRequests)

	// Check if error threshold is exceeded
	if errorRate >= b.config.ErrorThreshold {
		b.transitionToOpen()
	}
}

// transitionToOpen changes the circuit state to open
func (b *ErrorRateBreaker) transitionToOpen() {
	b.state = CircuitStateOpen
	b.lastStateChange = time.Now()
	b.halfOpenSuccessCount = 0
	b.halfOpenFailureCount = 0
}

// transitionToHalfOpen changes the circuit state to half-open
func (b *ErrorRateBreaker) transitionToHalfOpen() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Only transition if we're still in open state
	if b.state == CircuitStateOpen {
		b.state = CircuitStateHalfOpen
		b.lastStateChange = time.Now()
		b.halfOpenSuccessCount = 0
		b.halfOpenFailureCount = 0
	}
}

// transitionToClosed changes the circuit state to closed
func (b *ErrorRateBreaker) transitionToClosed() {
	b.state = CircuitStateClosed
	b.lastStateChange = time.Now()
	b.halfOpenSuccessCount = 0
	b.halfOpenFailureCount = 0

	// Clear request history
	b.requests = make([]request, 0, b.config.MinimumRequests*2)
}

// cleanup periodically prunes old requests
func (b *ErrorRateBreaker) cleanup() {
	ticker := time.NewTicker(b.samplingWindow / 2)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()

		// Only clean up if we have a significant number of requests
		if len(b.requests) > b.config.MinimumRequests*2 {
			now := time.Now()
			windowStart := now.Add(-b.samplingWindow)

			// Find the index of the first request within the window
			index := 0
			for i, req := range b.requests {
				if req.timestamp.After(windowStart) {
					index = i
					break
				}
			}

			// Prune requests before the window
			if index > 0 {
				b.requests = b.requests[index:]
			}
		}

		b.mu.Unlock()
	}
}
