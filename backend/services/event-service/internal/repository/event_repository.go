package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/payload/request"
)

var (
	ErrEventNotFound   = errors.New("event not found")
	ErrEventSlugExists = errors.New("event slug already exists")
)

// EventRepository defines interface for event data operations
type EventRepository interface {
	Create(ctx context.Context, event *entity.Event) error
	GetByID(ctx context.Context, id string) (*entity.Event, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Event, error)
	List(ctx context.Context, filters request.ListEventsRequest) ([]entity.Event, int64, error)
	Update(ctx context.Context, event *entity.Event) error
	Delete(ctx context.Context, id string) error
	GetByOrganizerID(ctx context.Context, organizerID string) ([]entity.Event, error)
}

// eventRepository implements EventRepository interface
type eventRepository struct {
	db *sql.DB
}

// NewEventRepository creates new event repository instance
func NewEventRepository(db *sql.DB) EventRepository {
	return &eventRepository{db: db}
}

// Create inserts new event into database
func (r *eventRepository) Create(ctx context.Context, event *entity.Event) error {
	query := `
		INSERT INTO events (id, organizer_id, title, slug, description, category, location, venue,
		                   start_date, end_date, timezone, banner_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	event.ID = uuid.New().String()

	err := r.db.QueryRowContext(
		ctx,
		query,
		event.ID,
		event.OrganizerID,
		event.Title,
		event.Slug,
		event.Description,
		event.Category,
		event.Location,
		event.Venue,
		event.StartDate,
		event.EndDate,
		event.Timezone,
		event.BannerURL,
		event.Status,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "events_slug_key") {
			return ErrEventSlugExists
		}
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetByID retrieves event by ID
func (r *eventRepository) GetByID(ctx context.Context, id string) (*entity.Event, error) {
	query := `
		SELECT id, organizer_id, title, slug, description, category, location, venue,
		       start_date, end_date, timezone, banner_url, status, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	event := &entity.Event{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.OrganizerID,
		&event.Title,
		&event.Slug,
		&event.Description,
		&event.Category,
		&event.Location,
		&event.Venue,
		&event.StartDate,
		&event.EndDate,
		&event.Timezone,
		&event.BannerURL,
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

// GetBySlug retrieves event by slug
func (r *eventRepository) GetBySlug(ctx context.Context, slug string) (*entity.Event, error) {
	query := `
		SELECT id, organizer_id, title, slug, description, category, location, venue,
		       start_date, end_date, timezone, banner_url, status, created_at, updated_at
		FROM events
		WHERE slug = $1
	`

	event := &entity.Event{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&event.ID,
		&event.OrganizerID,
		&event.Title,
		&event.Slug,
		&event.Description,
		&event.Category,
		&event.Location,
		&event.Venue,
		&event.StartDate,
		&event.EndDate,
		&event.Timezone,
		&event.BannerURL,
		&event.Status,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrEventNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get event by slug: %w", err)
	}

	return event, nil
}

// List retrieves events with filters and pagination
func (r *eventRepository) List(ctx context.Context, filters request.ListEventsRequest) ([]entity.Event, int64, error) {
	// Build WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argCount := 1

	if filters.Category != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("category = $%d", argCount))
		args = append(args, filters.Category)
		argCount++
	}

	if filters.Location != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("location ILIKE $%d", argCount))
		args = append(args, "%"+filters.Location+"%")
		argCount++
	}

	if filters.Status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, filters.Status)
		argCount++
	} else {
		// Default: only show published events if no status filter
		whereConditions = append(whereConditions, "status = 'published'")
	}

	if !filters.StartDate.IsZero() {
		whereConditions = append(whereConditions, fmt.Sprintf("start_date >= $%d", argCount))
		args = append(args, filters.StartDate)
		argCount++
	}

	if !filters.EndDate.IsZero() {
		whereConditions = append(whereConditions, fmt.Sprintf("end_date <= $%d", argCount))
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+filters.Search+"%")
		argCount++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Build ORDER BY clause
	sortBy := "start_date"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}

	sortOrder := "ASC"
	if filters.SortOrder != "" {
		sortOrder = strings.ToUpper(filters.SortOrder)
	}

	orderClause := fmt.Sprintf("ORDER BY %s %s", sortBy, sortOrder)

	// Pagination
	page := 1
	if filters.Page > 0 {
		page = filters.Page
	}

	limit := 10
	if filters.Limit > 0 {
		limit = filters.Limit
	}

	offset := (page - 1) * limit

	// Build final query
	query := fmt.Sprintf(`
		SELECT id, organizer_id, title, slug, description, category, location, venue,
		       start_date, end_date, timezone, banner_url, status, created_at, updated_at
		FROM events
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argCount, argCount+1)

	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	events := []entity.Event{}
	for rows.Next() {
		var event entity.Event
		err := rows.Scan(
			&event.ID,
			&event.OrganizerID,
			&event.Title,
			&event.Slug,
			&event.Description,
			&event.Category,
			&event.Location,
			&event.Venue,
			&event.StartDate,
			&event.EndDate,
			&event.Timezone,
			&event.BannerURL,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, total, nil
}

// Update updates event information
func (r *eventRepository) Update(ctx context.Context, event *entity.Event) error {
	query := `
		UPDATE events
		SET title = $1, description = $2, category = $3, location = $4, venue = $5,
		    start_date = $6, end_date = $7, timezone = $8, banner_url = $9, status = $10,
		    updated_at = NOW()
		WHERE id = $11
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		event.Title,
		event.Description,
		event.Category,
		event.Location,
		event.Venue,
		event.StartDate,
		event.EndDate,
		event.Timezone,
		event.BannerURL,
		event.Status,
		event.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrEventNotFound
	}

	return nil
}

// Delete soft deletes event
func (r *eventRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrEventNotFound
	}

	return nil
}

// GetByOrganizerID retrieves all events by organizer
func (r *eventRepository) GetByOrganizerID(ctx context.Context, organizerID string) ([]entity.Event, error) {
	query := `
		SELECT id, organizer_id, title, slug, description, category, location, venue,
		       start_date, end_date, timezone, banner_url, status, created_at, updated_at
		FROM events
		WHERE organizer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, organizerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by organizer: %w", err)
	}
	defer rows.Close()

	events := []entity.Event{}
	for rows.Next() {
		var event entity.Event
		err := rows.Scan(
			&event.ID,
			&event.OrganizerID,
			&event.Title,
			&event.Slug,
			&event.Description,
			&event.Category,
			&event.Location,
			&event.Venue,
			&event.StartDate,
			&event.EndDate,
			&event.Timezone,
			&event.BannerURL,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}
