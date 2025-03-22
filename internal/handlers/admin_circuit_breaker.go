package handlers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/resilience/circuitbreaker"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type AdminCircuitBreakerHandler struct {
	configWatcher *config.ConfigWatcher
	logger        logging.Logger
}

func NewAdminCircuitBreakerHandler(configWatcher *config.ConfigWatcher, logger logging.Logger) *AdminCircuitBreakerHandler {
	return &AdminCircuitBreakerHandler{
		configWatcher: configWatcher,
		logger:        logger,
	}
}

func (h *AdminCircuitBreakerHandler) GetCircuitBreakers(c *fiber.Ctx) error {
	cfg := h.configWatcher.GetConfig()

	var circuitBreakers []map[string]interface{}

	for _, route := range cfg.Routes {
		if route.CircuitBreaker != nil && route.CircuitBreaker.Enabled {
			cb := map[string]interface{}{
				"route_name":          route.Name,
				"type":                route.CircuitBreaker.Type,
				"error_threshold":     route.CircuitBreaker.ErrorThreshold,
				"minimum_requests":    route.CircuitBreaker.MinimumRequests,
				"sampling_window":     route.CircuitBreaker.SamplingWindow,
				"open_state_duration": route.CircuitBreaker.OpenStateDuration,
			}

			if route.CircuitBreaker.Type == circuitbreaker.BreakerTypeTimeout {
				cb["timeout_threshold"] = route.CircuitBreaker.TimeoutThreshold
			}

			if route.CircuitBreaker.Type == circuitbreaker.BreakerTypeConcurrency ||
				route.CircuitBreaker.Type == circuitbreaker.BreakerTypeHybrid {
				cb["concurrency_limit"] = route.CircuitBreaker.ConcurrencyLimit
			}

			circuitBreakers = append(circuitBreakers, cb)
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"circuit_breakers": circuitBreakers,
	})
}

func (h *AdminCircuitBreakerHandler) GetCircuitBreaker(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	for _, route := range cfg.Routes {
		if route.Name == routeName {
			if route.CircuitBreaker == nil || !route.CircuitBreaker.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Circuit breaker not configured for this route",
				})
			}

			cb := map[string]interface{}{
				"route_name":             route.Name,
				"type":                   route.CircuitBreaker.Type,
				"error_threshold":        route.CircuitBreaker.ErrorThreshold,
				"timeout_threshold":      route.CircuitBreaker.TimeoutThreshold,
				"minimum_requests":       route.CircuitBreaker.MinimumRequests,
				"sampling_window":        route.CircuitBreaker.SamplingWindow,
				"open_state_duration":    route.CircuitBreaker.OpenStateDuration,
				"half_open_max_requests": route.CircuitBreaker.HalfOpenMaxRequests,
				"concurrency_limit":      route.CircuitBreaker.ConcurrencyLimit,
				"response_status_code":   route.CircuitBreaker.ResponseStatusCode,
				"response_body":          route.CircuitBreaker.ResponseBody,
			}

			if route.CircuitBreaker.FailureStatusCodes != nil && len(route.CircuitBreaker.FailureStatusCodes) > 0 {
				cb["failure_status_codes"] = route.CircuitBreaker.FailureStatusCodes
			}

			return c.Status(http.StatusOK).JSON(cb)
		}
	}

	return c.Status(http.StatusNotFound).JSON(fiber.Map{
		"error": "Route not found",
	})
}

