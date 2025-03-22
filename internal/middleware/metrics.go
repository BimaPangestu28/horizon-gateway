package middleware

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bimapangestu28/horizon/internal/metrics"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// MetricsMiddleware creates middleware for recording metrics
func MetricsMiddleware(logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		path := c.Path()
		method := c.Method()

		reqSize := len(c.Request().Body())

		routeName := "unknown"
		if routeData := c.Locals("route"); routeData != nil {
			if route, ok := routeData.(map[string]string); ok {
				if name, exists := route["name"]; exists {
					routeName = name
				}
			}
		}

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()
		respSize := len(c.Response().Body())

		metrics.RecordRequest(method, path, routeName, status, duration, reqSize, respSize)

		return err
	}
}

// PrometheusHandler returns a handler for Prometheus metrics
func PrometheusHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		promHandler := promhttp.Handler()

		responseWriter := newResponseWriter(c)
		request := newRequest(c)

		promHandler.ServeHTTP(responseWriter, request)

		return nil
	}
}

// fiberResponseWriter adapts fiber.Ctx to http.ResponseWriter
type fiberResponseWriter struct {
	c *fiber.Ctx
}

// newResponseWriter creates a new response writer adapter
func newResponseWriter(c *fiber.Ctx) *fiberResponseWriter {
	return &fiberResponseWriter{c: c}
}

// Header returns the header map to update
func (f *fiberResponseWriter) Header() http.Header {
	h := http.Header{}
	f.c.Response().Header.VisitAll(func(key, value []byte) {
		h.Add(string(key), string(value))
	})
	return h
}

// Write writes data to the response
func (f *fiberResponseWriter) Write(data []byte) (int, error) {
	return f.c.Write(data)
}

// WriteHeader sets the status code for the response
func (f *fiberResponseWriter) WriteHeader(statusCode int) {
	f.c.Status(statusCode)
}

// newRequest creates a new http.Request from a fiber.Ctx
func newRequest(c *fiber.Ctx) *http.Request {
	r := &http.Request{
		Method:     c.Method(),
		URL:        &url.URL{Path: c.Path()},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       c.Hostname(),
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		r.Header.Add(string(key), string(value))
	})

	return r
}
