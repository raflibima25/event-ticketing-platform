package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
)

// Event represents event data from database
type Event struct {
	ID          string
	Name        string
	Description string
	Location    string
	StartDate   time.Time
	EndDate     time.Time
	CategoryID  string
	OrganizerID string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EventRepository defines interface for event data operations
type EventRepository interface {
	GetByID(ctx context.Context, id string) (*Event, error)
}

// eventRepository implements EventRepository interface
type eventRepository struct {
	db *sql.DB
}

// NewEventRepository creates new event repository instance
func NewEventRepository(db *sql.DB) EventRepository {
	return &eventRepository{db: db}
}

// GetByID retrieves event by ID
func (r *eventRepository) GetByID(ctx context.Context, id string) (*Event, error) {
	query := `
		SELECT id, title, description,
		       COALESCE(venue, location) as location,
		       start_date, end_date,
		       category, organizer_id, status, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	event := &Event{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Name,
		&event.Description,
		&event.Location,
		&event.StartDate,
		&event.EndDate,
		&event.CategoryID,
		&event.OrganizerID,
		&event.Status,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrEventNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}
