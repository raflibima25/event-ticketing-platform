package entity

import "time"

// Event represents event data from event service
type Event struct {
	ID          string    `db:"id"`
	Name        string    `db:"title"`
	Description string    `db:"description"`
	Location    string    `db:"location"`
	StartDate   time.Time `db:"start_date"`
	EndDate     time.Time `db:"end_date"`
	CategoryID  string    `db:"category"`
	OrganizerID string    `db:"organizer_id"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Event status constants
const (
	EventStatusDraft     = "draft"
	EventStatusPublished = "published"
	EventStatusCancelled = "cancelled"
	EventStatusCompleted = "completed"
)

// IsActive checks if event is currently active
func (e *Event) IsActive() bool {
	return e.Status == EventStatusPublished
}

// IsCancelled checks if event is cancelled
func (e *Event) IsCancelled() bool {
	return e.Status == EventStatusCancelled
}

// HasStarted checks if event has started
func (e *Event) HasStarted() bool {
	return time.Now().After(e.StartDate)
}

// HasEnded checks if event has ended
func (e *Event) HasEnded() bool {
	return time.Now().After(e.EndDate)
}
