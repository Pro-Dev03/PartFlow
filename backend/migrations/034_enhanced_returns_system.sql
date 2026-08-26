-- Enhanced Returns System for PartFlow
-- Based on comprehensive returns analysis
-- This migration implements a robust returns system that:
-- 1. Never deletes sales - creates independent return events
-- 2. Supports partial and full returns
-- 3. Integrates with debt system
-- 4. Tracks item condition after return
-- 5. Maintains complete audit trail
-- 6. Supports warranty claims
-- 7. Provides detailed return items tracking

-- First, let's backup existing returns data if any
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'returns') THEN
        CREATE TABLE returns_backup AS SELECT * FROM returns;
    END IF;
END $$;

-- Drop existing functions and triggers
DROP FUNCTION IF EXISTS generate_return_number() CASCADE;
DROP TRIGGER IF EXISTS generate_return_number_trigger ON returns;

-- Drop existing returns table (will be recreated with enhanced structure)
DROP TABLE IF EXISTS returns CASCADE;

-- Enhanced returns table
CREATE TABLE returns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    return_number VARCHAR(50) UNIQUE NOT NULL,
    reference_number VARCHAR(50) NOT NULL,
    
    -- Source information
    sale_id UUID REFERENCES sales(id) ON DELETE SET NULL,
    purchase_id UUID REFERENCES purchases(id) ON DELETE SET NULL,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    
    -- Return details
    return_date DATE NOT NULL DEFAULT CURRENT_DATE,
    return_type VARCHAR(20) NOT NULL CHECK (return_type IN ('FULL', 'PARTIAL', 'QUANTITY_PARTIAL')),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'PROCESSING', 'COMPLETED', 'REJECTED', 'CANCELLED')),
    
    -- Financial details
    total_refund_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    refund_method VARCHAR(20) CHECK (refund_method IN ('CASH', 'CREDIT', 'DEBT_ADJUSTMENT', 'EXCHANGE', 'BANK_TRANSFER', 'STORE_CREDIT')),
    refund_date DATE,
    refund_reference VARCHAR(100),
    
    -- Debt integration
    debt_id UUID REFERENCES debts(id) ON DELETE SET NULL,
    debt_adjustment NUMERIC(10,2) DEFAULT 0,
    customer_credit NUMERIC(10,2) DEFAULT 0,
    
    -- Return reason and condition
    reason VARCHAR(50) NOT NULL CHECK (reason IN ('DEFECTIVE', 'WRONG_ITEM', 'COMPATIBILITY_ISSUE', 'CUSTOMER_CHANGED_MIND', 'DAMAGED', 'WARRANTY', 'INCORRECT_SPECIFICATION', 'OTHER')),
    reason_detail TEXT,
    item_condition_after_return VARCHAR(20) CHECK (item_condition_after_return IN ('SELLABLE', 'NEEDS_INSPECTION', 'NEEDS_REPAIR', 'DAMAGED', 'USED', 'REFURBISHED', 'SUPPLIER_RETURN', 'WRITE_OFF', 'PARTS')),
    
    -- Warranty information
    is_warranty_claim BOOLEAN DEFAULT FALSE,
    warranty_id UUID,
    warranty_valid_until DATE,
    
    -- Approval workflow
    created_by UUID REFERENCES users(id),
    processed_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    
    -- Notes and audit
    notes TEXT,
    internal_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_returns_sale ON returns(sale_id);
CREATE INDEX idx_returns_purchase ON returns(purchase_id);
CREATE INDEX idx_returns_customer ON returns(customer_id);
CREATE INDEX idx_returns_date ON returns(return_date);
CREATE INDEX idx_returns_status ON returns(status);
CREATE INDEX idx_returns_type ON returns(return_type);
CREATE INDEX idx_returns_refund_method ON returns(refund_method);
CREATE INDEX idx_returns_debt ON returns(debt_id);
CREATE INDEX idx_returns_warranty ON returns(warranty_id);
CREATE INDEX idx_returns_number ON returns(return_number);

