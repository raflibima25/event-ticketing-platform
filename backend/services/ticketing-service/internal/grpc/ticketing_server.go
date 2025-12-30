package grpc

import (
	"context"
	"log"

	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/ticketing"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/request"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/service"
)

// TicketingGRPCServer implements ticketing gRPC service
type TicketingGRPCServer struct {
	pb.UnimplementedTicketingServiceServer
	confirmationService service.ConfirmationService
}

// NewTicketingGRPCServer creates new ticketing gRPC server instance
func NewTicketingGRPCServer(confirmationService service.ConfirmationService) *TicketingGRPCServer {
	return &TicketingGRPCServer{
		confirmationService: confirmationService,
	}
}

// ConfirmPayment confirms payment and generates tickets
func (s *TicketingGRPCServer) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	log.Printf("[gRPC] ConfirmPayment called for order: %s, payment: %s", req.OrderId, req.PaymentId)

	// Convert gRPC request to internal request
	confirmReq := &request.ConfirmOrderRequest{
		OrderID:       req.OrderId,
		PaymentID:     req.PaymentId,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
	}

	// Call confirmation service
	if err := s.confirmationService.ConfirmPayment(ctx, confirmReq); err != nil {
		log.Printf("[gRPC] ConfirmPayment failed for order %s: %v", req.OrderId, err)
		return &pb.ConfirmPaymentResponse{
			Success: false,
			Message: err.Error(),
		}, nil // Return nil error to avoid gRPC error, but set success=false
	}

	log.Printf("[gRPC] Payment confirmed successfully for order %s", req.OrderId)

	return &pb.ConfirmPaymentResponse{
		Success:          true,
		Message:          "Payment confirmed and tickets generated",
		TicketsGenerated: 0, // TODO: Return actual ticket count
	}, nil
}
