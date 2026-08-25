-- PartFlow Unified Ledger System
-- إنشاء جدول ledger_entries الموحد لتتبع جميع الحركات المالية

-- ============================================
-- Create Unified Ledger Entries Table
-- ============================================
CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- معلومات الهوية
    ledger_type VARCHAR(20) NOT NULL CHECK (ledger_type IN ('CUSTOMER', 'SUPPLIER', 'INVENTORY')),
    entity_id UUID NOT NULL, -- customer_id, supplier_id, or product_id
    
    -- معلومات المعاملة
    transaction_type VARCHAR(50) NOT NULL CHECK (transaction_type IN (
        'SALE', 'PAYMENT', 'RETURN', 'REFUND', 'ADJUSTMENT',
        'PURCHASE', 'PURCHASE_PAYMENT',
        'STOCK_IN', 'STOCK_OUT', 'STOCK_ADJUSTMENT', 'TRANSFER', 'DAMAGED', 'REPAIR'
    )),
    reference_id UUID, -- sale_id, payment_id, purchase_id, etc.
    reference_type VARCHAR(50), -- 'sale', 'payment', 'purchase', etc.
    
    -- المعلومات المالية
    amount DECIMAL(15,2) NOT NULL, -- positive for debit, negative for credit
    balance DECIMAL(15,2) NOT NULL, -- running balance after this transaction
    previous_balance DECIMAL(15,2) DEFAULT 0, -- balance before this transaction
    
    -- معلومات إضافية
    description TEXT,
    metadata JSONB DEFAULT '{}', -- additional data like quantities, unit prices, etc.
    
    -- معلومات التتبع
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- منع التكرار
    CONSTRAINT unique_reference UNIQUE (reference_type, reference_id)
);

-- ============================================
-- Create Indexes for Performance
-- ============================================
CREATE INDEX IF NOT EXISTS idx_ledger_entries_ledger_type ON ledger_entries(ledger_type);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_entity ON ledger_entries(entity_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_transaction_type ON ledger_entries(transaction_type);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_reference ON ledger_entries(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_created_at ON ledger_entries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_entity_created ON ledger_entries(entity_id, created_at DESC);

-- ============================================
-- Create Customer Ledger View
-- ============================================
CREATE OR REPLACE VIEW customer_ledger_view AS
SELECT 
    le.id,
    le.entity_id as customer_id,
    le.transaction_type,
    CASE 
        WHEN le.amount > 0 THEN 'debit'
        ELSE 'credit'
    END as type,
    ABS(le.amount) as amount,
    le.balance,
    le.description,
    le.reference_id,
    le.created_at,
    c.name as customer_name,
    c.code as customer_code
FROM ledger_entries le
LEFT JOIN customers c ON c.id = le.entity_id
WHERE le.ledger_type = 'CUSTOMER'
ORDER BY le.created_at DESC;

-- ============================================
-- Create Supplier Ledger View
-- ============================================
CREATE OR REPLACE VIEW supplier_ledger_view AS
SELECT 
    le.id,
    le.entity_id as supplier_id,
    le.transaction_type,
    CASE 
        WHEN le.amount > 0 THEN 'debit'
        ELSE 'credit'
    END as type,
    ABS(le.amount) as amount,
    le.balance,
    le.description,
    le.reference_id,
    le.created_at,
    s.name as supplier_name,
    s.code as supplier_code
FROM ledger_entries le
LEFT JOIN suppliers s ON s.id = le.entity_id
WHERE le.ledger_type = 'SUPPLIER'
ORDER BY le.created_at DESC;

-- ============================================
-- Create Inventory Ledger View
-- ============================================
CREATE OR REPLACE VIEW inventory_ledger_view AS
SELECT 
    le.id,
    le.entity_id as product_id,
    le.transaction_type,
    ABS(le.amount) as quantity,
    le.balance as current_quantity,
    le.previous_balance as previous_quantity,
    le.description,
    le.reference_id,
    le.reference_type,
    le.metadata,
    le.created_at,
    p.name as product_name,
    p.sku as product_sku,
    p.barcode as product_barcode
FROM ledger_entries le
LEFT JOIN products p ON p.id = le.entity_id
WHERE le.ledger_type = 'INVENTORY'
ORDER BY le.created_at DESC;

-- ============================================
-- Create Function to Update Customer Balance
-- ============================================
CREATE OR REPLACE FUNCTION update_customer_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.ledger_type = 'CUSTOMER' THEN
        UPDATE customers 
        SET current_balance = NEW.balance,
            updated_at = NOW()
        WHERE id = NEW.entity_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Create Function to Update Supplier Balance
-- ============================================
CREATE OR REPLACE FUNCTION update_supplier_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.ledger_type = 'SUPPLIER' THEN
        UPDATE suppliers 
        SET current_balance = NEW.balance,
            updated_at = NOW()
        WHERE id = NEW.entity_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Create Function to Update Inventory Quantity
-- ============================================
CREATE OR REPLACE FUNCTION update_inventory_quantity()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.ledger_type = 'INVENTORY' THEN
        UPDATE inventory 
        SET quantity = NEW.balance,
            updated_at = NOW()
        WHERE product_id = NEW.entity_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Apply Triggers
-- ============================================
DROP TRIGGER IF EXISTS trigger_update_customer_balance ON ledger_entries;
CREATE TRIGGER trigger_update_customer_balance
    AFTER INSERT ON ledger_entries
    FOR EACH ROW
    EXECUTE FUNCTION update_customer_balance();

DROP TRIGGER IF EXISTS trigger_update_supplier_balance ON ledger_entries;
CREATE TRIGGER trigger_update_supplier_balance
    AFTER INSERT ON ledger_entries
    FOR EACH ROW
    EXECUTE FUNCTION update_supplier_balance();

DROP TRIGGER IF EXISTS trigger_update_inventory_quantity ON ledger_entries;
CREATE TRIGGER trigger_update_inventory_quantity
    AFTER INSERT ON ledger_entries
    FOR EACH ROW
    EXECUTE FUNCTION update_inventory_quantity();

-- ============================================
-- Add Comments
-- ============================================
COMMENT ON TABLE ledger_entries IS 'Unified ledger table for tracking all financial and inventory transactions';
COMMENT ON COLUMN ledger_entries.ledger_type IS 'Type of ledger: CUSTOMER, SUPPLIER, or INVENTORY';
COMMENT ON COLUMN ledger_entries.entity_id IS 'ID of the related entity (customer, supplier, or product)';
COMMENT ON COLUMN ledger_entries.transaction_type IS 'Type of transaction: SALE, PAYMENT, PURCHASE, STOCK_IN, etc.';
COMMENT ON COLUMN ledger_entries.amount IS 'Transaction amount (positive for debit, negative for credit)';
COMMENT ON COLUMN ledger_entries.balance IS 'Running balance after this transaction';
COMMENT ON COLUMN ledger_entries.previous_balance IS 'Balance before this transaction';
COMMENT ON COLUMN ledger_entries.metadata IS 'Additional JSON data for the transaction';
