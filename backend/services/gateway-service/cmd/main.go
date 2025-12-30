package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/raflibima25/event-ticketing-platform/backend/services/gateway-service/config"
	"github.com/raflibima25/event-ticketing-platform/backend/services/gateway-service/internal/router"
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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Printf("Starting API Gateway on port %s...", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Rate limiting: %v (RPM: %d, Burst: %d)",
		cfg.RateLimit.Enabled,
		cfg.RateLimit.RequestsPerMinute,
		cfg.RateLimit.BurstSize,
	)

	// Setup router with all middleware and routes
	r := router.SetupRouter(cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 API Gateway running on port %s", cfg.Port)
		log.Println("📡 Proxying requests to:")
		log.Printf("   - Auth Service:        %s", cfg.Services.AuthService)
		log.Printf("   - Event Service:       %s", cfg.Services.EventService)
		log.Printf("   - Ticketing Service:   %s", cfg.Services.TicketingService)
		log.Printf("   - Payment Service:     %s", cfg.Services.PaymentService)
		log.Printf("   - Notification Service: %s", cfg.Services.NotificationService)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway stopped gracefully")
}
