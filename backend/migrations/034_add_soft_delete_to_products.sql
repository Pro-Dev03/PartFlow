-- Add soft delete capability to products table
-- This migration only affects products, not used parts (which remain isolated)

-- Add deleted_at column to products table
ALTER TABLE products 
ADD COLUMN deleted_at TIMESTAMP NULL;

-- Create index on deleted_at for better query performance
CREATE INDEX idx_products_deleted_at ON products(deleted_at);

-- Add comment explaining the soft delete logic
COMMENT ON COLUMN products.deleted_at IS 'Soft delete timestamp. NULL means the product is active, a timestamp means it has been logically deleted.';
