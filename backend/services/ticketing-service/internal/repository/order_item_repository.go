package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
)

// OrderItemRepository defines interface for order item data operations
type OrderItemRepository interface {
	Create(ctx context.Context, tx *sql.Tx, item *entity.OrderItem) error
	CreateBatch(ctx context.Context, tx *sql.Tx, items []entity.OrderItem) error
	GetByOrderID(ctx context.Context, orderID string) ([]entity.OrderItem, error)
	GetByID(ctx context.Context, id string) (*entity.OrderItem, error)
}

// orderItemRepository implements OrderItemRepository interface
type orderItemRepository struct {
	db *sql.DB
}

// NewOrderItemRepository creates new order item repository instance
func NewOrderItemRepository(db *sql.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

// Create inserts new order item (must be called within a transaction)
func (r *orderItemRepository) Create(ctx context.Context, tx *sql.Tx, item *entity.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, ticket_tier_id, quantity, price, subtotal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	item.ID = uuid.New().String()
	item.Subtotal = item.CalculateSubtotal()

	err := tx.QueryRowContext(
		ctx,
		query,
		item.ID,
		item.OrderID,
		item.TicketTierID,
		item.Quantity,
		item.Price,
		item.Subtotal,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}

	return nil
}

// CreateBatch inserts multiple order items in one transaction
func (r *orderItemRepository) CreateBatch(ctx context.Context, tx *sql.Tx, items []entity.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, ticket_tier_id, quantity, price, subtotal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i := range items {
		items[i].ID = uuid.New().String()
		items[i].Subtotal = items[i].CalculateSubtotal()

		_, err := stmt.ExecContext(
			ctx,
			items[i].ID,
			items[i].OrderID,
			items[i].TicketTierID,
			items[i].Quantity,
			items[i].Price,
			items[i].Subtotal,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return nil
}

// GetByOrderID retrieves all items for an order
func (r *orderItemRepository) GetByOrderID(ctx context.Context, orderID string) ([]entity.OrderItem, error) {
	query := `
		SELECT id, order_id, ticket_tier_id, quantity, price, subtotal, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	items := []entity.OrderItem{}
	for rows.Next() {
		var item entity.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.TicketTierID,
			&item.Quantity,
			&item.Price,
			&item.Subtotal,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// GetByID retrieves order item by ID
func (r *orderItemRepository) GetByID(ctx context.Context, id string) (*entity.OrderItem, error) {
	query := `
		SELECT id, order_id, ticket_tier_id, quantity, price, subtotal, created_at, updated_at
		FROM order_items
		WHERE id = $1
	`

	item := &entity.OrderItem{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.OrderID,
		&item.TicketTierID,
		&item.Quantity,
		&item.Price,
		&item.Subtotal,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order item not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get order item: %w", err)
	}

	return item, nil
}
