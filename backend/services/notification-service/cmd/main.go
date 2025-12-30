package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/notification"
	"github.com/raflibima25/event-ticketing-platform/backend/services/notification-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/notification-service/internal/client"
	grpcHandler "github.com/raflibima25/event-ticketing-platform/backend/services/notification-service/internal/grpc"
	"github.com/raflibima25/event-ticketing-platform/backend/services/notification-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load .env file from project root
	envPath := filepath.Join("..", "..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️  Warning: .env file not found at %s, using environment variables or defaults", envPath)
	} else {
		log.Println("✓ Loaded .env file")
	}

	// Load configuration
	cfg := config.Load()

	log.Printf("Starting Notification Service on gRPC port %s...", cfg.Server.GRPCPort)

	// Validate Resend configuration
	if cfg.Resend.APIKey == "" {
		log.Fatal("❌ RESEND_API_KEY is required. Please set it in .env file")
	}

	// Initialize Resend client
	resendClient := client.NewResendClient(cfg.Resend.APIKey)
	log.Println("✅ Resend client initialized")

	// Display test mode configuration
	if cfg.Resend.TestMode {
		if cfg.Resend.TestEmail == "" {
			log.Fatal("❌ RESEND_TEST_MODE is enabled but RESEND_TEST_EMAIL is not set")
		}
		log.Printf("🧪 Test mode enabled - all emails will be sent to: %s", cfg.Resend.TestEmail)
	} else {
		log.Println("📧 Production mode - emails will be sent to actual recipients")
	}

	// Initialize services
	emailService := service.NewEmailService(
		resendClient,
		cfg.Resend.FromName,
		cfg.Resend.FromEmail,
		cfg.Resend.TestMode,
		cfg.Resend.TestEmail,
	)
	log.Println("✅ Email service initialized")

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	notificationGRPCServer := grpcHandler.NewNotificationGRPCServer(emailService)
	pb.RegisterNotificationServiceServer(grpcServer, notificationGRPCServer)
	reflection.Register(grpcServer)

	log.Println("✅ gRPC server initialized")

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start gRPC server in goroutine
	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
		if err != nil {
			log.Fatalf("❌ Failed to listen on gRPC port: %v", err)
		}
		log.Printf("🚀 gRPC Server running on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("❌ Failed to start gRPC server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down Notification Service...")

	// Gracefully stop gRPC server
	grpcServer.GracefulStop()

	log.Println("✓ Notification Service stopped gracefully")
}
