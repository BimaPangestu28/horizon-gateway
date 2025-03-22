package discovery

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/hashicorp/consul/api"
)

type ConsulDiscovery struct {
	client           *api.Client
	config           *ConsulConfig
	logger           logging.Logger
	cache            map[string][]ServiceInstance
	cacheMu          sync.RWMutex
	lastRefresh      map[string]time.Time
	watchCh          map[string]chan []ServiceInstance
	watchCancelFuncs map[string]context.CancelFunc
	watchMu          sync.Mutex
}

type ConsulConfig struct {
	Address     string
	Datacenter  string
	Token       string
	TagFilter   string
	HealthCheck bool
	CacheTTL    time.Duration
}

func NewConsulDiscovery(config *ConsulConfig, logger logging.Logger) *ConsulDiscovery {
	if config.CacheTTL == 0 {
		config.CacheTTL = 30 * time.Second
	}

	return &ConsulDiscovery{
		config:           config,
		logger:           logger,
		cache:            make(map[string][]ServiceInstance),
		lastRefresh:      make(map[string]time.Time),
		watchCh:          make(map[string]chan []ServiceInstance),
		watchCancelFuncs: make(map[string]context.CancelFunc),
	}
}

func (d *ConsulDiscovery) Initialize(ctx context.Context) error {
	consulConfig := api.DefaultConfig()

	if d.config.Address != "" {
		consulConfig.Address = d.config.Address
	}

	if d.config.Datacenter != "" {
		consulConfig.Datacenter = d.config.Datacenter
	}

	if d.config.Token != "" {
		consulConfig.Token = d.config.Token
	}

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return fmt.Errorf("failed to create Consul client: %w", err)
	}

	d.client = client

	d.logger.Info("Consul service discovery initialized",
		"address", consulConfig.Address,
		"datacenter", consulConfig.Datacenter)

	return nil
}

