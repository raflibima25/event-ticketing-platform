package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/raflibima25/event-ticketing-platform/backend/pkg/cache"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/controller"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/router"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/service"
	"github.com/raflibima25/event-ticketing-platform/backend/services/event-service/internal/utility"
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

	// Initialize database connection
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

	log.Println("Successfully connected to database")

	// Run database migrations automatically
	// Production (Docker): ./migrations (copied to /root/migrations)
	// Development (local): ../../migrations (relative to services/event-service/cmd)
	migrationsPath := "../../migrations"
	if cfg.Environment == "production" {
		migrationsPath = "./migrations"
	}
	if err := utility.RunMigrations(db, migrationsPath); err != nil{
		log.Printf("⚠️  Migration error: %v", err)
		log.Println("⚠️  Continuing without migrations (ensure database schema is correct)")
	}

	// Initialize Redis with abstraction layer (auto-detects TCP or REST)
	redisClient, err := cache.NewRedisClient()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to Redis: %v", err)
		log.Println("⚠️  Continuing without Redis (caching disabled)")
		redisClient = nil
	} else {
		log.Printf("✓ Redis connected successfully (Environment: %s)", cfg.Environment)
		defer redisClient.Close()
	}

	// Initialize Repository Layer
	eventRepo := repository.NewEventRepository(db)
	ticketTierRepo := repository.NewTicketTierRepository(db)

	log.Println("Repository layer initialized")

	// Initialize Service Layer with Redis caching
	eventService := service.NewEventService(eventRepo, ticketTierRepo, redisClient)

	log.Println("Service layer initialized")

	// Initialize Controller Layer
	eventController := controller.NewEventController(eventService)

	log.Println("Controller layer initialized")

	// Setup Router
	r := router.SetupRouter(eventController, cfg.JWTSecret)

	log.Println("Router configured")

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Event Service starting on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
