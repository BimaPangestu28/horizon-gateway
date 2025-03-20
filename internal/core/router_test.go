package core

import (
	"testing"

	"github.com/bimapangestu28/horizon/internal/config"
)

// mockLogger is a simple logger for testing
type mockLogger struct{}

func (l *mockLogger) Debug(msg string, keyvals ...interface{}) {}
func (l *mockLogger) Info(msg string, keyvals ...interface{})  {}
func (l *mockLogger) Warn(msg string, keyvals ...interface{})  {}
func (l *mockLogger) Error(msg string, keyvals ...interface{}) {}
func (l *mockLogger) Fatal(msg string, keyvals ...interface{}) {}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{
			name:     "Exact match",
			path:     "/api/users",
			pattern:  "/api/users",
			expected: true,
		},
		{
			name:     "Wildcard match",
			path:     "/api/users/123",
			pattern:  "/api/*",
			expected: true,
		},
		{
			name:     "No match",
			path:     "/api/users",
			pattern:  "/api/products",
			expected: false,
		},
		{
			name:     "Root path",
			path:     "/",
			pattern:  "/*",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPath(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchPath(%s, %s) = %v; want %v", tt.path, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestIsMethodAllowed(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		allowedMethods []string
		expected       bool
	}{
		{
			name:           "Method explicitly allowed",
			method:         "GET",
			allowedMethods: []string{"GET", "POST"},
			expected:       true,
		},
		{
			name:           "Method not allowed",
			method:         "DELETE",
			allowedMethods: []string{"GET", "POST"},
			expected:       false,
		},
		{
			name:           "All methods allowed with wildcard",
			method:         "PUT",
			allowedMethods: []string{"*"},
			expected:       true,
		},
		{
			name:           "Empty allowed methods",
			method:         "GET",
			allowedMethods: []string{},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMethodAllowed(tt.method, tt.allowedMethods)
			if result != tt.expected {
				t.Errorf("isMethodAllowed(%s, %v) = %v; want %v", tt.method, tt.allowedMethods, result, tt.expected)
			}
		})
	}
}

func TestNewRouter(t *testing.T) {
	logger := &mockLogger{}

	routeConfigs := []config.RouteConfig{
		{
			Name:        "test-route",
			ListenPath:  "/api/*",
			UpstreamURL: "http://example.com",
			Methods:     []string{"GET", "POST"},
			StripPath:   true,
		},
	}

	router, err := NewRouter(routeConfigs, logger)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	if len(router.routes) != len(routeConfigs) {
		t.Errorf("NewRouter() created %d routes; want %d", len(router.routes), len(routeConfigs))
	}

	if router.routes[0].Name != routeConfigs[0].Name {
		t.Errorf("Route name = %s; want %s", router.routes[0].Name, routeConfigs[0].Name)
	}
}

func TestFindRoute(t *testing.T) {
	logger := &mockLogger{}

	routeConfigs := []config.RouteConfig{
		{
			Name:        "api-route",
			ListenPath:  "/api/*",
			UpstreamURL: "http://api.example.com",
			Methods:     []string{"GET", "POST"},
			StripPath:   true,
		},
		{
			Name:        "web-route",
			ListenPath:  "/web/*",
			UpstreamURL: "http://web.example.com",
			Methods:     []string{"*"},
			StripPath:   false,
		},
	}

	router, _ := NewRouter(routeConfigs, logger)

	tests := []struct {
		name          string
		path          string
		method        string
		expectRoute   string
		expectError   bool
		expectedError error
	}{
		{
			name:        "Find API route with GET",
			path:        "/api/users",
			method:      "GET",
			expectRoute: "api-route",
			expectError: false,
		},
		{
			name:        "Find web route with any method",
			path:        "/web/page",
			method:      "PUT",
			expectRoute: "web-route",
			expectError: false,
		},
		{
			name:          "Method not allowed",
			path:          "/api/users",
			method:        "DELETE",
			expectRoute:   "",
			expectError:   true,
			expectedError: ErrMethodNotAllowed,
		},
		{
			name:          "Route not found",
			path:          "/unknown/path",
			method:        "GET",
			expectRoute:   "",
			expectError:   true,
			expectedError: ErrRouteNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, err := router.FindRoute(tt.path, tt.method)

			if tt.expectError {
				if err == nil {
					t.Errorf("FindRoute() error = nil; want error")
				} else if err != tt.expectedError {
					t.Errorf("FindRoute() error = %v; want %v", err, tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("FindRoute() error = %v; want nil", err)
				return
			}

			if route.Name != tt.expectRoute {
				t.Errorf("FindRoute() found route = %s; want %s", route.Name, tt.expectRoute)
			}
		})
	}
}
