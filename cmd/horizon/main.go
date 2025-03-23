package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bimapangestu28/horizon/internal/config"
	"github.com/bimapangestu28/horizon/internal/core"
	"github.com/bimapangestu28/horizon/internal/proxy"
	"github.com/bimapangestu28/horizon/internal/server"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

func main() {
	// Parse command-line arguments
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Create logger
	logger := logging.NewLogger()
	logger.Info("Starting Horizon API Gateway")

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err.Error())
	}

	// Create router
	router, err := core.NewRouter(cfg.Routes, logger)
	if err != nil {
		logger.Fatal("Failed to create router", "error", err.Error())
	}

	// Create config watcher
	configWatcher, err := config.NewConfigWatcher(*configPath, logger)
	if err != nil {
		logger.Fatal("Failed to create config watcher", "error", err.Error())
	}

	// Create proxy handler
	proxyHandler := proxy.New(router, logger)

	// Create server
	srv, err := server.NewServer(cfg, configWatcher, logger, proxyHandler, nil)
	if err != nil {
		logger.Fatal("Failed to create server", "error", err.Error())
	}

	// Create admin server
	adminSrv, err := server.NewAdminServer(cfg, configWatcher, logger)
	if err != nil {
		logger.Fatal("Failed to create admin server", "error", err.Error())
	}

	// Start the config watcher
	if err := configWatcher.Start(); err != nil {
		logger.Fatal("Failed to start config watcher", "error", err.Error())
	}
	defer configWatcher.Stop()

	// Register config change callback
	configWatcher.RegisterCallback(func(newCfg *config.Config) error {
		logger.Info("Configuration changed, updating servers")

		if err := srv.UpdateConfig(newCfg); err != nil {
			logger.Error("Failed to update server config", "error", err.Error())
			return err
		}

		if err := adminSrv.UpdateConfig(newCfg); err != nil {
			logger.Error("Failed to update admin server config", "error", err.Error())
			return err
		}

		return nil
	})

	// Start servers in goroutines
	go func() {
		logger.Info("Starting main server", "port", cfg.Server.Port)
		if err := srv.Start(); err != nil {
			logger.Fatal("Server failed to start", "error", err.Error())
		}
	}()

	go func() {
		logger.Info("Starting admin server", "port", cfg.Server.AdminPort)
		if err := adminSrv.Start(); err != nil {
			logger.Fatal("Admin server failed to start", "error", err.Error())
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers")

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown both servers
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "error", err.Error())
	}

	if err := adminSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Admin server shutdown failed", "error", err.Error())
	}

	logger.Info("Servers stopped, goodbye!")
}
