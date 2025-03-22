package circuitbreaker

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half_open"
)

// BreakerType represents the type of circuit breaking to use
type BreakerType string

const (
	BreakerTypeError       BreakerType = "error"
	BreakerTypeTimeout     BreakerType = "timeout"
	BreakerTypeConcurrency BreakerType = "concurrency"
	BreakerTypeHybrid      BreakerType = "hybrid"
)

// CircuitBreakerConfig defines circuit breaking configuration
type CircuitBreakerConfig struct {
	Enabled             bool        `yaml:"enabled"`
	Type                BreakerType `yaml:"type" default:"error"`
	ErrorThreshold      int         `yaml:"error_threshold" default:"50"`
	TimeoutThreshold    int         `yaml:"timeout_threshold" default:"50"`
	MinimumRequests     int         `yaml:"minimum_requests" default:"20"`
	SamplingWindow      string      `yaml:"sampling_window" default:"1m"`
	OpenStateDuration   string      `yaml:"open_state_duration" default:"30s"`
	HalfOpenMaxRequests int         `yaml:"half_open_max_requests" default:"5"`
	ConcurrencyLimit    int         `yaml:"concurrency_limit" default:"100"`
	ResponseStatusCode  int         `yaml:"response_status_code" default:"503"`
	ResponseBody        string      `yaml:"response_body" default:"Service temporarily unavailable"`
	FailureStatusCodes  []int       `yaml:"failure_status_codes"`
}

// CircuitBreaker interface defines circuit breaking functionality
type CircuitBreaker interface {
	IsAllowed() bool
	RecordSuccess()
	RecordFailure()
	RecordTimeout()
	RecordRequestStarted()
	RecordRequestCompleted()
	GetState() CircuitState
	GetConfig() *CircuitBreakerConfig
}
