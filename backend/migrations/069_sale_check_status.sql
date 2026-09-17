ALTER TABLE sale_payment_allocations
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending';

CREATE INDEX IF NOT EXISTS idx_sale_payment_allocations_status
    ON sale_payment_allocations(status);