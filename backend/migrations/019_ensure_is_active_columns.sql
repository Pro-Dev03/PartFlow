-- ============================================
-- Ensure is_active columns exist in all required tables
-- ============================================
-- This migration ensures that is_active columns exist in tables that reference them
-- in the code but might have been missed in previous migrations

-- Add is_active to products if it doesn't exist
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Add is_active to customers if it doesn't exist  
ALTER TABLE customers
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Add is_active to suppliers if it doesn't exist
ALTER TABLE suppliers
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Add is_active to brands if it doesn't exist
ALTER TABLE brands
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Add is_active to part_types if it doesn't exist
ALTER TABLE part_types
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Add is_active to locations if it doesn't exist
ALTER TABLE locations
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Create indexes for is_active columns for better query performance
CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active);
CREATE INDEX IF NOT EXISTS idx_customers_is_active ON customers(is_active);
CREATE INDEX IF NOT EXISTS idx_suppliers_is_active ON suppliers(is_active);
CREATE INDEX IF NOT EXISTS idx_brands_is_active ON brands(is_active);
CREATE INDEX IF NOT EXISTS idx_part_types_is_active ON part_types(is_active);
CREATE INDEX IF NOT EXISTS idx_locations_is_active ON locations(is_active);

SELECT 'is_active columns ensured in all required tables' as status;
