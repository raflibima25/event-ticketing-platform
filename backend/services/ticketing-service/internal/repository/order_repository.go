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
	ErrOrderNotFound = errors.New("order not found")
)

// OrderRepository defines interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id string) (*entity.Order, error)
	GetByIDWithLock(ctx context.Context, tx *sql.Tx, id string) (*entity.Order, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]entity.Order, int64, error)
	Update(ctx context.Context, order *entity.Order) error
	UpdateWithTx(ctx context.Context, tx *sql.Tx, order *entity.Order) error
	GetExpiredReservations(ctx context.Context) ([]entity.Order, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

// orderRepository implements OrderRepository interface
type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates new order repository instance
func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

// BeginTx starts a new transaction
func (r *orderRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// Create inserts new order into database
func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	query := `
		INSERT INTO orders (
			id, user_id, event_id, total_amount, platform_fee, service_fee,
			grand_total, status, reservation_expires_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	order.ID = uuid.New().String()

	err := r.db.QueryRowContext(
		ctx,
		query,
		order.ID,
		order.UserID,
		order.EventID,
		order.TotalAmount,
		order.PlatformFee,
		order.ServiceFee,
		order.GrandTotal,
		order.Status,
		order.ReservationExpiresAt,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID retrieves order by ID
func (r *orderRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	query := `
		SELECT id, user_id, event_id, total_amount, platform_fee, service_fee,
		       grand_total, status, payment_id, payment_method, reservation_expires_at,
		       created_at, updated_at, completed_at
		FROM orders
		WHERE id = $1
	`

	order := &entity.Order{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.EventID,
		&order.TotalAmount,
		&order.PlatformFee,
		&order.ServiceFee,
		&order.GrandTotal,
		&order.Status,
		&order.PaymentID,
		&order.PaymentMethod,
		&order.ReservationExpiresAt,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrOrderNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetByIDWithLock retrieves order by ID with row-level lock (SELECT FOR UPDATE)
// CRITICAL: This must be called within a transaction
func (r *orderRepository) GetByIDWithLock(ctx context.Context, tx *sql.Tx, id string) (*entity.Order, error) {
	query := `
		SELECT id, user_id, event_id, total_amount, platform_fee, service_fee,
		       grand_total, status, payment_id, payment_method, reservation_expires_at,
		       created_at, updated_at, completed_at
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`

	order := &entity.Order{}
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.EventID,
		&order.TotalAmount,
		&order.PlatformFee,
		&order.ServiceFee,
		&order.GrandTotal,
		&order.Status,
		&order.PaymentID,
		&order.PaymentMethod,
		&order.ReservationExpiresAt,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrOrderNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get order with lock: %w", err)
	}

	return order, nil
}

// GetByUserID retrieves all orders for a user with pagination
func (r *orderRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]entity.Order, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get orders
	query := `
		SELECT id, user_id, event_id, total_amount, platform_fee, service_fee,
		       grand_total, status, payment_id, payment_method, reservation_expires_at,
		       created_at, updated_at, completed_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	orders := []entity.Order{}
	for rows.Next() {
		var order entity.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.EventID,
			&order.TotalAmount,
			&order.PlatformFee,
			&order.ServiceFee,
			&order.GrandTotal,
			&order.Status,
			&order.PaymentID,
			&order.PaymentMethod,
			&order.ReservationExpiresAt,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.CompletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

// Update updates order information
func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	query := `
		UPDATE orders
		SET status = $1, payment_id = $2, payment_method = $3,
		    completed_at = $4, updated_at = NOW()
		WHERE id = $5
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		order.Status,
		order.PaymentID,
		order.PaymentMethod,
		order.CompletedAt,
		order.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrOrderNotFound
	}

	return nil
}

// UpdateWithTx updates order within a transaction
func (r *orderRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, order *entity.Order) error {
	query := `
		UPDATE orders
		SET status = $1, payment_id = $2, payment_method = $3,
		    completed_at = $4, updated_at = NOW()
		WHERE id = $5
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		order.Status,
		order.PaymentID,
		order.PaymentMethod,
		order.CompletedAt,
		order.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrOrderNotFound
	}

	return nil
}

// GetExpiredReservations retrieves all orders with expired reservations
// Used by background worker to release inventory
func (r *orderRepository) GetExpiredReservations(ctx context.Context) ([]entity.Order, error) {
	query := `
		SELECT id, user_id, event_id, total_amount, platform_fee, service_fee,
		       grand_total, status, payment_id, payment_method, reservation_expires_at,
		       created_at, updated_at, completed_at
		FROM orders
		WHERE status = $1 AND reservation_expires_at < $2
		ORDER BY reservation_expires_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, entity.OrderStatusReserved, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get expired reservations: %w", err)
	}
	defer rows.Close()

	orders := []entity.Order{}
	for rows.Next() {
		var order entity.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.EventID,
			&order.TotalAmount,
			&order.PlatformFee,
			&order.ServiceFee,
			&order.GrandTotal,
			&order.Status,
			&order.PaymentID,
			&order.PaymentMethod,
			&order.ReservationExpiresAt,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expired order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}
