-- Add route_id to orders table for analytics
ALTER TABLE orders ADD COLUMN IF NOT EXISTS route_id UUID;
CREATE INDEX IF NOT EXISTS idx_orders_route_id ON orders(route_id);
