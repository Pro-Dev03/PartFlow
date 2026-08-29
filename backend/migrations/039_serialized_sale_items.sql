BEGIN;

ALTER TABLE sale_items
    ADD COLUMN IF NOT EXISTS inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_sale_items_inventory_item
    ON sale_items(inventory_item_id);

COMMIT;
