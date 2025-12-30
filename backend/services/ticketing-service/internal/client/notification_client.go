package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/notification"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NotificationClient handles gRPC communication with Notification Service
type NotificationClient struct {
	client pb.NotificationServiceClient
	conn   *grpc.ClientConn
}

// NewNotificationClient creates new notification gRPC client instance
// Connection is lazy and will auto-reconnect if service is unavailable
func NewNotificationClient(grpcURL string) (*NotificationClient, error) {
	// Use grpc.NewClient for lazy connection with auto-reconnect
	// No WithBlock() - this allows the client to connect lazily and reconnect automatically
	conn, err := grpc.NewClient(
		grpcURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification client: %w", err)
	}

	client := pb.NewNotificationServiceClient(conn)

	log.Printf("[NotificationGRPC] Notification client initialized for %s (lazy connection with auto-reconnect)", grpcURL)

	return &NotificationClient{
		client: client,
		conn:   conn,
	}, nil
}

// SendTicketEmailRequest represents request to send ticket email
type SendTicketEmailRequest struct {
	OrderID        string
	RecipientEmail string
	RecipientName  string
	EventName      string
	EventLocation  string
	EventStartTime string
	TotalAmount    float64
	PaymentMethod  string
	Tickets        []TicketInfo
}

// TicketInfo represents ticket information for email
type TicketInfo struct {
	TicketID string
	QRCode   string
	TierName string
	Price    float64
}

// SendTicketEmail sends e-ticket email via gRPC
func (c *NotificationClient) SendTicketEmail(ctx context.Context, req *SendTicketEmailRequest) error {
	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Convert tickets to protobuf format
	pbTickets := make([]*pb.Ticket, len(req.Tickets))
	for i, ticket := range req.Tickets {
		pbTickets[i] = &pb.Ticket{
			TicketId: ticket.TicketID,
			QrCode:   ticket.QRCode,
			TierName: ticket.TierName,
			Price:    ticket.Price,
		}
	}

	// Convert to gRPC request
	grpcReq := &pb.SendTicketEmailRequest{
		OrderId:        req.OrderID,
		RecipientEmail: req.RecipientEmail,
		RecipientName:  req.RecipientName,
		EventName:      req.EventName,
		EventLocation:  req.EventLocation,
		EventStartTime: req.EventStartTime,
		TotalAmount:    req.TotalAmount,
		PaymentMethod:  req.PaymentMethod,
		Tickets:        pbTickets,
	}

	// Call gRPC service
	resp, err := c.client.SendTicketEmail(callCtx, grpcReq)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %w", err)
	}

	// Check response success
	if !resp.Success {
		return fmt.Errorf("failed to send email: %s", resp.Message)
	}

	log.Printf("[NotificationGRPC] Email sent successfully for order %s, email ID: %s", req.OrderID, resp.EmailId)

	return nil
}

// Close closes the gRPC connection
func (c *NotificationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
