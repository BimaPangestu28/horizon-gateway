package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bimapangestu28/horizon/internal/cache"
	"github.com/bimapangestu28/horizon/internal/resilience/circuitbreaker"
	"github.com/bimapangestu28/horizon/internal/security/auth"
	"github.com/bimapangestu28/horizon/internal/security/ipfilter"
	"github.com/bimapangestu28/horizon/internal/security/ratelimit"
	"github.com/bimapangestu28/horizon/internal/transform"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultServerPort is the default port for the main server
	DefaultServerPort = 8080

	// DefaultAdminPort is the default port for the admin server
	DefaultAdminPort = 8081

	// DefaultReadTimeout is the default timeout for reading request
	DefaultReadTimeout = 30 * time.Second

	// DefaultWriteTimeout is the default timeout for writing response
	DefaultWriteTimeout = 30 * time.Second

	// DefaultIdleTimeout is the default timeout for idle connections
	DefaultIdleTimeout = 120 * time.Second
)

// Config represents the application configuration
type Config struct {
	// Server contains HTTP server configuration options
	Server ServerConfig `yaml:"server"`

	// Routes defines the API Gateway routing rules
	Routes []RouteConfig `yaml:"routes"`
}

// ServerConfig defines HTTP server settings
type ServerConfig struct {
	// Port is the port number for the main HTTP server
	Port int `yaml:"port"`

	// AdminPort is the port number for the admin API server
	AdminPort int `yaml:"admin_port"`

	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration `yaml:"read_timeout"`

	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration `yaml:"write_timeout"`

	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration `yaml:"idle_timeout"`

	// TLS configuration for HTTPS
	TLS *TLSConfig `yaml:"tls,omitempty"`
}

// TLSConfig defines TLS/SSL configuration
type TLSConfig struct {
	// Enabled indicates whether TLS is enabled
	Enabled bool `yaml:"enabled"`

	// CertFile is the path to the TLS certificate file
	CertFile string `yaml:"cert_file"`

	// KeyFile is the path to the TLS key file
	KeyFile string `yaml:"key_file"`
}

// LoadBalancingConfig defines load balancing settings for a route
type LoadBalancingConfig struct {
	// Type defines the load balancing algorithm
	Type string `yaml:"type"`

	// Targets is a list of upstream targets for this route
	Targets []TargetConfig `yaml:"targets,omitempty"`

	// HealthCheck defines health check settings for targets
	HealthCheck *HealthCheckConfig `yaml:"health_check,omitempty"`
}

// TargetConfig defines an upstream target for load balancing
type TargetConfig struct {
	// URL is the upstream service URL
	URL string `yaml:"url"`

	// Weight is used for weighted load balancing
	Weight int `yaml:"weight,omitempty"`
}

// HealthCheckConfig defines health check settings
type HealthCheckConfig struct {
	// Path is the endpoint path to check on the target
	Path string `yaml:"path"`

	// Interval is how often to check the target
	Interval string `yaml:"interval"`

	// Timeout is how long to wait for a response
	Timeout string `yaml:"timeout"`

	// HealthyThreshold is the number of consecutive successes required
	HealthyThreshold int `yaml:"healthy_threshold"`

	// UnhealthyThreshold is the number of consecutive failures required
	UnhealthyThreshold int `yaml:"unhealthy_threshold"`
}

// RouteConfig defines a single route mapping in the gateway
type RouteConfig struct {
	// Name is a unique identifier for the route
	Name string `yaml:"name"`

	// ListenPath is the path pattern to match for incoming requests
	ListenPath string `yaml:"listen_path"`

	// UpstreamURL is the target URL for proxying requests
	UpstreamURL string `yaml:"upstream_url"`

	// Methods is a list of HTTP methods this route allows
	Methods []string `yaml:"methods"`

	// StripPath indicates whether to strip the matched path prefix before forwarding
	StripPath bool `yaml:"strip_path"`

	// Headers defines required headers for matching this route
	Headers map[string]string `yaml:"headers,omitempty"`

	// QueryParams defines required query parameters for matching this route
	QueryParams map[string]string `yaml:"query_params,omitempty"`

	// Host is the hostname to match for this route
	Host string `yaml:"host,omitempty"`

	// Priority defines the route precedence when multiple routes match
	Priority int `yaml:"priority,omitempty"`

	// LoadBalancing defines load balancing settings for this route
	LoadBalancing *LoadBalancingConfig `yaml:"load_balancing,omitempty"`

	// IPFilter defines IP filtering settings for this route
	IPFilter *ipfilter.IPFilterConfig `yaml:"ip_filter,omitempty"`

	// Auth defines authentication settings for this route
	Auth *auth.AuthConfig `yaml:"auth,omitempty"`

	// RateLimiting defines rate limiting settings for this route
	RateLimiting *ratelimit.RateLimitConfig `yaml:"rate_limiting,omitempty"`

	// CircuitBreaker defines circuit breaking settings for this route
	CircuitBreaker *circuitbreaker.CircuitBreakerConfig `yaml:"circuit_breaker,omitempty"`

	// Caching defines response caching settings for this route
	Caching *cache.CacheConfig `yaml:"caching,omitempty"`

	// Transform defines request/response transformation settings for this route
	Transform *transform.TransformConfig `yaml:"transform,omitempty"`
}

