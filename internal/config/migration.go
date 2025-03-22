package config

import (
	"fmt"
)

// ConfigVersion represents a configuration format version
type ConfigVersion string

const (
	// CurrentVersion is the latest configuration version
	CurrentVersion ConfigVersion = "v1.0.0"

	// InitialVersion is the first version of the configuration
	InitialVersion ConfigVersion = "v0.9.0"
)

// Migration represents a function that migrates from one config version to another
type Migration func(*Config) (*Config, error)

// migrationMap maps from version to a migration function
var migrationMap = map[ConfigVersion]Migration{
	InitialVersion: migrateInitialToCurrent,
	// Add more migrations as new versions are released
}

// MigrateConfig upgrades a configuration to the current version if needed
func MigrateConfig(config *Config, version ConfigVersion) (*Config, error) {
	// If already at current version, no migration needed
	if version == CurrentVersion {
		return config, nil
	}

	// Find migration path
	migration, exists := migrationMap[version]
	if !exists {
		return nil, fmt.Errorf("no migration path from version %s to %s", version, CurrentVersion)
	}

	// Apply migration
	migratedConfig, err := migration(config)
	if err != nil {
		return nil, fmt.Errorf("migration from %s to %s failed: %w", version, CurrentVersion, err)
	}

	return migratedConfig, nil
}

// migrateInitialToCurrent migrates from the initial version to the current version
func migrateInitialToCurrent(config *Config) (*Config, error) {
	// Create a new config with the current structure
	newConfig := &Config{
		Server: config.Server,
		Routes: make([]RouteConfig, len(config.Routes)),
	}

	// Migrate routes
	for i, route := range config.Routes {
		// Copy basic properties
		newConfig.Routes[i] = RouteConfig{
			Name:        route.Name,
			ListenPath:  route.ListenPath,
			UpstreamURL: route.UpstreamURL,
			Methods:     route.Methods,
			StripPath:   route.StripPath,
			Host:        route.Host,
			Headers:     route.Headers,
			QueryParams: route.QueryParams,
			Priority:    route.Priority,
		}

		// Set up load balancing config if not present
		if route.LoadBalancing == nil && route.UpstreamURL != "" {
			newConfig.Routes[i].LoadBalancing = &LoadBalancingConfig{
				Type: "round_robin",
				Targets: []TargetConfig{
					{
						URL:    route.UpstreamURL,
						Weight: 1,
					},
				},
			}
		} else {
			newConfig.Routes[i].LoadBalancing = route.LoadBalancing
		}
	}

	return newConfig, nil
}

// ConvertConfigVersionToString converts a ConfigVersion to a string
func ConvertConfigVersionToString(version ConfigVersion) string {
	return string(version)
}
