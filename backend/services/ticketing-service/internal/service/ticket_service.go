package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/utility"
)

var (
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrTicketAlreadyUsed = errors.New("ticket has already been used")
	ErrTicketInvalid     = errors.New("ticket is invalid")
)

// TicketService handles e-ticket operations
type TicketService interface {
	GenerateTickets(ctx context.Context, orderID string) ([]response.TicketResponse, error)
	GetTicket(ctx context.Context, userID, ticketID string) (*response.TicketResponse, error)
	GetUserTickets(ctx context.Context, userID string) ([]response.TicketResponse, error)
	ValidateTicket(ctx context.Context, req *request.ValidateTicketRequest) (*response.TicketResponse, error)
}

// ticketService implements TicketService interface
type ticketService struct {
	ticketRepo    repository.TicketRepository
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
}

// NewTicketService creates new ticket service instance
func NewTicketService(
	ticketRepo repository.TicketRepository,
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
) TicketService {
	return &ticketService{
		ticketRepo:    ticketRepo,
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
	}
}

// GenerateTickets generates e-tickets for a paid order
// This is called after payment confirmation
func (s *ticketService) GenerateTickets(ctx context.Context, orderID string) ([]response.TicketResponse, error) {
	// Get order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Verify order is paid
	if order.Status != entity.OrderStatusPaid {
		return nil, fmt.Errorf("order is not in paid status")
	}

	// Get order items
	items, err := s.orderItemRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	// Start transaction
	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Generate tickets for each order item
	tickets := []entity.Ticket{}
	ticketCounter := 1

	for _, item := range items {
		for i := 0; i < item.Quantity; i++ {
			// Generate unique ticket ID and number
			ticketID := uuid.New().String()
			ticketNumber := fmt.Sprintf("TKT-%s-%03d", orderID[:8], ticketCounter)

			// Generate QR code data
			qrData := utility.GenerateTicketQRData(ticketID, order.EventID)

			// Generate QR code image (base64)
			qrCode, err := utility.GenerateQRCode(qrData)
			if err != nil {
				return nil, fmt.Errorf("failed to generate QR code: %w", err)
			}

			ticket := entity.Ticket{
				ID:           ticketID,
				OrderID:      orderID,
				OrderItemID:  item.ID,
				TicketTierID: item.TicketTierID,
				EventID:      order.EventID,
				UserID:       order.UserID,
				TicketNumber: ticketNumber,
				QRCode:       qrCode,
				QRData:       qrData,
				Status:       entity.TicketStatusValid,
			}

			tickets = append(tickets, ticket)
			ticketCounter++
		}
	}

	// Insert tickets in batch
	if err := s.ticketRepo.CreateBatch(ctx, tx, tickets); err != nil {
		return nil, fmt.Errorf("failed to create tickets: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert to response
	ticketResponses := make([]response.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		ticketResponses[i] = *response.ToTicketResponse(&ticket)
	}

	return ticketResponses, nil
}

// GetTicket retrieves a single ticket with authorization check
func (s *ticketService) GetTicket(ctx context.Context, userID, ticketID string) (*response.TicketResponse, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Check authorization
	if ticket.UserID != userID {
		return nil, ErrUnauthorized
	}

	return response.ToTicketResponse(ticket), nil
}

// GetUserTickets retrieves all tickets for a user
func (s *ticketService) GetUserTickets(ctx context.Context, userID string) ([]response.TicketResponse, error) {
	tickets, err := s.ticketRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tickets: %w", err)
	}

	// Convert to response
	ticketResponses := make([]response.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		ticketResponses[i] = *response.ToTicketResponse(&ticket)
	}

	return ticketResponses, nil
}

// ValidateTicket validates a ticket at event entrance
// This is called by event staff to scan and validate tickets
func (s *ticketService) ValidateTicket(ctx context.Context, req *request.ValidateTicketRequest) (*response.TicketResponse, error) {
	// Parse QR data to extract ticket ID and event ID
	ticketID, eventID, err := utility.ParseTicketQRData(req.QRData)
	if err != nil {
		return nil, ErrTicketInvalid
	}

	// Get ticket
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Verify ticket belongs to the event
	if ticket.EventID != eventID {
		return nil, ErrTicketInvalid
	}

	// Check if ticket can be used
	if !ticket.CanBeUsed() {
		if ticket.IsUsed() {
			return nil, ErrTicketAlreadyUsed
		}
		return nil, ErrTicketInvalid
	}

	// Mark ticket as used
	if err := s.ticketRepo.MarkAsUsed(ctx, ticketID); err != nil {
		return nil, fmt.Errorf("failed to mark ticket as used: %w", err)
	}

	// Get updated ticket
	ticket, err = s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated ticket: %w", err)
	}

	return response.ToTicketResponse(ticket), nil
}
