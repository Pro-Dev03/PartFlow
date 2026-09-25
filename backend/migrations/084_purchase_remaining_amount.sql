-- Keep the purchase balance explicit and consistent in cloud and local schemas.
BEGIN;

ALTER TABLE public.purchases
    ADD COLUMN IF NOT EXISTS remaining_amount NUMERIC(10,2) NOT NULL DEFAULT 0;

UPDATE public.purchases
SET remaining_amount = GREATEST(COALESCE(total_amount, 0) - COALESCE(paid_amount, 0), 0)
WHERE remaining_amount IS DISTINCT FROM GREATEST(COALESCE(total_amount, 0) - COALESCE(paid_amount, 0), 0);

COMMENT ON COLUMN public.purchases.remaining_amount IS
    'Unpaid purchase balance, maintained as max(total_amount - paid_amount, 0)';

COMMIT;
