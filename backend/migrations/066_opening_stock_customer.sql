ALTER TABLE inventory_items
    ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_inventory_items_customer
    ON inventory_items(customer_id);
