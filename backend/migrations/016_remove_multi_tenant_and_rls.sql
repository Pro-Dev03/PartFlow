-- ============================================
-- Remove Multi-tenant System and RLS Policies
-- ============================================
-- This migration removes the entire multi-tenant system and Row Level Security
-- Converting the database to a single-tenant system

-- ============================================
-- Step 1: Drop all RLS policies from all tables
-- ============================================

-- Drop policies from organizations table
DROP POLICY IF EXISTS "Owners can view their organization" ON organizations;
DROP POLICY IF EXISTS "Owners can update organization" ON organizations;

-- Drop policies from users table
DROP POLICY IF EXISTS "Owners can view organization users" ON users;
DROP POLICY IF EXISTS "Owners can manage users" ON users;

-- Drop policies from products table
DROP POLICY IF EXISTS "Owners can view organization products" ON products;
DROP POLICY IF EXISTS "Owners can manage organization products" ON products;

-- Drop policies from categories table
DROP POLICY IF EXISTS "Owners can view organization categories" ON categories;
DROP POLICY IF EXISTS "Owners can manage categories" ON categories;

-- Drop policies from inventory table
DROP POLICY IF EXISTS "Owners can view organization inventory" ON inventory;
DROP POLICY IF EXISTS "Owners can manage organization inventory" ON inventory;

-- Drop policies from warehouses table
DROP POLICY IF EXISTS "Owners can view organization warehouses" ON warehouses;
DROP POLICY IF EXISTS "Owners can manage warehouses" ON warehouses;

-- Drop policies from customers table
DROP POLICY IF EXISTS "Owners can view organization customers" ON customers;
DROP POLICY IF EXISTS "Owners can manage organization customers" ON customers;

-- Drop policies from suppliers table
DROP POLICY IF EXISTS "Owners can view organization suppliers" ON suppliers;
DROP POLICY IF EXISTS "Owners can manage organization suppliers" ON suppliers;

-- Drop policies from sales table
DROP POLICY IF EXISTS "Owners can view organization sales" ON sales;
DROP POLICY IF EXISTS "Owners can manage organization sales" ON sales;

-- Drop policies from sale_items table
DROP POLICY IF EXISTS "Owners can view organization sale items" ON sale_items;
DROP POLICY IF EXISTS "Owners can manage organization sale items" ON sale_items;

-- Drop policies from purchases table
DROP POLICY IF EXISTS "Owners can view organization purchases" ON purchases;
DROP POLICY IF EXISTS "Owners can manage organization purchases" ON purchases;

-- Drop policies from purchase_items table
DROP POLICY IF EXISTS "Owners can view organization purchase items" ON purchase_items;
DROP POLICY IF EXISTS "Owners can manage organization purchase items" ON purchase_items;

-- Drop policies from payments table
DROP POLICY IF EXISTS "Owners can view organization payments" ON payments;
DROP POLICY IF EXISTS "Owners can manage organization payments" ON payments;

-- Drop policies from debts table
DROP POLICY IF EXISTS "Owners can view organization debts" ON debts;
DROP POLICY IF EXISTS "Owners can manage organization debts" ON debts;

-- Drop policies from expenses table
DROP POLICY IF EXISTS "Owners can view organization expenses" ON expenses;
DROP POLICY IF EXISTS "Owners can manage organization expenses" ON expenses;

-- Drop policies from returns table
DROP POLICY IF EXISTS "Owners can view organization returns" ON returns;
DROP POLICY IF EXISTS "Owners can manage organization returns" ON returns;

-- Drop policies from inspections table
DROP POLICY IF EXISTS "Owners can view organization inspections" ON inspections;
DROP POLICY IF EXISTS "Owners can manage organization inspections" ON inspections;

-- Drop policies from audit_logs table
DROP POLICY IF EXISTS "Owners can view organization audit logs" ON audit_logs;
DROP POLICY IF EXISTS "System can insert audit logs" ON audit_logs;

