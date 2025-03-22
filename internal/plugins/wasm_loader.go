package plugins

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
	"github.com/wasmerio/wasmer-go/wasmer"
)

type WasmPluginLoader struct {
	logger      logging.Logger
	registry    *PluginRegistry
	loadedPaths map[string]string
	store       *wasmer.Store
	mu          sync.Mutex
}

func NewWasmPluginLoader(registry *PluginRegistry, logger logging.Logger) *WasmPluginLoader {
	config := wasmer.NewConfig()
	engine := wasmer.NewEngine(config)
	store := wasmer.NewStore(engine)

	return &WasmPluginLoader{
		logger:      logger,
		registry:    registry,
		loadedPaths: make(map[string]string),
		store:       store,
	}
}

func (l *WasmPluginLoader) LoadPlugin(path string) (interfaces.Plugin, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !fileExists(path) {
		return nil, fmt.Errorf("%w: file does not exist at %s", ErrInvalidPluginPath, path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPluginPath, err)
	}

	// Read WebAssembly module file
	wasmBytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error reading WASM file: %w", err)
	}

	// Compile the module
	module, err := wasmer.NewModule(l.store, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("error compiling WASM module: %w", err)
	}

	// Create imports object
	importObject := wasmer.NewImportObject()

	// Instantiate the module with imports
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		return nil, fmt.Errorf("error instantiating WASM module: %w", err)
	}

	// Get required exports
	nameFunc, err := instance.Exports.GetFunction("name")
	if err != nil {
		return nil, fmt.Errorf("WASM module does not export 'name' function: %w", err)
	}

	versionFunc, err := instance.Exports.GetFunction("version")
	if err != nil {
		return nil, fmt.Errorf("WASM module does not export 'version' function: %w", err)
	}

	descriptionFunc, err := instance.Exports.GetFunction("description")
	if err != nil {
		return nil, fmt.Errorf("WASM module does not export 'description' function: %w", err)
	}

	authorFunc, err := instance.Exports.GetFunction("author")
	if err != nil {
		return nil, fmt.Errorf("WASM module does not export 'author' function: %w", err)
	}

	initFunc, err := instance.Exports.GetFunction("init")
	if err != nil {
		return nil, fmt.Errorf("WASM module does not export 'init' function: %w", err)
	}

	executeFunc, err := instance.Exports.GetFunction("execute")
	if err != nil {
		return nil, fmt.Errorf("WASM module does not export 'execute' function: %w", err)
	}

	shutdownFunc, err := instance.Exports.GetFunction("shutdown")
	if err != nil {
		return nil, fmt.Errorf("WASM module does not export 'shutdown' function: %w", err)
	}

	// Execute name function to get plugin name
	nameResult, err := nameFunc()
	if err != nil {
		return nil, fmt.Errorf("error getting plugin name: %w", err)
	}

	// Convert pointer and length to string
	namePtr := nameResult.(int32)
	nameLenResult, err := instance.Exports.GetFunction("name_len")()
	if err != nil {
		return nil, fmt.Errorf("error getting plugin name length: %w", err)
	}
	nameLen := nameLenResult.(int32)

	memory := instance.Exports.GetMemory("memory")
	if memory == nil {
		return nil, fmt.Errorf("WASM module does not export 'memory'")
	}

	nameBytes := memory.Data()[namePtr : namePtr+nameLen]
	name := string(nameBytes)

	// Similarly get version, description and author
	versionResult, _ := versionFunc()
	versionPtr := versionResult.(int32)
	versionLenResult, _ := instance.Exports.GetFunction("version_len")()
	versionLen := versionLenResult.(int32)
	versionBytes := memory.Data()[versionPtr : versionPtr+versionLen]
	version := string(versionBytes)

	descriptionResult, _ := descriptionFunc()
	descriptionPtr := descriptionResult.(int32)
	descriptionLenResult, _ := instance.Exports.GetFunction("description_len")()
	descriptionLen := descriptionLenResult.(int32)
	descriptionBytes := memory.Data()[descriptionPtr : descriptionPtr+descriptionLen]
	description := string(descriptionBytes)

	authorResult, _ := authorFunc()
	authorPtr := authorResult.(int32)
	authorLenResult, _ := instance.Exports.GetFunction("author_len")()
	authorLen := authorLenResult.(int32)
	authorBytes := memory.Data()[authorPtr : authorPtr+authorLen]
	author := string(authorBytes)

	// Create WASM plugin
	plugin := &WasmPlugin{
		name:         name,
		version:      version,
		description:  description,
		author:       author,
		instance:     instance,
		initFunc:     initFunc,
		executeFunc:  executeFunc,
		shutdownFunc: shutdownFunc,
		memory:       memory,
		logger:       l.logger,
	}

	l.loadedPaths[plugin.Name()] = absPath
	l.logger.Info("WASM plugin loaded", "name", plugin.Name(), "version", plugin.Version(), "path", absPath)

	return plugin, nil
}

