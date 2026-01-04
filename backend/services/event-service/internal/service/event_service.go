package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/pkg/cache"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/payload/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/utility"
)

var (
	ErrUnauthorized        = errors.New("unauthorized to perform this action")
	ErrEventNotFound       = errors.New("event not found")
	ErrTicketTierNotFound  = errors.New("ticket tier not found")
	ErrInvalidDateRange    = errors.New("end date must be after start date")
	ErrCannotUpdateSlug    = errors.New("slug cannot be updated")
	ErrQuotaBelowSoldCount = errors.New("quota cannot be less than sold count")
)

// Cache TTL constants
const (
	cacheEventDetailTTL    = 5 * time.Minute  // Event detail cache TTL
	cacheTicketTiersTTL    = 30 * time.Second // Ticket tiers cache TTL (shorter because quota changes)
	cacheEventListingTTL   = 5 * time.Minute  // Event listing cache TTL
)

// EventService defines interface for event business logic
type EventService interface {
	// Event operations
	CreateEvent(ctx context.Context, organizerID string, req *request.CreateEventRequest) (*response.EventResponse, error)
	GetEventByID(ctx context.Context, id string) (*response.EventResponse, error)
	GetEventBySlug(ctx context.Context, slug string) (*response.EventResponse, error)
	ListEvents(ctx context.Context, filters request.ListEventsRequest) (*response.PaginatedEventsResponse, error)
	UpdateEvent(ctx context.Context, organizerID string, eventID string, req *request.UpdateEventRequest) (*response.EventResponse, error)
	DeleteEvent(ctx context.Context, organizerID string, eventID string) error
	GetOrganizerEvents(ctx context.Context, organizerID string) ([]response.EventResponse, error)

	// Ticket tier operations
	CreateTicketTier(ctx context.Context, organizerID string, req *request.CreateTicketTierRequest) (*response.TicketTierResponse, error)
	GetTicketTierByID(ctx context.Context, id string) (*response.TicketTierResponse, error)
	GetTicketTiersByEventID(ctx context.Context, eventID string) ([]response.TicketTierResponse, error)
	UpdateTicketTier(ctx context.Context, organizerID string, tierID string, req *request.UpdateTicketTierRequest) (*response.TicketTierResponse, error)
	DeleteTicketTier(ctx context.Context, organizerID string, tierID string) error
}

// eventService implements EventService interface
type eventService struct {
	eventRepo      repository.EventRepository
	ticketTierRepo repository.TicketTierRepository
	cache          cache.RedisClient
}

// NewEventService creates new event service instance
func NewEventService(
	eventRepo repository.EventRepository,
	ticketTierRepo repository.TicketTierRepository,
	redisClient cache.RedisClient,
) EventService {
	return &eventService{
		eventRepo:      eventRepo,
		ticketTierRepo: ticketTierRepo,
		cache:          redisClient,
	}
}

// CreateEvent creates new event
func (s *eventService) CreateEvent(ctx context.Context, organizerID string, req *request.CreateEventRequest) (*response.EventResponse, error) {
	// Validate date range
	if !req.EndDate.After(req.StartDate) {
		return nil, ErrInvalidDateRange
	}

	// Generate slug
	slug := utility.GenerateSlug(req.Title)

	// Create event entity
	event := &entity.Event{
		OrganizerID: organizerID,
		Title:       req.Title,
		Slug:        slug,
		Description: &req.Description,
		Category:    req.Category,
		Location:    req.Location,
		Venue:       &req.Venue,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Timezone:    req.Timezone,
		BannerURL:   &req.BannerURL,
		Status:      req.Status,
	}

	// Set default status if not provided
	if event.Status == "" {
		event.Status = "draft"
	}

	// Create event in repository
	if err := s.eventRepo.Create(ctx, event); err != nil {
		if errors.Is(err, repository.ErrEventSlugExists) {
			return nil, fmt.Errorf("slug already exists, please try again")
		}
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return response.ToEventResponse(event, nil), nil
}

// GetEventByID retrieves event by ID with caching
func (s *eventService) GetEventByID(ctx context.Context, id string) (*response.EventResponse, error) {
	cacheKey := fmt.Sprintf("event:id:%s", id)

	// Try to get from cache first
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var eventResp response.EventResponse
			if err := json.Unmarshal([]byte(cached), &eventResp); err == nil {
				return &eventResp, nil
			}
			// If unmarshal fails, continue to database
		}
	}

	// Cache miss or error - get from database
	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Get ticket tiers for this event
	tiers, err := s.ticketTierRepo.GetByEventID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tiers: %w", err)
	}

	eventResp := response.ToEventResponse(event, tiers)

	// Store in cache for next time
	if s.cache != nil {
		if data, err := json.Marshal(eventResp); err == nil {
			s.cache.Set(ctx, cacheKey, string(data), cacheEventDetailTTL)
		}
	}

	return eventResp, nil
}

