package interfaces

import (
	"context"
	"errors"
)

type PluginPhase string

const (
	PhaseRequest  PluginPhase = "request"
	PhaseProxy    PluginPhase = "proxy"
	PhaseResponse PluginPhase = "response"
	PhaseError    PluginPhase = "error"
)

var (
	ErrPluginNotFound      = errors.New("plugin not found")
	ErrPluginAlreadyExists = errors.New("plugin already exists")
	ErrPluginInitFailed    = errors.New("plugin initialization failed")
	ErrPluginExecuteFailed = errors.New("plugin execution failed")
)

type PluginData struct {
	Request      *Request
	Response     *Response
	Route        *Route
	Error        error
	Context      context.Context
	Metadata     map[string]interface{}
	AbortChain   bool
	AbortMessage string
}

type Request struct {
	Method      string
	URL         string
	Path        string
	Headers     map[string][]string
	QueryParams map[string][]string
	Body        []byte
	ClientIP    string
}

type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

type Route struct {
	Name        string
	ListenPath  string
	UpstreamURL string
	StripPath   bool
}

type Plugin interface {
	Name() string
	Version() string
	Description() string
	Author() string
	Init(config map[string]interface{}) error
	Execute(phase PluginPhase, data *PluginData) error
	Shutdown() error
}

type PluginRegistry interface {
	Register(plugin Plugin) error
	Get(name string) (Plugin, bool)
	GetAll() []Plugin
	GetByPhase(phase PluginPhase) []Plugin
	Execute(phase PluginPhase, data *PluginData) error
	Shutdown() error
}

type PluginLoader interface {
	Load(source string) ([]Plugin, error)
	LoadDirectory(dir string) ([]Plugin, error)
}

type PluginConfig struct {
	Name    string                 `yaml:"name" json:"name"`
	Enabled bool                   `yaml:"enabled" json:"enabled"`
	Phase   PluginPhase            `yaml:"phase" json:"phase"`
	Config  map[string]interface{} `yaml:"config" json:"config"`
}

type PluginInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Author      string     `json:"author"`
	Phases      []string   `json:"phases"`
	Config      ConfigInfo `json:"config"`
}

type ConfigInfo struct {
	Properties  map[string]PropertyInfo `json:"properties"`
	Required    []string                `json:"required"`
	Description string                  `json:"description"`
}

type PropertyInfo struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}
