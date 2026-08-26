-- ============================================
-- Enhanced Returns System
-- بناءً على تقرير نظام المرتجعات الشامل
-- ============================================

-- ============================================
-- 1. Create return_items table (جدول منتجات المرتجع)
-- ============================================
CREATE TABLE IF NOT EXISTS return_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Basic linkage
    return_id UUID NOT NULL REFERENCES returns(id) ON DELETE CASCADE,
    sale_item_id UUID NOT NULL REFERENCES sale_items(id),
    product_id UUID NOT NULL REFERENCES products(id),
    inventory_item_id UUID REFERENCES inventory_items(id), -- للقطع الفردية المستعملة
    
    -- Return details
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    
    -- Item condition after return
    condition VARCHAR(50) NOT NULL CHECK (condition IN ('new', 'used', 'damaged', 'refurbished', 'needs_repair', 'parts_only')),
    condition_reason TEXT,
    
    -- Inspection required
    inspection_required BOOLEAN DEFAULT false,
    inspection_status VARCHAR(50) DEFAULT 'pending' CHECK (inspection_status IN ('pending', 'passed', 'failed', 'needs_repair')),
    inspection_date DATE,
    inspection_notes TEXT,
    
    -- Inventory decision
    inventory_decision VARCHAR(50) CHECK (inventory_decision IN ('restock', 'repair', 'supplier_return', 'write_off', 'parts_only', 'awaiting_decision')),
    inventory_date DATE,
    inventory_notes TEXT,
    
    -- Warranty info
    is_warranty BOOLEAN DEFAULT false,
    warranty_id UUID REFERENCES warranties(id),
    
    -- Metadata
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_return_items_return ON return_items(return_id);
CREATE INDEX idx_return_items_sale_item ON return_items(sale_item_id);
CREATE INDEX idx_return_items_product ON return_items(product_id);
CREATE INDEX idx_return_items_inventory ON return_items(inventory_item_id);
CREATE INDEX idx_return_items_inspection ON return_items(inspection_status);
CREATE INDEX idx_return_items_decision ON return_items(inventory_decision);

-- ============================================
-- 2. Enhance returns table with additional fields
-- ============================================
ALTER TABLE returns
ADD COLUMN IF NOT EXISTS return_type VARCHAR(50) DEFAULT 'full' CHECK (return_type IN ('full', 'partial', 'partial_quantity')),
ADD COLUMN IF NOT EXISTS original_total_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS net_amount DECIMAL(10,2) DEFAULT 0, -- بعد خصوص المرتجعات السابقة
ADD COLUMN IF NOT EXISTS customer_credit DECIMAL(10,2) DEFAULT 0, -- رصيد للعميل
ADD COLUMN IF NOT EXISTS debt_adjustment DECIMAL(10,2) DEFAULT 0, -- تعديل على الدين
ADD COLUMN IF NOT EXISTS debt_id UUID REFERENCES debts(id), -- الديون المرتبطة
ADD COLUMN IF NOT EXISTS warranty_claim BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS supplier_return_id UUID, -- في حالة إرجاع للمورد
ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS rejected_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS rejected_reason TEXT,
ADD COLUMN IF NOT EXISTS reversal_of UUID REFERENCES returns(id), -- لفحص العكس
ADD COLUMN IF NOT EXISTS audit_log JSONB DEFAULT '{}';

CREATE INDEX idx_returns_return_type ON returns(return_type);
CREATE INDEX idx_returns_debt ON returns(debt_id);
CREATE INDEX idx_returns_approved_by ON returns(approved_by);
CREATE INDEX idx_returns_reversal ON returns(reversal_of);

-- ============================================
-- 3. Create return_refunds table (جدول رد الأموال)
-- ============================================
CREATE TABLE IF NOT EXISTS return_refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    return_id UUID NOT NULL REFERENCES returns(id) ON DELETE CASCADE,
    refund_type VARCHAR(50) NOT NULL CHECK (refund_type IN ('cash', 'card', 'bank_transfer', 'store_credit', 'debt_reduction', 'partial')),
    
    amount DECIMAL(10,2) NOT NULL,
    refund_date DATE NOT NULL,
    payment_method VARCHAR(50),
    transaction_reference VARCHAR(100),
    
    -- For store credit
    credit_balance_id UUID,
    credit_expiry_date DATE,
    
    -- For debt reduction
    debt_id UUID REFERENCES debts(id),
    debt_reduction_amount DECIMAL(10,2),
    
    -- Processing info
    processed_by UUID NOT NULL REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    notes TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_return_refunds_return ON return_refunds(return_id);