// GetEventBySlug retrieves event by slug with caching
func (s *eventService) GetEventBySlug(ctx context.Context, slug string) (*response.EventResponse, error) {
	cacheKey := fmt.Sprintf("event:slug:%s", slug)

	// Try to get from cache first
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var eventResp response.EventResponse
			if err := json.Unmarshal([]byte(cached), &eventResp); err == nil {
				return &eventResp, nil
			}
		}
	}

	// Cache miss - get from database
	event, err := s.eventRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Get ticket tiers for this event
	tiers, err := s.ticketTierRepo.GetByEventID(ctx, event.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tiers: %w", err)
	}

	eventResp := response.ToEventResponse(event, tiers)

	// Store in cache
	if s.cache != nil {
		if data, err := json.Marshal(eventResp); err == nil {
			s.cache.Set(ctx, cacheKey, string(data), cacheEventDetailTTL)
		}
	}

	return eventResp, nil
}

// ListEvents retrieves events with filters and pagination
func (s *eventService) ListEvents(ctx context.Context, filters request.ListEventsRequest) (*response.PaginatedEventsResponse, error) {
	events, total, err := s.eventRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Convert to response
	eventResponses := make([]response.EventResponse, 0, len(events))
	for _, event := range events {
		eventResponses = append(eventResponses, *response.ToEventResponse(&event, nil))
	}

	// Calculate pagination metadata
	page := 1
	if filters.Page > 0 {
		page = filters.Page
	}

	limit := 10
	if filters.Limit > 0 {
		limit = filters.Limit
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &response.PaginatedEventsResponse{
		Events: eventResponses,
		Pagination: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       total,
			TotalPages:  totalPages,
		},
	}, nil
}

// UpdateEvent updates event information
func (s *eventService) UpdateEvent(ctx context.Context, organizerID string, eventID string, req *request.UpdateEventRequest) (*response.EventResponse, error) {
	// Get existing event
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Check authorization
	if event.OrganizerID != organizerID {
		return nil, ErrUnauthorized
	}

	// Update fields if provided
	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != "" {
		event.Description = &req.Description
	}
	if req.Category != "" {
		event.Category = req.Category
	}
	if req.Location != "" {
		event.Location = req.Location
	}
	if req.Venue != "" {
		event.Venue = &req.Venue
	}
	if !req.StartDate.IsZero() {
		event.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		event.EndDate = req.EndDate
	}
	if req.Timezone != "" {
		event.Timezone = req.Timezone
	}
	if req.BannerURL != "" {
		event.BannerURL = &req.BannerURL
	}
	if req.Status != "" {
		event.Status = req.Status
	}

	// Validate date range
	if !event.EndDate.After(event.StartDate) {
		return nil, ErrInvalidDateRange
	}

	// Update in repository
	if err := s.eventRepo.Update(ctx, event); err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	// Invalidate cache (both ID and slug keys)
	if s.cache != nil {
		s.cache.Del(ctx, fmt.Sprintf("event:id:%s", eventID))
		s.cache.Del(ctx, fmt.Sprintf("event:slug:%s", event.Slug))
	}

	// Get ticket tiers
	tiers, err := s.ticketTierRepo.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tiers: %w", err)
	}

	return response.ToEventResponse(event, tiers), nil
}

// DeleteEvent deletes event
func (s *eventService) DeleteEvent(ctx context.Context, organizerID string, eventID string) error {
	// Get existing event
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return ErrEventNotFound
		}
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Check authorization
	if event.OrganizerID != organizerID {
		return ErrUnauthorized
	}

	// Delete event
	if err := s.eventRepo.Delete(ctx, eventID); err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return ErrEventNotFound
		}
		return fmt.Errorf("failed to delete event: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.Del(ctx, fmt.Sprintf("event:id:%s", eventID))
		s.cache.Del(ctx, fmt.Sprintf("event:slug:%s", event.Slug))
	}

	return nil
}

// GetOrganizerEvents retrieves all events for an organizer
func (s *eventService) GetOrganizerEvents(ctx context.Context, organizerID string) ([]response.EventResponse, error) {
	events, err := s.eventRepo.GetByOrganizerID(ctx, organizerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizer events: %w", err)
	}

	// Convert to response
	eventResponses := make([]response.EventResponse, 0, len(events))
	for _, event := range events {
		eventResponses = append(eventResponses, *response.ToEventResponse(&event, nil))
	}

	return eventResponses, nil
}

