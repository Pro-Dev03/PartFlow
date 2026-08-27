ALTER TABLE inspections
ADD COLUMN IF NOT EXISTS inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE SET NULL;

ALTER TABLE inspections
ALTER COLUMN product_id DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_inspections_inventory_item
ON inspections(inventory_item_id);
