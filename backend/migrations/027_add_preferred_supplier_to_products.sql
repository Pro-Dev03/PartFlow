-- Add preferred supplier to products table
-- This allows tracking which supplier is the preferred source for each product

ALTER TABLE products ADD COLUMN IF NOT EXISTS preferred_supplier_id UUID REFERENCES suppliers(id);

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_products_preferred_supplier ON products(preferred_supplier_id);

-- Add comment
COMMENT ON COLUMN products.preferred_supplier_id IS 'المورد المفضل للمنتج';
