package entity

import "time"

// Ticket represents an e-ticket
type Ticket struct {
	ID           string     `db:"id"`
	OrderID      string     `db:"order_id"`
	OrderItemID  string     `db:"order_item_id"`
	TicketTierID string     `db:"ticket_tier_id"`
	EventID      string     `db:"event_id"`
	UserID       string     `db:"user_id"`
	TicketNumber string     `db:"ticket_number"` // Unique ticket number (for display)
	QRCode       string     `db:"qr_code"` // Base64 encoded QR code
	QRData       string     `db:"qr_data"` // Data encoded in QR (for validation)
	Status       string     `db:"status"` // valid, used, cancelled, expired
	UsedAt       *time.Time `db:"validated_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// Ticket status constants
const (
	TicketStatusValid     = "valid"     // Ticket is valid and can be used
	TicketStatusUsed      = "used"      // Ticket has been scanned and used
	TicketStatusCancelled = "cancelled" // Ticket cancelled (refund)
	TicketStatusExpired   = "expired"   // Event has passed
)

// CanBeUsed checks if ticket can be used (scanned at event)
func (t *Ticket) CanBeUsed() bool {
	return t.Status == TicketStatusValid
}

// IsUsed checks if ticket has been used
func (t *Ticket) IsUsed() bool {
	return t.Status == TicketStatusUsed
}
