BEGIN;

ALTER TABLE payments ADD COLUMN IF NOT EXISTS type VARCHAR(50);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS reference_id UUID;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS method VARCHAR(50);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS reference TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS status VARCHAR(30) DEFAULT 'completed';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL;

UPDATE payments
SET method = COALESCE(method, payment_method),
    reference = COALESCE(reference, reference_number),
    type = COALESCE(type, CASE WHEN customer_id IS NOT NULL THEN 'customer' WHEN supplier_id IS NOT NULL THEN 'supplier' ELSE 'other' END),
    reference_id = COALESCE(reference_id, customer_id, supplier_id);

COMMIT;
