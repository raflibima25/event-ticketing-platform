package entity

import "time"

// WebhookEvent represents a webhook event for idempotency tracking
type WebhookEvent struct {
	ID          string
	WebhookID   string // Unique ID from Xendit
	EventType   string // invoice.paid, invoice.expired, etc.
	Payload     string // JSONB - full webhook payload
	ProcessedAt *time.Time
	Status      string // pending, processed, failed
	CreatedAt   time.Time
}

// Webhook status constants
const (
	WebhookStatusPending   = "pending"
	WebhookStatusProcessed = "processed"
	WebhookStatusFailed    = "failed"
)

// Event type constants
const (
	EventTypeInvoicePaid    = "invoice.paid"
	EventTypeInvoiceExpired = "invoice.expired"
	EventTypeInvoiceFailed  = "invoice.failed"
)

// IsProcessed checks if webhook has been processed
func (w *WebhookEvent) IsProcessed() bool {
	return w.Status == WebhookStatusProcessed
}
