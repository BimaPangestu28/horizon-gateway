package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/horizon-gateway/horizon/internal/security/auth"
	"github.com/horizon-gateway/horizon/internal/utils/logging"
)

// AuthMiddleware creates a middleware for authentication
func AuthMiddleware(authenticator auth.Authenticator, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Convert Fiber context to http.Request for the authenticator
		httpReq := &http.Request{
			Method: c.Method(),
			URL:    c.Request().URI().QueryArgs(),
			Header: make(http.Header),
		}

		// Copy headers
		c.Request().Header.VisitAll(func(key, value []byte) {
			httpReq.Header.Add(string(key), string(value))
		})

		// Authenticate the request
		metadata, err := authenticator.Authenticate(c.Context(), httpReq)
		if err != nil {
			// Handle authentication errors
			status := http.StatusUnauthorized
			message := "Unauthorized"

			switch err {
			case auth.ErrMissingAPIKey, auth.ErrMissingJWT:
				message = "Authentication credentials not provided"
			case auth.ErrInvalidAPIKey, auth.ErrInvalidJWT, auth.ErrInvalidSignature:
				message = "Invalid authentication credentials"
			case auth.ErrExpiredAPIKey, auth.ErrJWTExpired:
				message = "Authentication credentials expired"
			case auth.ErrInsufficientScope:
				status = http.StatusForbidden
				message = "Insufficient permissions"
			}

			logger.Warn("Authentication failed",
				"error", err.Error(),
				"path", c.Path(),
				"ip", c.IP(),
			)

			return c.Status(status).JSON(fiber.Map{
				"error": message,
			})
		}

		// Add authentication metadata to context
		c.Locals("auth", metadata)

		// If using JWT, add claims to headers if configured
		if jwt, ok := authenticator.(*auth.JWTAuthenticator); ok {
			jwt.AddClaimsToHeaders(httpReq, metadata)

			// Copy modified headers back to Fiber context
			for key, values := range httpReq.Header {
				for _, value := range values {
					c.Set(key, value)
				}
			}
		}

		// Call next handler
		return c.Next()
	}
}
