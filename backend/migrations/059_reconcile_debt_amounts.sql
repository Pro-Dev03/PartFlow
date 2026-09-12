-- Separate the deferred debt from the amount paid at the time of sale.
-- Older credit-sale rows stored the full sale total in debts.amount and the
-- sale advance in debts.paid_amount, which made the debt report overstate
-- collected debt.
UPDATE debts d
SET amount = GREATEST(s.total_amount - COALESCE(s.paid_amount, 0), 0),
    paid_amount = GREATEST(
        GREATEST(s.total_amount - COALESCE(s.paid_amount, 0), 0)
        - GREATEST(d.remaining_amount, 0),
        0
    ),
    remaining_amount = GREATEST(d.remaining_amount, 0),
    status = CASE
        WHEN GREATEST(d.remaining_amount, 0) = 0 THEN 'paid'
        ELSE 'partial'
    END,
    updated_at = CURRENT_TIMESTAMP
FROM sales s
WHERE d.sale_id = s.id
  AND LOWER(COALESCE(s.payment_method, '')) IN ('debt', 'credit', 'on_account');

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
