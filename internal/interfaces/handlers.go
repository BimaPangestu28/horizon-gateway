package interfaces

import (
	"github.com/gofiber/fiber/v2"
)

type ProxyHandlerInterface interface {
	HandleRequest(c *fiber.Ctx) error
}

type MetricsHandlerInterface interface {
	GetMetrics(c *fiber.Ctx) error
	RecordRequest(c *fiber.Ctx, routeName string) error
	RecordCacheActivity(routeName, cacheType string, hit bool)
	RecordRateLimit(routeName, limitType, scope string)
	RecordAuthentication(routeName, authType string, success bool, reason string)
	RecordCircuitBreakerTrip(routeName, breakerType string)
	SetCircuitBreakerState(routeName, breakerType string, state int)
	SetUpstreamHealth(upstream, routeName string, healthy bool)
	SetUpstreamConnections(upstream, routeName string, connections int)
	RecordConfigReload()
	RecordUpstreamRequest(upstream, method, routeName string, duration float64)
}

type ErrorHandlerInterface interface {
	HandleError(c *fiber.Ctx, err error) error
}
