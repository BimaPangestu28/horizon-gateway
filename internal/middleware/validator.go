package middleware

import (
	"net/http"

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

		httpReq := &http.Request{
			Method: c.Method(),
			URL:    createURL(c.Path(), c.Query()),
			Header: make(http.Header),
			Body:   c.Request().BodyStream(),
			Host:   c.Hostname(),
		}

		c.Request().Header.VisitAll(func(key, value []byte) {
			httpReq.Header.Add(string(key), string(value))
		})

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

		// Reset the body for downstream handlers
		c.Request().ResetBody()
		c.Request().SetBody(c.Body())

		return c.Next()
	}
}

func createURL(path string, queryString string) *http.URL {
	u := &http.URL{
		Path: path,
	}
	if queryString != "" {
		u.RawQuery = queryString
	}
	return u
}
