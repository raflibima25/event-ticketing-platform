package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/payload/entity"
)

var (
	ErrTicketTierNotFound = errors.New("ticket tier not found")
	ErrInsufficientQuota  = errors.New("insufficient ticket quota")
)

// TicketTierRepository defines interface for ticket tier data operations
type TicketTierRepository interface {
	Create(ctx context.Context, tier *entity.TicketTier) error
	GetByID(ctx context.Context, id string) (*entity.TicketTier, error)
	GetByEventID(ctx context.Context, eventID string) ([]entity.TicketTier, error)
	Update(ctx context.Context, tier *entity.TicketTier) error
	Delete(ctx context.Context, id string) error
	CheckAvailability(ctx context.Context, tierID string, quantity int) (bool, error)
	UpdateSoldCount(ctx context.Context, tierID string, quantity int) error
}

// ticketTierRepository implements TicketTierRepository interface
type ticketTierRepository struct {
	db *sql.DB
}

// NewTicketTierRepository creates new ticket tier repository instance
func NewTicketTierRepository(db *sql.DB) TicketTierRepository {
	return &ticketTierRepository{db: db}
}

// Create inserts new ticket tier into database
func (r *ticketTierRepository) Create(ctx context.Context, tier *entity.TicketTier) error {
	query := `
		INSERT INTO ticket_tiers (id, event_id, name, description, price, quota, sold_count,
		                         max_per_order, early_bird_price, early_bird_end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	tier.ID = uuid.New().String()
	tier.SoldCount = 0 // Initialize sold count

	err := r.db.QueryRowContext(
		ctx,
		query,
		tier.ID,
		tier.EventID,
		tier.Name,
		tier.Description,
		tier.Price,
		tier.Quota,
		tier.SoldCount,
		tier.MaxPerOrder,
		tier.EarlyBirdPrice,
		tier.EarlyBirdEndDate,
	).Scan(&tier.ID, &tier.CreatedAt, &tier.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create ticket tier: %w", err)
	}

	return nil
}

// GetByID retrieves ticket tier by ID
func (r *ticketTierRepository) GetByID(ctx context.Context, id string) (*entity.TicketTier, error) {
	query := `
		SELECT id, event_id, name, description, price, quota, sold_count, max_per_order,
		       early_bird_price, early_bird_end_date, created_at, updated_at
		FROM ticket_tiers
		WHERE id = $1
	`

	tier := &entity.TicketTier{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tier.ID,
		&tier.EventID,
		&tier.Name,
		&tier.Description,
		&tier.Price,
		&tier.Quota,
		&tier.SoldCount,
		&tier.MaxPerOrder,
		&tier.EarlyBirdPrice,
		&tier.EarlyBirdEndDate,
		&tier.CreatedAt,
		&tier.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTicketTierNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tier: %w", err)
	}

	return tier, nil
}

// GetByEventID retrieves all ticket tiers for an event
func (r *ticketTierRepository) GetByEventID(ctx context.Context, eventID string) ([]entity.TicketTier, error) {
	query := `
		SELECT id, event_id, name, description, price, quota, sold_count, max_per_order,
		       early_bird_price, early_bird_end_date, created_at, updated_at
		FROM ticket_tiers
		WHERE event_id = $1
		ORDER BY price ASC
	`

	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tiers by event: %w", err)
	}
	defer rows.Close()

	tiers := []entity.TicketTier{}
	for rows.Next() {
		var tier entity.TicketTier
		err := rows.Scan(
			&tier.ID,
			&tier.EventID,
			&tier.Name,
			&tier.Description,
			&tier.Price,
			&tier.Quota,
			&tier.SoldCount,
			&tier.MaxPerOrder,
			&tier.EarlyBirdPrice,
			&tier.EarlyBirdEndDate,
			&tier.CreatedAt,
			&tier.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket tier: %w", err)
		}
		tiers = append(tiers, tier)
	}

	return tiers, nil
}

// Update updates ticket tier information
func (r *ticketTierRepository) Update(ctx context.Context, tier *entity.TicketTier) error {
	query := `
		UPDATE ticket_tiers
		SET name = $1, description = $2, price = $3, quota = $4, max_per_order = $5,
		    early_bird_price = $6, early_bird_end_date = $7, updated_at = NOW()
		WHERE id = $8
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		tier.Name,
		tier.Description,
		tier.Price,
		tier.Quota,
		tier.MaxPerOrder,
		tier.EarlyBirdPrice,
		tier.EarlyBirdEndDate,
		tier.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update ticket tier: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrTicketTierNotFound
	}

	return nil
}

// Delete removes ticket tier from database
func (r *ticketTierRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM ticket_tiers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete ticket tier: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrTicketTierNotFound
	}

	return nil
}

// CheckAvailability checks if requested quantity is available for a ticket tier
func (r *ticketTierRepository) CheckAvailability(ctx context.Context, tierID string, quantity int) (bool, error) {
	query := `
		SELECT (quota - sold_count) >= $1 as available
		FROM ticket_tiers
		WHERE id = $2
	`

	var available bool
	err := r.db.QueryRowContext(ctx, query, quantity, tierID).Scan(&available)

	if err == sql.ErrNoRows {
		return false, ErrTicketTierNotFound
	}

	if err != nil {
		return false, fmt.Errorf("failed to check availability: %w", err)
	}

	return available, nil
}

// UpdateSoldCount increments sold count for a ticket tier
// This should be called within a transaction from the service layer
func (r *ticketTierRepository) UpdateSoldCount(ctx context.Context, tierID string, quantity int) error {
	query := `
		UPDATE ticket_tiers
		SET sold_count = sold_count + $1, updated_at = NOW()
		WHERE id = $2 AND (sold_count + $1) <= quota
	`

	result, err := r.db.ExecContext(ctx, query, quantity, tierID)
	if err != nil {
		return fmt.Errorf("failed to update sold count: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		// Either ticket tier not found or quota exceeded
		// Check which one it is
		tier, err := r.GetByID(ctx, tierID)
		if err != nil {
			return err
		}

		if tier.SoldCount+quantity > tier.Quota {
			return ErrInsufficientQuota
		}

		return ErrTicketTierNotFound
	}

	return nil
}
