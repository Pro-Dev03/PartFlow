-- PartFlow Enhanced Inventory Schema
-- Individual Items, Conditions, Grades, Movements, Reservations

-- ============================================
-- Inventory Items (Individual items tracking)
-- ============================================
CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    item_code VARCHAR(50) UNIQUE NOT NULL,
    barcode VARCHAR(100) UNIQUE,
    serial_number VARCHAR(100) UNIQUE,
    condition VARCHAR(20) NOT NULL CHECK (condition IN ('NEW', 'USED', 'REFURBISHED', 'DAMAGED', 'FOR_PARTS')),
    grade VARCHAR(20) CHECK (grade IN ('EXCELLENT', 'VERY_GOOD', 'GOOD', 'FAIR', 'POOR')),
    purchase_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    selling_price DECIMAL(10,2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'PURCHASED' CHECK (status IN ('PURCHASED', 'RECEIVED', 'INSPECTION', 'AVAILABLE', 'RESERVED', 'SOLD', 'DAMAGED', 'IN_REPAIR', 'RETURNED', 'WARRANTY', 'FOR_PARTS', 'ARCHIVED')),
    location_id UUID REFERENCES locations(id),
    supplier_id UUID REFERENCES suppliers(id),
    purchase_date DATE,
    sold_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_inventory_items_organization ON inventory_items(organization_id);
CREATE INDEX idx_inventory_items_product ON inventory_items(product_id);
CREATE INDEX idx_inventory_items_barcode ON inventory_items(barcode);
CREATE INDEX idx_inventory_items_serial ON inventory_items(serial_number);
CREATE INDEX idx_inventory_items_status ON inventory_items(status);
CREATE INDEX idx_inventory_items_condition ON inventory_items(condition);
CREATE INDEX idx_inventory_items_location ON inventory_items(location_id);

-- ============================================
-- Locations (Enhanced with hierarchy)
-- ============================================
CREATE TABLE IF NOT EXISTS locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('warehouse', 'shelf', 'box', 'display')),
    parent_id UUID REFERENCES locations(id),
    warehouse_id UUID REFERENCES locations(id),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_locations_organization ON locations(organization_id);
CREATE INDEX idx_locations_parent ON locations(parent_id);
CREATE INDEX idx_locations_warehouse ON locations(warehouse_id);

-- ============================================
-- Inventory Movements (Track all inventory changes)
-- ============================================
CREATE TABLE IF NOT EXISTS inventory_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    item_id UUID REFERENCES inventory_items(id),
    product_id UUID REFERENCES products(id),
    movement_type VARCHAR(20) NOT NULL CHECK (movement_type IN ('PURCHASE', 'SALE', 'RETURN', 'ADJUSTMENT', 'TRANSFER', 'RESERVATION', 'RELEASE', 'DAMAGE', 'REPAIR')),
    quantity INTEGER NOT NULL,
    before_quantity INTEGER NOT NULL DEFAULT 0,
    after_quantity INTEGER NOT NULL DEFAULT 0,
    reference_type VARCHAR(50), -- sale, purchase, return, adjustment, etc.
    reference_id UUID,
    reason TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_inventory_movements_organization ON inventory_movements(organization_id);
CREATE INDEX idx_inventory_movements_item ON inventory_movements(item_id);
CREATE INDEX idx_inventory_movements_product ON inventory_movements(product_id);
CREATE INDEX idx_inventory_movements_type ON inventory_movements(movement_type);
CREATE INDEX idx_inventory_movements_reference ON inventory_movements(reference_type, reference_id);
CREATE INDEX idx_inventory_movements_created_at ON inventory_movements(created_at);

-- ============================================
-- Reservations (Item reservations)
-- ============================================
CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    customer_id UUID REFERENCES customers(id),
    user_id UUID NOT NULL REFERENCES users(id),
    reserved_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'converted', 'cancelled')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_reservations_organization ON reservations(organization_id);
CREATE INDEX idx_reservations_item ON reservations(item_id);
CREATE INDEX idx_reservations_customer ON reservations(customer_id);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_reservations_expires_at ON reservations(expires_at);

-- ============================================
-- Customer Ledger (Track customer financial transactions)
-- ============================================
CREATE TABLE IF NOT EXISTS customer_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('SALE', 'PAYMENT', 'RETURN', 'REFUND', 'ADJUSTMENT')),
    reference_id UUID, -- sale_id, payment_id, return_id, etc.
    amount DECIMAL(10,2) NOT NULL, -- positive for debit, negative for credit
    balance DECIMAL(10,2) NOT NULL, -- running balance
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_customer_ledger_organization ON customer_ledger(organization_id);
CREATE INDEX idx_customer_ledger_customer ON customer_ledger(customer_id);
CREATE INDEX idx_customer_ledger_transaction_type ON customer_ledger(transaction_type);
CREATE INDEX idx_customer_ledger_created_at ON customer_ledger(created_at);

