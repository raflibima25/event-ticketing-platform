package grpc

import (
	"context"
	"log"

	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/notification"
	"github.com/raflibima25/event-ticketing-platform/backend/services/notification-service/internal/service"
)

// NotificationGRPCServer implements notification gRPC service
type NotificationGRPCServer struct {
	pb.UnimplementedNotificationServiceServer
	emailService service.EmailService
}

// NewNotificationGRPCServer creates new notification gRPC server instance
func NewNotificationGRPCServer(emailService service.EmailService) *NotificationGRPCServer {
	return &NotificationGRPCServer{
		emailService: emailService,
	}
}

// SendTicketEmail sends e-ticket email to customer
func (s *NotificationGRPCServer) SendTicketEmail(ctx context.Context, req *pb.SendTicketEmailRequest) (*pb.SendTicketEmailResponse, error) {
	log.Printf("[gRPC] SendTicketEmail called for order: %s, recipient: %s, tickets: %d",
		req.OrderId, req.RecipientEmail, len(req.Tickets))

	// Call email service
	resp, err := s.emailService.SendTicketEmail(ctx, req)
	if err != nil {
		log.Printf("[gRPC] SendTicketEmail failed for order %s: %v", req.OrderId, err)
		return &pb.SendTicketEmailResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("[gRPC] SendTicketEmail completed for order %s, success: %v", req.OrderId, resp.Success)

	return resp, nil
}
