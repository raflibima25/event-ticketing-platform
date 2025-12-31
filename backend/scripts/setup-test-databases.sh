#!/bin/bash

# Setup Test Databases for Testing
# Run after PostgreSQL is installed/upgraded

set -e

echo "🧪 Setting Up Test Databases"
echo "============================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check PostgreSQL is running
echo "📊 Step 1: Checking PostgreSQL status..."
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo "❌ PostgreSQL is not running on port 5432"
    echo "Start PostgreSQL with: sudo service postgresql start"
    exit 1
fi

echo "${GREEN}✅ PostgreSQL is running${NC}"
psql --version
echo ""

# Create test databases
echo "📦 Step 2: Creating test databases..."

# Drop existing test databases if they exist (optional)
echo "${YELLOW}Dropping existing test databases (if any)...${NC}"
sudo -u postgres dropdb --if-exists ticketing_test 2>/dev/null || true
sudo -u postgres dropdb --if-exists payment_test 2>/dev/null || true

# Create fresh test databases
echo "Creating ticketing_test database..."
sudo -u postgres createdb ticketing_test

echo "Creating payment_test database..."
sudo -u postgres createdb payment_test

echo "${GREEN}✅ Test databases created${NC}"
echo ""

# Verify databases exist
echo "📋 Step 3: Verifying databases..."
sudo -u postgres psql -l | grep -E "ticketing_test|payment_test"
echo ""

# Apply migrations
echo "🔄 Step 4: Applying migrations..."

cd "/mnt/c/Rafli Bima Pratandra (D)/Development/event-ticketing-platform/backend"

# Ticketing service migrations
echo "Applying ticketing service migrations..."
sudo -u postgres psql -d ticketing_test -f migrations/000001_initial_schema.up.sql
sudo -u postgres psql -d ticketing_test -f migrations/000002_update_ticketing_schema.up.sql

# Payment service migrations
echo "Applying payment service migrations..."
sudo -u postgres psql -d payment_test -f migrations/000001_initial_schema.up.sql
sudo -u postgres psql -d payment_test -f migrations/000003_create_payment_tables.up.sql

echo "${GREEN}✅ Migrations applied${NC}"
echo ""

# Verify tables
echo "✅ Step 5: Verifying tables..."

echo ""
echo "Tables in ticketing_test:"
sudo -u postgres psql -d ticketing_test -c "\dt" | grep -v "List of relations" | grep -v "(.*rows)" || echo "No tables found"

echo ""
echo "Tables in payment_test:"
sudo -u postgres psql -d payment_test -c "\dt" | grep -v "List of relations" | grep -v "(.*rows)" || echo "No tables found"

echo ""

# Set permissions (if needed)
echo "🔐 Step 6: Setting permissions..."
sudo -u postgres psql -d ticketing_test -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;"
sudo -u postgres psql -d payment_test -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;"

echo "${GREEN}✅ Permissions set${NC}"
echo ""

# Export environment variable
echo "📝 Step 7: Setting environment variables..."
echo ""
echo "Add this to your ~/.bashrc or ~/.zshrc:"
echo ""
echo "export TEST_DATABASE_URL=\"postgres://postgres:postgres@localhost:5432/ticketing_test\""
echo ""

# Summary
echo "======================================"
echo "${GREEN}✅ Test Databases Setup Complete!${NC}"
echo "======================================"
echo ""
echo "Summary:"
echo "  ✅ ticketing_test - Ready for testing"
echo "  ✅ payment_test - Ready for testing"
echo "  📊 Migrations applied"
echo "  🔐 Permissions configured"
echo ""
echo "Next Steps:"
echo "  1. Export TEST_DATABASE_URL (see above)"
echo "  2. Run tests:"
echo ""
echo "     cd backend/services/ticketing-service"
echo "     go test -v ./internal/repository -run TestConcurrentPurchase"
echo ""
echo "     cd backend/services/payment-service"
echo "     go test -v ./internal/repository -run TestWebhook"
echo ""
