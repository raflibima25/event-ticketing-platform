package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/entity"
)

var (
	ErrPaymentNotFound = errors.New("payment transaction not found")
)

// PaymentRepository defines interface for payment data operations
type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.PaymentTransaction) error
	GetByID(ctx context.Context, id string) (*entity.PaymentTransaction, error)
	GetByOrderID(ctx context.Context, orderID string) (*entity.PaymentTransaction, error)
	GetByExternalID(ctx context.Context, externalID string) (*entity.PaymentTransaction, error)
	GetByInvoiceID(ctx context.Context, invoiceID string) (*entity.PaymentTransaction, error)
	Update(ctx context.Context, payment *entity.PaymentTransaction) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

// paymentRepository implements PaymentRepository interface
type paymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates new payment repository instance
func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// Create inserts new payment transaction
func (r *paymentRepository) Create(ctx context.Context, payment *entity.PaymentTransaction) error {
	query := `
		INSERT INTO payment_transactions (
			id, order_id, external_id, invoice_id, invoice_url,
			amount, payment_method, status, paid_at, expires_at,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	payment.ID = uuid.New().String()

	err := r.db.QueryRowContext(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.ExternalID,
		payment.InvoiceID,
		payment.InvoiceURL,
		payment.Amount,
		payment.PaymentMethod,
		payment.Status,
		payment.PaidAt,
		payment.ExpiresAt,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create payment transaction: %w", err)
	}

	return nil
}

// GetByID retrieves payment transaction by ID
func (r *paymentRepository) GetByID(ctx context.Context, id string) (*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, external_id, invoice_id, invoice_url,
		       amount, payment_method, status, paid_at, expires_at,
		       created_at, updated_at
		FROM payment_transactions
		WHERE id = $1
	`

	payment := &entity.PaymentTransaction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.ExternalID,
		&payment.InvoiceID,
		&payment.InvoiceURL,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.PaidAt,
		&payment.ExpiresAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction: %w", err)
	}

	return payment, nil
}

// GetByOrderID retrieves payment transaction by order ID
func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID string) (*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, external_id, invoice_id, invoice_url,
		       amount, payment_method, status, paid_at, expires_at,
		       created_at, updated_at
		FROM payment_transactions
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	payment := &entity.PaymentTransaction{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.ExternalID,
		&payment.InvoiceID,
		&payment.InvoiceURL,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.PaidAt,
		&payment.ExpiresAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction: %w", err)
	}

	return payment, nil
}

// GetByExternalID retrieves payment transaction by external ID
func (r *paymentRepository) GetByExternalID(ctx context.Context, externalID string) (*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, external_id, invoice_id, invoice_url,
		       amount, payment_method, status, paid_at, expires_at,
		       created_at, updated_at
		FROM payment_transactions
		WHERE external_id = $1
	`

	payment := &entity.PaymentTransaction{}
	err := r.db.QueryRowContext(ctx, query, externalID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.ExternalID,
		&payment.InvoiceID,
		&payment.InvoiceURL,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.PaidAt,
		&payment.ExpiresAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction: %w", err)
	}

	return payment, nil
}

// GetByInvoiceID retrieves payment transaction by invoice ID
func (r *paymentRepository) GetByInvoiceID(ctx context.Context, invoiceID string) (*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, external_id, invoice_id, invoice_url,
		       amount, payment_method, status, paid_at, expires_at,
		       created_at, updated_at
		FROM payment_transactions
		WHERE invoice_id = $1
	`

	payment := &entity.PaymentTransaction{}
	err := r.db.QueryRowContext(ctx, query, invoiceID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.ExternalID,
		&payment.InvoiceID,
		&payment.InvoiceURL,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.PaidAt,
		&payment.ExpiresAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction: %w", err)
	}

	return payment, nil
}

// Update updates payment transaction
func (r *paymentRepository) Update(ctx context.Context, payment *entity.PaymentTransaction) error {
	query := `
		UPDATE payment_transactions
		SET invoice_id = $1, invoice_url = $2, payment_method = $3,
		    status = $4, paid_at = $5, updated_at = NOW()
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		payment.InvoiceID,
		payment.InvoiceURL,
		payment.PaymentMethod,
		payment.Status,
		payment.PaidAt,
		payment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment transaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

// BeginTx starts a new database transaction
func (r *paymentRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
