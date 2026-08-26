-- PartFlow Customer Acquisition System
-- Support for buying used parts from customers (USED-PARTS-ACQUISITION.md)

-- ============================================
-- Acquisitions Table (Unified acquisition system)
-- ============================================
CREATE TABLE IF NOT EXISTS acquisitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    type VARCHAR(20) NOT NULL CHECK (type IN ('SUPPLIER', 'CUSTOMER')),
    acquisition_date DATE NOT NULL,
    
    -- Seller Information (can be supplier or customer)
    supplier_id UUID REFERENCES suppliers(id),
    customer_id UUID REFERENCES customers(id),
    
    -- Financial Information
    total_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    paid_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'payable' CHECK (payment_status IN ('paid', 'payable', 'partial', 'overdue')),
    
    -- Status & Workflow
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'acquired', 'inspection', 'approved', 'rejected', 'cancelled', 'reversed')),
    
    -- Notes & Metadata
    notes TEXT,
    user_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Reversal tracking (ARCHITECTURE-PRINCIPLES.md)
    reversed_at TIMESTAMP WITH TIME ZONE,
    reversed_by UUID REFERENCES users(id),
    reversal_reason TEXT,

    -- Constraint: Either supplier_id or customer_id must be set
    CONSTRAINT acquisitions_seller_constraint CHECK (
        (type = 'SUPPLIER' AND supplier_id IS NOT NULL AND customer_id IS NULL) OR
        (type = 'CUSTOMER' AND customer_id IS NOT NULL AND supplier_id IS NULL)
    )
);

CREATE INDEX idx_acquisitions_type ON acquisitions(type);
CREATE INDEX idx_acquisitions_supplier ON acquisitions(supplier_id);
CREATE INDEX idx_acquisitions_customer ON acquisitions(customer_id);
CREATE INDEX idx_acquisitions_status ON acquisitions(status);
CREATE INDEX idx_acquisitions_payment_status ON acquisitions(payment_status);
CREATE INDEX idx_acquisitions_date ON acquisitions(acquisition_date);

-- ============================================
-- Acquisition Items Table
-- ============================================
CREATE TABLE IF NOT EXISTS acquisition_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    acquisition_id UUID NOT NULL REFERENCES acquisitions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    serial_number VARCHAR(100),
    condition VARCHAR(20) NOT NULL CHECK (condition IN ('new', 'used', 'refurbished')),
    grade VARCHAR(20) CHECK (grade IN ('excellent', 'very_good', 'good', 'fair', 'poor')),
    unit_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    
    -- Inspection Status
    inspection_id UUID REFERENCES inspections(id),
    inspection_status VARCHAR(20) DEFAULT 'pending' CHECK (inspection_status IN ('pending', 'passed', 'failed', 'needs_repair')),
    
    -- Inventory Status
    inventory_item_id UUID REFERENCES inventory_items(id),
    item_status VARCHAR(20) DEFAULT 'acquired' CHECK (item_status IN ('acquired', 'inspection', 'available', 'sold', 'rejected', 'for_parts', 'archived')),
    
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_acquisition_items_acquisition ON acquisition_items(acquisition_id);
CREATE INDEX idx_acquisition_items_product ON acquisition_items(product_id);
CREATE INDEX idx_acquisition_items_inspection ON acquisition_items(inspection_id);
CREATE INDEX idx_acquisition_items_inventory ON acquisition_items(inventory_item_id);
CREATE INDEX idx_acquisition_items_status ON acquisition_items(item_status);