-- Create return_items table for detailed item tracking
CREATE TABLE return_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    return_id UUID NOT NULL REFERENCES returns(id) ON DELETE CASCADE,
    
    -- Item identification
    sale_item_id UUID REFERENCES sale_items(id) ON DELETE SET NULL,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    inventory_item_id UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    serial_number VARCHAR(100),
    barcode VARCHAR(100),
    
    -- Quantity and pricing
    quantity_returned INTEGER NOT NULL CHECK (quantity_returned > 0),
    original_quantity INTEGER,
    unit_price NUMERIC(10,2) NOT NULL,
    total_refund_amount NUMERIC(10,2) NOT NULL,
    
    -- Item condition
    original_condition VARCHAR(20),
    returned_condition VARCHAR(20) CHECK (returned_condition IN ('NEW', 'USED', 'DAMAGED', 'DEFECTIVE', 'OPEN_BOX', 'REFURBISHED')),
    condition_notes TEXT,
    
    -- Resolution
    resolution VARCHAR(20) CHECK (resolution IN ('RESTOCK', 'REPAIR', 'SUPPLIER_RETURN', 'WRITE_OFF', 'PARTS', 'REPLACEMENT')),
    inventory_status VARCHAR(20) DEFAULT 'RETURNED' CHECK (inventory_status IN ('RETURNED', 'INSPECTION', 'REPAIRING', 'RESTOCKED', 'SUPPLIER_RETURNED', 'WRITTEN_OFF', 'DISMANTLED')),
    
    -- Inspection details
    inspection_required BOOLEAN DEFAULT TRUE,
    inspection_date DATE,
    inspection_result VARCHAR(20) CHECK (inspection_result IN ('PASSED', 'FAILED', 'PENDING')),
    inspection_notes TEXT,
    
    -- Cost tracking
    original_cost NUMERIC(10,2),
    repair_cost NUMERIC(10,2) DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for return_items
CREATE INDEX idx_return_items_return ON return_items(return_id);
CREATE INDEX idx_return_items_sale_item ON return_items(sale_item_id);
CREATE INDEX idx_return_items_product ON return_items(product_id);
CREATE INDEX idx_return_items_inventory ON return_items(inventory_item_id);
CREATE INDEX idx_return_items_serial ON return_items(serial_number);
CREATE INDEX idx_return_items_resolution ON return_items(resolution);
CREATE INDEX idx_return_items_status ON return_items(inventory_status);

-- Create trigger to update updated_at
CREATE TRIGGER update_returns_updated_at BEFORE UPDATE ON returns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_return_items_updated_at BEFORE UPDATE ON return_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to generate return numbers
CREATE OR REPLACE FUNCTION generate_return_number()
RETURNS TRIGGER AS $$
DECLARE
    return_num VARCHAR;
    date_part VARCHAR;
    seq_num INTEGER;
BEGIN
    date_part := TO_CHAR(CURRENT_DATE, 'YYYYMM');
    
    -- Get next sequence number for this month
    SELECT COALESCE(MAX(CAST(SUBSTRING(return_number FROM 9) AS INTEGER)), 0) + 1
    INTO seq_num
    FROM returns
    WHERE return_number LIKE 'RET-' || date_part || '%';
    
    return_num := 'RET-' || date_part || LPAD(seq_num::TEXT, 5, '0');
    
    NEW.return_number := return_num;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-generate return number
CREATE TRIGGER generate_return_number_trigger
    BEFORE INSERT ON returns
    FOR EACH ROW
    WHEN (NEW.return_number IS NULL OR NEW.return_number = '')
    EXECUTE FUNCTION generate_return_number();

-- Function to handle debt adjustment on return completion
CREATE OR REPLACE FUNCTION handle_return_debt_adjustment()
RETURNS TRIGGER AS $$
DECLARE
    current_debt NUMERIC;
    new_debt NUMERIC;
