package config

import (
	"time"

	"github.com/bimapangestu28/horizon/internal/cache"
	"github.com/bimapangestu28/horizon/internal/resilience/circuitbreaker"
	"github.com/bimapangestu28/horizon/internal/security/auth"
	"github.com/bimapangestu28/horizon/internal/security/ipfilter"
	"github.com/bimapangestu28/horizon/internal/security/ratelimit"
	"github.com/bimapangestu28/horizon/internal/transform"
	"github.com/bimapangestu28/horizon/internal/types"
	interfaces "github.com/bimapangestu28/horizon/plugins/interfaces"
)

type RouteConfig struct {
	Name             string                               `yaml:"name"`
	ListenPath       string                               `yaml:"listen_path"`
	UpstreamURL      string                               `yaml:"upstream_url"`
	Methods          []string                             `yaml:"methods"`
	StripPath        bool                                 `yaml:"strip_path"`
	Headers          map[string]string                    `yaml:"headers,omitempty"`
	QueryParams      map[string]string                    `yaml:"query_params,omitempty"`
	Host             string                               `yaml:"host,omitempty"`
	Priority         int                                  `yaml:"priority,omitempty"`
	LoadBalancing    *LoadBalancingConfig                 `yaml:"load_balancing,omitempty"`
	IPFilter         *ipfilter.IPFilterConfig             `yaml:"ip_filter,omitempty"`
	Auth             *auth.AuthConfig                     `yaml:"auth,omitempty"`
	RateLimiting     *ratelimit.RateLimitConfig           `yaml:"rate_limiting,omitempty"`
	CircuitBreaker   *circuitbreaker.CircuitBreakerConfig `yaml:"circuit_breaker,omitempty"`
	Caching          *cache.CacheConfig                   `yaml:"caching,omitempty"`
	Transform        *transform.TransformConfig           `yaml:"transform,omitempty"`
	WebSocket        *types.WebSocketConfig               `yaml:"websocket,omitempty"`
	GRPC             *GRPCConfig                          `yaml:"grpc,omitempty"`
	ServiceDiscovery *ServiceDiscoveryConfig              `yaml:"service_discovery,omitempty"`
	Plugins          []interfaces.PluginConfig            `yaml:"plugins,omitempty"`
	HealthCheck      *HealthCheckConfig                   `yaml:"health_check,omitempty"`
	Timeout          *TimeoutConfig                       `yaml:"timeout,omitempty"`
	Retry            *RetryConfig                         `yaml:"retry,omitempty"`
}

type LoadBalancingConfig struct {
	Type        string             `yaml:"type"`
	Targets     []TargetConfig     `yaml:"targets,omitempty"`
	HealthCheck *HealthCheckConfig `yaml:"health_check,omitempty"`
}

type TargetConfig struct {
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight,omitempty"`
}

type HealthCheckConfig struct {
	Path               string `yaml:"path"`
	Interval           string `yaml:"interval"`
	Timeout            string `yaml:"timeout"`
	HealthyThreshold   int    `yaml:"healthy_threshold"`
	UnhealthyThreshold int    `yaml:"unhealthy_threshold"`
}

type GRPCConfig struct {
	Enabled            bool          `yaml:"enabled"`
	UpstreamURL        string        `yaml:"upstream_url"`
	ProtoDescriptorDir string        `yaml:"proto_descriptor_dir"`
	ServiceName        string        `yaml:"service_name"`
	MaxMessageSize     int           `yaml:"max_message_size"`
	MaxConnectionIdle  time.Duration `yaml:"max_connection_idle"`
	MaxConnectionAge   time.Duration `yaml:"max_connection_age"`
	KeepAlive          time.Duration `yaml:"keep_alive"`
	KeepAliveTimeout   time.Duration `yaml:"keep_alive_timeout"`
	ForwardHeaders     []string      `yaml:"forward_headers"`
}

type ServiceDiscoveryConfig struct {
	Enabled          bool                              `yaml:"enabled"`
	Type             string                            `yaml:"type"`
	ServiceName      string                            `yaml:"service_name"`
	Namespace        string                            `yaml:"namespace"`
	RefreshInterval  time.Duration                     `yaml:"refresh_interval"`
	ConsulConfig     *ConsulServiceDiscoveryConfig     `yaml:"consul,omitempty"`
	KubernetesConfig *KubernetesServiceDiscoveryConfig `yaml:"kubernetes,omitempty"`
	EtcdConfig       *EtcdServiceDiscoveryConfig       `yaml:"etcd,omitempty"`
	DNSConfig        *DNSServiceDiscoveryConfig        `yaml:"dns,omitempty"`
}

type ConsulServiceDiscoveryConfig struct {
	Address     string `yaml:"address"`
	Datacenter  string `yaml:"datacenter"`
	Token       string `yaml:"token"`
	TagFilter   string `yaml:"tag_filter"`
	HealthCheck bool   `yaml:"health_check"`
}

type KubernetesServiceDiscoveryConfig struct {
	InCluster     bool   `yaml:"in_cluster"`
	KubeConfig    string `yaml:"kube_config"`
	LabelSelector string `yaml:"label_selector"`
	FieldSelector string `yaml:"field_selector"`
	Port          int    `yaml:"port"`
}

type EtcdServiceDiscoveryConfig struct {
	Endpoints []string   `yaml:"endpoints"`
	Username  string     `yaml:"username"`
	Password  string     `yaml:"password"`
	KeyPrefix string     `yaml:"key_prefix"`
	TLS       *TLSConfig `yaml:"tls,omitempty"`
}

type DNSServiceDiscoveryConfig struct {
	Domain    string `yaml:"domain"`
	QueryType string `yaml:"query_type"`
	Port      int    `yaml:"port"`
	Resolver  string `yaml:"resolver"`
}

type TimeoutConfig struct {
	RequestTimeout time.Duration `yaml:"request_timeout"`
	DialTimeout    time.Duration `yaml:"dial_timeout"`
	TLSHandshake   time.Duration `yaml:"tls_handshake"`
	ResponseHeader time.Duration `yaml:"response_header"`
	IdleConnection time.Duration `yaml:"idle_connection"`
}

type RetryConfig struct {
	Attempts      int           `yaml:"attempts"`
	Backoff       time.Duration `yaml:"backoff"`
	BackoffFactor float64       `yaml:"backoff_factor"`
	MaxBackoff    time.Duration `yaml:"max_backoff"`
	RetryOn       []int         `yaml:"retry_on"`
}