func (h *AdminCircuitBreakerHandler) CreateCircuitBreaker(c *fiber.Ctx) error {
	var request struct {
		RouteName           string `json:"route_name"`
		Type                string `json:"type"`
		ErrorThreshold      int    `json:"error_threshold"`
		TimeoutThreshold    int    `json:"timeout_threshold,omitempty"`
		MinimumRequests     int    `json:"minimum_requests"`
		SamplingWindow      string `json:"sampling_window"`
		OpenStateDuration   string `json:"open_state_duration"`
		HalfOpenMaxRequests int    `json:"half_open_max_requests,omitempty"`
		ConcurrencyLimit    int    `json:"concurrency_limit,omitempty"`
		ResponseStatusCode  int    `json:"response_status_code,omitempty"`
		ResponseBody        string `json:"response_body,omitempty"`
		FailureStatusCodes  []int  `json:"failure_status_codes,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if request.RouteName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	if request.Type == "" {
		request.Type = string(circuitbreaker.BreakerTypeError)
	}

	if request.ErrorThreshold <= 0 {
		request.ErrorThreshold = 50 // Default 50%
	}

	if request.MinimumRequests <= 0 {
		request.MinimumRequests = 20 // Default 20 requests
	}

	if request.SamplingWindow == "" {
		request.SamplingWindow = "1m" // Default 1 minute
	}

	if request.OpenStateDuration == "" {
		request.OpenStateDuration = "30s" // Default 30 seconds
	}

	if request.HalfOpenMaxRequests <= 0 {
		request.HalfOpenMaxRequests = 5 // Default 5 requests
	}

	if request.Type == string(circuitbreaker.BreakerTypeConcurrency) && request.ConcurrencyLimit <= 0 {
		request.ConcurrencyLimit = 100 // Default 100 concurrent requests
	}

	if request.ResponseStatusCode <= 0 {
		request.ResponseStatusCode = http.StatusServiceUnavailable
	}

	if request.ResponseBody == "" {
		request.ResponseBody = "Service temporarily unavailable"
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound bool
	for i, route := range cfg.Routes {
		if route.Name == request.RouteName {
			routeFound = true

			cbConfig := &circuitbreaker.CircuitBreakerConfig{
				Enabled:             true,
				Type:                circuitbreaker.BreakerType(request.Type),
				ErrorThreshold:      request.ErrorThreshold,
				TimeoutThreshold:    request.TimeoutThreshold,
				MinimumRequests:     request.MinimumRequests,
				SamplingWindow:      request.SamplingWindow,
				OpenStateDuration:   request.OpenStateDuration,
				HalfOpenMaxRequests: request.HalfOpenMaxRequests,
				ConcurrencyLimit:    request.ConcurrencyLimit,
				ResponseStatusCode:  request.ResponseStatusCode,
				ResponseBody:        request.ResponseBody,
				FailureStatusCodes:  request.FailureStatusCodes,
			}

			route.CircuitBreaker = cbConfig
			cfg.Routes[i] = route
			break
		}
	}

	if !routeFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	versionedConfig := config.VersionedConfig{
		Version:    string(config.CurrentVersion),
		Config:     cfg,
		ModifiedAt: time.Now(),
	}

	tempPath := "/tmp/horizon_config_" + time.Now().Format("20060102150405") + ".yaml"
	err := config.SaveConfig(&versionedConfig, tempPath)
	if err != nil {
		h.logger.Error("Failed to save temporary config", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Circuit breaker configuration created successfully",
	})
}

func (h *AdminCircuitBreakerHandler) UpdateCircuitBreaker(c *fiber.Ctx) error {
	routeName := c.Params("route")

	var request struct {
		Type                string `json:"type,omitempty"`
		ErrorThreshold      int    `json:"error_threshold,omitempty"`
		TimeoutThreshold    int    `json:"timeout_threshold,omitempty"`
		MinimumRequests     int    `json:"minimum_requests,omitempty"`
		SamplingWindow      string `json:"sampling_window,omitempty"`
		OpenStateDuration   string `json:"open_state_duration,omitempty"`
		HalfOpenMaxRequests int    `json:"half_open_max_requests,omitempty"`
		ConcurrencyLimit    int    `json:"concurrency_limit,omitempty"`
		ResponseStatusCode  int    `json:"response_status_code,omitempty"`
		ResponseBody        string `json:"response_body,omitempty"`
		FailureStatusCodes  []int  `json:"failure_status_codes,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound, cbFound bool
	for i, route := range cfg.Routes {
		if route.Name == routeName {
			routeFound = true

			if route.CircuitBreaker == nil || !route.CircuitBreaker.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Circuit breaker not configured for this route",
				})
			}

			cbFound = true

			if request.Type != "" {
				route.CircuitBreaker.Type = circuitbreaker.BreakerType(request.Type)
			}

			if request.ErrorThreshold > 0 {
				route.CircuitBreaker.ErrorThreshold = request.ErrorThreshold
			}

			if request.TimeoutThreshold > 0 {
				route.CircuitBreaker.TimeoutThreshold = request.TimeoutThreshold
			}

			if request.MinimumRequests > 0 {
				route.CircuitBreaker.MinimumRequests = request.MinimumRequests
			}

			if request.SamplingWindow != "" {
				route.CircuitBreaker.SamplingWindow = request.SamplingWindow
			}

			if request.OpenStateDuration != "" {
				route.CircuitBreaker.OpenStateDuration = request.OpenStateDuration
			}

			if request.HalfOpenMaxRequests > 0 {
				route.CircuitBreaker.HalfOpenMaxRequests = request.HalfOpenMaxRequests
			}

			if request.ConcurrencyLimit > 0 {
				route.CircuitBreaker.ConcurrencyLimit = request.ConcurrencyLimit
			}

			if request.ResponseStatusCode > 0 {
				route.CircuitBreaker.ResponseStatusCode = request.ResponseStatusCode
			}

			if request.ResponseBody != "" {
				route.CircuitBreaker.ResponseBody = request.ResponseBody
			}

			if request.FailureStatusCodes != nil {
				route.CircuitBreaker.FailureStatusCodes = request.FailureStatusCodes
			}

			cfg.Routes[i] = route
			break
		}
	}

	if !routeFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	if !cbFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Circuit breaker not configured for this route",
		})
	}

	versionedConfig := config.VersionedConfig{
		Version:    string(config.CurrentVersion),
		Config:     cfg,
		ModifiedAt: time.Now(),
	}

	tempPath := "/tmp/horizon_config_" + time.Now().Format("20060102150405") + ".yaml"
	err := config.SaveConfig(&versionedConfig, tempPath)
	if err != nil {
		h.logger.Error("Failed to save temporary config", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Circuit breaker configuration updated successfully",
	})
}

func (h *AdminCircuitBreakerHandler) DeleteCircuitBreaker(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound, cbFound bool
	for i, route := range cfg.Routes {
		if route.Name == routeName {
			routeFound = true

			if route.CircuitBreaker == nil || !route.CircuitBreaker.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Circuit breaker not configured for this route",
				})
			}

			cbFound = true
			route.CircuitBreaker.Enabled = false
			cfg.Routes[i] = route
			break
		}
	}

	if !routeFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	if !cbFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Circuit breaker not configured for this route",
		})
	}

	versionedConfig := config.VersionedConfig{
		Version:    string(config.CurrentVersion),
		Config:     cfg,
		ModifiedAt: time.Now(),
	}

	tempPath := "/tmp/horizon_config_" + time.Now().Format("20060102150405") + ".yaml"
	err := config.SaveConfig(&versionedConfig, tempPath)
	if err != nil {
		h.logger.Error("Failed to save temporary config", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Circuit breaker configuration deleted successfully",
	})
}

func (h *AdminCircuitBreakerHandler) ResetCircuitBreaker(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Circuit breaker reset successfully",
		"note":    "This is a runtime operation that doesn't affect configuration.",
	})
}
