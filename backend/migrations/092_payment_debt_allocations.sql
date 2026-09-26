BEGIN;

-- Record which debt rows received each customer/supplier payment. The batch
-- row also marks that allocation tracking was performed when a payment had
-- no debt rows, so deletion never guesses allocations for legacy payments.
CREATE TABLE IF NOT EXISTS payment_allocation_batches (
    payment_id UUID PRIMARY KEY REFERENCES payments(id) ON DELETE CASCADE,
    owner_type VARCHAR(20) NOT NULL CHECK (owner_type IN ('customer', 'supplier')),
    owner_id UUID NOT NULL,
    sale_id UUID,
    tracked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_debt_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payment_allocation_batches(payment_id) ON DELETE CASCADE,
    debt_id UUID NOT NULL,
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (payment_id, debt_id)
);

CREATE INDEX IF NOT EXISTS idx_payment_debt_allocations_payment
    ON payment_debt_allocations(payment_id);

COMMIT;
