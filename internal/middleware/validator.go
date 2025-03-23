package middleware

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/internal/validator"
	"github.com/gofiber/fiber/v2"
)

func RequestValidatorMiddleware(validator *validator.RequestValidator, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		routeData, ok := c.Locals("route").(map[string]string)
		if !ok {
			return c.Next()
		}

		routeName, ok := routeData["name"]
		if !ok {
			return c.Next()
		}

		// Create HTTP request for validation
		httpReq, err := createRequestValidatorHTTPRequest(c)
		if err != nil {
			logger.Error("Failed to create HTTP request for validation",
				"route", routeName,
				"error", err)
			return c.Next()
		}

		result, err := validator.ValidateRequest(httpReq, routeName)
		if err != nil {
			logger.Error("Request validation error",
				"route", routeName,
				"error", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Request validation error",
				"details": err.Error(),
			})
		}

		if !result.Valid {
			logger.Warn("Request validation failed",
				"route", routeName,
				"errors", result.Errors)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation failed",
				"message": result.Message,
				"details": result.Errors,
			})
		}

		// Set validation result in context for downstream handlers
		c.Locals("validation_result", result)

		return c.Next()
	}
}

func createRequestValidatorHTTPRequest(c *fiber.Ctx) (*http.Request, error) {
	method := c.Method()

	// Create URL object
	parsedURL, err := url.Parse(c.Path())
	if err != nil {
		return nil, err
	}

	// Add query string
	parsedURL.RawQuery = string(c.Request().URI().QueryString())

	// Prepare request body
	var bodyReader *bytes.Reader
	if len(c.Body()) > 0 {
		bodyReader = bytes.NewReader(c.Body())
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	// Create a new HTTP request
	httpReq, err := http.NewRequest(method, parsedURL.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	httpReq.Host = c.Hostname()

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Add(string(key), string(value))
	})

	return httpReq, nil
}
