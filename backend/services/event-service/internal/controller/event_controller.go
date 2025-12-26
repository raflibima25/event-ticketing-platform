package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/message"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/service"
)

// EventController handles HTTP requests for events
type EventController struct {
	eventService service.EventService
}

// NewEventController creates new event controller instance
func NewEventController(eventService service.EventService) *EventController {
	return &EventController{
		eventService: eventService,
	}
}

// CreateEvent handles POST /events
func (c *EventController) CreateEvent(ctx *gin.Context) {
	var req request.CreateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   message.ErrInvalidRequest,
			"details": err.Error(),
		})
		return
	}

	// Get organizer ID from context (set by auth middleware)
	organizerID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Create event
	event, err := c.eventService.CreateEvent(ctx.Request.Context(), organizerID.(string), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidDateRange) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": message.ErrInvalidDateRange,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": message.MsgEventCreated,
		"data":    event,
	})
}

// GetEvent handles GET /events/:id
func (c *EventController) GetEvent(ctx *gin.Context) {
	id := ctx.Param("id")

	event, err := c.eventService.GetEventByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": message.ErrEventNotFound,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgEventRetrieved,
		"data":    event,
	})
}

// GetEventBySlug handles GET /events/slug/:slug
func (c *EventController) GetEventBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")

	event, err := c.eventService.GetEventBySlug(ctx.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": message.ErrEventNotFound,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgEventRetrieved,
		"data":    event,
	})
}

// ListEvents handles GET /events
func (c *EventController) ListEvents(ctx *gin.Context) {
	var filters request.ListEventsRequest
	if err := ctx.ShouldBindQuery(&filters); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   message.ErrInvalidRequest,
			"details": err.Error(),
		})
		return
	}

	events, err := c.eventService.ListEvents(ctx.Request.Context(), filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgEventsRetrieved,
		"data":    events,
	})
}

// UpdateEvent handles PUT /events/:id
func (c *EventController) UpdateEvent(ctx *gin.Context) {
	id := ctx.Param("id")

	var req request.UpdateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   message.ErrInvalidRequest,
			"details": err.Error(),
		})
		return
	}

	// Get organizer ID from context
	organizerID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Update event
	event, err := c.eventService.UpdateEvent(ctx.Request.Context(), organizerID.(string), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": message.ErrEventNotFound,
			})
			return
		}

		if errors.Is(err, service.ErrUnauthorized) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": message.ErrForbidden,
			})
			return
		}

		if errors.Is(err, service.ErrInvalidDateRange) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": message.ErrInvalidDateRange,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgEventUpdated,
		"data":    event,
	})
}

// DeleteEvent handles DELETE /events/:id
func (c *EventController) DeleteEvent(ctx *gin.Context) {
	id := ctx.Param("id")

	// Get organizer ID from context
	organizerID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Delete event
	err := c.eventService.DeleteEvent(ctx.Request.Context(), organizerID.(string), id)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": message.ErrEventNotFound,
			})
			return
		}

		if errors.Is(err, service.ErrUnauthorized) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": message.ErrForbidden,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgEventDeleted,
	})
}

// GetOrganizerEvents handles GET /organizer/events
func (c *EventController) GetOrganizerEvents(ctx *gin.Context) {
	// Get organizer ID from context
	organizerID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	events, err := c.eventService.GetOrganizerEvents(ctx.Request.Context(), organizerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgEventsRetrieved,
		"data":    events,
	})
}

// CreateTicketTier handles POST /ticket-tiers
func (c *EventController) CreateTicketTier(ctx *gin.Context) {
	var req request.CreateTicketTierRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   message.ErrInvalidRequest,
			"details": err.Error(),
		})
		return
	}

	// Get organizer ID from context
	organizerID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Create ticket tier
	tier, err := c.eventService.CreateTicketTier(ctx.Request.Context(), organizerID.(string), &req)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": message.ErrEventNotFound,
			})
			return
		}

		if errors.Is(err, service.ErrUnauthorized) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": message.ErrForbidden,
			})
			return
		}

		// Check for validation errors
		if errors.Is(err, request.ErrInvalidEarlyBirdSettings) ||
			errors.Is(err, request.ErrInvalidEarlyBirdPrice) ||
			errors.Is(err, request.ErrInvalidEarlyBirdEndDate) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": message.MsgTicketTierCreated,
		"data":    tier,
	})
}

// GetTicketTier handles GET /ticket-tiers/:id
func (c *EventController) GetTicketTier(ctx *gin.Context) {
	id := ctx.Param("id")

	tier, err := c.eventService.GetTicketTierByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTicketTierNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": message.ErrTicketTierNotFound,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": tier,
	})
}

// GetEventTicketTiers handles GET /events/:id/ticket-tiers
func (c *EventController) GetEventTicketTiers(ctx *gin.Context) {
	eventID := ctx.Param("id")

	tiers, err := c.eventService.GetTicketTiersByEventID(ctx.Request.Context(), eventID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": tiers,
	})
}

// UpdateTicketTier handles PUT /ticket-tiers/:id
func (c *EventController) UpdateTicketTier(ctx *gin.Context) {
	id := ctx.Param("id")

	var req request.UpdateTicketTierRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   message.ErrInvalidRequest,
			"details": err.Error(),
		})
		return
	}

	// Get organizer ID from context
	organizerID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Update ticket tier
	tier, err := c.eventService.UpdateTicketTier(ctx.Request.Context(), organizerID.(string), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrTicketTierNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": message.ErrTicketTierNotFound,
			})
			return
		}

		if errors.Is(err, service.ErrUnauthorized) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": message.ErrForbidden,
			})
			return
		}

		if errors.Is(err, service.ErrQuotaBelowSoldCount) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": message.ErrQuotaBelowSoldCount,
			})
			return
		}

		// Check for validation errors
		if errors.Is(err, request.ErrInvalidEarlyBirdSettings) ||
			errors.Is(err, request.ErrInvalidEarlyBirdPrice) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgTicketTierUpdated,
		"data":    tier,
	})
}

// DeleteTicketTier handles DELETE /ticket-tiers/:id
func (c *EventController) DeleteTicketTier(ctx *gin.Context) {
	id := ctx.Param("id")

	// Get organizer ID from context
	organizerID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Delete ticket tier
	err := c.eventService.DeleteTicketTier(ctx.Request.Context(), organizerID.(string), id)
	if err != nil {
		if errors.Is(err, service.ErrTicketTierNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": message.ErrTicketTierNotFound,
			})
			return
		}

		if errors.Is(err, service.ErrUnauthorized) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": message.ErrForbidden,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgTicketTierDeleted,
	})
}
