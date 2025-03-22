package discovery

import (
	"context"
	"fmt"
	"strconv"
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

type KubernetesDiscovery struct {
	client           *kubernetes.Clientset
	config           *KubernetesConfig
	logger           logging.Logger
	cache            map[string][]ServiceInstance
	cacheMu          sync.RWMutex
	lastRefresh      map[string]time.Time
	watchCh          map[string]chan []ServiceInstance
	watchCancelFuncs map[string]context.CancelFunc
	watchMu          sync.Mutex
}

type KubernetesConfig struct {
	InCluster     bool
	KubeConfig    string
	Namespace     string
	LabelSelector string
	FieldSelector string
	PortName      string
	DefaultPort   int
	CacheTTL      time.Duration
}

func NewKubernetesDiscovery(config *KubernetesConfig, logger logging.Logger) *KubernetesDiscovery {
	if config.CacheTTL == 0 {
		config.CacheTTL = 30 * time.Second
	}

	if config.Namespace == "" {
		config.Namespace = "default"
	}

	if config.DefaultPort == 0 {
		config.DefaultPort = 80
	}

	return &KubernetesDiscovery{
		config:           config,
		logger:           logger,
		cache:            make(map[string][]ServiceInstance),
		lastRefresh:      make(map[string]time.Time),
		watchCh:          make(map[string]chan []ServiceInstance),
		watchCancelFuncs: make(map[string]context.CancelFunc),
	}
}

func (d *KubernetesDiscovery) Initialize(ctx context.Context) error {
	var config *rest.Config
	var err error

	if d.config.InCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", d.config.KubeConfig)
		if err != nil {
			return fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	d.client = clientset

	d.logger.Info("Kubernetes service discovery initialized",
		"in_cluster", d.config.InCluster,
		"namespace", d.config.Namespace)

	return nil
}

func (d *KubernetesDiscovery) GetService(ctx context.Context, name string) ([]ServiceInstance, error) {
	d.cacheMu.RLock()
	instances, found := d.cache[name]
	lastRefresh, _ := d.lastRefresh[name]
	d.cacheMu.RUnlock()

	if found && time.Since(lastRefresh) < d.config.CacheTTL {
		return instances, nil
	}

	// Get service
	service, err := d.client.CoreV1().Services(d.config.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// Get service selector
	selector := labels.SelectorFromSet(service.Spec.Selector)
	if selector.Empty() {
		return nil, fmt.Errorf("service has no selector")
	}

	// Get pods
	pods, err := d.client.CoreV1().Pods(d.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
		FieldSelector: d.config.FieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, ErrNoHealthyInstances
	}

	instances = d.podsToInstances(service, pods.Items)

	d.cacheMu.Lock()
	d.cache[name] = instances
	d.lastRefresh[name] = time.Now()
	d.cacheMu.Unlock()

	return instances, nil
}

func (d *KubernetesDiscovery) GetInstance(ctx context.Context, name string) (*ServiceInstance, error) {
	instances, err := d.GetService(ctx, name)
	if err != nil {
		return nil, err
	}

	// Filter running pods
	var runningInstances []ServiceInstance
	for _, instance := range instances {
		if instance.Healthy {
			runningInstances = append(runningInstances, instance)
		}
	}

	if len(runningInstances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// Return first instance (in production, use better selection strategy)
	return &runningInstances[0], nil
}

func (d *KubernetesDiscovery) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	return fmt.Errorf("service registration not supported in Kubernetes service discovery")
}

func (d *KubernetesDiscovery) DeregisterService(ctx context.Context, instanceID string) error {
	return fmt.Errorf("service deregistration not supported in Kubernetes service discovery")
}

func (d *KubernetesDiscovery) Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error) {
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

	// Start the watch goroutine
	go func() {
		defer close(ch)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-watchCtx.Done():
				return
			case <-ticker.C:
				instances, err := d.GetService(watchCtx, name)
				if err != nil {
					d.logger.Error("Error watching service", "name", name, "error", err)
					continue
				}

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

func (d *KubernetesDiscovery) GetHealthStatus(ctx context.Context, name string) (*HealthStatus, error) {
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

func (d *KubernetesDiscovery) Close() error {
	d.watchMu.Lock()
	defer d.watchMu.Unlock()

	for _, cancel := range d.watchCancelFuncs {
		cancel()
	}

	return nil
}

func (d *KubernetesDiscovery) podsToInstances(service *corev1.Service, pods []corev1.Pod) []ServiceInstance {
	instances := make([]ServiceInstance, 0, len(pods))

	for _, pod := range pods {
		if !isPodReady(&pod) {
			continue
		}

		// Determine port
		port := d.config.DefaultPort
		if d.config.PortName != "" {
			for _, container := range pod.Spec.Containers {
				for _, containerPort := range container.Ports {
					if containerPort.Name == d.config.PortName {
						port = int(containerPort.ContainerPort)
						break
					}
				}
			}
		} else if len(service.Spec.Ports) > 0 {
			port = int(service.Spec.Ports[0].Port)
		}

		// Get pod IP
		address := pod.Status.PodIP

		// Build metadata
		metadata := make(map[string]string)
		for k, v := range pod.Labels {
			metadata[k] = v
		}

		metadata["node"] = pod.Spec.NodeName
		metadata["namespace"] = pod.Namespace

		// Get version from label
		version := pod.Labels["version"]
		if version == "" {
			version = pod.Labels["app.kubernetes.io/version"]
		}

		// Determine zone from node labels
		zone := "default"
		if pod.Spec.NodeName != "" {
			if node, err := d.client.CoreV1().Nodes().Get(context.Background(), pod.Spec.NodeName, metav1.GetOptions{}); err == nil {
				if z, ok := node.Labels["topology.kubernetes.io/zone"]; ok {
					zone = z
				} else if z, ok := node.Labels["failure-domain.beta.kubernetes.io/zone"]; ok {
					zone = z
				}
			}
		}

		// Get weight from annotation or label
		weight := 1
		if w, ok := pod.Annotations["weight"]; ok {
			if wInt, err := strconv.Atoi(w); err == nil && wInt > 0 {
				weight = wInt
			}
		} else if w, ok := pod.Labels["weight"]; ok {
			if wInt, err := strconv.Atoi(w); err == nil && wInt > 0 {
				weight = wInt
			}
		}

		instance := ServiceInstance{
			ID:        string(pod.UID),
			Name:      service.Name,
			Address:   address,
			Port:      port,
			Secure:    service.Spec.Ports[0].Name == "https",
			Metadata:  metadata,
			Tags:      []string{},
			Healthy:   isPodReady(&pod),
			Weight:    weight,
			Zone:      zone,
			Version:   version,
			LastCheck: time.Now(),
		}

		instances = append(instances, instance)
	}

	return instances
}

func isPodReady(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
