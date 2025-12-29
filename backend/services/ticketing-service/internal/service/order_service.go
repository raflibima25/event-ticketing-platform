package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/repository"
)

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderExpired          = errors.New("order has expired")
	ErrOrderAlreadyPaid      = errors.New("order has already been paid")
	ErrOrderAlreadyCancelled = errors.New("order has already been cancelled")
	ErrCannotCancelOrder     = errors.New("cannot cancel order at this stage")
	ErrUnauthorized          = errors.New("unauthorized to access this order")
)

// OrderService handles order operations
type OrderService interface {
	GetOrderByID(ctx context.Context, userID, orderID string) (*response.OrderResponse, error)
	GetUserOrders(ctx context.Context, userID string, page, limit int) ([]response.OrderResponse, int64, error)
	CancelOrder(ctx context.Context, userID, orderID string) error
}

// orderService implements OrderService interface
type orderService struct {
	orderRepo         repository.OrderRepository
	orderItemRepo     repository.OrderItemRepository
	reservationService ReservationService
}

// NewOrderService creates new order service instance
func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	reservationService ReservationService,
) OrderService {
	return &orderService{
		orderRepo:         orderRepo,
		orderItemRepo:     orderItemRepo,
		reservationService: reservationService,
	}
}

// GetOrderByID retrieves order by ID with authorization check
func (s *orderService) GetOrderByID(ctx context.Context, userID, orderID string) (*response.OrderResponse, error) {
	// Get order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check authorization
	if order.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Get order items
	items, err := s.orderItemRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	return response.ToOrderResponse(order, items), nil
}

// GetUserOrders retrieves all orders for a user with pagination
func (s *orderService) GetUserOrders(ctx context.Context, userID string, page, limit int) ([]response.OrderResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Get orders
	orders, total, err := s.orderRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user orders: %w", err)
	}

	// Convert to response
	orderResponses := make([]response.OrderResponse, 0, len(orders))
	for _, order := range orders {
		// Get items for this order
		items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get order items: %w", err)
		}

		orderResponses = append(orderResponses, *response.ToOrderResponse(&order, items))
	}

	return orderResponses, total, nil
}

// CancelOrder cancels an order and releases inventory
func (s *orderService) CancelOrder(ctx context.Context, userID, orderID string) error {
	// Get order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check authorization
	if order.UserID != userID {
		return ErrUnauthorized
	}

	// Check if order can be cancelled (only reserved orders)
	if !order.CanBeCancelled() {
		return ErrCannotCancelOrder
	}

	// Release the reservation with "cancelled" status (quota will be returned)
	if err := s.reservationService.ReleaseReservation(ctx, orderID, "cancelled"); err != nil {
		return fmt.Errorf("failed to release reservation: %w", err)
	}

	// NOTE: Paid orders cannot be cancelled via this endpoint
	// Refund requests must go through Payment Service which will:
	// 1. Validate refund eligibility (event date, refund policy)
	// 2. Process refund via Xendit
	// 3. Update order status to 'refunded'
	// 4. Cancel tickets
	// 5. Adjust inventory if needed

	return nil
}
