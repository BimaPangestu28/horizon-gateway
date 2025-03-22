package plugins

import (
	"context"
	"fmt"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
)

type PluginManager struct {
	registry       *PluginRegistry
	loader         *PluginLoader
	logger         logging.Logger
	pluginDirs     []string
	enabledPlugins map[string]bool
	pluginConfigs  map[string]map[string]interface{}
	mu             sync.RWMutex
	initialized    bool
}

func NewPluginManager(logger logging.Logger) *PluginManager {
	registry := NewPluginRegistry(logger)
	loader := NewPluginLoader(registry, logger)

	return &PluginManager{
		registry:       registry,
		loader:         loader,
		logger:         logger,
		pluginDirs:     []string{"./plugins"},
		enabledPlugins: make(map[string]bool),
		pluginConfigs:  make(map[string]map[string]interface{}),
		initialized:    false,
	}
}

func (m *PluginManager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	for _, dir := range m.pluginDirs {
		plugins, err := m.loader.LoadDirectory(dir)
		if err != nil {
			m.logger.Warn("Error loading plugins from directory", "dir", dir, "error", err)
			continue
		}

		for _, p := range plugins {
			config := m.getPluginConfig(p.Name())
			if err := m.loader.InitializePlugin(p, config); err != nil {
				m.logger.Error("Failed to initialize plugin", "name", p.Name(), "error", err)
				continue
			}

			if err := m.loader.RegisterPlugin(p); err != nil {
				m.logger.Error("Failed to register plugin", "name", p.Name(), "error", err)
				continue
			}

			if m.isPluginEnabled(p.Name()) {
				m.enabledPlugins[p.Name()] = true
			}
		}
	}

	if err := m.registry.MapToPhases(); err != nil {
		m.logger.Error("Failed to map plugins to phases", "error", err)
		return err
	}

	m.initialized = true
	return nil
}

func (m *PluginManager) SetPluginDirectories(dirs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pluginDirs = dirs
}

func (m *PluginManager) AddPluginDirectory(dir string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pluginDirs = append(m.pluginDirs, dir)
}

func (m *PluginManager) SetPluginConfig(name string, config map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pluginConfigs[name] = config
}

func (m *PluginManager) SetPluginEnabled(name string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.registry.Get(name)
	if !exists {
		return fmt.Errorf("%w: %s", ErrPluginNotFound, name)
	}

	m.enabledPlugins[name] = enabled

	basePlugin, ok := plugin.(*BasePlugin)
	if ok {
		basePlugin.SetEnabled(enabled)
	}

	return nil
}

func (m *PluginManager) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	return m.registry.Execute(phase, data)
}

func (m *PluginManager) GetPlugin(name string) (interfaces.Plugin, bool) {
	return m.registry.Get(name)
}

func (m *PluginManager) GetAllPlugins() []interfaces.Plugin {
	return m.registry.GetAll()
}

func (m *PluginManager) GetPluginsByPhase(phase interfaces.PluginPhase) []interfaces.Plugin {
	return m.registry.GetByPhase(phase)
}

func (m *PluginManager) Shutdown() error {
	return m.registry.ShutdownAll()
}

func (m *PluginManager) getPluginConfig(name string) map[string]interface{} {
	config, exists := m.pluginConfigs[name]
	if !exists {
		return nil
	}
	return config
}

func (m *PluginManager) isPluginEnabled(name string) bool {
	enabled, exists := m.enabledPlugins[name]
	if !exists {
		return true
	}
	return enabled
}
