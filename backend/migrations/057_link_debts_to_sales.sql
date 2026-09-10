-- Link customer debts to the originating sale and preserve any advance payment.
ALTER TABLE debts ADD COLUMN IF NOT EXISTS sale_id UUID REFERENCES sales(id) ON DELETE SET NULL;
ALTER TABLE debts ADD COLUMN IF NOT EXISTS paid_amount NUMERIC(15,2) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_debts_sale_id ON debts(sale_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_debts_sale_id
    ON debts(sale_id)
    WHERE sale_id IS NOT NULL;

-- Backfill credit sales created before sale creation started inserting a debt row.
-- Only sales with an explicit credit payment method and a positive outstanding
-- amount are included, and the unique sale index makes this idempotent.
INSERT INTO debts (
        id,
        customer_id,
        sale_id,
        amount,
        paid_amount,
        remaining_amount,
        due_date,
        status,
        notes,
        created_at,
        updated_at
)
SELECT
        gen_random_uuid(),
        s.customer_id,
        s.id,
        s.total_amount,
        LEAST(GREATEST(COALESCE(s.paid_amount, 0), 0), s.total_amount),
        GREATEST(s.total_amount - COALESCE(s.paid_amount, 0), 0),
        COALESCE(s.sale_date::date + 30, CURRENT_DATE + 30),
        CASE WHEN COALESCE(s.paid_amount, 0) > 0 THEN 'partial' ELSE 'pending' END,
        'Backfilled from credit sale ' || COALESCE(s.invoice_number, s.id::text),
        s.created_at,
        s.updated_at
FROM sales s
WHERE s.customer_id IS NOT NULL
    AND LOWER(COALESCE(s.payment_method, '')) IN ('debt', 'credit', 'on_account')
    AND s.total_amount > COALESCE(s.paid_amount, 0)
    AND NOT EXISTS (SELECT 1 FROM debts d WHERE d.sale_id = s.id);

-- Reconcile affected customer balances with the debt records used by cards,
-- the Debts page, and reports.
UPDATE customers c
SET current_balance = COALESCE((
                SELECT SUM(d.remaining_amount)
                FROM debts d
                WHERE d.customer_id = c.id
                    AND d.remaining_amount > 0
                    AND LOWER(COALESCE(d.status, 'pending')) NOT IN ('paid', 'completed', 'settled', 'cancelled', 'canceled', 'reversed')
        ), 0),
        updated_at = CURRENT_TIMESTAMP
WHERE EXISTS (
        SELECT 1
        FROM debts d
        WHERE d.customer_id = c.id
            AND d.sale_id IS NOT NULL
);
