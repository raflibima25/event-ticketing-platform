package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port               string
	Database           DatabaseConfig
	Redis              RedisConfig
	JWTSecret          string
	JWTExpiry          string
	RefreshTokenExpiry string
	BcryptCost         int
	Environment        string
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

// Load loads configuration from environment variables
func Load() *Config {
	bcryptCost, _ := strconv.Atoi(getEnv("BCRYPT_COST", "10"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		Port: getEnv("AUTH_SERVER_PORT", "8081"),
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
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-key"),
		JWTExpiry:          getEnv("JWT_EXPIRY", "24h"),
		RefreshTokenExpiry: getEnv("REFRESH_TOKEN_EXPIRY", "168h"), // 7 days
		BcryptCost:         bcryptCost,
		Environment:        getEnv("ENVIRONMENT", "development"),
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
		return fmt.Sprintf(
			"redis://:%s@%s:%s/%d",
			c.Redis.Password,
			c.Redis.Host,
			c.Redis.Port,
			c.Redis.DB,
		)
	}
	return fmt.Sprintf(
		"redis://%s:%s/%d",
		c.Redis.Host,
		c.Redis.Port,
		c.Redis.DB,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
