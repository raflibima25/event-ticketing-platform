package main

import (
	"fmt"
	"log"
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/controller"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/repository"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/router"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/service"
	"github.com/raflibima25/event-ticketing-platform/backend/services/auth-service/internal/utility"
)

func main() {
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
	authService := service.NewAuthService(userRepo, jwtUtil, cfg.BcryptCost)
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
