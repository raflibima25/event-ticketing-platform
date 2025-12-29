package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
)

// TicketRepository defines interface for ticket data operations
type TicketRepository interface {
	Create(ctx context.Context, tx *sql.Tx, ticket *entity.Ticket) error
	CreateBatch(ctx context.Context, tx *sql.Tx, tickets []entity.Ticket) error
	GetByID(ctx context.Context, id string) (*entity.Ticket, error)
	GetByOrderID(ctx context.Context, orderID string) ([]entity.Ticket, error)
	GetByUserID(ctx context.Context, userID string) ([]entity.Ticket, error)
	Update(ctx context.Context, ticket *entity.Ticket) error
	MarkAsUsed(ctx context.Context, ticketID string) error
}

// ticketRepository implements TicketRepository interface
type ticketRepository struct {
	db *sql.DB
}

// NewTicketRepository creates new ticket repository instance
func NewTicketRepository(db *sql.DB) TicketRepository {
	return &ticketRepository{db: db}
}

// Create inserts new ticket (must be called within a transaction)
func (r *ticketRepository) Create(ctx context.Context, tx *sql.Tx, ticket *entity.Ticket) error {
	query := `
		INSERT INTO tickets (
			id, order_id, order_item_id, ticket_tier_id, event_id, user_id,
			ticket_number, qr_code, qr_data, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	ticket.ID = uuid.New().String()

	err := tx.QueryRowContext(
		ctx,
		query,
		ticket.ID,
		ticket.OrderID,
		ticket.OrderItemID,
		ticket.TicketTierID,
		ticket.EventID,
		ticket.UserID,
		ticket.TicketNumber,
		ticket.QRCode,
		ticket.QRData,
		ticket.Status,
	).Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	return nil
}

// CreateBatch inserts multiple tickets in one transaction
func (r *ticketRepository) CreateBatch(ctx context.Context, tx *sql.Tx, tickets []entity.Ticket) error {
	query := `
		INSERT INTO tickets (
			id, order_id, order_item_id, ticket_tier_id, event_id, user_id,
			ticket_number, qr_code, qr_data, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i := range tickets {
		// Don't generate new ID - use the one already set in the ticket
		// The ID is already generated in the service layer with QR data
		if tickets[i].ID == "" {
			tickets[i].ID = uuid.New().String()
		}

		_, err := stmt.ExecContext(
			ctx,
			tickets[i].ID,
			tickets[i].OrderID,
			tickets[i].OrderItemID,
			tickets[i].TicketTierID,
			tickets[i].EventID,
			tickets[i].UserID,
			tickets[i].TicketNumber,
			tickets[i].QRCode,
			tickets[i].QRData,
			tickets[i].Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert ticket: %w", err)
		}
	}

	return nil
}

// GetByID retrieves ticket by ID
func (r *ticketRepository) GetByID(ctx context.Context, id string) (*entity.Ticket, error) {
	query := `
		SELECT id, order_id, order_item_id, ticket_tier_id, event_id, user_id,
		       ticket_number, qr_code, qr_data, status, validated_at, created_at, updated_at
		FROM tickets
		WHERE id = $1
	`

	ticket := &entity.Ticket{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ticket.ID,
		&ticket.OrderID,
		&ticket.OrderItemID,
		&ticket.TicketTierID,
		&ticket.EventID,
		&ticket.UserID,
		&ticket.TicketNumber,
		&ticket.QRCode,
		&ticket.QRData,
		&ticket.Status,
		&ticket.UsedAt,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTicketNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return ticket, nil
}

// GetByOrderID retrieves all tickets for an order
func (r *ticketRepository) GetByOrderID(ctx context.Context, orderID string) ([]entity.Ticket, error) {
	query := `
		SELECT id, order_id, order_item_id, ticket_tier_id, event_id, user_id,
		       ticket_number, qr_code, qr_data, status, validated_at, created_at, updated_at
		FROM tickets
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets by order: %w", err)
	}
	defer rows.Close()

	tickets := []entity.Ticket{}
	for rows.Next() {
		var ticket entity.Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.OrderID,
			&ticket.OrderItemID,
			&ticket.TicketTierID,
			&ticket.EventID,
			&ticket.UserID,
			&ticket.TicketNumber,
			&ticket.QRCode,
			&ticket.QRData,
			&ticket.Status,
			&ticket.UsedAt,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

// GetByUserID retrieves all tickets for a user
func (r *ticketRepository) GetByUserID(ctx context.Context, userID string) ([]entity.Ticket, error) {
	query := `
		SELECT id, order_id, order_item_id, ticket_tier_id, event_id, user_id,
		       ticket_number, qr_code, qr_data, status, validated_at, created_at, updated_at
		FROM tickets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tickets: %w", err)
	}
	defer rows.Close()

	tickets := []entity.Ticket{}
	for rows.Next() {
		var ticket entity.Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.OrderID,
			&ticket.OrderItemID,
			&ticket.TicketTierID,
			&ticket.EventID,
			&ticket.UserID,
			&ticket.TicketNumber,
			&ticket.QRCode,
			&ticket.QRData,
			&ticket.Status,
			&ticket.UsedAt,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

// Update updates ticket information
func (r *ticketRepository) Update(ctx context.Context, ticket *entity.Ticket) error {
	query := `
		UPDATE tickets
		SET status = $1, validated_at = $2, updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		ticket.Status,
		ticket.UsedAt,
		ticket.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrTicketNotFound
	}

	return nil
}

// MarkAsUsed marks a ticket as used (scanned at event entrance)
func (r *ticketRepository) MarkAsUsed(ctx context.Context, ticketID string) error {
	query := `
		UPDATE tickets
		SET status = $1, validated_at = $2, updated_at = NOW()
		WHERE id = $3 AND status = $4
	`

	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		query,
		entity.TicketStatusUsed,
		now,
		ticketID,
		entity.TicketStatusValid,
	)

	if err != nil {
		return fmt.Errorf("failed to mark ticket as used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("ticket not found or already used")
	}

	return nil
}
