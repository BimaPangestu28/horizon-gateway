package server

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
	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/middleware"
	"github.com/bimapangestu28/horizon/internal/ssl"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type Server struct {
	app            *fiber.App
	config         *config.Config
	configWatcher  *config.ConfigWatcher
	router         interfaces.Router
	logger         logging.Logger
	proxyHandler   interfaces.ProxyHandlerInterface
	metricsHandler interfaces.MetricsHandlerInterface
	sslManager     *ssl.CertbotManager
	mu             sync.RWMutex
}

func NewServer(cfg *config.Config, configWatcher *config.ConfigWatcher, logger logging.Logger,
	proxyHandler interfaces.ProxyHandlerInterface, metricsHandler interfaces.MetricsHandlerInterface) (*Server, error) {

	app := fiber.New(fiber.Config{
		ReadTimeout:             cfg.Server.ReadTimeout,
		WriteTimeout:            cfg.Server.WriteTimeout,
		IdleTimeout:             cfg.Server.IdleTimeout,
		EnableTrustedProxyCheck: true,
		ServerHeader:            "Horizon API Gateway",
	})

	server := &Server{
		app:            app,
		config:         cfg,
		configWatcher:  configWatcher,
		logger:         logger,
		proxyHandler:   proxyHandler,
		metricsHandler: metricsHandler,
	}

	// Initialize SSL manager if enabled
	if cfg.Server.TLS != nil && cfg.Server.TLS.Enabled && cfg.SSLConfig != nil && cfg.SSLConfig.Certbot != nil && cfg.SSLConfig.Certbot.Enabled {
		// Create reload hook to update server on certificate changes
		reloadHook := func() error {
			return server.reloadTLSConfig()
		}

		sslManager, err := ssl.NewCertbotManager(cfg.SSLConfig.Certbot, logger, reloadHook)
		if err != nil {
			return nil, fmt.Errorf("initializing SSL manager: %w", err)
		}

		server.sslManager = sslManager

		// Update server TLS config with certificate paths
		if err := sslManager.UpdateServerConfig(&cfg.Server); err != nil {
			logger.Warn("Failed to update TLS config with Certbot certificates", "error", err)
		}
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	return server, nil
}

func (s *Server) reloadTLSConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Reloading TLS configuration with renewed certificates")

	if err := s.sslManager.UpdateServerConfig(&s.config.Server); err != nil {
		return fmt.Errorf("updating server config with new certificates: %w", err)
	}

	// We need to restart the server to apply new certificates
	// This is simplified - in production you would want a more graceful approach
	// that doesn't drop active connections
	go func() {
		time.Sleep(100 * time.Millisecond)
		// Properly shutdown and restart the server
		err := s.app.Shutdown()
		if err != nil {
			s.logger.Error("Error shutting down server for certificate reload", "error", err)
			return
		}

		port := s.config.Server.Port
		addr := fmt.Sprintf(":%d", port)

		if s.config.Server.TLS != nil && s.config.Server.TLS.Enabled {
			s.logger.Info("Restarting HTTPS server with new certificates",
				"port", port,
				"cert_file", s.config.Server.TLS.CertFile,
				"key_file", s.config.Server.TLS.KeyFile)
			err = s.app.ListenTLS(addr, s.config.Server.TLS.CertFile, s.config.Server.TLS.KeyFile)
		} else {
			s.logger.Info("Restarting HTTP server", "port", port)
			err = s.app.Listen(addr)
		}

		if err != nil {
			s.logger.Error("Error restarting server after certificate reload", "error", err)
		}
	}()

	return nil
}

func (s *Server) SetRouter(router interfaces.Router) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.router = router
}

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

func (s *Server) setupRoutes() {
	// Health check endpoint
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

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

func (s *Server) Start() error {
	// Start SSL manager if configured
	if s.sslManager != nil {
		if err := s.sslManager.Start(); err != nil {
			return fmt.Errorf("starting SSL manager: %w", err)
		}
	}

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

func (s *Server) Shutdown(ctx context.Context) error {
	// Stop SSL manager if running
	if s.sslManager != nil {
		s.sslManager.Stop()
	}

	return s.app.ShutdownWithContext(ctx)
}

func (s *Server) UpdateConfig(cfg *config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Updating server configuration")

	// Store the new config
	s.config = cfg

	// In a complete implementation, we would:
	// 1. Create a new app with updated routes and middleware
	// 2. Gracefully swap the old app with the new one
	// 3. Shutdown the old app

	// For now, just return success
	return nil
}