func (l *WasmPluginLoader) LoadDirectory(dirPath string) ([]interfaces.Plugin, error) {
	if !dirExists(dirPath) {
		return nil, fmt.Errorf("%w: directory does not exist at %s", ErrInvalidPluginPath, dirPath)
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var loadedPlugins []interfaces.Plugin
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".wasm" {
			continue
		}

		fullPath := filepath.Join(dirPath, file.Name())
		plugin, err := l.LoadPlugin(fullPath)
		if err != nil {
			l.logger.Error("Failed to load WASM plugin", "path", fullPath, "error", err)
			continue
		}

		loadedPlugins = append(loadedPlugins, plugin)
	}

	if len(loadedPlugins) == 0 {
		return nil, ErrNoPluginsFound
	}

	return loadedPlugins, nil
}

func (l *WasmPluginLoader) InitializePlugin(p interfaces.Plugin, config map[string]interface{}) error {
	wasmPlugin, ok := p.(*WasmPlugin)
	if !ok {
		return fmt.Errorf("not a WASM plugin")
	}

	return wasmPlugin.Init(config)
}

func (l *WasmPluginLoader) RegisterPlugin(p interfaces.Plugin) error {
	return l.registry.Register(p)
}

func (l *WasmPluginLoader) GetLoadedPluginPath(name string) (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	path, ok := l.loadedPaths[name]
	return path, ok
}

type WasmPlugin struct {
	name         string
	version      string
	description  string
	author       string
	instance     *wasmer.Instance
	initFunc     *wasmer.Function
	executeFunc  *wasmer.Function
	shutdownFunc *wasmer.Function
	memory       *wasmer.Memory
	logger       logging.Logger
	config       map[string]interface{}
	mu           sync.RWMutex
}

func (p *WasmPlugin) Name() string {
	return p.name
}

func (p *WasmPlugin) Version() string {
	return p.version
}

func (p *WasmPlugin) Description() string {
	return p.description
}

func (p *WasmPlugin) Author() string {
	return p.author
}

func (p *WasmPlugin) Init(config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config

	// Convert config to JSON
	configJSON := "{}"
	if config != nil {
		// Implementation would convert map to JSON string
	}

	// Allocate memory for config JSON
	configPtr, err := p.allocateStringInWasm(configJSON)
	if err != nil {
		return fmt.Errorf("error allocating memory for config: %w", err)
	}

	// Call init function
	result, err := p.initFunc(configPtr, int32(len(configJSON)))
	if err != nil {
		return fmt.Errorf("error initializing WASM plugin: %w", err)
	}

	status := result.(int32)
	if status != 0 {
		return fmt.Errorf("WASM plugin initialization failed with status %d", status)
	}

	return nil
}

func (p *WasmPlugin) Execute(phase interfaces.PluginPhase, data *interfaces.PluginData) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Convert data to format that can be passed to WASM
	// This is a simplified version, in practice you'd need to serialize the data
	phaseStr := string(phase)
	phasePtr, err := p.allocateStringInWasm(phaseStr)
	if err != nil {
		return fmt.Errorf("error allocating memory for phase: %w", err)
	}

	// Serialize data to JSON
	dataJSON := "{}"
	// Implementation would serialize data to JSON

	dataPtr, err := p.allocateStringInWasm(dataJSON)
	if err != nil {
		return fmt.Errorf("error allocating memory for data: %w", err)
	}

	// Call execute function
	result, err := p.executeFunc(phasePtr, int32(len(phaseStr)), dataPtr, int32(len(dataJSON)))
	if err != nil {
		return fmt.Errorf("error executing WASM plugin: %w", err)
	}

	status := result.(int32)
	if status != 0 {
		return fmt.Errorf("WASM plugin execution failed with status %d", status)
	}

	// In a real implementation, you'd read the results back from memory
	// and update the data parameter

	return nil
}

func (p *WasmPlugin) Shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, err := p.shutdownFunc()
	if err != nil {
		return fmt.Errorf("error shutting down WASM plugin: %w", err)
	}

	return nil
}

func (p *WasmPlugin) allocateStringInWasm(str string) (int32, error) {
	// This is a simplified implementation
	// In practice, you'd need to use memory allocation functions exported by the WASM module
	// or implement your own allocator

	// For now, we'll just assume there's an exported "allocate" function
	allocateFunc, err := p.instance.Exports.GetFunction("allocate")
	if err != nil {
		return 0, fmt.Errorf("WASM module does not export 'allocate' function: %w", err)
	}

	result, err := allocateFunc(int32(len(str)))
	if err != nil {
		return 0, fmt.Errorf("error allocating memory in WASM: %w", err)
	}

	ptr := result.(int32)
	if ptr == 0 {
		return 0, fmt.Errorf("WASM memory allocation failed")
	}

	// Copy string to WASM memory
	copy(p.memory.Data()[ptr:ptr+int32(len(str))], []byte(str))

	return ptr, nil
}
