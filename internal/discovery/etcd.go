package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"
)

type EtcdServiceDiscovery struct {
	client          *clientv3.Client
	config          *EtcdConfig
	logger          logging.Logger
	watchCh         map[string]chan []ServiceInstance
	watchCancelFunc map[string]context.CancelFunc
	mu              sync.RWMutex
	serviceMap      map[string][]ServiceInstance
	lastRefresh     map[string]time.Time
	initialized     bool
}

type EtcdConfig struct {
	Endpoints        []string      `yaml:"endpoints"`
	Username         string        `yaml:"username"`
	Password         string        `yaml:"password"`
	KeyPrefix        string        `yaml:"key_prefix"`
	TLS              *TLSConfig    `yaml:"tls"`
	DialTimeout      time.Duration `yaml:"dial_timeout"`
	RequestTimeout   time.Duration `yaml:"request_timeout"`
	AutoSyncInterval time.Duration `yaml:"auto_sync_interval"`
	CacheTTL         time.Duration `yaml:"cache_ttl"`
}

func NewEtcdServiceDiscovery(config *EtcdConfig, logger logging.Logger) (*EtcdServiceDiscovery, error) {
	if len(config.Endpoints) == 0 {
		config.Endpoints = []string{"localhost:2379"}
	}

	if config.KeyPrefix == "" {
		config.KeyPrefix = "/services"
	}

	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}

	if config.RequestTimeout == 0 {
		config.RequestTimeout = 3 * time.Second
	}

	if config.AutoSyncInterval == 0 {
		config.AutoSyncInterval = 1 * time.Minute
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = 30 * time.Second
	}

	return &EtcdServiceDiscovery{
		config:          config,
		logger:          logger,
		watchCh:         make(map[string]chan []ServiceInstance),
		watchCancelFunc: make(map[string]context.CancelFunc),
		serviceMap:      make(map[string][]ServiceInstance),
		lastRefresh:     make(map[string]time.Time),
		initialized:     false,
	}, nil
}

func (e *EtcdServiceDiscovery) Initialize(ctx context.Context) error {
	etcdConfig := clientv3.Config{
		Endpoints:            e.config.Endpoints,
		Username:             e.config.Username,
		Password:             e.config.Password,
		DialTimeout:          e.config.DialTimeout,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		AutoSyncInterval:     e.config.AutoSyncInterval,
	}

	if e.config.TLS != nil && e.config.TLS.Enabled {
		tlsConfig, err := createTLSConfigFromFiles(e.config.TLS.CertFile, e.config.TLS.KeyFile, e.config.TLS.CAFile)
		if err != nil {
			return fmt.Errorf("error creating TLS config: %w", err)
		}
		etcdConfig.TLS = tlsConfig
	}

	client, err := clientv3.New(etcdConfig)
	if err != nil {
		return fmt.Errorf("error creating etcd client: %w", err)
	}

	if e.config.KeyPrefix != "" {
		client.KV = namespace.NewKV(client.KV, e.config.KeyPrefix)
		client.Watcher = namespace.NewWatcher(client.Watcher, e.config.KeyPrefix)
	}

	e.client = client
	e.initialized = true

	e.logger.Info("Etcd service discovery initialized",
		"endpoints", strings.Join(e.config.Endpoints, ","),
		"key_prefix", e.config.KeyPrefix)

	return nil
}

