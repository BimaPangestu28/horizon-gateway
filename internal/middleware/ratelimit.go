package middleware

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/horizon-gateway/horizon/internal/security/ratelimit"
	"github.com/horizon-gateway/horizon/internal/utils/logging"
)

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware(limiter ratelimit.RateLimiter, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if rate limiting is not configured
		if limiter == nil {
			return c.Next()
		}

		// Convert Fiber context to http.Request for the limiter
		httpReq := &http.Request{
			Method:     c.Method(),
			RemoteAddr: c.IP(),
			Header:     make(http.Header),
		}

		// Copy headers
		c.Request().Header.VisitAll(func(key, value []byte) {
			httpReq.Header.Add(string(key), string(value))
		})

		// Get authentication data from context (if available)
		var authData map[string]interface{}
		if auth, ok := c.Locals("auth").(map[string]interface{}); ok {
			authData = auth
		}

		// Extract the rate limiting key
		key := limiter.ExtractKey(httpReq, authData)

		// Check if request is allowed
		allowed, remaining, resetTime := limiter.Allow(key)

		// Add rate limit headers if enabled
		config := limiter.GetConfig()
		if config.EnableHeaders {
			c.Set("X-RateLimit-Limit", strconv.Itoa(config.Limit))
			c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		}

		// If not allowed, return rate limit error
		if !allowed {
			logger.Warn("Rate limit exceeded",
				"key", key,
				"path", c.Path(),
				"ip", c.IP(),
			)

			// Use configured status code and response
			status := config.ResponseStatusCode
			if status == 0 {
				status = http.StatusTooManyRequests
			}

			message := config.ResponseBody
			if message == "" {
				message = "Rate limit exceeded"
			}

			return c.Status(status).SendString(message)
		}

		// Call next handler
		return c.Next()
	}
}
