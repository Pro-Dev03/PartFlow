CREATE TABLE IF NOT EXISTS sale_payment_allocations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sale_id UUID NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    amount DECIMAL(12,2) NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(30) NOT NULL,
    check_number VARCHAR(100),
    bank_name VARCHAR(255),
    check_date DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sale_payment_allocations_sale ON sale_payment_allocations(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_payment_allocations_check ON sale_payment_allocations(check_number) WHERE check_number IS NOT NULL;
