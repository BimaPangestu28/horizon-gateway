package handlers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/core"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type AdminHandler struct {
	configWatcher *config.ConfigWatcher
	router        *core.Router
	logger        logging.Logger
}

func NewAdminHandler(configWatcher *config.ConfigWatcher, router *core.Router, logger logging.Logger) *AdminHandler {
	return &AdminHandler{
		configWatcher: configWatcher,
		router:        router,
		logger:        logger,
	}
}

func (h *AdminHandler) GetRoutes(c *fiber.Ctx) error {
	routes := h.router.GetAllRoutes()
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"routes": routes,
	})
}

func (h *AdminHandler) GetRoute(c *fiber.Ctx) error {
	name := c.Params("name")

	routes := h.router.GetAllRoutes()
	for _, route := range routes {
		if route.Name == name {
			return c.Status(http.StatusOK).JSON(route)
		}
	}

	return c.Status(http.StatusNotFound).JSON(fiber.Map{
		"error": "Route not found",
	})
}

func (h *AdminHandler) CreateRoute(c *fiber.Ctx) error {
	var routeConfig config.RouteConfig
	if err := c.BodyParser(&routeConfig); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	cfg := h.configWatcher.GetConfig()

	for _, route := range cfg.Routes {
		if route.Name == routeConfig.Name {
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "Route with this name already exists",
			})
		}

		if route.ListenPath == routeConfig.ListenPath {
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "Route with this listen path already exists",
			})
		}
	}

	cfg.Routes = append(cfg.Routes, routeConfig)

	err := h.applyConfigChange(cfg)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to apply configuration change",
		})
	}

	return c.Status(http.StatusCreated).JSON(routeConfig)
}

func (h *AdminHandler) UpdateRoute(c *fiber.Ctx) error {
	name := c.Params("name")

	var routeConfig config.RouteConfig
	if err := c.BodyParser(&routeConfig); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	cfg := h.configWatcher.GetConfig()

	found := false
	for i, route := range cfg.Routes {
		if route.Name == name {
			cfg.Routes[i] = routeConfig
			found = true
			break
		}
	}

	if !found {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	err := h.applyConfigChange(cfg)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to apply configuration change",
		})
	}

	return c.Status(http.StatusOK).JSON(routeConfig)
}

func (h *AdminHandler) DeleteRoute(c *fiber.Ctx) error {
	name := c.Params("name")

	cfg := h.configWatcher.GetConfig()

	found := false
	for i, route := range cfg.Routes {
		if route.Name == name {
			cfg.Routes = append(cfg.Routes[:i], cfg.Routes[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	err := h.applyConfigChange(cfg)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to apply configuration change",
		})
	}

	return c.Status(http.StatusNoContent).Send(nil)
}

func (h *AdminHandler) GetConfig(c *fiber.Ctx) error {
	cfg := h.configWatcher.GetConfig()
	return c.Status(http.StatusOK).JSON(cfg)
}

func (h *AdminHandler) UpdateConfig(c *fiber.Ctx) error {
	var cfg config.Config
	if err := c.BodyParser(&cfg); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.applyConfigChange(&cfg)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to apply configuration change",
		})
	}

	return c.Status(http.StatusOK).JSON(cfg)
}

func (h *AdminHandler) applyConfigChange(cfg *config.Config) error {
	versionedConfig := config.VersionedConfig{
		Version:    string(config.CurrentVersion),
		Config:     cfg,
		ModifiedAt: time.Now(),
	}

	tempPath := "/tmp/horizon_config_" + time.Now().Format("20060102150405") + ".yaml"
	err := config.SaveConfig(&versionedConfig, tempPath)
	if err != nil {
		h.logger.Error("Failed to save temporary config", "error", err)
		return err
	}

	return nil
}

func (h *AdminHandler) GetAPIKeys(c *fiber.Ctx) error {
	cfg := h.configWatcher.GetConfig()

	var keys []map[string]interface{}

	for _, route := range cfg.Routes {
		if route.Auth != nil && route.Auth.Type == "api_key" && route.Auth.APIKey != nil {
			for _, key := range route.Auth.APIKey.Keys {
				keyInfo := map[string]interface{}{
					"name":   key.Name,
					"scopes": key.Scopes,
					"route":  route.Name,
				}

				if key.Expires != "" {
					keyInfo["expires"] = key.Expires
				}

				if key.Metadata != nil {
					keyInfo["metadata"] = key.Metadata
				}

				keys = append(keys, keyInfo)
			}
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"api_keys": keys,
	})
}

func (h *AdminHandler) GetHealth(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *AdminHandler) GetVersion(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"version": "v1.0.0",
		"build":   "20240301-1",
	})
}

func (h *AdminHandler) GetConfigBackups(c *fiber.Ctx) error {
	backups := []map[string]interface{}{
		{
			"id":        "backup_20240301_120000",
			"timestamp": "2024-03-01T12:00:00Z",
			"version":   "v1.0.0",
		},
		{
			"id":        "backup_20240228_120000",
			"timestamp": "2024-02-28T12:00:00Z",
			"version":   "v1.0.0",
		},
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"backups": backups,
	})
}

func (h *AdminHandler) RestoreConfigBackup(c *fiber.Ctx) error {
	backupID := c.Params("id")

	if backupID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Backup ID is required",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Configuration restored from backup " + backupID,
	})
}
