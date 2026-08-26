-- Fix reversal fields and create aggregation tables

-- Add reversal fields to payments
ALTER TABLE payments 
ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS reversed_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS reversal_reason TEXT;

-- Add reversal fields to sales
ALTER TABLE sales 
ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS reversed_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS reversal_reason TEXT;

-- Add reversal fields to purchases
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS reversed_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS reversal_reason TEXT;

-- Create aggregation tables for sales
CREATE TABLE IF NOT EXISTS daily_sales_summary (
    summary_date DATE PRIMARY KEY,
    total_sales NUMERIC(15,2) DEFAULT 0,
    total_orders INTEGER DEFAULT 0,
    total_customers INTEGER DEFAULT 0,
    avg_order_value NUMERIC(15,2) DEFAULT 0,
    total_paid NUMERIC(15,2) DEFAULT 0,
    total_credit NUMERIC(15,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS monthly_sales_summary (
    summary_month DATE PRIMARY KEY,
    total_sales NUMERIC(15,2) DEFAULT 0,
    total_orders INTEGER DEFAULT 0,
    total_customers INTEGER DEFAULT 0,
    avg_order_value NUMERIC(15,2) DEFAULT 0,
    total_paid NUMERIC(15,2) DEFAULT 0,
    total_credit NUMERIC(15,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create aggregation tables for inventory
CREATE TABLE IF NOT EXISTS daily_inventory_summary (
    summary_date DATE PRIMARY KEY,
    total_products INTEGER DEFAULT 0,
    total_stock INTEGER DEFAULT 0,
    total_value NUMERIC(15,2) DEFAULT 0,
    low_stock_items INTEGER DEFAULT 0,
    out_of_stock_items INTEGER DEFAULT 0,
    new_items INTEGER DEFAULT 0,
    sold_items INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS monthly_inventory_summary (
    summary_month DATE PRIMARY KEY,
    total_products INTEGER DEFAULT 0,
    avg_stock_level INTEGER DEFAULT 0,
    total_value NUMERIC(15,2) DEFAULT 0,
    low_stock_items INTEGER DEFAULT 0,
    new_items INTEGER DEFAULT 0,
    sold_items INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create aggregation tables for debts
CREATE TABLE IF NOT EXISTS daily_debt_summary (
    summary_date DATE PRIMARY KEY,
    total_debts INTEGER DEFAULT 0,
    total_debt_amount NUMERIC(15,2) DEFAULT 0,
    total_remaining NUMERIC(15,2) DEFAULT 0,
    overdue_debts INTEGER DEFAULT 0,
    overdue_amount NUMERIC(15,2) DEFAULT 0,
    paid_today NUMERIC(15,2) DEFAULT 0,
    new_debts_today NUMERIC(15,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS monthly_debt_summary (
    summary_month DATE PRIMARY KEY,
    total_debts INTEGER DEFAULT 0,
    total_debt_amount NUMERIC(15,2) DEFAULT 0,
    total_collected NUMERIC(15,2) DEFAULT 0,
    overdue_debts INTEGER DEFAULT 0,
    overdue_amount NUMERIC(15,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create aggregation tables for profits
CREATE TABLE IF NOT EXISTS daily_profit_summary (
    summary_date DATE PRIMARY KEY,
    total_revenue NUMERIC(15,2) DEFAULT 0,
    total_cost NUMERIC(15,2) DEFAULT 0,
    gross_profit NUMERIC(15,2) DEFAULT 0,
    net_profit NUMERIC(15,2) DEFAULT 0,
    profit_margin NUMERIC(5,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS monthly_profit_summary (
    summary_month DATE PRIMARY KEY,
    total_revenue NUMERIC(15,2) DEFAULT 0,
    total_cost NUMERIC(15,2) DEFAULT 0,
    gross_profit NUMERIC(15,2) DEFAULT 0,
    net_profit NUMERIC(15,2) DEFAULT 0,
    profit_margin NUMERIC(5,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create function to update daily sales summary
CREATE OR REPLACE FUNCTION update_daily_sales_summary()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO daily_sales_summary (summary_date, total_sales, total_orders, total_customers, avg_order_value, total_paid, total_credit)
    VALUES (
        CURRENT_DATE,
        NEW.total_amount,
        1,
        1,
        NEW.total_amount,
        CASE WHEN NEW.payment_method = 'cash' THEN NEW.paid_amount ELSE 0 END,
        CASE WHEN NEW.payment_method = 'credit' THEN NEW.total_amount ELSE 0 END
    )
    ON CONFLICT (summary_date) DO UPDATE SET
        total_sales = daily_sales_summary.total_sales + NEW.total_amount,
        total_orders = daily_sales_summary.total_orders + 1,
        total_customers = daily_sales_summary.total_customers + 1,
        avg_order_value = (daily_sales_summary.total_sales + NEW.total_amount) / (daily_sales_summary.total_orders + 1),
        total_paid = daily_sales_summary.total_paid + CASE WHEN NEW.payment_method = 'cash' THEN NEW.paid_amount ELSE 0 END,
        total_credit = daily_sales_summary.total_credit + CASE WHEN NEW.payment_method = 'credit' THEN NEW.total_amount ELSE 0 END,
        updated_at = NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for sales
DROP TRIGGER IF EXISTS trigger_update_daily_sales_summary ON sales;
CREATE TRIGGER trigger_update_daily_sales_summary
AFTER INSERT ON sales
FOR EACH ROW
EXECUTE FUNCTION update_daily_sales_summary();

-- Create function to update daily debt summary
CREATE OR REPLACE FUNCTION update_daily_debt_summary()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO daily_debt_summary (summary_date, total_debts, total_debt_amount, total_remaining, overdue_debts, overdue_amount, new_debts_today)
    VALUES (
        CURRENT_DATE,
        1,
        NEW.amount,
        NEW.remaining_amount,
        CASE WHEN NEW.due_date < CURRENT_DATE THEN 1 ELSE 0 END,
        CASE WHEN NEW.due_date < CURRENT_DATE THEN NEW.remaining_amount ELSE 0 END,
        NEW.amount
    )
    ON CONFLICT (summary_date) DO UPDATE SET
        total_debts = daily_debt_summary.total_debts + 1,
        total_debt_amount = daily_debt_summary.total_debt_amount + NEW.amount,
        total_remaining = daily_debt_summary.total_remaining + NEW.remaining_amount,
        overdue_debts = daily_debt_summary.overdue_debts + CASE WHEN NEW.due_date < CURRENT_DATE THEN 1 ELSE 0 END,
        overdue_amount = daily_debt_summary.overdue_amount + CASE WHEN NEW.due_date < CURRENT_DATE THEN NEW.remaining_amount ELSE 0 END,
        new_debts_today = daily_debt_summary.new_debts_today + NEW.amount,
        updated_at = NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for debts
DROP TRIGGER IF EXISTS trigger_update_daily_debt_summary ON debts;
CREATE TRIGGER trigger_update_daily_debt_summary
AFTER INSERT ON debts
FOR EACH ROW
EXECUTE FUNCTION update_daily_debt_summary();

-- Create indexes for aggregation tables
CREATE INDEX IF NOT EXISTS idx_daily_sales_summary_date ON daily_sales_summary(summary_date);
CREATE INDEX IF NOT EXISTS idx_monthly_sales_summary_month ON monthly_sales_summary(summary_month);
CREATE INDEX IF NOT EXISTS idx_daily_inventory_summary_date ON daily_inventory_summary(summary_date);
CREATE INDEX IF NOT EXISTS idx_monthly_inventory_summary_month ON monthly_inventory_summary(summary_month);
CREATE INDEX IF NOT EXISTS idx_daily_debt_summary_date ON daily_debt_summary(summary_date);
CREATE INDEX IF NOT EXISTS idx_monthly_debt_summary_month ON monthly_debt_summary(summary_month);
CREATE INDEX IF NOT EXISTS idx_daily_profit_summary_date ON daily_profit_summary(summary_date);
CREATE INDEX IF NOT EXISTS idx_monthly_profit_summary_month ON monthly_profit_summary(summary_month);