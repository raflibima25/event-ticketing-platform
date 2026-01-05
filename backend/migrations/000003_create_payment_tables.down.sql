-- Drop triggers
DROP TRIGGER IF EXISTS update_payment_transactions_updated_at ON payment_transactions;
DROP FUNCTION IF EXISTS update_payment_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_refunds_status;
DROP INDEX IF EXISTS idx_refunds_payment;
DROP INDEX IF EXISTS idx_refunds_order;

-- Note: webhook_events indexes managed by migration 000001, don't drop
DROP INDEX IF EXISTS idx_payment_transactions_status;
DROP INDEX IF EXISTS idx_payment_transactions_invoice;
DROP INDEX IF EXISTS idx_payment_transactions_external;
DROP INDEX IF EXISTS idx_payment_transactions_order;

-- Drop tables
DROP TABLE IF EXISTS refunds;
-- Note: webhook_events table created in migration 000001, don't drop, just revert changes
ALTER TABLE webhook_events DROP CONSTRAINT IF EXISTS webhook_events_status_check;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS status;

DROP TABLE IF EXISTS payment_transactions;

-- Restore old payments table
ALTER TABLE IF EXISTS payments_legacy RENAME TO payments;