// LoadConfig reads configuration from the specified file and environment variables
// It returns a parsed Config struct or an error if loading fails
func LoadConfig(path string) (*Config, error) {
	// Read configuration file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Substitute environment variables
	data = substituteEnvVars(data)

	// First try to parse as a versioned config
	var versionedConfig VersionedConfig
	err = yaml.Unmarshal(data, &versionedConfig)

	// Check if this is a versioned config
	if err == nil && versionedConfig.Version != "" && versionedConfig.Config != nil {
		// Apply migration if needed
		config, err := MigrateConfig(versionedConfig.Config, ConfigVersion(versionedConfig.Version))
		if err != nil {
			return nil, fmt.Errorf("migrating config: %w", err)
		}

		// Apply default values
		applyDefaults(config)

		// Validate configuration
		if err := validateConfig(config); err != nil {
			return nil, fmt.Errorf("validating config: %w", err)
		}

		return config, nil
	}

	// Fall back to parsing as a non-versioned config
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Apply default values
	applyDefaults(&config)

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &config, nil
}

// substituteEnvVars replaces environment variable references in the configuration
func substituteEnvVars(data []byte) []byte {
	content := string(data)

	// Replace ${VAR} or $VAR with the environment variable value
	for _, match := range envVarPattern.FindAllStringSubmatch(content, -1) {
		varName := match[1]
		if varName == "" {
			varName = match[2]
		}

		// Get environment variable value, use empty string if not found
		value := os.Getenv(varName)

		// Replace in the content
		content = strings.Replace(content, match[0], value, -1)
	}

	return []byte(content)
}

// applyDefaults applies default values to the configuration
func applyDefaults(config *Config) {
	// Server defaults
	if config.Server.Port == 0 {
		config.Server.Port = DefaultServerPort
	}

	if config.Server.AdminPort == 0 {
		config.Server.AdminPort = DefaultAdminPort
	}

	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = DefaultReadTimeout
	}

	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = DefaultWriteTimeout
	}

	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = DefaultIdleTimeout
	}

	// Route defaults
	for i := range config.Routes {
		// Default to all methods if none specified
		if len(config.Routes[i].Methods) == 0 {
			config.Routes[i].Methods = []string{"*"}
		}
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate server configuration
	if config.Server.Port < 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Server.AdminPort < 0 || config.Server.AdminPort > 65535 {
		return fmt.Errorf("invalid admin port: %d", config.Server.AdminPort)
	}

	if config.Server.Port == config.Server.AdminPort {
		return fmt.Errorf("server port and admin port must be different")
	}

	// Validate TLS configuration
	if config.Server.TLS != nil && config.Server.TLS.Enabled {
		if config.Server.TLS.CertFile == "" {
			return fmt.Errorf("TLS certificate file path is required when TLS is enabled")
		}

		if config.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file path is required when TLS is enabled")
		}
	}

	// Validate routes
	routeNames := make(map[string]bool)
	routePaths := make(map[string]bool)

	for _, route := range config.Routes {
		// Check required fields
		if route.Name == "" {
			return fmt.Errorf("route name is required")
		}

		if route.ListenPath == "" {
			return fmt.Errorf("route listen path is required")
		}

		if route.UpstreamURL == "" {
			return fmt.Errorf("route upstream URL is required")
		}

		// Check for duplicates
		if routeNames[route.Name] {
			return fmt.Errorf("duplicate route name: %s", route.Name)
		}
		routeNames[route.Name] = true

		if routePaths[route.ListenPath] {
			return fmt.Errorf("duplicate route listen path: %s", route.ListenPath)
		}
		routePaths[route.ListenPath] = true
	}

	return nil
}
