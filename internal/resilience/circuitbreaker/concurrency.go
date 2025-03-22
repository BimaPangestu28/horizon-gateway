package circuitbreaker

import (
	"sync"
	"sync/atomic"
	"time"
)

// ConcurrencyBreaker implements circuit breaking based on concurrency limits
type ConcurrencyBreaker struct {
	config          *CircuitBreakerConfig
	currentCount    int32
	mu              sync.RWMutex
	state           CircuitState
	lastStateChange time.Time
}

// NewConcurrencyBreaker creates a new concurrency circuit breaker
func NewConcurrencyBreaker(config *CircuitBreakerConfig) (*ConcurrencyBreaker, error) {
	return &ConcurrencyBreaker{
		config:          config,
		currentCount:    0,
		state:           CircuitStateClosed,
		lastStateChange: time.Now(),
	}, nil
}

// IsAllowed checks if a request should be allowed through
func (b *ConcurrencyBreaker) IsAllowed() bool {
	// Quick check without locking
	if atomic.LoadInt32(&b.currentCount) >= int32(b.config.ConcurrencyLimit) {
		b.mu.Lock()
		defer b.mu.Unlock()

		// Double-check after acquiring lock
		if b.state == CircuitStateClosed && atomic.LoadInt32(&b.currentCount) >= int32(b.config.ConcurrencyLimit) {
			b.state = CircuitStateOpen
			b.lastStateChange = time.Now()
		}

		return b.state == CircuitStateClosed
	}

	return true
}

// RecordSuccess records a successful request (no-op for concurrency breaker)
func (b *ConcurrencyBreaker) RecordSuccess() {
	// No-op for concurrency breaker
}

// RecordFailure records a failed request (no-op for concurrency breaker)
func (b *ConcurrencyBreaker) RecordFailure() {
	// No-op for concurrency breaker
}

// RecordTimeout records a request timeout (no-op for concurrency breaker)
func (b *ConcurrencyBreaker) RecordTimeout() {
	// No-op for concurrency breaker
}

// RecordRequestStarted increments the concurrency counter
func (b *ConcurrencyBreaker) RecordRequestStarted() {
	atomic.AddInt32(&b.currentCount, 1)
}

// RecordRequestCompleted decrements the concurrency counter
func (b *ConcurrencyBreaker) RecordRequestCompleted() {
	newCount := atomic.AddInt32(&b.currentCount, -1)

	// If count drops below limit and circuit is open, consider closing it
	if newCount < int32(b.config.ConcurrencyLimit) && b.GetState() == CircuitStateOpen {
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.state == CircuitStateOpen {
			b.state = CircuitStateClosed
			b.lastStateChange = time.Now()
		}
	}
}

// GetState returns the current state of the circuit breaker
func (b *ConcurrencyBreaker) GetState() CircuitState {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.state
}

// GetConfig returns the circuit breaker configuration
func (b *ConcurrencyBreaker) GetConfig() *CircuitBreakerConfig {
	return b.config
}
