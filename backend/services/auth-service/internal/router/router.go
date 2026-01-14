package router

import (
	"github.com/gin-gonic/gin"

	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/controller"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/middleware"
)

// SetupRouter configures all routes for the service
func SetupRouter(authController *controller.AuthController, jwtSecret string) *gin.Engine {
	router := gin.Default()

	// NOTE: CORS is handled by API Gateway - do not add CORS middleware here
	// Adding CORS here causes duplicate Access-Control-Allow-Origin headers

	// Health check (public)
	router.GET("/health", authController.Health)

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
		}

		// Protected routes (require authentication)
		protected := api.Group("/auth")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.GET("/profile", authController.GetProfile)
		}
	}

	return router
}
