package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/client"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/response"
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
	orderRepo          repository.OrderRepository
	orderItemRepo      repository.OrderItemRepository
	ticketTierRepo     repository.TicketTierRepository
	eventRepo          repository.EventRepository
	userRepo           repository.UserRepository
	ticketService      TicketService
	notificationClient *client.NotificationClient
}

// NewConfirmationService creates new confirmation service instance
func NewConfirmationService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	ticketTierRepo repository.TicketTierRepository,
	eventRepo repository.EventRepository,
	userRepo repository.UserRepository,
	ticketService TicketService,
	notificationClient *client.NotificationClient,
) ConfirmationService {
	return &confirmationService{
		orderRepo:          orderRepo,
		orderItemRepo:      orderItemRepo,
		ticketTierRepo:     ticketTierRepo,
		eventRepo:          eventRepo,
		userRepo:           userRepo,
		ticketService:      ticketService,
		notificationClient: notificationClient,
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
	tickets, err := s.ticketService.GenerateTickets(ctx, req.OrderID)
	if err != nil {
		// Log error but don't fail - tickets can be regenerated later
		// TODO: Add to retry queue
		return fmt.Errorf("warning: failed to generate tickets: %w", err)
	}

	log.Printf("[ConfirmationService] Generated %d tickets for order %s", len(tickets), req.OrderID)

	// Send e-ticket email via notification service (async with auto-reconnect)
	go s.sendTicketEmail(context.Background(), order, tickets)

	return nil
}

// sendTicketEmail sends e-ticket email asynchronously
func (s *confirmationService) sendTicketEmail(ctx context.Context, order *entity.Order, tickets []response.TicketResponse) {
	// Get order items
	orderItems, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		log.Printf("[ConfirmationService] Failed to get order items for email: %v", err)
		return
	}

	// Get event details
	event, err := s.eventRepo.GetByID(ctx, order.EventID)
	if err != nil {
		log.Printf("[ConfirmationService] Failed to get event details for %s: %v", order.EventID, err)
		// Use fallback values if event not found
		event = &repository.Event{
			Name:      "Event",
			Location:  "TBA",
			StartDate: time.Now().Add(24 * time.Hour),
		}
	} else {
		log.Printf("[ConfirmationService] ✓ Event retrieved: ID=%s, Name=%s, Location=%s", event.ID, event.Name, event.Location)
	}

	eventName := event.Name
	eventLocation := event.Location
	eventStartTime := event.StartDate.Format("Monday, 02 Jan 2006 15:04 WIB")

	// Create maps for tier prices and names from order items
	tierPrices := make(map[string]float64)
	tierNames := make(map[string]string)

	for _, item := range orderItems {
		tierPrices[item.TicketTierID] = item.Price

		// Fetch tier name if not already in map
		if _, exists := tierNames[item.TicketTierID]; !exists {
			tier, err := s.ticketTierRepo.GetByID(ctx, item.TicketTierID)
			if err != nil {
				log.Printf("[ConfirmationService] Warning: Failed to get tier name for %s: %v", item.TicketTierID, err)
				tierNames[item.TicketTierID] = "Unknown Tier"
			} else {
				tierNames[item.TicketTierID] = tier.Name
			}
		}
	}

	// Prepare ticket info for email
	ticketInfos := make([]client.TicketInfo, len(tickets))
	for i, ticket := range tickets {
		price := tierPrices[ticket.TicketTierID]
		tierName := tierNames[ticket.TicketTierID]
		if tierName == "" {
			tierName = "Unknown Tier"
		}

		ticketInfos[i] = client.TicketInfo{
			TicketID: ticket.ID,
			QRCode:   ticket.QRCode,
			TierName: tierName,
			Price:    price,
		}
	}

	// Get recipient details from user profile
	user, err := s.userRepo.GetByID(ctx, order.UserID)
	if err != nil {
		log.Printf("[ConfirmationService] Failed to get user details for %s: %v", order.UserID, err)
		// Use fallback values if user not found
		user = &repository.User{
			Email:    "customer@example.com",
			FullName: "Customer",
		}
	} else {
		log.Printf("[ConfirmationService] ✓ User retrieved: ID=%s, Email=%s, FullName=%s", user.ID, user.Email, user.FullName)
	}

	recipientEmail := user.Email
	recipientName := user.FullName
	if recipientName == "" {
		log.Printf("[ConfirmationService] ⚠️ FullName is empty, using fallback 'Customer'")
		recipientName = "Customer"
	}

	paymentMethod := "QRIS"
	if order.PaymentMethod != nil {
		paymentMethod = *order.PaymentMethod
	}

	// Send email request
	emailReq := &client.SendTicketEmailRequest{
		OrderID:        order.ID,
		RecipientEmail: recipientEmail,
		RecipientName:  recipientName,
		EventName:      eventName,
		EventLocation:  eventLocation,
		EventStartTime: eventStartTime,
		TotalAmount:    order.GrandTotal,
		PaymentMethod:  paymentMethod,
		Tickets:        ticketInfos,
	}

	log.Printf("[ConfirmationService] 📧 Sending email to: %s (%s) for event: %s at %s", recipientEmail, recipientName, eventName, eventLocation)

	if err := s.notificationClient.SendTicketEmail(ctx, emailReq); err != nil {
		log.Printf("[ConfirmationService] Failed to send ticket email for order %s: %v", order.ID, err)
		// TODO: Add to retry queue
	} else {
		log.Printf("[ConfirmationService] ✅ Ticket email sent for order %s", order.ID)
	}
}