-- ============================================
-- Supplier Ledger (Track supplier financial transactions)
-- ============================================
CREATE TABLE IF NOT EXISTS supplier_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('PURCHASE', 'PAYMENT', 'RETURN', 'ADJUSTMENT')),
    reference_id UUID, -- purchase_id, payment_id, return_id, etc.
    amount DECIMAL(10,2) NOT NULL, -- positive for debit, negative for credit
    balance DECIMAL(10,2) NOT NULL, -- running balance
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_supplier_ledger_organization ON supplier_ledger(organization_id);
CREATE INDEX idx_supplier_ledger_supplier ON supplier_ledger(supplier_id);
CREATE INDEX idx_supplier_ledger_transaction_type ON supplier_ledger(transaction_type);
CREATE INDEX idx_supplier_ledger_created_at ON supplier_ledger(created_at);

-- ============================================
-- Warranty Claims (Detailed warranty tracking)
-- ============================================
CREATE TABLE IF NOT EXISTS warranty_claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    warranty_id UUID NOT NULL REFERENCES warranties(id) ON DELETE CASCADE,
    sale_id UUID NOT NULL REFERENCES sales(id),
    product_id UUID NOT NULL REFERENCES products(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    claim_date DATE NOT NULL DEFAULT CURRENT_DATE,
    issue_description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'in_progress', 'completed')),
    resolution_type VARCHAR(20) CHECK (resolution_type IN ('repair', 'replace', 'refund', 'reject')),
    resolution_notes TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_warranty_claims_organization ON warranty_claims(organization_id);
CREATE INDEX idx_warranty_claims_warranty ON warranty_claims(warranty_id);
CREATE INDEX idx_warranty_claims_sale ON warranty_claims(sale_id);
CREATE INDEX idx_warranty_claims_customer ON warranty_claims(customer_id);
CREATE INDEX idx_warranty_claims_status ON warranty_claims(status);

-- ============================================
-- Inspection Items (Detailed inspection checkpoints)
-- ============================================
CREATE TABLE IF NOT EXISTS inspection_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inspection_id UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
    item_id UUID REFERENCES inventory_items(id),
    checkpoint_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pass', 'fail', 'pending')),
    notes TEXT,
    images JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_inspection_items_inspection ON inspection_items(inspection_id);
CREATE INDEX idx_inspection_items_item ON inspection_items(item_id);

-- ============================================
-- Barcodes (Central barcode management)
-- ============================================
CREATE TABLE IF NOT EXISTS barcodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('EXTERNAL', 'INTERNAL', 'SKU', 'SERIAL', 'ITEM_CODE')),
    product_id UUID REFERENCES products(id),
    inventory_item_id UUID REFERENCES inventory_items(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, code)
);

CREATE INDEX idx_barcodes_organization ON barcodes(organization_id);
CREATE INDEX idx_barcodes_code ON barcodes(code);
CREATE INDEX idx_barcodes_type ON barcodes(type);
CREATE INDEX idx_barcodes_product ON barcodes(product_id);
CREATE INDEX idx_barcodes_item ON barcodes(inventory_item_id);

-- ============================================
-- Functions and Triggers
-- ============================================

-- Trigger for inventory_items updated_at
CREATE TRIGGER update_inventory_items_updated_at BEFORE UPDATE ON inventory_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for locations updated_at
CREATE TRIGGER update_locations_updated_at BEFORE UPDATE ON locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for reservations updated_at
CREATE TRIGGER update_reservations_updated_at BEFORE UPDATE ON reservations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for warranty_claims updated_at
CREATE TRIGGER update_warranty_claims_updated_at BEFORE UPDATE ON warranty_claims
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for barcodes updated_at
CREATE TRIGGER update_barcodes_updated_at BEFORE UPDATE ON barcodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to create inventory movement automatically
CREATE OR REPLACE FUNCTION create_inventory_movement()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO inventory_movements (organization_id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at)
    VALUES (
        NEW.organization_id,
        NEW.id,
        NEW.product_id,
        'PURCHASE',
        1,
        0,
        1,
        'item_creation',
        NEW.id,
        'Initial item creation',
        NEW.created_by,
        NOW()
    );
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to create movement on item creation
CREATE TRIGGER create_movement_on_item_creation AFTER INSERT ON inventory_items
    FOR EACH ROW EXECUTE FUNCTION create_inventory_movement();
