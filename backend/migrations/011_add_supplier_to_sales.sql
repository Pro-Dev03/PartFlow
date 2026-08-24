-- Add supplier_id to sale_items for tracking supplier origin
-- This allows tracking which supplier a sold item came from

-- Add supplier_id column to sale_items
ALTER TABLE sale_items
ADD COLUMN supplier_id UUID REFERENCES suppliers(id);

-- Add index for performance
CREATE INDEX idx_sale_items_supplier_id ON sale_items(supplier_id);

-- Add comment
COMMENT ON COLUMN sale_items.supplier_id IS 'ID of the supplier who provided this item (links back to inventory_items.supplier_id)';
