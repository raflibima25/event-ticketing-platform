package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	sharedresponse "github.com/raflibima25/event-ticketing-platform/backend/pkg/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/message"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/service"
)

// PaymentController handles HTTP requests for payments
type PaymentController struct {
	paymentService service.PaymentService
}

// NewPaymentController creates new payment controller instance
func NewPaymentController(paymentService service.PaymentService) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
	}
}

// CreateInvoice handles POST /invoices - Create payment invoice
func (c *PaymentController) CreateInvoice(ctx *gin.Context) {
	var req request.CreateInvoiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Create invoice
	invoice, err := c.paymentService.CreateInvoice(ctx.Request.Context(), &req)
	if err != nil {
		log.Printf("[ERROR] CreateInvoice failed for order %s: %v", req.OrderID, err)

		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrPaymentAlreadyPaid) {
			statusCode = http.StatusConflict
			errorMessage = message.ErrPaymentAlreadyPaid
		} else if errors.Is(err, service.ErrXenditAPIError) {
			statusCode = http.StatusBadGateway
			errorMessage = message.ErrXenditAPIError
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, sharedresponse.Success(message.MsgInvoiceCreated, invoice))
}

// GetInvoice handles GET /invoices/:orderId - Get invoice by order ID
func (c *PaymentController) GetInvoice(ctx *gin.Context) {
	orderID := ctx.Param("orderId")

	invoice, err := c.paymentService.GetInvoice(ctx.Request.Context(), orderID)
	if err != nil {
		log.Printf("[ERROR] GetInvoice failed for order %s: %v", orderID, err)

		statusCode := http.StatusInternalServerError
		errorMessage := message.ErrInternalServer

		if errors.Is(err, service.ErrPaymentNotFound) {
			statusCode = http.StatusNotFound
			errorMessage = message.ErrPaymentNotFound
		}

		ctx.JSON(statusCode, sharedresponse.Error(errorMessage, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, sharedresponse.Success(message.MsgInvoiceRetrieved, invoice))
}
