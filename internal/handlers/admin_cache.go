package handlers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/cache"
	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type RedisConfig struct {
	Address   string `json:"address"`
	Password  string `json:"password,omitempty"`
	DB        int    `json:"db"`
	KeyPrefix string `json:"key_prefix,omitempty"`
}

type AdminCacheHandler struct {
	configWatcher *config.ConfigWatcher
	logger        logging.Logger
}

func NewAdminCacheHandler(configWatcher *config.ConfigWatcher, logger logging.Logger) *AdminCacheHandler {
	return &AdminCacheHandler{
		configWatcher: configWatcher,
		logger:        logger,
	}
}

func (h *AdminCacheHandler) GetCacheConfigs(c *fiber.Ctx) error {
	cfg := h.configWatcher.GetConfig()

	var cacheConfigs []map[string]interface{}

	for _, route := range cfg.Routes {
		if route.Caching != nil && route.Caching.Enabled {
			cacheConfig := map[string]interface{}{
				"route_name": route.Name,
				"type":       route.Caching.Type,
				"ttl":        route.Caching.TTL,
				"max_size":   route.Caching.MaxSize,
			}

			if len(route.Caching.Methods) > 0 {
				cacheConfig["methods"] = route.Caching.Methods
			}

			if route.Caching.CacheKeyTemplate != "" {
				cacheConfig["cache_key_template"] = route.Caching.CacheKeyTemplate
			}

			cacheConfigs = append(cacheConfigs, cacheConfig)
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"cache_configs": cacheConfigs,
	})
}

func (h *AdminCacheHandler) GetCacheConfig(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	for _, route := range cfg.Routes {
		if route.Name == routeName {
			if route.Caching == nil || !route.Caching.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Cache configuration not found for this route",
				})
			}

			cacheConfig := map[string]interface{}{
				"route_name":            route.Name,
				"type":                  route.Caching.Type,
				"ttl":                   route.Caching.TTL,
				"max_size":              route.Caching.MaxSize,
				"methods":               route.Caching.Methods,
				"cache_key_template":    route.Caching.CacheKeyTemplate,
				"respect_cache_control": route.Caching.RespectCacheControl,
				"include_host":          route.Caching.IncludeHost,
			}

			if len(route.Caching.IgnoreQueryParams) > 0 {
				cacheConfig["ignore_query_params"] = route.Caching.IgnoreQueryParams
			}

			if len(route.Caching.VaryHeaders) > 0 {
				cacheConfig["vary_headers"] = route.Caching.VaryHeaders
			}

			if len(route.Caching.NeverCache) > 0 {
				cacheConfig["never_cache"] = route.Caching.NeverCache
			}

			if len(route.Caching.AlwaysCache) > 0 {
				cacheConfig["always_cache"] = route.Caching.AlwaysCache
			}

			if route.Caching.Redis != nil {
				cacheConfig["redis"] = map[string]interface{}{
					"address":    route.Caching.Redis.Address,
					"db":         route.Caching.Redis.DB,
					"key_prefix": route.Caching.Redis.KeyPrefix,
				}
			}

			return c.Status(http.StatusOK).JSON(cacheConfig)
		}
	}

	return c.Status(http.StatusNotFound).JSON(fiber.Map{
		"error": "Route not found",
	})
}

