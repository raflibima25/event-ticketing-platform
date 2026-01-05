-- Rename old payments table to payments_legacy (from migration 000001)
ALTER TABLE IF EXISTS payments RENAME TO payments_legacy;

-- Create payment_transactions table (new payment tracking table)
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    external_id VARCHAR(255) UNIQUE NOT NULL,
    invoice_id VARCHAR(255) UNIQUE,
    invoice_url TEXT,
    amount DECIMAL(12,2) NOT NULL,
    payment_method VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT payment_transactions_amount_check CHECK (amount >= 0),
    CONSTRAINT payment_transactions_status_check CHECK (status IN ('pending', 'paid', 'expired', 'failed'))
);

-- Create indexes for payment_transactions
CREATE INDEX IF NOT EXISTS idx_payment_transactions_order ON payment_transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_external ON payment_transactions(external_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_invoice ON payment_transactions(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions(status);

-- Update webhook_events table (add status column to existing table from migration 000001)
-- First, add status column if it doesn't exist
ALTER TABLE webhook_events
  ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending';

-- Add constraint for status if it doesn't exist (drop first to avoid duplicate)
ALTER TABLE webhook_events
  DROP CONSTRAINT IF EXISTS webhook_events_status_check;

ALTER TABLE webhook_events
  ADD CONSTRAINT webhook_events_status_check CHECK (status IN ('pending', 'processed', 'failed'));

-- Migrate data: set status based on processed column
UPDATE webhook_events
SET status = CASE
  WHEN processed = true THEN 'processed'
  ELSE 'pending'
END
WHERE status = 'pending' OR status IS NULL;

-- Create indexes for webhook_events (safe to run even if they exist)
CREATE INDEX IF NOT EXISTS idx_webhook_events_webhook ON webhook_events(webhook_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_status ON webhook_events(status);
CREATE INDEX IF NOT EXISTS idx_webhook_events_type ON webhook_events(event_type);

-- Create refunds table
CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    payment_transaction_id UUID REFERENCES payment_transactions(id),
    amount DECIMAL(12,2) NOT NULL,
    reason VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending',
    disbursement_id VARCHAR(255),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT refunds_amount_check CHECK (amount >= 0),
    CONSTRAINT refunds_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

-- Create indexes for refunds
CREATE INDEX IF NOT EXISTS idx_refunds_order ON refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_refunds_payment ON refunds(payment_transaction_id);
CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status);

-- Add updated_at trigger for payment_transactions
CREATE OR REPLACE FUNCTION update_payment_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_payment_transactions_updated_at ON payment_transactions;
CREATE TRIGGER update_payment_transactions_updated_at
    BEFORE UPDATE ON payment_transactions
    FOR EACH ROW EXECUTE FUNCTION update_payment_updated_at();
