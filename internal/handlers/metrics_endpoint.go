package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/interfaces"
)

// MetricsEndpointHandler handles the metrics endpoints for the admin API
type MetricsEndpointHandler struct {
	metricsHandler interfaces.MetricsHandlerInterface
}

// NewMetricsEndpointHandler creates a new metrics endpoint handler
func NewMetricsEndpointHandler(metricsHandler interfaces.MetricsHandlerInterface) *MetricsEndpointHandler {
	return &MetricsEndpointHandler{
		metricsHandler: metricsHandler,
	}
}

// GetMetrics handles the /metrics endpoint
func (h *MetricsEndpointHandler) GetMetrics(c *fiber.Ctx) error {
	return h.metricsHandler.GetMetrics(c)
}

// GetRouteMetrics returns metrics for a specific route
func (h *MetricsEndpointHandler) GetRouteMetrics(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	// For a real implementation, this would query metrics from Prometheus
	// or another metrics store. For now, we'll return sample data.

	// Generate sample metrics
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	metrics := []map[string]interface{}{
		{
			"timestamp":     oneHourAgo.Add(10 * time.Minute).Unix(),
			"requests":      123,
			"errors":        5,
			"response_time": 45.3,
		},
		{
			"timestamp":     oneHourAgo.Add(20 * time.Minute).Unix(),
			"requests":      142,
			"errors":        2,
			"response_time": 38.7,
		},
		{
			"timestamp":     oneHourAgo.Add(30 * time.Minute).Unix(),
			"requests":      156,
			"errors":        3,
			"response_time": 42.1,
		},
		{
			"timestamp":     oneHourAgo.Add(40 * time.Minute).Unix(),
			"requests":      168,
			"errors":        4,
			"response_time": 44.6,
		},
		{
			"timestamp":     oneHourAgo.Add(50 * time.Minute).Unix(),
			"requests":      173,
			"errors":        1,
			"response_time": 39.3,
		},
		{
			"timestamp":     now.Unix(),
			"requests":      185,
			"errors":        2,
			"response_time": 41.8,
		},
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"route":   routeName,
		"metrics": metrics,
		"summary": fiber.Map{
			"total_requests":    947,
			"total_errors":      17,
			"error_rate":        1.79,
			"avg_response_time": 41.97,
		},
	})
}
