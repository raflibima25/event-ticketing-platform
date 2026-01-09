package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	sharedresponse "github.com/raflibima25/event-ticketing-platform/backend/pkg/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/message"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/service"
)

// TicketController handles HTTP requests for tickets
type TicketController struct {
	ticketService service.TicketService
}

// NewTicketController creates new ticket controller instance
func NewTicketController(ticketService service.TicketService) *TicketController {
	return &TicketController{
		ticketService: ticketService,
	}
}

// GetTicket handles GET /tickets/:id - Get ticket by ID
func (c *TicketController) GetTicket(ctx *gin.Context) {
	ticketID := ctx.Param("id")

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, sharedresponse.Error(message.ErrUnauthorized, nil))
		return
	}

	// Get ticket
	ticket, err := c.ticketService.GetTicket(ctx.Request.Context(), userID.(string), ticketID)
	if err != nil {
		log.Printf("[ERROR] GetTicket failed for user %s, ticket %s: %v", userID.(string), ticketID, err)

		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrTicketNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrTicketNotFound
		} else if errors.Is(err, service.ErrUnauthorized) {
			statusCode = http.StatusForbidden
			errorMessage = message.ErrForbidden
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, sharedresponse.Success(message.MsgTicketRetrieved, ticket))
}

// GetUserTickets handles GET /tickets - Get user's tickets
func (c *TicketController) GetUserTickets(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, sharedresponse.Error(message.ErrUnauthorized, nil))
		return
	}

	// Get tickets
	tickets, err := c.ticketService.GetUserTickets(ctx.Request.Context(), userID.(string))
	if err != nil {
		log.Printf("[ERROR] GetUserTickets failed for user %s: %v", userID.(string), err)

		ctx.JSON(http.StatusInternalServerError, sharedresponse.Error(message.ErrInternalServer, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, sharedresponse.Success(message.MsgTicketsRetrieved, tickets))
}

// ValidateTicket handles POST /tickets/validate - Validate ticket at event entrance
func (c *TicketController) ValidateTicket(ctx *gin.Context) {
	var req request.ValidateTicketRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Validate ticket
	ticket, err := c.ticketService.ValidateTicket(ctx.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrTicketNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrTicketNotFound
		} else if errors.Is(err, service.ErrTicketAlreadyUsed) {
			statusCode = http.StatusConflict
			errorMessage = message.ErrTicketAlreadyUsed
		} else if errors.Is(err, service.ErrTicketInvalid) {
			statusCode = http.StatusBadRequest
			errorMessage = message.ErrTicketInvalid
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, sharedresponse.Success(message.MsgTicketValidated, ticket))
}
