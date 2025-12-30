package repository

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
)

var (
	ErrEventNotFound = errors.New("event not found")
)

// EventRepository defines interface for event data operations
type EventRepository interface {
	GetByID(ctx context.Context, id string) (*entity.Event, error)
}

// eventRepository implements EventRepository interface
type eventRepository struct {
	db *sqlx.DB
}

// NewEventRepository creates new event repository instance
func NewEventRepository(db *sqlx.DB) EventRepository {
	return &eventRepository{db: db}
}

// GetByID retrieves event by ID using sqlx
func (r *eventRepository) GetByID(ctx context.Context, id string) (*entity.Event, error) {
	var event entity.Event
	query := `
		SELECT id, title, description,
		       COALESCE(venue, location) as location,
		       start_date, end_date,
		       category, organizer_id, status, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &event, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	return &event, nil
}
