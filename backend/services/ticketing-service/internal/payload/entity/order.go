package entity

import "time"

// Order represents a ticket order
type Order struct {
	ID                   string
	UserID               string
	EventID              string
	TotalAmount          float64
	PlatformFee          float64
	ServiceFee           float64
	GrandTotal           float64
	Status               string // reserved, paid, expired, cancelled, completed
	PaymentID            *string
	PaymentMethod        *string
	ReservationExpiresAt *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
	CompletedAt          *time.Time
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
