// Package main is the entry point for the API server application.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"voconsteroid/internal/config"
	"voconsteroid/internal/server"
	"voconsteroid/pkg/logger"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Only log this as info, not fatal, since .env is optional
		log.Printf("No .env file found: %v", err)
	}

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
	logConfig := logger.DefaultConfig()
	logConfig.Level = cfg.LogLevel
	
	log := logger.NewWithConfig(logConfig)
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