BEGIN
    -- Only process when return is completed and has debt adjustment
    IF NEW.status = 'COMPLETED' AND OLD.status != 'COMPLETED' AND NEW.debt_id IS NOT NULL AND NEW.debt_adjustment != 0 THEN
        -- Get current debt amount
        SELECT remaining_amount INTO current_debt
        FROM debts
        WHERE id = NEW.debt_id;
        
        IF current_debt IS NOT NULL THEN
            -- Calculate new debt
            new_debt := current_debt - NEW.debt_adjustment;
            
            -- Ensure debt doesn't go negative
            IF new_debt < 0 THEN
                -- Create customer credit for excess
                UPDATE returns
                SET customer_credit = ABS(new_debt),
                    debt_adjustment = current_debt,
                    updated_at = NOW()
                WHERE id = NEW.id;
                
                new_debt := 0;
            END IF;
            
            -- Update debt
            UPDATE debts
            SET remaining_amount = new_debt,
                updated_at = NOW()
            WHERE id = NEW.debt_id;
            
            -- Update debt status if fully paid
            IF new_debt = 0 THEN
                UPDATE debts
                SET status = 'paid',
                    updated_at = NOW()
                WHERE id = NEW.debt_id;
            END IF;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to handle debt adjustment
CREATE TRIGGER handle_return_debt_trigger
    AFTER UPDATE OF status, debt_adjustment, debt_id ON returns
    FOR EACH ROW
    WHEN (OLD.status != 'COMPLETED' AND NEW.status = 'COMPLETED')
    EXECUTE FUNCTION handle_return_debt_adjustment();

-- Function to validate return quantities
CREATE OR REPLACE FUNCTION validate_return_quantity()
RETURNS TRIGGER AS $$
DECLARE
    original_qty INTEGER;
    total_returned INTEGER;
BEGIN
    -- Check if this is a sale item return
    IF NEW.sale_item_id IS NOT NULL THEN
        -- Get original quantity from sale
        SELECT quantity INTO original_qty
        FROM sale_items
        WHERE id = NEW.sale_item_id;
        
        -- Calculate total returned for this sale item
        SELECT COALESCE(SUM(quantity_returned), 0) INTO total_returned
        FROM return_items
        WHERE sale_item_id = NEW.sale_item_id
        AND id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::UUID);
        
        -- Add current return quantity
        total_returned := total_returned + NEW.quantity_returned;
        
        -- Validate
        IF original_qty IS NOT NULL AND total_returned > original_qty THEN
            RAISE EXCEPTION 'Cannot return more items than originally sold. Original: %, Already returned: %, Current return: %',
                original_qty, total_returned - NEW.quantity_returned, NEW.quantity_returned;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to validate return quantities
CREATE TRIGGER validate_return_quantity_trigger
    BEFORE INSERT OR UPDATE OF quantity_returned ON return_items
    FOR EACH ROW
    EXECUTE FUNCTION validate_return_quantity();

-- Create view for returns summary
CREATE OR REPLACE VIEW returns_summary AS
SELECT 
    r.id,
    r.return_number,
    r.reference_number,
    r.sale_id,
    s.invoice_number as sale_invoice,
    r.customer_id,
    c.name as customer_name,
    r.return_date,
    r.return_type,
    r.status,
    r.total_refund_amount,
    r.refund_method,
    r.reason,
    r.item_condition_after_return,
    COUNT(ri.id) as total_items,
    SUM(ri.quantity_returned) as total_quantity_returned,
    r.created_at,
    r.updated_at
FROM returns r
LEFT JOIN sales s ON r.sale_id = s.id
LEFT JOIN customers c ON r.customer_id = c.id
LEFT JOIN return_items ri ON r.id = ri.return_id
GROUP BY r.id, r.return_number, r.reference_number, r.sale_id, s.invoice_number, 
         r.customer_id, c.name, r.return_date, r.return_type, r.status, 
         r.total_refund_amount, r.refund_method, r.reason, r.item_condition_after_return,
         r.created_at, r.updated_at;

