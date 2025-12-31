package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
// Uses environment variable TEST_DATABASE_URL or falls back to default
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	// Get database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		// Default test database URL
		dbURL = "postgres://postgres:postgres@localhost:5432/ticketing_test?sslmode=disable"
		t.Logf("⚠️  TEST_DATABASE_URL not set, using default: %s", dbURL)
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v\nMake sure PostgreSQL is running and TEST_DATABASE_URL is set", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("Failed to ping test database: %v", err)
	}

	t.Logf("✅ Test database connected successfully")

	return db
}

// CleanupTestDB closes database connection and cleans up test data
func CleanupTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()

	if db != nil {
		db.Close()
		t.Logf("✅ Test database connection closed")
	}
}

// TruncateTables truncates specified tables for clean test state
func TruncateTables(t *testing.T, db *sqlx.DB, tables ...string) {
	t.Helper()

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		if _, err := db.Exec(query); err != nil {
			t.Logf("⚠️  Warning: Failed to truncate table %s: %v", table, err)
		}
	}

	t.Logf("✅ Truncated tables: %v", tables)
}

// CreateTestTicketTier creates a test ticket tier for testing
func CreateTestTicketTier(t *testing.T, db *sqlx.DB, eventID string, quota int) string {
	t.Helper()

	tierID := uuid.New().String()

	query := `
		INSERT INTO ticket_tiers (id, event_id, name, price, quota, sold_count, max_per_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	_, err := db.Exec(query, tierID, eventID, "Test Tier", 100000.0, quota, 0, 5)
	if err != nil {
		t.Fatalf("Failed to create test ticket tier: %v", err)
	}

	t.Logf("✅ Created test ticket tier: %s (quota: %d)", tierID, quota)

	return tierID
}

// CreateTestEvent creates a test event for testing
func CreateTestEvent(t *testing.T, db *sqlx.DB) string {
	t.Helper()

	eventID := uuid.New().String()
	organizerID := uuid.New().String()
	slug := fmt.Sprintf("test-event-%d", time.Now().UnixNano())

	// First create the organizer user (required by foreign key)
	userQuery := `
		INSERT INTO users (id, email, full_name, phone, role, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`
	_, _ = db.Exec(userQuery,
		organizerID,
		fmt.Sprintf("organizer-%s@test.com", organizerID[:8]),
		"Test Organizer",
		"081234567890",
		"organizer",
		"$2a$10$test.hash",
	)

	// Create the event
	query := `
		INSERT INTO events (id, organizer_id, title, slug, description, category,
		                   location, start_date, end_date, timezone, status,
		                   created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
	`

	startDate := time.Now().Add(24 * time.Hour)
	endDate := startDate.Add(3 * time.Hour)

	_, err := db.Exec(query,
		eventID,
		organizerID,
		"Test Event",
		slug,
		"Test event description",
		"music",
		"Test Location",
		startDate,
		endDate,
		"Asia/Jakarta",
		"published",
	)

	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	t.Logf("✅ Created test event: %s", eventID)

	return eventID
}
