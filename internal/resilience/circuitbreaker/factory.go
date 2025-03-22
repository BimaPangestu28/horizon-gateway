package circuitbreaker

import (
	"fmt"
)

// NewCircuitBreaker creates a circuit breaker based on configuration
func NewCircuitBreaker(config *CircuitBreakerConfig) (CircuitBreaker, error) {
	if !config.Enabled {
		return nil, nil
	}

	switch config.Type {
	case BreakerTypeError:
		return NewErrorRateBreaker(config)

	case BreakerTypeTimeout:
		timeoutConfig := *config
		timeoutConfig.ErrorThreshold = config.TimeoutThreshold
		return NewErrorRateBreaker(&timeoutConfig)

	case BreakerTypeConcurrency:
		return NewConcurrencyBreaker(config)

	case BreakerTypeHybrid:
		return NewHybridBreaker(config)

	default:
		return nil, fmt.Errorf("unsupported circuit breaker type: %s", config.Type)
	}
}
