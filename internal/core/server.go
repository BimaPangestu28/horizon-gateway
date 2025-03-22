package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/httphandlers"
	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/middleware"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// Server represents the main API Gateway server
type Server struct {
	app            *fiber.App
	config         *config.Config
	configWatcher  *config.ConfigWatcher
	router         *Router
	logger         logging.Logger
	proxyHandler   interfaces.ProxyHandlerInterface   // Use interface
	metricsHandler interfaces.MetricsHandlerInterface // Use interface
	mu             sync.RWMutex
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, configWatcher *config.ConfigWatcher, router *Router, logger logging.Logger,
	proxyHandler interfaces.ProxyHandlerInterface, metricsHandler interfaces.MetricsHandlerInterface) (*Server, error) {
	// Create fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:             cfg.Server.ReadTimeout,
		WriteTimeout:            cfg.Server.WriteTimeout,
		IdleTimeout:             cfg.Server.IdleTimeout,
		EnableTrustedProxyCheck: true,
		ServerHeader:            "Horizon API Gateway",
	})

	// Create server
	server := &Server{
		app:            app,
		config:         cfg,
		configWatcher:  configWatcher,
		router:         router,
		logger:         logger,
		proxyHandler:   proxyHandler,
		metricsHandler: metricsHandler,
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	return server, nil
}

// setupMiddleware sets up the middleware for the server
func (s *Server) setupMiddleware() {
	// Basic middleware
	s.app.Use(recover.New())
	s.app.Use(requestid.New())
	s.app.Use(middleware.NewLogger(s.logger))
	s.app.Use(cors.New())

	// Add request start time for metrics
	s.app.Use(func(c *fiber.Ctx) error {
		c.Locals("start", time.Now())
		return c.Next()
	})

	// Add metrics middleware
	s.app.Use(middleware.MetricsMiddleware(s.logger))
}

// setupRoutes sets up the routes for the server
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.app.Get("/health", httphandlers.HealthCheck)

	// Register route handlers for all configured routes
	for _, route := range s.config.Routes {
		handlers := make([]fiber.Handler, 0)

		// Add route name to context
		handlers = append(handlers, func(routeName string) fiber.Handler {
			return func(c *fiber.Ctx) error {
				c.Locals("route", map[string]string{"name": routeName})
				return c.Next()
			}
		}(route.Name))

		// Add IP filter middleware if configured
		if route.IPFilter != nil && route.IPFilter.Enabled {
			handlers = append(handlers, middleware.IPFilterMiddleware(nil, s.logger))
		}

		// Add authentication middleware if configured
		if route.Auth != nil && route.Auth.Enabled {
			handlers = append(handlers, middleware.AuthMiddleware(nil, s.logger))
		}

		// Add rate limiting middleware if configured
		if route.RateLimiting != nil && route.RateLimiting.Enabled {
			handlers = append(handlers, middleware.RateLimitMiddleware(nil, s.logger))
		}

		// Add circuit breaker middleware if configured
		if route.CircuitBreaker != nil && route.CircuitBreaker.Enabled {
			handlers = append(handlers, middleware.CircuitBreakerMiddleware(nil, s.logger))
		}

		// Add cache middleware if configured
		if route.Caching != nil && route.Caching.Enabled {
			handlers = append(handlers, middleware.CacheMiddleware(nil, s.logger))
		}

		// Add transformation middleware if configured
		if route.Transform != nil && route.Transform.Enabled {
			handlers = append(handlers, middleware.TransformMiddleware(nil, nil, s.logger))
		}

		// Add proxy handler (using interface)
		handlers = append(handlers, func(c *fiber.Ctx) error {
			return s.proxyHandler.HandleRequest(c)
		})

		// Register the route with all middleware
		routePath := route.ListenPath
		s.app.All(routePath, handlers...)

		s.logger.Info("Registered route",
			"route", route.Name,
			"path", routePath,
			"methods", route.Methods,
			"middleware_count", len(handlers)-2) // -2 for route name handler and proxy handler
	}

	// Catch-all route
	s.app.All("/*", func(c *fiber.Ctx) error {
		return s.proxyHandler.HandleRequest(c)
	})
}

// Start starts the server
func (s *Server) Start() error {
	s.mu.RLock()
	port := s.config.Server.Port
	tlsConfig := s.config.Server.TLS
	s.mu.RUnlock()

	addr := fmt.Sprintf(":%d", port)

	if tlsConfig != nil && tlsConfig.Enabled {
		s.logger.Info("Starting HTTPS server",
			"port", port,
			"cert_file", tlsConfig.CertFile,
			"key_file", tlsConfig.KeyFile)
		return s.app.ListenTLS(addr, tlsConfig.CertFile, tlsConfig.KeyFile)
	}

	s.logger.Info("Starting HTTP server", "port", port)
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

// UpdateConfig updates the server configuration
func (s *Server) UpdateConfig(cfg *config.Config) error {
	// Implementation simplified for brevity
	// Real implementation would update all components and restart
	return nil
}
