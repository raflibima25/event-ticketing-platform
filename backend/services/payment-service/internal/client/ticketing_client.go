package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/ticketing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TicketingClient handles gRPC communication with Ticketing Service
type TicketingClient struct {
	client pb.TicketingServiceClient
	conn   *grpc.ClientConn
}

// ConfirmPaymentRequest represents request to confirm payment
type ConfirmPaymentRequest struct {
	PaymentID     string  `json:"payment_id"`
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
}

// NewTicketingClient creates new ticketing gRPC client instance
// Connection is non-blocking and will auto-reconnect when ticketing service becomes available
func NewTicketingClient(grpcURL string) (*TicketingClient, error) {
	// Use TLS for Cloud Run services (production) or insecure for localhost (development)
	var creds credentials.TransportCredentials
	if grpcURL == "localhost:50052" || grpcURL == "127.0.0.1:50052" {
		creds = insecure.NewCredentials()
		log.Printf("[TicketingGRPC] Using insecure connection for local development")
	} else {
		// Use TLS for Cloud Run
		creds = credentials.NewClientTLSFromCert(nil, "")
		log.Printf("[TicketingGRPC] Using TLS connection for Cloud Run")
	}

	// Use grpc.NewClient for lazy connection with auto-reconnect
	// No WithBlock() - this allows the client to connect lazily and reconnect automatically
	conn, err := grpc.NewClient(
		grpcURL,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticketing client: %w", err)
	}

	client := pb.NewTicketingServiceClient(conn)
	log.Printf("[TicketingGRPC] Ticketing client initialized for %s (lazy connection with auto-reconnect)", grpcURL)

	return &TicketingClient{
		client: client,
		conn:   conn,
	}, nil
}

// ConfirmPayment confirms payment via gRPC
func (c *TicketingClient) ConfirmPayment(orderID string, req *ConfirmPaymentRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert to gRPC request
	grpcReq := &pb.ConfirmPaymentRequest{
		OrderId:       orderID,
		PaymentId:     req.PaymentID,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
	}

	// Call gRPC service
	resp, err := c.client.ConfirmPayment(ctx, grpcReq)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %w", err)
	}

	// Check response success
	if !resp.Success {
		return fmt.Errorf("payment confirmation failed: %s", resp.Message)
	}

	log.Printf("[TicketingGRPC] Payment confirmed for order %s, %d tickets generated", orderID, resp.TicketsGenerated)

	return nil
}

// Close closes the gRPC connection
func (c *TicketingClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
