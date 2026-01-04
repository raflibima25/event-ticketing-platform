package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TCPRedisClient implements RedisClient using traditional TCP connection
// Used for local development with Docker Redis
type TCPRedisClient struct {
	client *redis.Client
}

// NewTCPRedisClient creates a new TCP-based Redis client
// This is used for local development connecting to Docker Redis
func NewTCPRedisClient(host, port, password string, db int) (RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis (TCP): %w", err)
	}

	return &TCPRedisClient{client: client}, nil
}

// Set stores a key-value pair with expiration
func (c *TCPRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (c *TCPRedisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key doesn't exist, return empty string (not an error)
		return "", nil
	}
	return result, err
}

// Del deletes one or more keys
func (c *TCPRedisClient) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// SetNX sets a key only if it doesn't exist (for distributed locks)
func (c *TCPRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// Exists checks if key exists
func (c *TCPRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration on existing key
func (c *TCPRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// Ping checks connection health
func (c *TCPRedisClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the connection
func (c *TCPRedisClient) Close() error {
	return c.client.Close()
}