func (e *EtcdServiceDiscovery) GetService(ctx context.Context, name string) ([]ServiceInstance, error) {
	if !e.initialized {
		return nil, fmt.Errorf("etcd service discovery not initialized")
	}

	e.mu.RLock()
	instances, found := e.serviceMap[name]
	lastRefresh, hasRefresh := e.lastRefresh[name]
	e.mu.RUnlock()

	if found && hasRefresh && time.Since(lastRefresh) < e.config.CacheTTL {
		return instances, nil
	}

	servicePath := path.Join("/", name)
	timeoutCtx, cancel := context.WithTimeout(ctx, e.config.RequestTimeout)
	defer cancel()

	resp, err := e.client.Get(timeoutCtx, servicePath, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("error getting service from etcd: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrServiceNotFound
	}

	var serviceInstances []ServiceInstance
	for _, kv := range resp.Kvs {
		instance, err := e.parseInstanceData(kv.Value)
		if err != nil {
			e.logger.Error("Error parsing instance data",
				"key", string(kv.Key),
				"error", err)
			continue
		}

		serviceInstances = append(serviceInstances, *instance)
	}

	if len(serviceInstances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	e.mu.Lock()
	e.serviceMap[name] = serviceInstances
	e.lastRefresh[name] = time.Now()
	e.mu.Unlock()

	return serviceInstances, nil
}

func (e *EtcdServiceDiscovery) GetInstance(ctx context.Context, name string) (*ServiceInstance, error) {
	instances, err := e.GetService(ctx, name)
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

func (e *EtcdServiceDiscovery) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	if !e.initialized {
		return fmt.Errorf("etcd service discovery not initialized")
	}

	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("error marshaling instance data: %w", err)
	}

	instancePath := path.Join("/", instance.Name, instance.ID)
	timeoutCtx, cancel := context.WithTimeout(ctx, e.config.RequestTimeout)
	defer cancel()

	var leaseID clientv3.LeaseID
	if instance.TTL > 0 {
		lease, err := e.client.Grant(timeoutCtx, int64(instance.TTL.Seconds()))
		if err != nil {
			return fmt.Errorf("error creating lease: %w", err)
		}
		leaseID = lease.ID
	}

	var opts []clientv3.OpOption
	if leaseID != 0 {
		opts = append(opts, clientv3.WithLease(leaseID))
	}

	_, err = e.client.Put(timeoutCtx, instancePath, string(data), opts...)
	if err != nil {
		return fmt.Errorf("error registering service in etcd: %w", err)
	}

	return nil
}

func (e *EtcdServiceDiscovery) DeregisterService(ctx context.Context, instanceID string) error {
	if !e.initialized {
		return fmt.Errorf("etcd service discovery not initialized")
	}

	parts := strings.Split(instanceID, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid instance ID format: %s", instanceID)
	}

	serviceName := parts[0]
	id := parts[1]
	instancePath := path.Join("/", serviceName, id)

	timeoutCtx, cancel := context.WithTimeout(ctx, e.config.RequestTimeout)
	defer cancel()

	_, err := e.client.Delete(timeoutCtx, instancePath)
	if err != nil {
		return fmt.Errorf("error deregistering service from etcd: %w", err)
	}

	return nil
}

func (e *EtcdServiceDiscovery) Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error) {
	if !e.initialized {
		return nil, fmt.Errorf("etcd service discovery not initialized")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if ch, exists := e.watchCh[name]; exists {
		return ch, nil
	}

	ch := make(chan []ServiceInstance, 1)
	e.watchCh[name] = ch

	watchCtx, cancel := context.WithCancel(context.Background())
	e.watchCancelFunc[name] = cancel

	servicePath := path.Join("/", name)
	watchChan := e.client.Watch(watchCtx, servicePath, clientv3.WithPrefix())

	go e.watchHandler(watchCtx, watchChan, name, ch)

	return ch, nil
}

func (e *EtcdServiceDiscovery) watchHandler(ctx context.Context, watchChan clientv3.WatchChan, serviceName string, ch chan<- []ServiceInstance) {
	defer close(ch)

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Stopping watch for service", "service", serviceName)
			return
		case watchResp := <-watchChan:
			if watchResp.Canceled {
				e.logger.Warn("Watch canceled", "service", serviceName)
				return
			}

			if err := watchResp.Err(); err != nil {
				e.logger.Error("Watch error", "service", serviceName, "error", err)
				continue
			}

			e.logger.Debug("Received watch event", "service", serviceName, "event_count", len(watchResp.Events))

			if len(watchResp.Events) > 0 {
				instances, err := e.GetService(ctx, serviceName)
				if err != nil {
					e.logger.Error("Error refreshing service instances",
						"service", serviceName,
						"error", err)
					continue
				}

				select {
				case ch <- instances:
				default:
				}
			}
		}
	}
}

func (e *EtcdServiceDiscovery) GetHealthStatus(ctx context.Context, name string) (*HealthStatus, error) {
	instances, err := e.GetService(ctx, name)
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

func (e *EtcdServiceDiscovery) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, cancel := range e.watchCancelFunc {
		cancel()
	}

	e.watchCancelFunc = make(map[string]context.CancelFunc)
	e.watchCh = make(map[string]chan []ServiceInstance)

	if e.client != nil {
		return e.client.Close()
	}

	return nil
}

func (e *EtcdServiceDiscovery) parseInstanceData(data []byte) (*ServiceInstance, error) {
	var instance ServiceInstance
	if err := json.Unmarshal(data, &instance); err != nil {
		return nil, fmt.Errorf("error unmarshaling instance data: %w", err)
	}

	return &instance, nil
}
