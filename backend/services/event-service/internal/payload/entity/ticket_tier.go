package entity

import "time"

// TicketTier represents ticket tier entity in database
type TicketTier struct {
	ID               string     `json:"id" db:"id"`
	EventID          string     `json:"event_id" db:"event_id"`
	Name             string     `json:"name" db:"name"`
	Description      *string    `json:"description,omitempty" db:"description"`
	Price            float64    `json:"price" db:"price"`
	Quota            int        `json:"quota" db:"quota"`
	SoldCount        int        `json:"sold_count" db:"sold_count"`
	MaxPerOrder      int        `json:"max_per_order" db:"max_per_order"`
	EarlyBirdPrice   *float64   `json:"early_bird_price,omitempty" db:"early_bird_price"`
	EarlyBirdEndDate *time.Time `json:"early_bird_end_date,omitempty" db:"early_bird_end_date"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// AvailableCount returns available tickets
func (t *TicketTier) AvailableCount() int {
	return t.Quota - t.SoldCount
}

// IsAvailable checks if tickets are available
func (t *TicketTier) IsAvailable() bool {
	return t.AvailableCount() > 0
}

// CurrentPrice returns current price (early bird or regular)
func (t *TicketTier) CurrentPrice() float64 {
	if t.EarlyBirdPrice != nil && t.EarlyBirdEndDate != nil {
		if time.Now().Before(*t.EarlyBirdEndDate) {
			return *t.EarlyBirdPrice
		}
	}
	return t.Price
}

// IsSoldOut checks if tier is sold out
func (t *TicketTier) IsSoldOut() bool {
	return t.SoldCount >= t.Quota
}
