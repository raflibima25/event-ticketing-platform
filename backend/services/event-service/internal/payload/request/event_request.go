package request

import "time"

// CreateEventRequest represents create event request
type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required,min=3,max=255"`
	Description string    `json:"description"`
	Category    string    `json:"category" binding:"required,oneof=music sports arts technology food business education other"`
	Location    string    `json:"location" binding:"required"`
	Venue       string    `json:"venue"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required,gtfield=StartDate"`
	Timezone    string    `json:"timezone" binding:"required"`
	BannerURL   string    `json:"banner_url"`
	Status      string    `json:"status" binding:"omitempty,oneof=draft published"`
}

// UpdateEventRequest represents update event request
type UpdateEventRequest struct {
	Title       string    `json:"title" binding:"omitempty,min=3,max=255"`
	Description string    `json:"description"`
	Category    string    `json:"category" binding:"omitempty,oneof=music sports arts technology food business education other"`
	Location    string    `json:"location"`
	Venue       string    `json:"venue"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Timezone    string    `json:"timezone"`
	BannerURL   string    `json:"banner_url"`
	Status      string    `json:"status" binding:"omitempty,oneof=draft published cancelled"`
}

// ListEventsRequest represents list events with filters
type ListEventsRequest struct {
	Category  string    `form:"category"`
	Location  string    `form:"location"`
	StartDate time.Time `form:"start_date"`
	EndDate   time.Time `form:"end_date"`
	Status    string    `form:"status" binding:"omitempty,oneof=draft published cancelled"`
	Search    string    `form:"search"`
	Page      int       `form:"page" binding:"omitempty,min=1"`
	Limit     int       `form:"limit" binding:"omitempty,min=1,max=100"`
	SortBy    string    `form:"sort_by" binding:"omitempty,oneof=start_date created_at title"`
	SortOrder string    `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// CreateTicketTierRequest represents create ticket tier request
type CreateTicketTierRequest struct {
	EventID          string     `json:"event_id" binding:"required,uuid"`
	Name             string     `json:"name" binding:"required,min=3,max=100"`
	Description      string     `json:"description"`
	Price            float64    `json:"price" binding:"required,min=0"`
	Quota            int        `json:"quota" binding:"required,min=1"`
	MaxPerOrder      int        `json:"max_per_order" binding:"omitempty,min=1"`
	EarlyBirdPrice   *float64   `json:"early_bird_price" binding:"omitempty,min=0"`
	EarlyBirdEndDate *time.Time `json:"early_bird_end_date"`
}

// UpdateTicketTierRequest represents update ticket tier request
type UpdateTicketTierRequest struct {
	Name             string     `json:"name" binding:"omitempty,min=3,max=100"`
	Description      string     `json:"description"`
	Price            float64    `json:"price" binding:"omitempty,min=0"`
	Quota            int        `json:"quota" binding:"omitempty,min=1"`
	MaxPerOrder      int        `json:"max_per_order" binding:"omitempty,min=1"`
	EarlyBirdPrice   *float64   `json:"early_bird_price" binding:"omitempty,min=0"`
	EarlyBirdEndDate *time.Time `json:"early_bird_end_date"`
}

// Validate validates CreateTicketTierRequest business rules
func (r *CreateTicketTierRequest) Validate() error {
	// If early bird price is set, early bird end date must be set
	if r.EarlyBirdPrice != nil && r.EarlyBirdEndDate == nil {
		return ErrInvalidEarlyBirdSettings
	}

	// Early bird price must be less than regular price
	if r.EarlyBirdPrice != nil && *r.EarlyBirdPrice >= r.Price {
		return ErrInvalidEarlyBirdPrice
	}

	// Early bird end date must be in the future
	if r.EarlyBirdEndDate != nil && r.EarlyBirdEndDate.Before(time.Now()) {
		return ErrInvalidEarlyBirdEndDate
	}

	return nil
}

// Validate validates UpdateTicketTierRequest business rules
func (r *UpdateTicketTierRequest) Validate() error {
	// If early bird price is set, early bird end date must be set
	if r.EarlyBirdPrice != nil && r.EarlyBirdEndDate == nil {
		return ErrInvalidEarlyBirdSettings
	}

	// Early bird price must be less than regular price
	if r.EarlyBirdPrice != nil && *r.EarlyBirdPrice >= r.Price {
		return ErrInvalidEarlyBirdPrice
	}

	return nil
}
