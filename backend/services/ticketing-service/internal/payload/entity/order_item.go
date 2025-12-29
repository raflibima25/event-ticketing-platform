package entity

import "time"

// OrderItem represents an item in an order
type OrderItem struct {
	ID           string
	OrderID      string
	TicketTierID string
	Quantity     int
	Price        float64 // Price per ticket at time of purchase
	Subtotal     float64 // Price * Quantity
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CalculateSubtotal calculates subtotal for the order item
func (oi *OrderItem) CalculateSubtotal() float64 {
	return oi.Price * float64(oi.Quantity)
}
