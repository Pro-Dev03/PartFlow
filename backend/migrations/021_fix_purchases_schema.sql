-- ============================================
-- Fix purchases table schema to match code expectations
-- ============================================
-- The Go code expects different column names than the current schema

-- The current schema already has most columns needed
-- Add any missing columns if the code expects them
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS subtotal DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS tax_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS notes TEXT,
ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);

SELECT 'Purchases schema verified/fixed' as status;
