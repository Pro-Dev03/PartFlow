-- Row Level Security (RLS) Policies for PartFlow
-- Multi-tenant data isolation

-- Enable RLS on all tables
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory ENABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses ENABLE ROW LEVEL SECURITY;
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE suppliers ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales ENABLE ROW LEVEL SECURITY;
ALTER TABLE sale_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchases ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE debts ENABLE ROW LEVEL SECURITY;
ALTER TABLE expenses ENABLE ROW LEVEL SECURITY;
ALTER TABLE returns ENABLE ROW LEVEL SECURITY;
ALTER TABLE warranties ENABLE ROW LEVEL SECURITY;
ALTER TABLE inspections ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE automations ENABLE ROW LEVEL SECURITY;

-- ============================================
-- Organizations Policies
-- ============================================
-- Organization members can view their organization
CREATE POLICY "Users can view their organization" ON organizations
    FOR SELECT
    USING (
        id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Organization owners can update their organization
CREATE POLICY "Owners can update organization" ON organizations
    FOR UPDATE
    USING (
        id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Users Policies
-- ============================================
-- Users can view users in their organization
CREATE POLICY "Users can view organization users" ON users
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can update their own profile
CREATE POLICY "Users can update own profile" ON users
    FOR UPDATE
    USING (id = auth.uid());

-- Admins can manage users in their organization
CREATE POLICY "Admins can manage users" ON users
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Roles Policies
-- ============================================
-- Users can view roles in their organization
CREATE POLICY "Users can view organization roles" ON roles
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Admins can manage roles in their organization
CREATE POLICY "Admins can manage roles" ON roles
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Products Policies
-- ============================================
-- Users can view products in their organization
CREATE POLICY "Users can view organization products" ON products
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage products in their organization
CREATE POLICY "Users can manage organization products" ON products
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Categories Policies
-- ============================================
-- Users can view categories in their organization
CREATE POLICY "Users can view organization categories" ON categories
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Admins can manage categories
CREATE POLICY "Admins can manage categories" ON categories
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Inventory Policies
-- ============================================
-- Users can view inventory in their organization
CREATE POLICY "Users can view organization inventory" ON inventory
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage inventory
CREATE POLICY "Users can manage organization inventory" ON inventory
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Warehouses Policies
-- ============================================
-- Users can view warehouses in their organization
CREATE POLICY "Users can view organization warehouses" ON warehouses
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Admins can manage warehouses
CREATE POLICY "Admins can manage warehouses" ON warehouses
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Customers Policies
-- ============================================
-- Users can view customers in their organization
CREATE POLICY "Users can view organization customers" ON customers
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage customers
CREATE POLICY "Users can manage organization customers" ON customers
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Suppliers Policies
-- ============================================
-- Users can view suppliers in their organization
CREATE POLICY "Users can view organization suppliers" ON suppliers
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage suppliers
CREATE POLICY "Users can manage organization suppliers" ON suppliers
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Sales Policies
-- ============================================
-- Users can view sales in their organization
CREATE POLICY "Users can view organization sales" ON sales
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage sales
CREATE POLICY "Users can manage organization sales" ON sales
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Sale Items Policies
-- ============================================
-- Users can view sale items in their organization
CREATE POLICY "Users can view organization sale items" ON sale_items
    FOR SELECT
    USING (
        sale_id IN (
            SELECT id FROM sales
            WHERE organization_id IN (
                SELECT organization_id FROM users
                WHERE id = auth.uid()
            )
        )
    );

-- Users can manage sale items
CREATE POLICY "Users can manage organization sale items" ON sale_items
    FOR ALL
    USING (
        sale_id IN (
            SELECT id FROM sales
            WHERE organization_id IN (
                SELECT organization_id FROM users
                WHERE id = auth.uid()
            )
        )
    );

-- ============================================
-- Purchases Policies
-- ============================================
-- Users can view purchases in their organization
CREATE POLICY "Users can view organization purchases" ON purchases
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage purchases
CREATE POLICY "Users can manage organization purchases" ON purchases
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Purchase Items Policies
-- ============================================
-- Users can view purchase items in their organization
CREATE POLICY "Users can view organization purchase items" ON purchase_items
    FOR SELECT
    USING (
        purchase_id IN (
            SELECT id FROM purchases
            WHERE organization_id IN (
                SELECT organization_id FROM users
                WHERE id = auth.uid()
            )
        )
    );

-- Users can manage purchase items
CREATE POLICY "Users can manage organization purchase items" ON purchase_items
    FOR ALL
    USING (
        purchase_id IN (
            SELECT id FROM purchases
            WHERE organization_id IN (
                SELECT organization_id FROM users
                WHERE id = auth.uid()
            )
        )
    );

-- ============================================
-- Payments Policies
-- ============================================
-- Users can view payments in their organization
CREATE POLICY "Users can view organization payments" ON payments
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage payments
CREATE POLICY "Users can manage organization payments" ON payments
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Debts Policies
-- ============================================
-- Users can view debts in their organization
CREATE POLICY "Users can view organization debts" ON debts
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage debts
CREATE POLICY "Users can manage organization debts" ON debts
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Expenses Policies
-- ============================================
-- Users can view expenses in their organization
CREATE POLICY "Users can view organization expenses" ON expenses
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage expenses
CREATE POLICY "Users can manage organization expenses" ON expenses
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Returns Policies
-- ============================================
-- Users can view returns in their organization
CREATE POLICY "Users can view organization returns" ON returns
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage returns
CREATE POLICY "Users can manage organization returns" ON returns
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Warranties Policies
-- ============================================
-- Users can view warranties in their organization
CREATE POLICY "Users can view organization warranties" ON warranties
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage warranties
CREATE POLICY "Users can manage organization warranties" ON warranties
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Inspections Policies
-- ============================================
-- Users can view inspections in their organization
CREATE POLICY "Users can view organization inspections" ON inspections
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Users can manage inspections
CREATE POLICY "Users can manage organization inspections" ON inspections
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- ============================================
-- Audit Logs Policies
-- ============================================
-- Users can view audit logs in their organization
CREATE POLICY "Users can view organization audit logs" ON audit_logs
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Only system can insert audit logs
CREATE POLICY "System can insert audit logs" ON audit_logs
    FOR INSERT
    WITH CHECK (true);

-- ============================================
-- Notifications Policies
-- ============================================
-- Users can view their own notifications
CREATE POLICY "Users can view own notifications" ON notifications
    FOR SELECT
    USING (user_id = auth.uid());

-- Users can update their own notifications
CREATE POLICY "Users can update own notifications" ON notifications
    FOR UPDATE
    USING (user_id = auth.uid());

-- System can insert notifications
CREATE POLICY "System can insert notifications" ON notifications
    FOR INSERT
    WITH CHECK (true);

-- ============================================
-- Automations Policies
-- ============================================
-- Users can view automations in their organization
CREATE POLICY "Users can view organization automations" ON automations
    FOR SELECT
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );

-- Admins can manage automations
CREATE POLICY "Admins can manage automations" ON automations
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM users
            WHERE id = auth.uid()
        )
    );
