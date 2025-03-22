package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/bimapangestu28/horizon/internal/cache"
	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/handlers"
	"github.com/bimapangestu28/horizon/internal/httphandlers"
	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/middleware"
	"github.com/bimapangestu28/horizon/internal/proxy"
	"github.com/bimapangestu28/horizon/internal/resilience/circuitbreaker"
	"github.com/bimapangestu28/horizon/internal/security/auth"
	"github.com/bimapangestu28/horizon/internal/security/ipfilter"
	"github.com/bimapangestu28/horizon/internal/security/ratelimit"
	"github.com/bimapangestu28/horizon/internal/transform"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type Server struct {
	app          *fiber.App
	adminApp     *fiber.App
	config       *config.Config
	router       interfaces.Router
	logger       logging.Logger
	proxyHandler *proxy.Handler
	mu           sync.RWMutex
}

func NewServer(cfg *config.Config, logger logging.Logger) (*Server, error) {

	router, err := NewRouter(cfg.Routes, logger)
	if err != nil {
		return nil, fmt.Errorf("creating router: %w", err)
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:             cfg.Server.ReadTimeout,
		WriteTimeout:            cfg.Server.WriteTimeout,
		IdleTimeout:             cfg.Server.IdleTimeout,
		ErrorHandler:            handlers.CustomErrorHandler,
		EnableTrustedProxyCheck: true,
		ServerHeader:            "Horizon API Gateway",
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middleware.NewLogger(logger))

	adminApp := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: handlers.CustomErrorHandler,
	})

	adminApp.Use(recover.New())
	adminApp.Use(requestid.New())
	adminApp.Use(middleware.NewLogger(logger))

	proxyHandler := proxy.New(router, logger)

	server := &Server{
		app:          app,
		adminApp:     adminApp,
		config:       cfg,
		router:       router,
		logger:       logger,
		proxyHandler: proxyHandler,
	}

	server.registerRoutes()
	server.registerAdminRoutes()

	return server, nil
}

