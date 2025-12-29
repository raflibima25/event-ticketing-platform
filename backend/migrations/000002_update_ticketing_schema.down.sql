-- Rollback ticketing service schema changes

-- Drop indexes
DROP INDEX IF EXISTS idx_orders_cleanup;
DROP INDEX IF EXISTS idx_tickets_number;
DROP INDEX IF EXISTS idx_tickets_status;
DROP INDEX IF EXISTS idx_tickets_tier;
DROP INDEX IF EXISTS idx_tickets_order;

-- Remove new columns from orders
ALTER TABLE orders
  DROP COLUMN IF EXISTS service_fee,
  DROP COLUMN IF EXISTS grand_total,
  DROP COLUMN IF EXISTS payment_id,
  DROP COLUMN IF EXISTS paid_at;

-- Restore original orders status constraint
ALTER TABLE orders
  DROP CONSTRAINT IF EXISTS orders_status_check;

ALTER TABLE orders
  ADD CONSTRAINT orders_status_check CHECK (status IN ('reserved', 'pending_payment', 'paid', 'expired', 'refunded', 'cancelled'));

-- Remove new columns from tickets
ALTER TABLE tickets
  DROP COLUMN IF EXISTS order_id,
  DROP COLUMN IF EXISTS ticket_tier_id,
  DROP COLUMN IF EXISTS ticket_number,
  DROP COLUMN IF EXISTS qr_data,
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS validated_at;

-- Restore original tickets columns
ALTER TABLE tickets
  ADD COLUMN IF NOT EXISTS validation_hash VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS pdf_url VARCHAR(500),
  ADD COLUMN IF NOT EXISTS is_validated BOOLEAN DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS validated_by UUID REFERENCES users(id);

ALTER TABLE tickets
  ALTER COLUMN qr_code SET NOT NULL,
  ALTER COLUMN qr_code TYPE VARCHAR(255);

-- Remove subtotal from order_items
ALTER TABLE order_items
  DROP COLUMN IF EXISTS subtotal;

-- Restore attendee fields to order_items
ALTER TABLE order_items
  ADD COLUMN IF NOT EXISTS attendee_name VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS attendee_email VARCHAR(255) NOT NULL DEFAULT '';
