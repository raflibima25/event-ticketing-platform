package cache

import (
	"fmt"
	"os"
	"strconv"
)

// NewRedisClient creates appropriate Redis client based on environment
// - Development: TCP connection to localhost Docker Redis
// - Production: REST connection to Upstash
//
// Environment is determined by ENVIRONMENT variable:
//   - "development" or empty: Use TCP (local Redis)
//   - "production": Use REST (Upstash)
func NewRedisClient() (RedisClient, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development" // Default to development
	}

	switch env {
	case "development":
		return newLocalRedisClient()
	case "production":
		return newUpstashRedisClient()
	default:
		return nil, fmt.Errorf("unknown environment: %s (expected 'development' or 'production')", env)
	}
}

// newLocalRedisClient creates TCP client for local development
func newLocalRedisClient() (RedisClient, error) {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost" // Default
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379" // Default Redis port
	}

	password := os.Getenv("REDIS_PASSWORD")
	// Password can be empty for local dev

	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		var err error
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB value '%s': %w", dbStr, err)
		}
	}

	fmt.Printf("📡 Connecting to local Redis (TCP): %s:%s (DB: %d)\n", host, port, db)
	client, err := NewTCPRedisClient(host, port, password, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create local Redis client: %w", err)
	}

	fmt.Println("✅ Connected to local Redis successfully")
	return client, nil
}

// newUpstashRedisClient creates REST client for production
func newUpstashRedisClient() (RedisClient, error) {
	url := os.Getenv("UPSTASH_REDIS_REST_URL")
	token := os.Getenv("UPSTASH_REDIS_REST_TOKEN")

	if url == "" || token == "" {
		return nil, fmt.Errorf("UPSTASH_REDIS_REST_URL and UPSTASH_REDIS_REST_TOKEN must be set in production environment")
	}

	fmt.Printf("📡 Connecting to Upstash Redis (REST): %s\n", url)
	client, err := NewRESTRedisClient(url, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Upstash Redis client: %w", err)
	}

	fmt.Println("✅ Connected to Upstash Redis successfully")
	return client, nil
}