-- ============================================
-- Seller Payments Table (Payments to customers who sold items)
-- ============================================
CREATE TABLE IF NOT EXISTS seller_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    acquisition_id UUID NOT NULL REFERENCES acquisitions(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(50) NOT NULL, -- cash, transfer, etc.
    payment_date DATE NOT NULL,
    notes TEXT,
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_seller_payments_acquisition ON seller_payments(acquisition_id);
CREATE INDEX idx_seller_payments_customer ON seller_payments(customer_id);
CREATE INDEX idx_seller_payments_date ON seller_payments(payment_date);

-- ============================================
-- Acquisition Reversals Table
-- ============================================
CREATE TABLE IF NOT EXISTS acquisition_reversals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    acquisition_id UUID NOT NULL REFERENCES acquisitions(id),
    reason TEXT NOT NULL,
    reversed_by UUID NOT NULL REFERENCES users(id),
    reversed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    original_total DECIMAL(10,2) NOT NULL,
    inventory_adjustment_ids UUID[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_acquisition_reversals_acquisition ON acquisition_reversals(acquisition_id);
CREATE INDEX idx_acquisition_reversals_reversed_by ON acquisition_reversals(reversed_by);

-- ============================================
-- Repair Costs Table (Track repair costs per item)
-- ============================================
CREATE TABLE IF NOT EXISTS item_repair_costs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
    acquisition_item_id UUID REFERENCES acquisition_items(id),
    repair_date DATE NOT NULL,
    repair_type VARCHAR(50) NOT NULL, -- cleaning, part_replacement, fan_replacement, battery_replacement, etc.
    cost DECIMAL(10,2) NOT NULL,
    description TEXT,
    performed_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_item_repair_costs_item ON item_repair_costs(inventory_item_id);
CREATE INDEX idx_item_repair_costs_acquisition_item ON item_repair_costs(acquisition_item_id);
CREATE INDEX idx_item_repair_costs_date ON item_repair_costs(repair_date);

-- ============================================
-- Item History Timeline Table (Complete item lifecycle)
-- ============================================
CREATE TABLE IF NOT EXISTS item_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
    event_type VARCHAR(50) NOT NULL, -- acquired, inspection_passed, inspection_failed, repair, priced, sold, returned, etc.
    event_date TIMESTAMP WITH TIME ZONE NOT NULL,
    reference_type VARCHAR(50), -- acquisition, inspection, repair, sale, etc.
    reference_id UUID,
    description TEXT,
    metadata JSONB DEFAULT '{}', -- Additional event-specific data
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_item_history_item ON item_history(inventory_item_id);
CREATE INDEX idx_item_history_type ON item_history(event_type);
CREATE INDEX idx_item_history_date ON item_history(event_date);
CREATE INDEX idx_item_history_reference ON item_history(reference_type, reference_id);

-- ============================================
-- Functions and Triggers
-- ============================================

-- Trigger for acquisitions updated_at
CREATE TRIGGER update_acquisitions_updated_at BEFORE UPDATE ON acquisitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for acquisition_items updated_at
CREATE TRIGGER update_acquisition_items_updated_at BEFORE UPDATE ON acquisition_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to calculate total cost of acquisition items
CREATE OR REPLACE FUNCTION calculate_acquisition_total()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE acquisitions 
    SET total_cost = (
        SELECT COALESCE(SUM(total_cost), 0) 
        FROM acquisition_items 
        WHERE acquisition_id = NEW.acquisition_id
    )
    WHERE id = NEW.acquisition_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to update acquisition total when items change
CREATE TRIGGER update_acquisition_total AFTER INSERT OR UPDATE OR DELETE ON acquisition_items
    FOR EACH ROW EXECUTE FUNCTION calculate_acquisition_total();

-- Function to create item history entry
CREATE OR REPLACE FUNCTION create_item_history_entry()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO item_history (inventory_item_id, event_type, event_date, reference_type, reference_id, description, metadata, created_by, created_at)
    VALUES (
        NEW.inventory_item_id,
        'acquired',
        NOW(),
        'acquisition',
        NEW.acquisition_id,
        'Item acquired through ' || (SELECT type FROM acquisitions WHERE id = NEW.acquisition_id),
        jsonb_build_object(
            'acquisition_id', NEW.acquisition_id,
            'condition', NEW.condition,
            'grade', NEW.grade,
            'cost', NEW.unit_cost
        ),
        (SELECT user_id FROM acquisitions WHERE id = NEW.acquisition_id),
        NOW()
    );
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to create history when item is linked to acquisition
CREATE TRIGGER create_item_history_on_link AFTER UPDATE OF inventory_item_id ON acquisition_items
    FOR EACH ROW WHEN (NEW.inventory_item_id IS NOT NULL AND OLD.inventory_item_id IS NULL)
    EXECUTE FUNCTION create_item_history_entry();

-- ============================================
-- Views for Reporting
-- ============================================

-- View for used parts aging analysis
CREATE OR REPLACE VIEW used_parts_aging AS
SELECT 
    ii.id as item_id,
    ai.acquisition_id,
    a.acquisition_date,
    (CURRENT_DATE - a.acquisition_date)::integer as days_in_stock,
    ii.status,
    ii.condition,
    ii.purchase_cost as cost,
    ii.selling_price as current_price,
    CASE 
        WHEN (CURRENT_DATE - a.acquisition_date)::integer <= 30 THEN 'fresh'
        WHEN (CURRENT_DATE - a.acquisition_date)::integer <= 60 THEN 'normal'
        WHEN (CURRENT_DATE - a.acquisition_date)::integer <= 90 THEN 'aged'
        ELSE 'long_aged'
    END as aging_category,
    CASE 
        WHEN (CURRENT_DATE - a.acquisition_date)::integer > 90 THEN 'critical'
        WHEN (CURRENT_DATE - a.acquisition_date)::integer > 60 THEN 'warning'
        ELSE 'none'
    END as alert_level
FROM inventory_items ii
JOIN acquisition_items ai ON ii.id = ai.inventory_item_id
JOIN acquisitions a ON ai.acquisition_id = a.id
WHERE a.type = 'CUSTOMER' AND ii.status IN ('AVAILABLE', 'RESERVED');

-- View for seller balances (amounts store owes to customers)
CREATE OR REPLACE VIEW seller_balances AS
SELECT 
    a.customer_id,
    c.name as customer_name,
    c.phone,
    COUNT(a.id) as total_acquisitions,
    SUM(a.total_cost) as total_owed,
    SUM(a.paid_amount) as total_paid,
    SUM(a.total_cost - a.paid_amount) as remaining_balance,
    CASE 
        WHEN SUM(a.total_cost - a.paid_amount) > 0 THEN 'payable'
        ELSE 'settled'
    END as payment_status
FROM acquisitions a
JOIN customers c ON a.customer_id = c.id
WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
GROUP BY a.customer_id, c.name, c.phone
HAVING SUM(a.total_cost - a.paid_amount) > 0;

-- View for customer acquisition summary
CREATE OR REPLACE VIEW customer_acquisition_summary AS
SELECT 
    a.customer_id,
    c.name as customer_name,
    COUNT(DISTINCT a.id) as total_acquisitions,
    COUNT(DISTINCT ai.id) as total_items,
    SUM(a.total_cost) as total_spent,
    AVG(a.total_cost) as avg_acquisition_value,
    COUNT(CASE WHEN ai.inspection_status = 'passed' THEN 1 END) as items_passed,
    COUNT(CASE WHEN ai.inspection_status = 'failed' THEN 1 END) as items_failed,
    COUNT(CASE WHEN ai.item_status = 'sold' THEN 1 END) as items_sold,
    SUM(CASE WHEN ai.item_status = 'sold' THEN ii.selling_price - ii.purchase_cost ELSE 0 END) as total_profit
FROM acquisitions a
JOIN customers c ON a.customer_id = c.id
JOIN acquisition_items ai ON a.id = ai.acquisition_id
LEFT JOIN inventory_items ii ON ai.inventory_item_id = ii.id
WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
GROUP BY a.customer_id, c.name;