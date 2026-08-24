-- ============================================
-- Fix returns table schema to match code expectations
-- ============================================
-- The Go code expects different column names than the current schema

-- The current returns table has a different structure than what the code expects
-- The code expects: return_number, sale_id, customer_id, return_date, reason, condition, status, refund_amount, refund_method, refund_date, notes, processed_by
-- Current table has: reference_number, type, sale_id, purchase_id, product_id, quantity, reason, status

-- Add missing columns
ALTER TABLE returns
ADD COLUMN IF NOT EXISTS return_number VARCHAR(50),
ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id),
ADD COLUMN IF NOT EXISTS return_date DATE DEFAULT CURRENT_DATE,
ADD COLUMN IF NOT EXISTS condition VARCHAR(50),
ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS refund_method VARCHAR(50),
ADD COLUMN IF NOT EXISTS refund_date DATE,
ADD COLUMN IF NOT EXISTS notes TEXT,
ADD COLUMN IF NOT EXISTS processed_by UUID REFERENCES users(id);

-- Migrate existing data
UPDATE returns 
SET return_number = COALESCE(return_number, reference_number),
    return_date = COALESCE(return_date, created_at::date)
WHERE return_number IS NULL OR return_date IS NULL;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_returns_customer ON returns(customer_id);
CREATE INDEX IF NOT EXISTS idx_returns_date ON returns(return_date);
CREATE INDEX IF NOT EXISTS idx_returns_status ON returns(status);

SELECT 'Returns schema fixed' as status;
