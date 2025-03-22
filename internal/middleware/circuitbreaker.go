package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/resilience/circuitbreaker"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// CircuitBreakerMiddleware creates a middleware for circuit breaking
func CircuitBreakerMiddleware(breaker circuitbreaker.CircuitBreaker, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if circuit breaking is not configured
		if breaker == nil {
			return c.Next()
		}

		// Check if request is allowed
		if !breaker.IsAllowed() {
			logger.Warn("Circuit breaker open, rejecting request",
				"path", c.Path(),
				"state", string(breaker.GetState()),
			)

			// Use configured status code and response
			config := breaker.GetConfig()
			status := config.ResponseStatusCode
			if status == 0 {
				status = fiber.StatusServiceUnavailable
			}

			message := config.ResponseBody
			if message == "" {
				message = "Service temporarily unavailable"
			}

			c.Set("X-Circuit-Breaker", string(breaker.GetState()))

			return c.Status(status).SendString(message)
		}

		// Record request and handle response
		breaker.RecordRequestStarted()
		startTime := time.Now()
		err := c.Next()
		breaker.RecordRequestCompleted()

		// Check for timeout
		if time.Since(startTime) > 30*time.Second {
			breaker.RecordTimeout()
			return err
		}

		// Check response status
		status := c.Response().StatusCode()
		config := breaker.GetConfig()

		// Determine failure status codes
		failureCodes := config.FailureStatusCodes
		if len(failureCodes) == 0 {
			for i := 500; i < 600; i++ {
				failureCodes = append(failureCodes, i)
			}
		}

		// Record success or failure based on status code
		isFailure := false
		for _, code := range failureCodes {
			if status == code {
				isFailure = true
				break
			}
		}

		if isFailure {
			breaker.RecordFailure()
		} else {
			breaker.RecordSuccess()
		}

		return err
	}
}
