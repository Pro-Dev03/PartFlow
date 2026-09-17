ALTER TABLE payment_transactions
    ADD COLUMN IF NOT EXISTS order_id UUID;

CREATE INDEX IF NOT EXISTS idx_payment_transactions_order ON payment_transactions(order_id);
