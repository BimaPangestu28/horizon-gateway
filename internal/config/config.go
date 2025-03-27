package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultServerPort   = 8080
	DefaultAdminPort    = 8081
	DefaultReadTimeout  = 30 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultIdleTimeout  = 120 * time.Second
)

// Regular expression to match environment variables in the form of ${VAR} or $VAR
var envVarPattern = regexp.MustCompile(`\${([a-zA-Z0-9_]+)}|\$([a-zA-Z0-9_]+)`)

type Config struct {
	Server    ServerConfig  `yaml:"server"`
	Routes    []RouteConfig `yaml:"routes"`
	SSLConfig *SSLConfig    `yaml:"ssl,omitempty"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	AdminPort    int           `yaml:"admin_port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	TLS          *TLSConfig    `yaml:"tls,omitempty"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type SSLConfig struct {
	Enabled bool           `yaml:"enabled"`
	Certbot *CertbotConfig `yaml:"certbot,omitempty"`
	Static  *TLSConfig     `yaml:"static,omitempty"`
}

type CertbotConfig struct {
	Enabled       bool          `yaml:"enabled"`
	Domains       []string      `yaml:"domains"`
	Email         string        `yaml:"email"`
	CertPath      string        `yaml:"cert_path"`
	RenewInterval time.Duration `yaml:"renew_interval"`
	CertbotBin    string        `yaml:"certbot_bin"`
	Staging       bool          `yaml:"staging"`
	WebRootPath   string        `yaml:"webroot_path"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	data = substituteEnvVars(data)

	// First check if this is a versioned config
	var versionedConfig struct {
		Version string  `yaml:"version"`
		Config  *Config `yaml:"config"`
	}

	if err := yaml.Unmarshal(data, &versionedConfig); err == nil && versionedConfig.Version != "" && versionedConfig.Config != nil {
		config := versionedConfig.Config
		applyDefaults(config)
		if err := validateConfig(config); err != nil {
			return nil, fmt.Errorf("validating config: %w", err)
		}
		return config, nil
	}

	// Fall back to standard config
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	applyDefaults(&config)
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &config, nil
}

func substituteEnvVars(data []byte) []byte {
	content := string(data)

	matches := envVarPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		// match[0] is the full match
		// match[1] or match[2] is the variable name (depending on which pattern matched)
		varName := match[1]
		if varName == "" {
			varName = match[2]
		}

		value := os.Getenv(varName)
		content = strings.Replace(content, match[0], value, -1)
	}

	return []byte(content)
}

func applyDefaults(config *Config) {
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

	for i := range config.Routes {
		if len(config.Routes[i].Methods) == 0 {
			config.Routes[i].Methods = []string{"*"}
		}
	}

	// If we have SSL config but no TLS config, create it
	if config.SSLConfig != nil && config.SSLConfig.Enabled {
		if config.Server.TLS == nil {
			config.Server.TLS = &TLSConfig{
				Enabled: true,
			}
		}

		// If we have static certificate config, use it
		if config.SSLConfig.Static != nil {
			config.Server.TLS.CertFile = config.SSLConfig.Static.CertFile
			config.Server.TLS.KeyFile = config.SSLConfig.Static.KeyFile
		}
	}
}

func validateConfig(config *Config) error {
	if config.Server.Port < 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Server.AdminPort < 0 || config.Server.AdminPort > 65535 {
		return fmt.Errorf("invalid admin port: %d", config.Server.AdminPort)
	}

	if config.Server.Port == config.Server.AdminPort {
		return fmt.Errorf("server port and admin port must be different")
	}

	if config.Server.TLS != nil && config.Server.TLS.Enabled {
		if config.SSLConfig != nil && config.SSLConfig.Certbot != nil && config.SSLConfig.Certbot.Enabled {
			// Certbot will handle certificates, so no validation needed for cert files
		} else if config.Server.TLS.CertFile == "" {
			return fmt.Errorf("TLS certificate file path is required when TLS is enabled")
		} else if config.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file path is required when TLS is enabled")
		}
	}

	routeNames := make(map[string]bool)
	routePaths := make(map[string]bool)

	for _, route := range config.Routes {
		if route.Name == "" {
			return fmt.Errorf("route name is required")
		}

		if route.ListenPath == "" {
			return fmt.Errorf("route listen path is required")
		}

		if route.UpstreamURL == "" {
			return fmt.Errorf("route upstream URL is required")
		}

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
