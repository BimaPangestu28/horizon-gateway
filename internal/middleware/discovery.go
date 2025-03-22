package middleware

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/discovery"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type ServiceDiscoveryMiddleware struct {
	discoveryClient discovery.ServiceDiscovery
	cache           map[string][]discovery.ServiceInstance
	cacheLock       sync.RWMutex
	cacheTTL        time.Duration
	lastRefresh     map[string]time.Time
	logger          logging.Logger
}

func NewServiceDiscoveryMiddleware(discoveryClient discovery.ServiceDiscovery, cacheTTL time.Duration, logger logging.Logger) *ServiceDiscoveryMiddleware {
	if cacheTTL == 0 {
		cacheTTL = 30 * time.Second
	}

	middleware := &ServiceDiscoveryMiddleware{
		discoveryClient: discoveryClient,
		cache:           make(map[string][]discovery.ServiceInstance),
		lastRefresh:     make(map[string]time.Time),
		cacheTTL:        cacheTTL,
		logger:          logger,
	}

	return middleware
}

func ServiceDiscoveryHandler(sdm *ServiceDiscoveryMiddleware) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if sdm == nil || sdm.discoveryClient == nil {
			return c.Next()
		}

		routeData, ok := c.Locals("route").(map[string]interface{})
		if !ok {
			return c.Next()
		}

		serviceName, ok := routeData["service_name"].(string)
		if !ok || serviceName == "" {
			return c.Next()
		}

		instances, err := sdm.getServiceInstances(c.Context(), serviceName)
		if err != nil {
			sdm.logger.Error("Failed to get service instances",
				"service", serviceName,
				"error", err,
			)

			if strings.Contains(err.Error(), "not found") {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": fmt.Sprintf("Service '%s' not found", serviceName),
				})
			}

			if strings.Contains(err.Error(), "no healthy instances") {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"error": fmt.Sprintf("No healthy instances available for service '%s'", serviceName),
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Service discovery error: %s", err.Error()),
			})
		}

		if len(instances) == 0 {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fmt.Sprintf("No instances available for service '%s'", serviceName),
			})
		}

		// Store service instances in context for use by load balancer
		c.Locals("service_instances", instances)

		// Continue with next handlers
		return c.Next()
	}
}

func (m *ServiceDiscoveryMiddleware) getServiceInstances(ctx context.Context, serviceName string) ([]discovery.ServiceInstance, error) {
	m.cacheLock.RLock()
	instances, found := m.cache[serviceName]
	lastRefresh, _ := m.lastRefresh[serviceName]
	m.cacheLock.RUnlock()

	// If in cache and not expired, return cached instances
	if found && time.Since(lastRefresh) < m.cacheTTL {
		return instances, nil
	}

	// Get fresh instances from discovery service
	freshInstances, err := m.discoveryClient.GetService(ctx, serviceName)
	if err != nil {
		return instances, err // Return cached instances if available, even if expired
	}

	// Update cache
	m.cacheLock.Lock()
	m.cache[serviceName] = freshInstances
	m.lastRefresh[serviceName] = time.Now()
	m.cacheLock.Unlock()

	return freshInstances, nil
}
