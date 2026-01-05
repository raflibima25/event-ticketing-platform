package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	Port                string
	GRPCPort            string
	Database            DatabaseConfig
	Redis               RedisConfig
	JWTSecret           string
	Reservation         ReservationConfig
	PaymentService      PaymentServiceConfig
	NotificationService NotificationServiceConfig
	AuthService         AuthServiceConfig
	Environment         string
}

// PaymentServiceConfig holds payment service gRPC configuration
type PaymentServiceConfig struct {
	GRPCAddress string
}

// NotificationServiceConfig holds notification service gRPC configuration
type NotificationServiceConfig struct {
	GRPCAddress string
}

// AuthServiceConfig holds auth service HTTP configuration
type AuthServiceConfig struct {
	BaseURL string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// ReservationConfig holds reservation timeout configuration
type ReservationConfig struct {
	Timeout         time.Duration // Default: 15 minutes
	CleanupInterval time.Duration // Background job interval
}

// Load loads configuration from environment variables
func Load() *Config {
	// Parse reservation timeout (default 15 minutes)
	timeout := 15 * time.Minute
	if timeoutStr := os.Getenv("RESERVATION_TIMEOUT"); timeoutStr != "" {
		if d, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = d
		}
	}

	// Parse cleanup interval (default 1 minute)
	cleanupInterval := 1 * time.Minute
	if intervalStr := os.Getenv("CLEANUP_INTERVAL"); intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			cleanupInterval = d
		}
	}

	// Parse Redis DB (default 0)
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	return &Config{
		Port:     getEnv("TICKETING_SERVER_PORT", "8083"),
		GRPCPort: getEnv("TICKETING_GRPC_PORT", "50053"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "ticketing_platform"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		Reservation: ReservationConfig{
			Timeout:         timeout,
			CleanupInterval: cleanupInterval,
		},
		PaymentService: PaymentServiceConfig{
			GRPCAddress: getEnv("PAYMENT_SERVICE_GRPC_ADDR", "localhost:50054"),
		},
		NotificationService: NotificationServiceConfig{
			GRPCAddress: getEnv("NOTIFICATION_SERVICE_GRPC_ADDR", "localhost:50055"),
		},
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// GetDatabaseURL constructs PostgreSQL connection URL
func (c *Config) GetDatabaseURL() string {
	// Check if using Cloud SQL Unix socket (path starts with /)
	if len(c.Database.Host) > 0 && c.Database.Host[0] == '/' {
		// Unix socket format: postgres://user:password@/dbname?host=/cloudsql/INSTANCE
		return fmt.Sprintf(
			"postgres://%s:%s@/%s?host=%s&sslmode=%s",
			c.Database.User,
			c.Database.Password,
			c.Database.Name,
			c.Database.Host,
			c.Database.SSLMode,
		)
	}

	// TCP connection format: postgres://user:password@host:port/dbname?sslmode=disable
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRedisURL constructs Redis connection URL
func (c *Config) GetRedisURL() string {
	if c.Redis.Password != "" {
		return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
	}
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
