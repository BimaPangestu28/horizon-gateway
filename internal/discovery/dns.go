package discovery

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type DNSServiceDiscovery struct {
	config        *DNSConfig
	logger        logging.Logger
	resolver      *net.Resolver
	cache         map[string][]ServiceInstance
	cacheMu       sync.RWMutex
	lastRefresh   map[string]time.Time
	refreshTicker *time.Ticker
	stopCh        chan struct{}
	watchChans    map[string][]chan []ServiceInstance
	watchChanMu   sync.RWMutex
	initialized   bool
}

type DNSConfig struct {
	Domain        string        `yaml:"domain"`
	QueryType     string        `yaml:"query_type"`
	Port          int           `yaml:"port"`
	ResolverAddr  string        `yaml:"resolver"`
	RefreshRate   time.Duration `yaml:"refresh_rate"`
	CacheTTL      time.Duration `yaml:"cache_ttl"`
	DefaultWeight int           `yaml:"default_weight"`
}

func NewDNSServiceDiscovery(config *DNSConfig, logger logging.Logger) (*DNSServiceDiscovery, error) {
	if config.Domain == "" {
		return nil, fmt.Errorf("DNS domain is required")
	}

	if config.QueryType == "" {
		config.QueryType = "A"
	}

	if config.Port <= 0 {
		config.Port = 80
	}

	if config.RefreshRate == 0 {
		config.RefreshRate = 30 * time.Second
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = 60 * time.Second
	}

	if config.DefaultWeight == 0 {
		config.DefaultWeight = 1
	}

	var resolver *net.Resolver
	if config.ResolverAddr != "" {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialer := net.Dialer{
					Timeout: 5 * time.Second,
				}
				return dialer.DialContext(ctx, "udp", config.ResolverAddr)
			},
		}
	} else {
		resolver = net.DefaultResolver
	}

	return &DNSServiceDiscovery{
		config:      config,
		logger:      logger,
		resolver:    resolver,
		cache:       make(map[string][]ServiceInstance),
		lastRefresh: make(map[string]time.Time),
		stopCh:      make(chan struct{}),
		watchChans:  make(map[string][]chan []ServiceInstance),
	}, nil
}

func (d *DNSServiceDiscovery) Initialize(ctx context.Context) error {
	d.initialized = true
	d.refreshTicker = time.NewTicker(d.config.RefreshRate)

	go d.refreshLoop()

	d.logger.Info("DNS service discovery initialized",
		"domain", d.config.Domain,
		"query_type", d.config.QueryType,
		"refresh_rate", d.config.RefreshRate)

	return nil
}

func (d *DNSServiceDiscovery) refreshLoop() {
	for {
		select {
		case <-d.stopCh:
			return
		case <-d.refreshTicker.C:
			d.refreshAllServices()
		}
	}
}

func (d *DNSServiceDiscovery) refreshAllServices() {
	d.cacheMu.RLock()
	services := make([]string, 0, len(d.cache))
	for service := range d.cache {
		services = append(services, service)
	}
	d.cacheMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, service := range services {
		_, err := d.GetService(ctx, service)
		if err != nil {
			d.logger.Error("Error refreshing service",
				"service", service,
				"error", err)
		}
	}
}

func (d *DNSServiceDiscovery) GetService(ctx context.Context, name string) ([]ServiceInstance, error) {
	if !d.initialized {
		return nil, fmt.Errorf("DNS service discovery not initialized")
	}

	d.cacheMu.RLock()
	instances, found := d.cache[name]
	lastRefresh, _ := d.lastRefresh[name]
	d.cacheMu.RUnlock()

	if found && time.Since(lastRefresh) < d.config.CacheTTL {
		return instances, nil
	}

	var host string
	if strings.HasSuffix(name, "."+d.config.Domain) || name == d.config.Domain {
		host = name
	} else {
		host = name + "." + d.config.Domain
	}

	addrs, err := d.resolveHost(ctx, host)
	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		return nil, ErrServiceNotFound
	}

	instances := make([]ServiceInstance, 0, len(addrs))
	for i, addr := range addrs {
		instance := ServiceInstance{
			ID:        fmt.Sprintf("%s-%d", name, i),
			Name:      name,
			Address:   addr,
			Port:      d.config.Port,
			Healthy:   true,
			Weight:    d.config.DefaultWeight,
			Zone:      "default",
			LastCheck: time.Now(),
			Metadata: map[string]string{
				"discovery": "dns",
				"host":      host,
			},
		}
		instances = append(instances, instance)
	}

	d.cacheMu.Lock()
	d.cache[name] = instances
	d.lastRefresh[name] = time.Now()
	d.cacheMu.Unlock()

	d.notifyWatchers(name, instances)

	return instances, nil
}

func (d *DNSServiceDiscovery) GetInstance(ctx context.Context, name string) (*ServiceInstance, error) {
	instances, err := d.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	index := time.Now().UnixNano() % int64(len(instances))
	return &instances[index], nil
}

func (d *DNSServiceDiscovery) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	return fmt.Errorf("service registration not supported in DNS service discovery")
}

func (d *DNSServiceDiscovery) DeregisterService(ctx context.Context, instanceID string) error {
	return fmt.Errorf("service deregistration not supported in DNS service discovery")
}

func (d *DNSServiceDiscovery) Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error) {
	if !d.initialized {
		return nil, fmt.Errorf("DNS service discovery not initialized")
	}

	ch := make(chan []ServiceInstance, 1)

	d.watchChanMu.Lock()
	if _, ok := d.watchChans[name]; !ok {
		d.watchChans[name] = make([]chan []ServiceInstance, 0)
	}
	d.watchChans[name] = append(d.watchChans[name], ch)
	d.watchChanMu.Unlock()

	instances, err := d.GetService(ctx, name)
	if err == nil && len(instances) > 0 {
		select {
		case ch <- instances:
		default:
		}
	}

	return ch, nil
}

func (d *DNSServiceDiscovery) notifyWatchers(name string, instances []ServiceInstance) {
	d.watchChanMu.RLock()
	defer d.watchChanMu.RUnlock()

	chans, ok := d.watchChans[name]
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

func (d *DNSServiceDiscovery) GetHealthStatus(ctx context.Context, name string) (*HealthStatus, error) {
	instances, err := d.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	status := &HealthStatus{
		ServiceName:     name,
		InstanceCount:   len(instances),
		HealthyCount:    len(instances), // All DNS instances are considered healthy
		UnhealthyCount:  0,
		LastRefreshed:   time.Now(),
		Status:          "healthy",
		InstancesByZone: make(map[string]int),
	}

	for _, instance := range instances {
		status.InstancesByZone[instance.Zone]++
	}

	return status, nil
}

func (d *DNSServiceDiscovery) Close() error {
	close(d.stopCh)
	if d.refreshTicker != nil {
		d.refreshTicker.Stop()
	}

	d.watchChanMu.Lock()
	for _, chans := range d.watchChans {
		for _, ch := range chans {
			close(ch)
		}
	}
	d.watchChans = make(map[string][]chan []ServiceInstance)
	d.watchChanMu.Unlock()

	return nil
}
