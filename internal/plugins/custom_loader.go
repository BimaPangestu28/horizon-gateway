package plugins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
)

// CustomPluginLoader implements the plugin loading functionality
// for various plugin types (Go plugins, WebAssembly, JavaScript, etc.)
type CustomPluginLoader struct {
	registry interfaces.PluginRegistry
	logger   logging.Logger
	mu       sync.RWMutex
}

// NewCustomPluginLoader creates a new custom plugin loader
func NewCustomPluginLoader(registry interfaces.PluginRegistry, logger logging.Logger) *CustomPluginLoader {
	return &CustomPluginLoader{
		registry: registry,
		logger:   logger,
	}
}

// LoadPluginFromFile loads a plugin from a file path
func (l *CustomPluginLoader) LoadPluginFromFile(path string) (interfaces.Plugin, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ext := filepath.Ext(path)
	switch ext {
	case ".so", ".dylib":
		return l.loadGoPlugin(path)
	case ".wasm":
		return l.loadWasmPlugin(path)
	case ".js":
		return l.loadJavaScriptPlugin(path)
	case ".lua":
		return l.loadLuaPlugin(path)
	case ".plugin":
		return l.loadCustomPlugin(path)
	default:
		return nil, fmt.Errorf("unsupported plugin format: %s", ext)
	}
}

// loadGoPlugin loads a Go plugin using the plugin package
func (l *CustomPluginLoader) loadGoPlugin(path string) (interfaces.Plugin, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening Go plugin: %w", err)
	}

	symCreate, err := p.Lookup("CreatePlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export 'CreatePlugin' symbol: %w", err)
	}

	createFunc, ok := symCreate.(func() interfaces.Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin 'CreatePlugin' symbol has wrong type: %T", symCreate)
	}

	plugin := createFunc()
	if plugin == nil {
		return nil, fmt.Errorf("plugin constructor returned nil")
	}

	l.logger.Info("Loaded Go plugin",
		"path", path,
		"name", plugin.Name(),
		"version", plugin.Version())

	return plugin, nil
}

// loadWasmPlugin loads a WebAssembly plugin
func (l *CustomPluginLoader) loadWasmPlugin(path string) (interfaces.Plugin, error) {
	// WebAssembly support will be implemented in a future phase
	return nil, fmt.Errorf("WebAssembly plugin support not implemented yet")
}

// loadJavaScriptPlugin loads a JavaScript plugin
func (l *CustomPluginLoader) loadJavaScriptPlugin(path string) (interfaces.Plugin, error) {
	// JavaScript plugins will be implemented using goja (a Go JavaScript interpreter)
	return nil, fmt.Errorf("JavaScript plugin support not implemented yet")
}

// loadLuaPlugin loads a Lua plugin
func (l *CustomPluginLoader) loadLuaPlugin(path string) (interfaces.Plugin, error) {
	// Lua plugins will be implemented using gopher-lua
	return nil, fmt.Errorf("Lua plugin support not implemented yet")
}

// loadCustomPlugin loads a custom plugin format (JSON/YAML definition)
func (l *CustomPluginLoader) loadCustomPlugin(path string) (interfaces.Plugin, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading custom plugin file: %w", err)
	}

	var definition struct {
		Name        string                 `json:"name"`
		Version     string                 `json:"version"`
		Description string                 `json:"description"`
		Author      string                 `json:"author"`
		Type        string                 `json:"type"`
		Phases      []string               `json:"phases"`
		Handlers    map[string]interface{} `json:"handlers"`
		Config      map[string]interface{} `json:"config"`
	}

	if err := json.Unmarshal(data, &definition); err != nil {
		return nil, fmt.Errorf("error parsing custom plugin definition: %w", err)
	}

	var plugin interfaces.Plugin

	switch definition.Type {
	case "script":
		plugin = NewScriptPlugin(definition.Name, definition.Version,
			definition.Description, definition.Author, definition.Handlers)
	case "http":
		plugin = NewHTTPPlugin(definition.Name, definition.Version,
			definition.Description, definition.Author, definition.Handlers)
	case "composite":
		plugin = NewCompositePlugin(definition.Name, definition.Version,
			definition.Description, definition.Author, definition.Handlers)
	default:
		return nil, fmt.Errorf("unsupported custom plugin type: %s", definition.Type)
	}

	l.logger.Info("Loaded custom plugin",
		"path", path,
		"name", definition.Name,
		"version", definition.Version,
		"type", definition.Type)

	return plugin, nil
}

// ScriptPlugin represents a plugin implemented using a scripting language
type ScriptPlugin struct {
	name        string
	version     string
	description string
	author      string
	handlers    map[string]interface{}
	config      map[string]interface{}
}

// NewScriptPlugin creates a new script plugin
func NewScriptPlugin(name, version, description, author string,
	handlers map[string]interface{}) *ScriptPlugin {
	return &ScriptPlugin{
		name:        name,
		version:     version,
		description: description,
		author:      author,
		handlers:    handlers,
		config:      make(map[string]interface{}),
	}
}

