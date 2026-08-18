-- PartFlow Sales Enhancements
-- Add profit tracking and automation support to sales

-- ============================================
-- Enhance Sales Table
-- ============================================
ALTER TABLE sales 
ADD COLUMN IF NOT EXISTS cost_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS gross_profit DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS net_profit DECIMAL(10,2) DEFAULT 0;

-- ============================================
-- Enhance Sale Items Table
-- ============================================
ALTER TABLE sale_items
ADD COLUMN IF NOT EXISTS unit_cost DECIMAL(10,2) DEFAULT 0;

-- ============================================
-- Create Indexes for Performance
-- ============================================
CREATE INDEX IF NOT EXISTS idx_sales_organization_date ON sales(organization_id, sale_date);
CREATE INDEX IF NOT EXISTS idx_sales_customer ON sales(customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sale_items_product ON sale_items(product_id);

-- ============================================
-- Add Comments
-- ============================================
COMMENT ON COLUMN sales.cost_amount IS 'Total cost of goods sold at time of sale';
COMMENT ON COLUMN sales.gross_profit IS 'Gross profit (revenue - cost)';
COMMENT ON COLUMN sales.net_profit IS 'Net profit (gross profit - taxes - discounts)';
COMMENT ON COLUMN sale_items.unit_cost IS 'Unit cost at time of sale for profit calculation';
