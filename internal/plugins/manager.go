package plugins

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
	"github.com/bimapangestu28/horizon/plugins/loader"
	"github.com/bimapangestu28/horizon/plugins/registry"
)

type Manager struct {
	registry       interfaces.PluginRegistry
	loader         interfaces.PluginLoader
	logger         logging.Logger
	loadedPlugins  map[string]interfaces.Plugin
	pluginDirs     []string
	enabledPlugins map[string]bool
	mu             sync.RWMutex
}

type PluginManagerConfig struct {
	PluginDirectories []string               `yaml:"plugin_directories"`
	EnabledPlugins    []string               `yaml:"enabled_plugins"`
	PluginConfigs     map[string]interface{} `yaml:"plugin_configs"`
}

func NewManager(config PluginManagerConfig, logger logging.Logger) (*Manager, error) {
	pluginRegistry := registry.NewRegistry(logger)
	pluginLoader := loader.NewLoader(pluginRegistry, logger)

	manager := &Manager{
		registry:       pluginRegistry,
		loader:         pluginLoader,
		logger:         logger,
		loadedPlugins:  make(map[string]interfaces.Plugin),
		pluginDirs:     config.PluginDirectories,
		enabledPlugins: make(map[string]bool),
	}

	// Mark enabled plugins
	for _, name := range config.EnabledPlugins {
		manager.enabledPlugins[name] = true
	}

	return manager, nil
}

func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load plugins from all configured directories
	for _, dir := range m.pluginDirs {
		plugins, err := m.loader.LoadDirectory(dir)
		if err != nil {
			m.logger.Error("Failed to load plugins from directory", "dir", dir, "error", err)
			continue
		}

		for _, plugin := range plugins {
			m.loadedPlugins[plugin.Name()] = plugin
			m.logger.Info("Loaded plugin",
				"name", plugin.Name(),
				"version", plugin.Version(),
				"author", plugin.Author())
		}
	}

	// Initialize enabled plugins
	for name, plugin := range m.loadedPlugins {
		if !m.enabledPlugins[name] {
			m.logger.Info("Plugin loaded but not enabled", "name", name)
			continue
		}

		// Initialize plugin
		err := plugin.Init(nil) // TODO: Pass plugin-specific config
		if err != nil {
			m.logger.Error("Failed to initialize plugin", "name", name, "error", err)
			continue
		}

		m.logger.Info("Initialized plugin", "name", name)
	}

	return nil
}

func (m *Manager) LoadPlugin(path string) (interfaces.Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugins, err := m.loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	if len(plugins) == 0 {
		return nil, fmt.Errorf("no plugins found at %s", path)
	}

	plugin := plugins[0]
	m.loadedPlugins[plugin.Name()] = plugin

	// Initialize if enabled
	if m.enabledPlugins[plugin.Name()] {
		err := plugin.Init(nil) // TODO: Pass plugin-specific config
		if err != nil {
			m.logger.Error("Failed to initialize plugin", "name", plugin.Name(), "error", err)
			return nil, fmt.Errorf("failed to initialize plugin: %w", err)
		}
	}

	return plugin, nil
}

func (m *Manager) GetPlugin(name string) (interfaces.Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.loadedPlugins[name]
	return plugin, exists
}

func (m *Manager) GetAllPlugins() []interfaces.Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]interfaces.Plugin, 0, len(m.loadedPlugins))
	for _, plugin := range m.loadedPlugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

func (m *Manager) GetEnabledPlugins() []interfaces.Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]interfaces.Plugin, 0)
	for name, plugin := range m.loadedPlugins {
		if m.enabledPlugins[name] {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

func (m *Manager) EnablePlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.loadedPlugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if m.enabledPlugins[name] {
		return nil // Already enabled
	}

	// Initialize plugin
	err := plugin.Init(nil) // TODO: Pass plugin-specific config
	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	m.enabledPlugins[name] = true
	m.logger.Info("Enabled plugin", "name", name)

	return nil
}

func (m *Manager) DisablePlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.loadedPlugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if !m.enabledPlugins[name] {
		return nil // Already disabled
	}

	// Shutdown plugin
	err := plugin.Shutdown()
	if err != nil {
		m.logger.Error("Error shutting down plugin", "name", name, "error", err)
		// Continue anyway to mark as disabled
	}

	delete(m.enabledPlugins, name)
	m.logger.Info("Disabled plugin", "name", name)

	return nil
}

func (m *Manager) ExecutePhase(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	return m.registry.Execute(phase, data)
}

func (m *Manager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, plugin := range m.loadedPlugins {
		if !m.enabledPlugins[name] {
			continue
		}

		err := plugin.Shutdown()
		if err != nil {
			m.logger.Error("Error shutting down plugin", "name", name, "error", err)
			lastErr = err
		}
	}

	return lastErr
}

func (m *Manager) ReloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.loadedPlugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Get the plugin path
	pluginPath := ""
	for _, dir := range m.pluginDirs {
		// Check if the plugin exists in this directory
		possiblePath := filepath.Join(dir, name+".so")
		// TODO: Check if the file exists
		pluginPath = possiblePath
		break
	}

	if pluginPath == "" {
		return fmt.Errorf("could not find plugin file for %s", name)
	}

	// Shutdown the plugin if enabled
	if m.enabledPlugins[name] {
		err := plugin.Shutdown()
		if err != nil {
			m.logger.Error("Error shutting down plugin for reload", "name", name, "error", err)
			// Continue anyway
		}
	}

	// Reload the plugin
	newPlugins, err := m.loader.Load(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to reload plugin: %w", err)
	}

	if len(newPlugins) == 0 {
		return fmt.Errorf("no plugins found at %s", pluginPath)
	}

	newPlugin := newPlugins[0]
	m.loadedPlugins[name] = newPlugin

	// Initialize if enabled
	if m.enabledPlugins[name] {
		err := newPlugin.Init(nil) // TODO: Pass plugin-specific config
		if err != nil {
			m.logger.Error("Failed to initialize reloaded plugin", "name", name, "error", err)
			return fmt.Errorf("failed to initialize reloaded plugin: %w", err)
		}
	}

	m.logger.Info("Reloaded plugin", "name", name)
	return nil
}

func (m *Manager) GetPluginInfo(name string) (*interfaces.PluginInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.loadedPlugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	info := &interfaces.PluginInfo{
		Name:        plugin.Name(),
		Version:     plugin.Version(),
		Description: plugin.Description(),
		Author:      plugin.Author(),
		Phases:      []string{"request", "proxy", "response", "error"}, // TODO: Get from plugin metadata
		Config: interfaces.ConfigInfo{
			Properties:  make(map[string]interfaces.PropertyInfo),
			Required:    []string{},
			Description: "Plugin configuration",
		},
	}

	return info, nil
}

func (m *Manager) IsPluginEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.enabledPlugins[name]
}
