package handlers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/security/ratelimit"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type AdminRateLimitHandler struct {
	configWatcher *config.ConfigWatcher
	logger        logging.Logger
}

func NewAdminRateLimitHandler(configWatcher *config.ConfigWatcher, logger logging.Logger) *AdminRateLimitHandler {
	return &AdminRateLimitHandler{
		configWatcher: configWatcher,
		logger:        logger,
	}
}

func (h *AdminRateLimitHandler) GetRateLimits(c *fiber.Ctx) error {
	cfg := h.configWatcher.GetConfig()

	var rateLimits []map[string]interface{}

	for _, route := range cfg.Routes {
		if route.RateLimiting != nil && route.RateLimiting.Enabled {
			rateLimit := map[string]interface{}{
				"route_name": route.Name,
				"type":       route.RateLimiting.Type,
				"scope":      route.RateLimiting.Scope,
				"limit":      route.RateLimiting.Limit,
				"window":     route.RateLimiting.Window,
			}

			if route.RateLimiting.BurstSize > 0 {
				rateLimit["burst_size"] = route.RateLimiting.BurstSize
			}

			if route.RateLimiting.RefillRate > 0 {
				rateLimit["refill_rate"] = route.RateLimiting.RefillRate
			}

			rateLimits = append(rateLimits, rateLimit)
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"rate_limits": rateLimits,
	})
}

func (h *AdminRateLimitHandler) GetRateLimit(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	for _, route := range cfg.Routes {
		if route.Name == routeName {
			if route.RateLimiting == nil || !route.RateLimiting.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Rate limiting not configured for this route",
				})
			}

			rateLimit := map[string]interface{}{
				"route_name":           route.Name,
				"type":                 route.RateLimiting.Type,
				"scope":                route.RateLimiting.Scope,
				"limit":                route.RateLimiting.Limit,
				"window":               route.RateLimiting.Window,
				"response_status_code": route.RateLimiting.ResponseStatusCode,
				"response_body":        route.RateLimiting.ResponseBody,
				"enable_headers":       route.RateLimiting.EnableHeaders,
			}

			if route.RateLimiting.BurstSize > 0 {
				rateLimit["burst_size"] = route.RateLimiting.BurstSize
			}

			if route.RateLimiting.RefillRate > 0 {
				rateLimit["refill_rate"] = route.RateLimiting.RefillRate
			}

			if route.RateLimiting.Header != "" {
				rateLimit["header"] = route.RateLimiting.Header
			}

			return c.Status(http.StatusOK).JSON(rateLimit)
		}
	}

	return c.Status(http.StatusNotFound).JSON(fiber.Map{
		"error": "Route not found",
	})
}

func (h *AdminRateLimitHandler) CreateRateLimit(c *fiber.Ctx) error {
	var request struct {
		RouteName          string  `json:"route_name"`
		Type               string  `json:"type"`
		Scope              string  `json:"scope"`
		Limit              int     `json:"limit"`
		Window             string  `json:"window"`
		Header             string  `json:"header,omitempty"`
		BurstSize          int     `json:"burst_size,omitempty"`
		RefillRate         float64 `json:"refill_rate,omitempty"`
		ResponseStatusCode int     `json:"response_status_code,omitempty"`
		ResponseBody       string  `json:"response_body,omitempty"`
		EnableHeaders      bool    `json:"enable_headers,omitempty"`
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
		request.Type = string(ratelimit.RateLimitTypeFixedWindow)
	}

	if request.Scope == "" {
		request.Scope = string(ratelimit.RateLimitScopeIP)
	}

	if request.Limit <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Limit must be greater than 0",
		})
	}

	if request.Window == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Window is required",
		})
	}

	if request.Scope == string(ratelimit.RateLimitScopeHeader) && request.Header == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Header is required when scope is 'header'",
		})
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound bool
	for i, route := range cfg.Routes {
		if route.Name == request.RouteName {
			routeFound = true

			rateLimitConfig := &ratelimit.RateLimitConfig{
				Enabled:            true,
				Type:               ratelimit.RateLimitType(request.Type),
				Scope:              ratelimit.RateLimitScope(request.Scope),
				Limit:              request.Limit,
				Window:             request.Window,
				Header:             request.Header,
				BurstSize:          request.BurstSize,
				RefillRate:         request.RefillRate,
				ResponseStatusCode: request.ResponseStatusCode,
				ResponseBody:       request.ResponseBody,
				EnableHeaders:      request.EnableHeaders,
			}

			if rateLimitConfig.ResponseStatusCode == 0 {
				rateLimitConfig.ResponseStatusCode = http.StatusTooManyRequests
			}

			if rateLimitConfig.ResponseBody == "" {
				rateLimitConfig.ResponseBody = "Rate limit exceeded"
			}

			route.RateLimiting = rateLimitConfig
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
		"message": "Rate limit configuration created successfully",
	})
}

func (h *AdminRateLimitHandler) UpdateRateLimit(c *fiber.Ctx) error {
	routeName := c.Params("route")

	var request struct {
		Type               string  `json:"type"`
		Scope              string  `json:"scope"`
		Limit              int     `json:"limit"`
		Window             string  `json:"window"`
		Header             string  `json:"header,omitempty"`
		BurstSize          int     `json:"burst_size,omitempty"`
		RefillRate         float64 `json:"refill_rate,omitempty"`
		ResponseStatusCode int     `json:"response_status_code,omitempty"`
		ResponseBody       string  `json:"response_body,omitempty"`
		EnableHeaders      bool    `json:"enable_headers,omitempty"`
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

	var routeFound, rateLimitFound bool
	for i, route := range cfg.Routes {
		if route.Name == routeName {
			routeFound = true

			if route.RateLimiting == nil || !route.RateLimiting.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Rate limiting not configured for this route",
				})
			}

			rateLimitFound = true

			if request.Type != "" {
				route.RateLimiting.Type = ratelimit.RateLimitType(request.Type)
			}

			if request.Scope != "" {
				route.RateLimiting.Scope = ratelimit.RateLimitScope(request.Scope)
			}

			if request.Limit > 0 {
				route.RateLimiting.Limit = request.Limit
			}

			if request.Window != "" {
				route.RateLimiting.Window = request.Window
			}

			if request.Scope == string(ratelimit.RateLimitScopeHeader) && request.Header != "" {
				route.RateLimiting.Header = request.Header
			}

			if request.BurstSize > 0 {
				route.RateLimiting.BurstSize = request.BurstSize
			}

			if request.RefillRate > 0 {
				route.RateLimiting.RefillRate = request.RefillRate
			}

			if request.ResponseStatusCode > 0 {
				route.RateLimiting.ResponseStatusCode = request.ResponseStatusCode
			}

			if request.ResponseBody != "" {
				route.RateLimiting.ResponseBody = request.ResponseBody
			}

			route.RateLimiting.EnableHeaders = request.EnableHeaders

			cfg.Routes[i] = route
			break
		}
	}

	if !routeFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	if !rateLimitFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Rate limiting not configured for this route",
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
		"message": "Rate limit configuration updated successfully",
	})
}

func (h *AdminRateLimitHandler) DeleteRateLimit(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound, rateLimitFound bool
	for i, route := range cfg.Routes {
		if route.Name == routeName {
			routeFound = true

			if route.RateLimiting == nil || !route.RateLimiting.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Rate limiting not configured for this route",
				})
			}

			rateLimitFound = true
			route.RateLimiting.Enabled = false
			cfg.Routes[i] = route
			break
		}
	}

	if !routeFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	if !rateLimitFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Rate limiting not configured for this route",
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
		"message": "Rate limit configuration deleted successfully",
	})
}
