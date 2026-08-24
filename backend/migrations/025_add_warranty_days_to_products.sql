-- Add warranty_days column to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS warranty_days INTEGER DEFAULT 0;