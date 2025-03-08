package server

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// setupSwagger configures the Swagger documentation routes
func (s *Server) setupSwagger() {
	// Only enable Swagger UI in development mode
	if s.cfg.LogLevel == "debug" || gin.Mode() == gin.DebugMode {
		s.log.Info().Msg("Enabling Swagger UI in development mode")

		// Configure Swagger URL
		url := ginSwagger.URL("/swagger.yaml")
		
		// Serve the Swagger UI with the specified URL
		s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

		// Serve the swagger.yaml file
		s.router.GET("/swagger.yaml", func(c *gin.Context) {
			c.File(filepath.Join("api", "swagger.yaml"))
		})
		
		// Log Swagger UI URL
		s.log.Info().Msg("Swagger UI available at: /swagger/index.html")
	} else {
		s.log.Info().Msg("Swagger UI disabled in production mode")
	}
}
