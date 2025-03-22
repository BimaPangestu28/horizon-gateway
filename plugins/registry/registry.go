package registry

import (
	"fmt"
	"sort"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
)

type RegistryImpl struct {
	plugins     map[string]interfaces.Plugin
	phaseMap    map[interfaces.PluginPhase][]interfaces.Plugin
	phaseOrder  map[string]int
	logger      logging.Logger
	mu          sync.RWMutex
	initialized bool
}

func NewRegistry(logger logging.Logger) *RegistryImpl {
	phaseOrder := map[string]int{
		"cors":               10,
		"ip-filter":          20,
		"rate-limit":         30,
		"authentication":     40,
		"authorization":      50,
		"request-validator":  60,
		"request-transform":  70,
		"circuit-breaker":    10,
		"load-balancer":      20,
		"timeout":            30,
		"retry":              40,
		"response-transform": 10,
		"cache":              20,
		"compression":        30,
		"cors-response":      40,
		"error-handler":      10,
		"fallback":           20,
	}

	return &RegistryImpl{
		plugins:     make(map[string]interfaces.Plugin),
		phaseMap:    make(map[interfaces.PluginPhase][]interfaces.Plugin),
		phaseOrder:  phaseOrder,
		logger:      logger,
		initialized: false,
	}
}

func (r *RegistryImpl) Register(plugin interfaces.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("%w: %s", interfaces.ErrPluginAlreadyExists, name)
	}

	r.plugins[name] = plugin
	r.logger.Info("Plugin registered", "name", name, "version", plugin.Version())
	r.initialized = false

	return nil
}

func (r *RegistryImpl) Get(name string) (interfaces.Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, found := r.plugins[name]
	return plugin, found
}

func (r *RegistryImpl) GetAll() []interfaces.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]interfaces.Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name() < plugins[j].Name()
	})

	return plugins
}

func (r *RegistryImpl) GetByPhase(phase interfaces.PluginPhase) []interfaces.Plugin {
	r.initializePhases()

	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins, exists := r.phaseMap[phase]
	if !exists {
		return []interfaces.Plugin{}
	}

	result := make([]interfaces.Plugin, len(plugins))
	copy(result, plugins)
	return result
}

func (r *RegistryImpl) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	r.initializePhases()

	r.mu.RLock()
	plugins, exists := r.phaseMap[phase]
	r.mu.RUnlock()

	if !exists || len(plugins) == 0 {
		return nil
	}

	for _, p := range plugins {
		if data.AbortChain {
			r.logger.Debug("Plugin chain aborted",
				"phase", phase,
				"plugin", p.Name(),
				"message", data.AbortMessage)
			break
		}

		if err := p.Execute(phase, data); err != nil {
			r.logger.Error("Plugin execution failed",
				"phase", phase,
				"plugin", p.Name(),
				"error", err)

			data.Error = err
		}
	}

	return nil
}

func (r *RegistryImpl) Shutdown() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastError error
	for name, plugin := range r.plugins {
		if err := plugin.Shutdown(); err != nil {
			r.logger.Error("Plugin shutdown failed", "plugin", name, "error", err)
			lastError = err
		}
	}

	return lastError
}

func (r *RegistryImpl) initializePhases() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return
	}

	r.phaseMap = make(map[interfaces.PluginPhase][]interfaces.Plugin)
	r.phaseMap[interfaces.PhaseRequest] = []interfaces.Plugin{}
	r.phaseMap[interfaces.PhaseProxy] = []interfaces.Plugin{}
	r.phaseMap[interfaces.PhaseResponse] = []interfaces.Plugin{}
	r.phaseMap[interfaces.PhaseError] = []interfaces.Plugin{}

	for _, p := range r.plugins {
		var phase interfaces.PluginPhase
		switch {
		case contains([]string{"authentication", "authorization", "validator", "cors", "ip-filter", "rate-limit"}, p.Name()):
			phase = interfaces.PhaseRequest
		case contains([]string{"circuit", "timeout", "retry", "loadbalancer"}, p.Name()):
			phase = interfaces.PhaseProxy
		case contains([]string{"cache", "transform", "compression"}, p.Name()):
			phase = interfaces.PhaseResponse
		case contains([]string{"error", "fallback"}, p.Name()):
			phase = interfaces.PhaseError
		default:
			phase = interfaces.PhaseRequest
		}

		r.phaseMap[phase] = append(r.phaseMap[phase], p)
	}

	for phase, plugins := range r.phaseMap {
		sort.Slice(plugins, func(i, j int) bool {
			orderI := r.getPluginOrder(plugins[i].Name())
			orderJ := r.getPluginOrder(plugins[j].Name())
			return orderI < orderJ
		})
		r.phaseMap[phase] = plugins
	}

	r.initialized = true
}

func (r *RegistryImpl) getPluginOrder(name string) int {
	for key, order := range r.phaseOrder {
		if name == key || contains([]string{name}, key) {
			return order
		}
	}
	return 1000
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
