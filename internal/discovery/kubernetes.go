package discovery

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"path/filepath"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubernetesServiceDiscovery struct {
	client          *kubernetes.Clientset
	config          *KubernetesConfig
	logger          logging.Logger
	watchCh         map[string]chan []ServiceInstance
	watchCancelFunc map[string]context.CancelFunc
	mu              sync.RWMutex
	serviceMap      map[string][]ServiceInstance
	lastRefresh     map[string]time.Time
	initialized     bool
}

type KubernetesConfig struct {
	InCluster     bool          `yaml:"in_cluster"`
	KubeConfig    string        `yaml:"kube_config"`
	Namespace     string        `yaml:"namespace"`
	LabelSelector string        `yaml:"label_selector"`
	FieldSelector string        `yaml:"field_selector"`
	Port          int           `yaml:"port"`
	RefreshRate   time.Duration `yaml:"refresh_rate"`
	CacheTTL      time.Duration `yaml:"cache_ttl"`
}

func NewKubernetesServiceDiscovery(config *KubernetesConfig, logger logging.Logger) (*KubernetesServiceDiscovery, error) {
	if config.Namespace == "" {
		config.Namespace = "default"
	}

	if config.RefreshRate == 0 {
		config.RefreshRate = 30 * time.Second
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = 60 * time.Second
	}

	if config.KubeConfig == "" && !config.InCluster {
		if home := homedir.HomeDir(); home != "" {
			config.KubeConfig = filepath.Join(home, ".kube", "config")
		}
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

func (k *KubernetesServiceDiscovery) Initialize(ctx context.Context) error {
	var config *rest.Config
	var err error

	if k.config.InCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("error creating in-cluster config: %w", err)
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", k.config.KubeConfig)
		if err != nil {
			return fmt.Errorf("error building kubeconfig: %w", err)
		}
	}

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

func (k *KubernetesServiceDiscovery) GetService(ctx context.Context, name string) ([]ServiceInstance, error) {
	if !k.initialized {
		return nil, fmt.Errorf("kubernetes service discovery not initialized")
	}

	k.mu.RLock()
	instances, found := k.serviceMap[name]
	lastRefresh, hasRefresh := k.lastRefresh[name]
	k.mu.RUnlock()

	if found && hasRefresh && time.Since(lastRefresh) < k.config.CacheTTL {
		return instances, nil
	}

	service, err := k.client.CoreV1().Services(k.config.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting service from Kubernetes: %w", err)
	}

	if len(service.Spec.Selector) == 0 {
		return nil, fmt.Errorf("service %s has no selectors", name)
	}

	selector := labels.SelectorFromSet(service.Spec.Selector)
	labelSelector := selector.String()
	if k.config.LabelSelector != "" {
		labelSelector = fmt.Sprintf("%s,%s", labelSelector, k.config.LabelSelector)
	}

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

	instances = k.processPodsToInstances(service, pods.Items)

	k.mu.Lock()
	k.serviceMap[name] = instances
	k.lastRefresh[name] = time.Now()
	k.mu.Unlock()

	return instances, nil
}

func (k *KubernetesServiceDiscovery) GetInstance(ctx context.Context, name string) (*ServiceInstance, error) {
	instances, err := k.GetService(ctx, name)
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

func (k *KubernetesServiceDiscovery) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	return fmt.Errorf("service registration not supported in Kubernetes service discovery")
}

func (k *KubernetesServiceDiscovery) DeregisterService(ctx context.Context, instanceID string) error {
	return fmt.Errorf("service deregistration not supported in Kubernetes service discovery")
}

func (k *KubernetesServiceDiscovery) Watch(ctx context.Context, name string) (<-chan []ServiceInstance, error) {
	if !k.initialized {
		return nil, fmt.Errorf("kubernetes service discovery not initialized")
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	if ch, exists := k.watchCh[name]; exists {
		return ch, nil
	}

	ch := make(chan []ServiceInstance, 1)
	k.watchCh[name] = ch

	watchCtx, cancel := context.WithCancel(context.Background())
	k.watchCancelFunc[name] = cancel

	go k.watchService(watchCtx, name, ch)

	return ch, nil
}

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

			select {
			case ch <- instances:
			default:
			}
		}
	}
}

func (k *KubernetesServiceDiscovery) GetHealthStatus(ctx context.Context, name string) (*HealthStatus, error) {
	instances, err := k.GetService(ctx, name)
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

func (k *KubernetesServiceDiscovery) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	for _, cancel := range k.watchCancelFunc {
		cancel()
	}

	k.watchCancelFunc = make(map[string]context.CancelFunc)
	k.watchCh = make(map[string]chan []ServiceInstance)

	return nil
}

func (k *KubernetesServiceDiscovery) processPodsToInstances(service *corev1.Service, pods []corev1.Pod) []ServiceInstance {
	instances := make([]ServiceInstance, 0, len(pods))

	for _, pod := range pods {
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

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

		port := k.config.Port
		if port == 0 && len(service.Spec.Ports) > 0 {
			port = int(service.Spec.Ports[0].Port)
		}

		zone := "default"
		for k, v := range pod.Labels {
			if k == "topology.kubernetes.io/zone" || k == "failure-domain.beta.kubernetes.io/zone" {
				zone = v
				break
			}
		}

		metadata := make(map[string]string)
		for k, v := range pod.Labels {
			metadata[k] = v
		}

		metadata["namespace"] = pod.Namespace
		metadata["node"] = pod.Spec.NodeName
		metadata["pod_ip"] = pod.Status.PodIP
		metadata["host_ip"] = pod.Status.HostIP

		secure := false
		for _, servicePort := range service.Spec.Ports {
			if strings.HasPrefix(servicePort.Name, "https") || strings.HasPrefix(servicePort.Name, "tls") {
				secure = true
				break
			}
		}

		tags := make([]string, 0)
		for k, v := range pod.Annotations {
			tags = append(tags, fmt.Sprintf("%s:%s", k, v))
		}

		version := pod.Labels["version"]
		if version == "" {
			version = pod.Labels["app.kubernetes.io/version"]
		}

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
