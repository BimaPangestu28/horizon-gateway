package circuitbreaker

import (
	"sync"
)

// HybridBreaker combines multiple circuit breaking strategies
type HybridBreaker struct {
	config             *CircuitBreakerConfig
	errorBreaker       *ErrorRateBreaker
	timeoutBreaker     *ErrorRateBreaker
	concurrencyBreaker *ConcurrencyBreaker
	mu                 sync.RWMutex
}

// NewHybridBreaker creates a new hybrid circuit breaker
func NewHybridBreaker(config *CircuitBreakerConfig) (*HybridBreaker, error) {
	// Create error rate breaker
	errorBreaker, err := NewErrorRateBreaker(config)
	if err != nil {
		return nil, err
	}

	// Create timeout breaker with modified config
	timeoutConfig := *config
	timeoutConfig.ErrorThreshold = config.TimeoutThreshold
	timeoutBreaker, err := NewErrorRateBreaker(&timeoutConfig)
	if err != nil {
		return nil, err
	}

	// Create concurrency breaker
	concurrencyBreaker, err := NewConcurrencyBreaker(config)
	if err != nil {
		return nil, err
	}

	return &HybridBreaker{
		config:             config,
		errorBreaker:       errorBreaker,
		timeoutBreaker:     timeoutBreaker,
		concurrencyBreaker: concurrencyBreaker,
	}, nil
}

// IsAllowed checks if a request should be allowed through
func (b *HybridBreaker) IsAllowed() bool {
	// Check all breakers - request is allowed only if all breakers allow it
	return b.errorBreaker.IsAllowed() &&
		b.timeoutBreaker.IsAllowed() &&
		b.concurrencyBreaker.IsAllowed()
}

// RecordSuccess records a successful request
func (b *HybridBreaker) RecordSuccess() {
	b.errorBreaker.RecordSuccess()
	b.timeoutBreaker.RecordSuccess()
	b.concurrencyBreaker.RecordSuccess()
}

// RecordFailure records a failed request
func (b *HybridBreaker) RecordFailure() {
	b.errorBreaker.RecordFailure()
}

// RecordTimeout records a request timeout
func (b *HybridBreaker) RecordTimeout() {
	b.timeoutBreaker.RecordFailure()
}

// RecordRequestStarted records the start of a request
func (b *HybridBreaker) RecordRequestStarted() {
	b.concurrencyBreaker.RecordRequestStarted()
}

// RecordRequestCompleted records the completion of a request
func (b *HybridBreaker) RecordRequestCompleted() {
	b.concurrencyBreaker.RecordRequestCompleted()
}

// GetState returns the current state of the circuit breaker
func (b *HybridBreaker) GetState() CircuitState {
	// Return the most restrictive state
	if b.errorBreaker.GetState() == CircuitStateOpen ||
		b.timeoutBreaker.GetState() == CircuitStateOpen ||
		b.concurrencyBreaker.GetState() == CircuitStateOpen {
		return CircuitStateOpen
	}

	if b.errorBreaker.GetState() == CircuitStateHalfOpen ||
		b.timeoutBreaker.GetState() == CircuitStateHalfOpen {
		return CircuitStateHalfOpen
	}

	return CircuitStateClosed
}

// GetConfig returns the circuit breaker configuration
func (b *HybridBreaker) GetConfig() *CircuitBreakerConfig {
	return b.config
}

// Reset resets all circuit breakers to their initial state
func (b *HybridBreaker) Reset() {
	// No built-in reset function for the individual breakers,
	// but we could add one if needed and call it here
	b.mu.Lock()
	defer b.mu.Unlock()

	// For now, we'll just create new breakers
	newErrorBreaker, _ := NewErrorRateBreaker(b.config)
	b.errorBreaker = newErrorBreaker

	timeoutConfig := *b.config
	timeoutConfig.ErrorThreshold = b.config.TimeoutThreshold
	newTimeoutBreaker, _ := NewErrorRateBreaker(&timeoutConfig)
	b.timeoutBreaker = newTimeoutBreaker

	newConcurrencyBreaker, _ := NewConcurrencyBreaker(b.config)
	b.concurrencyBreaker = newConcurrencyBreaker
}
