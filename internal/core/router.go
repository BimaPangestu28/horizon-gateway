package core

import (
	"net/http"
	"sort"
	"strings"

	"github.com/bimapangestu28/horizon/internal/config"
	apierrors "github.com/bimapangestu28/horizon/internal/errors"
	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/types"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// Router handles route matching and management
type Router struct {
	routes []*types.Route
	logger logging.Logger
}

// Ensure Router implements interfaces.Router
var _ interfaces.Router = (*Router)(nil)

// NewRouter creates a new router with the given route configurations
func NewRouter(routeConfigs []config.RouteConfig, logger logging.Logger) (*Router, error) {
	router := &Router{
		routes: make([]*types.Route, 0, len(routeConfigs)),
		logger: logger,
	}

	// Process route configurations
	for _, rc := range routeConfigs {
		// Create the route
		route := &types.Route{
			Name:        rc.Name,
			ListenPath:  rc.ListenPath,
			UpstreamURL: rc.UpstreamURL,
			Methods:     rc.Methods,
			StripPath:   rc.StripPath,
			Headers:     rc.Headers,
			QueryParams: rc.QueryParams,
			Host:        rc.Host,
			Priority:    rc.Priority,
		}

		router.routes = append(router.routes, route)
		logger.Info("Route registered",
			"name", route.Name,
			"path", route.ListenPath,
			"methods", strings.Join(route.Methods, ","),
			"priority", rc.Priority,
		)
	}

	return router, nil
}

// FindRoute finds the matching route for the given request details
func (r *Router) FindRoute(req *http.Request) (*types.Route, error) {
	path := req.URL.Path
	method := req.Method
	host := req.Host

	// Store matching routes to handle priority
	var matchingRoutes []*types.Route

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
func (r *Router) GetAllRoutes() []*types.Route {
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
