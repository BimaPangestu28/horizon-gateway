package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "horizon_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"method", "path", "route_name", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "horizon_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "route_name", "status"},
	)

	UpstreamRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "horizon_upstream_request_duration_seconds",
			Help:    "Upstream request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"upstream", "method", "route_name"},
	)

	RequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "horizon_request_size_bytes",
			Help:    "Request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path", "route_name"},
	)

	ResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "horizon_response_size_bytes",
			Help:    "Response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path", "route_name", "status"},
	)

	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "horizon_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"route_name", "cache_type"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "horizon_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"route_name", "cache_type"},
	)

	RateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "horizon_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"route_name", "limit_type", "scope"},
	)

	AuthSuccesses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "horizon_auth_successes_total",
			Help: "Total number of successful authentications",
		},
		[]string{"route_name", "auth_type"},
	)

	AuthFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "horizon_auth_failures_total",
			Help: "Total number of failed authentications",
		},
		[]string{"route_name", "auth_type", "reason"},
	)

	CircuitBreakerTrips = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "horizon_circuit_breaker_trips_total",
			Help: "Total number of circuit breaker trips",
		},
		[]string{"route_name", "breaker_type"},
	)

	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "horizon_circuit_breaker_state",
			Help: "Current state of circuit breakers (0=closed, 1=half-open, 2=open)",
		},
		[]string{"route_name", "breaker_type"},
	)

	UpstreamHealthy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "horizon_upstream_healthy",
			Help: "Health status of upstream services (0=unhealthy, 1=healthy)",
		},
		[]string{"upstream", "route_name"},
	)

	UpstreamActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "horizon_upstream_active_connections",
			Help: "Number of active connections to upstream services",
		},
		[]string{"upstream", "route_name"},
	)

	ConfigReloads = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "horizon_config_reloads_total",
			Help: "Total number of configuration reloads",
		},
	)
)

func RecordRequest(method, path, routeName string, status int, duration float64, requestSize, responseSize int) {
	statusStr := string(status)

	RequestTotal.WithLabelValues(method, path, routeName, statusStr).Inc()
	RequestDuration.WithLabelValues(method, path, routeName, statusStr).Observe(duration)
	RequestSize.WithLabelValues(method, path, routeName).Observe(float64(requestSize))
	ResponseSize.WithLabelValues(method, path, routeName, statusStr).Observe(float64(responseSize))
}

func RecordCacheActivity(routeName, cacheType string, hit bool) {
	if hit {
		CacheHits.WithLabelValues(routeName, cacheType).Inc()
	} else {
		CacheMisses.WithLabelValues(routeName, cacheType).Inc()
	}
}

func RecordRateLimit(routeName, limitType, scope string) {
	RateLimitHits.WithLabelValues(routeName, limitType, scope).Inc()
}

func RecordAuthentication(routeName, authType string, success bool, reason string) {
	if success {
		AuthSuccesses.WithLabelValues(routeName, authType).Inc()
	} else {
		AuthFailures.WithLabelValues(routeName, authType, reason).Inc()
	}
}

func RecordCircuitBreakerTrip(routeName, breakerType string) {
	CircuitBreakerTrips.WithLabelValues(routeName, breakerType).Inc()
}

func SetCircuitBreakerState(routeName, breakerType string, state int) {
	CircuitBreakerState.WithLabelValues(routeName, breakerType).Set(float64(state))
}

func SetUpstreamHealth(upstream, routeName string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	UpstreamHealthy.WithLabelValues(upstream, routeName).Set(value)
}

func SetUpstreamConnections(upstream, routeName string, connections int) {
	UpstreamActiveConnections.WithLabelValues(upstream, routeName).Set(float64(connections))
}

func RecordConfigReload() {
	ConfigReloads.Inc()
}

func RecordUpstreamRequest(upstream, method, routeName string, duration float64) {
	UpstreamRequestDuration.WithLabelValues(upstream, method, routeName).Observe(duration)
}