func (d *ConsulDiscovery) GetService(ctx context.Context, name string) ([]ServiceInstance, error) {
	d.cacheMu.RLock()
	instances, found := d.cache[name]
	lastRefresh, _ := d.lastRefresh[name]
	d.cacheMu.RUnlock()

	if found && time.Since(lastRefresh) < d.config.CacheTTL {
		return instances, nil
	}

	var queryOptions api.QueryOptions
	queryOptions.WithContext(ctx)

	var serviceEntries []*api.ServiceEntry
	var err error

	if d.config.HealthCheck {
		serviceEntries, _, err = d.client.Health().Service(name, d.config.TagFilter, true, &queryOptions)
	} else {
		var services []*api.CatalogService
		services, _, err = d.client.Catalog().Service(name, d.config.TagFilter, &queryOptions)
		if err == nil {
			serviceEntries = make([]*api.ServiceEntry, len(services))
			for i, svc := range services {
				serviceEntries[i] = &api.ServiceEntry{
					Node: &api.Node{
						Node:    svc.Node,
						Address: svc.Address,
					},
					Service: &api.AgentService{
						ID:      svc.ServiceID,
						Service: svc.ServiceName,
						Tags:    svc.ServiceTags,
						Port:    svc.ServicePort,
						Address: svc.ServiceAddress,
					},
				}
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error getting service from Consul: %w", err)
	}

	if len(serviceEntries) == 0 {
		return nil, ErrServiceNotFound
	}

	instances = make([]ServiceInstance, 0, len(serviceEntries))

	for _, entry := range serviceEntries {
		address := entry.Service.Address
		if address == "" {
			address = entry.Node.Address
		}

		metadata := make(map[string]string)
		for k, v := range entry.Service.Meta {
			metadata[k] = v
		}

		weight := 1 // Default weight
		if wStr, ok := metadata["weight"]; ok {
			if w, err := strconv.Atoi(wStr); err == nil && w > 0 {
				weight = w
			}
		}

		zone := "" // Default zone
		if z, ok := metadata["zone"]; ok {
			zone = z
		} else if z, ok := metadata["datacenter"]; ok {
			zone = z
		} else {
			zone = entry.Node.Datacenter
		}

		version := "" // Default version
		if v, ok := metadata["version"]; ok {
			version = v
		}

		healthy := true
		if d.config.HealthCheck {
			healthy = isHealthy(entry)
		}

		instance := ServiceInstance{
			ID:        entry.Service.ID,
			Name:      entry.Service.Service,
			Address:   address,
			Port:      entry.Service.Port,
			Secure:    hasSecureTag(entry.Service.Tags),
			Metadata:  metadata,
			Tags:      entry.Service.Tags,
			Healthy:   healthy,
			Weight:    weight,
			Zone:      zone,
			Version:   version,
			LastCheck: time.Now(),
		}

		instances = append(instances, instance)
	}

	d.cacheMu.Lock()
	d.cache[name] = instances
	d.lastRefresh[name] = time.Now()
	d.cacheMu.Unlock()

	return instances, nil
}

func (d *ConsulDiscovery) GetInstance(ctx context.Context, name string) (*ServiceInstance, error) {
	instances, err := d.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	// Filter healthy instances
	var healthyInstances []ServiceInstance
	for _, instance := range instances {
		if instance.Healthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// Return random instance (this could be replaced with more sophisticated selection)
	return &healthyInstances[0], nil
}

func (d *ConsulDiscovery) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	registration := &api.AgentServiceRegistration{
		ID:      instance.ID,
		Name:    instance.Name,
		Address: instance.Address,
		Port:    instance.Port,
		Tags:    instance.Tags,
		Meta:    instance.Metadata,
	}

	if instance.Secure {
		registration.Tags = append(registration.Tags, "secure")
	}

	// Add health check if enabled
	if d.config.HealthCheck {
		registration.Check = &api.AgentServiceCheck{
			HTTP:          fmt.Sprintf("http://%s:%d/health", instance.Address, instance.Port),
			Interval:      "10s",
			Timeout:       "5s",
			TLSSkipVerify: true,
		}
	}

	err := d.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service with Consul: %w", err)
	}

	return nil
}

func (d *ConsulDiscovery) DeregisterService(ctx context.Context, instanceID string) error {
	err := d.client.Agent().ServiceDeregister(instanceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service from Consul: %w", err)
	}

	return nil
}

func (d *ConsulDiscovery) Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error) {
	d.watchMu.Lock()
	defer d.watchMu.Unlock()

	// Check if a watch already exists
	if ch, exists := d.watchCh[name]; exists {
		return ch, nil
	}

	// Create a new channel for the watch
	ch := make(chan []ServiceInstance, 1)
	d.watchCh[name] = ch

	// Create a new context with cancel for the watch
	watchCtx, cancel := context.WithCancel(context.Background())
	d.watchCancelFuncs[name] = cancel

	var lastIndex uint64

	// Start the watch goroutine
	go func() {
		defer close(ch)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-watchCtx.Done():
				return
			case <-ticker.C:
				var queryOptions api.QueryOptions
				queryOptions.WithContext(watchCtx)
				queryOptions.WaitIndex = lastIndex

				services, meta, err := d.client.Health().Service(name, d.config.TagFilter, true, &queryOptions)
				if err != nil {
					d.logger.Error("Error watching service", "name", name, "error", err)
					continue
				}

				if meta.LastIndex == lastIndex {
					continue
				}

				lastIndex = meta.LastIndex

				// Convert services to instances
				instances := make([]ServiceInstance, 0, len(services))
				for _, entry := range services {
					address := entry.Service.Address
					if address == "" {
						address = entry.Node.Address
					}

					metadata := make(map[string]string)
					for k, v := range entry.Service.Meta {
						metadata[k] = v
					}

					instance := ServiceInstance{
						ID:        entry.Service.ID,
						Name:      entry.Service.Service,
						Address:   address,
						Port:      entry.Service.Port,
						Secure:    hasSecureTag(entry.Service.Tags),
						Metadata:  metadata,
						Tags:      entry.Service.Tags,
						Healthy:   isHealthy(entry),
						LastCheck: time.Now(),
					}

					instances = append(instances, instance)
				}

				// Update cache
				d.cacheMu.Lock()
				d.cache[name] = instances
				d.lastRefresh[name] = time.Now()
				d.cacheMu.Unlock()

				// Send update
				select {
				case ch <- instances:
				default:
					// If the channel buffer is full, wait and try again
					select {
					case <-watchCtx.Done():
						return
					case ch <- instances:
					}
				}
			}
		}
	}()

	return ch, nil
}

func (d *ConsulDiscovery) GetHealthStatus(ctx context.Context, name string) (*HealthStatus, error) {
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

func (d *ConsulDiscovery) Close() error {
	d.watchMu.Lock()
	defer d.watchMu.Unlock()

	for _, cancel := range d.watchCancelFuncs {
		cancel()
	}

	return nil
}

func isHealthy(entry *api.ServiceEntry) bool {
	if entry == nil {
		return false
	}

	for _, check := range entry.Checks {
		if check.Status == api.HealthCritical {
			return false
		}
	}

	return true
}

func hasSecureTag(tags []string) bool {
	for _, tag := range tags {
		if strings.ToLower(tag) == "secure" || strings.ToLower(tag) == "ssl" || strings.ToLower(tag) == "tls" {
			return true
		}
	}
	return false
}
