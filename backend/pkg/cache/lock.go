package cache

import (
	"context"
	"time"
)

// DistributedLockClient wraps RedisClient with convenience methods for distributed locking
type DistributedLockClient struct {
	RedisClient
}

// NewDistributedLockClient creates a wrapper with distributed locking methods
func NewDistributedLockClient(client RedisClient) *DistributedLockClient {
	return &DistributedLockClient{RedisClient: client}
}

// AcquireLock acquires a distributed lock using SetNX
// Returns true if lock was acquired, false if lock already exists
func (c *DistributedLockClient) AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	// Use SET NX (set if not exists) with expiration
	// Value is set to "locked" but could be any value (e.g., owner ID)
	return c.SetNX(ctx, key, "locked", expiration)
}

// ReleaseLock releases a distributed lock by deleting the key
func (c *DistributedLockClient) ReleaseLock(ctx context.Context, key string) error {
	return c.Del(ctx, key)
}
