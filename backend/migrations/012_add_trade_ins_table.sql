-- Add trade_ins table to track used items purchased from customers
-- This allows tracking which customer we bought the used item from

CREATE TABLE IF NOT EXISTS trade_ins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    purchase_price DECIMAL(10,2) NOT NULL, -- in shekels
    purchase_date TIMESTAMP NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add index for performance
CREATE INDEX idx_trade_ins_customer_id ON trade_ins(customer_id);
CREATE INDEX idx_trade_ins_inventory_item_id ON trade_ins(inventory_item_id);

-- Add comment
COMMENT ON TABLE trade_ins IS 'Tracks used items purchased from customers (trade-ins)';
COMMENT ON COLUMN trade_ins.purchase_price IS 'Price paid to customer in shekels';