func (s *Server) Start() error {

	go func() {
		addr := fmt.Sprintf(":%d", s.config.Server.AdminPort)
		if err := s.adminApp.Listen(addr); err != nil {
			s.logger.Error("Admin server failed", "error", err)
		}
	}()

	addr := fmt.Sprintf(":%d", s.config.Server.Port)

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

func (s *Server) Shutdown() error {

	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.adminApp.ShutdownWithContext(ctx); err != nil {
		s.logger.Error("Admin server shutdown failed", "error", err)
	}

	return s.app.ShutdownWithContext(ctx)
}

func (s *Server) registerRoutes() {

	s.app.Get("/health", httphandlers.HealthCheck)

	for _, route := range s.config.Routes {

		handlers := make([]fiber.Handler, 0)

		if route.IPFilter != nil && route.IPFilter.Enabled {
			ipFilter, err := ipfilter.NewIPFilter(route.IPFilter)
			if err != nil {
				s.logger.Error("Failed to create IP filter", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.IPFilterMiddleware(ipFilter, s.logger))
		}

		if route.Auth != nil && route.Auth.Enabled {
			authenticator, err := auth.NewAuthenticator(route.Auth)
			if err != nil {
				s.logger.Error("Failed to create authenticator", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.AuthMiddleware(authenticator, s.logger))
		}

		if route.RateLimiting != nil && route.RateLimiting.Enabled {
			rateLimiter, err := ratelimit.NewRateLimiter(route.RateLimiting)
			if err != nil {
				s.logger.Error("Failed to create rate limiter", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.RateLimitMiddleware(rateLimiter, s.logger))
		}

		if route.CircuitBreaker != nil && route.CircuitBreaker.Enabled {
			breaker, err := circuitbreaker.NewCircuitBreaker(route.CircuitBreaker)
			if err != nil {
				s.logger.Error("Failed to create circuit breaker", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.CircuitBreakerMiddleware(breaker, s.logger))
		}

		if route.Caching != nil && route.Caching.Enabled {
			cacheStore, err := cache.NewCache(route.Caching)
			if err != nil {
				s.logger.Error("Failed to create cache", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.CacheMiddleware(cacheStore, s.logger))
		}

		if route.Transform != nil && route.Transform.Enabled {
			reqTransformer, err := transform.NewRequestTransformer(route.Transform)
			if err != nil {
				s.logger.Error("Failed to create request transformer", "route", route.Name, "error", err)
				continue
			}

			respTransformer, err := transform.NewResponseTransformer(route.Transform)
			if err != nil {
				s.logger.Error("Failed to create response transformer", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.TransformMiddleware(reqTransformer, respTransformer, s.logger))
		}

		handlers = append(handlers, s.proxyHandler.HandleRequest)

		routePath := route.ListenPath
		s.app.All(routePath, handlers...)

		s.logger.Info("Registered route with middleware",
			"route", route.Name,
			"path", routePath,
			"middleware_count", len(handlers)-1)
	}

	s.app.All("/*", s.proxyHandler.HandleRequest)
}

func (s *Server) registerAdminRoutes() {

	s.adminApp.Get("/health", httphandlers.HealthCheck)

	admin := s.adminApp.Group("/admin")

	admin.Get("/routes", handlers.GetRoutes)

	admin.Get("/config", handlers.GetConfig)
}

func (s *Server) UpdateConfig(cfg *config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Starting configuration update")

	s.config = cfg

	router, err := NewRouter(cfg.Routes, s.logger)
	if err != nil {
		return fmt.Errorf("creating router: %w", err)
	}

	s.router = router
	s.proxyHandler = proxy.New(router, s.logger)

	app := fiber.New(fiber.Config{
		ReadTimeout:             cfg.Server.ReadTimeout,
		WriteTimeout:            cfg.Server.WriteTimeout,
		IdleTimeout:             cfg.Server.IdleTimeout,
		ErrorHandler:            handlers.CustomErrorHandler,
		EnableTrustedProxyCheck: true,
		ServerHeader:            "Horizon API Gateway",
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middleware.NewLogger(s.logger))

	app.Get("/health", httphandlers.HealthCheck)

	for _, route := range cfg.Routes {

		handlers := make([]fiber.Handler, 0)

		if route.IPFilter != nil && route.IPFilter.Enabled {
			ipFilter, err := ipfilter.NewIPFilter(route.IPFilter)
			if err != nil {
				s.logger.Error("Failed to create IP filter", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.IPFilterMiddleware(ipFilter, s.logger))
		}

		if route.Auth != nil && route.Auth.Enabled {
			authenticator, err := auth.NewAuthenticator(route.Auth)
			if err != nil {
				s.logger.Error("Failed to create authenticator", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.AuthMiddleware(authenticator, s.logger))
		}

		if route.RateLimiting != nil && route.RateLimiting.Enabled {
			rateLimiter, err := ratelimit.NewRateLimiter(route.RateLimiting)
			if err != nil {
				s.logger.Error("Failed to create rate limiter", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.RateLimitMiddleware(rateLimiter, s.logger))
		}

		if route.CircuitBreaker != nil && route.CircuitBreaker.Enabled {
			breaker, err := circuitbreaker.NewCircuitBreaker(route.CircuitBreaker)
			if err != nil {
				s.logger.Error("Failed to create circuit breaker", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.CircuitBreakerMiddleware(breaker, s.logger))
		}

		if route.Caching != nil && route.Caching.Enabled {
			cacheStore, err := cache.NewCache(route.Caching)
			if err != nil {
				s.logger.Error("Failed to create cache", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.CacheMiddleware(cacheStore, s.logger))
		}

		if route.Transform != nil && route.Transform.Enabled {
			reqTransformer, err := transform.NewRequestTransformer(route.Transform)
			if err != nil {
				s.logger.Error("Failed to create request transformer", "route", route.Name, "error", err)
				continue
			}

			respTransformer, err := transform.NewResponseTransformer(route.Transform)
			if err != nil {
				s.logger.Error("Failed to create response transformer", "route", route.Name, "error", err)
				continue
			}

			handlers = append(handlers, middleware.TransformMiddleware(reqTransformer, respTransformer, s.logger))
		}

		handlers = append(handlers, s.proxyHandler.HandleRequest)

		routePath := route.ListenPath
		app.All(routePath, handlers...)

		s.logger.Info("Configured route with middleware during hot reload",
			"route", route.Name,
			"path", routePath,
			"middleware_count", len(handlers)-1)
	}

	app.All("/*", s.proxyHandler.HandleRequest)

	adminApp := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: handlers.CustomErrorHandler,
	})

	adminApp.Use(recover.New())
	adminApp.Use(requestid.New())
	adminApp.Use(middleware.NewLogger(s.logger))

	adminApp.Get("/health", httphandlers.HealthCheck)

	admin := adminApp.Group("/admin")

	admin.Get("/routes", handlers.GetRoutes)

	admin.Get("/config", handlers.GetConfig)

	oldApp := s.app
	oldAdminApp := s.adminApp

	s.app = app
	s.adminApp = adminApp

	go func() {

		time.Sleep(100 * time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := oldApp.ShutdownWithContext(ctx); err != nil {
			s.logger.Error("Error shutting down old app during hot reload", "error", err)
		}

		if err := oldAdminApp.ShutdownWithContext(ctx); err != nil {
			s.logger.Error("Error shutting down old admin app during hot reload", "error", err)
		}

		addr := fmt.Sprintf(":%d", s.config.Server.Port)
		if s.config.Server.TLS != nil && s.config.Server.TLS.Enabled {
			s.logger.Info("Restarting server with TLS after config update")
			if err := s.app.ListenTLS(addr, s.config.Server.TLS.CertFile, s.config.Server.TLS.KeyFile); err != nil {
				s.logger.Error("Failed to restart main server after config update", "error", err)
			}
		} else {
			s.logger.Info("Restarting server after config update")
			if err := s.app.Listen(addr); err != nil {
				s.logger.Error("Failed to restart main server after config update", "error", err)
			}
		}
	}()

	go func() {

		time.Sleep(500 * time.Millisecond)

		adminAddr := fmt.Sprintf(":%d", s.config.Server.AdminPort)
		s.logger.Info("Restarting admin server after config update")
		if err := s.adminApp.Listen(adminAddr); err != nil {
			s.logger.Error("Failed to restart admin server after config update", "error", err)
		}
	}()

	s.logger.Info("Server configuration updated, restarting servers")
	return nil
}
