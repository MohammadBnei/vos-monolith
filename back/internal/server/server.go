// Package server provides HTTP server implementation following clean architecture principles.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"voconsteroid/internal/config"
)

// Server represents the HTTP server with all its dependencies.
type Server struct {
	cfg    *config.Config
	log    zerolog.Logger
	router *gin.Engine
	srv    *http.Server

	// Channels for managing server lifecycle
	serverErrors chan error
	shutdown     chan struct{}
}

// NewServer creates a new server instance with the provided configuration and logger.
func NewServer(cfg *config.Config, log zerolog.Logger) *Server {
	// Set gin mode based on log level
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	
	return &Server{
		cfg:          cfg,
		log:          log,
		router:       router,
		serverErrors: make(chan error, 1),
		shutdown:     make(chan struct{}, 1),
	}
}

// Run initializes and starts the HTTP server.
func (s *Server) Run() error {
	s.setupMiddleware()
	s.setupRoutes()

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%s", s.cfg.HTTPPort),
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		s.log.Info().
			Str("port", s.cfg.HTTPPort).
			Str("app", s.cfg.AppName).
			Msg("Starting server")
		
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.serverErrors <- fmt.Errorf("server error: %w", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	// Notify shutdown
	go func() {
		s.shutdown <- struct{}{}
	}()

	// Blocking main and waiting for shutdown
	select {
	case err := <-s.serverErrors:
		return fmt.Errorf("server error during shutdown: %w", err)

	case <-s.shutdown:
		s.log.Info().Msg("Starting graceful shutdown")

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		
		s.log.Info().Msg("Server stopped")
		return nil
	}
}

// setupMiddleware configures all middleware for the server.
func (s *Server) setupMiddleware() {
	// Add recovery middleware
	s.router.Use(gin.Recovery())
	
	// Add logger middleware
	s.router.Use(s.loggerMiddleware())
}

// setupRoutes configures all routes for the server.
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)
	
	// API routes would be registered here or in separate handler files
}

// loggerMiddleware creates a middleware that logs HTTP requests.
func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add logger to context
		c.Set("logger", s.log)

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Log request start
		s.log.Debug().
			Str("method", method).
			Str("path", path).
			Msg("Request started")

		c.Next()

		// Log request completion with duration
		s.log.Debug().
			Str("method", method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("duration", time.Since(start)).
			Msg("Request completed")
	}
}

// healthCheck handles health check requests.
func (s *Server) healthCheck(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)
	log.Debug().Msg("Health check request")
	
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"app":    s.cfg.AppName,
	})
}
