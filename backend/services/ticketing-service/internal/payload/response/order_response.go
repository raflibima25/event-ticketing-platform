package response

import (
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
)

// OrderResponse represents order information in response
type OrderResponse struct {
	ID                   string              `json:"id"`
	UserID               string              `json:"user_id"`
	EventID              string              `json:"event_id"`
	Items                []OrderItemResponse `json:"items"`
	TotalAmount          float64             `json:"total_amount"`
	PlatformFee          float64             `json:"platform_fee"`
	ServiceFee           float64             `json:"service_fee"`
	GrandTotal           float64             `json:"grand_total"`
	Status               string              `json:"status"`
	PaymentID            *string             `json:"payment_id,omitempty"`
	PaymentMethod        *string             `json:"payment_method,omitempty"`
	ReservationExpiresAt *time.Time          `json:"reservation_expires_at,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
	CompletedAt          *time.Time          `json:"completed_at,omitempty"`
}

// OrderItemResponse represents order item in response
type OrderItemResponse struct {
	ID           string  `json:"id"`
	TicketTierID string  `json:"ticket_tier_id"`
	TierName     string  `json:"tier_name,omitempty"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	Subtotal     float64 `json:"subtotal"`
}

// TicketResponse represents ticket information
type TicketResponse struct {
	ID           string     `json:"id"`
	OrderID      string     `json:"order_id"`
	TicketTierID string     `json:"ticket_tier_id"`
	EventID      string     `json:"event_id"`
	TicketNumber string     `json:"ticket_number"`
	QRCode       string     `json:"qr_code"` // Base64 encoded
	Status       string     `json:"status"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// AvailabilityResponse represents ticket availability info
type AvailabilityResponse struct {
	TicketTierID string `json:"ticket_tier_id"`
	TierName     string `json:"tier_name"`
	Quota        int    `json:"quota"`
	SoldCount    int    `json:"sold_count"`
	Available    int    `json:"available"`
	IsAvailable  bool   `json:"is_available"`
	MaxPerOrder  int    `json:"max_per_order"`
}

// ToOrderResponse converts Order entity to OrderResponse
func ToOrderResponse(order *entity.Order, items []entity.OrderItem) *OrderResponse {
	itemResponses := make([]OrderItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, OrderItemResponse{
			ID:           item.ID,
			TicketTierID: item.TicketTierID,
			Quantity:     item.Quantity,
			Price:        item.Price,
			Subtotal:     item.Subtotal,
		})
	}

	return &OrderResponse{
		ID:                   order.ID,
		UserID:               order.UserID,
		EventID:              order.EventID,
		Items:                itemResponses,
		TotalAmount:          order.TotalAmount,
		PlatformFee:          order.PlatformFee,
		ServiceFee:           order.ServiceFee,
		GrandTotal:           order.GrandTotal,
		Status:               order.Status,
		PaymentID:            order.PaymentID,
		PaymentMethod:        order.PaymentMethod,
		ReservationExpiresAt: order.ReservationExpiresAt,
		CreatedAt:            order.CreatedAt,
		UpdatedAt:            order.UpdatedAt,
		CompletedAt:          order.CompletedAt,
	}
}

// ToTicketResponse converts Ticket entity to TicketResponse
func ToTicketResponse(ticket *entity.Ticket) *TicketResponse {
	return &TicketResponse{
		ID:           ticket.ID,
		OrderID:      ticket.OrderID,
		TicketTierID: ticket.TicketTierID,
		EventID:      ticket.EventID,
		TicketNumber: ticket.TicketNumber,
		QRCode:       ticket.QRCode,
		Status:       ticket.Status,
		UsedAt:       ticket.UsedAt,
		CreatedAt:    ticket.CreatedAt,
	}
}
