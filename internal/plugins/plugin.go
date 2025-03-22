package plugins

import (
	"errors"
	"fmt"
	"path/filepath"
	"plugin"
	"strings"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
)

var (
	ErrPluginNotFound      = errors.New("plugin not found")
	ErrPluginLoadFailed    = errors.New("failed to load plugin")
	ErrNoPluginsFound      = errors.New("no plugins found in directory")
	ErrPluginInitFailed    = errors.New("plugin initialization failed")
	ErrUnsupportedFormat   = errors.New("unsupported plugin format")
	ErrExecutionFailed     = errors.New("plugin execution failed")
	ErrInvalidPluginPath   = errors.New("invalid plugin path")
	ErrPluginSymbolMissing = errors.New("plugin symbol not found")
)

type PluginRegistry struct {
	plugins      map[string]interfaces.Plugin
	phaseMapping map[interfaces.PluginPhase][]interfaces.Plugin
	logger       logging.Logger
	mu           sync.RWMutex
}

func NewPluginRegistry(logger logging.Logger) *PluginRegistry {
	return &PluginRegistry{
		plugins:      make(map[string]interfaces.Plugin),
		phaseMapping: make(map[interfaces.PluginPhase][]interfaces.Plugin),
		logger:       logger,
	}
}

func (r *PluginRegistry) Register(plugin interfaces.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("%w: %s", interfaces.ErrPluginAlreadyExists, name)
	}

	r.plugins[name] = plugin
	r.logger.Info("Plugin registered", "name", name, "version", plugin.Version())

	return nil
}

func (r *PluginRegistry) Get(name string) (interfaces.Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, found := r.plugins[name]
	return plugin, found
}

func (r *PluginRegistry) GetAll() []interfaces.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]interfaces.Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, p)
	}
	return result
}

func (r *PluginRegistry) MapToPhases() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mapping := make(map[interfaces.PluginPhase][]interfaces.Plugin)

	for _, p := range r.plugins {
		assignedPhase := getPluginDefaultPhase(p.Name())
		if plugins, ok := mapping[assignedPhase]; ok {
			mapping[assignedPhase] = append(plugins, p)
		} else {
			mapping[assignedPhase] = []interfaces.Plugin{p}
		}
	}

	r.phaseMapping = mapping
	return nil
}

func (r *PluginRegistry) GetByPhase(phase interfaces.PluginPhase) []interfaces.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if plugins, ok := r.phaseMapping[phase]; ok {
		return plugins
	}
	return []interfaces.Plugin{}
}

func (r *PluginRegistry) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	plugins := r.GetByPhase(phase)
	if len(plugins) == 0 {
		return nil
	}

	for _, p := range plugins {
		if data.AbortChain {
			break
		}

		if err := p.Execute(phase, data); err != nil {
			r.logger.Error("Plugin execution failed",
				"phase", string(phase),
				"plugin", p.Name(),
				"error", err)

			if data.Error == nil {
				data.Error = err
			}
		}
	}

	return nil
}

func (r *PluginRegistry) ShutdownAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, p := range r.plugins {
		if err := p.Shutdown(); err != nil {
			r.logger.Error("Plugin shutdown failed", "name", name, "error", err)
		}
	}

	return nil
}

func (r *PluginRegistry) LoadPlugin(path string) (interfaces.Plugin, error) {
	plug, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPluginLoadFailed, err)
	}

	createSym, err := plug.Lookup("CreatePlugin")
	if err != nil {
		return nil, fmt.Errorf("%w: CreatePlugin", ErrPluginSymbolMissing)
	}

	createFunc, ok := createSym.(func() interfaces.Plugin)
	if !ok {
		return nil, fmt.Errorf("%w: invalid CreatePlugin signature", ErrPluginLoadFailed)
	}

	p := createFunc()
	if p == nil {
		return nil, fmt.Errorf("%w: CreatePlugin returned nil", ErrPluginLoadFailed)
	}

	return p, nil
}

func (r *PluginRegistry) LoadDirectory(dirPath string) ([]interfaces.Plugin, error) {
	matches, err := filepath.Glob(filepath.Join(dirPath, "*.so"))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPluginLoadFailed, err)
	}

	if len(matches) == 0 {
		return nil, ErrNoPluginsFound
	}

	var loadedPlugins []interfaces.Plugin
	for _, path := range matches {
		p, err := r.LoadPlugin(path)
		if err != nil {
			r.logger.Error("Failed to load plugin", "path", path, "error", err)
			continue
		}

		if err := r.Register(p); err != nil {
			r.logger.Error("Failed to register plugin", "name", p.Name(), "error", err)
			continue
		}

		loadedPlugins = append(loadedPlugins, p)
	}

	if len(loadedPlugins) == 0 {
		return nil, ErrNoPluginsFound
	}

	return loadedPlugins, nil
}

func getPluginDefaultPhase(name string) interfaces.PluginPhase {
	switch {
	case containsAny(name, []string{"auth", "security", "rate", "limit", "filter", "cors"}):
		return interfaces.PhaseRequest
	case containsAny(name, []string{"transform", "convert", "modify"}):
		return interfaces.PhaseRequest
	case containsAny(name, []string{"cache", "compress", "encode"}):
		return interfaces.PhaseResponse
	case containsAny(name, []string{"log", "metric", "track", "monitor"}):
		return interfaces.PhaseRequest
	case containsAny(name, []string{"error", "recovery", "fallback"}):
		return interfaces.PhaseError
	default:
		return interfaces.PhaseRequest
	}
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(strings.ToLower(s), strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