CREATE INDEX idx_return_refunds_debt ON return_refunds(debt_id);
CREATE INDEX idx_return_refunds_date ON return_refunds(refund_date);

-- ============================================
-- 4. Create return_inspection table (جدول فحص المرتجعات)
-- ============================================
CREATE TABLE IF NOT EXISTS return_inspection (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    return_item_id UUID NOT NULL REFERENCES return_items(id) ON DELETE CASCADE,
    
    inspection_date DATE NOT NULL,
    inspector_id UUID NOT NULL REFERENCES users(id),
    
    -- Inspection results
    inspection_result VARCHAR(50) NOT NULL CHECK (inspection_result IN ('passed', 'failed', 'needs_repair', 'condemned')),
    inspection_notes TEXT,
    
    -- Condition assessment
    physical_condition VARCHAR(50),
    functionality_status VARCHAR(50),
    defects_found TEXT[],
    accessories_present TEXT[],
    
    -- Repair recommendation
    repair_required BOOLEAN DEFAULT false,
    repair_type VARCHAR(50),
    estimated_repair_cost DECIMAL(10,2),
    repair_recommendation TEXT,
    
    -- Final decision
    recommended_action VARCHAR(50) CHECK (recommended_action IN ('restock', 'repair', 'supplier_return', 'write_off', 'parts_only')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_return_inspection_return_item ON return_inspection(return_item_id);
CREATE INDEX idx_return_inspection_inspector ON return_inspection(inspector_id);
CREATE INDEX idx_return_inspection_result ON return_inspection(inspection_result);

-- ============================================
-- 5. Create return_audit_log table (سجل تتبع المرتجعات)
-- ============================================
CREATE TABLE IF NOT EXISTS return_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    return_id UUID NOT NULL REFERENCES returns(id) ON DELETE CASCADE,
    return_item_id UUID REFERENCES return_items(id),
    
    action VARCHAR(50) NOT NULL, -- created, approved, rejected, refunded, inspected, inventory_updated, cancelled
    action_by UUID NOT NULL REFERENCES users(id),
    action_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Action details
    previous_status VARCHAR(50),
    new_status VARCHAR(50),
    amount_changed DECIMAL(10,2),
    inventory_quantity INT,
    
    notes TEXT,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_return_audit_log_return ON return_audit_log(return_id);
CREATE INDEX idx_return_audit_log_action_by ON return_audit_log(action_by);
CREATE INDEX idx_return_audit_log_action_at ON return_audit_log(action_at);

-- ============================================
-- 6. Triggers and Functions
-- ============================================

-- Trigger for return_items updated_at
CREATE TRIGGER update_return_items_updated_at BEFORE UPDATE ON return_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for return_refunds updated_at
CREATE TRIGGER update_return_refunds_updated_at BEFORE UPDATE ON return_refunds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for return_inspection updated_at
CREATE TRIGGER update_return_inspection_updated_at BEFORE UPDATE ON return_inspection
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to calculate return total amount from items
CREATE OR REPLACE FUNCTION calculate_return_total()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE returns 
    SET refund_amount = (
        SELECT COALESCE(SUM(total_price), 0) 
        FROM return_items 
        WHERE return_id = NEW.return_id
    ),
    updated_at = NOW()
    WHERE id = NEW.return_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to update return total when items change
CREATE TRIGGER update_return_total_items AFTER INSERT OR UPDATE OR DELETE ON return_items
    FOR EACH ROW EXECUTE FUNCTION calculate_return_total();

-- Function to create audit log entry
CREATE OR REPLACE FUNCTION create_return_audit_log()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO return_audit_log (return_id, action, action_by, action_at, previous_status, new_status, notes, metadata)
    VALUES (
        NEW.id,
        'created',
        NEW.processed_by,
        NOW(),
        NULL,
        NEW.status,
        'Return created',
        jsonb_build_object(
            'refund_amount', NEW.refund_amount,
            'refund_method', NEW.refund_method,
            'return_type', NEW.return_type
        )
    );
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to create audit log when return is created
CREATE TRIGGER create_return_audit_log_trigger AFTER INSERT ON returns
    FOR EACH ROW WHEN (NEW.processed_by IS NOT NULL)
    EXECUTE FUNCTION create_return_audit_log();

-- ============================================
-- 7. Views for Reporting
-- ============================================

-- View for returns summary
CREATE OR REPLACE VIEW returns_summary AS
SELECT 
    r.id,
    r.return_number,
    r.sale_id,
    s.invoice_number,
    r.customer_id,
    c.name as customer_name,
    r.return_date,
    r.return_type,
    r.status,
    r.refund_amount,
    r.refund_method,
    r.original_total_amount,
    r.net_amount,
    r.customer_credit,
    r.debt_adjustment,
    COUNT(ri.id) as total_items,
    SUM(ri.quantity) as total_quantity,
    COUNT(CASE WHEN ri.inspection_required = true THEN 1 END) as items_needing_inspection,
    COUNT(CASE WHEN ri.inspection_status = 'passed' THEN 1 END) as items_passed,
    COUNT(CASE WHEN ri.inspection_status = 'failed' THEN 1 END) as items_failed,
    COUNT(CASE WHEN ri.inventory_decision = 'restock' THEN 1 END) as items_restocked,
    COUNT(CASE WHEN ri.inventory_decision = 'supplier_return' THEN 1 END) as items_returned_to_supplier
FROM returns r
LEFT JOIN sales s ON r.sale_id = s.id
LEFT JOIN customers c ON r.customer_id = c.id
LEFT JOIN return_items ri ON r.id = ri.return_id
GROUP BY r.id, s.invoice_number, c.name;

-- View for returns impact on sales (gross vs net)
CREATE OR REPLACE VIEW returns_sales_impact AS
SELECT 
    s.id as sale_id,
    s.invoice_number,
    s.sale_date,
    s.total_amount as gross_sale_amount,
    COALESCE(SUM(r.refund_amount), 0) as total_returns_amount,
    s.total_amount - COALESCE(SUM(r.refund_amount), 0) as net_sale_amount,
    COUNT(r.id) as return_count,
    COUNT(DISTINCT r.customer_id) as customers_who_returned
FROM sales s
LEFT JOIN returns r ON s.id = r.sale_id AND r.status = 'completed'
GROUP BY s.id, s.invoice_number, s.sale_date, s.total_amount;

-- View for return reasons analysis
CREATE OR REPLACE VIEW return_reasons_analysis AS
SELECT 
    r.reason,
    r.condition,
    COUNT(r.id) as return_count,
    SUM(r.refund_amount) as total_refund_amount,
    AVG(r.refund_amount) as avg_refund_amount,
    COUNT(DISTINCT r.customer_id) as unique_customers,
    COUNT(DISTINCT r.sale_id) as unique_sales,
    EXTRACT(DAY FROM (MAX(r.return_date) - MIN(r.return_date))) as days_span
FROM returns r
WHERE r.status = 'completed'
GROUP BY r.reason, r.condition
ORDER BY return_count DESC;

-- View for monthly returns summary
CREATE OR REPLACE VIEW monthly_returns_summary AS
SELECT 
    DATE_TRUNC('month', r.return_date) as month,
    COUNT(r.id) as total_returns,
    SUM(r.refund_amount) as total_refund_amount,
    AVG(r.refund_amount) as avg_refund_amount,
    COUNT(DISTINCT r.customer_id) as unique_customers,
    COUNT(CASE WHEN r.return_type = 'full' THEN 1 END) as full_returns,
    COUNT(CASE WHEN r.return_type = 'partial' THEN 1 END) as partial_returns,
    COUNT(CASE WHEN r.refund_method = 'store_credit' THEN 1 END) as store_credit_returns,
    COUNT(CASE WHEN r.refund_method = 'debt_reduction' THEN 1 END) as debt_reduction_returns
FROM returns r
WHERE r.status = 'completed'
GROUP BY DATE_TRUNC('month', r.return_date)
ORDER BY month DESC;

-- ============================================
-- 8. Functions for Debt Integration
-- ============================================

-- Function to adjust debt when return is processed with debt reduction
CREATE OR REPLACE FUNCTION adjust_debt_for_return()
RETURNS TRIGGER AS $$
DECLARE
    current_debt DECIMAL(10,2);
    new_debt DECIMAL(10,2);
BEGIN
    -- Only process if refund method is debt_reduction and debt_id is set
    IF NEW.refund_method = 'debt_reduction' AND NEW.debt_id IS NOT NULL THEN
        -- Get current debt
        SELECT remaining_amount INTO current_debt 
        FROM debts 
        WHERE id = NEW.debt_id;
        
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
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to adjust debt when return is updated
CREATE TRIGGER adjust_debt_trigger AFTER UPDATE OF refund_method, debt_adjustment, debt_id ON returns
    FOR EACH ROW WHEN (OLD.status != 'completed' AND NEW.status = 'completed')
    EXECUTE FUNCTION adjust_debt_for_return();

SELECT 'Enhanced returns system created successfully' as status;