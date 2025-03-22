package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/gofiber/fiber/v2"
)

type GraphQLHandler struct {
	router       interfaces.Router
	logger       logging.Logger
	clients      map[string]*http.Client
	clientsMutex sync.RWMutex
	config       *GraphQLConfig
}

type GraphQLConfig struct {
	QueryTimeout        time.Duration `yaml:"query_timeout"`
	MaxQuerySize        int           `yaml:"max_query_size"`
	EnableIntrospection bool          `yaml:"enable_introspection"`
	EnableValidation    bool          `yaml:"enable_validation"`
	EnableCaching       bool          `yaml:"enable_caching"`
}

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Extensions    map[string]interface{} `json:"extensions,omitempty"`
}

type GraphQLResponse struct {
	Data       interface{}            `json:"data,omitempty"`
	Errors     []GraphQLError         `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLLocation      `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type GraphQLLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func NewGraphQLHandler(router interfaces.Router, logger logging.Logger, config *GraphQLConfig) *GraphQLHandler {
	if config == nil {
		config = &GraphQLConfig{
			QueryTimeout:        30 * time.Second,
			MaxQuerySize:        1024 * 1024, // 1MB
			EnableIntrospection: true,
			EnableValidation:    true,
			EnableCaching:       false,
		}
	}

	return &GraphQLHandler{
		router:  router,
		logger:  logger,
		clients: make(map[string]*http.Client),
		config:  config,
	}
}

func (h *GraphQLHandler) HandleGraphQL(c *fiber.Ctx) error {
	// Check content type and request method
	contentType := string(c.Request().Header.ContentType())
	if !strings.Contains(contentType, "application/json") &&
		!strings.Contains(contentType, "application/graphql") {
		return c.Status(http.StatusUnsupportedMediaType).JSON(fiber.Map{
			"errors": []GraphQLError{
				{
					Message: "Unsupported content type",
					Extensions: map[string]interface{}{
						"code": "UNSUPPORTED_MEDIA_TYPE",
					},
				},
			},
		})
	}

	// Check request size
	if c.Request().Header.ContentLength() > int64(h.config.MaxQuerySize) {
		return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"errors": []GraphQLError{
				{
					Message: "GraphQL query is too large",
					Extensions: map[string]interface{}{
						"code": "REQUEST_TOO_LARGE",
					},
				},
			},
		})
	}

	// Parse GraphQL request
	var graphqlRequest GraphQLRequest
	if err := c.BodyParser(&graphqlRequest); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"errors": []GraphQLError{
				{
					Message: "Invalid GraphQL request",
					Extensions: map[string]interface{}{
						"code":    "BAD_REQUEST",
						"details": err.Error(),
					},
				},
			},
		})
	}

	// Check if introspection is disabled
	if !h.config.EnableIntrospection && isIntrospectionQuery(graphqlRequest.Query) {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"errors": []GraphQLError{
				{
					Message: "Introspection queries are disabled",
					Extensions: map[string]interface{}{
						"code": "INTROSPECTION_DISABLED",
					},
				},
			},
		})
	}

	// Find route
	route, err := h.findGraphQLRoute(c)
	if err != nil {
		h.logger.Error("GraphQL route not found", "path", c.Path(), "error", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"errors": []GraphQLError{
				{
					Message: "GraphQL endpoint not found",
					Extensions: map[string]interface{}{
						"code": "NOT_FOUND",
					},
				},
			},
		})
	}

	// Forward request to upstream service
	upstreamResp, err := h.forwardGraphQLRequest(c, route, graphqlRequest)
	if err != nil {
		h.logger.Error("GraphQL proxy error", "error", err, "route", route.Name)
		return c.Status(http.StatusBadGateway).JSON(fiber.Map{
			"errors": []GraphQLError{
				{
					Message: "Error proxying to GraphQL service",
					Extensions: map[string]interface{}{
						"code":    "UPSTREAM_ERROR",
						"details": err.Error(),
					},
				},
			},
		})
	}
	defer upstreamResp.Body.Close()

	// Read response body
	respBody, err := ioutil.ReadAll(upstreamResp.Body)
	if err != nil {
		h.logger.Error("Failed to read upstream response", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"errors": []GraphQLError{
				{
					Message: "Failed to read upstream response",
					Extensions: map[string]interface{}{
						"code": "INTERNAL_SERVER_ERROR",
					},
				},
			},
		})
	}

	// Set headers from upstream
	for key, values := range upstreamResp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Ensure content type is application/json
	c.Set("Content-Type", "application/json; charset=utf-8")

	// Return response
	return c.Status(upstreamResp.StatusCode).Send(respBody)
}

