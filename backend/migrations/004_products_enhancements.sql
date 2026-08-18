-- PartFlow Products Enhancements
-- Add product types and individual tracking support

-- ============================================
-- Add product_type and individual tracking to products
-- ============================================
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS product_type VARCHAR(20) DEFAULT 'quantity' CHECK (product_type IN ('quantity', 'individual')),
ADD COLUMN IF NOT EXISTS track_serial BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS track_individual BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS default_cost DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS default_price DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS warranty_policy TEXT;

-- ============================================
-- Add brands table if not exists
-- ============================================
CREATE TABLE IF NOT EXISTS brands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    logo_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, name)
);

CREATE INDEX IF NOT EXISTS idx_brands_organization ON brands(organization_id);

-- ============================================
-- Add brand_id to products
-- ============================================
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS brand_id UUID REFERENCES brands(id);

CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand_id);

-- ============================================
-- Update products table to remove brand column (use brand_id instead)
-- ============================================
-- This is a migration step - data migration would be needed in production
-- ALTER TABLE products DROP COLUMN IF EXISTS brand;

-- ============================================
-- Add updated_at trigger for brands
-- ============================================
CREATE TRIGGER update_brands_updated_at BEFORE UPDATE ON brands
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Add function to generate item codes
-- ============================================
CREATE OR REPLACE FUNCTION generate_item_code()
RETURNS VARCHAR(50) AS $$
BEGIN
    RETURN 'ITEM-' || LPAD(nextval('item_code_seq')::TEXT, 6, '0');
END;
$$ LANGUAGE plpgsql;

-- Create sequence for item codes if not exists
CREATE SEQUENCE IF NOT EXISTS item_code_seq START 1;

-- ============================================
-- Add function to generate internal barcodes
-- ============================================
CREATE OR REPLACE FUNCTION generate_internal_barcode(product_id UUID)
RETURNS VARCHAR(50) AS $$
BEGIN
    RETURN 'PF-' || SUBSTRING(product_id::TEXT, 1, 8);
END;
$$ LANGUAGE plpgsql;
