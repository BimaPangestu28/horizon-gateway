package ratelimit

import (
	"net/http"
	"time"
)

// RateLimitType represents the type of rate limiting algorithm
type RateLimitType string

const (
	// RateLimitTypeFixedWindow represents fixed window rate limiting
	RateLimitTypeFixedWindow RateLimitType = "fixed_window"

	// RateLimitTypeSlidingWindow represents sliding window rate limiting
	RateLimitTypeSlidingWindow RateLimitType = "sliding_window"

	// RateLimitTypeTokenBucket represents token bucket rate limiting
	RateLimitTypeTokenBucket RateLimitType = "token_bucket"
)

// RateLimitScope represents the scope of rate limiting
type RateLimitScope string

const (
	// RateLimitScopeGlobal applies rate limits across all clients
	RateLimitScopeGlobal RateLimitScope = "global"

	// RateLimitScopeIP applies rate limits per client IP
	RateLimitScopeIP RateLimitScope = "ip"

	// RateLimitScopeHeader applies rate limits based on a header value
	RateLimitScopeHeader RateLimitScope = "header"

	// RateLimitScopeAuth applies rate limits based on authentication
	RateLimitScopeAuth RateLimitScope = "auth"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	// Enabled indicates whether rate limiting is enabled
	Enabled bool `yaml:"enabled"`

	// Type is the rate limiting algorithm to use
	Type RateLimitType `yaml:"type" default:"fixed_window"`

	// Scope is the scope of rate limiting
	Scope RateLimitScope `yaml:"scope" default:"ip"`

	// Header is the header name to use when scope is "header"
	Header string `yaml:"header,omitempty"`

	// Limit is the maximum number of requests allowed in the window
	Limit int `yaml:"limit"`

	// Window is the time window for rate limiting (e.g., "1m", "1h")
	Window string `yaml:"window"`

	// BurstSize is the maximum bucket size for token bucket algorithm
	BurstSize int `yaml:"burst_size,omitempty"`

	// RefillRate is the tokens per second for token bucket algorithm
	RefillRate float64 `yaml:"refill_rate,omitempty"`

	// ResponseStatusCode is the HTTP status code to return when rate limited
	ResponseStatusCode int `yaml:"response_status_code" default:"429"`

	// ResponseBody is the body to return when rate limited
	ResponseBody string `yaml:"response_body" default:"Rate limit exceeded"`

	// EnableHeaders indicates whether to add rate limit headers to responses
	EnableHeaders bool `yaml:"enable_headers" default:"true"`
}

// RateLimiter interface defines rate limiting functionality
type RateLimiter interface {
	// Allow checks if a request is allowed based on rate limiting rules
	// Returns allowed (true/false), remaining count, and reset time
	Allow(key string) (bool, int, time.Time)

	// GetConfig returns the rate limiter configuration
	GetConfig() *RateLimitConfig

	// ExtractKey extracts the rate limiting key from the request
	ExtractKey(req *http.Request, authData map[string]interface{}) string
}
