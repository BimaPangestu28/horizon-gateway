package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/security/auth"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type AdminAuthHandler struct {
	configWatcher *config.ConfigWatcher
	logger        logging.Logger
}

func NewAdminAuthHandler(configWatcher *config.ConfigWatcher, logger logging.Logger) *AdminAuthHandler {
	return &AdminAuthHandler{
		configWatcher: configWatcher,
		logger:        logger,
	}
}

func (h *AdminAuthHandler) GetAPIKeys(c *fiber.Ctx) error {
	cfg := h.configWatcher.GetConfig()

	var keys []map[string]interface{}

	for _, route := range cfg.Routes {
		if route.Auth != nil && route.Auth.Type == "api_key" && route.Auth.APIKey != nil {
			for _, key := range route.Auth.APIKey.Keys {
				keyInfo := map[string]interface{}{
					"key_id": key.Name,
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

func (h *AdminAuthHandler) CreateAPIKey(c *fiber.Ctx) error {
	var request struct {
		RouteName string            `json:"route_name"`
		Name      string            `json:"name"`
		Scopes    []string          `json:"scopes"`
		Expires   string            `json:"expires,omitempty"`
		Metadata  map[string]string `json:"metadata,omitempty"`
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

	if request.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "API key name is required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound bool
	var generatedKey string // Declare the variable to store the generated key

	for i, route := range cfg.Routes {
		if route.Name == request.RouteName {
			routeFound = true

			if route.Auth == nil {
				route.Auth = &auth.AuthConfig{
					Enabled: true,
					Type:    auth.AuthTypeAPIKey,
					APIKey: &auth.APIKeyConfig{
						Header: "X-API-Key",
						Keys:   []auth.APIKey{},
					},
				}
			} else if route.Auth.Type != auth.AuthTypeAPIKey {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Route uses a different authentication type",
				})
			}

			generatedKey = generateAPIKey() // Store the generated key

			apiKey := auth.APIKey{
				Key:      generatedKey,
				Name:     request.Name,
				Scopes:   request.Scopes,
				Expires:  request.Expires,
				Metadata: request.Metadata,
			}

			route.Auth.APIKey.Keys = append(route.Auth.APIKey.Keys, apiKey)
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
		"message": "API key created successfully",
		"key":     generatedKey,
	})
}

func (h *AdminAuthHandler) DeleteAPIKey(c *fiber.Ctx) error {
	routeName := c.Params("route")
	keyName := c.Params("key")

	if routeName == "" || keyName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Route name and key name are required",
		})
	}

	cfg := h.configWatcher.GetConfig()

	var routeFound, keyFound bool
	for i, route := range cfg.Routes {
		if route.Name == routeName {
			routeFound = true

			if route.Auth == nil || route.Auth.Type != auth.AuthTypeAPIKey || route.Auth.APIKey == nil {
				break
			}

			for j, key := range route.Auth.APIKey.Keys {
				if key.Name == keyName {
					keyFound = true

					route.Auth.APIKey.Keys = append(route.Auth.APIKey.Keys[:j], route.Auth.APIKey.Keys[j+1:]...)
					cfg.Routes[i] = route
					break
				}
			}

			break
		}
	}

	if !routeFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	if !keyFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
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
		"message": "API key deleted successfully",
	})
}

func (h *AdminAuthHandler) GetJWTConfigs(c *fiber.Ctx) error {
	cfg := h.configWatcher.GetConfig()

	var jwtConfigs []map[string]interface{}

	for _, route := range cfg.Routes {
		if route.Auth != nil && route.Auth.Type == "jwt" && route.Auth.JWT != nil {
			jwtConfig := map[string]interface{}{
				"route_name": route.Name,
				"algorithm":  route.Auth.JWT.Algorithm,
				"issuer":     route.Auth.JWT.Issuer,
				"audience":   route.Auth.JWT.Audience,
			}

			if route.Auth.JWT.ClaimsToHeaders != nil {
				jwtConfig["claims_to_headers"] = route.Auth.JWT.ClaimsToHeaders
			}

			jwtConfigs = append(jwtConfigs, jwtConfig)
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"jwt_configs": jwtConfigs,
	})
}

func (h *AdminAuthHandler) UpdateJWTConfig(c *fiber.Ctx) error {
	routeName := c.Params("route")

	var request struct {
		Algorithm       string            `json:"algorithm"`
		Secret          string            `json:"secret,omitempty"`
		PublicKey       string            `json:"public_key,omitempty"`
		PublicKeyFile   string            `json:"public_key_file,omitempty"`
		Issuer          string            `json:"issuer,omitempty"`
		Audience        string            `json:"audience,omitempty"`
		ClaimsToHeaders map[string]string `json:"claims_to_headers,omitempty"`
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

	var routeFound bool
	for i, route := range cfg.Routes {
		if route.Name == routeName {
			routeFound = true

			if route.Auth == nil {
				route.Auth = &auth.AuthConfig{
					Enabled: true,
					Type:    auth.AuthTypeJWT,
					JWT:     &auth.JWTConfig{},
				}
			} else if route.Auth.Type != auth.AuthTypeJWT {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Route uses a different authentication type",
				})
			}

			if route.Auth.JWT == nil {
				route.Auth.JWT = &auth.JWTConfig{}
			}

			jwt := route.Auth.JWT
			jwt.Algorithm = request.Algorithm

			if request.Secret != "" {
				jwt.Secret = request.Secret
			}

			if request.PublicKey != "" {
				jwt.PublicKey = request.PublicKey
			}

			if request.PublicKeyFile != "" {
				jwt.PublicKeyFile = request.PublicKeyFile
			}

			if request.Issuer != "" {
				jwt.Issuer = request.Issuer
			}

			if request.Audience != "" {
				jwt.Audience = request.Audience
			}

			if request.ClaimsToHeaders != nil {
				jwt.ClaimsToHeaders = request.ClaimsToHeaders
			}

			route.Auth.JWT = jwt
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

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "JWT configuration updated successfully",
	})
}

func generateAPIKey() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("horizon.%d.%s", timestamp, "api-key")
}
