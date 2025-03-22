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
	"github.com/bimapangestu28/horizon/internal/core"
	"github.com/bimapangestu28/horizon/internal/handlers"
	"github.com/bimapangestu28/horizon/internal/middleware"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// AdminServer represents the admin API server
type AdminServer struct {
	app           *fiber.App
	config        *config.Config
	configWatcher *config.ConfigWatcher
	logger        logging.Logger
	mu            sync.RWMutex
}

// NewAdminServer creates a new admin API server
func NewAdminServer(cfg *config.Config, configWatcher *config.ConfigWatcher, logger logging.Logger) (*AdminServer, error) {
	// Create fiber app for admin API
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: handlers.CustomErrorHandler,
	})

	// Create server
	server := &AdminServer{
		app:           app,
		config:        cfg,
		configWatcher: configWatcher,
		logger:        logger,
	}

	// Setup middleware and routes
	server.setupMiddleware()
	if err := server.setupRoutes(); err != nil {
		return nil, fmt.Errorf("setting up admin routes: %w", err)
	}

	return server, nil
}

// setupMiddleware sets up the middleware for the admin server
func (s *AdminServer) setupMiddleware() {
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

	// Add auth middleware for admin API if configured
	// TODO: Add admin authentication
}

// setupRoutes sets up the routes for the admin server
func (s *AdminServer) setupRoutes() error {
	// Create router for admin handlers
	router, err := core.NewRouter(s.config.Routes, s.logger)
	if err != nil {
		return fmt.Errorf("creating router for admin handler: %w", err)
	}

	// Create admin handlers
	adminHandler := handlers.NewAdminHandler(s.configWatcher, router, s.logger)
	adminAuthHandler := handlers.NewAdminAuthHandler(s.configWatcher, s.logger)
	adminRateLimitHandler := handlers.NewAdminRateLimitHandler(s.configWatcher, s.logger)
	adminCircuitBreakerHandler := handlers.NewAdminCircuitBreakerHandler(s.configWatcher, s.logger)
	adminCacheHandler := handlers.NewAdminCacheHandler(s.configWatcher, s.logger)
	metricsHandler := handlers.NewMetricsHandler(s.logger)

	// Health check
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Metrics endpoint
	s.app.Get("/metrics", metricsHandler.GetMetrics)

	// Admin API group
	admin := s.app.Group("/admin")

	// Routes management
	admin.Get("/routes", adminHandler.GetRoutes)
	admin.Get("/routes/:name", adminHandler.GetRoute)
	admin.Post("/routes", adminHandler.CreateRoute)
	admin.Put("/routes/:name", adminHandler.UpdateRoute)
	admin.Delete("/routes/:name", adminHandler.DeleteRoute)

	// Config management
	admin.Get("/config", adminHandler.GetConfig)
	admin.Put("/config", adminHandler.UpdateConfig)
	admin.Get("/config/backups", adminHandler.GetConfigBackups)
	admin.Post("/config/backups/:id/restore", adminHandler.RestoreConfigBackup)

	// Authentication management
	admin.Get("/auth/api-keys", adminAuthHandler.GetAPIKeys)
	admin.Post("/auth/api-keys", adminAuthHandler.CreateAPIKey)
	admin.Delete("/auth/api-keys/:route/:key", adminAuthHandler.DeleteAPIKey)
	admin.Get("/auth/jwt", adminAuthHandler.GetJWTConfigs)
	admin.Put("/auth/jwt/:route", adminAuthHandler.UpdateJWTConfig)

	// Rate limiting management
	admin.Get("/rate-limits", adminRateLimitHandler.GetRateLimits)
	admin.Get("/rate-limits/:route", adminRateLimitHandler.GetRateLimit)
	admin.Post("/rate-limits", adminRateLimitHandler.CreateRateLimit)
	admin.Put("/rate-limits/:route", adminRateLimitHandler.UpdateRateLimit)
	admin.Delete("/rate-limits/:route", adminRateLimitHandler.DeleteRateLimit)

	// Circuit breaker management
	admin.Get("/circuit-breakers", adminCircuitBreakerHandler.GetCircuitBreakers)
	admin.Get("/circuit-breakers/:route", adminCircuitBreakerHandler.GetCircuitBreaker)
	admin.Post("/circuit-breakers", adminCircuitBreakerHandler.CreateCircuitBreaker)
	admin.Put("/circuit-breakers/:route", adminCircuitBreakerHandler.UpdateCircuitBreaker)
	admin.Delete("/circuit-breakers/:route", adminCircuitBreakerHandler.DeleteCircuitBreaker)
	admin.Post("/circuit-breakers/:route/reset", adminCircuitBreakerHandler.ResetCircuitBreaker)

	// Cache management
	admin.Get("/cache", adminCacheHandler.GetCacheConfigs)
	admin.Get("/cache/:route", adminCacheHandler.GetCacheConfig)
	admin.Post("/cache", adminCacheHandler.CreateCacheConfig)
	admin.Put("/cache/:route", adminCacheHandler.UpdateCacheConfig)
	admin.Delete("/cache/:route", adminCacheHandler.DeleteCacheConfig)
	admin.Post("/cache/:route/clear", adminCacheHandler.ClearCache)

	// Version info
	admin.Get("/version", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"version": "v1.0.0",
			"build":   "20240301-1",
		})
	})

	return nil
}

// Start starts the admin server
func (s *AdminServer) Start() error {
	s.mu.RLock()
	port := s.config.Server.AdminPort
	s.mu.RUnlock()

	addr := fmt.Sprintf(":%d", port)

	s.logger.Info("Starting admin server", "port", port)
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

	s.logger.Info("Updating admin server configuration")

	// Store the new config
	s.config = cfg

	// Create a new app
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: handlers.CustomErrorHandler,
	})

	// Update app
	s.app = app

	// Setup middleware and routes
	s.setupMiddleware()
	if err := s.setupRoutes(); err != nil {
		return fmt.Errorf("setting up admin routes: %w", err)
	}

	// Restart the server
	go func() {
		// Wait a moment to ensure any ongoing requests are completed
		time.Sleep(100 * time.Millisecond)

		// Start the new server
		port := cfg.Server.AdminPort
		addr := fmt.Sprintf(":%d", port)

		s.logger.Info("Restarting admin server", "port", port)
		if err := s.app.Listen(addr); err != nil {
			s.logger.Error("Failed to restart admin server", "error", err)
		}
	}()

	return nil
}
