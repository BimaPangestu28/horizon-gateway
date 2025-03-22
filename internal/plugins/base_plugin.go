package plugins

import (
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
)

type BasePlugin struct {
	name        string
	version     string
	description string
	author      string
	config      map[string]interface{}
	enabled     bool
	logger      logging.Logger
	mu          sync.RWMutex
}

func NewBasePlugin(name, version, description, author string, logger logging.Logger) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
		author:      author,
		config:      make(map[string]interface{}),
		enabled:     true,
		logger:      logger,
	}
}

func (p *BasePlugin) Name() string {
	return p.name
}

func (p *BasePlugin) Version() string {
	return p.version
}

func (p *BasePlugin) Description() string {
	return p.description
}

func (p *BasePlugin) Author() string {
	return p.author
}

func (p *BasePlugin) Init(config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if config != nil {
		p.config = config
	}

	if enabled, ok := p.config["enabled"].(bool); ok {
		p.enabled = enabled
	}

	return nil
}

func (p *BasePlugin) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.enabled {
		return nil
	}

	return nil
}

func (p *BasePlugin) Shutdown() error {
	return nil
}

func (p *BasePlugin) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

func (p *BasePlugin) SetEnabled(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = enabled
}

func (p *BasePlugin) GetConfig() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	config := make(map[string]interface{})
	for k, v := range p.config {
		config[k] = v
	}

	return config
}

func (p *BasePlugin) SetConfig(config map[string]interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config
}

func (p *BasePlugin) GetConfigValue(key string) (interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	value, ok := p.config[key]
	return value, ok
}

func (p *BasePlugin) SetConfigValue(key string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config[key] = value
}
