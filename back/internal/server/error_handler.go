// Package server provides HTTP server implementation.
package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// errorHandler is a middleware that handles errors and logs them with stack traces
func (s *Server) errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last().Err
			
			// Default status code
			statusCode := http.StatusInternalServerError
			
			// Try to get status code from gin.Error
			if c.Writer.Status() != http.StatusOK {
				statusCode = c.Writer.Status()
			}
			
			// Create error response
			response := ErrorResponse{
				Status:  statusCode,
				Message: "An error occurred while processing your request",
				Error:   err.Error(),
			}
			
			// Log the error with stack trace
			log := c.MustGet("logger").(zerolog.Logger)
			log.Error().Stack().Err(err).Int("status", statusCode).Msg("Request error")
			
			// Send JSON response
			c.JSON(statusCode, response)
		}
	}
}

// DemoError demonstrates error handling with stack traces
func (s *Server) demoError(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)
	
	// Create an error with stack trace
	err := errors.New("this is a demo error")
	
	// Log the error with stack trace
	log.Error().Stack().Err(err).Msg("Demo error occurred")
	
	// Return error response
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Status:  http.StatusInternalServerError,
		Message: "Demo error",
		Error:   err.Error(),
	})
}
