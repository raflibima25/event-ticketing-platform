package entity

import "time"

// OrderItem represents an item in an order
type OrderItem struct {
	ID           string    `db:"id"`
	OrderID      string    `db:"order_id"`
	TicketTierID string    `db:"ticket_tier_id"`
	Quantity     int       `db:"quantity"`
	Price        float64   `db:"price"` // Price per ticket at time of purchase
	Subtotal     float64   `db:"subtotal"` // Price * Quantity
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// CalculateSubtotal calculates subtotal for the order item
func (oi *OrderItem) CalculateSubtotal() float64 {
	return oi.Price * float64(oi.Quantity)
}
