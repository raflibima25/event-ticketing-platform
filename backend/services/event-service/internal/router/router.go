package router

import (
	"github.com/gin-gonic/gin"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/controller"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/middleware"
)

// SetupRouter configures all routes
func SetupRouter(eventController *controller.EventController, jwtSecret string) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "event-service",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public event routes
		events := v1.Group("/events")
		{
			events.GET("", eventController.ListEvents)                      // List events with filters
			events.GET("/slug/:slug", eventController.GetEventBySlug)       // Get event by slug (must be before /:id)
			events.GET("/:id", eventController.GetEvent)                    // Get event by ID
			events.GET("/:id/ticket-tiers", eventController.GetEventTicketTiers) // Get ticket tiers for event
		}

		// Public ticket tier routes
		ticketTiers := v1.Group("/ticket-tiers")
		{
			ticketTiers.GET("/:id", eventController.GetTicketTier) // Get ticket tier by ID
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			// Organizer-only event routes
			organizerEvents := protected.Group("/events")
			organizerEvents.Use(middleware.OrganizerOnly())
			{
				organizerEvents.POST("", eventController.CreateEvent)       // Create event
				organizerEvents.PUT("/:id", eventController.UpdateEvent)    // Update event
				organizerEvents.DELETE("/:id", eventController.DeleteEvent) // Delete event
			}

			// Organizer dashboard
			organizer := protected.Group("/organizer")
			organizer.Use(middleware.OrganizerOnly())
			{
				organizer.GET("/events", eventController.GetOrganizerEvents) // Get organizer's events
			}

			// Organizer-only ticket tier routes
			organizerTicketTiers := protected.Group("/ticket-tiers")
			organizerTicketTiers.Use(middleware.OrganizerOnly())
			{
				organizerTicketTiers.POST("", eventController.CreateTicketTier)       // Create ticket tier
				organizerTicketTiers.PUT("/:id", eventController.UpdateTicketTier)    // Update ticket tier
				organizerTicketTiers.DELETE("/:id", eventController.DeleteTicketTier) // Delete ticket tier
			}
		}
	}

	return r
}
