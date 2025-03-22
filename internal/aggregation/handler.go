package aggregation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/gofiber/fiber/v2"
)

type AggregationHandler struct {
	router     interfaces.Router
	logger     logging.Logger
	aggregator *Aggregator
}

func NewAggregationHandler(router interfaces.Router, logger logging.Logger) *AggregationHandler {
	aggregator := NewAggregator(logger, 30*time.Second, 10*1024*1024)

	return &AggregationHandler{
		router:     router,
		logger:     logger,
		aggregator: aggregator,
	}
}

func (h *AggregationHandler) HandleRequest(c *fiber.Ctx) error {
	// Find route configuration
	route, err := h.findRoute(c)
	if err != nil {
		h.logger.Error("Route not found", "path", c.Path(), "error", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	// Get aggregation configuration from route
	aggConfig, ok := route["aggregation"].(map[string]interface{})
	if !ok || aggConfig == nil {
		h.logger.Error("Aggregation configuration not found", "route", route["name"])
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Aggregation configuration not found",
		})
	}

	// Convert configuration to AggregationConfig
	config, err := h.parseAggregationConfig(aggConfig)
	if err != nil {
		h.logger.Error("Invalid aggregation configuration",
			"route", route["name"],
			"error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid aggregation configuration",
		})
	}

	// Prepare template data from request
	templateData, err := h.prepareTemplateData(c)
	if err != nil {
		h.logger.Error("Failed to prepare template data",
			"route", route["name"],
			"error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to prepare template data",
		})
	}

	// Execute aggregation
	ctx, cancel := context.WithTimeout(c.Context(), config.Timeout)
	defer cancel()

	result, err := h.aggregator.Aggregate(ctx, config, templateData)
	if err != nil {
		h.logger.Error("Aggregation failed",
			"route", route["name"],
			"error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Aggregation failed",
			"details": err.Error(),
		})
	}

	// Set response headers
	for name, values := range result.Headers {
		// Skip certain headers
		if name == "Content-Length" || name == "Transfer-Encoding" {
			continue
		}
		for _, value := range values {
			c.Set(name, value)
		}
	}

	// Set content type
	c.Set("Content-Type", "application/json")

	// Return aggregated result
	return c.Status(result.StatusCode).JSON(result.Combined)
}

func (h *AggregationHandler) findRoute(c *fiber.Ctx) (map[string]interface{}, error) {
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

	return h.router.FindRoute(httpReq)
}

func (h *AggregationHandler) parseAggregationConfig(rawConfig map[string]interface{}) (*AggregationConfig, error) {
	// Convert raw config to JSON and then unmarshal to typed config
	jsonData, err := json.Marshal(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to encode aggregation config: %w", err)
	}

	var config AggregationConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to decode aggregation config: %w", err)
	}

	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &config, nil
}

func (h *AggregationHandler) prepareTemplateData(c *fiber.Ctx) (map[string]interface{}, error) {
	templateData := make(map[string]interface{})

	// Add query parameters
	queryParams := make(map[string]string)
	c.QueryParser(&queryParams)
	templateData["query"] = queryParams

	// Add path parameters
	pathParams := make(map[string]string)
	// In a real implementation, you'd extract path params from the route pattern
	templateData["params"] = pathParams

	// Add request headers
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	templateData["headers"] = headers

	// Add request body if present
	if len(c.Body()) > 0 {
		var bodyData map[string]interface{}
		if err := json.Unmarshal(c.Body(), &bodyData); err == nil {
			templateData["body"] = bodyData
		} else {
			// If not JSON, add as string
			templateData["bodyText"] = string(c.Body())
		}
	}

	// Add request metadata
	templateData["method"] = c.Method()
	templateData["path"] = c.Path()
	templateData["ip"] = c.IP()
	templateData["host"] = c.Hostname()
	templateData["timestamp"] = time.Now().Format(time.RFC3339)

	return templateData, nil
}
