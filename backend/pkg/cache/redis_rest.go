package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RESTRedisClient implements RedisClient using Upstash REST API
// Used for production deployment on Cloud Run (serverless-friendly)
type RESTRedisClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// upstashResponse represents the standard response from Upstash REST API
type upstashResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// NewRESTRedisClient creates a new REST-based Redis client
// This is used for production connecting to Upstash
func NewRESTRedisClient(url, token string) (RedisClient, error) {
	if url == "" || token == "" {
		return nil, fmt.Errorf("UPSTASH_REDIS_REST_URL and UPSTASH_REDIS_REST_TOKEN must be set")
	}

	client := &RESTRedisClient{
		baseURL: strings.TrimSuffix(url, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Test connection with PING
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Upstash Redis (REST): %w", err)
	}

	return client, nil
}

// executeCommand executes a Redis command via Upstash REST API
func (c *RESTRedisClient) executeCommand(ctx context.Context, command string, args ...interface{}) (interface{}, error) {
	// Build request body
	requestBody := []interface{}{command}
	requestBody = append(requestBody, args...)

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result upstashResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("redis error: %s", result.Error)
	}

	return result.Result, nil
}

// Set stores a key-value pair with expiration
func (c *RESTRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if expiration > 0 {
		// Use SETEX for expiration in seconds
		_, err := c.executeCommand(ctx, "SETEX", key, int(expiration.Seconds()), value)
		return err
	}
	// Use SET for no expiration
	_, err := c.executeCommand(ctx, "SET", key, value)
	return err
}

// Get retrieves a value by key
func (c *RESTRedisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := c.executeCommand(ctx, "GET", key)
	if err != nil {
		return "", err
	}

	// Upstash returns nil for non-existent keys
	if result == nil {
		return "", nil
	}

	// Type assertion to string
	if str, ok := result.(string); ok {
		return str, nil
	}

	return fmt.Sprintf("%v", result), nil
}

// Del deletes one or more keys
func (c *RESTRedisClient) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	_, err := c.executeCommand(ctx, "DEL", args...)
	return err
}

// SetNX sets a key only if it doesn't exist (for distributed locks)
func (c *RESTRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	// Use SET with NX and EX options
	if expiration > 0 {
		result, err := c.executeCommand(ctx, "SET", key, value, "NX", "EX", int(expiration.Seconds()))
		if err != nil {
			return false, err
		}
		// SET NX returns "OK" if successful, nil if key exists
		return result != nil, nil
	}

	// Use SETNX without expiration
	result, err := c.executeCommand(ctx, "SETNX", key, value)
	if err != nil {
		return false, err
	}

	// SETNX returns 1 if set, 0 if not set
	if num, ok := result.(float64); ok {
		return num == 1, nil
	}

	return false, fmt.Errorf("unexpected SETNX response type: %T", result)
}

// Exists checks if key exists
func (c *RESTRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	result, err := c.executeCommand(ctx, "EXISTS", args...)
	if err != nil {
		return 0, err
	}

	// Upstash returns number as float64
	if num, ok := result.(float64); ok {
		return int64(num), nil
	}

	return 0, fmt.Errorf("unexpected EXISTS response type: %T", result)
}

// Expire sets expiration on existing key
func (c *RESTRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := c.executeCommand(ctx, "EXPIRE", key, int(expiration.Seconds()))
	return err
}

// Ping checks connection health
func (c *RESTRedisClient) Ping(ctx context.Context) error {
	result, err := c.executeCommand(ctx, "PING")
	if err != nil {
		return err
	}

	// PING should return "PONG"
	if str, ok := result.(string); ok && str == "PONG" {
		return nil
	}

	return fmt.Errorf("unexpected PING response: %v", result)
}

// Close closes the connection
func (c *RESTRedisClient) Close() error {
	// HTTP client doesn't need explicit close for REST API
	// No persistent connection to close
	return nil
}
