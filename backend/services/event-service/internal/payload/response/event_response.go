package response

import (
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/payload/entity"
)

// EventResponse represents event information in response
type EventResponse struct {
	ID          string               `json:"id"`
	OrganizerID string               `json:"organizer_id"`
	Title       string               `json:"title"`
	Slug        string               `json:"slug"`
	Description *string              `json:"description,omitempty"`
	Category    string               `json:"category"`
	Location    string               `json:"location"`
	Venue       *string              `json:"venue,omitempty"`
	StartDate   time.Time            `json:"start_date"`
	EndDate     time.Time            `json:"end_date"`
	Timezone    string               `json:"timezone"`
	BannerURL   *string              `json:"banner_url,omitempty"`
	Status      string               `json:"status"`
	TicketTiers []TicketTierResponse `json:"ticket_tiers,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// TicketTierResponse represents ticket tier information
type TicketTierResponse struct {
	ID               string     `json:"id"`
	EventID          string     `json:"event_id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description,omitempty"`
	Price            float64    `json:"price"`
	Quota            int        `json:"quota"`
	SoldCount        int        `json:"sold_count"`
	Available        int        `json:"available"` // Calculated field
	MaxPerOrder      int        `json:"max_per_order"`
	EarlyBirdPrice   *float64   `json:"early_bird_price,omitempty"`
	EarlyBirdEndDate *time.Time `json:"early_bird_end_date,omitempty"`
	CurrentPrice     float64    `json:"current_price"` // Calculated field
	IsSoldOut        bool       `json:"is_sold_out"`   // Calculated field
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// PaginatedEventsResponse represents paginated events response
type PaginatedEventsResponse struct {
	Events []EventResponse `json:"events"`
	Meta   PaginationMeta  `json:"meta"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
}

// ToEventResponse converts Event entity to EventResponse
func ToEventResponse(event *entity.Event, tiers []entity.TicketTier) *EventResponse {
	response := &EventResponse{
		ID:          event.ID,
		OrganizerID: event.OrganizerID,
		Title:       event.Title,
		Slug:        event.Slug,
		Description: event.Description,
		Category:    event.Category,
		Location:    event.Location,
		Venue:       event.Venue,
		StartDate:   event.StartDate,
		EndDate:     event.EndDate,
		Timezone:    event.Timezone,
		BannerURL:   event.BannerURL,
		Status:      event.Status,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}

	// Convert ticket tiers if provided
	if tiers != nil {
		tierResponses := make([]TicketTierResponse, 0, len(tiers))
		for _, tier := range tiers {
			tierResponses = append(tierResponses, *ToTicketTierResponse(&tier))
		}
		response.TicketTiers = tierResponses
	}

	return response
}

// ToTicketTierResponse converts TicketTier entity to TicketTierResponse
func ToTicketTierResponse(tier *entity.TicketTier) *TicketTierResponse {
	available := tier.Quota - tier.SoldCount
	currentPrice := tier.CurrentPrice()
	isSoldOut := tier.SoldCount >= tier.Quota

	return &TicketTierResponse{
		ID:               tier.ID,
		EventID:          tier.EventID,
		Name:             tier.Name,
		Description:      tier.Description,
		Price:            tier.Price,
		Quota:            tier.Quota,
		SoldCount:        tier.SoldCount,
		Available:        available,
		MaxPerOrder:      tier.MaxPerOrder,
		EarlyBirdPrice:   tier.EarlyBirdPrice,
		EarlyBirdEndDate: tier.EarlyBirdEndDate,
		CurrentPrice:     currentPrice,
		IsSoldOut:        isSoldOut,
		CreatedAt:        tier.CreatedAt,
		UpdatedAt:        tier.UpdatedAt,
	}
}
