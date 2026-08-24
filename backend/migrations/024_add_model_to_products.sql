-- Add model column to products table
-- This column is used in the Go model but was missing from the database schema

ALTER TABLE products 
ADD COLUMN IF NOT EXISTS model VARCHAR(100);

-- Add index for model column for better search performance
CREATE INDEX IF NOT EXISTS idx_products_model ON products(model);
