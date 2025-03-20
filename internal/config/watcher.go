package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// ConfigWatcher watches a configuration file for changes and triggers reloading
type ConfigWatcher struct {
	configPath string
	logger     logging.Logger
	watcher    *fsnotify.Watcher
	callbacks  []func(*Config) error
	config     *Config
	mu         sync.RWMutex
	stopCh     chan struct{}
	debounce   time.Duration
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(configPath string, logger logging.Logger) (*ConfigWatcher, error) {
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	// Load initial configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("loading initial configuration: %w", err)
	}

	// Create config watcher
	cw := &ConfigWatcher{
		configPath: configPath,
		logger:     logger,
		watcher:    watcher,
		callbacks:  make([]func(*Config) error, 0),
		config:     config,
		stopCh:     make(chan struct{}),
		debounce:   500 * time.Millisecond, // Debounce interval for multiple change events
	}

	return cw, nil
}

// Start begins watching for configuration file changes
func (cw *ConfigWatcher) Start() error {
	// Add the configuration file to the watcher
	if err := cw.watcher.Add(cw.configPath); err != nil {
		return fmt.Errorf("watching config file: %w", err)
	}

	// Start the watcher goroutine
	go cw.watchConfig()

	cw.logger.Info("Configuration watcher started", "path", cw.configPath)
	return nil
}

// Stop stops watching for configuration file changes
func (cw *ConfigWatcher) Stop() {
	close(cw.stopCh)
	cw.watcher.Close()
	cw.logger.Info("Configuration watcher stopped")
}

// GetConfig returns the current configuration
func (cw *ConfigWatcher) GetConfig() *Config {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	return cw.config
}

// RegisterCallback registers a function to be called when the configuration changes
func (cw *ConfigWatcher) RegisterCallback(callback func(*Config) error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	cw.callbacks = append(cw.callbacks, callback)
}

// watchConfig watches for file changes and triggers reload
func (cw *ConfigWatcher) watchConfig() {
	// Use timer for debouncing
	var timer *time.Timer

	for {
		select {
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Check if the event is a write/create event for our config file
			if (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) &&
				event.Name == cw.configPath {

				// Debounce multiple events
				if timer != nil {
					timer.Stop()
				}

				timer = time.AfterFunc(cw.debounce, func() {
					cw.reloadConfig()
				})
			}

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}

			cw.logger.Error("Config watcher error", "error", err)

		case <-cw.stopCh:
			if timer != nil {
				timer.Stop()
			}
			return
		}
	}
}

// reloadConfig reloads the configuration file and notifies callbacks
func (cw *ConfigWatcher) reloadConfig() {
	cw.logger.Info("Configuration file changed, reloading...")

	// Check if file exists
	_, err := os.Stat(cw.configPath)
	if err != nil {
		cw.logger.Error("Config file access error", "error", err)
		return
	}

	// Load new configuration
	newConfig, err := LoadConfig(cw.configPath)
	if err != nil {
		cw.logger.Error("Failed to reload configuration", "error", err)
		return
	}

	// Update the configuration
	cw.mu.Lock()
	cw.config = newConfig
	callbacks := append([]func(*Config) error{}, cw.callbacks...) // Create a copy for safe iteration
	cw.mu.Unlock()

	// Notify callbacks
	for i, callback := range callbacks {
		if err := callback(newConfig); err != nil {
			cw.logger.Error("Error in config change callback", "callback", i, "error", err)
		}
	}

	cw.logger.Info("Configuration reloaded successfully")
}
