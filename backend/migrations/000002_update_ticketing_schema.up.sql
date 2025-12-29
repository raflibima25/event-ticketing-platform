-- Add missing columns to orders table for ticketing service
ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS service_fee DECIMAL(12,2) DEFAULT 0 CHECK (service_fee >= 0),
  ADD COLUMN IF NOT EXISTS grand_total DECIMAL(12,2) DEFAULT 0 CHECK (grand_total >= 0);

-- Update orders status enum to match ticketing service
ALTER TABLE orders
  DROP CONSTRAINT IF EXISTS orders_status_check;

ALTER TABLE orders
  ADD CONSTRAINT orders_status_check CHECK (status IN ('reserved', 'paid', 'expired', 'cancelled', 'completed', 'refunded'));

-- Update order_items table structure for ticketing service
-- Remove attendee fields (moved to tickets table)
ALTER TABLE order_items
  DROP COLUMN IF EXISTS attendee_name,
  DROP COLUMN IF EXISTS attendee_email;

-- Drop subtotal if it exists as generated column
ALTER TABLE order_items DROP COLUMN IF EXISTS subtotal;

-- Add missing columns (subtotal as regular column, not generated)
ALTER TABLE order_items
  ADD COLUMN IF NOT EXISTS subtotal DECIMAL(12,2) DEFAULT 0 CHECK (subtotal >= 0),
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

-- Update tickets table to match ticketing service expectations
ALTER TABLE tickets
  DROP CONSTRAINT IF EXISTS tickets_order_item_id_fkey;

-- Add new columns to tickets table
ALTER TABLE tickets
  ADD COLUMN IF NOT EXISTS order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS event_id UUID REFERENCES events(id),
  ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id),
  ADD COLUMN IF NOT EXISTS ticket_tier_id UUID REFERENCES ticket_tiers(id),
  ADD COLUMN IF NOT EXISTS ticket_number VARCHAR(50) UNIQUE,
  ADD COLUMN IF NOT EXISTS qr_data TEXT,
  ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'valid' CHECK (status IN ('valid', 'used', 'cancelled', 'expired')),
  ADD COLUMN IF NOT EXISTS validated_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

-- Rename/update QR code column
ALTER TABLE tickets
  ALTER COLUMN qr_code DROP NOT NULL,
  ALTER COLUMN qr_code TYPE TEXT;

-- Drop old validation columns if they exist
ALTER TABLE tickets
  DROP COLUMN IF EXISTS validation_hash,
  DROP COLUMN IF EXISTS pdf_url,
  DROP COLUMN IF EXISTS is_validated,
  DROP COLUMN IF EXISTS validated_by;

-- Create indexes for tickets
CREATE INDEX IF NOT EXISTS idx_tickets_order ON tickets(order_id);
CREATE INDEX IF NOT EXISTS idx_tickets_event ON tickets(event_id);
CREATE INDEX IF NOT EXISTS idx_tickets_user ON tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_tier ON tickets(ticket_tier_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tickets_number ON tickets(ticket_number);

-- Add updated_at trigger for tickets
DROP TRIGGER IF EXISTS update_tickets_updated_at ON tickets;
CREATE TRIGGER update_tickets_updated_at
  BEFORE UPDATE ON tickets
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add payment confirmation columns to orders
ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS payment_id VARCHAR(255),
  ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50),
  ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;

-- Create index for reservation cleanup
CREATE INDEX IF NOT EXISTS idx_orders_cleanup ON orders(status, reservation_expires_at)
  WHERE status = 'reserved' AND reservation_expires_at IS NOT NULL;

-- Update grand_total for existing orders (backfill)
UPDATE orders
SET grand_total = total_amount + platform_fee + COALESCE(service_fee, 0)
WHERE grand_total = 0 OR grand_total IS NULL;
