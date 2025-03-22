package handlers

import (
	"bytes"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/metrics"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// MetricsHandler implements metrics handling functionality
type MetricsHandler struct {
	logger logging.Logger
}

// Ensure MetricsHandler implements interfaces.MetricsHandlerInterface
var _ interfaces.MetricsHandlerInterface = (*MetricsHandler)(nil)

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger logging.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

// GetMetrics returns Prometheus metrics
func (h *MetricsHandler) GetMetrics(c *fiber.Ctx) error {
	promHandler := promhttp.Handler()

	responseWriter := &responseWriter{c: c}
	request := &http.Request{
		Method:     c.Method(),
		URL:        &url.URL{Path: c.Path()},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       c.Hostname(),
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		request.Header.Add(string(key), string(value))
	})

	promHandler.ServeHTTP(responseWriter, request)

	return nil
}

// RecordRequest records request metrics
func (h *MetricsHandler) RecordRequest(c *fiber.Ctx, routeName string) error {
	path := c.Path()
	method := c.Method()

	start, ok := c.Locals("start").(time.Time)
	if !ok {
		start = time.Now()
	}

	duration := time.Since(start).Seconds()
	status := c.Response().StatusCode()
	reqSize := len(c.Request().Body())
	respSize := len(c.Response().Body())

	metrics.RecordRequest(method, path, routeName, status, duration, reqSize, respSize)

	return nil
}

// RecordCacheActivity records cache hit/miss metrics
func (h *MetricsHandler) RecordCacheActivity(routeName, cacheType string, hit bool) {
	metrics.RecordCacheActivity(routeName, cacheType, hit)
}

// RecordRateLimit records rate limiting metrics
func (h *MetricsHandler) RecordRateLimit(routeName, limitType, scope string) {
	metrics.RecordRateLimit(routeName, limitType, scope)
}

// RecordAuthentication records authentication metrics
func (h *MetricsHandler) RecordAuthentication(routeName, authType string, success bool, reason string) {
	metrics.RecordAuthentication(routeName, authType, success, reason)
}

// RecordCircuitBreakerTrip records circuit breaker trip metrics
func (h *MetricsHandler) RecordCircuitBreakerTrip(routeName, breakerType string) {
	metrics.RecordCircuitBreakerTrip(routeName, breakerType)
}

// SetCircuitBreakerState sets circuit breaker state metrics
func (h *MetricsHandler) SetCircuitBreakerState(routeName, breakerType string, state int) {
	metrics.SetCircuitBreakerState(routeName, breakerType, state)
}

// SetUpstreamHealth sets upstream health metrics
func (h *MetricsHandler) SetUpstreamHealth(upstream, routeName string, healthy bool) {
	metrics.SetUpstreamHealth(upstream, routeName, healthy)
}

// SetUpstreamConnections sets upstream connection metrics
func (h *MetricsHandler) SetUpstreamConnections(upstream, routeName string, connections int) {
	metrics.SetUpstreamConnections(upstream, routeName, connections)
}

// RecordConfigReload records config reload metrics
func (h *MetricsHandler) RecordConfigReload() {
	metrics.RecordConfigReload()
}

// RecordUpstreamRequest records upstream request metrics
func (h *MetricsHandler) RecordUpstreamRequest(upstream, method, routeName string, duration float64) {
	metrics.RecordUpstreamRequest(upstream, method, routeName, duration)
}

// responseWriter adapts fiber.Ctx to http.ResponseWriter for Prometheus handler
type responseWriter struct {
	c      *fiber.Ctx
	status int
	buffer bytes.Buffer
}

func (w *responseWriter) Header() http.Header {
	header := http.Header{}
	w.c.Response().Header.VisitAll(func(key, value []byte) {
		header.Add(string(key), string(value))
	})
	return header
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.buffer.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.status == 0 {
		w.status = statusCode
		w.c.Status(statusCode)
	}
}

func (w *responseWriter) Flush() {
	if w.buffer.Len() > 0 {
		w.c.Send(w.buffer.Bytes())
		w.buffer.Reset()
	}
}
