package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/raflibima25/event-ticketing-platform/backend/services/payment-service/internal/payload/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates a test database connection for payment service
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Get database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/payment_test?sslmode=disable"
		t.Logf("⚠️  TEST_DATABASE_URL not set, using default: %s", dbURL)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v\nMake sure PostgreSQL is running", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("Failed to ping test database: %v", err)
	}

	t.Logf("✅ Test database connected successfully")
	return db
}

// CleanupTestDB closes database connection
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if db != nil {
		db.Close()
		t.Logf("✅ Test database connection closed")
	}
}

// TruncateTables truncates specified tables
func TruncateTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		if _, err := db.Exec(query); err != nil {
			t.Logf("⚠️  Warning: Failed to truncate table %s: %v", table, err)
		}
	}
	t.Logf("✅ Truncated tables: %v", tables)
}

// TestWebhook_Idempotency tests that duplicate webhooks are rejected
// This is CRITICAL to prevent double ticket generation or double charges
func TestWebhook_Idempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "webhook_events")

	repo := NewWebhookRepository(db)
	ctx := context.Background()

	// Create a webhook event
	webhookID := "xendit-webhook-12345"
	webhook := &entity.WebhookEvent{
		WebhookID: webhookID,
		EventType: "invoice.paid",
		Payload:   `{"id":"12345","status":"PAID"}`,
		Status:    entity.WebhookStatusPending,
	}

	// First attempt: Should succeed
	err := repo.Create(ctx, webhook)
	require.NoError(t, err, "First webhook creation should succeed")
	t.Logf("✅ First webhook created successfully: %s", webhookID)

	// Second attempt with same webhook_id: Should fail with ErrDuplicateWebhook
	duplicateWebhook := &entity.WebhookEvent{
		WebhookID: webhookID,
		EventType: "invoice.paid",
		Payload:   `{"id":"12345","status":"PAID"}`,
		Status:    entity.WebhookStatusPending,
	}

	err = repo.Create(ctx, duplicateWebhook)

	// CRITICAL ASSERTION: Must return ErrDuplicateWebhook
	assert.Error(t, err, "CRITICAL: Duplicate webhook MUST be rejected")
	assert.Equal(t, ErrDuplicateWebhook, err, "Should return ErrDuplicateWebhook")

	t.Logf("✅ Duplicate webhook correctly rejected")

	// Verify only 1 webhook exists in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM webhook_events WHERE webhook_id = $1", webhookID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "CRITICAL: Only 1 webhook should exist in database")

	t.Logf("✅ WEBHOOK IDEMPOTENCY TEST PASSED!")
}

// TestConcurrentWebhook_Idempotency tests concurrent duplicate webhook handling
// Simulates Xendit sending the same webhook multiple times concurrently
func TestConcurrentWebhook_Idempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "webhook_events")

	repo := NewWebhookRepository(db)

	// Test: 10 concurrent attempts to create the same webhook
	webhookID := "xendit-webhook-concurrent-12345"
	concurrentAttempts := 10

	successCount := 0
	duplicateCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	t.Logf("🔥 Testing concurrent webhook handling: %d attempts", concurrentAttempts)

	for i := 0; i < concurrentAttempts; i++ {
		wg.Add(1)
		go func(attemptID int) {
			defer wg.Done()

			ctx := context.Background()
			webhook := &entity.WebhookEvent{
				WebhookID: webhookID,
				EventType: "invoice.paid",
				Payload:   fmt.Sprintf(`{"id":"%s","attempt":%d}`, webhookID, attemptID),
				Status:    entity.WebhookStatusPending,
			}

			err := repo.Create(ctx, webhook)

			mu.Lock()
			defer mu.Unlock()

			if err == nil {
				successCount++
				t.Logf("Attempt %d: ✅ Successfully created webhook", attemptID)
			} else if err == ErrDuplicateWebhook {
				duplicateCount++
				t.Logf("Attempt %d: ❌ Duplicate webhook rejected", attemptID)
			} else {
				t.Errorf("Attempt %d: Unexpected error: %v", attemptID, err)
			}
		}(i)
	}

	wg.Wait()

	// CRITICAL ASSERTIONS
	t.Logf("\n📊 RESULTS:")
	t.Logf("   Success:    %d", successCount)
	t.Logf("   Duplicates: %d", duplicateCount)

	// ASSERTION 1: Exactly 1 webhook should succeed
	assert.Equal(t, 1, successCount, "CRITICAL: Exactly 1 webhook creation should succeed")

	// ASSERTION 2: All others should be rejected as duplicates
	assert.Equal(t, concurrentAttempts-1, duplicateCount, "All other attempts should be rejected")

	// ASSERTION 3: Verify database state
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM webhook_events WHERE webhook_id = $1", webhookID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "CRITICAL: Only 1 webhook should exist in database")

	t.Logf("\n✅ CONCURRENT WEBHOOK IDEMPOTENCY TEST PASSED!")
	t.Logf("   ✓ No duplicate webhooks processed")
	t.Logf("   ✓ Database unique constraint works correctly")
	t.Logf("   ✓ Concurrent safety verified")
}

