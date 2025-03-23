package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

type AdminServer struct {
	app           *fiber.App
	config        *config.Config
	configWatcher *config.ConfigWatcher
	logger        logging.Logger
	mu            sync.RWMutex
}

func NewAdminServer(cfg *config.Config, configWatcher *config.ConfigWatcher, logger logging.Logger) (*AdminServer, error) {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: handlers.CustomErrorHandler,
	})

	server := &AdminServer{
		app:           app,
		config:        cfg,
		configWatcher: configWatcher,
		logger:        logger,
	}

	server.setupMiddleware()
	if err := server.setupRoutes(); err != nil {
		return nil, fmt.Errorf("setting up admin routes: %w", err)
	}

	return server, nil
}

func (s *AdminServer) setupMiddleware() {
	s.app.Use(recover.New())
	s.app.Use(requestid.New())
	s.app.Use(middleware.NewLogger(s.logger))
	s.app.Use(cors.New())

	s.app.Use(func(c *fiber.Ctx) error {
		c.Locals("start", time.Now())
		return c.Next()
	})

	s.app.Use(middleware.MetricsMiddleware(s.logger))
}

func (s *AdminServer) setupRoutes() error {
	s.logger.Info("Setting up admin routes")

	router, err := core.NewRouter(s.config.Routes, s.logger)
	if err != nil {
		return fmt.Errorf("creating router for admin handler: %w", err)
	}

	adminHandler := handlers.NewAdminHandler(s.configWatcher, router, s.logger)
	adminAuthHandler := handlers.NewAdminAuthHandler(s.configWatcher, s.logger)
	adminRateLimitHandler := handlers.NewAdminRateLimitHandler(s.configWatcher, s.logger)
	adminCircuitBreakerHandler := handlers.NewAdminCircuitBreakerHandler(s.configWatcher, s.logger)
	adminCacheHandler := handlers.NewAdminCacheHandler(s.configWatcher, s.logger)
	metricsHandler := handlers.NewMetricsHandler(s.logger)

	// CRITICAL CHANGE: Move static file serving BEFORE API routes
	// Serve static files for Admin UI
	uiPath := os.Getenv("ADMIN_UI_PATH")
	if uiPath == "" {
		uiPath = "./ui/build"
	}

	s.logger.Info("Checking for UI directory", "path", uiPath)

	if _, err := os.Stat(uiPath); err == nil {
		s.logger.Info("Serving Admin UI", "path", uiPath)

		// Serve static files
		s.app.Static("/static", filepath.Join(uiPath, "static"))

		// API Routes
		s.app.Get("/health", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "ok",
				"time":   time.Now().Format(time.RFC3339),
			})
		})

		s.app.Get("/metrics", metricsHandler.GetMetrics)

		admin := s.app.Group("/admin")

		admin.Get("/routes", adminHandler.GetRoutes)
		admin.Get("/routes/:name", adminHandler.GetRoute)
		admin.Post("/routes", adminHandler.CreateRoute)
		admin.Put("/routes/:name", adminHandler.UpdateRoute)
		admin.Delete("/routes/:name", adminHandler.DeleteRoute)

		admin.Get("/config", adminHandler.GetConfig)
		admin.Put("/config", adminHandler.UpdateConfig)
		admin.Get("/config/backups", adminHandler.GetConfigBackups)
		admin.Post("/config/backups/:id/restore", adminHandler.RestoreConfigBackup)

		admin.Get("/auth/api-keys", adminAuthHandler.GetAPIKeys)
		admin.Post("/auth/api-keys", adminAuthHandler.CreateAPIKey)
		admin.Delete("/auth/api-keys/:route/:key", adminAuthHandler.DeleteAPIKey)
		admin.Get("/auth/jwt", adminAuthHandler.GetJWTConfigs)
		admin.Put("/auth/jwt/:route", adminAuthHandler.UpdateJWTConfig)

		admin.Get("/rate-limits", adminRateLimitHandler.GetRateLimits)
		admin.Get("/rate-limits/:route", adminRateLimitHandler.GetRateLimit)
		admin.Post("/rate-limits", adminRateLimitHandler.CreateRateLimit)
		admin.Put("/rate-limits/:route", adminRateLimitHandler.UpdateRateLimit)
		admin.Delete("/rate-limits/:route", adminRateLimitHandler.DeleteRateLimit)

		admin.Get("/circuit-breakers", adminCircuitBreakerHandler.GetCircuitBreakers)
		admin.Get("/circuit-breakers/:route", adminCircuitBreakerHandler.GetCircuitBreaker)
		admin.Post("/circuit-breakers", adminCircuitBreakerHandler.CreateCircuitBreaker)
		admin.Put("/circuit-breakers/:route", adminCircuitBreakerHandler.UpdateCircuitBreaker)
		admin.Delete("/circuit-breakers/:route", adminCircuitBreakerHandler.DeleteCircuitBreaker)
		admin.Post("/circuit-breakers/:route/reset", adminCircuitBreakerHandler.ResetCircuitBreaker)

		admin.Get("/cache", adminCacheHandler.GetCacheConfigs)
		admin.Get("/cache/:route", adminCacheHandler.GetCacheConfig)
		admin.Post("/cache", adminCacheHandler.CreateCacheConfig)
		admin.Put("/cache/:route", adminCacheHandler.UpdateCacheConfig)
		admin.Delete("/cache/:route", adminCacheHandler.DeleteCacheConfig)
		admin.Post("/cache/:route/clear", adminCacheHandler.ClearCache)

		admin.Get("/version", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"version": "v1.0.0",
				"build":   "20240301-1",
			})
		})

		// Set the root path AFTER registering API routes but BEFORE catch-all
		s.app.Get("/", func(c *fiber.Ctx) error {
			s.logger.Info("Serving index.html for root path")
			return c.SendFile(filepath.Join(uiPath, "index.html"))
		})

		// Catch-all route must be registered last
		s.app.Get("/*", func(c *fiber.Ctx) error {
			path := c.Path()
			s.logger.Info("Catch-all route", "path", path)

			// Skip API routes
			if strings.HasPrefix(path, "/admin/") ||
				path == "/health" ||
				path == "/metrics" ||
				strings.HasPrefix(path, "/static/") {
				return c.Next()
			}

			return c.SendFile(filepath.Join(uiPath, "index.html"))
		})
	} else {
		s.logger.Warn("Admin UI not found, serving API only", "path", uiPath, "error", err.Error())

		// If UI not found, still serve API endpoints
		s.app.Get("/health", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "ok",
				"time":   time.Now().Format(time.RFC3339),
			})
		})

		s.app.Get("/metrics", metricsHandler.GetMetrics)

		admin := s.app.Group("/admin")

		admin.Get("/routes", adminHandler.GetRoutes)
		admin.Get("/routes/:name", adminHandler.GetRoute)
		admin.Post("/routes", adminHandler.CreateRoute)
		admin.Put("/routes/:name", adminHandler.UpdateRoute)
		admin.Delete("/routes/:name", adminHandler.DeleteRoute)

		admin.Get("/config", adminHandler.GetConfig)
		admin.Put("/config", adminHandler.UpdateConfig)
		admin.Get("/config/backups", adminHandler.GetConfigBackups)
		admin.Post("/config/backups/:id/restore", adminHandler.RestoreConfigBackup)

		admin.Get("/auth/api-keys", adminAuthHandler.GetAPIKeys)
		admin.Post("/auth/api-keys", adminAuthHandler.CreateAPIKey)
		admin.Delete("/auth/api-keys/:route/:key", adminAuthHandler.DeleteAPIKey)
		admin.Get("/auth/jwt", adminAuthHandler.GetJWTConfigs)
		admin.Put("/auth/jwt/:route", adminAuthHandler.UpdateJWTConfig)

		admin.Get("/rate-limits", adminRateLimitHandler.GetRateLimits)
		admin.Get("/rate-limits/:route", adminRateLimitHandler.GetRateLimit)
		admin.Post("/rate-limits", adminRateLimitHandler.CreateRateLimit)
		admin.Put("/rate-limits/:route", adminRateLimitHandler.UpdateRateLimit)
		admin.Delete("/rate-limits/:route", adminRateLimitHandler.DeleteRateLimit)

		admin.Get("/circuit-breakers", adminCircuitBreakerHandler.GetCircuitBreakers)
		admin.Get("/circuit-breakers/:route", adminCircuitBreakerHandler.GetCircuitBreaker)
		admin.Post("/circuit-breakers", adminCircuitBreakerHandler.CreateCircuitBreaker)
		admin.Put("/circuit-breakers/:route", adminCircuitBreakerHandler.UpdateCircuitBreaker)
		admin.Delete("/circuit-breakers/:route", adminCircuitBreakerHandler.DeleteCircuitBreaker)
		admin.Post("/circuit-breakers/:route/reset", adminCircuitBreakerHandler.ResetCircuitBreaker)

		admin.Get("/cache", adminCacheHandler.GetCacheConfigs)
		admin.Get("/cache/:route", adminCacheHandler.GetCacheConfig)
		admin.Post("/cache", adminCacheHandler.CreateCacheConfig)
		admin.Put("/cache/:route", adminCacheHandler.UpdateCacheConfig)
		admin.Delete("/cache/:route", adminCacheHandler.DeleteCacheConfig)
		admin.Post("/cache/:route/clear", adminCacheHandler.ClearCache)

		admin.Get("/version", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"version": "v1.0.0",
				"build":   "20240301-1",
			})
		})
	}

	return nil
}

func (s *AdminServer) Start() error {
	s.mu.RLock()
	port := s.config.Server.AdminPort
	s.mu.RUnlock()

	addr := fmt.Sprintf(":%d", port)

	s.logger.Info("Starting admin server", "port", port)
	return s.app.Listen(addr)
}

func (s *AdminServer) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

func (s *AdminServer) UpdateConfig(cfg *config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Updating admin server configuration")

	s.config = cfg

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: handlers.CustomErrorHandler,
	})

	s.app = app

	s.setupMiddleware()
	if err := s.setupRoutes(); err != nil {
		return fmt.Errorf("setting up admin routes: %w", err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)

		port := cfg.Server.AdminPort
		addr := fmt.Sprintf(":%d", port)

		s.logger.Info("Restarting admin server", "port", port)
		if err := s.app.Listen(addr); err != nil {
			s.logger.Error("Failed to restart admin server", "error", err)
		}
	}()

	return nil
}
