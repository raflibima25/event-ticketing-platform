package request

// CreateOrderRequest represents create order from cart or direct purchase
type CreateOrderRequest struct {
	EventID       string      `json:"event_id" binding:"required,uuid"`
	Items         []OrderItem `json:"items" binding:"required,min=1,dive"`
	Email         string      `json:"email,omitempty"`          // Optional - will use user profile if not provided
	CustomerName  string      `json:"customer_name,omitempty"`  // Optional - will use user profile if not provided
	PaymentMethod string      `json:"payment_method,omitempty"` // Will be set later before payment
}

// OrderItem represents an item to order
type OrderItem struct {
	TicketTierID string `json:"ticket_tier_id" binding:"required,uuid"`
	Quantity     int    `json:"quantity" binding:"required,min=1"`
}

// ConfirmOrderRequest represents payment confirmation (from webhook)
type ConfirmOrderRequest struct {
	OrderID       string  `json:"order_id"` // Set from URL path parameter, not required in body
	PaymentID     string  `json:"payment_id" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,min=0"`
}

// CancelOrderRequest represents order cancellation
type CancelOrderRequest struct {
	Reason string `json:"reason"`
}

// ValidateTicketRequest represents ticket validation at event entrance
type ValidateTicketRequest struct {
	QRData string `json:"qr_data" binding:"required"`
}
