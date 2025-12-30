package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/entity"
)

var (
	ErrWebhookNotFound      = errors.New("webhook event not found")
	ErrDuplicateWebhook     = errors.New("webhook already processed")
)

// WebhookRepository defines interface for webhook data operations
type WebhookRepository interface {
	Create(ctx context.Context, webhook *entity.WebhookEvent) error
	GetByWebhookID(ctx context.Context, webhookID string) (*entity.WebhookEvent, error)
	MarkAsProcessed(ctx context.Context, webhookID string) error
	MarkAsFailed(ctx context.Context, webhookID string) error
}

// webhookRepository implements WebhookRepository interface
type webhookRepository struct {
	db *sql.DB
}

// NewWebhookRepository creates new webhook repository instance
func NewWebhookRepository(db *sql.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

// Create inserts new webhook event (idempotency check via unique constraint)
func (r *webhookRepository) Create(ctx context.Context, webhook *entity.WebhookEvent) error {
	query := `
		INSERT INTO webhook_events (
			id, webhook_id, event_type, payload, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`

	webhook.ID = uuid.New().String()

	err := r.db.QueryRowContext(
		ctx,
		query,
		webhook.ID,
		webhook.WebhookID,
		webhook.EventType,
		webhook.Payload,
		webhook.Status,
	).Scan(&webhook.ID, &webhook.CreatedAt)

	if err != nil {
		// Check for duplicate webhook (unique constraint violation)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateWebhook
		}
		return fmt.Errorf("failed to create webhook event: %w", err)
	}

	return nil
}

// GetByWebhookID retrieves webhook event by webhook ID
func (r *webhookRepository) GetByWebhookID(ctx context.Context, webhookID string) (*entity.WebhookEvent, error) {
	query := `
		SELECT id, webhook_id, event_type, payload, processed_at, status, created_at
		FROM webhook_events
		WHERE webhook_id = $1
	`

	webhook := &entity.WebhookEvent{}
	err := r.db.QueryRowContext(ctx, query, webhookID).Scan(
		&webhook.ID,
		&webhook.WebhookID,
		&webhook.EventType,
		&webhook.Payload,
		&webhook.ProcessedAt,
		&webhook.Status,
		&webhook.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrWebhookNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get webhook event: %w", err)
	}

	return webhook, nil
}

// MarkAsProcessed marks webhook as successfully processed
func (r *webhookRepository) MarkAsProcessed(ctx context.Context, webhookID string) error {
	query := `
		UPDATE webhook_events
		SET status = $1, processed_at = NOW()
		WHERE webhook_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, entity.WebhookStatusProcessed, webhookID)
	if err != nil {
		return fmt.Errorf("failed to mark webhook as processed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrWebhookNotFound
	}

	return nil
}

// MarkAsFailed marks webhook as failed
func (r *webhookRepository) MarkAsFailed(ctx context.Context, webhookID string) error {
	query := `
		UPDATE webhook_events
		SET status = $1
		WHERE webhook_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, entity.WebhookStatusFailed, webhookID)
	if err != nil {
		return fmt.Errorf("failed to mark webhook as failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrWebhookNotFound
	}

	return nil
}
