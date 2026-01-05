package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/payment"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/client"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/controller"
	grpcHandler "github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/grpc"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/service"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/utility"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/router"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load .env file from project root
	envPath := filepath.Join("..", "..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️  Warning: .env file not found at %s", envPath)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := utility.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to database")

	// Run migrations
	// Production (Docker): ./migrations (copied to /root/migrations)
	// Development (local): ../../migrations (relative to services/payment-service/cmd)
	migrationsPath := "../../migrations"
	// Note: payment-service doesn't have Environment in config, checking via env var
	if os.Getenv("ENVIRONMENT") == "production" {
		migrationsPath = "./migrations"
	}
	if err := utility.RunMigrations(db, migrationsPath); err != nil {
		log.Printf("⚠️  Migration error: %v", err)
		log.Println("⚠️  Continuing without migrations (ensure database schema is correct)")
	}

	// Initialize repositories
	paymentRepo := repository.NewPaymentRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)
	log.Println("✅ Repositories initialized")

	// Initialize clients
	xenditClient := client.NewXenditClient(&cfg.Xendit)

	// Initialize ticketing gRPC client (non-blocking with auto-reconnect)
	ticketingClient, err := client.NewTicketingClient(cfg.TicketingService.GRPCAddress)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize Ticketing Service gRPC client: %v", err)
		log.Println("⚠️  Payment service will continue without ticketing client")
		ticketingClient = nil
	} else {
		defer ticketingClient.Close()
	}

	log.Println("✅ External clients initialized")

	// Initialize services
	paymentService := service.NewPaymentService(paymentRepo, xenditClient, cfg)
	webhookService := service.NewWebhookService(webhookRepo, paymentRepo, ticketingClient)
	log.Println("✅ Services initialized")

	// Initialize controllers
	paymentController := controller.NewPaymentController(paymentService)
	webhookController := controller.NewWebhookController(webhookService, cfg)
	log.Println("✅ Controllers initialized")

	// Setup HTTP router
	r := router.SetupRouter(cfg, paymentController, webhookController)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	paymentGRPCServer := grpcHandler.NewPaymentGRPCServer(paymentService)
	pb.RegisterPaymentServiceServer(grpcServer, paymentGRPCServer)

	// Enable gRPC reflection for debugging (optional)
	reflection.Register(grpcServer)
	log.Println("✅ gRPC server initialized")

	// Start HTTP server in goroutine
	go func() {
		log.Printf("🚀 HTTP Server running on port %s", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start HTTP server: %v", err)
		}
	}()

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

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down payment service...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("❌ HTTP server forced to shutdown: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	log.Println("✅ Payment service stopped gracefully")
}