-- Drop policies from notifications table
DROP POLICY IF EXISTS "Owners can view own notifications" ON notifications;
DROP POLICY IF EXISTS "Users can update own notifications" ON notifications;
DROP POLICY IF EXISTS "System can insert notifications" ON notifications;

-- Drop policies from automations table
DROP POLICY IF EXISTS "Owners can view organization automations" ON automations;
DROP POLICY IF EXISTS "Owners can manage automations" ON automations;

-- ============================================
-- Step 2: Disable RLS on all tables
-- ============================================

ALTER TABLE organizations DISABLE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE products DISABLE ROW LEVEL SECURITY;
ALTER TABLE categories DISABLE ROW LEVEL SECURITY;
ALTER TABLE inventory DISABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses DISABLE ROW LEVEL SECURITY;
ALTER TABLE customers DISABLE ROW LEVEL SECURITY;
ALTER TABLE suppliers DISABLE ROW LEVEL SECURITY;
ALTER TABLE sales DISABLE ROW LEVEL SECURITY;
ALTER TABLE sale_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE purchases DISABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE payments DISABLE ROW LEVEL SECURITY;
ALTER TABLE debts DISABLE ROW LEVEL SECURITY;
ALTER TABLE expenses DISABLE ROW LEVEL SECURITY;
ALTER TABLE returns DISABLE ROW LEVEL SECURITY;
ALTER TABLE inspections DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE automations DISABLE ROW LEVEL SECURITY;

-- Tables from migration 003
ALTER TABLE inventory_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE locations DISABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_movements DISABLE ROW LEVEL SECURITY;
ALTER TABLE reservations DISABLE ROW LEVEL SECURITY;
ALTER TABLE customer_ledger DISABLE ROW LEVEL SECURITY;
ALTER TABLE supplier_ledger DISABLE ROW LEVEL SECURITY;
ALTER TABLE inspection_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE barcodes DISABLE ROW LEVEL SECURITY;

-- Tables from migration 004
ALTER TABLE brands DISABLE ROW LEVEL SECURITY;

-- Tables from migration 008
ALTER TABLE idempotency_keys DISABLE ROW LEVEL SECURITY;

-- Tables from migration 012
ALTER TABLE trade_ins DISABLE ROW LEVEL SECURITY;

-- Tables from migration 013
ALTER TABLE part_types DISABLE ROW LEVEL SECURITY;
ALTER TABLE part_specifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE type_specifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE item_specification_values DISABLE ROW LEVEL SECURITY;

-- ============================================
-- Step 3: Remove organization_id from all tables
-- ============================================

-- From users table
ALTER TABLE users DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_users_organization;

-- From categories table
ALTER TABLE categories DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_categories_organization;
ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_organization_id_fkey;

-- From products table
ALTER TABLE products DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_products_organization;
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_organization_id_fkey;

-- From warehouses table
ALTER TABLE warehouses DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_warehouses_organization;
ALTER TABLE warehouses DROP CONSTRAINT IF EXISTS warehouses_organization_id_fkey;

-- From inventory table
ALTER TABLE inventory DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_inventory_organization;
ALTER TABLE inventory DROP CONSTRAINT IF EXISTS inventory_organization_id_fkey;

-- From customers table
ALTER TABLE customers DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_customers_organization;
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_organization_id_fkey;

-- From suppliers table
ALTER TABLE suppliers DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_suppliers_organization;
ALTER TABLE suppliers DROP CONSTRAINT IF EXISTS suppliers_organization_id_fkey;

-- From sales table
ALTER TABLE sales DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_sales_organization;
ALTER TABLE sales DROP CONSTRAINT IF EXISTS sales_organization_id_fkey;

-- From purchases table
ALTER TABLE purchases DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_purchases_organization;
ALTER TABLE purchases DROP CONSTRAINT IF EXISTS purchases_organization_id_fkey;

-- From payments table
ALTER TABLE payments DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_payments_organization;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_organization_id_fkey;

-- From debts table
ALTER TABLE debts DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_debts_organization;
ALTER TABLE debts DROP CONSTRAINT IF EXISTS debts_organization_id_fkey;

