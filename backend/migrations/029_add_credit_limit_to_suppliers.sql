-- Add credit_limit and current_balance columns to suppliers table
ALTER TABLE suppliers 
ADD COLUMN IF NOT EXISTS credit_limit DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS current_balance DECIMAL(10,2) DEFAULT 0;
