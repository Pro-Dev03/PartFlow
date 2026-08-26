-- Migration: Apply ARCHITECTURE-PRINCIPLES.md (Current State + Immutable History + Reverse + Aggregations)
-- This migration adds fields and tables to support the new architecture philosophy
-- Date: 2026-08-26

-- ============================================================
-- 1. Inventory Current State Fields
-- ============================================================

-- Add current state fields to inventory (the actual table name)
ALTER TABLE inventory
ADD COLUMN IF NOT EXISTS current_quantity INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS available_quantity INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS current_cost DECIMAL(15,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS current_value DECIMAL(15,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS last_movement_id UUID;

-- Create index for last_movement_id
CREATE INDEX IF NOT EXISTS idx_inventory_last_movement ON inventory(last_movement_id);

-- ============================================================
-- 2. Enhanced Ledger Entries for Immutable History
-- ============================================================

-- Add fields to ledger_entries for better history tracking (ARCHITECTURE-PRINCIPLES.md)
ALTER TABLE ledger_entries
ADD COLUMN IF NOT EXISTS cost_before DECIMAL(15,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS cost_after DECIMAL(15,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS value_before DECIMAL(15,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS value_after DECIMAL(15,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS reversed_by UUID,
ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS reversal_reason TEXT,
ADD COLUMN IF NOT EXISTS product_id UUID;

-- Create indexes for movement tracking
CREATE INDEX IF NOT EXISTS idx_ledger_reversed ON ledger_entries(is_reversed);
CREATE INDEX IF NOT EXISTS idx_ledger_reversed_by ON ledger_entries(reversed_by);
CREATE INDEX IF NOT EXISTS idx_ledger_product_id ON ledger_entries(product_id);

-- ============================================================
-- 3. Purchase Reversal Tracking
-- ============================================================

-- Add reversal fields to purchases
ALTER TABLE purchases
ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS reversed_by UUID,
ADD COLUMN IF NOT EXISTS reversal_reason TEXT;

-- Create index for purchase reversals
CREATE INDEX IF NOT EXISTS idx_purchases_reversed_by ON purchases(reversed_by);

-- Create purchase_reversals table
CREATE TABLE IF NOT EXISTS purchase_reversals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_id UUID NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    reversed_by UUID NOT NULL,
    reversed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    original_total DECIMAL(15,2) NOT NULL,
    inventory_adjustment_ids UUID[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_purchase_reversals_purchase ON purchase_reversals(purchase_id);
CREATE INDEX IF NOT EXISTS idx_purchase_reversals_reversed_by ON purchase_reversals(reversed_by);

-- ============================================================
-- 4. Payment Reversal Tracking
-- ============================================================

-- Add reversal fields to payments
ALTER TABLE payments
ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS reversed_by UUID,
ADD COLUMN IF NOT EXISTS reversal_reason TEXT,
ADD COLUMN IF NOT EXISTS reversal_payment_id UUID;

-- Create indexes for payment reversals
CREATE INDEX IF NOT EXISTS idx_payments_reversed ON payments(is_reversed);
CREATE INDEX IF NOT EXISTS idx_payments_reversed_by ON payments(reversed_by);
CREATE INDEX IF NOT EXISTS idx_payments_reversal_payment ON payments(reversal_payment_id);

-- Create payment_reversals table
CREATE TABLE IF NOT EXISTS payment_reversals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    reversed_by UUID NOT NULL,
    reversed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    original_amount DECIMAL(15,2) NOT NULL,
    debt_adjustment_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_reversals_payment ON payment_reversals(payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_reversals_reversed_by ON payment_reversals(reversed_by);

-- ============================================================
-- 5. Sales Reversal Tracking (if sales table exists)
-- ============================================================

-- Add reversal fields to sales (if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sales') THEN
        ALTER TABLE sales
        ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP,
        ADD COLUMN IF NOT EXISTS reversed_by UUID,
        ADD COLUMN IF NOT EXISTS reversal_reason TEXT;

        CREATE INDEX IF NOT EXISTS idx_sales_reversed_by ON sales(reversed_by);
    END IF;
END $$;

-- ============================================================
-- 6. Aggregation Tables for Dashboard Performance
-- ============================================================

-- Daily Sales Summary
CREATE TABLE IF NOT EXISTS daily_sales_summary (
    date DATE PRIMARY KEY,
    total_sales BIGINT DEFAULT 0,
    total_revenue DECIMAL(15,2) DEFAULT 0.00,
    total_profit DECIMAL(15,2) DEFAULT 0.00,
    total_customers INTEGER DEFAULT 0,
    average_order_value DECIMAL(15,2) DEFAULT 0.00,
    total_items_sold INTEGER DEFAULT 0,
    cash_sales DECIMAL(15,2) DEFAULT 0.00,
    card_sales DECIMAL(15,2) DEFAULT 0.00,
    debt_sales DECIMAL(15,2) DEFAULT 0.00,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Monthly Sales Summary
CREATE TABLE IF NOT EXISTS monthly_sales_summary (
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    total_sales BIGINT DEFAULT 0,
    total_revenue DECIMAL(15,2) DEFAULT 0.00,
    total_profit DECIMAL(15,2) DEFAULT 0.00,
    total_customers INTEGER DEFAULT 0,
    average_order_value DECIMAL(15,2) DEFAULT 0.00,
    total_items_sold INTEGER DEFAULT 0,
    cash_sales DECIMAL(15,2) DEFAULT 0.00,
    card_sales DECIMAL(15,2) DEFAULT 0.00,
    debt_sales DECIMAL(15,2) DEFAULT 0.00,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (year, month)
);

-- Daily Inventory Summary
CREATE TABLE IF NOT EXISTS daily_inventory_summary (
    date DATE PRIMARY KEY,
    total_items INTEGER DEFAULT 0,
    total_value DECIMAL(15,2) DEFAULT 0.00,
    low_stock_count INTEGER DEFAULT 0,
    out_of_stock_count INTEGER DEFAULT 0,
    new_items_added INTEGER DEFAULT 0,
    items_sold INTEGER DEFAULT 0,
    items_returned INTEGER DEFAULT 0,
    items_damaged INTEGER DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Monthly Inventory Summary
CREATE TABLE IF NOT EXISTS monthly_inventory_summary (
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    total_items INTEGER DEFAULT 0,
    total_value DECIMAL(15,2) DEFAULT 0.00,
    low_stock_count INTEGER DEFAULT 0,
    out_of_stock_count INTEGER DEFAULT 0,
    new_items_added INTEGER DEFAULT 0,
    items_sold INTEGER DEFAULT 0,
    items_returned INTEGER DEFAULT 0,
    items_damaged INTEGER DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (year, month)
);

-- Daily Debt Summary
CREATE TABLE IF NOT EXISTS daily_debt_summary (
    date DATE PRIMARY KEY,
    total_debt DECIMAL(15,2) DEFAULT 0.00,
    new_debt DECIMAL(15,2) DEFAULT 0.00,
    payments_received DECIMAL(15,2) DEFAULT 0.00,
    overdue_debt DECIMAL(15,2) DEFAULT 0.00,
    overdue_count INTEGER DEFAULT 0,
    paid_debt DECIMAL(15,2) DEFAULT 0.00,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Monthly Debt Summary
CREATE TABLE IF NOT EXISTS monthly_debt_summary (
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    total_debt DECIMAL(15,2) DEFAULT 0.00,
    new_debt DECIMAL(15,2) DEFAULT 0.00,
    payments_received DECIMAL(15,2) DEFAULT 0.00,
    overdue_debt DECIMAL(15,2) DEFAULT 0.00,
    overdue_count INTEGER DEFAULT 0,
    paid_debt DECIMAL(15,2) DEFAULT 0.00,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (year, month)
);

-- Daily Profit Summary
CREATE TABLE IF NOT EXISTS daily_profit_summary (
    date DATE PRIMARY KEY,
    gross_profit DECIMAL(15,2) DEFAULT 0.00,
    net_profit DECIMAL(15,2) DEFAULT 0.00,
    total_revenue DECIMAL(15,2) DEFAULT 0.00,
    total_cost DECIMAL(15,2) DEFAULT 0.00,
    profit_margin DECIMAL(15,2) DEFAULT 0.00,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Monthly Profit Summary
CREATE TABLE IF NOT EXISTS monthly_profit_summary (
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    gross_profit DECIMAL(15,2) DEFAULT 0.00,
    net_profit DECIMAL(15,2) DEFAULT 0.00,
    total_revenue DECIMAL(15,2) DEFAULT 0.00,
    total_cost DECIMAL(15,2) DEFAULT 0.00,
    profit_margin DECIMAL(15,2) DEFAULT 0.00,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (year, month)
);

-- ============================================================
-- 7. Performance Indexes for Aggregation Tables
-- ============================================================

-- Sales aggregation indexes
CREATE INDEX IF NOT EXISTS idx_daily_sales_date ON daily_sales_summary(date);
CREATE INDEX IF NOT EXISTS idx_monthly_sales_year_month ON monthly_sales_summary(year, month);

-- Inventory aggregation indexes
CREATE INDEX IF NOT EXISTS idx_daily_inventory_date ON daily_inventory_summary(date);
CREATE INDEX IF NOT EXISTS idx_monthly_inventory_year_month ON monthly_inventory_summary(year, month);

-- Debt aggregation indexes
CREATE INDEX IF NOT EXISTS idx_daily_debt_date ON daily_debt_summary(date);
CREATE INDEX IF NOT EXISTS idx_monthly_debt_year_month ON monthly_debt_summary(year, month);

-- Profit aggregation indexes
CREATE INDEX IF NOT EXISTS idx_daily_profit_date ON daily_profit_summary(date);
CREATE INDEX IF NOT EXISTS idx_monthly_profit_year_month ON monthly_profit_summary(year, month);

-- ============================================================
-- 8. Archive Status Tracking (Future-Ready)
-- ============================================================

-- Create a table to track archive status (for future implementation)
CREATE TABLE IF NOT EXISTS archive_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name TEXT NOT NULL UNIQUE,
    last_archive_date TIMESTAMP,
    archive_threshold_days INTEGER DEFAULT 730, -- 2 years default
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================
-- 9. Triggers for Automatic Current State Updates
-- ============================================================

-- Trigger to update inventory current state when ledger entry occurs
CREATE OR REPLACE FUNCTION update_inventory_current_state()
RETURNS TRIGGER AS $$
BEGIN
    -- Only process inventory ledger entries
    IF NEW.ledger_type = 'INVENTORY' AND NEW.product_id IS NOT NULL THEN
        -- Update the inventory item's current state
        UPDATE inventory
        SET
            current_quantity = NEW.balance,
            current_cost = NEW.cost_after,
            current_value = NEW.balance * NEW.cost_after,
            last_movement_id = NEW.id,
            updated_at = NOW()
        WHERE product_id = NEW.product_id;

        -- Update available quantity (current - reserved)
        UPDATE inventory
        SET available_quantity = current_quantity - reserved_quantity
        WHERE product_id = NEW.product_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for ledger entries
DROP TRIGGER IF EXISTS trigger_update_inventory_state ON ledger_entries;
CREATE TRIGGER trigger_update_inventory_state
    AFTER INSERT ON ledger_entries
    FOR EACH ROW
    EXECUTE FUNCTION update_inventory_current_state();

-- ============================================================
-- 10. Functions for Aggregation Updates
-- ============================================================

-- Function to update daily sales summary
CREATE OR REPLACE FUNCTION update_daily_sales_summary(target_date DATE DEFAULT CURRENT_DATE)
RETURNS VOID AS $$
BEGIN
    INSERT INTO daily_sales_summary (date, total_sales, total_revenue, total_profit, total_customers, average_order_value, total_items_sold, cash_sales, card_sales, debt_sales, updated_at)
    SELECT
        target_date,
        COUNT(*) as total_sales,
        COALESCE(SUM(total_amount), 0) as total_revenue,
        COALESCE(SUM(profit), 0) as total_profit,
        COUNT(DISTINCT customer_id) as total_customers,
        COALESCE(AVG(total_amount), 0) as average_order_value,
        COALESCE(SUM(total_items), 0) as total_items_sold,
        COALESCE(SUM(CASE WHEN payment_method = 'cash' THEN total_amount ELSE 0 END), 0) as cash_sales,
        COALESCE(SUM(CASE WHEN payment_method = 'card' THEN total_amount ELSE 0 END), 0) as card_sales,
        COALESCE(SUM(CASE WHEN payment_method = 'debt' THEN total_amount ELSE 0 END), 0) as debt_sales,
        NOW()
    FROM sales
    WHERE DATE(created_at) = target_date
    ON CONFLICT (date) DO UPDATE SET
        total_sales = EXCLUDED.total_sales,
        total_revenue = EXCLUDED.total_revenue,
        total_profit = EXCLUDED.total_profit,
        total_customers = EXCLUDED.total_customers,
        average_order_value = EXCLUDED.average_order_value,
        total_items_sold = EXCLUDED.total_items_sold,
        cash_sales = EXCLUDED.cash_sales,
        card_sales = EXCLUDED.card_sales,
        debt_sales = EXCLUDED.debt_sales,
        updated_at = EXCLUDED.updated_at;
END;
$$ LANGUAGE plpgsql;

-- Function to update monthly sales summary
CREATE OR REPLACE FUNCTION update_monthly_sales_summary(target_year INTEGER DEFAULT EXTRACT(YEAR FROM CURRENT_DATE), target_month INTEGER DEFAULT EXTRACT(MONTH FROM CURRENT_DATE))
RETURNS VOID AS $$
BEGIN
    INSERT INTO monthly_sales_summary (year, month, total_sales, total_revenue, total_profit, total_customers, average_order_value, total_items_sold, cash_sales, card_sales, debt_sales, updated_at)
    SELECT
        target_year,
        target_month,
        COUNT(*) as total_sales,
        COALESCE(SUM(total_amount), 0) as total_revenue,
        COALESCE(SUM(profit), 0) as total_profit,
        COUNT(DISTINCT customer_id) as total_customers,
        COALESCE(AVG(total_amount), 0) as average_order_value,
        COALESCE(SUM(total_items), 0) as total_items_sold,
        COALESCE(SUM(CASE WHEN payment_method = 'cash' THEN total_amount ELSE 0 END), 0) as cash_sales,
        COALESCE(SUM(CASE WHEN payment_method = 'card' THEN total_amount ELSE 0 END), 0) as card_sales,
        COALESCE(SUM(CASE WHEN payment_method = 'debt' THEN total_amount ELSE 0 END), 0) as debt_sales,
        NOW()
    FROM sales
    WHERE EXTRACT(YEAR FROM created_at) = target_year
    AND EXTRACT(MONTH FROM created_at) = target_month
    ON CONFLICT (year, month) DO UPDATE SET
        total_sales = EXCLUDED.total_sales,
        total_revenue = EXCLUDED.total_revenue,
        total_profit = EXCLUDED.total_profit,
        total_customers = EXCLUDED.total_customers,
        average_order_value = EXCLUDED.average_order_value,
        total_items_sold = EXCLUDED.total_items_sold,
        cash_sales = EXCLUDED.cash_sales,
        card_sales = EXCLUDED.card_sales,
        debt_sales = EXCLUDED.debt_sales,
        updated_at = EXCLUDED.updated_at;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Comments
-- ============================================================

COMMENT ON TABLE daily_sales_summary IS 'Daily sales aggregation for Dashboard performance (ARCHITECTURE-PRINCIPLES.md)';
COMMENT ON TABLE monthly_sales_summary IS 'Monthly sales aggregation for Dashboard performance (ARCHITECTURE-PRINCIPLES.md)';
COMMENT ON TABLE daily_inventory_summary IS 'Daily inventory aggregation for Dashboard performance (ARCHITECTURE-PRINCIPLES.md)';
COMMENT ON TABLE monthly_inventory_summary IS 'Monthly inventory aggregation for Dashboard performance (ARCHITECTURE-PRINCIPLES.md)';
COMMENT ON TABLE daily_debt_summary IS 'Daily debt aggregation for Dashboard performance (ARCHITECTURE-PRINCIPLES.md)';
COMMENT ON TABLE monthly_debt_summary IS 'Monthly debt aggregation for Dashboard performance (ARCHITECTURE-PRINCIPLES.md)';
COMMENT ON TABLE daily_profit_summary IS 'Daily profit aggregation for Dashboard performance (ARCHITECTURE-PRINCIPLES.md)';
COMMENT ON TABLE monthly_profit_summary IS 'Monthly profit aggregation for Dashboard performance (ARCHITECTURE-PRINCIPLES.md)';
COMMENT ON TABLE purchase_reversals IS 'Purchase reversal records (ARCHITECTURE-PRINCIPLES.md - Reverse instead of Delete)';
COMMENT ON TABLE payment_reversals IS 'Payment reversal records (ARCHITECTURE-PRINCIPLES.md - Reverse instead of Delete)';
COMMENT ON TABLE archive_status IS 'Archive status tracking for future implementation (ARCHITECTURE-PRINCIPLES.md)';