-- From expenses table
ALTER TABLE expenses DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_expenses_organization;
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS expenses_organization_id_fkey;

-- From returns table
ALTER TABLE returns DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_returns_organization;
ALTER TABLE returns DROP CONSTRAINT IF EXISTS returns_organization_id_fkey;

-- From inspections table
ALTER TABLE inspections DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_inspections_organization;
ALTER TABLE inspections DROP CONSTRAINT IF EXISTS inspections_organization_id_fkey;

-- From audit_logs table
ALTER TABLE audit_logs DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_audit_logs_organization;
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_organization_id_fkey;

-- From notifications table
ALTER TABLE notifications DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_notifications_organization;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_organization_id_fkey;

-- From automations table
ALTER TABLE automations DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_automations_organization;
ALTER TABLE automations DROP CONSTRAINT IF EXISTS automations_organization_id_fkey;

-- From migration 003 tables
ALTER TABLE inventory_items DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_inventory_items_organization;
ALTER TABLE inventory_items DROP CONSTRAINT IF EXISTS inventory_items_organization_id_fkey;

ALTER TABLE locations DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_locations_organization;
ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_organization_id_fkey;

ALTER TABLE inventory_movements DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_inventory_movements_organization;
ALTER TABLE inventory_movements DROP CONSTRAINT IF EXISTS inventory_movements_organization_id_fkey;

ALTER TABLE reservations DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_reservations_organization;
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS reservations_organization_id_fkey;

ALTER TABLE customer_ledger DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_customer_ledger_organization;
ALTER TABLE customer_ledger DROP CONSTRAINT IF EXISTS customer_ledger_organization_id_fkey;

ALTER TABLE supplier_ledger DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_supplier_ledger_organization;
ALTER TABLE supplier_ledger DROP CONSTRAINT IF EXISTS supplier_ledger_organization_id_fkey;

ALTER TABLE barcodes DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_barcodes_organization;
ALTER TABLE barcodes DROP CONSTRAINT IF EXISTS barcodes_organization_id_fkey;

-- From migration 004 tables
ALTER TABLE brands DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_brands_organization;
ALTER TABLE brands DROP CONSTRAINT IF EXISTS brands_organization_id_fkey;

-- From migration 008 tables
ALTER TABLE idempotency_keys DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_idempotency_keys_organization;
ALTER TABLE idempotency_keys DROP CONSTRAINT IF EXISTS idempotency_keys_organization_id_fkey;

-- From migration 012 tables
ALTER TABLE trade_ins DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_trade_ins_organization_id;
ALTER TABLE trade_ins DROP CONSTRAINT IF EXISTS trade_ins_organization_id_fkey;

-- From migration 013 tables
ALTER TABLE part_types DROP COLUMN IF EXISTS organization_id;
DROP INDEX IF EXISTS idx_part_types_org;
ALTER TABLE part_types DROP CONSTRAINT IF EXISTS part_types_organization_id_fkey;

-- ============================================
-- Step 4: Update unique constraints that included organization_id
-- ============================================

-- Update categories unique constraint
ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_organization_id_name_key;
ALTER TABLE categories ADD CONSTRAINT categories_name_key UNIQUE (name);

-- Update products unique constraint
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_organization_id_sku_key;
ALTER TABLE products ADD CONSTRAINT products_sku_key UNIQUE (sku);

-- Update customers unique constraint
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_organization_id_code_key;
ALTER TABLE customers ADD CONSTRAINT customers_code_key UNIQUE (code);

-- Update suppliers unique constraint
ALTER TABLE suppliers DROP CONSTRAINT IF EXISTS suppliers_organization_id_code_key;
ALTER TABLE suppliers ADD CONSTRAINT suppliers_code_key UNIQUE (code);

-- Update sales unique constraint
ALTER TABLE sales DROP CONSTRAINT IF EXISTS sales_organization_id_invoice_number_key;
ALTER TABLE sales ADD CONSTRAINT sales_invoice_number_key UNIQUE (invoice_number);

