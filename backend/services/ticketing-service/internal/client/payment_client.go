package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PaymentClient handles communication with payment service via gRPC
type PaymentClient struct {
	client pb.PaymentServiceClient
	conn   *grpc.ClientConn
}

// NewPaymentClient creates new payment gRPC client
func NewPaymentClient(grpcURL string) (*PaymentClient, error) {
	// Create gRPC connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		grpcURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	client := pb.NewPaymentServiceClient(conn)
	log.Printf("✅ Connected to Payment Service gRPC at %s", grpcURL)

	return &PaymentClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *PaymentClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateInvoiceRequest contains data for creating an invoice
type CreateInvoiceRequest struct {
	OrderID      string
	UserID       string
	Email        string
	CustomerName string
	Amount       float64
	Description  string
	Items        []InvoiceItem
}

// InvoiceItem represents a line item
type InvoiceItem struct {
	Name     string
	Quantity int
	Price    float64
}

// CreateInvoiceResponse contains invoice creation result
type CreateInvoiceResponse struct {
	PaymentID  string
	InvoiceID  string
	InvoiceURL string
	ExternalID string
	Amount     float64
	Status     string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

// CreateInvoice creates a payment invoice via gRPC
func (c *PaymentClient) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	// Convert items to protobuf format
	pbItems := make([]*pb.InvoiceItem, len(req.Items))
	for i, item := range req.Items {
		pbItems[i] = &pb.InvoiceItem{
			Name:     item.Name,
			Quantity: int32(item.Quantity),
			Price:    item.Price,
		}
	}

	// Create gRPC request
	grpcReq := &pb.CreateInvoiceRequest{
		OrderId:      req.OrderID,
		UserId:       req.UserID,
		Email:        req.Email,
		CustomerName: req.CustomerName,
		Amount:       req.Amount,
		Description:  req.Description,
		Items:        pbItems,
	}

	// Call gRPC endpoint with timeout
	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := c.client.CreateInvoice(callCtx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice via gRPC: %w", err)
	}

	// Parse timestamps
	expiresAt, _ := time.Parse(time.RFC3339, resp.ExpiresAt)
	createdAt, _ := time.Parse(time.RFC3339, resp.CreatedAt)

	// Convert response
	return &CreateInvoiceResponse{
		PaymentID:  resp.PaymentId,
		InvoiceID:  resp.InvoiceId,
		InvoiceURL: resp.InvoiceUrl,
		ExternalID: resp.ExternalId,
		Amount:     resp.Amount,
		Status:     resp.Status,
		ExpiresAt:  expiresAt,
		CreatedAt:  createdAt,
	}, nil
}

// GetPaymentStatus retrieves payment status via gRPC
func (c *PaymentClient) GetPaymentStatus(ctx context.Context, orderID string) (*CreateInvoiceResponse, error) {
	grpcReq := &pb.GetPaymentStatusRequest{
		OrderId: orderID,
	}

	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.client.GetPaymentStatus(callCtx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment status via gRPC: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, resp.CreatedAt)
	paidAt, _ := time.Parse(time.RFC3339, resp.PaidAt)

	return &CreateInvoiceResponse{
		PaymentID:  resp.PaymentId,
		InvoiceID:  resp.InvoiceId,
		Amount:     resp.Amount,
		Status:     resp.Status,
		CreatedAt:  createdAt,
		ExpiresAt:  paidAt,
	}, nil
}
