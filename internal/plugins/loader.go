package plugins

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/plugins/interfaces"
)

type PluginLoader struct {
	logger      logging.Logger
	registry    *PluginRegistry
	loadedPaths map[string]string
	mu          sync.Mutex
}

func NewPluginLoader(registry *PluginRegistry, logger logging.Logger) *PluginLoader {
	return &PluginLoader{
		logger:      logger,
		registry:    registry,
		loadedPaths: make(map[string]string),
	}
}

func (l *PluginLoader) LoadPlugin(path string) (interfaces.Plugin, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !fileExists(path) {
		return nil, fmt.Errorf("%w: file does not exist at %s", ErrInvalidPluginPath, path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPluginPath, err)
	}

	plug, err := plugin.Open(absPath)
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

	l.loadedPaths[p.Name()] = absPath
	l.logger.Info("Plugin loaded", "name", p.Name(), "version", p.Version(), "path", absPath)

	return p, nil
}

func (l *PluginLoader) LoadDirectory(dirPath string) ([]interfaces.Plugin, error) {
	if !dirExists(dirPath) {
		return nil, fmt.Errorf("%w: directory does not exist at %s", ErrInvalidPluginPath, dirPath)
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var loadedPlugins []interfaces.Plugin
	for _, file := range files {
		if file.IsDir() || !isSupportedPluginFile(file.Name()) {
			continue
		}

		fullPath := filepath.Join(dirPath, file.Name())
		plugin, err := l.LoadPlugin(fullPath)
		if err != nil {
			l.logger.Error("Failed to load plugin", "path", fullPath, "error", err)
			continue
		}

		loadedPlugins = append(loadedPlugins, plugin)
	}

	if len(loadedPlugins) == 0 {
		return nil, ErrNoPluginsFound
	}

	return loadedPlugins, nil
}

func (l *PluginLoader) InitializePlugin(p interfaces.Plugin, config map[string]interface{}) error {
	if err := p.Init(config); err != nil {
		l.logger.Error("Failed to initialize plugin", "name", p.Name(), "error", err)
		return fmt.Errorf("%w: %v", ErrPluginInitFailed, err)
	}

	return nil
}

func (l *PluginLoader) RegisterPlugin(p interfaces.Plugin) error {
	return l.registry.Register(p)
}

func (l *PluginLoader) GetLoadedPluginPath(name string) (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	path, ok := l.loadedPaths[name]
	return path, ok
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func isSupportedPluginFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".so" || ext == ".dylib" || ext == ".plugin"
}
