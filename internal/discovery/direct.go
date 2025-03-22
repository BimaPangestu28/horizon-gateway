package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type DirectServiceDiscovery struct {
	logger      logging.Logger
	serviceMap  map[string][]ServiceInstance
	mu          sync.RWMutex
	watchCh     map[string][]chan []ServiceInstance
	watchChanMu sync.RWMutex
	initialized bool
}

func NewDirectServiceDiscovery(logger logging.Logger) *DirectServiceDiscovery {
	return &DirectServiceDiscovery{
		logger:      logger,
		serviceMap:  make(map[string][]ServiceInstance),
		watchCh:     make(map[string][]chan []ServiceInstance),
		initialized: false,
	}
}

func (d *DirectServiceDiscovery) Initialize(ctx context.Context) error {
	d.initialized = true
	d.logger.Info("Direct service discovery initialized")
	return nil
}

func (d *DirectServiceDiscovery) GetService(ctx context.Context, name string) ([]ServiceInstance, error) {
	if !d.initialized {
		return nil, fmt.Errorf("direct service discovery not initialized")
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	instances, exists := d.serviceMap[name]
	if !exists || len(instances) == 0 {
		return nil, ErrServiceNotFound
	}

	return instances, nil
}

func (d *DirectServiceDiscovery) GetInstance(ctx context.Context, name string) (*ServiceInstance, error) {
	instances, err := d.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	var healthyInstances []ServiceInstance
	for _, instance := range instances {
		if instance.Healthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	index := time.Now().UnixNano() % int64(len(healthyInstances))
	return &healthyInstances[index], nil
}

func (d *DirectServiceDiscovery) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	if !d.initialized {
		return fmt.Errorf("direct service discovery not initialized")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	serviceName := instance.Name
	instances, exists := d.serviceMap[serviceName]

	if !exists {
		instances = []ServiceInstance{}
	}

	// Check if instance already exists
	exists = false
	for i, existingInstance := range instances {
		if existingInstance.ID == instance.ID {
			// Update existing instance
			instances[i] = *instance
			exists = true
			break
		}
	}

	// Add new instance if it doesn't exist
	if !exists {
		instances = append(instances, *instance)
	}

	d.serviceMap[serviceName] = instances
	d.notifyWatchers(serviceName, instances)

	return nil
}

func (d *DirectServiceDiscovery) DeregisterService(ctx context.Context, instanceID string) error {
	if !d.initialized {
		return fmt.Errorf("direct service discovery not initialized")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for serviceName, instances := range d.serviceMap {
		for i, instance := range instances {
			if instance.ID == instanceID {
				// Remove instance
				instances = append(instances[:i], instances[i+1:]...)
				d.serviceMap[serviceName] = instances
				d.notifyWatchers(serviceName, instances)
				return nil
			}
		}
	}

	return fmt.Errorf("instance not found: %s", instanceID)
}

func (d *DirectServiceDiscovery) Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error) {
	if !d.initialized {
		return nil, fmt.Errorf("direct service discovery not initialized")
	}

	ch := make(chan []ServiceInstance, 1)

	d.watchChanMu.Lock()
	if _, ok := d.watchCh[name]; !ok {
		d.watchCh[name] = make([]chan []ServiceInstance, 0)
	}
	d.watchCh[name] = append(d.watchCh[name], ch)
	d.watchChanMu.Unlock()

	// Send initial set of instances if available
	instances, err := d.GetService(ctx, name)
	if err == nil && len(instances) > 0 {
		select {
		case ch <- instances:
		default:
		}
	}

	return ch, nil
}

func (d *DirectServiceDiscovery) notifyWatchers(name string, instances []ServiceInstance) {
	d.watchChanMu.RLock()
	defer d.watchChanMu.RUnlock()

	chans, ok := d.watchCh[name]
	if !ok {
		return
	}

	for _, ch := range chans {
		select {
		case ch <- instances:
		default:
		}
	}
}

func (d *DirectServiceDiscovery) GetHealthStatus(ctx context.Context, name string) (*HealthStatus, error) {
	instances, err := d.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	status := &HealthStatus{
		ServiceName:     name,
		InstanceCount:   len(instances),
		HealthyCount:    0,
		UnhealthyCount:  0,
		LastRefreshed:   time.Now(),
		Status:          "unknown",
		InstancesByZone: make(map[string]int),
	}

	for _, instance := range instances {
		if instance.Healthy {
			status.HealthyCount++
		} else {
			status.UnhealthyCount++
		}

		status.InstancesByZone[instance.Zone]++
	}

	if status.HealthyCount == 0 {
		status.Status = "critical"
	} else if status.HealthyCount < status.InstanceCount {
		status.Status = "warning"
	} else {
		status.Status = "healthy"
	}

	return status, nil
}

func (d *DirectServiceDiscovery) Close() error {
	d.watchChanMu.Lock()
	for _, chans := range d.watchCh {
		for _, ch := range chans {
			close(ch)
		}
	}
	d.watchCh = make(map[string][]chan []ServiceInstance)
	d.watchChanMu.Unlock()

	return nil
}

func (d *DirectServiceDiscovery) SetServices(services map[string][]ServiceInstance) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.serviceMap = make(map[string][]ServiceInstance)
	for name, instances := range services {
		d.serviceMap[name] = instances
		d.notifyWatchers(name, instances)
	}
}
