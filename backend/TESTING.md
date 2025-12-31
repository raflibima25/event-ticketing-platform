# Testing Guide - Event Ticketing Platform

## Critical Tests untuk MVP

Dokumen ini menjelaskan 3 **critical tests** yang WAJIB pass sebelum production deployment.

---

## Prerequisites

### 1. PostgreSQL Test Database

Buat 2 test database untuk testing:

```bash
# For ticketing-service tests
createdb ticketing_test

# For payment-service tests
createdb payment_test
```

**Schema Setup:**
```bash
# Apply migrations to test databases
cd backend/services/ticketing-service
psql -d ticketing_test < ../../migrations/ticketing/*.sql

cd backend/services/payment-service
psql -d payment_test < ../../migrations/payment/*.sql
```

### 2. Environment Variables (Optional)

```bash
# Default: postgres://postgres:postgres@localhost:5432/ticketing_test
export TEST_DATABASE_URL="postgres://user:pass@localhost:5432/ticketing_test"
```

---

## Running Critical Tests

### 🔥 Test 1: Overselling Prevention (MOST CRITICAL)

**What it tests:**
- Concurrent ticket purchases don't cause overselling
- Row-level locking (`SELECT FOR UPDATE`) works correctly
- Database constraints prevent quota violations

**Run:**
```bash
cd backend/services/ticketing-service
go test -v ./internal/repository -run TestConcurrentPurchase_NoOverselling
```

**Expected output:**
```
=== RUN   TestConcurrentPurchase_NoOverselling
🔥 Starting concurrent purchase test: 20 buyers competing for 10 tickets
✅ Buyer 0: Successfully purchased ticket
✅ Buyer 1: Successfully purchased ticket
...
📊 RESULTS:
   Success: 10
   Failed:  10
✅ OVERSELLING PREVENTION TEST PASSED!
   ✓ No overselling occurred
   ✓ Row-level locking works correctly
   ✓ Database constraints enforced
--- PASS: TestConcurrentPurchase_NoOverselling (0.15s)
```

**If this test fails:** 🚨 **DO NOT DEPLOY TO PRODUCTION** - Overselling akan terjadi!

---

### 💰 Test 2: Payment Webhook Idempotency (CRITICAL)

**What it tests:**
- Duplicate webhooks are rejected
- No double ticket generation
- Database unique constraint on webhook_id works

**Run:**
```bash
cd backend/services/payment-service
go test -v ./internal/repository -run TestWebhook_Idempotency
go test -v ./internal/repository -run TestConcurrentWebhook_Idempotency
```

**Expected output:**
```
=== RUN   TestWebhook_Idempotency
✅ First webhook created successfully
✅ Duplicate webhook correctly rejected
✅ WEBHOOK IDEMPOTENCY TEST PASSED!
--- PASS: TestWebhook_Idempotency (0.05s)

=== RUN   TestConcurrentWebhook_Idempotency
🔥 Testing concurrent webhook handling: 10 attempts
📊 RESULTS:
   Success:    1
   Duplicates: 9
✅ CONCURRENT WEBHOOK IDEMPOTENCY TEST PASSED!
--- PASS: TestConcurrentWebhook_Idempotency (0.08s)
```

**If this test fails:** 🚨 Risk of double billing or duplicate tickets!

---

### ⏰ Test 3: Reservation Timeout (IMPORTANT)

**What it tests:**
- Expired reservations are correctly identified
- Inventory is released after timeout
- No double-release of inventory

**Run:**
```bash
cd backend/services/ticketing-service
go test -v ./internal/repository -run TestReservationTimeout
```

**Expected output:**
```
=== RUN   TestReservationTimeout
✅ Created 4 test orders
📊 Found 2 expired reservations
✅ RESERVATION TIMEOUT TEST PASSED!
   ✓ Expired reservations correctly identified
   ✓ Future reservations excluded
   ✓ Paid orders excluded
--- PASS: TestReservationTimeout (0.10s)
```

---

## Run All Critical Tests

```bash
# Ticketing Service Tests
cd backend/services/ticketing-service
go test -v ./internal/repository -run "TestConcurrentPurchase|TestReservationTimeout"

# Payment Service Tests
cd backend/services/payment-service
go test -v ./internal/repository -run "TestWebhook"
```

---

## Quick Test (No Database)

Skip integration tests jika tidak ada database:

```bash
go test -v -short ./...
```

---

## Test Coverage

Generate coverage report:

```bash
cd backend/services/ticketing-service
go test ./internal/repository -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Critical Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Setup test databases
        run: |
          createdb -h localhost -U postgres ticketing_test
          createdb -h localhost -U postgres payment_test
          psql -h localhost -U postgres -d ticketing_test < migrations/ticketing/*.sql

      - name: Run critical tests
        env:
          TEST_DATABASE_URL: "postgres://postgres:postgres@localhost:5432/ticketing_test"
        run: |
          cd backend/services/ticketing-service
          go test -v ./internal/repository -run "TestConcurrentPurchase|TestReservationTimeout"
```

---

## Troubleshooting

### Error: "Failed to connect to test database"

**Solution:**
```bash
# Check PostgreSQL is running
pg_ctl status

# Create test databases
createdb ticketing_test
createdb payment_test

# Set correct connection string
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/ticketing_test"
```

### Error: "table does not exist"

**Solution:** Run migrations on test database
```bash
psql -d ticketing_test < migrations/*.sql
```

### Test hangs or timeouts

**Reason:** Deadlock in locking tests (ini seharusnya tidak terjadi!)

**Debug:**
```sql
-- Check for locks
SELECT * FROM pg_locks WHERE NOT granted;

-- Check for blocking queries
SELECT * FROM pg_stat_activity WHERE state = 'active';
```

---

## MVP Testing Checklist

Sebelum deploy ke production:

- [ ] ✅ `TestConcurrentPurchase_NoOverselling` - **MUST PASS**
- [ ] ✅ `TestUpdateSoldCount_DatabaseConstraintPreventsOverselling` - **MUST PASS**
- [ ] ✅ `TestWebhook_Idempotency` - **MUST PASS**
- [ ] ✅ `TestConcurrentWebhook_Idempotency` - **MUST PASS**
- [ ] ✅ `TestReservationTimeout` - **SHOULD PASS**
- [ ] Manual test: End-to-end ticket purchase flow
- [ ] Manual test: Payment webhook dengan Xendit sandbox
- [ ] Manual test: Reservation timeout (wait 15 minutes)

---

## Post-MVP: Additional Tests

Setelah launch, tambahkan:
- [ ] E2E tests (Playwright/Cypress)
- [ ] Load testing (K6) - 1000 concurrent users
- [ ] Security testing (OWASP ZAP)
- [ ] API integration tests
- [ ] Frontend unit tests

---

## Contact

Jika ada pertanyaan tentang testing:
- Create issue di GitHub
- Tag: `testing`, `critical-path`
