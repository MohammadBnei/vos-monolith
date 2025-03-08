package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"voconsteroid/internal/config"
	"voconsteroid/internal/server"
	"voconsteroid/pkg/logger"
)

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)

	// Create and start server
	srv := server.NewServer(cfg, log)

	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	if err := srv.Shutdown(); err != nil {
		log.Fatal().Err(err).Msg("Failed to shutdown server")
	}
}