// CreateTicketTier creates new ticket tier for an event
func (s *eventService) CreateTicketTier(ctx context.Context, organizerID string, req *request.CreateTicketTierRequest) (*response.TicketTierResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if event exists and user is the organizer
	event, err := s.eventRepo.GetByID(ctx, req.EventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if event.OrganizerID != organizerID {
		return nil, ErrUnauthorized
	}

	// Create ticket tier entity
	tier := &entity.TicketTier{
		EventID:          req.EventID,
		Name:             req.Name,
		Description:      &req.Description,
		Price:            req.Price,
		Quota:            req.Quota,
		MaxPerOrder:      req.MaxPerOrder,
		EarlyBirdPrice:   req.EarlyBirdPrice,
		EarlyBirdEndDate: req.EarlyBirdEndDate,
	}

	// Create in repository
	if err := s.ticketTierRepo.Create(ctx, tier); err != nil {
		return nil, fmt.Errorf("failed to create ticket tier: %w", err)
	}

	return response.ToTicketTierResponse(tier), nil
}

// GetTicketTierByID retrieves ticket tier by ID
func (s *eventService) GetTicketTierByID(ctx context.Context, id string) (*response.TicketTierResponse, error) {
	tier, err := s.ticketTierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTicketTierNotFound) {
			return nil, ErrTicketTierNotFound
		}
		return nil, fmt.Errorf("failed to get ticket tier: %w", err)
	}

	return response.ToTicketTierResponse(tier), nil
}

// GetTicketTiersByEventID retrieves all ticket tiers for an event
func (s *eventService) GetTicketTiersByEventID(ctx context.Context, eventID string) ([]response.TicketTierResponse, error) {
	tiers, err := s.ticketTierRepo.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket tiers: %w", err)
	}

	// Convert to response
	tierResponses := make([]response.TicketTierResponse, 0, len(tiers))
	for _, tier := range tiers {
		tierResponses = append(tierResponses, *response.ToTicketTierResponse(&tier))
	}

	return tierResponses, nil
}

// UpdateTicketTier updates ticket tier information
func (s *eventService) UpdateTicketTier(ctx context.Context, organizerID string, tierID string, req *request.UpdateTicketTierRequest) (*response.TicketTierResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get existing ticket tier
	tier, err := s.ticketTierRepo.GetByID(ctx, tierID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketTierNotFound) {
			return nil, ErrTicketTierNotFound
		}
		return nil, fmt.Errorf("failed to get ticket tier: %w", err)
	}

	// Check if user is the event organizer
	event, err := s.eventRepo.GetByID(ctx, tier.EventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if event.OrganizerID != organizerID {
		return nil, ErrUnauthorized
	}

	// Validate quota is not less than sold count
	if req.Quota < tier.SoldCount {
		return nil, ErrQuotaBelowSoldCount
	}

	// Update fields
	tier.Name = req.Name
	tier.Description = &req.Description
	tier.Price = req.Price
	tier.Quota = req.Quota
	tier.MaxPerOrder = req.MaxPerOrder
	tier.EarlyBirdPrice = req.EarlyBirdPrice
	tier.EarlyBirdEndDate = req.EarlyBirdEndDate

	// Update in repository
	if err := s.ticketTierRepo.Update(ctx, tier); err != nil {
		if errors.Is(err, repository.ErrTicketTierNotFound) {
			return nil, ErrTicketTierNotFound
		}
		return nil, fmt.Errorf("failed to update ticket tier: %w", err)
	}

	return response.ToTicketTierResponse(tier), nil
}

// DeleteTicketTier deletes ticket tier
func (s *eventService) DeleteTicketTier(ctx context.Context, organizerID string, tierID string) error {
	// Get existing ticket tier
	tier, err := s.ticketTierRepo.GetByID(ctx, tierID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketTierNotFound) {
			return ErrTicketTierNotFound
		}
		return fmt.Errorf("failed to get ticket tier: %w", err)
	}

	// Check if user is the event organizer
	event, err := s.eventRepo.GetByID(ctx, tier.EventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return ErrEventNotFound
		}
		return fmt.Errorf("failed to get event: %w", err)
	}

	if event.OrganizerID != organizerID {
		return ErrUnauthorized
	}

	// TODO: Check if there are existing orders for this ticket tier
	// For now, we'll allow deletion but in production this should be prevented

	// Delete ticket tier
	if err := s.ticketTierRepo.Delete(ctx, tierID); err != nil {
		if errors.Is(err, repository.ErrTicketTierNotFound) {
			return ErrTicketTierNotFound
		}
		return fmt.Errorf("failed to delete ticket tier: %w", err)
	}

	return nil
}
