package handlers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// DefaultErrorHandler provides a default handler for HTTP errors
func DefaultErrorHandler(c *fiber.Ctx, err error) error {
	// Default to internal server error
	code := http.StatusInternalServerError

	// Check if it's a fiber error
	if fiberErr, ok := err.(*fiber.Error); ok {
		code = fiberErr.Code
	}

	// Return error as JSON
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

// HealthCheck provides a simple health check endpoint
func HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// GetRoutes returns all configured routes (for admin API)
func GetRoutes(c *fiber.Ctx) error {
	// NOTE: In a real implementation, this would get routes from the router
	// For Phase 1, we'll return a simple placeholder

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Route listing is not implemented yet",
	})
}

// GetConfig returns the current configuration (for admin API)
func GetConfig(c *fiber.Ctx) error {
	// NOTE: In a real implementation, this would get the current configuration
	// For Phase 1, we'll return a simple placeholder

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Configuration API is not implemented yet",
	})
}
