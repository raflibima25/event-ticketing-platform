package router

import (
	"github.com/gin-gonic/gin"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/controller"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/middleware"
)

// SetupRouter configures all routes
func SetupRouter(
	orderController *controller.OrderController,
	ticketController *controller.TicketController,
	jwtSecret string,
) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "ticketing-service",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			// Order endpoints
			orders := protected.Group("/orders")
			{
				orders.POST("", orderController.CreateOrder)           // Create order (reserve tickets)
				orders.GET("", orderController.GetUserOrders)          // Get user's orders
				orders.GET("/:id", orderController.GetOrder)           // Get order detail
				orders.POST("/:id/cancel", orderController.CancelOrder) // Cancel order
			}

			// Ticket endpoints
			tickets := protected.Group("/tickets")
			{
				tickets.GET("", ticketController.GetUserTickets)      // Get user's tickets
				tickets.GET("/:id", ticketController.GetTicket)       // Get ticket detail
			}
		}

		// Internal/Webhook endpoints (should be called by Payment Service)
		// In production, these should be protected by API key or internal network
		internal := v1.Group("/internal")
		{
			internal.POST("/orders/:id/confirm", orderController.ConfirmPayment) // Confirm payment
		}

		// Public endpoints (for event staff to validate tickets)
		// In production, this should be protected by staff authentication
		public := v1.Group("/public")
		{
			public.POST("/tickets/validate", ticketController.ValidateTicket) // Validate ticket at entrance
		}
	}

	return r
}
