package cache

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestTCPRedisClient_BasicOperations tests TCP Redis client with local Docker Redis
func TestTCPRedisClient_BasicOperations(t *testing.T) {
	// Skip if Redis is not available
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	client, err := NewTCPRedisClient(host, port, "", 0)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Test 1: Ping
	t.Run("Ping", func(t *testing.T) {
		err := client.Ping(ctx)
		if err != nil {
			t.Fatalf("Ping failed: %v", err)
		}
	})

	// Test 2: Set and Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "test:key:1"
		value := "test-value"

		err := client.Set(ctx, key, value, 10*time.Second)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		result, err := client.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if result != value {
			t.Errorf("Expected %s, got %s", value, result)
		}

		// Cleanup
		client.Del(ctx, key)
	})

	// Test 3: SetNX (distributed lock)
	t.Run("SetNX", func(t *testing.T) {
		key := "test:lock:1"
		value := "lock-value"

		// First SetNX should succeed
		ok, err := client.SetNX(ctx, key, value, 10*time.Second)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if !ok {
			t.Error("Expected SetNX to succeed")
		}

		// Second SetNX should fail (key exists)
		ok, err = client.SetNX(ctx, key, "another-value", 10*time.Second)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if ok {
			t.Error("Expected SetNX to fail when key exists")
		}

		// Cleanup
		client.Del(ctx, key)
	})

	// Test 4: Exists
	t.Run("Exists", func(t *testing.T) {
		key := "test:exists:1"

		// Key should not exist
		count, err := client.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}

		// Set key
		client.Set(ctx, key, "value", 10*time.Second)

		// Key should exist
		count, err = client.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}

		// Cleanup
		client.Del(ctx, key)
	})

	// Test 5: Expire
	t.Run("Expire", func(t *testing.T) {
		key := "test:expire:1"

		// Set key without expiration
		client.Set(ctx, key, "value", 0)

		// Set expiration
		err := client.Expire(ctx, key, 2*time.Second)
		if err != nil {
			t.Fatalf("Expire failed: %v", err)
		}

		// Key should still exist
		count, _ := client.Exists(ctx, key)
		if count != 1 {
			t.Error("Expected key to exist before expiration")
		}

		// Wait for expiration
		time.Sleep(3 * time.Second)

		// Key should be gone
		count, _ = client.Exists(ctx, key)
		if count != 0 {
			t.Error("Expected key to be expired")
		}
	})

	// Test 6: Del
	t.Run("Del", func(t *testing.T) {
		key1 := "test:del:1"
		key2 := "test:del:2"

		// Set keys
		client.Set(ctx, key1, "value1", 10*time.Second)
		client.Set(ctx, key2, "value2", 10*time.Second)

		// Delete keys
		err := client.Del(ctx, key1, key2)
		if err != nil {
			t.Fatalf("Del failed: %v", err)
		}

		// Keys should not exist
		count, _ := client.Exists(ctx, key1, key2)
		if count != 0 {
			t.Errorf("Expected keys to be deleted, found %d", count)
		}
	})
}

// TestRESTRedisClient_BasicOperations tests REST Redis client with Upstash
func TestRESTRedisClient_BasicOperations(t *testing.T) {
	// Skip if Upstash credentials not set
	url := os.Getenv("UPSTASH_REDIS_REST_URL")
	token := os.Getenv("UPSTASH_REDIS_REST_TOKEN")

	if url == "" || token == "" {
		t.Skip("Skipping test: UPSTASH_REDIS_REST_URL and UPSTASH_REDIS_REST_TOKEN not set")
		return
	}

	client, err := NewRESTRedisClient(url, token)
	if err != nil {
		t.Fatalf("Failed to create REST client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test 1: Ping
	t.Run("Ping", func(t *testing.T) {
		err := client.Ping(ctx)
		if err != nil {
			t.Fatalf("Ping failed: %v", err)
		}
	})

	// Test 2: Set and Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "test:rest:key:1"
		value := "test-rest-value"

		err := client.Set(ctx, key, value, 10*time.Second)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		result, err := client.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if result != value {
			t.Errorf("Expected %s, got %s", value, result)
		}

		// Cleanup
		client.Del(ctx, key)
	})

	// Test 3: SetNX (distributed lock)
	t.Run("SetNX", func(t *testing.T) {
		key := "test:rest:lock:1"
		value := "lock-value"

		// Cleanup first
		client.Del(ctx, key)

		// First SetNX should succeed
		ok, err := client.SetNX(ctx, key, value, 10*time.Second)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if !ok {
			t.Error("Expected SetNX to succeed")
		}

		// Second SetNX should fail (key exists)
		ok, err = client.SetNX(ctx, key, "another-value", 10*time.Second)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if ok {
			t.Error("Expected SetNX to fail when key exists")
		}

		// Cleanup
		client.Del(ctx, key)
	})

	// Test 4: Exists
	t.Run("Exists", func(t *testing.T) {
		key := "test:rest:exists:1"

		// Cleanup first
		client.Del(ctx, key)

		// Key should not exist
		count, err := client.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}

		// Set key
		client.Set(ctx, key, "value", 10*time.Second)

		// Key should exist
		count, err = client.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}

		// Cleanup
		client.Del(ctx, key)
	})
}

// TestFactory_Environment tests factory function with different environments
func TestFactory_Environment(t *testing.T) {
	// Save original env
	originalEnv := os.Getenv("ENVIRONMENT")
	defer os.Setenv("ENVIRONMENT", originalEnv)

	t.Run("Development environment", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "development")

		// This will try to connect to local Redis
		// Skip if not available
		_, err := NewRedisClient()
		if err != nil {
			t.Logf("Development Redis not available (expected in CI): %v", err)
		}
	})

	t.Run("Production environment without credentials", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "production")

		// Temporarily clear Upstash credentials
		originalURL := os.Getenv("UPSTASH_REDIS_REST_URL")
		originalToken := os.Getenv("UPSTASH_REDIS_REST_TOKEN")
		os.Unsetenv("UPSTASH_REDIS_REST_URL")
		os.Unsetenv("UPSTASH_REDIS_REST_TOKEN")

		_, err := NewRedisClient()
		if err == nil {
			t.Error("Expected error when Upstash credentials not set")
		}

		// Restore credentials
		if originalURL != "" {
			os.Setenv("UPSTASH_REDIS_REST_URL", originalURL)
		}
		if originalToken != "" {
			os.Setenv("UPSTASH_REDIS_REST_TOKEN", originalToken)
		}
	})

	t.Run("Unknown environment", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "staging")

		_, err := NewRedisClient()
		if err == nil {
			t.Error("Expected error for unknown environment")
		}
	})
}
