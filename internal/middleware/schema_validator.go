package middleware

import (
	"bytes"
	"net/http"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/bimapangestu28/horizon/internal/validator"
	"github.com/gofiber/fiber/v2"
)

func SchemaValidatorMiddleware(validator *validator.SchemaValidator, logger logging.Logger) fiber.Handler {
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
		httpReq, err := createHTTPRequest(c)
		if err != nil {
			logger.Error("Failed to create HTTP request for validation",
				"route", routeName,
				"error", err)
			return c.Next()
		}

		// Validate the request
		result, err := validator.ValidateRequest(httpReq)
		if err != nil {
			logger.Warn("Schema validation skipped",
				"route", routeName,
				"error", err)
			return c.Next()
		}

		if !result.Valid {
			logger.Warn("Schema validation failed",
				"route", routeName,
				"errors", result.Errors)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Schema validation failed",
				"message": result.Message,
				"details": result.Errors,
			})
		}

		// Set validation result in context for downstream handlers
		c.Locals("schema_validation_result", result)

		// Now handle the response validation
		// We need to store the original response handler and replace it
		// with our own to validate the response before sending it
		c.Response().SetStatusCode(fiber.StatusOK)

		// Store the original handlers
		originalBody := c.Response().Body()

		// Process the request
		err = c.Next()

		// Check for errors from the handler
		if err != nil {
			return err // Let Fiber handle the error
		}

		// If the status is not an error, validate the response
		statusCode := c.Response().StatusCode()
		if statusCode >= 200 && statusCode < 300 {
			responseBody := c.Response().Body()

			// Don't validate empty responses
			if len(responseBody) > 0 {
				responseResult, validationErr := validator.ValidateResponse(httpReq, statusCode, responseBody)
				if validationErr != nil {
					logger.Warn("Response schema validation skipped",
						"route", routeName,
						"error", validationErr)
				} else if !responseResult.Valid {
					logger.Error("Response schema validation failed",
						"route", routeName,
						"errors", responseResult.Errors)

					// You could either just log the error or actually return an error to the client
					// For now, we'll just log and let the response proceed
				}
			}
		}

		return nil
	}
}

func createHTTPRequest(c *fiber.Ctx) (*http.Request, error) {
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
