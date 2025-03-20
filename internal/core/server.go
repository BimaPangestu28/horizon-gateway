package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/handlers"
	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/middleware"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

// Server represents the API Gateway server
type Server struct {
	app      *fiber.App
	adminApp *fiber.App
	config   *config.Config
	router   interfaces.Router
	logger   logging.Logger
	mu       sync.RWMutex
}

// NewServer creates a new API Gateway server instance
func NewServer(cfg *config.Config, logger logging.Logger) (*Server, error) {
	// Create router
	router, err := NewRouter(cfg.Routes, logger)
	if err != nil {
		return nil, fmt.Errorf("creating router: %w", err)
	}

	// Create main server app with corrected config
	app := fiber.New(fiber.Config{
		ReadTimeout:             cfg.Server.ReadTimeout,
		WriteTimeout:            cfg.Server.WriteTimeout,
		IdleTimeout:             cfg.Server.IdleTimeout,
		ErrorHandler:            handlers.CustomErrorHandler,
		EnableTrustedProxyCheck: true,
		// Removed the EnableHTTP2 field as it's not available in the current version
		ServerHeader: "Horizon API Gateway",
	})

	// Add global middlewares
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middleware.NewLogger(logger))

	// Create admin server app
	adminApp := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: handlers.CustomErrorHandler,
	})

	// Add global middlewares for admin app
	adminApp.Use(recover.New())
	adminApp.Use(requestid.New())
	adminApp.Use(middleware.NewLogger(logger))

	// Create server instance
	server := &Server{
		app:      app,
		adminApp: adminApp,
		config:   cfg,
		router:   router,
		logger:   logger,
	}

	// Register routes
	server.registerRoutes()
	server.registerAdminRoutes()

	return server, nil
}

// Start starts the API Gateway server
func (s *Server) Start() error {
	// Start admin server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", s.config.Server.AdminPort)
		if err := s.adminApp.Listen(addr); err != nil {
			s.logger.Error("Admin server failed", "error", err)
		}
	}()

	// Start main server
	addr := fmt.Sprintf(":%d", s.config.Server.Port)

	// Check if TLS is enabled
	if s.config.Server.TLS != nil && s.config.Server.TLS.Enabled {
		s.logger.Info("Starting server with TLS",
			"cert_file", s.config.Server.TLS.CertFile,
			"key_file", s.config.Server.TLS.KeyFile,
		)
		return s.app.ListenTLS(addr, s.config.Server.TLS.CertFile, s.config.Server.TLS.KeyFile)
	}

	s.logger.Info("Starting server without TLS")
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	// Shutdown with timeout
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Shutdown admin server
	if err := s.adminApp.ShutdownWithContext(ctx); err != nil {
		s.logger.Error("Admin server shutdown failed", "error", err)
	}

	// Shutdown main server
	return s.app.ShutdownWithContext(ctx)
}

// registerRoutes sets up the main API Gateway routes
func (s *Server) registerRoutes() {
	// Create proxy handler
	proxyHandler := handlers.NewProxyHandler(s.router, s.logger)

	// Set up health check endpoint
	s.app.Get("/health", handlers.HealthCheck)

	// Set up catch-all route for API Gateway proxying
	s.app.All("/*", proxyHandler.HandleRequest)
}

// registerAdminRoutes sets up the admin API routes
func (s *Server) registerAdminRoutes() {
	// Set up admin API endpoints
	s.adminApp.Get("/health", handlers.HealthCheck)

	// Admin routes group
	admin := s.adminApp.Group("/admin")

	// Routes management
	admin.Get("/routes", handlers.GetRoutes)

	// Configuration management
	admin.Get("/config", handlers.GetConfig)
}

func (s *Server) UpdateConfig(cfg *config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update server config
	s.config = cfg

	// Create new router with updated routes
	router, err := NewRouter(cfg.Routes, s.logger)
	if err != nil {
		return fmt.Errorf("creating router: %w", err)
	}

	// Update router
	s.router = router

	s.logger.Info("Server configuration updated")
	return nil
}
