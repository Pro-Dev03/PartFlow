CREATE TABLE IF NOT EXISTS supplier_returns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_return_id UUID REFERENCES returns(id),
    sale_id UUID REFERENCES sales(id),
    purchase_id UUID NOT NULL REFERENCES purchases(id),
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    return_number VARCHAR(50) NOT NULL UNIQUE,
    status VARCHAR(30) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','PENDING','SHIPPED','RECEIVED','COMPLETED','REJECTED')),
    reason TEXT NOT NULL,
    refund_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    return_reason TEXT,
    return_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS supplier_return_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_return_id UUID REFERENCES returns(id),
    sale_id UUID REFERENCES sales(id),
    sale_item_id UUID REFERENCES sale_items(id),
    inventory_item_id UUID REFERENCES inventory_items(id),
    supplier_return_id UUID NOT NULL REFERENCES supplier_returns(id) ON DELETE CASCADE,
    purchase_item_id UUID NOT NULL REFERENCES purchase_items(id),
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_cost DECIMAL(10,2) NOT NULL,
    barcode VARCHAR(100),
    serial_number VARCHAR(100),
    purchase_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    return_reason TEXT,
    return_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_supplier_returns_customer_return_active
    ON supplier_returns(customer_return_id)
    WHERE customer_return_id IS NOT NULL AND status IN ('PENDING','SHIPPED','RECEIVED');

CREATE UNIQUE INDEX IF NOT EXISTS idx_supplier_return_items_customer_return_item
    ON supplier_return_items(customer_return_id, inventory_item_id)
    WHERE customer_return_id IS NOT NULL AND inventory_item_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_supplier_returns_purchase ON supplier_returns(purchase_id);
CREATE INDEX IF NOT EXISTS idx_supplier_returns_status ON supplier_returns(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_supplier_return_item_once
    ON supplier_return_items(supplier_return_id, purchase_item_id);
