package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/utility"
)

var (
	ErrInsufficientQuota     = errors.New("insufficient ticket quota available")
	ErrInvalidQuantity       = errors.New("invalid quantity")
	ErrMaxPerOrderExceeded   = errors.New("maximum tickets per order exceeded")
	ErrLockAcquisitionFailed = errors.New("failed to acquire lock, please try again")
	ErrTicketTierNotFound    = errors.New("ticket tier not found")
)

// ReservationService handles ticket reservation with distributed locking
type ReservationService interface {
	CreateReservation(ctx context.Context, userID string, req *request.CreateOrderRequest) (*response.OrderResponse, error)
	ReleaseReservation(ctx context.Context, orderID string, newStatus string) error
	CleanupExpiredReservations(ctx context.Context) (int, error)
}

// reservationService implements ReservationService interface
type reservationService struct {
	orderRepo      repository.OrderRepository
	orderItemRepo  repository.OrderItemRepository
	ticketTierRepo repository.TicketTierRepository
	redisClient    *utility.RedisClient
	timeout        time.Duration
}

// NewReservationService creates new reservation service instance
func NewReservationService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	ticketTierRepo repository.TicketTierRepository,
	redisClient *utility.RedisClient,
	timeout time.Duration,
) ReservationService {
	return &reservationService{
		orderRepo:      orderRepo,
		orderItemRepo:  orderItemRepo,
		ticketTierRepo: ticketTierRepo,
		redisClient:    redisClient,
		timeout:        timeout,
	}
}

// CreateReservation creates a ticket reservation with distributed + database locking
// This is the CRITICAL function that prevents overselling
func (s *reservationService) CreateReservation(ctx context.Context, userID string, req *request.CreateOrderRequest) (*response.OrderResponse, error) {
	// Step 1: Validate request
	if len(req.Items) == 0 {
		return nil, ErrInvalidQuantity
	}

	// Step 2: Acquire distributed locks for all ticket tiers (Redis)
	// Skip if Redis is not available (development mode)
	var lockKeys []string
	if s.redisClient != nil {
		lockKeys = make([]string, len(req.Items))
		for i, item := range req.Items {
			lockKeys[i] = fmt.Sprintf("lock:tier:%s", item.TicketTierID)
		}

		// Try to acquire all locks with timeout
		lockCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		for _, key := range lockKeys {
			acquired, err := s.redisClient.AcquireLock(lockCtx, key, 10*time.Second)
			if err != nil || !acquired {
				// Release any acquired locks
				for _, k := range lockKeys {
					s.redisClient.ReleaseLock(ctx, k)
				}
				return nil, ErrLockAcquisitionFailed
			}
		}

		// Ensure locks are released when done
		defer func() {
			for _, key := range lockKeys {
				s.redisClient.ReleaseLock(context.Background(), key)
			}
		}()
	}

	// Step 3: Start database transaction
	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is rolled back on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Step 4: Calculate totals and validate availability
	var totalAmount float64
	tierPrices := make(map[string]float64) // Store tier prices

	for _, item := range req.Items {
		// Get tier with row-level lock (SELECT FOR UPDATE)
		tier, err := s.ticketTierRepo.GetByIDWithLock(ctx, tx, item.TicketTierID)
		if err != nil {
			if errors.Is(err, repository.ErrTicketTierNotFound) {
				return nil, ErrTicketTierNotFound
			}
			return nil, fmt.Errorf("failed to get ticket tier: %w", err)
		}

		// Validate quantity
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}

		// Check max per order
		if item.Quantity > tier.MaxPerOrder {
			return nil, ErrMaxPerOrderExceeded
		}

		// Check availability
		available := tier.Quota - tier.SoldCount
		if available < item.Quantity {
			return nil, ErrInsufficientQuota
		}

		// Calculate subtotal
		subtotal := tier.Price * float64(item.Quantity)
		totalAmount += subtotal
		tierPrices[item.TicketTierID] = tier.Price

		// Update sold count (reserve inventory)
		if err := s.ticketTierRepo.UpdateSoldCount(ctx, tx, item.TicketTierID, item.Quantity); err != nil {
			if errors.Is(err, repository.ErrInsufficientQuota) {
				return nil, ErrInsufficientQuota
			}
			return nil, fmt.Errorf("failed to update sold count: %w", err)
		}
	}

	// Step 5: Calculate fees
	platformFee := totalAmount * 0.05  // 5% platform fee
	serviceFee := 2500.0               // Rp 2,500 service fee
	grandTotal := totalAmount + platformFee + serviceFee

	// Step 6: Create order
	expiresAt := time.Now().Add(s.timeout)
	order := &entity.Order{
		UserID:               userID,
		EventID:              req.EventID,
		TotalAmount:          totalAmount,
		PlatformFee:          platformFee,
		ServiceFee:           serviceFee,
		GrandTotal:           grandTotal,
		Status:               entity.OrderStatusReserved,
		ReservationExpiresAt: &expiresAt,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Step 7: Create order items
	orderItems := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		orderItems[i] = entity.OrderItem{
			OrderID:      order.ID,
			TicketTierID: item.TicketTierID,
			Quantity:     item.Quantity,
			Price:        tierPrices[item.TicketTierID],
		}
	}

	if err := s.orderItemRepo.CreateBatch(ctx, tx, orderItems); err != nil {
		return nil, fmt.Errorf("failed to create order items: %w", err)
	}

	// Step 8: Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Step 9: Return response
	return response.ToOrderResponse(order, orderItems), nil
}

// ReleaseReservation releases a reservation and returns inventory
// newStatus can be either "cancelled" (manual) or "expired" (automatic)
func (s *reservationService) ReleaseReservation(ctx context.Context, orderID string, newStatus string) error {
	// Start transaction
	tx, err := s.orderRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get order with lock
	order, err := s.orderRepo.GetByIDWithLock(ctx, tx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Only release if status is reserved
	if order.Status != entity.OrderStatusReserved {
		return fmt.Errorf("order is not in reserved status")
	}

	// Get order items
	items, err := s.orderItemRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	// Release inventory for each item
	for _, item := range items {
		if err := s.ticketTierRepo.ReleaseSoldCount(ctx, tx, item.TicketTierID, item.Quantity); err != nil {
			return fmt.Errorf("failed to release sold count: %w", err)
		}
	}

	// Update order status (cancelled or expired)
	order.Status = newStatus
	if err := s.orderRepo.UpdateWithTx(ctx, tx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CleanupExpiredReservations releases expired reservations (called by background worker)
func (s *reservationService) CleanupExpiredReservations(ctx context.Context) (int, error) {
	// Get expired reservations
	expiredOrders, err := s.orderRepo.GetExpiredReservations(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired reservations: %w", err)
	}

	if len(expiredOrders) == 0 {
		return 0, nil
	}

	releasedCount := 0

	// Process each expired order
	for _, order := range expiredOrders {
		// Acquire lock for this order
		lockKey := fmt.Sprintf("lock:order:%s", order.ID)
		acquired, err := s.redisClient.AcquireLock(ctx, lockKey, 10*time.Second)
		if err != nil || !acquired {
			// Skip if can't acquire lock (might be processing payment)
			continue
		}

		// Release reservation with "expired" status
		if err := s.ReleaseReservation(ctx, order.ID, entity.OrderStatusExpired); err != nil {
			// Log error but continue processing other orders
			s.redisClient.ReleaseLock(ctx, lockKey)
			continue
		}

		s.redisClient.ReleaseLock(ctx, lockKey)
		releasedCount++
	}

	return releasedCount, nil
}
