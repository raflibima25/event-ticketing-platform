package controller

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	sharedresponse "github.com/raflibima25/event-ticketing-platform/backend/pkg/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/message"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/service"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/utility"
)

// WebhookController handles HTTP requests for webhooks
type WebhookController struct {
	webhookService service.WebhookService
	webhookToken   string
}

// NewWebhookController creates new webhook controller instance
func NewWebhookController(webhookService service.WebhookService, cfg *config.Config) *WebhookController {
	return &WebhookController{
		webhookService: webhookService,
		webhookToken:   cfg.Xendit.WebhookToken,
	}
}

// HandleXenditWebhook handles POST /webhooks/xendit - Xendit webhook callback
func (c *WebhookController) HandleXenditWebhook(ctx *gin.Context) {
	// Step 1: Verify callback token (Xendit uses x-callback-token header)
	callbackToken := ctx.GetHeader("x-callback-token")
	if err := utility.VerifyCallbackToken(callbackToken, c.webhookToken); err != nil {
		log.Printf("[ERROR] Invalid webhook signature/token")
		ctx.JSON(http.StatusUnauthorized, sharedresponse.Error(message.ErrInvalidSignature, err.Error()))
		return
	}

	// Step 2: Read request body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read webhook body: %v", err)
		ctx.JSON(http.StatusBadRequest, sharedresponse.Error(message.ErrInvalidRequest, err.Error()))
		return
	}

	// Step 3: Get webhook ID and event type from body
	// For Xendit, webhook ID is the invoice ID
	// We'll extract it from the parsed JSON in the service layer
	// For now, we'll use a timestamp-based ID if not available
	webhookID := ctx.GetHeader("webhook-id")
	if webhookID == "" {
		// Xendit doesn't send webhook-id header, we'll use invoice ID from payload
		// This will be handled in service layer
		webhookID = ctx.GetHeader("x-request-id")
		if webhookID == "" {
			webhookID = "XENDIT-" + ctx.GetString("request_id")
		}
	}

	// Event type from Xendit is in the body (status field)
	// We'll determine it in service layer based on status
	eventType := "invoice.paid" // Default, will be determined by service

	// Step 4: Process webhook
	if err := c.webhookService.ProcessWebhook(ctx.Request.Context(), webhookID, eventType, body); err != nil {
		// Handle duplicate webhooks (idempotency)
		if errors.Is(err, service.ErrDuplicateWebhook) {
			log.Printf("[INFO] Duplicate webhook: %s", webhookID)
			ctx.JSON(http.StatusOK, sharedresponse.Success("Webhook already processed", nil))
			return
		}

		// Handle payment not found (test webhooks or race conditions)
		if errors.Is(err, repository.ErrPaymentNotFound) || strings.Contains(err.Error(), "payment not found") {
			log.Printf("[WARN] Payment not found for webhook %s - possibly test webhook or race condition", webhookID)
			ctx.JSON(http.StatusOK, sharedresponse.Success("Webhook received but payment not found (possibly test webhook)", nil))
			return
		}

		// Log actual errors but still return 200 to prevent Xendit retries
		// Only critical errors should return 500
		log.Printf("[ERROR] Failed to process webhook %s: %v", webhookID, err)
		ctx.JSON(http.StatusOK, sharedresponse.Success("Webhook received with errors", map[string]string{
			"warning": "Payment processing may have failed - check logs",
		}))
		return
	}

	// Step 5: Return success response
	ctx.JSON(http.StatusOK, sharedresponse.Success(message.MsgWebhookProcessed, nil))
}
