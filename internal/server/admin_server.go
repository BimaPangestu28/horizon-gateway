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
	"github.com/bimapangestu28/horizon/internal/handlers"
	"github.com/bimapangestu28/horizon/internal/httphandlers"
	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/middleware"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// AdminServer represents the admin API server
type AdminServer struct {
	app            *fiber.App
	config         *config.Config
	configWatcher  *config.ConfigWatcher
	router         interfaces.Router
	logger         logging.Logger
	metricsHandler interfaces.MetricsHandlerInterface
	mu             sync.RWMutex
}

// NewAdminServer creates a new admin server instance
func NewAdminServer(cfg *config.Config, configWatcher *config.ConfigWatcher, router interfaces.Router,
	logger logging.Logger, metricsHandler interfaces.MetricsHandlerInterface) (*AdminServer, error) {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: handlers.CustomErrorHandler,
	})

	server := &AdminServer{
		app:            app,
		config:         cfg,
		configWatcher:  configWatcher,
		router:         router,
		logger:         logger,
		metricsHandler: metricsHandler,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server, nil
}

// setupMiddleware sets up the middleware for the admin server
func (s *AdminServer) setupMiddleware() {
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

// setupRoutes sets up the routes for the admin server
func (s *AdminServer) setupRoutes() {
	// Health check endpoint
	s.app.Get("/health", httphandlers.HealthCheck)

	// Admin API
	admin := s.app.Group("/admin")

	// Routes management
	routesHandler := handlers.NewAdminHandler(s.configWatcher, s.logger)
	admin.Get("/routes", routesHandler.GetRoutes)
	admin.Get("/routes/:name", routesHandler.GetRoute)
	admin.Post("/routes", routesHandler.CreateRoute)
	admin.Put("/routes/:name", routesHandler.UpdateRoute)
	admin.Delete("/routes/:name", routesHandler.DeleteRoute)

	// Authentication management
	authHandler := handlers.NewAdminAuthHandler(s.configWatcher, s.logger)
	admin.Get("/auth/api-keys", authHandler.GetAPIKeys)
	admin.Post("/auth/api-keys", authHandler.CreateAPIKey)
	admin.Delete("/auth/api-keys/:route/:key", authHandler.DeleteAPIKey)
	admin.Get("/auth/jwt", authHandler.GetJWTConfigs)
	admin.Put("/auth/jwt/:route", authHandler.UpdateJWTConfig)

	// Rate limiting management
	rateLimitHandler := handlers.NewAdminRateLimitHandler(s.configWatcher, s.logger)
	admin.Get("/rate-limits", rateLimitHandler.GetRateLimits)
	admin.Get("/rate-limits/:route", rateLimitHandler.GetRateLimit)
	admin.Post("/rate-limits", rateLimitHandler.CreateRateLimit)
	admin.Put("/rate-limits/:route", rateLimitHandler.UpdateRateLimit)
	admin.Delete("/rate-limits/:route", rateLimitHandler.DeleteRateLimit)

	// Circuit breaker management
	circuitBreakerHandler := handlers.NewAdminCircuitBreakerHandler(s.configWatcher, s.logger)
	admin.Get("/circuit-breakers", circuitBreakerHandler.GetCircuitBreakers)
	admin.Get("/circuit-breakers/:route", circuitBreakerHandler.GetCircuitBreaker)
	admin.Post("/circuit-breakers", circuitBreakerHandler.CreateCircuitBreaker)
	admin.Put("/circuit-breakers/:route", circuitBreakerHandler.UpdateCircuitBreaker)
	admin.Delete("/circuit-breakers/:route", circuitBreakerHandler.DeleteCircuitBreaker)
	admin.Post("/circuit-breakers/:route/reset", circuitBreakerHandler.ResetCircuitBreaker)

	// Cache management
	cacheHandler := handlers.NewAdminCacheHandler(s.configWatcher, s.logger)
	admin.Get("/cache", cacheHandler.GetCacheConfigs)
	admin.Get("/cache/:route", cacheHandler.GetCacheConfig)
	admin.Post("/cache", cacheHandler.CreateCacheConfig)
	admin.Put("/cache/:route", cacheHandler.UpdateCacheConfig)
	admin.Delete("/cache/:route", cacheHandler.DeleteCacheConfig)
	admin.Post("/cache/:route/clear", cacheHandler.ClearCache)

	// Configuration management
	admin.Get("/config", handlers.GetConfig)

	// Metrics endpoint
	admin.Get("/metrics", func(c *fiber.Ctx) error {
		return s.metricsHandler.GetMetrics(c)
	})
}

// Start starts the admin server
func (s *AdminServer) Start() error {
	s.mu.RLock()
	port := s.config.Server.AdminPort
	s.mu.RUnlock()

	addr := fmt.Sprintf(":%d", port)
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the admin server
func (s *AdminServer) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

// UpdateConfig updates the admin server configuration
func (s *AdminServer) UpdateConfig(cfg *config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = cfg

	// For full implementation, would recreate and restart server
	// with new configuration here

	return nil
}
