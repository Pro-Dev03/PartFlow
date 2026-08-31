-- Make supplier ledger writes portable across databases created by the
-- historical 003 (transaction_type) and 030/045 (type) migrations.  The
-- application writes both canonical fields after this migration.
BEGIN;

ALTER TABLE supplier_ledger ADD COLUMN IF NOT EXISTS type VARCHAR(20);
ALTER TABLE supplier_ledger ADD COLUMN IF NOT EXISTS transaction_type VARCHAR(20);
ALTER TABLE supplier_ledger ADD COLUMN IF NOT EXISTS reference_type VARCHAR(40);
ALTER TABLE supplier_ledger ADD COLUMN IF NOT EXISTS created_by UUID;

-- Legacy installations made one or both columns mandatory.  A ledger entry
-- can be created by an automated/local operation, so the actor is optional;
-- transaction_type/type are still populated by every application write.
ALTER TABLE supplier_ledger ALTER COLUMN type DROP NOT NULL;
ALTER TABLE supplier_ledger ALTER COLUMN transaction_type DROP NOT NULL;
ALTER TABLE supplier_ledger ALTER COLUMN created_by DROP NOT NULL;

UPDATE supplier_ledger
SET type = CASE
    WHEN transaction_type IN ('PAYMENT', 'RETURN', 'SUPPLIER_RETURN') THEN 'credit'
    ELSE 'debit'
END
WHERE type IS NULL AND transaction_type IS NOT NULL;

UPDATE supplier_ledger
SET transaction_type = CASE
    WHEN type = 'credit' THEN 'PAYMENT'
    ELSE 'PURCHASE'
END
WHERE transaction_type IS NULL AND type IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_supplier_ledger_runtime_type
    ON supplier_ledger(type);
CREATE INDEX IF NOT EXISTS idx_supplier_ledger_runtime_transaction_type
    ON supplier_ledger(transaction_type);

COMMIT;
