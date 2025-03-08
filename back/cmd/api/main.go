// Package main is the entry point for the API server application.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"voconsteroid/internal/config"
	"voconsteroid/internal/server"
	"voconsteroid/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

// run encapsulates the startup and shutdown logic, returning any error encountered.
func run() error {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	log.Info().Str("app", cfg.AppName).Msg("Starting application")

	// Create and start server
	srv := server.NewServer(cfg, log)
	if err := srv.Run(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")

	// Graceful shutdown
	log.Info().Msg("Shutting down server...")
	if err := srv.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	log.Info().Msg("Application stopped")
	return nil
}
