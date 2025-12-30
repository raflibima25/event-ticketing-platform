package entity

// TicketTier represents ticket tier data (read-only from event service)
type TicketTier struct {
	ID          string  `db:"id"`
	EventID     string  `db:"event_id"`
	Name        string  `db:"name"`
	Price       float64 `db:"price"`
	Quota       int     `db:"quota"`
	SoldCount   int     `db:"sold_count"`
	MaxPerOrder int     `db:"max_per_order"`
}

// GetAvailableQuota returns remaining ticket quota
func (tt *TicketTier) GetAvailableQuota() int {
	remaining := tt.Quota - tt.SoldCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsSoldOut checks if all tickets are sold
func (tt *TicketTier) IsSoldOut() bool {
	return tt.SoldCount >= tt.Quota
}

// CanPurchase checks if requested quantity can be purchased
func (tt *TicketTier) CanPurchase(quantity int) bool {
	// Check if quantity exceeds max per order
	if quantity > tt.MaxPerOrder {
		return false
	}

	// Check if enough quota available
	return tt.GetAvailableQuota() >= quantity
}

// GetPercentageSold returns percentage of tickets sold
func (tt *TicketTier) GetPercentageSold() float64 {
	if tt.Quota == 0 {
		return 0
	}
	return (float64(tt.SoldCount) / float64(tt.Quota)) * 100
}
