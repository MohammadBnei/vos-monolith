package server

import (
	"io/fs"
	"net/http"

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

		// Get the embedded swagger files
		swaggerFS, err := fs.Sub(swaggerContent, "swagger_files")
		if err != nil {
			s.log.Error().Err(err).Msg("Failed to create sub-filesystem for swagger files")
			return
		}

		// Serve the swagger.yaml file from embedded content
		s.router.GET("/swagger.yaml", func(c *gin.Context) {
			yamlContent, err := fs.ReadFile(swaggerFS, "swagger.yaml")
			if err != nil {
				s.log.Error().Err(err).Msg("Failed to read embedded swagger.yaml")
				c.String(http.StatusInternalServerError, "Failed to read swagger.yaml")
				return
			}
			c.Data(http.StatusOK, "application/yaml", yamlContent)
		})
		
		// Serve the index.html file from embedded content
		s.router.GET("/", func(c *gin.Context) {
			htmlContent, err := fs.ReadFile(swaggerFS, "index.html")
			if err != nil {
				s.log.Error().Err(err).Msg("Failed to read embedded index.html")
				c.String(http.StatusInternalServerError, "Failed to read index.html")
				return
			}
			c.Data(http.StatusOK, "text/html", htmlContent)
		})
		
		// Log Swagger UI URL
		s.log.Info().Msg("Swagger UI available at: /swagger/index.html")
	} else {
		s.log.Info().Msg("Swagger UI disabled in production mode")
	}
}
