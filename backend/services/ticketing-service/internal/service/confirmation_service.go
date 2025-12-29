package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/repository"
)

var (
	ErrOrderNotInReservedStatus = errors.New("order is not in reserved status")
	ErrAmountMismatch           = errors.New("payment amount mismatch")
)

// ConfirmationService handles order confirmation after payment
type ConfirmationService interface {
	ConfirmPayment(ctx context.Context, req *request.ConfirmOrderRequest) error
}

// confirmationService implements ConfirmationService interface
type confirmationService struct {
	orderRepo      repository.OrderRepository
	ticketService  TicketService
}

// NewConfirmationService creates new confirmation service instance
func NewConfirmationService(
	orderRepo repository.OrderRepository,
	ticketService TicketService,
) ConfirmationService {
	return &confirmationService{
		orderRepo:     orderRepo,
		ticketService: ticketService,
	}
}

// ConfirmPayment confirms payment and generates tickets
// This is called by Payment Service after successful payment
func (s *confirmationService) ConfirmPayment(ctx context.Context, req *request.ConfirmOrderRequest) error {
	// Start transaction
	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get order with lock
	order, err := s.orderRepo.GetByIDWithLock(ctx, tx, req.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Verify order is in reserved status
	if order.Status != entity.OrderStatusReserved {
		return ErrOrderNotInReservedStatus
	}

	// Verify order hasn't expired
	if order.IsExpired() {
		return ErrOrderExpired
	}

	// Verify amount matches
	if req.Amount != order.GrandTotal {
		return fmt.Errorf("%w: expected %.2f, got %.2f", ErrAmountMismatch, order.GrandTotal, req.Amount)
	}

	// Update order status to paid
	paymentID := req.PaymentID
	paymentMethod := req.PaymentMethod
	completedAt := time.Now()

	order.Status = entity.OrderStatusPaid
	order.PaymentID = &paymentID
	order.PaymentMethod = &paymentMethod
	order.CompletedAt = &completedAt

	if err := s.orderRepo.UpdateWithTx(ctx, tx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Generate e-tickets (outside transaction for better performance)
	if _, err := s.ticketService.GenerateTickets(ctx, req.OrderID); err != nil {
		// Log error but don't fail - tickets can be regenerated later
		// TODO: Add to retry queue
		return fmt.Errorf("warning: failed to generate tickets: %w", err)
	}

	// TODO: Publish event to notification service
	// PublishEvent("ticket.created", {orderID, userID, email})

	return nil
}