// Name returns the plugin name
func (p *ScriptPlugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *ScriptPlugin) Version() string {
	return p.version
}

// Description returns the plugin description
func (p *ScriptPlugin) Description() string {
	return p.description
}

// Author returns the plugin author
func (p *ScriptPlugin) Author() string {
	return p.author
}

// Init initializes the plugin
func (p *ScriptPlugin) Init(config map[string]interface{}) error {
	if config != nil {
		p.config = config
	}
	return nil
}

// Execute executes the plugin for a specific phase
func (p *ScriptPlugin) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	handlerName := string(phase)
	handler, exists := p.handlers[handlerName]
	if !exists {
		return nil // No handler for this phase
	}

	// Execute script handler (implementation will depend on the script engine)
	// For now, we just return nil since script execution is not implemented yet
	return nil
}

// Shutdown cleans up resources used by the plugin
func (p *ScriptPlugin) Shutdown() error {
	return nil
}

// HTTPPlugin represents a plugin that calls an HTTP endpoint for processing
type HTTPPlugin struct {
	name        string
	version     string
	description string
	author      string
	handlers    map[string]interface{}
	config      map[string]interface{}
}

// NewHTTPPlugin creates a new HTTP plugin
func NewHTTPPlugin(name, version, description, author string,
	handlers map[string]interface{}) *HTTPPlugin {
	return &HTTPPlugin{
		name:        name,
		version:     version,
		description: description,
		author:      author,
		handlers:    handlers,
		config:      make(map[string]interface{}),
	}
}

// Name returns the plugin name
func (p *HTTPPlugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *HTTPPlugin) Version() string {
	return p.version
}

// Description returns the plugin description
func (p *HTTPPlugin) Description() string {
	return p.description
}

// Author returns the plugin author
func (p *HTTPPlugin) Author() string {
	return p.author
}

// Init initializes the plugin
func (p *HTTPPlugin) Init(config map[string]interface{}) error {
	if config != nil {
		p.config = config
	}
	return nil
}

// Execute executes the plugin for a specific phase
func (p *HTTPPlugin) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	handlerName := string(phase)
	handlerConfig, exists := p.handlers[handlerName]
	if !exists {
		return nil // No handler for this phase
	}

	// Convert handler configuration to endpoint details
	handlerConfigMap, ok := handlerConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid handler configuration format")
	}

	url, ok := handlerConfigMap["url"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid URL in handler configuration")
	}

	// Make HTTP request to the handler endpoint
	// For now, we just return nil since HTTP execution is not implemented yet
	return nil
}

// Shutdown cleans up resources used by the plugin
func (p *HTTPPlugin) Shutdown() error {
	return nil
}

// CompositePlugin represents a plugin composed of multiple other plugins
type CompositePlugin struct {
	name        string
	version     string
	description string
	author      string
	handlers    map[string]interface{}
	config      map[string]interface{}
	subPlugins  []interfaces.Plugin
}

// NewCompositePlugin creates a new composite plugin
func NewCompositePlugin(name, version, description, author string,
	handlers map[string]interface{}) *CompositePlugin {
	return &CompositePlugin{
		name:        name,
		version:     version,
		description: description,
		author:      author,
		handlers:    handlers,
		config:      make(map[string]interface{}),
		subPlugins:  make([]interfaces.Plugin, 0),
	}
}

// Name returns the plugin name
func (p *CompositePlugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *CompositePlugin) Version() string {
	return p.version
}

// Description returns the plugin description
func (p *CompositePlugin) Description() string {
	return p.description
}

// Author returns the plugin author
func (p *CompositePlugin) Author() string {
	return p.author
}

// Init initializes the plugin
func (p *CompositePlugin) Init(config map[string]interface{}) error {
	if config != nil {
		p.config = config
	}

	// Initialize sub-plugins if any
	for _, plugin := range p.subPlugins {
		if err := plugin.Init(nil); err != nil {
			return fmt.Errorf("error initializing sub-plugin %s: %w", plugin.Name(), err)
		}
	}

	return nil
}

// Execute executes the plugin for a specific phase
func (p *CompositePlugin) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	// Execute all sub-plugins for this phase
	for _, plugin := range p.subPlugins {
		if err := plugin.Execute(phase, data); err != nil {
			return err
		}

		// If chain was aborted, stop execution
		if data.AbortChain {
			break
		}
	}

	return nil
}

// Shutdown cleans up resources used by the plugin
func (p *CompositePlugin) Shutdown() error {
	// Shutdown all sub-plugins
	for _, plugin := range p.subPlugins {
		if err := plugin.Shutdown(); err != nil {
			return err
		}
	}

	return nil
}

// AddSubPlugin adds a sub-plugin to the composite plugin
func (p *CompositePlugin) AddSubPlugin(plugin interfaces.Plugin) {
	p.subPlugins = append(p.subPlugins, plugin)
}
