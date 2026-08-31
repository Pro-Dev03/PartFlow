-- Normalize supplier ledger column names for databases created by the
-- earlier migrations. The application supports both names while existing
-- rows are copied into the canonical type column.
BEGIN;

ALTER TABLE supplier_ledger ADD COLUMN IF NOT EXISTS type VARCHAR(20);
ALTER TABLE supplier_ledger ADD COLUMN IF NOT EXISTS transaction_type VARCHAR(20);

UPDATE supplier_ledger
SET type = CASE
    WHEN transaction_type IN ('PAYMENT', 'RETURN') THEN 'credit'
    ELSE 'debit'
END
WHERE type IS NULL AND transaction_type IS NOT NULL;

UPDATE supplier_ledger
SET transaction_type = CASE
    WHEN type = 'credit' THEN 'PAYMENT'
    ELSE 'PURCHASE'
END
WHERE transaction_type IS NULL AND type IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_supplier_ledger_type_normalized
    ON supplier_ledger(type);

COMMIT;
