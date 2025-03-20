package httphandlers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthCheck provides a simple health check endpoint
func HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
