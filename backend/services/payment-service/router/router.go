package router

import (
	"github.com/gin-gonic/gin"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/controller"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/middleware"
)

// SetupRouter configures all routes for the payment service
func SetupRouter(
	cfg *config.Config,
	paymentController *controller.PaymentController,
	webhookController *controller.WebhookController,
) *gin.Engine {
	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "payment-service",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Payment routes (protected with JWT)
		payments := v1.Group("/payments")
		payments.Use(middleware.JWTAuth(&cfg.JWT))
		{
			payments.POST("/invoices", paymentController.CreateInvoice)
			payments.GET("/invoices/:orderId", paymentController.GetInvoice)
		}

		// Webhook routes (public - no JWT, uses signature verification)
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/xendit", webhookController.HandleXenditWebhook)
		}
	}

	return router
}
