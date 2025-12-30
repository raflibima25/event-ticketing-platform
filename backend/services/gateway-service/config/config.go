package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Config holds gateway configuration
type Config struct {
	Port        string
	Environment string
	JWTSecret   string
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Services    ServiceURLs
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	Enabled           bool
}

// ServiceURLs holds backend service URLs
type ServiceURLs struct {
	AuthService        string
	EventService       string
	TicketingService   string
	PaymentService     string
	NotificationService string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvAsInt("RATE_LIMIT_RPM", 100),
			BurstSize:         getEnvAsInt("RATE_LIMIT_BURST", 20),
			Enabled:           getEnv("RATE_LIMIT_ENABLED", "true") == "true",
		},
		Services: ServiceURLs{
			AuthService:        getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
			EventService:       getEnv("EVENT_SERVICE_URL", "http://localhost:8082"),
			TicketingService:   getEnv("TICKETING_SERVICE_URL", "http://localhost:8083"),
			PaymentService:     getEnv("PAYMENT_SERVICE_URL", "http://localhost:8084"),
			NotificationService: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8085"),
		},
	}
}

// Validate validates configuration
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		log.Println("⚠️  Warning: JWT_SECRET not set - authentication will not work")
	}
	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets environment variable as integer with fallback
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return fallback
}

// getEnvAsSlice gets environment variable as slice (comma-separated)
func getEnvAsSlice(key, fallback string) []string {
	value := getEnv(key, fallback)
	return strings.Split(value, ",")
}
