-- A debt with no collection is pending, not partially paid.
UPDATE debts
SET status = CASE
    WHEN remaining_amount <= 0 THEN 'paid'
    WHEN COALESCE(paid_amount, 0) > 0 THEN 'partial'
    ELSE 'pending'
END,
updated_at = CURRENT_TIMESTAMP
WHERE sale_id IS NOT NULL;
