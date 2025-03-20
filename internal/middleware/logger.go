package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// NewLogger creates a new logger middleware that logs HTTP requests
func NewLogger(logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Get request ID if available
		requestID := c.Get(fiber.HeaderXRequestID)
		if requestID == "" {
			requestID = "unknown"
		}

		// Process request
		err := c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Log request details
		logger.Info("HTTP Request",
			"request_id", requestID,
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
			"ip", c.IP(),
			"user_agent", c.Get(fiber.HeaderUserAgent),
		)

		return err
	}
}
