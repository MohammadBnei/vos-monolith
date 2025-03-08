package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"voconsteroid/internal/config"
)

type Server struct {
	cfg    *config.Config
	log    zerolog.Logger
	router *gin.Engine

	srv *http.Server

	serverErrors chan error
	shutdown     chan struct{}
}

func NewServer(cfg *config.Config, log zerolog.Logger) *Server {

	return &Server{
		cfg:    cfg,
		log:    log,
		router: gin.Default(),

		serverErrors: make(chan error),
		shutdown:     make(chan struct{}),
	}
}

func (s *Server) Run() error {
	s.setupRoutes()

	s.srv = &http.Server{
		Addr:    ":" + s.cfg.HTTPPort,
		Handler: s.router,
	}

	// Start the server
	go func() {
		s.log.Info().
			Str("port", s.cfg.HTTPPort).
			Msg("Starting server")
		s.serverErrors <- s.srv.ListenAndServe()
	}()

	// Channel to listen for interrupt or terminate signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	return nil
}

func (s *Server) Shutdown() error {
	// Notify shutdown
	go func() {
		s.shutdown <- struct{}{}
	}()

	// Blocking main and waiting for shutdown
	select {
	case err := <-s.serverErrors:
		return err

	case <-s.shutdown:
		s.log.Info().Msg("Starting graceful shutdown")

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		defer s.log.Info().Msg("Server stopped")
		return s.srv.Shutdown(ctx)
	}
}

func (s *Server) setupRoutes() {
	// Add logger middleware
	s.router.Use(s.loggerMiddleware())

	// Setup API routes
	s.router.GET("/health", s.healthCheck)
}

func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add logger to context
		c.Set("logger", s.log)

		// Log request start
		s.log.Debug().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Msg("Request started")

		c.Next()

		// Log request completion
		s.log.Debug().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Msg("Request completed")
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	// Get logger from context
	log := c.MustGet("logger").(zerolog.Logger)

	log.Debug().Msg("Health check request")
	c.JSON(200, gin.H{
		"status": "ok",
		"app":    s.cfg.AppName,
	})
}
