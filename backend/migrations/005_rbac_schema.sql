-- PartFlow RBAC Schema
-- Role-Based Access Control

-- ============================================
-- Permissions Table
-- ============================================
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action);

-- ============================================
-- Role Permissions Table
-- ============================================
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_name VARCHAR(100) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_name)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_name);

-- ============================================
-- Update roles table to use JSONB for permissions (optional)
-- ============================================
-- This is for backward compatibility - role_permissions is the authoritative source
ALTER TABLE roles 
ADD COLUMN IF NOT EXISTS permissions_jsonb JSONB DEFAULT '[]';

-- ============================================
-- Function to sync role_permissions to roles.permissions_jsonb
-- ============================================
CREATE OR REPLACE FUNCTION sync_role_permissions()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE roles 
        SET permissions_jsonb = (
            SELECT COALESCE(jsonb_agg(permission_name), '[]'::jsonb)
            FROM role_permissions 
            WHERE role_id = NEW.role_id
        )
        WHERE id = NEW.role_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE roles 
        SET permissions_jsonb = (
            SELECT COALESCE(jsonb_agg(permission_name), '[]'::jsonb)
            FROM role_permissions 
            WHERE role_id = OLD.role_id
        )
        WHERE id = OLD.role_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for role_permissions
CREATE TRIGGER sync_role_permissions_insert
AFTER INSERT ON role_permissions
FOR EACH ROW EXECUTE FUNCTION sync_role_permissions();

CREATE TRIGGER sync_role_permissions_delete
AFTER DELETE ON role_permissions
FOR EACH ROW EXECUTE FUNCTION sync_role_permissions();

-- ============================================
-- Insert standard permissions
-- ============================================
INSERT INTO permissions (name, description, resource, action) VALUES
-- Products
('products.read', 'Read products', 'products', 'read'),
('products.create', 'Create products', 'products', 'create'),
('products.update', 'Update products', 'products', 'update'),
('products.delete', 'Delete products', 'products', 'delete'),
('products.archive', 'Archive products', 'products', 'archive'),

-- Inventory
('inventory.read', 'Read inventory', 'inventory', 'read'),
('inventory.adjust', 'Adjust inventory', 'inventory', 'adjust'),
('inventory.transfer', 'Transfer inventory', 'inventory', 'transfer'),
('inventory.inspect', 'Inspect inventory items', 'inventory', 'inspect'),

-- Sales
('sales.read', 'Read sales', 'sales', 'read'),
('sales.create', 'Create sales', 'sales', 'create'),
('sales.cancel', 'Cancel sales', 'sales', 'cancel'),
('sales.refund', 'Refund sales', 'sales', 'refund'),

-- Customers
('customers.read', 'Read customers', 'customers', 'read'),
('customers.create', 'Create customers', 'customers', 'create'),
('customers.update', 'Update customers', 'customers', 'update'),
('customers.delete', 'Delete customers', 'customers', 'delete'),

-- Debts
('debts.read', 'Read debts', 'debts', 'read'),
('debts.manage', 'Manage debts', 'debts', 'manage'),

-- Suppliers
('suppliers.read', 'Read suppliers', 'suppliers', 'read'),
('suppliers.create', 'Create suppliers', 'suppliers', 'create'),
('suppliers.update', 'Update suppliers', 'suppliers', 'update'),
('suppliers.delete', 'Delete suppliers', 'suppliers', 'delete'),

-- Purchases
('purchases.read', 'Read purchases', 'purchases', 'read'),
('purchases.create', 'Create purchases', 'purchases', 'create'),
('purchases.receive', 'Receive purchases', 'purchases', 'receive'),

-- Expenses
('expenses.read', 'Read expenses', 'expenses', 'read'),
('expenses.create', 'Create expenses', 'expenses', 'create'),
('expenses.update', 'Update expenses', 'expenses', 'update'),
('expenses.delete', 'Delete expenses', 'expenses', 'delete'),

-- Returns
('returns.read', 'Read returns', 'returns', 'read'),
('returns.create', 'Create returns', 'returns', 'create'),
('returns.approve', 'Approve returns', 'returns', 'approve'),

-- Warranty
('warranties.read', 'Read warranties', 'warranties', 'read'),
('warranties.claim', 'Claim warranties', 'warranties', 'claim'),

-- Reports
('reports.read', 'Read reports', 'reports', 'read'),
('reports.export', 'Export reports', 'reports', 'export'),

-- Users
('users.read', 'Read users', 'users', 'read'),
('users.create', 'Create users', 'users', 'create'),
('users.update', 'Update users', 'users', 'update'),
('users.delete', 'Delete users', 'users', 'delete'),

-- Settings
('settings.manage', 'Manage settings', 'settings', 'manage'),

-- Audit
('audit.read', 'Read audit logs', 'audit', 'read')
ON CONFLICT (name) DO NOTHING;
