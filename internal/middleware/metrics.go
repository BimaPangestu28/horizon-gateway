package middleware

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bimapangestu28/horizon/internal/metrics"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

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

func PrometheusHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		promHandler := prometheus.Handler()

		responseWriter := newResponseWriter(c)
		request := newRequest(c)

		promHandler.ServeHTTP(responseWriter, request)

		return nil
	}
}

type fiberResponseWriter struct {
	c *fiber.Ctx
}

func newResponseWriter(c *fiber.Ctx) *fiberResponseWriter {
	return &fiberResponseWriter{c: c}
}

func (f *fiberResponseWriter) Header() http.Header {
	h := http.Header{}
	f.c.Response().Header.VisitAll(func(key, value []byte) {
		h.Add(string(key), string(value))
	})
	return h
}

func (f *fiberResponseWriter) Write(data []byte) (int, error) {
	return f.c.Write(data)
}

func (f *fiberResponseWriter) WriteHeader(statusCode int) {
	f.c.Status(statusCode)
}

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
