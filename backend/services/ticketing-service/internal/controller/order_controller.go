package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/message"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/service"
)

// OrderController handles HTTP requests for orders
type OrderController struct {
	reservationService  service.ReservationService
	orderService        service.OrderService
	confirmationService service.ConfirmationService
}

// NewOrderController creates new order controller instance
func NewOrderController(
	reservationService service.ReservationService,
	orderService service.OrderService,
	confirmationService service.ConfirmationService,
) *OrderController {
	return &OrderController{
		reservationService:  reservationService,
		orderService:        orderService,
		confirmationService: confirmationService,
	}
}

// CreateOrder handles POST /orders - Create order (reserve tickets)
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req request.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   message.ErrInvalidRequest,
			"details": err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Create reservation
	order, err := c.reservationService.CreateReservation(ctx.Request.Context(), userID.(string), &req)
	if err != nil {
		// Log the actual error for debugging
		log.Printf("[ERROR] CreateOrder failed for user %s: %v", userID.(string), err)

		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		// Handle specific errors
		if errors.Is(err, service.ErrInsufficientQuota) {
			statusCode = http.StatusConflict
			errorMessage = message.ErrInsufficientQuota
		} else if errors.Is(err, service.ErrInvalidQuantity) {
			statusCode = http.StatusBadRequest
			errorMessage = message.ErrInvalidQuantity
		} else if errors.Is(err, service.ErrMaxPerOrderExceeded) {
			statusCode = http.StatusBadRequest
			errorMessage = message.ErrMaxPerOrderExceeded
		} else if errors.Is(err, service.ErrLockAcquisitionFailed) {
			statusCode = http.StatusConflict
			errorMessage = message.ErrLockAcquisitionFailed
		} else if errors.Is(err, service.ErrTicketTierNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrTicketTierNotFound
		}

		ctx.JSON(statusCode, gin.H{
			"error": errorMessage,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": message.MsgOrderCreated,
		"data":    order,
	})
}

// GetOrder handles GET /orders/:id - Get order by ID
func (c *OrderController) GetOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Get order
	order, err := c.orderService.GetOrderByID(ctx.Request.Context(), userID.(string), orderID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrOrderNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrOrderNotFound
		} else if errors.Is(err, service.ErrUnauthorized) {
			statusCode = http.StatusForbidden
			errorMessage = message.ErrForbidden
		}

		ctx.JSON(statusCode, gin.H{
			"error": errorMessage,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgOrderRetrieved,
		"data":    order,
	})
}

// GetUserOrders handles GET /orders - Get user's orders
func (c *OrderController) GetUserOrders(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Get orders
	orders, total, err := c.orderService.GetUserOrders(ctx.Request.Context(), userID.(string), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": message.ErrInternalServer,
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgOrdersRetrieved,
		"data": gin.H{
			"orders": orders,
			"pagination": gin.H{
				"current_page": page,
				"per_page":     limit,
				"total":        total,
				"total_pages":  totalPages,
			},
		},
	})
}

// CancelOrder handles POST /orders/:id/cancel - Cancel order
func (c *OrderController) CancelOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": message.ErrUnauthorized,
		})
		return
	}

	// Cancel order
	if err := c.orderService.CancelOrder(ctx.Request.Context(), userID.(string), orderID); err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrOrderNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrOrderNotFound
		} else if errors.Is(err, service.ErrUnauthorized) {
			statusCode = http.StatusForbidden
			errorMessage = message.ErrForbidden
		} else if errors.Is(err, service.ErrCannotCancelOrder) {
			statusCode = http.StatusBadRequest
			errorMessage = message.ErrCannotCancelOrder
		}

		ctx.JSON(statusCode, gin.H{
			"error": errorMessage,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgOrderCancelled,
	})
}

// ConfirmPayment handles POST /orders/:id/confirm - Confirm payment (webhook/internal)
func (c *OrderController) ConfirmPayment(ctx *gin.Context) {
	// Get order ID from URL path parameter
	orderID := ctx.Param("id")
	if orderID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Order ID is required",
		})
		return
	}

	var req request.ConfirmOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("[DEBUG] ConfirmPayment - Bind JSON failed. Error: %v", err)
		log.Printf("[DEBUG] Request body: %+v", req)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   message.ErrInvalidRequest,
			"details": err.Error(),
		})
		return
	}

	log.Printf("[DEBUG] ConfirmPayment - Request parsed successfully: %+v", req)

	// Set order ID from URL parameter
	req.OrderID = orderID

	// Confirm payment and generate tickets
	if err := c.confirmationService.ConfirmPayment(ctx.Request.Context(), &req); err != nil {
		log.Printf("[ERROR] ConfirmPayment failed for order %s: %v", orderID, err)

		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrOrderNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrOrderNotFound
		} else if errors.Is(err, service.ErrOrderExpired) {
			statusCode = http.StatusBadRequest
			errorMessage = message.ErrOrderExpired
		} else if errors.Is(err, service.ErrOrderNotInReservedStatus) {
			statusCode = http.StatusBadRequest
			errorMessage = "Order is not in reserved status"
		} else if errors.Is(err, service.ErrAmountMismatch) {
			statusCode = http.StatusBadRequest
			errorMessage = err.Error()
		}

		ctx.JSON(statusCode, gin.H{
			"error": errorMessage,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message.MsgOrderConfirmed,
	})
}
