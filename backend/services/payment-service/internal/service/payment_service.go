package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/client"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/entity"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/response"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/repository"
)

var (
	ErrPaymentNotFound    = errors.New("payment transaction not found")
	ErrPaymentAlreadyPaid = errors.New("payment already completed")
	ErrXenditAPIError     = errors.New("xendit API error")
)

// PaymentService handles payment operations
type PaymentService interface {
	CreateInvoice(ctx context.Context, req *request.CreateInvoiceRequest) (*response.InvoiceResponse, error)
	GetInvoice(ctx context.Context, orderID string) (*response.InvoiceResponse, error)
}

// paymentService implements PaymentService interface
type paymentService struct {
	paymentRepo   repository.PaymentRepository
	xenditClient  *client.XenditClient
	invoiceExpiry int
}

// NewPaymentService creates new payment service instance
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	xenditClient *client.XenditClient,
	cfg *config.Config,
) PaymentService {
	return &paymentService{
		paymentRepo:   paymentRepo,
		xenditClient:  xenditClient,
		invoiceExpiry: cfg.Xendit.InvoiceExpiry,
	}
}

// CreateInvoice creates a new payment invoice via Xendit
func (s *paymentService) CreateInvoice(ctx context.Context, req *request.CreateInvoiceRequest) (*response.InvoiceResponse, error) {
	// Check if payment already exists for this order
	existingPayment, err := s.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if err == nil {
		// Payment exists
		if existingPayment.IsPaid() {
			return nil, ErrPaymentAlreadyPaid
		}
		// If pending, return existing invoice
		return response.ToInvoiceResponse(existingPayment), nil
	}

	// Create external ID (format: ORDER-{order_id})
	externalID := fmt.Sprintf("ORDER-%s", req.OrderID)

	// Prepare Xendit invoice request
	xenditReq := &request.XenditCreateInvoiceRequest{
		ExternalID:         externalID,
		Amount:             req.Amount,
		PayerEmail:         req.PayerEmail,
		Description:        req.Description,
		InvoiceDuration:    s.invoiceExpiry,
		SuccessRedirectURL: req.SuccessRedirectURL,
		FailureRedirectURL: req.FailureRedirectURL,
		Currency:           "IDR",
	}

	// Create invoice in Xendit
	xenditResp, err := s.xenditClient.CreateInvoice(xenditReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrXenditAPIError, err)
	}

	// Save payment transaction to database
	invoiceID := xenditResp.ID
	invoiceURL := xenditResp.InvoiceURL
	expiresAt := xenditResp.ExpiryDate

	payment := &entity.PaymentTransaction{
		OrderID:    req.OrderID,
		ExternalID: externalID,
		InvoiceID:  &invoiceID,
		InvoiceURL: &invoiceURL,
		Amount:     req.Amount,
		Status:     entity.PaymentStatusPending,
		ExpiresAt:  &expiresAt,
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment transaction: %w", err)
	}

	return response.ToInvoiceResponse(payment), nil
}

// GetInvoice retrieves invoice by order ID
func (s *paymentService) GetInvoice(ctx context.Context, orderID string) (*response.InvoiceResponse, error) {
	payment, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// If payment is pending, sync with Xendit to get latest status
	if payment.Status == entity.PaymentStatusPending && payment.InvoiceID != nil {
		xenditInvoice, err := s.xenditClient.GetInvoice(*payment.InvoiceID)
		if err == nil {
			// Update local status based on Xendit response
			if xenditInvoice.Status == "PAID" && payment.Status != entity.PaymentStatusPaid {
				paidAt := time.Now()
				payment.Status = entity.PaymentStatusPaid
				payment.PaidAt = &paidAt
				paymentMethod := xenditInvoice.Status
				payment.PaymentMethod = &paymentMethod
				s.paymentRepo.Update(ctx, payment)
			} else if xenditInvoice.Status == "EXPIRED" && payment.Status != entity.PaymentStatusExpired {
				payment.Status = entity.PaymentStatusExpired
				s.paymentRepo.Update(ctx, payment)
			}
		}
	}

	return response.ToInvoiceResponse(payment), nil
}
