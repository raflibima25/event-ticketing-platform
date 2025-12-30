package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/ticketing"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/client"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/controller"
	grpcHandler "github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/grpc"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/router"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/service"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/utility"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/worker"
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

	log.Printf("Starting Ticketing Service on port %s...", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Reservation timeout: %v", cfg.Reservation.Timeout)

	// Initialize database
	db, err := utility.NewDatabase(utility.DatabaseConfig{
		URL:             cfg.GetDatabaseURL(),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// Run migrations
	migrationsPath := "../../migrations"
	if err := utility.RunMigrations(db, migrationsPath); err != nil {
		log.Printf("⚠️  Migration error: %v", err)
		log.Println("⚠️  Continuing without migrations (ensure database schema is correct)")
	}

	log.Println("Database migrations completed")

	// Initialize Redis (optional - graceful degradation)
	redisClient, err := utility.NewRedisClient(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to Redis: %v", err)
		log.Println("⚠️  Continuing without Redis (distributed locking disabled)")
		log.Println("⚠️  NOTE: Do not run multiple instances without Redis!")
		redisClient = nil
	} else {
		log.Println("✓ Redis connected successfully")
		defer redisClient.Close()
	}

	// Initialize repositories
	orderRepo := repository.NewOrderRepository(db)
	orderItemRepo := repository.NewOrderItemRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	ticketTierRepo := repository.NewTicketTierRepository(db)
	eventRepo := repository.NewEventRepository(db)
	userRepo := repository.NewUserRepository(db)

	log.Println("Repositories initialized")

	// Initialize payment gRPC client (with auto-reconnect)
	paymentClient, err := client.NewPaymentClient(cfg.PaymentService.GRPCAddress)
	if err != nil {
		log.Fatalf("Failed to create payment client: %v", err)
	}
	defer paymentClient.Close()
	log.Println("✓ Payment client initialized (will auto-reconnect if service unavailable)")

	// Initialize notification gRPC client (with auto-reconnect)
	notificationClient, err := client.NewNotificationClient(cfg.NotificationService.GRPCAddress)
	if err != nil {
		log.Fatalf("Failed to create notification client: %v", err)
	}
	defer notificationClient.Close()
	log.Println("✓ Notification client initialized (will auto-reconnect if service unavailable)")

	// Initialize services with dependency injection
	ticketService := service.NewTicketService(
		ticketRepo,
		orderRepo,
		orderItemRepo,
	)

	reservationService := service.NewReservationService(
		orderRepo,
		orderItemRepo,
		ticketTierRepo,
		redisClient,
		paymentClient,
		cfg.Reservation.Timeout,
	)

	orderService := service.NewOrderService(
		orderRepo,
		orderItemRepo,
		reservationService,
	)

	confirmationService := service.NewConfirmationService(
		orderRepo,
		orderItemRepo,
		ticketTierRepo,
		eventRepo,
		userRepo,
		ticketService,
		notificationClient,
	)

	log.Println("Services initialized")

	// Initialize controllers
	orderController := controller.NewOrderController(
		reservationService,
		orderService,
		confirmationService,
	)

	ticketController := controller.NewTicketController(
		ticketService,
	)

	log.Println("Controllers initialized")

	// Setup router
	r := router.SetupRouter(
		orderController,
		ticketController,
		cfg.JWTSecret,
	)

	log.Println("Router configured")

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	ticketingGRPCServer := grpcHandler.NewTicketingGRPCServer(confirmationService)
	pb.RegisterTicketingServiceServer(grpcServer, ticketingGRPCServer)
	reflection.Register(grpcServer)

	log.Println("gRPC server initialized")

	// Start background worker for reservation cleanup
	cleanupWorker := worker.NewReservationCleanupWorker(
		reservationService,
		cfg.Reservation.CleanupInterval,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker in goroutine
	go cleanupWorker.Start(ctx)

	log.Println("Background worker started")

	// Start HTTP server
	serverAddr := ":" + cfg.Port

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run HTTP server in goroutine
	go func() {
		log.Printf("🚀 HTTP Server running on port %s", cfg.Port)
		if err := r.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Run gRPC server in goroutine
	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port: %v", err)
		}
		log.Printf("🚀 gRPC Server running on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Stop background worker
	cleanupWorker.Stop()

	// Give outstanding operations a deadline for completion
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Wait for shutdown context
	<-shutdownCtx.Done()

	log.Println("Server stopped gracefully")
}
