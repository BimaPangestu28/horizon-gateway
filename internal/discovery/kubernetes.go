package discovery

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesServiceDiscovery implements service discovery using Kubernetes API
type KubernetesServiceDiscovery struct {
	client          *kubernetes.Clientset
	config          *KubernetesServiceDiscoveryConfig
	logger          logging.Logger
	watchCh         map[string]chan []ServiceInstance
	watchCancelFunc map[string]context.CancelFunc
	mu              sync.RWMutex
	serviceMap      map[string][]ServiceInstance
	lastRefresh     map[string]time.Time
	initialized     bool
}

// KubernetesServiceDiscoveryConfig defines configuration for Kubernetes service discovery
type KubernetesServiceDiscoveryConfig struct {
	InCluster     bool          `yaml:"in_cluster"`
	KubeConfig    string        `yaml:"kube_config"`
	Namespace     string        `yaml:"namespace"`
	LabelSelector string        `yaml:"label_selector"`
	FieldSelector string        `yaml:"field_selector"`
	Port          int           `yaml:"port"`
	RefreshRate   time.Duration `yaml:"refresh_rate"`
	CacheTTL      time.Duration `yaml:"cache_ttl"`
}

// NewKubernetesServiceDiscovery creates a new Kubernetes service discovery client
func NewKubernetesServiceDiscovery(config *KubernetesServiceDiscoveryConfig, logger logging.Logger) (*KubernetesServiceDiscovery, error) {
	if config.Namespace == "" {
		config.Namespace = "default"
	}

	if config.RefreshRate == 0 {
		config.RefreshRate = 30 * time.Second
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = 60 * time.Second
	}

	return &KubernetesServiceDiscovery{
		config:          config,
		logger:          logger,
		watchCh:         make(map[string]chan []ServiceInstance),
		watchCancelFunc: make(map[string]context.CancelFunc),
		serviceMap:      make(map[string][]ServiceInstance),
		lastRefresh:     make(map[string]time.Time),
		initialized:     false,
	}, nil
}

// Initialize initializes the Kubernetes client
func (k *KubernetesServiceDiscovery) Initialize(ctx context.Context) error {
	var config *rest.Config
	var err error

	if k.config.InCluster {
		// For in-cluster deployment
		config, err = rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("error creating in-cluster config: %w", err)
		}
	} else {
		// For outside-cluster deployment
		config, err = clientcmd.BuildConfigFromFlags("", k.config.KubeConfig)
		if err != nil {
			return fmt.Errorf("error building kubeconfig: %w", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes client: %w", err)
	}

	k.client = clientset
	k.initialized = true

	k.logger.Info("Kubernetes service discovery initialized",
		"namespace", k.config.Namespace,
		"in_cluster", k.config.InCluster)

	return nil
}

// GetService returns all instances of a service
func (k *KubernetesServiceDiscovery) GetService(ctx context.Context, name string) ([]ServiceInstance, error) {
	if !k.initialized {
		return nil, fmt.Errorf("kubernetes service discovery not initialized")
	}

	k.mu.RLock()
	instances, found := k.serviceMap[name]
	lastRefresh, hasRefresh := k.lastRefresh[name]
	k.mu.RUnlock()

	// Return cached instances if available and not expired
	if found && hasRefresh && time.Since(lastRefresh) < k.config.CacheTTL {
		return instances, nil
	}

	// Get service from Kubernetes API
	service, err := k.client.CoreV1().Services(k.config.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting service from Kubernetes: %w", err)
	}

	// Get selector for this service
	if len(service.Spec.Selector) == 0 {
		return nil, fmt.Errorf("service %s has no selectors", name)
	}

	selector := labels.SelectorFromSet(service.Spec.Selector)
	labelSelector := selector.String()
	if k.config.LabelSelector != "" {
		labelSelector = fmt.Sprintf("%s,%s", labelSelector, k.config.LabelSelector)
	}

	// List pods that match the service selector
	pods, err := k.client.CoreV1().Pods(k.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: k.config.FieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// Process all pods into service instances
	instances = k.processPodsToInstances(service, pods.Items)

	// Update cache
	k.mu.Lock()
	k.serviceMap[name] = instances
	k.lastRefresh[name] = time.Now()
	k.mu.Unlock()

	return instances, nil
}

// GetInstance returns a single instance of a service
func (k *KubernetesServiceDiscovery) GetInstance(ctx context.Context, name string) (*ServiceInstance, error) {
	instances, err := k.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	// Filter to only healthy instances
	var healthyInstances []ServiceInstance
	for _, instance := range instances {
		if instance.Healthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// Simple round-robin selection - in production we could use a more sophisticated algorithm
	// or delegate to a load balancer
	index := time.Now().UnixNano() % int64(len(healthyInstances))
	return &healthyInstances[index], nil
}

// Watch watches for changes to a service
func (k *KubernetesServiceDiscovery) Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error) {
	if !k.initialized {
		return nil, fmt.Errorf("kubernetes service discovery not initialized")
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if already watching
	if ch, exists := k.watchCh[name]; exists {
		return ch, nil
	}

	// Create new channel and watch
	ch := make(chan []ServiceInstance, 1)
	k.watchCh[name] = ch

	// Create context with cancel for this watch
	watchCtx, cancel := context.WithCancel(context.Background())
	k.watchCancelFunc[name] = cancel

	// Start watch goroutine
	go k.watchService(watchCtx, name, ch)

	return ch, nil
}

// watchService watches a service for changes
func (k *KubernetesServiceDiscovery) watchService(ctx context.Context, name string, ch chan<- []ServiceInstance) {
	ticker := time.NewTicker(k.config.RefreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			k.logger.Info("Stopping watch for service", "service", name)
			return
		case <-ticker.C:
			instances, err := k.GetService(ctx, name)
			if err != nil {
				k.logger.Error("Error refreshing service instances",
					"service", name,
					"error", err)
				continue
			}

			// Send update
			select {
			case ch <- instances:
			default:
				// Channel full, this is ok as it's a buffered channel
				// and we're only interested in the latest state
			}
		}
	}
}

// RegisterService registers a service instance
func (k *KubernetesServiceDiscovery) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	// Kubernetes service registration is handled by Kubernetes itself
	return fmt.Errorf("service registration not supported in Kubernetes service discovery")
}

// DeregisterService deregisters a service instance
func (k *KubernetesServiceDiscovery) DeregisterService(ctx context.Context, instanceID string) error {
	// Kubernetes service deregistration is handled by Kubernetes itself
	return fmt.Errorf("service deregistration not supported in Kubernetes service discovery")
}

// GetHealthStatus returns health status of a service
func (k *KubernetesServiceDiscovery) GetHealthStatus(ctx context.Context, name string) (*HealthStatus, error) {
	instances, err := k.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	// Calculate health status
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

		// Increment zone count
		status.InstancesByZone[instance.Zone]++
	}

	// Determine overall status
	if status.HealthyCount == 0 {
		status.Status = "critical"
	} else if status.HealthyCount < status.InstanceCount {
		status.Status = "warning"
	} else {
		status.Status = "healthy"
	}

	return status, nil
}

