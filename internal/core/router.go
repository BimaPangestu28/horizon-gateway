package core

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bimapangestu28/horizon/internal/config"
	apierrors "github.com/bimapangestu28/horizon/internal/errors"
	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/types"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

var (
	// ErrRouteNotFound is returned when no matching route is found
	ErrRouteNotFound = errors.New("route not found")

	// ErrMethodNotAllowed is returned when the route is found but the method is not allowed
	ErrMethodNotAllowed = errors.New("method not allowed")
)

// Router handles route matching and management
type Router struct {
	routes []*types.Route
	logger logging.Logger
}

var _ interfaces.Router = (*Router)(nil)

// Route represents a configured API route
type Route struct {
	Name         string
	ListenPath   string
	UpstreamURL  string
	Methods      []string
	StripPath    bool
	LoadBalancer *LoadBalancer
	Headers      map[string]string // Required headers for matching
	QueryParams  map[string]string // Required query parameters for matching
	Host         string            // Host-based routing
	Priority     int               // Route priority for conflict resolution
}

// NewRouter creates a new router with the given route configurations
func NewRouter(routeConfigs []config.RouteConfig, logger logging.Logger) (*Router, error) {
	router := &Router{
		routes: make([]*Route, 0, len(routeConfigs)),
		logger: logger,
	}

	// Process route configurations
	for _, rc := range routeConfigs {
		// Create load balancer for this route
		lbType := LBTypeRoundRobin
		if rc.LoadBalancing != nil && rc.LoadBalancing.Type != "" {
			switch rc.LoadBalancing.Type {
			case "weighted":
				lbType = LBTypeWeighted
			case "least_conn":
				lbType = LBTypeLeastConn
			default:
				lbType = LBTypeRoundRobin
			}
		}

		lb := NewLoadBalancer(lbType)

		// Add targets for load balancing
		if rc.LoadBalancing != nil && len(rc.LoadBalancing.Targets) > 0 {
			// Add each configured target
			for _, target := range rc.LoadBalancing.Targets {
				weight := target.Weight
				if weight <= 0 {
					weight = 1 // Default weight
				}

				err := lb.AddTarget(target.URL, weight)
				if err != nil {
					return nil, fmt.Errorf("adding target %s for route %s: %w", target.URL, rc.Name, err)
				}
			}
		} else {
			// Add the upstream URL as a single target if no targets defined
			err := lb.AddTarget(rc.UpstreamURL, 1)
			if err != nil {
				return nil, fmt.Errorf("adding target %s for route %s: %w", rc.UpstreamURL, rc.Name, err)
			}
		}

		// Create the route
		route := &Route{
			Name:         rc.Name,
			ListenPath:   rc.ListenPath,
			UpstreamURL:  rc.UpstreamURL,
			Methods:      rc.Methods,
			StripPath:    rc.StripPath,
			LoadBalancer: lb,
			Headers:      rc.Headers,
			QueryParams:  rc.QueryParams,
			Host:         rc.Host,
			Priority:     rc.Priority,
		}

		router.routes = append(router.routes, route)
		logger.Info("Route registered",
			"name", route.Name,
			"path", route.ListenPath,
			"methods", strings.Join(route.Methods, ","),
			"priority", rc.Priority,
		)
	}

	// Set up health checks for all routes with configured health checks
	for _, rc := range routeConfigs {
		if rc.LoadBalancing != nil && rc.LoadBalancing.HealthCheck != nil {
			// Find the corresponding route
			var route *Route
			for _, r := range router.routes {
				if r.Name == rc.Name {
					route = r
					break
				}
			}

			if route != nil {
				// Parse interval and timeout
				interval, _ := time.ParseDuration(rc.LoadBalancing.HealthCheck.Interval)
				timeout, _ := time.ParseDuration(rc.LoadBalancing.HealthCheck.Timeout)

				// Use defaults if parsing fails
				if interval <= 0 {
					interval = 10 * time.Second
				}
				if timeout <= 0 {
					timeout = 5 * time.Second
				}

				// Create health check config
				healthCheckConfig := HealthCheckConfig{
					Path:               rc.LoadBalancing.HealthCheck.Path,
					Interval:           interval,
					Timeout:            timeout,
					HealthyThreshold:   rc.LoadBalancing.HealthCheck.HealthyThreshold,
					UnhealthyThreshold: rc.LoadBalancing.HealthCheck.UnhealthyThreshold,
				}

				// Apply defaults if needed
				if healthCheckConfig.Path == "" {
					healthCheckConfig.Path = "/health"
				}
				if healthCheckConfig.HealthyThreshold <= 0 {
					healthCheckConfig.HealthyThreshold = 2
				}
				if healthCheckConfig.UnhealthyThreshold <= 0 {
					healthCheckConfig.UnhealthyThreshold = 3
				}

				// Create and start health checker
				healthChecker := NewHealthChecker(route.LoadBalancer, healthCheckConfig, logger)
				healthChecker.Start()

				logger.Info("Health checker started for route",
					"route", route.Name,
					"path", healthCheckConfig.Path,
					"interval", interval,
				)
			}
		}
	}

	return router, nil
}

// FindRoute finds the matching route for the given request details
func (r *Router) FindRoute(req *http.Request) (*Route, error) {
	path := req.URL.Path
	method := req.Method
	host := req.Host

	// Store matching routes to handle priority
	var matchingRoutes []*Route

	for _, route := range r.routes {
		// Check path match first (most common filter)
		if !matchPath(path, route.ListenPath) {
			continue
		}

		// Check if method is allowed
		if !isMethodAllowed(method, route.Methods) {
			return nil, apierrors.ErrMethodNotAllowed
		}

		// Check host if specified
		if route.Host != "" && route.Host != host {
			continue
		}

		// Check required headers
		headersMatch := true
		for headerKey, headerValue := range route.Headers {
			if req.Header.Get(headerKey) != headerValue {
				headersMatch = false
				break
			}
		}
		if !headersMatch {
			continue
		}

		// Check required query parameters
		queryParamsMatch := true
		query := req.URL.Query()
		for paramKey, paramValue := range route.QueryParams {
			if query.Get(paramKey) != paramValue {
				queryParamsMatch = false
				break
			}
		}
		if !queryParamsMatch {
			continue
		}

		// All criteria matched, add to matching routes
		matchingRoutes = append(matchingRoutes, route)
	}

	// No matches found
	if len(matchingRoutes) == 0 {
		return nil, apierrors.ErrRouteNotFound
	}

	// If multiple matches, select the one with highest priority
	if len(matchingRoutes) > 1 {
		// Sort by priority (highest first)
		sort.Slice(matchingRoutes, func(i, j int) bool {
			return matchingRoutes[i].Priority > matchingRoutes[j].Priority
		})
	}

	return matchingRoutes[0], nil
}

// GetAllRoutes returns all configured routes
func (r *Router) GetAllRoutes() []*Route {
	return r.routes
}

// matchPath checks if the path matches the route pattern
// This is a simple implementation that supports wildcard matching
func matchPath(path, pattern string) bool {
	// Convert pattern to prefix match if it ends with a wildcard
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}

	// Exact match for non-wildcard patterns
	return path == pattern
}

// isMethodAllowed checks if the HTTP method is allowed for the route
func isMethodAllowed(method string, allowedMethods []string) bool {
	// If the route has "*" in allowed methods, all methods are allowed
	for _, m := range allowedMethods {
		if m == "*" {
			return true
		}

		if m == method {
			return true
		}
	}

	return false
}
