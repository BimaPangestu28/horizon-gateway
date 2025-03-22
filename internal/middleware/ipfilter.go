package middleware

import (
	"net"

	"github.com/gofiber/fiber/v2"

	"github.com/horizon-gateway/horizon/internal/security/ipfilter"
	"github.com/horizon-gateway/horizon/internal/utils/logging"
)

// IPFilterMiddleware creates a middleware for IP filtering
func IPFilterMiddleware(filter ipfilter.IPFilter, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if IP filtering is not configured
		if filter == nil {
			return c.Next()
		}

		// Extract client IP
		ipStr := c.IP()

		// Parse IP
		ip := net.ParseIP(ipStr)
		if ip == nil {
			// Can't parse IP, log and default to deny for safety
			logger.Warn("Failed to parse client IP", "ip", ipStr)

			config := filter.GetConfig()
			if config.LogOnly {
				return c.Next()
			}

			return c.Status(config.ResponseStatusCode).SendString(config.ResponseBody)
		}

		// Check if allowed
		allowed := filter.IsAllowed(ip)

		// Log if not allowed
		if !allowed {
			country := filter.GetCountryForIP(ip)
			logger.Warn("IP filtered",
				"ip", ipStr,
				"country", country,
				"mode", filter.GetConfig().Mode,
				"path", c.Path(),
			)

			// If LogOnly is true, still allow the request
			if filter.GetConfig().LogOnly {
				return c.Next()
			}

			// Return forbidden response
			config := filter.GetConfig()
			return c.Status(config.ResponseStatusCode).SendString(config.ResponseBody)
		}

		// Request is allowed
		return c.Next()
	}
}
