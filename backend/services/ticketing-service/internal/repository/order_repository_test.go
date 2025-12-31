package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/payload/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReservationTimeout tests that expired reservations are properly identified
// This is IMPORTANT to prevent inventory from being locked forever
func TestReservationTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "orders", "events")

	repo := NewOrderRepository(db)
	ctx := context.Background()

	// Create test event
	eventID := CreateTestEvent(t, db)

	// Create 3 orders with different expiration times
	now := time.Now()

	// Order 1: Expired 5 minutes ago (should be returned)
	expiredOrder1 := createTestOrder(t, db, eventID, now.Add(-5*time.Minute))

	// Order 2: Expired 1 minute ago (should be returned)
	expiredOrder2 := createTestOrder(t, db, eventID, now.Add(-1*time.Minute))

	// Order 3: Expires in 10 minutes (should NOT be returned)
	futureOrder := createTestOrder(t, db, eventID, now.Add(10*time.Minute))

	// Order 4: Already paid (should NOT be returned even if expired)
	paidOrder := createTestOrder(t, db, eventID, now.Add(-10*time.Minute))
	_, err := db.Exec("UPDATE orders SET status = $1 WHERE id = $2", entity.OrderStatusPaid, paidOrder)
	require.NoError(t, err)

	t.Logf("✅ Created 4 test orders:")
	t.Logf("   Order 1: Expired 5 min ago (reserved)")
	t.Logf("   Order 2: Expired 1 min ago (reserved)")
	t.Logf("   Order 3: Expires in 10 min (reserved)")
	t.Logf("   Order 4: Expired but already paid")

	// Get expired reservations
	expiredOrders, err := repo.GetExpiredReservations(ctx)
	require.NoError(t, err, "Failed to get expired reservations")

	t.Logf("\n📊 Found %d expired reservations", len(expiredOrders))

	// CRITICAL ASSERTIONS
	// Should return exactly 2 expired orders (Order 1 and Order 2)
	assert.Equal(t, 2, len(expiredOrders), "Should return exactly 2 expired reservations")

	// Verify the expired orders are the correct ones
	expiredIDs := make(map[string]bool)
	for _, order := range expiredOrders {
		expiredIDs[order.ID] = true
		assert.Equal(t, entity.OrderStatusReserved, order.Status, "Expired order should have reserved status")
		assert.True(t, order.ReservationExpiresAt.Before(now), "Reservation should be expired")

		t.Logf("   ✓ Found expired order: %s (expired at: %s)",
			order.ID, order.ReservationExpiresAt.Format("15:04:05"))
	}

	// Verify correct orders were returned
	assert.True(t, expiredIDs[expiredOrder1], "Order 1 (expired 5 min ago) should be returned")
	assert.True(t, expiredIDs[expiredOrder2], "Order 2 (expired 1 min ago) should be returned")
	assert.False(t, expiredIDs[futureOrder], "Future order should NOT be returned")
	assert.False(t, expiredIDs[paidOrder], "Paid order should NOT be returned even if expired")

	t.Logf("\n✅ RESERVATION TIMEOUT TEST PASSED!")
	t.Logf("   ✓ Expired reservations correctly identified")
	t.Logf("   ✓ Future reservations excluded")
	t.Logf("   ✓ Paid orders excluded")
}

