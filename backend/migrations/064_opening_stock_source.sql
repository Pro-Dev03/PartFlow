ALTER TABLE inventory_movements
    ADD COLUMN IF NOT EXISTS source_type VARCHAR(50);

ALTER TABLE inventory_movements
    ADD COLUMN IF NOT EXISTS business_date DATE;

CREATE INDEX IF NOT EXISTS idx_inventory_movements_source_date
    ON inventory_movements(source_type, business_date);

ALTER TABLE inventory_items
    ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_inventory_items_customer
    ON inventory_items(customer_id);