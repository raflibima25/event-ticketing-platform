package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentPurchase_NoOverselling tests that concurrent purchases don't cause overselling
// This is the MOST CRITICAL test for the ticketing system
// MUST PASS before production deployment
func TestConcurrentPurchase_NoOverselling(t *testing.T) {
	// Skip if running in CI without database
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Clean tables
	TruncateTables(t, db, "ticket_tiers", "events")

	// Create test data
	eventID := CreateTestEvent(t, db)
	quota := 10 // Only 10 tickets available
	tierID := CreateTestTicketTier(t, db, eventID, quota)

	repo := NewTicketTierRepository(db)

	// Test: 20 concurrent goroutines trying to buy 1 ticket each
	concurrentBuyers := 20
	successCount := 0
	failCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	t.Logf("🔥 Starting concurrent purchase test: %d buyers competing for %d tickets", concurrentBuyers, quota)

	for i := 0; i < concurrentBuyers; i++ {
		wg.Add(1)
		go func(buyerID int) {
			defer wg.Done()

			// Simulate real-world purchase: check availability then buy
			ctx := context.Background()

			// Begin transaction
			tx, err := db.DB.BeginTx(ctx, nil)
			if err != nil {
				t.Errorf("Buyer %d: Failed to begin transaction: %v", buyerID, err)
				return
			}
			defer tx.Rollback()

			// CRITICAL: Lock the tier row (SELECT FOR UPDATE)
			tier, err := repo.GetByIDWithLock(ctx, tx, tierID)
			if err != nil {
				t.Logf("Buyer %d: Failed to get tier with lock: %v", buyerID, err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			// Check if ticket available
			if tier.SoldCount >= tier.Quota {
				t.Logf("Buyer %d: ❌ Sold out (sold: %d, quota: %d)", buyerID, tier.SoldCount, tier.Quota)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			// Simulate processing time (network latency, validation, etc.)
			time.Sleep(10 * time.Millisecond)

			// CRITICAL: Update sold count with database constraint check
			err = repo.UpdateSoldCount(ctx, tx, tierID, 1)
			if err != nil {
				t.Logf("Buyer %d: ❌ Failed to update sold count: %v", buyerID, err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				t.Logf("Buyer %d: ❌ Failed to commit: %v", buyerID, err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			t.Logf("Buyer %d: ✅ Successfully purchased ticket", buyerID)
			mu.Lock()
			successCount++
			mu.Unlock()
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// CRITICAL ASSERTIONS
	t.Logf("\n📊 RESULTS:")
	t.Logf("   Success: %d", successCount)
	t.Logf("   Failed:  %d", failCount)
	t.Logf("   Total:   %d", successCount+failCount)

	// ASSERTION 1: Exactly {quota} purchases should succeed
	assert.Equal(t, quota, successCount, "CRITICAL: Exactly %d purchases should succeed (no overselling!)", quota)

	// ASSERTION 2: Remaining buyers should fail
	expectedFails := concurrentBuyers - quota
	assert.Equal(t, expectedFails, failCount, "Expected %d buyers to fail", expectedFails)

	// ASSERTION 3: Verify database state
	var finalSoldCount int
	err := db.Get(&finalSoldCount, "SELECT sold_count FROM ticket_tiers WHERE id = $1", tierID)
	require.NoError(t, err, "Failed to query final sold count")

	assert.Equal(t, quota, finalSoldCount, "CRITICAL: Database sold_count must equal quota (no overselling in DB!)")

	// ASSERTION 4: Verify quota constraint is still intact
	var availableCount int
	err = db.Get(&availableCount, "SELECT (quota - sold_count) as available FROM ticket_tiers WHERE id = $1", tierID)
	require.NoError(t, err, "Failed to query available count")

	assert.Equal(t, 0, availableCount, "No tickets should be available after selling out")

	t.Logf("\n✅ OVERSELLING PREVENTION TEST PASSED!")
	t.Logf("   ✓ No overselling occurred")
	t.Logf("   ✓ Row-level locking works correctly")
	t.Logf("   ✓ Database constraints enforced")
}

// TestUpdateSoldCount_DatabaseConstraintPreventsOverselling tests database-level constraint
func TestUpdateSoldCount_DatabaseConstraintPreventsOverselling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "ticket_tiers", "events")

	// Create test data
	eventID := CreateTestEvent(t, db)
	quota := 5
	tierID := CreateTestTicketTier(t, db, eventID, quota)

	repo := NewTicketTierRepository(db)

	// First, sell 4 tickets (should succeed)
	ctx := context.Background()
	tx, _ := db.DB.BeginTx(ctx, nil)
	err := repo.UpdateSoldCount(ctx, tx, tierID, 4)
	require.NoError(t, err, "Selling 4 tickets should succeed")
	tx.Commit()

	// Now try to sell 2 more tickets (should fail - only 1 left)
	tx, _ = db.DB.BeginTx(ctx, nil)
	err = repo.UpdateSoldCount(ctx, tx, tierID, 2)
	tx.Rollback()

	// CRITICAL ASSERTION: This MUST fail
	assert.Error(t, err, "CRITICAL: Selling 2 tickets when only 1 available MUST fail")
	assert.Equal(t, ErrInsufficientQuota, err, "Should return ErrInsufficientQuota")

	// Verify sold count is still 4
	var soldCount int
	db.Get(&soldCount, "SELECT sold_count FROM ticket_tiers WHERE id = $1", tierID)
	assert.Equal(t, 4, soldCount, "Sold count should remain 4 after failed purchase")

	t.Logf("✅ Database constraint correctly prevents overselling")
}

// TestReleaseSoldCount tests releasing tickets (for cancellation/timeout)
func TestReleaseSoldCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	TruncateTables(t, db, "ticket_tiers", "events")

	// Create test data
	eventID := CreateTestEvent(t, db)
	quota := 10
	tierID := CreateTestTicketTier(t, db, eventID, quota)

	repo := NewTicketTierRepository(db)
	ctx := context.Background()

	// Sell 5 tickets
	tx, _ := db.DB.BeginTx(ctx, nil)
	err := repo.UpdateSoldCount(ctx, tx, tierID, 5)
	require.NoError(t, err)
	tx.Commit()

	// Verify sold count is 5
	tier, _ := repo.GetByID(ctx, tierID)
	assert.Equal(t, 5, tier.SoldCount)

	// Release 2 tickets (reservation timeout or cancellation)
	tx, _ = db.DB.BeginTx(ctx, nil)
	err = repo.ReleaseSoldCount(ctx, tx, tierID, 2)
	require.NoError(t, err)
	tx.Commit()

	// Verify sold count is now 3
	tier, _ = repo.GetByID(ctx, tierID)
	assert.Equal(t, 3, tier.SoldCount, "After releasing 2 tickets, sold_count should be 3")

	// Test: Release more than sold (should not go negative)
	tx, _ = db.DB.BeginTx(ctx, nil)
	err = repo.ReleaseSoldCount(ctx, tx, tierID, 10)
	require.NoError(t, err)
	tx.Commit()

	tier, _ = repo.GetByID(ctx, tierID)
	assert.Equal(t, 0, tier.SoldCount, "Sold count should not go below 0")

	t.Logf("✅ Release sold count works correctly")
}

// BenchmarkConcurrentPurchase benchmarks concurrent purchase performance
func BenchmarkConcurrentPurchase(b *testing.B) {
	db := SetupTestDB(&testing.T{})
	defer CleanupTestDB(&testing.T{}, db)

	eventID := CreateTestEvent(&testing.T{}, db)
	tierID := CreateTestTicketTier(&testing.T{}, db, eventID, 1000000) // Large quota for benchmarking

	repo := NewTicketTierRepository(db)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tx, _ := db.DB.BeginTx(ctx, nil)
		_, _ = repo.GetByIDWithLock(ctx, tx, tierID)
		_ = repo.UpdateSoldCount(ctx, tx, tierID, 1)
		tx.Commit()
	}
}
