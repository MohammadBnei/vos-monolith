package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// setupSwagger configures the Swagger documentation routes
func (s *Server) setupSwagger() {
	// Only enable Swagger UI in development mode
	if s.cfg.LogLevel == "debug" || gin.Mode() == gin.DebugMode {
		s.log.Info().Msg("Enabling Swagger UI in development mode")

		// Serve the Swagger UI
		s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		
		// Serve the index.html file that redirects to Swagger UI
		s.router.GET("/", func(c *gin.Context) {
			c.File("./internal/server/swagger_files/index.html")
		})
		
		// Log Swagger UI URL
		s.log.Info().Msg("Swagger UI available at: /swagger/index.html")
	} else {
		s.log.Info().Msg("Swagger UI disabled in production mode")
	}
}
