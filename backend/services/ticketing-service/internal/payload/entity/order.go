package entity

import "time"

// Order represents a ticket order
type Order struct {
	ID                   string     `db:"id"`
	UserID               string     `db:"user_id"`
	EventID              string     `db:"event_id"`
	TotalAmount          float64    `db:"total_amount"`
	PlatformFee          float64    `db:"platform_fee"`
	ServiceFee           float64    `db:"service_fee"`
	GrandTotal           float64    `db:"grand_total"`
	Status               string     `db:"status"` // reserved, paid, expired, cancelled, completed
	PaymentID            *string    `db:"payment_id"`
	PaymentMethod        *string    `db:"payment_method"`
	ReservationExpiresAt *time.Time `db:"reservation_expires_at"`
	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
	CompletedAt          *time.Time `db:"completed_at"`
}

// Order status constants
const (
	OrderStatusReserved  = "reserved"  // Tickets reserved, waiting for payment
	OrderStatusPaid      = "paid"      // Payment received, tickets issued
	OrderStatusExpired   = "expired"   // Reservation timeout reached
	OrderStatusCancelled = "cancelled" // Manually cancelled by user
	OrderStatusCompleted = "completed" // Event finished, tickets used
)

// IsExpired checks if order reservation has expired
func (o *Order) IsExpired() bool {
	if o.Status != OrderStatusReserved {
		return false
	}
	if o.ReservationExpiresAt == nil {
		return false
	}
	return time.Now().After(*o.ReservationExpiresAt)
}

// CanBeCancelled checks if order can be cancelled
// Only reserved orders can be cancelled
// Paid orders require refund process (handled separately)
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusReserved
}

// IsPaid checks if order has been paid
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid || o.Status == OrderStatusCompleted
}
