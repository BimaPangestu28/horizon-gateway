package middleware

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/horizon-gateway/horizon/internal/transform"
	"github.com/horizon-gateway/horizon/internal/utils/logging"
)

// TransformMiddleware creates middleware for request/response transformation
func TransformMiddleware(reqTransformer *transform.RequestTransformer, respTransformer *transform.ResponseTransformer, logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if no transformations configured
		if reqTransformer == nil && respTransformer == nil {
			return c.Next()
		}

		// Apply request transformations
		if reqTransformer != nil {
			if err := transformRequest(c, reqTransformer, logger); err != nil {
				logger.Error("Request transformation failed", "error", err)
				return fiber.NewError(fiber.StatusInternalServerError, "Request transformation failed")
			}
		}

		// For response transformations, we need to capture the response
		if respTransformer != nil {
			// Store the original response
			originalBody := c.Response().Body()
			originalStatusCode := c.Response().StatusCode()
			originalHeaders := make(map[string]string)

			c.Response().Header.VisitAll(func(key, value []byte) {
				originalHeaders[string(key)] = string(value)
			})

			// Process the request
			err := c.Next()

			// Check for errors from the handler
			if err != nil {
				return err // Let Fiber handle the error
			}

			// Now transform the response
			if err := transformResponse(c, respTransformer, logger); err != nil {
				logger.Error("Response transformation failed", "error", err)

				// Restore original response in case of transformation error
				c.Status(originalStatusCode)

				for key, value := range originalHeaders {
					c.Set(key, value)
				}

				return c.Send(originalBody)
			}

			return nil // Response already sent
		}

		// If no response transformation, just call next handler
		return c.Next()
	}
}

// transformRequest applies request transformations
func transformRequest(c *fiber.Ctx, transformer *transform.RequestTransformer, logger logging.Logger) error {
	// Convert Fiber context to http.Request for transformation
	httpReq, err := convertFiberToHTTPRequest(c)
	if err != nil {
		return fmt.Errorf("failed to convert Fiber context to HTTP request: %w", err)
	}

	// Apply transformations
	if err := transformer.TransformRequest(httpReq); err != nil {
		return fmt.Errorf("request transformation error: %w", err)
	}

	// Apply transformed request back to Fiber context
	if err := applyHTTPRequestToFiber(httpReq, c); err != nil {
		return fmt.Errorf("failed to apply transformed request to Fiber context: %w", err)
	}

	return nil
}

// transformResponse applies response transformations
func transformResponse(c *fiber.Ctx, transformer *transform.ResponseTransformer, logger logging.Logger) error {
	// Convert Fiber response to http.Response for transformation
	httpResp, err := convertFiberToHTTPResponse(c)
	if err != nil {
		return fmt.Errorf("failed to convert Fiber response to HTTP response: %w", err)
	}

	// Apply transformations
	if err := transformer.TransformResponse(httpResp); err != nil {
		return fmt.Errorf("response transformation error: %w", err)
	}

	// Apply transformed response back to Fiber context
	if err := applyHTTPResponseToFiber(httpResp, c); err != nil {
		return fmt.Errorf("failed to apply transformed response to Fiber context: %w", err)
	}

	return nil
}

// convertFiberToHTTPRequest converts a Fiber context to a standard http.Request
func convertFiberToHTTPRequest(c *fiber.Ctx) (*http.Request, error) {
	// Parse URL
	reqURL, err := url.Parse(string(c.Request().URI().FullURI()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse request URL: %w", err)
	}

	// Create request with body
	httpReq := &http.Request{
		Method: c.Method(),
		URL:    reqURL,
		Header: make(http.Header),
		Host:   c.Hostname(),
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Add(string(key), string(value))
	})

	// Add body if present
	if len(c.Body()) > 0 {
		httpReq.Body = ioutil.NopCloser(bytes.NewReader(c.Body()))
		httpReq.ContentLength = int64(len(c.Body()))
	}

	return httpReq, nil
}

// applyHTTPRequestToFiber applies a transformed http.Request back to the Fiber context
func applyHTTPRequestToFiber(req *http.Request, c *fiber.Ctx) error {
	// Set method if changed
	if c.Method() != req.Method {
		c.Method(req.Method)
	}

	// Set path if changed
	if string(c.Request().RequestURI()) != req.URL.Path {
		// In Fiber, we can't directly change the path, but we can set the original URL
		// This might need adjustment based on your routing requirements
		c.Path(req.URL.Path)
	}

	// Set query if changed
	if req.URL.RawQuery != "" {
		// In Fiber, we can update the query string using the original request
		c.Request().URI().SetQueryString(req.URL.RawQuery)
	}

	// Set headers
	// First clear existing headers (except those Fiber needs)
	c.Request().Header.VisitAll(func(key, _ []byte) {
		headerKey := string(key)
		// Skip necessary Fiber headers
		if !strings.HasPrefix(headerKey, "Fiber-") {
			c.Request().Header.Del(headerKey)
		}
	})

	// Then add new headers
	for key, values := range req.Header {
		for _, value := range values {
			c.Request().Header.Set(key, value)
		}
	}

	// Set body if present
	if req.Body != nil {
		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("failed to read transformed request body: %w", err)
		}
		req.Body.Close() // Close original body reader

		if len(bodyBytes) > 0 {
			c.Request().SetBody(bodyBytes)
		}
	}

	return nil
}

// convertFiberToHTTPResponse converts a Fiber response to a standard http.Response
func convertFiberToHTTPResponse(c *fiber.Ctx) (*http.Response, error) {
	// Create response with body
	httpResp := &http.Response{
		StatusCode: c.Response().StatusCode(),
		Header:     make(http.Header),
		Request:    &http.Request{}, // Empty request to avoid nil pointer
	}

	// Copy headers
	c.Response().Header.VisitAll(func(key, value []byte) {
		httpResp.Header.Add(string(key), string(value))
	})

	// Add body if present
	body := c.Response().Body()
	if len(body) > 0 {
		httpResp.Body = ioutil.NopCloser(bytes.NewReader(body))
		httpResp.ContentLength = int64(len(body))
	} else {
		// Ensure we always have a body, even if empty
		httpResp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
		httpResp.ContentLength = 0
	}

	return httpResp, nil
}

// applyHTTPResponseToFiber applies a transformed http.Response back to the Fiber context
func applyHTTPResponseToFiber(resp *http.Response, c *fiber.Ctx) error {
	// Set status code
	c.Status(resp.StatusCode)

	// Clear existing headers (except necessary ones)
	c.Response().Header.VisitAll(func(key, _ []byte) {
		headerKey := string(key)
		// Skip necessary Fiber headers
		if !strings.HasPrefix(headerKey, "Fiber-") {
			c.Response().Header.Del(headerKey)
		}
	})

	// Set headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Set body if present
	if resp.Body != nil {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read transformed response body: %w", err)
		}
		resp.Body.Close() // Close original body reader

		if len(bodyBytes) > 0 {
			// Send the response body - this will write to the client
			return c.Send(bodyBytes)
		}
	}

	return nil
}
