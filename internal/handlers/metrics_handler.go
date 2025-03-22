package handlers

import (
	"bytes"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bimapangestu28/horizon/internal/metrics"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type MetricsHandler struct {
	logger logging.Logger
}

func NewMetricsHandler(logger logging.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

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

func (h *MetricsHandler) RecordRequest(c *fiber.Ctx, routeName string) error {
	path := c.Path()
	method := c.Method()

	duration := time.Since(time.Time(c.Locals("start").(time.Time))).Seconds()
	status := c.Response().StatusCode()
	reqSize := len(c.Request().Body())
	respSize := len(c.Response().Body())

	metrics.RecordRequest(method, path, routeName, status, duration, reqSize, respSize)

	return nil
}

func (h *MetricsHandler) RecordCacheActivity(routeName, cacheType string, hit bool) {
	metrics.RecordCacheActivity(routeName, cacheType, hit)
}

func (h *MetricsHandler) RecordRateLimit(routeName, limitType, scope string) {
	metrics.RecordRateLimit(routeName, limitType, scope)
}

func (h *MetricsHandler) RecordAuthentication(routeName, authType string, success bool, reason string) {
	metrics.RecordAuthentication(routeName, authType, success, reason)
}

func (h *MetricsHandler) RecordCircuitBreakerTrip(routeName, breakerType string) {
	metrics.RecordCircuitBreakerTrip(routeName, breakerType)
}

func (h *MetricsHandler) SetCircuitBreakerState(routeName, breakerType string, state int) {
	metrics.SetCircuitBreakerState(routeName, breakerType, state)
}

func (h *MetricsHandler) SetUpstreamHealth(upstream, routeName string, healthy bool) {
	metrics.SetUpstreamHealth(upstream, routeName, healthy)
}

func (h *MetricsHandler) SetUpstreamConnections(upstream, routeName string, connections int) {
	metrics.SetUpstreamConnections(upstream, routeName, connections)
}

func (h *MetricsHandler) RecordConfigReload() {
	metrics.RecordConfigReload()
}

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