func (h *GraphQLHandler) findGraphQLRoute(c *fiber.Ctx) (map[string]interface{}, error) {
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

func (h *GraphQLHandler) forwardGraphQLRequest(c *fiber.Ctx, route map[string]interface{}, req GraphQLRequest) (*http.Response, error) {
	upstreamURL, ok := route["upstream_url"].(string)
	if !ok || upstreamURL == "" {
		return nil, fmt.Errorf("invalid upstream URL configuration")
	}

	// Add GraphQL endpoint path if not present in URL
	if !strings.Contains(upstreamURL, "/graphql") {
		upstreamURL = strings.TrimSuffix(upstreamURL, "/") + "/graphql"
	}

	// Serialize request to JSON
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error serializing GraphQL request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Copy headers from original request
	c.Request().Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Add(string(key), string(value))
	})

	// Make sure content type is set
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Add proxy headers
	httpReq.Header.Set("X-Forwarded-For", c.IP())
	httpReq.Header.Set("X-Forwarded-Proto", c.Protocol())
	httpReq.Header.Set("X-Forwarded-Host", c.Hostname())
	httpReq.Header.Set("X-Gateway-Name", "Horizon")
	httpReq.Header.Set("X-Gateway-Route", route["name"].(string))

	// Get or create client
	client := h.getHTTPClient(upstreamURL)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), h.config.QueryTimeout)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	// Send request
	start := time.Now()
	resp, err := client.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		h.logger.Error("GraphQL upstream request failed",
			"url", upstreamURL,
			"error", err,
			"duration_ms", duration.Milliseconds())
		return nil, err
	}

	h.logger.Info("GraphQL upstream request completed",
		"url", upstreamURL,
		"status", resp.StatusCode,
		"duration_ms", duration.Milliseconds(),
		"operation", req.OperationName)

	return resp, nil
}

func (h *GraphQLHandler) getHTTPClient(upstreamURL string) *http.Client {
	h.clientsMutex.RLock()
	client, exists := h.clients[upstreamURL]
	h.clientsMutex.RUnlock()

	if exists {
		return client
	}

	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()

	// Check again in case another goroutine created the client
	client, exists = h.clients[upstreamURL]
	if exists {
		return client
	}

	// Create new client with appropriate timeouts
	client = &http.Client{
		Timeout: h.config.QueryTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	h.clients[upstreamURL] = client
	return client
}

func isIntrospectionQuery(query string) bool {
	// Check for common introspection query patterns
	introspectionPatterns := []string{
		"__schema",
		"__type",
		"IntrospectionQuery",
	}

	for _, pattern := range introspectionPatterns {
		if strings.Contains(query, pattern) {
			return true
		}
	}

	return false
}

// ValidateGraphQLQuery performs basic validation on a GraphQL query
func (h *GraphQLHandler) ValidateGraphQLQuery(query string) []GraphQLError {
	var errors []GraphQLError

	// Check for empty query
	if strings.TrimSpace(query) == "" {
		errors = append(errors, GraphQLError{
			Message: "Query is empty",
			Extensions: map[string]interface{}{
				"code": "SYNTAX_ERROR",
			},
		})
		return errors
	}

	// Check basic syntax (opening/closing braces)
	openBraces := strings.Count(query, "{")
	closeBraces := strings.Count(query, "}")
	if openBraces != closeBraces {
		errors = append(errors, GraphQLError{
			Message: "Query has mismatched braces",
			Extensions: map[string]interface{}{
				"code":    "SYNTAX_ERROR",
				"details": fmt.Sprintf("Found %d opening braces and %d closing braces", openBraces, closeBraces),
			},
		})
	}

	// Check for missing operation type (query, mutation, subscription)
	if !strings.Contains(query, "query") &&
		!strings.Contains(query, "mutation") &&
		!strings.Contains(query, "subscription") &&
		!strings.HasPrefix(strings.TrimSpace(query), "{") {
		errors = append(errors, GraphQLError{
			Message: "Query is missing operation type",
			Extensions: map[string]interface{}{
				"code":    "SYNTAX_ERROR",
				"details": "Operation type should be 'query', 'mutation', or 'subscription'",
			},
		})
	}

	return errors
}
