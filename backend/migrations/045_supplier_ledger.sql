BEGIN;

CREATE TABLE IF NOT EXISTS supplier_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('debit', 'credit')),
    amount DECIMAL(10,2) NOT NULL,
    balance DECIMAL(10,2) NOT NULL,
    description TEXT,
    reference_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 003 created supplier_ledger with transaction_type instead of type. Keep the
-- legacy schema compatible while later runtime reconciliation backfills both.
ALTER TABLE supplier_ledger ADD COLUMN IF NOT EXISTS type VARCHAR(20);

CREATE INDEX IF NOT EXISTS idx_supplier_ledger_supplier ON supplier_ledger(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_ledger_type ON supplier_ledger(type);
CREATE INDEX IF NOT EXISTS idx_supplier_ledger_created_at ON supplier_ledger(created_at);

COMMIT;