-- Update purchases unique constraint
ALTER TABLE purchases DROP CONSTRAINT IF EXISTS purchases_organization_id_invoice_number_key;
ALTER TABLE purchases ADD CONSTRAINT purchases_invoice_number_key UNIQUE (invoice_number);

-- Update brands unique constraint
ALTER TABLE brands DROP CONSTRAINT IF EXISTS brands_organization_id_name_key;
ALTER TABLE brands ADD CONSTRAINT brands_name_key UNIQUE (name);

-- Update part_types unique constraint
ALTER TABLE part_types DROP CONSTRAINT IF EXISTS part_types_organization_id_name_ar_key;
ALTER TABLE part_types ADD CONSTRAINT part_types_name_ar_key UNIQUE (name_ar);

-- Update barcodes unique constraint
ALTER TABLE barcodes DROP CONSTRAINT IF EXISTS barcodes_organization_id_code_key;
ALTER TABLE barcodes ADD CONSTRAINT barcodes_code_key UNIQUE (code);

-- Update idempotency_keys unique constraint
ALTER TABLE idempotency_keys DROP CONSTRAINT IF EXISTS idempotency_keys_organization_id_idempotency_key_key;
ALTER TABLE idempotency_keys ADD CONSTRAINT idempotency_keys_idempotency_key_key UNIQUE (idempotency_key);

-- ============================================
-- Step 5: Drop the organizations table
-- ============================================
DROP TABLE IF EXISTS organizations CASCADE;

-- ============================================
-- Step 6: Update users table (remove organization_id constraint)
-- ============================================
-- Users table no longer needs organization_id, but we need to handle existing data
-- All users will now be global users in a single-tenant system

-- ============================================
-- Step 7: Update indexes that were organization-based
-- ============================================

-- Drop organization-based compound indexes and create simpler ones
DROP INDEX IF EXISTS idx_sales_organization_date;
CREATE INDEX IF NOT EXISTS idx_sales_date ON sales(sale_date);

-- ============================================
-- Step 8: Verification
-- ============================================

-- Check that RLS is disabled on all tables
DO $$
DECLARE
    table_name text;
    rls_enabled boolean;
BEGIN
    FOR table_name IN 
        SELECT tablename FROM pg_tables WHERE schemaname = 'public'
        AND tablename IN (
            'users', 'products', 'categories', 'inventory', 'warehouses',
            'customers', 'suppliers', 'sales', 'sale_items', 'purchases',
            'purchase_items', 'payments', 'debts', 'expenses', 'returns',
            'inspections', 'audit_logs', 'notifications', 'automations',
            'inventory_items', 'locations', 'inventory_movements', 'reservations',
            'customer_ledger', 'supplier_ledger', 'inspection_items', 'barcodes',
            'brands', 'idempotency_keys', 'trade_ins', 'part_types',
            'part_specifications', 'type_specifications', 'item_specification_values'
        )
    LOOP
        SELECT relrowsecurity INTO rls_enabled 
        FROM pg_class 
        WHERE relname = table_name;
        
        IF rls_enabled THEN
            RAISE NOTICE 'RLS still enabled on table: %', table_name;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'RLS check completed';
END $$;

-- Check that organization_id columns are removed
DO $$
DECLARE
    table_name text;
    column_name text;
BEGIN
    FOR table_name, column_name IN 
        SELECT table_name, column_name FROM information_schema.columns
        WHERE column_name = 'organization_id'
        AND table_schema = 'public'
    LOOP
        RAISE NOTICE 'organization_id still exists in table: %', table_name;
    END LOOP;
    
    IF NOT FOUND THEN
        RAISE NOTICE 'All organization_id columns successfully removed';
    END IF;
END $$;

-- Verify organizations table is dropped
DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'organizations') THEN
        RAISE NOTICE 'organizations table still exists';
    ELSE
        RAISE NOTICE 'organizations table successfully dropped';
    END IF;
END $$;

SELECT 'Multi-tenant system and RLS policies removal completed' as status;
