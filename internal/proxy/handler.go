package proxy

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

	apierrors "github.com/bimapangestu28/horizon/internal/errors"
	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/types"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

const RequestTimeout = 30 * time.Second

type Handler struct {
	router interfaces.Router
	logger logging.Logger
	client *http.Client
}

func New(router interfaces.Router, logger logging.Logger) *Handler {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: RequestTimeout,
	}

	return &Handler{
		router: router,
		logger: logger,
		client: client,
	}
}

func (h *Handler) HandleRequest(c *fiber.Ctx) error {
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

	c.Request().Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Add(string(key), string(value))
	})

	route, err := h.router.FindRoute(httpReq)
	if err != nil {
		if errors.Is(err, apierrors.ErrRouteNotFound) {
			h.logger.Error("Route not found", "path", c.Path(), "method", c.Method())
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Route not found",
			})
		}

		if errors.Is(err, apierrors.ErrMethodNotAllowed) {
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

	ctx, cancel := context.WithTimeout(c.Context(), RequestTimeout)
	defer cancel()

	resp, err := h.forwardRequest(ctx, c, route)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, apierrors.ErrProxyTimeout) {
			return c.Status(http.StatusGatewayTimeout).JSON(fiber.Map{
				"error": "Upstream service timeout",
			})
		}

		h.logger.Error("Proxy error", "error", err, "route", route.Name)
		return c.Status(http.StatusBadGateway).JSON(fiber.Map{
			"error": "Upstream service error",
		})
	}

	c.Status(resp.StatusCode)

	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read upstream response", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read upstream response",
		})
	}

	return c.Send(body)
}

func (h *Handler) forwardRequest(ctx context.Context, c *fiber.Ctx, route *types.Route) (*http.Response, error) {
	targetURL, err := url.Parse(route.UpstreamURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL: %w", err)
	}

	// Handle path rewriting
	requestPath := c.Path()
	if route.StripPath {
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
		targetURL.Path = path.Join(targetURL.Path, requestPath)
		// Preserve trailing slash which path.Join removes
		if strings.HasSuffix(requestPath, "/") && !strings.HasSuffix(targetURL.Path, "/") {
			targetURL.Path += "/"
		}
	}

	targetURL.RawQuery = string(c.Request().URI().QueryString())

	req, err := http.NewRequestWithContext(ctx, c.Method(), targetURL.String(), strings.NewReader(string(c.Body())))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	// Set proxy headers
	req.Header.Set("X-Forwarded-For", c.IP())
	req.Header.Set("X-Forwarded-Proto", c.Protocol())
	req.Header.Set("X-Forwarded-Host", c.Hostname())
	req.Header.Set("X-Gateway-Name", "Horizon")
	req.Header.Set("X-Gateway-Route", route.Name)

	h.logger.Info("Forwarding request",
		"method", c.Method(),
		"path", c.Path(),
		"target", targetURL.String(),
		"route", route.Name,
	)

	resp, err := h.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, apierrors.ErrProxyTimeout
		}
		return nil, fmt.Errorf("forwarding request: %w", err)
	}

	return resp, nil
}

func (h *Handler) GetRouter() interfaces.Router {
	return h.router
}
