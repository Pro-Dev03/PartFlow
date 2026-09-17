CREATE TABLE IF NOT EXISTS return_payment_refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    return_id UUID NOT NULL UNIQUE REFERENCES returns(id) ON DELETE RESTRICT,
    payment_transaction_id UUID NOT NULL REFERENCES payment_transactions(id) ON DELETE RESTRICT,
    payment_refund_id UUID REFERENCES payment_refunds(id) ON DELETE SET NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_return_payment_refunds_transaction ON return_payment_refunds(payment_transaction_id);