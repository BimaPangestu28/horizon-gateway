package discovery

import (
	"context"
	"errors"
	"time"
)

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrNoHealthyInstances   = errors.New("no healthy instances available")
	ErrDiscoveryFailed      = errors.New("service discovery failed")
	ErrInvalidConfiguration = errors.New("invalid service discovery configuration")
)

type ServiceInstance struct {
	ID        string
	Name      string
	Address   string
	Port      int
	Secure    bool
	Metadata  map[string]string
	Tags      []string
	Healthy   bool
	Weight    int
	Zone      string
	Version   string
	LastCheck time.Time
}

type ServiceDiscovery interface {
	// Initialize the service discovery client
	Initialize(ctx context.Context) error

	// GetService returns all instances of a service
	GetService(ctx context.Context, name string) ([]ServiceInstance, error)

	// GetInstance returns a single instance of a service
	GetInstance(ctx context.Context, name string) (*ServiceInstance, error)

	// RegisterService registers a service instance
	RegisterService(ctx context.Context, instance *ServiceInstance) error

	// DeregisterService deregisters a service instance
	DeregisterService(ctx context.Context, instanceID string) error

	// Watch for service changes
	Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error)

	// GetHealthStatus returns health status of a service
	GetHealthStatus(ctx context.Context, name string) (*HealthStatus, error)

	// Close closes the service discovery client
	Close() error
}

type HealthStatus struct {
	ServiceName     string
	InstanceCount   int
	HealthyCount    int
	UnhealthyCount  int
	LastRefreshed   time.Time
	Status          string
	InstancesByZone map[string]int
}

type ServiceRegistry struct {
	instances   map[string][]ServiceInstance
	discoveries map[string]ServiceDiscovery
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		instances:   make(map[string][]ServiceInstance),
		discoveries: make(map[string]ServiceDiscovery),
	}
}

func (r *ServiceRegistry) RegisterDiscovery(name string, discovery ServiceDiscovery) {
	r.discoveries[name] = discovery
}

func (r *ServiceRegistry) GetDiscovery(name string) (ServiceDiscovery, bool) {
	discovery, ok := r.discoveries[name]
	return discovery, ok
}

func (r *ServiceRegistry) Initialize(ctx context.Context) error {
	for name, discovery := range r.discoveries {
		if err := discovery.Initialize(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *ServiceRegistry) GetService(ctx context.Context, name string, discoveryType string) ([]ServiceInstance, error) {
	discovery, ok := r.discoveries[discoveryType]
	if !ok {
		return nil, errors.New("discovery type not found: " + discoveryType)
	}

	return discovery.GetService(ctx, name)
}

func (r *ServiceRegistry) GetInstance(ctx context.Context, name string, discoveryType string) (*ServiceInstance, error) {
	discovery, ok := r.discoveries[discoveryType]
	if !ok {
		return nil, errors.New("discovery type not found: " + discoveryType)
	}

	return discovery.GetInstance(ctx, name)
}

func (r *ServiceRegistry) Close() error {
	var errs []error
	for name, discovery := range r.discoveries {
		if err := discovery.Close(); err != nil {
			errs = append(errs, errors.New("error closing discovery "+name+": "+err.Error()))
		}
	}

	if len(errs) > 0 {
		return errors.New("errors closing service registry")
	}

	return nil
}
