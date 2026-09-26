-- Add an explicit relationship from a customer ledger debit to its debt row.
-- The association is safe to backfill only when customer, amount and timestamp
-- identify exactly one debt and one ledger entry.
BEGIN;

ALTER TABLE customer_ledger
    ADD COLUMN IF NOT EXISTS debt_id UUID;
-- Older PostgreSQL schemas store debit/credit in transaction_type, while the
-- active backend and SQLite use type. Add/backfill the compatibility field
-- before matching legacy manual-debt entries below.
ALTER TABLE customer_ledger
    ADD COLUMN IF NOT EXISTS type VARCHAR(20);

UPDATE customer_ledger
SET type = CASE
    WHEN UPPER(COALESCE(transaction_type, '')) IN ('PAYMENT', 'RETURN', 'REFUND') THEN 'credit'
    WHEN UPPER(COALESCE(transaction_type, '')) = 'ADJUSTMENT' AND amount < 0 THEN 'credit'
    ELSE 'debit'
END
WHERE type IS NULL AND transaction_type IS NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'customer_ledger_debt_id_fkey'
          AND conrelid = 'customer_ledger'::regclass
    ) THEN
        ALTER TABLE customer_ledger
            ADD CONSTRAINT customer_ledger_debt_id_fkey
            FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE RESTRICT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_customer_ledger_debt_id
    ON customer_ledger(debt_id) WHERE debt_id IS NOT NULL;

UPDATE customer_ledger AS ledger
SET debt_id = debt.id
FROM debts AS debt
WHERE ledger.debt_id IS NULL
  AND ledger.customer_id = debt.customer_id
  AND debt.sale_id IS NULL
  AND LOWER(COALESCE(ledger.type, '')) = 'debit'
  AND ABS(ledger.amount - debt.amount) < 0.000001
  AND ledger.created_at = debt.created_at
  AND (
      SELECT COUNT(*) FROM debts AS candidate
      WHERE candidate.customer_id = ledger.customer_id
        AND candidate.sale_id IS NULL
        AND ABS(candidate.amount - ledger.amount) < 0.000001
        AND candidate.created_at = ledger.created_at
  ) = 1
  AND (
      SELECT COUNT(*) FROM customer_ledger AS candidate
      WHERE candidate.customer_id = ledger.customer_id
        AND LOWER(COALESCE(candidate.type, '')) = 'debit'
        AND ABS(candidate.amount - ledger.amount) < 0.000001
        AND candidate.created_at = ledger.created_at
  ) = 1;

COMMIT;
