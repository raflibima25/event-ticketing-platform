package cache

import (
	"context"
	"time"
)

// RedisClient defines the interface for Redis operations
// This abstraction allows switching between TCP (local) and REST (production) implementations
type RedisClient interface {
	// Set stores a key-value pair with expiration
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get retrieves a value by key
	// Returns empty string if key doesn't exist
	Get(ctx context.Context, key string) (string, error)

	// Del deletes one or more keys
	// Returns error if deletion fails
	Del(ctx context.Context, keys ...string) error

	// SetNX sets a key only if it doesn't exist (for distributed locks)
	// Returns true if key was set, false if key already exists
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)

	// Exists checks if key exists
	// Returns the number of keys that exist
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Expire sets expiration on existing key
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// Ping checks connection health
	// Returns error if connection is not healthy
	Ping(ctx context.Context) error

	// Close closes the connection
	Close() error
}
