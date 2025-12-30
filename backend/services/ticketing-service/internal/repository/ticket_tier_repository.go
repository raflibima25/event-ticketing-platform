package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
)

var (
	ErrTicketTierNotFound = errors.New("ticket tier not found")
	ErrInsufficientQuota  = errors.New("insufficient ticket quota")
)

// TicketTierRepository defines interface for ticket tier operations
type TicketTierRepository interface {
	GetByID(ctx context.Context, id string) (*entity.TicketTier, error)
	GetByIDWithLock(ctx context.Context, tx *sql.Tx, id string) (*entity.TicketTier, error)
	GetByEventID(ctx context.Context, eventID string) ([]entity.TicketTier, error)
	CheckAvailability(ctx context.Context, tierID string, quantity int) (bool, error)
	UpdateSoldCount(ctx context.Context, tx *sql.Tx, tierID string, quantity int) error
	ReleaseSoldCount(ctx context.Context, tx *sql.Tx, tierID string, quantity int) error
}

// ticketTierRepository implements TicketTierRepository interface
type ticketTierRepository struct {
	db *sqlx.DB
}

// NewTicketTierRepository creates new ticket tier repository instance
func NewTicketTierRepository(db *sqlx.DB) TicketTierRepository {
	return &ticketTierRepository{db: db}
}

// GetByID retrieves ticket tier by ID using sqlx
func (r *ticketTierRepository) GetByID(ctx context.Context, id string) (*entity.TicketTier, error) {
	var tier entity.TicketTier
	query := `
		SELECT id, event_id, name, price, quota, sold_count, max_per_order
		FROM ticket_tiers
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &tier, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrTicketTierNotFound
		}
		return nil, fmt.Errorf("failed to get ticket tier: %w", err)
	}

	return &tier, nil
}

// GetByIDWithLock retrieves ticket tier with row-level lock (SELECT FOR UPDATE)
// CRITICAL PATH: Uses raw SQL for explicit locking control
// PREVENTS RACE CONDITIONS in concurrent reservations
// MUST be called within a transaction
func (r *ticketTierRepository) GetByIDWithLock(ctx context.Context, tx *sql.Tx, id string) (*entity.TicketTier, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold_count, max_per_order
		FROM ticket_tiers
		WHERE id = $1
		FOR UPDATE
	`

	tier := &entity.TicketTier{}
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&tier.ID,
		&tier.EventID,
		&tier.Name,
		&tier.Price,
		&tier.Quota,
		&tier.SoldCount,
		&tier.MaxPerOrder,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTicketTierNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tier with lock: %w", err)
	}

	return tier, nil
}

// GetByEventID retrieves all ticket tiers for an event using sqlx
func (r *ticketTierRepository) GetByEventID(ctx context.Context, eventID string) ([]entity.TicketTier, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold_count, max_per_order
		FROM ticket_tiers
		WHERE event_id = $1
		ORDER BY price ASC
	`

	tiers := []entity.TicketTier{}
	err := r.db.SelectContext(ctx, &tiers, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tiers: %w", err)
	}

	return tiers, nil
}

// CheckAvailability checks if requested quantity is available using sqlx
func (r *ticketTierRepository) CheckAvailability(ctx context.Context, tierID string, quantity int) (bool, error) {
	var available bool
	query := `
		SELECT (quota - sold_count) >= $1 as available
		FROM ticket_tiers
		WHERE id = $2
	`

	err := r.db.GetContext(ctx, &available, query, quantity, tierID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, ErrTicketTierNotFound
		}
		return false, fmt.Errorf("failed to check availability: %w", err)
	}

	return available, nil
}

// UpdateSoldCount increments sold count (for reservation/payment)
// CRITICAL PATH: Uses raw SQL transaction for atomic operation
// Database constraint prevents overselling: (sold_count + $1) <= quota
// MUST be called within a transaction with row-level lock
func (r *ticketTierRepository) UpdateSoldCount(ctx context.Context, tx *sql.Tx, tierID string, quantity int) error {
	query := `
		UPDATE ticket_tiers
		SET sold_count = sold_count + $1, updated_at = NOW()
		WHERE id = $2 AND (sold_count + $1) <= quota
	`

	result, err := tx.ExecContext(ctx, query, quantity, tierID)
	if err != nil {
		return fmt.Errorf("failed to update sold count: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		// Check if tier exists or quota exceeded
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

// ReleaseSoldCount decrements sold count (for cancellation/expiration)
// CRITICAL PATH: Uses raw SQL transaction for atomic operation
// MUST be called within a transaction
func (r *ticketTierRepository) ReleaseSoldCount(ctx context.Context, tx *sql.Tx, tierID string, quantity int) error {
	query := `
		UPDATE ticket_tiers
		SET sold_count = GREATEST(sold_count - $1, 0), updated_at = NOW()
		WHERE id = $2
	`

	result, err := tx.ExecContext(ctx, query, quantity, tierID)
	if err != nil {
		return fmt.Errorf("failed to release sold count: %w", err)
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
