// Package server provides HTTP server implementation.
package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// HealthResponse represents the health check response structure.
type HealthResponse struct {
	Status string `json:"status"`
	App    string `json:"app"`
}

// healthCheck handles health check requests.
// It returns a 200 OK response with basic application information.
func (s *Server) healthCheck(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)
	log.Debug().Msg("Health check request")
	
	response := HealthResponse{
		Status: "ok",
		App:    s.cfg.AppName,
	}
	
	c.JSON(http.StatusOK, response)
}

// Additional handler methods would be defined here
