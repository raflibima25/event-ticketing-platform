package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/raflibima25/event-ticketing-platform/backend/pkg/cache"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/controller"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/router"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/service"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/utility"
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

	log.Println("✓ Database connected successfully")

	// Run database migrations automatically
	migrationsPath := "../../migrations"
	if err := utility.RunMigrations(db, migrationsPath); err != nil {
		log.Printf("⚠️  Migration error: %v", err)
		log.Println("⚠️  Continuing without migrations (ensure database schema is correct)")
	}

	// Initialize Redis with abstraction layer (auto-detects TCP or REST)
	// Used for future features: token blacklist, rate limiting, session cache
	redisClient, err := cache.NewRedisClient()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to Redis: %v", err)
		log.Println("⚠️  Continuing without Redis (advanced features disabled)")
		redisClient = nil
	} else {
		log.Printf("✓ Redis connected successfully (Environment: %s)", cfg.Environment)
		defer redisClient.Close()
	}

	// Initialize JWT utility
	jwtUtil, err := utility.NewJWTUtil(cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		log.Fatalf("Failed to initialize JWT utility: %v", err)
	}

	log.Println("✓ JWT utility initialized")

	// === Dependency Injection (following SOLID principles) ===

	// 1. Initialize Repository Layer (Data Access)
	userRepo := repository.NewUserRepository(db)
	log.Println("✓ Repository layer initialized")

	// 2. Initialize Service Layer (Business Logic)
	authService := service.NewAuthService(userRepo, jwtUtil, redisClient, cfg.BcryptCost)
	log.Println("✓ Service layer initialized")

	// 3. Initialize Controller Layer (HTTP Handlers)
	authController := controller.NewAuthController(authService)
	log.Println("✓ Controller layer initialized")

	// 4. Setup Router with all routes
	r := router.SetupRouter(authController, cfg.JWTSecret)
	log.Println("✓ Router configured")

	// Start HTTP server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Auth Service starting on port %s", cfg.Port)
	log.Printf("📝 Environment: %s", cfg.Environment)
	log.Println("=====================================")

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
