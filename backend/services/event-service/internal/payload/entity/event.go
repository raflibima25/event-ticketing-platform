package entity

import "time"

// Event represents the event entity in database
type Event struct {
	ID          string    `json:"id" db:"id"`
	OrganizerID string    `json:"organizer_id" db:"organizer_id"`
	Title       string    `json:"title" db:"title"`
	Slug        string    `json:"slug" db:"slug"`
	Description *string   `json:"description,omitempty" db:"description"`
	Category    string    `json:"category" db:"category"`
	Location    string    `json:"location" db:"location"`
	Venue       *string   `json:"venue,omitempty" db:"venue"`
	StartDate   time.Time `json:"start_date" db:"start_date"`
	EndDate     time.Time `json:"end_date" db:"end_date"`
	Timezone    string    `json:"timezone" db:"timezone"`
	BannerURL   *string   `json:"banner_url,omitempty" db:"banner_url"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// EventStatus constants
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusCancelled = "cancelled"
)

// EventCategory constants
const (
	CategoryMusic      = "music"
	CategorySports     = "sports"
	CategoryArts       = "arts"
	CategoryTechnology = "technology"
	CategoryFood       = "food"
	CategoryBusiness   = "business"
	CategoryEducation  = "education"
	CategoryOther      = "other"
)

// IsValidStatus checks if status is valid
func IsValidStatus(status string) bool {
	switch status {
	case StatusDraft, StatusPublished, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsValidCategory checks if category is valid
func IsValidCategory(category string) bool {
	switch category {
	case CategoryMusic, CategorySports, CategoryArts, CategoryTechnology,
		CategoryFood, CategoryBusiness, CategoryEducation, CategoryOther:
		return true
	default:
		return false
	}
}
