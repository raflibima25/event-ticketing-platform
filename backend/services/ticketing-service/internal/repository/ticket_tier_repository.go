package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrTicketTierNotFound = errors.New("ticket tier not found")
	ErrInsufficientQuota  = errors.New("insufficient ticket quota")
)

// TicketTier represents ticket tier data (read-only from event service)
type TicketTier struct {
	ID          string
	EventID     string
	Name        string
	Price       float64
	Quota       int
	SoldCount   int
	MaxPerOrder int
}

// TicketTierRepository defines interface for ticket tier operations
type TicketTierRepository interface {
	GetByID(ctx context.Context, id string) (*TicketTier, error)
	GetByIDWithLock(ctx context.Context, tx *sql.Tx, id string) (*TicketTier, error)
	GetByEventID(ctx context.Context, eventID string) ([]TicketTier, error)
	CheckAvailability(ctx context.Context, tierID string, quantity int) (bool, error)
	UpdateSoldCount(ctx context.Context, tx *sql.Tx, tierID string, quantity int) error
	ReleaseSoldCount(ctx context.Context, tx *sql.Tx, tierID string, quantity int) error
}

// ticketTierRepository implements TicketTierRepository interface
type ticketTierRepository struct {
	db *sql.DB
}

// NewTicketTierRepository creates new ticket tier repository instance
func NewTicketTierRepository(db *sql.DB) TicketTierRepository {
	return &ticketTierRepository{db: db}
}

// GetByID retrieves ticket tier by ID
func (r *ticketTierRepository) GetByID(ctx context.Context, id string) (*TicketTier, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold_count, max_per_order
		FROM ticket_tiers
		WHERE id = $1
	`

	tier := &TicketTier{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
		return nil, fmt.Errorf("failed to get ticket tier: %w", err)
	}

	return tier, nil
}

// GetByIDWithLock retrieves ticket tier with row-level lock (SELECT FOR UPDATE)
// CRITICAL: Prevents race conditions in concurrent reservations
// MUST be called within a transaction
func (r *ticketTierRepository) GetByIDWithLock(ctx context.Context, tx *sql.Tx, id string) (*TicketTier, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold_count, max_per_order
		FROM ticket_tiers
		WHERE id = $1
		FOR UPDATE
	`

	tier := &TicketTier{}
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

// GetByEventID retrieves all ticket tiers for an event
func (r *ticketTierRepository) GetByEventID(ctx context.Context, eventID string) ([]TicketTier, error) {
	query := `
		SELECT id, event_id, name, price, quota, sold_count, max_per_order
		FROM ticket_tiers
		WHERE event_id = $1
		ORDER BY price ASC
	`

	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tiers: %w", err)
	}
	defer rows.Close()

	tiers := []TicketTier{}
	for rows.Next() {
		var tier TicketTier
		err := rows.Scan(
			&tier.ID,
			&tier.EventID,
			&tier.Name,
			&tier.Price,
			&tier.Quota,
			&tier.SoldCount,
			&tier.MaxPerOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket tier: %w", err)
		}
		tiers = append(tiers, tier)
	}

	return tiers, nil
}

// CheckAvailability checks if requested quantity is available
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

// UpdateSoldCount increments sold count (for reservation/payment)
// CRITICAL: Uses database constraint to prevent overselling
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
