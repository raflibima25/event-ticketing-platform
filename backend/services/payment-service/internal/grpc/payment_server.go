package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/payment"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/service"
)

// PaymentGRPCServer implements the gRPC PaymentService interface
type PaymentGRPCServer struct {
	pb.UnimplementedPaymentServiceServer
	paymentService service.PaymentService
}

// NewPaymentGRPCServer creates new gRPC server instance
func NewPaymentGRPCServer(paymentService service.PaymentService) *PaymentGRPCServer {
	return &PaymentGRPCServer{
		paymentService: paymentService,
	}
}

// CreateInvoice creates a new payment invoice for an order (gRPC endpoint)
func (s *PaymentGRPCServer) CreateInvoice(ctx context.Context, req *pb.CreateInvoiceRequest) (*pb.CreateInvoiceResponse, error) {
	log.Printf("[gRPC] CreateInvoice request for order: %s", req.OrderId)

	// Create internal request (map gRPC request to service request)
	createInvoiceReq := &request.CreateInvoiceRequest{
		OrderID:            req.OrderId,
		Amount:             req.Amount,
		PayerEmail:         req.Email,
		Description:        req.Description,
		SuccessRedirectURL: "",
		FailureRedirectURL: "",
	}

	// Call service layer
	invoiceResp, err := s.paymentService.CreateInvoice(ctx, createInvoiceReq)
	if err != nil {
		log.Printf("[gRPC] CreateInvoice failed for order %s: %v", req.OrderId, err)
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Convert internal response to protobuf response
	expiresAt := ""
	if invoiceResp.ExpiresAt != nil {
		expiresAt = invoiceResp.ExpiresAt.Format(time.RFC3339)
	}

	response := &pb.CreateInvoiceResponse{
		PaymentId:  invoiceResp.ID,
		InvoiceId:  invoiceResp.ExternalID, // Using external ID as invoice ID
		InvoiceUrl: invoiceResp.InvoiceURL,
		ExternalId: invoiceResp.ExternalID,
		Amount:     invoiceResp.Amount,
		Status:     invoiceResp.Status,
		ExpiresAt:  expiresAt,
		CreatedAt:  invoiceResp.CreatedAt.Format(time.RFC3339),
	}

	log.Printf("[gRPC] CreateInvoice success for order %s - Invoice URL: %s", req.OrderId, invoiceResp.InvoiceURL)
	return response, nil
}

// GetPaymentStatus retrieves payment status by order ID (gRPC endpoint)
func (s *PaymentGRPCServer) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.GetPaymentStatusResponse, error) {
	log.Printf("[gRPC] GetPaymentStatus request for order: %s", req.OrderId)

	// Get invoice by order ID
	invoice, err := s.paymentService.GetInvoice(ctx, req.OrderId)
	if err != nil {
		log.Printf("[gRPC] GetPaymentStatus failed for order %s: %v", req.OrderId, err)
		return nil, fmt.Errorf("failed to get payment status: %w", err)
	}

	// Convert to protobuf response
	response := &pb.GetPaymentStatusResponse{
		PaymentId: invoice.ID,
		OrderId:   invoice.OrderID,
		InvoiceId: invoice.ExternalID,
		Amount:    invoice.Amount,
		Status:    invoice.Status,
		CreatedAt: invoice.CreatedAt.Format(time.RFC3339),
	}

	// Note: PaymentMethod and PaidAt are not in InvoiceResponse
	// They would need to be added to the response struct if needed

	log.Printf("[gRPC] GetPaymentStatus success for order %s - Status: %s", req.OrderId, invoice.Status)
	return response, nil
}
