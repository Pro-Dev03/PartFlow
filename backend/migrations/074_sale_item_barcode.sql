ALTER TABLE sale_items ADD COLUMN IF NOT EXISTS barcode VARCHAR(100);
CREATE INDEX IF NOT EXISTS idx_sale_items_barcode ON sale_items(barcode);
