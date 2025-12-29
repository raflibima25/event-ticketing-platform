package entity

import "time"

// Ticket represents an e-ticket
type Ticket struct {
	ID           string
	OrderID      string
	OrderItemID  string
	TicketTierID string
	EventID      string
	UserID       string
	TicketNumber string // Unique ticket number (for display)
	QRCode       string // Base64 encoded QR code
	QRData       string // Data encoded in QR (for validation)
	Status       string // valid, used, cancelled, expired
	UsedAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