// Close closes the service discovery client
func (k *KubernetesServiceDiscovery) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Cancel all watches
	for name, cancel := range k.watchCancelFunc {
		cancel()
		delete(k.watchCancelFunc, name)
	}

	// Clear state
	k.serviceMap = make(map[string][]ServiceInstance)
	k.lastRefresh = make(map[string]time.Time)

	return nil
}

// processPodsToInstances converts Kubernetes pods to service instances
func (k *KubernetesServiceDiscovery) processPodsToInstances(service *corev1.Service, pods []corev1.Pod) []ServiceInstance {
	instances := make([]ServiceInstance, 0, len(pods))

	for _, pod := range pods {
		// Skip pods that aren't running
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		// Check pod readiness
		isReady := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				isReady = true
				break
			}
		}

		if !isReady {
			continue
		}

		// Determine port
		port := k.config.Port
		if port == 0 && len(service.Spec.Ports) > 0 {
			port = int(service.Spec.Ports[0].Port)
		}

		// Determine zone
		zone := "default"
		for _, node := range pod.Spec.NodeSelector {
			if node == "topology.kubernetes.io/zone" || node == "failure-domain.beta.kubernetes.io/zone" {
				zone = pod.Spec.NodeSelector[node]
				break
			}
		}

		// Get metadata from pod labels
		metadata := make(map[string]string)
		for k, v := range pod.Labels {
			metadata[k] = v
		}

		// Add some additional metadata
		metadata["namespace"] = pod.Namespace
		metadata["node"] = pod.Spec.NodeName
		metadata["pod_ip"] = pod.Status.PodIP
		metadata["host_ip"] = pod.Status.HostIP

		// Determine if secure based on port name convention
		secure := false
		for _, servicePort := range service.Spec.Ports {
			if strings.HasPrefix(servicePort.Name, "https") || strings.HasPrefix(servicePort.Name, "tls") {
				secure = true
				break
			}
		}

		// Create tags from annotations
		tags := make([]string, 0)
		for k, v := range pod.Annotations {
			tags = append(tags, fmt.Sprintf("%s:%s", k, v))
		}

		// Extract version from labels
		version := pod.Labels["version"]
		if version == "" {
			version = pod.Labels["app.kubernetes.io/version"]
		}

		// Get weight from annotation or default to 1
		weight := 1
		if weightStr, ok := pod.Annotations["weight"]; ok {
			if w, err := strconv.Atoi(weightStr); err == nil && w > 0 {
				weight = w
			}
		}

		instance := ServiceInstance{
			ID:        string(pod.UID),
			Name:      service.Name,
			Address:   pod.Status.PodIP,
			Port:      port,
			Secure:    secure,
			Metadata:  metadata,
			Tags:      tags,
			Healthy:   isReady,
			Weight:    weight,
			Zone:      zone,
			Version:   version,
			LastCheck: time.Now(),
		}

		instances = append(instances, instance)
	}

	return instances
}