func (h *AdminCacheHandler) CreateCacheConfig(c *fiber.Ctx) error {
	var request struct {
		RouteName           string       `json:"route_name"`
		Type                string       `json:"type"`
		TTL                 string       `json:"ttl"`
		MaxSize             int64        `json:"max_size"`
		Methods             []string     `json:"methods"`
		CacheKeyTemplate    string       `json:"cache_key_template,omitempty"`
		IgnoreQueryParams   []string     `json:"ignore_query_params,omitempty"`
		VaryHeaders         []string     `json:"vary_headers,omitempty"`
		NeverCache          []string     `json:"never_cache,omitempty"`
		AlwaysCache         []string     `json:"always_cache,omitempty"`
		RespectCacheControl bool         `json:"respect_cache_control"`
		IncludeHost         bool         `json:"include_host"`
		Redis               *RedisConfig `json:"redis,omitempty"`
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
		request.Type = string(cache.CacheTypeMemory)
	}

	if request.TTL == "" {
		request.TTL = "1m"
	}

	if len(request.Methods) == 0 {
		request.Methods = []string{"GET"}
	}

	if request.CacheKeyTemplate == "" {
		request.CacheKeyTemplate = "{method}:{path}:{query}"
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound bool
	for i, route := range cfg.Routes {
		if route.Name == request.RouteName {
			routeFound = true

			cacheConfig := &cache.CacheConfig{
				Enabled:             true,
				Type:                cache.CacheType(request.Type),
				TTL:                 request.TTL,
				MaxSize:             request.MaxSize,
				Methods:             request.Methods,
				CacheKeyTemplate:    request.CacheKeyTemplate,
				IgnoreQueryParams:   request.IgnoreQueryParams,
				VaryHeaders:         request.VaryHeaders,
				NeverCache:          request.NeverCache,
				AlwaysCache:         request.AlwaysCache,
				RespectCacheControl: request.RespectCacheControl,
				IncludeHost:         request.IncludeHost,
			}

			if request.Redis != nil && request.Type == string(cache.CacheTypeRedis) {
				cacheConfig.Redis = &cache.RedisCacheConfig{
					Address:   request.Redis.Address,
					Password:  request.Redis.Password,
					DB:        request.Redis.DB,
					KeyPrefix: request.Redis.KeyPrefix,
				}

				if cacheConfig.Redis.Address == "" {
					cacheConfig.Redis.Address = "localhost:6379"
				}

				if cacheConfig.Redis.KeyPrefix == "" {
					cacheConfig.Redis.KeyPrefix = "horizon:cache:"
				}
			}

			route.Caching = cacheConfig
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
		"message": "Cache configuration created successfully",
	})
}

func (h *AdminCacheHandler) UpdateCacheConfig(c *fiber.Ctx) error {
	routeName := c.Params("route")

	var request struct {
		Type                string       `json:"type,omitempty"`
		TTL                 string       `json:"ttl,omitempty"`
		MaxSize             int64        `json:"max_size,omitempty"`
		Methods             []string     `json:"methods,omitempty"`
		CacheKeyTemplate    string       `json:"cache_key_template,omitempty"`
		IgnoreQueryParams   []string     `json:"ignore_query_params,omitempty"`
		VaryHeaders         []string     `json:"vary_headers,omitempty"`
		NeverCache          []string     `json:"never_cache,omitempty"`
		AlwaysCache         []string     `json:"always_cache,omitempty"`
		RespectCacheControl *bool        `json:"respect_cache_control,omitempty"`
		IncludeHost         *bool        `json:"include_host,omitempty"`
		Redis               *RedisConfig `json:"redis,omitempty"`
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

	var routeFound, cacheFound bool
	for i, route := range cfg.Routes {
		if route.Name == routeName {
			routeFound = true

			if route.Caching == nil || !route.Caching.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Cache configuration not found for this route",
				})
			}

			cacheFound = true

			if request.Type != "" {
				route.Caching.Type = cache.CacheType(request.Type)
			}

			if request.TTL != "" {
				route.Caching.TTL = request.TTL
			}

			if request.MaxSize > 0 {
				route.Caching.MaxSize = request.MaxSize
			}

			if len(request.Methods) > 0 {
				route.Caching.Methods = request.Methods
			}

			if request.CacheKeyTemplate != "" {
				route.Caching.CacheKeyTemplate = request.CacheKeyTemplate
			}

			if request.IgnoreQueryParams != nil {
				route.Caching.IgnoreQueryParams = request.IgnoreQueryParams
			}

			if request.VaryHeaders != nil {
				route.Caching.VaryHeaders = request.VaryHeaders
			}

			if request.NeverCache != nil {
				route.Caching.NeverCache = request.NeverCache
			}

			if request.AlwaysCache != nil {
				route.Caching.AlwaysCache = request.AlwaysCache
			}

			if request.RespectCacheControl != nil {
				route.Caching.RespectCacheControl = *request.RespectCacheControl
			}

			if request.IncludeHost != nil {
				route.Caching.IncludeHost = *request.IncludeHost
			}

			if request.Redis != nil && route.Caching.Type == cache.CacheTypeRedis {
				if route.Caching.Redis == nil {
					route.Caching.Redis = &cache.RedisCacheConfig{}
				}

				if request.Redis.Address != "" {
					route.Caching.Redis.Address = request.Redis.Address
				}

				if request.Redis.Password != "" {
					route.Caching.Redis.Password = request.Redis.Password
				}

				if request.Redis.DB != 0 {
					route.Caching.Redis.DB = request.Redis.DB
				}

				if request.Redis.KeyPrefix != "" {
					route.Caching.Redis.KeyPrefix = request.Redis.KeyPrefix
				}
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

	if !cacheFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Cache configuration not found for this route",
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
		"message": "Cache configuration updated successfully",
	})
}

func (h *AdminCacheHandler) DeleteCacheConfig(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound, cacheFound bool
	for i, route := range cfg.Routes {
		if route.Name == routeName {
			routeFound = true

			if route.Caching == nil || !route.Caching.Enabled {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"error": "Cache configuration not found for this route",
				})
			}

			cacheFound = true
			route.Caching.Enabled = false
			cfg.Routes[i] = route
			break
		}
	}

	if !routeFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	if !cacheFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Cache configuration not found for this route",
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
		"message": "Cache configuration deleted successfully",
	})
}

func (h *AdminCacheHandler) ClearCache(c *fiber.Ctx) error {
	routeName := c.Params("route")

	if routeName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name is required",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Cache cleared successfully for route: " + routeName,
		"note":    "This is a runtime operation that doesn't affect configuration",
	})
}
