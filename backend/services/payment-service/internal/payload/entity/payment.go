package entity

import "time"

// PaymentTransaction represents a payment transaction record
type PaymentTransaction struct {
	ID            string
	OrderID       string
	ExternalID    string // ORDER-{order_id}
	InvoiceID     *string
	InvoiceURL    *string
	Amount        float64
	PaymentMethod *string
	Status        string // pending, paid, expired, failed
	PaidAt        *time.Time
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Payment status constants
const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusExpired = "expired"
	PaymentStatusFailed  = "failed"
)

// IsPaid checks if payment has been completed
func (p *PaymentTransaction) IsPaid() bool {
	return p.Status == PaymentStatusPaid
}

// IsExpired checks if payment has expired
func (p *PaymentTransaction) IsExpired() bool {
	if p.Status != PaymentStatusPending {
		return false
	}
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}
