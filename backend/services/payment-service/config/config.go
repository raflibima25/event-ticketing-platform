package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Server           ServerConfig
	Database         DatabaseConfig
	JWT              JWTConfig
	Xendit           XenditConfig
	TicketingService TicketingServiceConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port     string
	GRPCPort string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
	Expiry string
}

// XenditConfig holds Xendit API configuration
type XenditConfig struct {
	APIKey        string
	WebhookToken  string
	BaseURL       string
	InvoiceExpiry int // in seconds
}

// TicketingServiceConfig holds ticketing service configuration
type TicketingServiceConfig struct {
	BaseURL     string
	GRPCAddress string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:     getEnv("PAYMENT_SERVER_PORT", "8084"),
			GRPCPort: getEnv("PAYMENT_GRPC_PORT", "50054"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5433"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "ticketing_platform"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Expiry: getEnv("JWT_EXPIRY", "24h"),
		},
		Xendit: XenditConfig{
			APIKey:        getEnv("XENDIT_API_KEY", ""),
			WebhookToken:  getEnv("XENDIT_WEBHOOK_TOKEN", ""),
			BaseURL:       getEnv("XENDIT_BASE_URL", "https://api.xendit.co"),
			InvoiceExpiry: getEnvAsInt("XENDIT_INVOICE_EXPIRY", 1800), // 30 minutes default
		},
		TicketingService: TicketingServiceConfig{
			BaseURL:     getEnv("TICKETING_SERVICE_URL", "http://localhost:8083"),
			GRPCAddress: getEnv("TICKETING_SERVICE_GRPC_ADDR", "localhost:50053"),
		},
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid integer value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return value
}
