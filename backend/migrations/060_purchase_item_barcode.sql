BEGIN;

ALTER TABLE purchase_items
    ADD COLUMN IF NOT EXISTS barcode VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_purchase_items_barcode
    ON purchase_items(barcode);

COMMIT;