// TestReservationTimeout_ReleaseInventory tests complete flow of releasing inventory
// This simulates the background worker that releases expired reservations
func TestReservationTimeout_ReleaseInventory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "orders", "order_items", "ticket_tiers", "events")

	orderRepo := NewOrderRepository(db)
	tierRepo := NewTicketTierRepository(db)
	ctx := context.Background()

	// Create test data
	eventID := CreateTestEvent(t, db)
	tierID := CreateTestTicketTier(t, db, eventID, 10) // 10 tickets available

	// Scenario: User reserved 3 tickets but didn't pay
	// 1. Create order with 3 tickets
	now := time.Now()
	expiredTime := now.Add(-5 * time.Minute) // Expired 5 minutes ago

	orderID := createTestOrder(t, db, eventID, expiredTime)
	quantity := 3

	// 2. Simulate reservation: Update sold_count
	tx, _ := db.DB.BeginTx(ctx, nil)
	err := tierRepo.UpdateSoldCount(ctx, tx, tierID, quantity)
	require.NoError(t, err)
	tx.Commit()

	// Verify sold_count is 3
	tier, _ := tierRepo.GetByID(ctx, tierID)
	assert.Equal(t, 3, tier.SoldCount, "Sold count should be 3 after reservation")
	t.Logf("✅ Reservation created: 3 tickets reserved")

	// 3. Simulate background worker: Get expired reservations
	expiredOrders, err := orderRepo.GetExpiredReservations(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(expiredOrders), "Should find 1 expired order")

	t.Logf("✅ Found expired reservation: %s", expiredOrders[0].ID)

	// 4. Release inventory (what the background worker would do)
	tx, _ = db.DB.BeginTx(ctx, nil)

	// Cancel order
	_, err = db.Exec("UPDATE orders SET status = $1 WHERE id = $2", entity.OrderStatusCancelled, orderID)
	require.NoError(t, err)

	// Release tickets
	err = tierRepo.ReleaseSoldCount(ctx, tx, tierID, quantity)
	require.NoError(t, err)

	tx.Commit()

	// CRITICAL ASSERTIONS
	// 5. Verify inventory was released
	tier, _ = tierRepo.GetByID(ctx, tierID)
	assert.Equal(t, 0, tier.SoldCount, "CRITICAL: Sold count should return to 0 after release")

	// 6. Verify order status updated
	var orderStatus string
	db.Get(&orderStatus, "SELECT status FROM orders WHERE id = $1", orderID)
	assert.Equal(t, entity.OrderStatusCancelled, orderStatus, "Order should be cancelled")

	t.Logf("\n✅ RESERVATION RELEASE TEST PASSED!")
	t.Logf("   ✓ Expired reservation detected")
	t.Logf("   ✓ Inventory released (sold_count: 3 → 0)")
	t.Logf("   ✓ Order status updated to cancelled")
}

// TestReservationTimeout_NoDoubleRelease tests that inventory isn't released twice
func TestReservationTimeout_NoDoubleRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "orders", "events")

	repo := NewOrderRepository(db)
	ctx := context.Background()

	eventID := CreateTestEvent(t, db)
	now := time.Now()

	// Create expired order
	orderID := createTestOrder(t, db, eventID, now.Add(-10*time.Minute))

	// First release: Get expired reservations
	expiredOrders, _ := repo.GetExpiredReservations(ctx)
	assert.Equal(t, 1, len(expiredOrders))

	// Mark as cancelled
	db.Exec("UPDATE orders SET status = $1 WHERE id = $2", entity.OrderStatusCancelled, orderID)

	// Second attempt: Should not return cancelled order
	expiredOrders, _ = repo.GetExpiredReservations(ctx)
	assert.Equal(t, 0, len(expiredOrders), "Should not return already cancelled order")

	t.Logf("✅ Double release prevention works correctly")
}

// Helper function to create test order
func createTestOrder(t *testing.T, db *sqlx.DB, eventID string, expiresAt time.Time) string {
	t.Helper()

	orderID := uuid.New().String()
	userID := uuid.New().String()

	// First create the user (required by foreign key)
	userQuery := `
		INSERT INTO users (id, email, full_name, phone, role, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`
	_, _ = db.Exec(userQuery,
		userID,
		fmt.Sprintf("user-%s@test.com", userID[:8]),
		"Test User",
		"081234567890",
		"customer",
		"$2a$10$test.hash",
	)

	// Create the order
	query := `
		INSERT INTO orders (
			id, user_id, event_id, total_amount, platform_fee, service_fee,
			grand_total, status, reservation_expires_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`

	_, err := db.Exec(query,
		orderID,
		userID,
		eventID,
		100000.0,   // total_amount
		5000.0,     // platform_fee
		2500.0,     // service_fee
		107500.0,   // grand_total
		entity.OrderStatusReserved,
		expiresAt,
	)

	require.NoError(t, err, "Failed to create test order")
	return orderID
}
