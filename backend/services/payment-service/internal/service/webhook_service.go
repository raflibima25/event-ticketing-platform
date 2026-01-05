package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/client"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/repository"
)

var (
	ErrDuplicateWebhook = errors.New("webhook already processed")
	ErrWebhookNotFound  = errors.New("webhook event not found")
)

// WebhookService handles webhook event processing
type WebhookService interface {
	ProcessWebhook(ctx context.Context, webhookID string, eventType string, payload []byte) error
}

// webhookService implements WebhookService interface
type webhookService struct {
	webhookRepo      repository.WebhookRepository
	paymentRepo      repository.PaymentRepository
	ticketingClient  *client.TicketingClient
}

// NewWebhookService creates new webhook service instance
func NewWebhookService(
	webhookRepo repository.WebhookRepository,
	paymentRepo repository.PaymentRepository,
	ticketingClient *client.TicketingClient,
) WebhookService {
	return &webhookService{
		webhookRepo:     webhookRepo,
		paymentRepo:     paymentRepo,
		ticketingClient: ticketingClient,
	}
}

// ProcessWebhook processes incoming webhook with idempotency
func (s *webhookService) ProcessWebhook(ctx context.Context, webhookID string, eventType string, payload []byte) error {
	// Step 1: Idempotency check - Save webhook event (will fail if duplicate)
	webhookEvent := &entity.WebhookEvent{
		WebhookID: webhookID,
		EventType: eventType,
		Payload:   string(payload),
		Status:    entity.WebhookStatusPending,
	}

	if err := s.webhookRepo.Create(ctx, webhookEvent); err != nil {
		if errors.Is(err, repository.ErrDuplicateWebhook) {
			log.Printf("[INFO] Duplicate webhook received: %s (already processed)", webhookID)
			return ErrDuplicateWebhook
		}
		return fmt.Errorf("failed to save webhook event: %w", err)
	}

	// Step 2: Parse webhook payload
	var webhookPayload response.XenditWebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		s.webhookRepo.MarkAsFailed(ctx, webhookID)
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Step 3: Process based on event type
	var err error
	switch eventType {
	case entity.EventTypeInvoicePaid:
		err = s.handleInvoicePaid(ctx, &webhookPayload)
	case entity.EventTypeInvoiceExpired:
		err = s.handleInvoiceExpired(ctx, &webhookPayload)
	default:
		log.Printf("[INFO] Unhandled webhook event type: %s", eventType)
		err = nil // Not an error, just ignore
	}

	// Step 4: Mark webhook as processed or failed
	if err != nil {
		log.Printf("[ERROR] Failed to process webhook %s: %v", webhookID, err)
		s.webhookRepo.MarkAsFailed(ctx, webhookID)
		return err
	}

	if err := s.webhookRepo.MarkAsProcessed(ctx, webhookID); err != nil {
		return fmt.Errorf("failed to mark webhook as processed: %w", err)
	}

	log.Printf("[INFO] Successfully processed webhook: %s (type: %s)", webhookID, eventType)
	return nil
}

// handleInvoicePaid handles invoice.paid webhook event
func (s *webhookService) handleInvoicePaid(ctx context.Context, payload *response.XenditWebhookPayload) error {
	log.Printf("[INFO] Processing invoice.paid webhook for invoice: %s", payload.ID)

	// Step 1: Get payment transaction by invoice ID
	payment, err := s.paymentRepo.GetByInvoiceID(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("payment not found for invoice %s: %w", payload.ID, err)
	}

	// Step 2: Check if already paid (double webhook prevention)
	if payment.IsPaid() {
		log.Printf("[INFO] Payment already marked as paid: %s", payment.ID)
		return nil
	}

	// Step 3: Update payment status to paid
	paidAt := payload.PaidAt
	paymentMethod := payload.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = payload.PaymentChannel
	}

	payment.Status = entity.PaymentStatusPaid
	payment.PaidAt = &paidAt
	payment.PaymentMethod = &paymentMethod

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	log.Printf("[INFO] Payment marked as paid: %s (order: %s)", payment.ID, payment.OrderID)

	// Step 4: Call Ticketing Service to confirm payment and generate tickets
	confirmReq := &client.ConfirmPaymentRequest{
		PaymentID:     payload.ID,
		PaymentMethod: paymentMethod,
		Amount:        payload.PaidAmount,
	}

	// Check if ticketing client is available
	if s.ticketingClient == nil {
		log.Printf("[WARNING] Ticketing Service gRPC client not available, cannot confirm payment for order %s", payment.OrderID)
		log.Printf("[WARNING] Payment is marked as paid, but tickets need to be generated manually or via retry")
		// Payment is already marked as paid - this should be retried via background job
		// TODO: Add to retry queue
		return nil
	}

	if err := s.ticketingClient.ConfirmPayment(payment.OrderID, confirmReq); err != nil {
		log.Printf("[ERROR] Failed to confirm payment with ticketing service: %v", err)
		// Don't return error - payment is already marked as paid
		// This should be retried via background job
		// TODO: Add to retry queue
		return nil
	}

	log.Printf("[INFO] Successfully confirmed payment with ticketing service (order: %s)", payment.OrderID)
	return nil
}

// handleInvoiceExpired handles invoice.expired webhook event
func (s *webhookService) handleInvoiceExpired(ctx context.Context, payload *response.XenditWebhookPayload) error {
	log.Printf("[INFO] Processing invoice.expired webhook for invoice: %s", payload.ID)

	// Get payment transaction by invoice ID
	payment, err := s.paymentRepo.GetByInvoiceID(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("payment not found for invoice %s: %w", payload.ID, err)
	}

	// Only update if still pending
	if payment.Status == entity.PaymentStatusPending {
		payment.Status = entity.PaymentStatusExpired
		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			return fmt.Errorf("failed to update payment status: %w", err)
		}
		log.Printf("[INFO] Payment marked as expired: %s (order: %s)", payment.ID, payment.OrderID)
	}

	return nil
}
