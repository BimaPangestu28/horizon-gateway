package loader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin" as goplugin
	"strings"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
)

type LoaderImpl struct {
	registry interfaces.PluginRegistry
	logger   logging.Logger
}

func NewLoader(registry interfaces.PluginRegistry, logger logging.Logger) *LoaderImpl {
	return &LoaderImpl{
		registry: registry,
		logger:   logger,
	}
}

func (l *LoaderImpl) Load(source string) ([]interfaces.Plugin, error) {
	fileInfo, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("error checking plugin source: %w", err)
	}

	if fileInfo.IsDir() {
		return l.LoadDirectory(source)
	}

	plugin, err := l.loadPlugin(source)
	if err != nil {
		return nil, err
	}

	return []interfaces.Plugin{plugin}, nil
}

func (l *LoaderImpl) LoadDirectory(dir string) ([]interfaces.Plugin, error) {
	l.logger.Info("Loading plugins from directory", "dir", dir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin directory does not exist: %s", dir)
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading plugin directory: %w", err)
	}

	var loadedPlugins []interfaces.Plugin

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext != ".so" && ext != ".plugin" && ext != ".wasm" {
			continue
		}

		pluginPath := filepath.Join(dir, file.Name())
		l.logger.Debug("Loading plugin file", "path", pluginPath)

		plugin, err := l.loadPlugin(pluginPath)
		if err != nil {
			l.logger.Error("Failed to load plugin", "path", pluginPath, "error", err)
			continue
		}

		loadedPlugins = append(loadedPlugins, plugin)
		
		if err := l.registry.Register(plugin); err != nil {
			l.logger.Error("Failed to register plugin", "name", plugin.Name(), "error", err)
		}
	}

	l.logger.Info("Loaded plugins", "count", len(loadedPlugins))
	return loadedPlugins, nil
}

func (l *LoaderImpl) loadPlugin(path string) (interfaces.Plugin, error) {
	ext := filepath.Ext(path)

	switch ext {
	case ".so":
		return l.loadGoPlugin(path)
	case ".wasm":
		return l.loadWasmPlugin(path)
	case ".plugin":
		return l.loadCustomPlugin(path)
	default:
		return nil, fmt.Errorf("unsupported plugin format: %s", ext)
	}
}

func (l *LoaderImpl) loadGoPlugin(path string) (interfaces.Plugin, error) {
	p, err := goplugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening Go plugin: %w", err)
	}

	sym, err := p.Lookup("CreatePlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export 'CreatePlugin' symbol: %w", err)
	}

	createPlugin, ok := sym.(func() interfaces.Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin 'CreatePlugin' symbol is not a valid constructor")
	}

	plugin := createPlugin()
	if plugin == nil {
		return nil, fmt.Errorf("plugin constructor returned nil")
	}

	l.logger.Info("Loaded Go plugin", "path", path, "name", plugin.Name())
	return plugin, nil
}

func (l *LoaderImpl) loadWasmPlugin(path string) (interfaces.Plugin, error) {
	return nil, fmt.Errorf("WebAssembly plugin loading not implemented yet")
}

func (l *LoaderImpl) loadCustomPlugin(path string) (interfaces.Plugin, error) {
	return nil, fmt.Errorf("Custom plugin loading not implemented yet")
}

func (l *LoaderImpl) validatePlugin(plugin interfaces.Plugin) error {
	if plugin.Name() == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if plugin.Version() == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	return nil
}