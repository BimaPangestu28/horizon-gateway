package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/core"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// RequestTimeout defines how long a proxied request can take before timing out
const RequestTimeout = 30 * time.Second

// ProxyError represents an error that occurred during request proxying
var ErrProxyTimeout = errors.New("proxy request timed out")

// ProxyHandler handles proxying HTTP requests to upstream services
type ProxyHandler struct {
	router *core.Router
	logger logging.Logger
	client *http.Client
}

// NewProxyHandler creates a new proxy handler with the given router
func NewProxyHandler(router *core.Router, logger logging.Logger) *ProxyHandler {
	// Create HTTP client with connection pooling
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: RequestTimeout,
	}

	return &ProxyHandler{
		router: router,
		logger: logger,
		client: client,
	}
}

// HandleRequest processes incoming HTTP requests and forwards them to the
// appropriate upstream service based on route configuration
func (h *ProxyHandler) HandleRequest(c *fiber.Ctx) error {
	// Convert fiber context to http.Request for router matching
	httpReq := &http.Request{
		Method: c.Method(),
		URL: &url.URL{
			Path:     c.Path(),
			RawQuery: string(c.Request().URI().QueryString()),
		},
		Header: make(http.Header),
		Host:   c.Hostname(),
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Add(string(key), string(value))
	})

	// Find matching route using enhanced routing criteria
	route, err := h.router.FindRoute(httpReq)
	if err != nil {
		if errors.Is(err, core.ErrRouteNotFound) {
			h.logger.Error("Route not found", "path", c.Path(), "method", c.Method())
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Route not found",
			})
		}

		if errors.Is(err, core.ErrMethodNotAllowed) {
			h.logger.Error("Method not allowed", "path", c.Path(), "method", c.Method())
			return c.Status(http.StatusMethodNotAllowed).JSON(fiber.Map{
				"error": "Method not allowed",
			})
		}

		h.logger.Error("Error finding route", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Context(), RequestTimeout)
	defer cancel()

	// Forward the request to the upstream service
	resp, err := h.forwardRequest(ctx, c, route)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrProxyTimeout) {
			return c.Status(http.StatusGatewayTimeout).JSON(fiber.Map{
				"error": "Upstream service timeout",
			})
		}

		h.logger.Error("Proxy error", "error", err, "route", route.Name)
		return c.Status(http.StatusBadGateway).JSON(fiber.Map{
			"error": "Upstream service error",
		})
	}

	// Write the response status code
	c.Status(resp.StatusCode)

	// Copy headers from the upstream response
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Copy the response body
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read upstream response", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read upstream response",
		})
	}

	// Return the response body
	return c.Send(body)
}

// forwardRequest forwards the request to the upstream service
func (h *ProxyHandler) forwardRequest(ctx context.Context, c *fiber.Ctx, route *core.Route) (*http.Response, error) {
	// Get the next target from the load balancer
	target, ok := route.LoadBalancer.GetNextTarget()
	if !ok {
		return nil, fmt.Errorf("no available targets for route %s", route.Name)
	}

	// Track this connection
	route.LoadBalancer.IncrementConn(target)
	defer route.LoadBalancer.DecrementConn(target)

	// Create target URL
	targetURL := &url.URL{
		Scheme: target.URL.Scheme,
		Host:   target.URL.Host,
		Path:   target.URL.Path,
	}

	// Handle path rewriting
	requestPath := c.Path()
	if route.StripPath {
		// Strip the matching part of the path
		listenPath := strings.TrimSuffix(route.ListenPath, "/*")
		requestPath = strings.TrimPrefix(requestPath, listenPath)
		if requestPath == "" {
			requestPath = "/"
		}
	}

	// Join paths properly
	if targetURL.Path == "" {
		targetURL.Path = requestPath
	} else {
		// Ensure we don't have double slashes
		targetURL.Path = path.Join(targetURL.Path, requestPath)
	}

	// Copy query parameters
	targetURL.RawQuery = c.Request().URI().QueryArgs().String()

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, c.Method(), targetURL.String(), strings.NewReader(string(c.Body())))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Copy headers from the original request
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	// Set X-Forwarded headers
	req.Header.Set("X-Forwarded-For", c.IP())
	req.Header.Set("X-Forwarded-Proto", c.Protocol())
	req.Header.Set("X-Forwarded-Host", c.Hostname())

	// Add gateway-specific headers
	req.Header.Set("X-Gateway-Name", "Horizon")
	req.Header.Set("X-Gateway-Route", route.Name)

	// Send the request to the upstream service
	h.logger.Info("Forwarding request",
		"method", c.Method(),
		"path", c.Path(),
		"target", targetURL.String(),
		"route", route.Name,
	)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("forwarding request: %w", err)
	}

	return resp, nil
}