-- Create view for monthly returns analysis
CREATE OR REPLACE VIEW monthly_returns_analysis AS
SELECT 
    DATE_TRUNC('month', r.return_date) as month,
    COUNT(DISTINCT r.id) as total_returns,
    COUNT(DISTINCT r.customer_id) as unique_customers,
    SUM(r.total_refund_amount) as total_refund_amount,
    AVG(r.total_refund_amount) as avg_refund_amount,
    COUNT(CASE WHEN r.return_type = 'FULL' THEN 1 END) as full_returns,
    COUNT(CASE WHEN r.return_type = 'PARTIAL' THEN 1 END) as partial_returns,
    COUNT(CASE WHEN r.return_type = 'QUANTITY_PARTIAL' THEN 1 END) as quantity_partial_returns,
    COUNT(CASE WHEN r.reason = 'DEFECTIVE' THEN 1 END) as defective_returns,
    COUNT(CASE WHEN r.reason = 'WARRANTY' THEN 1 END) as warranty_returns,
    COUNT(CASE WHEN r.is_warranty_claim = TRUE THEN 1 END) as warranty_claims,
    SUM(CASE WHEN r.item_condition_after_return = 'SELLABLE' THEN 1 ELSE 0 END) as sellable_items,
    SUM(CASE WHEN r.item_condition_after_return = 'NEEDS_REPAIR' THEN 1 ELSE 0 END) as repair_needed,
    SUM(CASE WHEN r.item_condition_after_return = 'WRITE_OFF' THEN 1 ELSE 0 END) as written_off
FROM returns r
WHERE r.status = 'COMPLETED'
GROUP BY DATE_TRUNC('month', r.return_date)
ORDER BY month DESC;

-- Create view for sales vs returns analysis
CREATE OR REPLACE VIEW sales_returns_analysis AS
SELECT 
    DATE_TRUNC('month', s.sale_date) as month,
    COUNT(DISTINCT s.id) as total_sales,
    COALESCE(SUM(s.total_amount), 0) as gross_sales,
    COALESCE(SUM(s.cost_amount), 0) as total_cost,
    COALESCE(SUM(s.gross_profit), 0) as gross_profit,
    COALESCE(SUM(r.total_refund_amount), 0) as returns_amount,
    COALESCE(COUNT(r.id), 0) as return_count,
    COALESCE(SUM(s.total_amount), 0) - COALESCE(SUM(r.total_refund_amount), 0) as net_sales
FROM sales s
LEFT JOIN returns r ON r.sale_id = s.id 
    AND r.status = 'COMPLETED' 
    AND DATE_TRUNC('month', r.return_date) = DATE_TRUNC('month', s.sale_date)
WHERE s.status = 'completed'
GROUP BY DATE_TRUNC('month', s.sale_date)
ORDER BY month DESC;

-- Create function to handle inventory status changes based on return item resolution
CREATE OR REPLACE FUNCTION update_return_item_inventory_status()
RETURNS TRIGGER AS $$
BEGIN
    -- Only process when resolution is set
    IF NEW.resolution IS NOT NULL AND (OLD.resolution IS NULL OR OLD.resolution != NEW.resolution) THEN
        CASE NEW.resolution
            WHEN 'RESTOCK' THEN
                NEW.inventory_status := 'RESTOCKED';
            WHEN 'REPAIR' THEN
                NEW.inventory_status := 'REPAIRING';
            WHEN 'SUPPLIER_RETURN' THEN
                NEW.inventory_status := 'SUPPLIER_RETURNED';
            WHEN 'WRITE_OFF' THEN
                NEW.inventory_status := 'WRITTEN_OFF';
            WHEN 'PARTS' THEN
                NEW.inventory_status := 'DISMANTLED';
            ELSE
                NEW.inventory_status := 'RETURNED';
        END CASE;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update inventory status based on resolution
CREATE TRIGGER update_return_item_inventory_status_trigger
    BEFORE UPDATE OF resolution ON return_items
    FOR EACH ROW
    EXECUTE FUNCTION update_return_item_inventory_status();

SELECT 'Enhanced returns system created successfully' as result;