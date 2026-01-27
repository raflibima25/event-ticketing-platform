package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/raflibima25/event-ticketing-platform/backend/services/gateway-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/gateway-service/middleware"
	"github.com/raflibima25/event-ticketing-platform/backend/services/gateway-service/pkg"
	"net/http"
	"time"
)

// SetupRouter configures all routes for the API Gateway
func SetupRouter(cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS middleware
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Rate limiting middleware (if enabled)
	if cfg.RateLimit.Enabled {
		rateLimiter := middleware.NewRateLimiter(
			cfg.RateLimit.RequestsPerMinute,
			cfg.RateLimit.BurstSize,
		)
		router.Use(rateLimiter.Middleware())
	}

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// ============================================================
		// AUTH SERVICE ROUTES
		// ============================================================
		auth := v1.Group("/auth")
		{
			// Public routes
			auth.POST("/register", pkg.ProxyHandler(cfg.Services.AuthService))
			auth.POST("/login", pkg.ProxyHandler(cfg.Services.AuthService))
			auth.POST("/refresh", pkg.ProxyHandler(cfg.Services.AuthService))
			auth.POST("/forgot-password", pkg.ProxyHandler(cfg.Services.AuthService))
			auth.POST("/reset-password", pkg.ProxyHandler(cfg.Services.AuthService))

			// Protected routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				authProtected.GET("/profile", pkg.ProxyHandler(cfg.Services.AuthService))
				authProtected.POST("/change-password", pkg.ProxyHandler(cfg.Services.AuthService))
			}
		}

		// ============================================================
		// EVENT SERVICE ROUTES
		// ============================================================

		// Public event routes
		events := v1.Group("/events")
		{
			events.GET("", pkg.ProxyHandler(cfg.Services.EventService))                    // List events
			events.GET("/slug/:slug", pkg.ProxyHandler(cfg.Services.EventService))         // Get by slug
			events.GET("/:id", pkg.ProxyHandler(cfg.Services.EventService))                // Get by ID
			events.GET("/:id/ticket-tiers", pkg.ProxyHandler(cfg.Services.EventService))   // Get ticket tiers
		}

		// Protected event routes (organizer only)
		eventsProtected := v1.Group("/events")
		eventsProtected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		eventsProtected.Use(middleware.RoleMiddleware("organizer", "admin"))
		{
			eventsProtected.POST("", pkg.ProxyHandler(cfg.Services.EventService))          // Create event
			eventsProtected.PUT("/:id", pkg.ProxyHandler(cfg.Services.EventService))       // Update event
			eventsProtected.DELETE("/:id", pkg.ProxyHandler(cfg.Services.EventService))    // Delete event
		}

		// Public ticket tier routes
		ticketTiers := v1.Group("/ticket-tiers")
		{
			ticketTiers.GET("/:id", pkg.ProxyHandler(cfg.Services.EventService))           // Get tier by ID
		}

		// Protected ticket tier routes (organizer only)
		ticketTiersProtected := v1.Group("/ticket-tiers")
		ticketTiersProtected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		ticketTiersProtected.Use(middleware.RoleMiddleware("organizer", "admin"))
		{
			ticketTiersProtected.POST("", pkg.ProxyHandler(cfg.Services.EventService))     // Create tier
			ticketTiersProtected.PUT("/:id", pkg.ProxyHandler(cfg.Services.EventService))  // Update tier
			ticketTiersProtected.DELETE("/:id", pkg.ProxyHandler(cfg.Services.EventService)) // Delete tier
		}

		// Organizer dashboard
		organizer := v1.Group("/organizer")
		organizer.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		organizer.Use(middleware.RoleMiddleware("organizer", "admin"))
		{
			organizer.GET("/events", pkg.ProxyHandler(cfg.Services.EventService))          // Get organizer's events
		}

		// ============================================================
		// TICKETING SERVICE ROUTES
		// ============================================================

		// Protected order routes
		orders := v1.Group("/orders")
		orders.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			orders.POST("", pkg.ProxyHandler(cfg.Services.TicketingService))               // Create order (reserve)
			orders.GET("", pkg.ProxyHandler(cfg.Services.TicketingService))                // Get user orders
			orders.GET("/:id", pkg.ProxyHandler(cfg.Services.TicketingService))            // Get order detail
			orders.POST("/:id/cancel", pkg.ProxyHandler(cfg.Services.TicketingService))    // Cancel order
		}

		// Protected ticket routes
		tickets := v1.Group("/tickets")
		tickets.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			tickets.GET("", pkg.ProxyHandler(cfg.Services.TicketingService))               // Get user tickets
			tickets.GET("/:id", pkg.ProxyHandler(cfg.Services.TicketingService))           // Get ticket detail
		}

		// Internal routes (for inter-service communication)
		// These should ideally be on a separate internal network or use API keys
		internal := v1.Group("/internal")
		{
			internal.POST("/orders/:id/confirm", pkg.ProxyHandler(cfg.Services.TicketingService)) // Confirm payment
		}

		// Public ticket validation (for event staff)
		public := v1.Group("/public")
		{
			public.POST("/tickets/validate", pkg.ProxyHandler(cfg.Services.TicketingService)) // Validate ticket
		}

		// ============================================================
		// PAYMENT SERVICE ROUTES
		// ============================================================

		// Protected payment routes
		payments := v1.Group("/payments")
		payments.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			payments.POST("/invoices", pkg.ProxyHandler(cfg.Services.PaymentService))      // Create invoice
			payments.GET("/invoices/:orderId", pkg.ProxyHandler(cfg.Services.PaymentService)) // Get invoice
		}

		// Webhook routes (no auth - signature verified by service)
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/xendit", pkg.ProxyHandler(cfg.Services.PaymentService))        // Xendit webhook
		}
	}

	return router
}