// TestWebhook_MarkAsProcessed tests marking webhook as processed
func TestWebhook_MarkAsProcessed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "webhook_events")

	repo := NewWebhookRepository(db)
	ctx := context.Background()

	// Create webhook
	webhookID := "xendit-webhook-process-test"
	webhook := &entity.WebhookEvent{
		WebhookID: webhookID,
		EventType: "invoice.paid",
		Payload:   `{"id":"12345"}`,
		Status:    entity.WebhookStatusPending,
	}

	err := repo.Create(ctx, webhook)
	require.NoError(t, err)

	// Verify initial status is pending
	retrieved, err := repo.GetByWebhookID(ctx, webhookID)
	require.NoError(t, err)
	assert.Equal(t, entity.WebhookStatusPending, retrieved.Status)
	assert.Nil(t, retrieved.ProcessedAt, "ProcessedAt should be nil initially")

	// Mark as processed
	err = repo.MarkAsProcessed(ctx, webhookID)
	require.NoError(t, err)

	// Verify status updated
	retrieved, err = repo.GetByWebhookID(ctx, webhookID)
	require.NoError(t, err)
	assert.Equal(t, entity.WebhookStatusProcessed, retrieved.Status)
	assert.NotNil(t, retrieved.ProcessedAt, "ProcessedAt should be set")

	t.Logf("✅ Webhook status correctly updated to processed")
}

// TestWebhook_GetByWebhookID tests retrieving webhook by ID
func TestWebhook_GetByWebhookID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "webhook_events")

	repo := NewWebhookRepository(db)
	ctx := context.Background()

	// Test: Get non-existent webhook
	_, err := repo.GetByWebhookID(ctx, "non-existent-webhook")
	assert.Equal(t, ErrWebhookNotFound, err, "Should return ErrWebhookNotFound for non-existent webhook")

	// Create webhook
	webhookID := "xendit-webhook-get-test"
	webhook := &entity.WebhookEvent{
		WebhookID: webhookID,
		EventType: "invoice.paid",
		Payload:   `{"id":"12345","amount":100000}`,
		Status:    entity.WebhookStatusPending,
	}

	err = repo.Create(ctx, webhook)
	require.NoError(t, err)

	// Retrieve webhook
	retrieved, err := repo.GetByWebhookID(ctx, webhookID)
	require.NoError(t, err)
	assert.Equal(t, webhookID, retrieved.WebhookID)
	assert.Equal(t, "invoice.paid", retrieved.EventType)
	assert.Contains(t, retrieved.Payload, `"id":"12345"`)

	t.Logf("✅ Webhook retrieval works correctly")
}
