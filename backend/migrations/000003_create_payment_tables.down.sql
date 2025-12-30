-- Drop triggers
DROP TRIGGER IF EXISTS update_payment_transactions_updated_at ON payment_transactions;
DROP FUNCTION IF EXISTS update_payment_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_refunds_status;
DROP INDEX IF EXISTS idx_refunds_payment;
DROP INDEX IF EXISTS idx_refunds_order;

DROP INDEX IF EXISTS idx_webhook_events_type;
DROP INDEX IF EXISTS idx_webhook_events_status;
DROP INDEX IF EXISTS idx_webhook_events_webhook;

DROP INDEX IF EXISTS idx_payment_transactions_status;
DROP INDEX IF EXISTS idx_payment_transactions_invoice;
DROP INDEX IF EXISTS idx_payment_transactions_external;
DROP INDEX IF EXISTS idx_payment_transactions_order;

-- Drop tables
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS webhook_events;
DROP TABLE IF EXISTS payment_transactions;
