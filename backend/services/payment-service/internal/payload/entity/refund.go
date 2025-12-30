package entity

import "time"

// Refund represents a refund transaction
type Refund struct {
	ID                   string
	OrderID              string
	PaymentTransactionID string
	Amount               float64
	Reason               string
	Status               string // pending, processing, completed, failed
	DisbursementID       *string
	ProcessedAt          *time.Time
	CreatedAt            time.Time
}

// Refund status constants
const (
	RefundStatusPending    = "pending"
	RefundStatusProcessing = "processing"
	RefundStatusCompleted  = "completed"
	RefundStatusFailed     = "failed"
)

// IsCompleted checks if refund has been completed
func (r *Refund) IsCompleted() bool {
	return r.Status == RefundStatusCompleted
}
