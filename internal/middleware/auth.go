package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/security/auth"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// AuthMiddleware creates a middleware for authentication
func AuthMiddleware(authenticator auth.Authenticator, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Convert Fiber context to http.Request for the authenticator
		httpReq := &http.Request{
			Method: c.Method(),
			URL: &url.URL{
				Path:     c.Path(),
				RawQuery: string(c.Request().URI().QueryString()),
			},
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
		if jwtAuth, ok := authenticator.(*auth.JWTAuthenticator); ok {
			// Call the exported method or apply headers manually based on the config
			for k, v := range metadata {
				// Check if we have a header mapping for this claim
				if header, exists := getHeaderForClaim(jwtAuth, k); exists {
					// Convert value to string
					var strValue string
					switch val := v.(type) {
					case string:
						strValue = val
					default:
						strValue = strings.TrimSpace(strings.Replace(strings.Replace(
							strings.Replace(
								strings.TrimSpace(v.(string)),
								"\n", " ",
								-1,
							), "  ", " ",
							-1,
						), "  ", " ",
							-1,
						))
					}
					c.Set(header, strValue)
				}
			}
		}

		// Call next handler
		return c.Next()
	}
}

// Helper function to get the header name for a JWT claim
func getHeaderForClaim(jwtAuth *auth.JWTAuthenticator, claim string) (string, bool) {
	// Access the ClaimsToHeaders map via reflection or provide a public getter
	// For now, we'll use a simple map with common claim mappings
	commonMappings := map[string]string{
		"sub":   "X-User-ID",
		"name":  "X-User-Name",
		"email": "X-User-Email",
		"roles": "X-User-Roles",
		"role":  "X-User-Role",
	}

	header, exists := commonMappings[claim]
	return header, exists
}
