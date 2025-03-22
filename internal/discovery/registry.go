package discovery

import (
	"context"
	"fmt"
	"sync"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type DiscoveryType string

const (
	DiscoveryTypeDirect     DiscoveryType = "direct"
	DiscoveryTypeConsul     DiscoveryType = "consul"
	DiscoveryTypeKubernetes DiscoveryType = "kubernetes"
	DiscoveryTypeEtcd       DiscoveryType = "etcd"
	DiscoveryTypeDNS        DiscoveryType = "dns"
)

type DiscoveryRegistry struct {
	discoveries map[DiscoveryType]ServiceDiscovery
	logger      logging.Logger
	mu          sync.RWMutex
	initialized bool
}

func NewDiscoveryRegistry(logger logging.Logger) *DiscoveryRegistry {
	return &DiscoveryRegistry{
		discoveries: make(map[DiscoveryType]ServiceDiscovery),
		logger:      logger,
		initialized: false,
	}
}

func (r *DiscoveryRegistry) Initialize(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for typ, discovery := range r.discoveries {
		if err := discovery.Initialize(ctx); err != nil {
			r.logger.Error("Failed to initialize discovery",
				"type", string(typ),
				"error", err)
			return fmt.Errorf("failed to initialize %s discovery: %w", typ, err)
		}
		r.logger.Info("Initialized service discovery", "type", string(typ))
	}

	r.initialized = true
	return nil
}

func (r *DiscoveryRegistry) Register(typ DiscoveryType, discovery ServiceDiscovery) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.discoveries[typ] = discovery
	r.logger.Info("Registered service discovery", "type", string(typ))
}

func (r *DiscoveryRegistry) Get(typ DiscoveryType) (ServiceDiscovery, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	discovery, exists := r.discoveries[typ]
	return discovery, exists
}

func (r *DiscoveryRegistry) GetAll() map[DiscoveryType]ServiceDiscovery {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[DiscoveryType]ServiceDiscovery, len(r.discoveries))
	for k, v := range r.discoveries {
		result[k] = v
	}
	return result
}

func (r *DiscoveryRegistry) GetService(ctx context.Context, serviceName string, discoveryType DiscoveryType) ([]ServiceInstance, error) {
	if discoveryType == "" {
		return r.GetServiceFromAny(ctx, serviceName)
	}

	discovery, exists := r.Get(discoveryType)
	if !exists {
		return nil, fmt.Errorf("discovery type %s not found", discoveryType)
	}

	return discovery.GetService(ctx, serviceName)
}

func (r *DiscoveryRegistry) GetServiceFromAny(ctx context.Context, serviceName string) ([]ServiceInstance, error) {
	r.mu.RLock()
	discoveries := make([]ServiceDiscovery, 0, len(r.discoveries))
	for _, discovery := range r.discoveries {
		discoveries = append(discoveries, discovery)
	}
	r.mu.RUnlock()

	var lastErr error
	for _, discovery := range discoveries {
		instances, err := discovery.GetService(ctx, serviceName)
		if err == nil {
			return instances, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, ErrServiceNotFound
}

func (r *DiscoveryRegistry) GetInstance(ctx context.Context, serviceName string, discoveryType DiscoveryType) (*ServiceInstance, error) {
	if discoveryType == "" {
		return r.GetInstanceFromAny(ctx, serviceName)
	}

	discovery, exists := r.Get(discoveryType)
	if !exists {
		return nil, fmt.Errorf("discovery type %s not found", discoveryType)
	}

	return discovery.GetInstance(ctx, serviceName)
}

func (r *DiscoveryRegistry) GetInstanceFromAny(ctx context.Context, serviceName string) (*ServiceInstance, error) {
	r.mu.RLock()
	discoveries := make([]ServiceDiscovery, 0, len(r.discoveries))
	for _, discovery := range r.discoveries {
		discoveries = append(discoveries, discovery)
	}
	r.mu.RUnlock()

	var lastErr error
	for _, discovery := range discoveries {
		instance, err := discovery.GetInstance(ctx, serviceName)
		if err == nil {
			return instance, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, ErrServiceNotFound
}

func (r *DiscoveryRegistry) RegisterService(ctx context.Context, instance *ServiceInstance, discoveryType DiscoveryType) error {
	discovery, exists := r.Get(discoveryType)
	if !exists {
		return fmt.Errorf("discovery type %s not found", discoveryType)
	}

	return discovery.RegisterService(ctx, instance)
}

func (r *DiscoveryRegistry) DeregisterService(ctx context.Context, instanceID string, discoveryType DiscoveryType) error {
	discovery, exists := r.Get(discoveryType)
	if !exists {
		return fmt.Errorf("discovery type %s not found", discoveryType)
	}

	return discovery.DeregisterService(ctx, instanceID)
}

func (r *DiscoveryRegistry) Watch(ctx context.Context, serviceName string, discoveryType DiscoveryType) (<-chan []ServiceInstance, error) {
	discovery, exists := r.Get(discoveryType)
	if !exists {
		return nil, fmt.Errorf("discovery type %s not found", discoveryType)
	}

	return discovery.Watch(ctx, serviceName)
}

func (r *DiscoveryRegistry) GetHealthStatus(ctx context.Context, serviceName string, discoveryType DiscoveryType) (*HealthStatus, error) {
	discovery, exists := r.Get(discoveryType)
	if !exists {
		return nil, fmt.Errorf("discovery type %s not found", discoveryType)
	}

	return discovery.GetHealthStatus(ctx, serviceName)
}

func (r *DiscoveryRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for typ, discovery := range r.discoveries {
		if err := discovery.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing %s discovery: %w", typ, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing service discovery registry: %v", errs)
	}

	return nil
}
