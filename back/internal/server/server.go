// Package server provides HTTP server implementation following clean architecture principles.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"voconsteroid/internal/config"
	"voconsteroid/internal/domain/word"
)

// Server represents the HTTP server with all its dependencies.
type Server struct {
	cfg    *config.Config
	log    zerolog.Logger
	router *gin.Engine
	srv    *http.Server

	// Services
	wordService word.Service
}

// NewServer creates a new server instance with the provided configuration and logger.
func NewServer(cfg *config.Config, log zerolog.Logger, wordService word.Service) *Server {
	// Set gin mode based on log level
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	return &Server{
		cfg:         cfg,
		log:         log,
		router:      router,
		wordService: wordService,
	}
}

// Run initializes and starts the HTTP server.
func (s *Server) Run() error {
	s.setupMiddleware()
	s.setupRoutes()

	// Check if port is already in use before starting the server
	if err := s.checkPortAvailable(s.cfg.HTTPPort); err != nil {
		return fmt.Errorf("port check failed: %w", err)
	}

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%s", s.cfg.HTTPPort),
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Log server start
	s.log.Info().
		Str("port", s.cfg.HTTPPort).
		Str("app", s.cfg.AppName).
		Msg("Starting server")

	// Start the server in the main goroutine so errors are returned immediately
	err := s.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
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

// setupMiddleware configures all middleware for the server.
func (s *Server) setupMiddleware() {
	// Add recovery middleware
	s.router.Use(gin.Recovery())

	// Add logger middleware
	s.router.Use(s.loggerMiddleware())

	// Add error handler middleware
	s.router.Use(s.errorHandler())
}

// setupRoutes configures all routes for the server.
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)

	// Demo error endpoint
	s.router.GET("/error", s.demoError)

	// Setup Swagger documentation
	s.setupSwagger()
	
	// Serve API documentation index
	s.router.GET("/api/docs", func(c *gin.Context) {
		c.File(filepath.Join("api", "index.html"))
	})

	// Word API routes
	api := s.router.Group("/api/v1")
	{
		words := api.Group("/words")
		{
			words.POST("/search", s.SearchWord)
			words.GET("/recent", s.GetRecentWords)
		}
	}
}

// checkPortAvailable checks if the specified port is available for binding
func (s *Server) checkPortAvailable(port string) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("port %s is not available: %w", port, err)
	}
	ln.Close()
	return nil
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
