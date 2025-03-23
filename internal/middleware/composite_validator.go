package middleware

import (
	"bytes"
	"context"
	"net/http"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/internal/validator"
	"github.com/gofiber/fiber/v2"
)

type ValidatorMiddlewareConfig struct {
	Validators    map[string]validator.Validator
	StopOnFirst   bool
	RouteMappings map[string][]string
}

func CompositeValidatorMiddleware(config *ValidatorMiddlewareConfig, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		routeData, ok := c.Locals("route").(map[string]string)
		if !ok {
			return c.Next()
		}

		routeName, ok := routeData["name"]
		if !ok {
			return c.Next()
		}

		// Get validators for this route
		validatorNames, ok := config.RouteMappings[routeName]
		if !ok {
			// No validators configured for this route
			return c.Next()
		}

		// Collect validators
		validators := make([]validator.Validator, 0, len(validatorNames))
		for _, name := range validatorNames {
			val, exists := config.Validators[name]
			if !exists {
				logger.Warn("Validator not found", "name", name, "route", routeName)
				continue
			}
			validators = append(validators, val)
		}

		if len(validators) == 0 {
			// No validators found
			return c.Next()
		}

		// Create composite validator
		compositeValidator := validator.NewCompositeValidator(validators, logger, config.StopOnFirst)

		// Prepare validation context with route info
		ctx := context.WithValue(c.Context(), "route_name", routeName)
		ctx = context.WithValue(ctx, "validator_names", validatorNames)

		// Create HTTP request for validation
		httpReq, err := createCompositeHTTPRequest(c)
		if err != nil {
			logger.Error("Failed to create HTTP request for validation",
				"route", routeName,
				"error", err)
			return c.Next()
		}

		// Validate request
		result, err := compositeValidator.Validate(ctx, httpReq)
		if err != nil {
			logger.Error("Validation error",
				"route", routeName,
				"error", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Validation error",
				"details": err.Error(),
			})
		}

		if !result.Valid {
			logger.Warn("Validation failed",
				"route", routeName,
				"errors", result.Errors)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation failed",
				"message": result.Message,
				"details": result.Errors,
			})
		}

		// Store validation result in context
		c.Locals("validation_result", result)

		// Continue with next handler
		return c.Next()
	}
}

func ValidateJWTMiddleware(jwtValidator *validator.JWTValidator, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		result := jwtValidator.ValidateToken(tokenString)
		if !result.Valid {
			logger.Warn("JWT validation failed",
				"error", result.Error.Error())

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Invalid JWT token",
				"details": result.Error.Error(),
			})
		}

		// Store token claims in context
		c.Locals("jwt_claims", result.Claims)
		c.Locals("subject_id", result.SubjectID)
		c.Locals("roles", result.Roles)

		return c.Next()
	}
}

func ValidateURLMiddleware(urlValidator *validator.URLValidator, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get URL parameter from query string or body
		url := c.Query("url")
		if url == "" {
			// Try to get from body
			body := make(map[string]interface{})
			if err := c.BodyParser(&body); err == nil {
				if urlVal, ok := body["url"].(string); ok {
					url = urlVal
				}
			}
		}

		if url == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "URL parameter is required",
			})
		}

		result := urlValidator.ValidateURL(url)
		if !result.Valid {
			logger.Warn("URL validation failed",
				"url", url,
				"errors", result.Errors)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid URL",
				"message": result.Message,
				"details": result.Errors,
			})
		}

		// Store validated URL in context
		c.Locals("validated_url", result.URL)

		return c.Next()
	}
}

func createCompositeHTTPRequest(c *fiber.Ctx) (*http.Request, error) {
	method := c.Method()
	url := c.BaseURL() + c.OriginalURL()

	var body []byte
	if len(c.Body()) > 0 {
		body = c.Body()
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	return req, nil
}
