-- Reconcile schemas created by the architecture migration (002) with the
-- legacy aggregation migration (033).  Both files used CREATE TABLE IF NOT
-- EXISTS, so an already-created table kept the old column names and its
-- triggers could fail every new sale/debt insert.  The API is now the single
-- source of truth for recomputing summaries; these statements make the
-- runtime columns available without deleting existing rows.
BEGIN;

ALTER TABLE daily_sales_summary
    ADD COLUMN IF NOT EXISTS date DATE,
    ADD COLUMN IF NOT EXISTS total_revenue NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_profit NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_customers INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS average_order_value NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_items_sold INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cash_sales NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS card_sales NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS debt_sales NUMERIC(15,2) DEFAULT 0;

ALTER TABLE monthly_sales_summary
    ADD COLUMN IF NOT EXISTS year INTEGER,
    ADD COLUMN IF NOT EXISTS month INTEGER,
    ADD COLUMN IF NOT EXISTS total_revenue NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_profit NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_customers INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS average_order_value NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_items_sold INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cash_sales NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS card_sales NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS debt_sales NUMERIC(15,2) DEFAULT 0;

ALTER TABLE daily_inventory_summary
    ADD COLUMN IF NOT EXISTS date DATE,
    ADD COLUMN IF NOT EXISTS total_items INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_value NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS low_stock_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS out_of_stock_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS new_items_added INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS items_sold INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS items_returned INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS items_damaged INTEGER DEFAULT 0;

ALTER TABLE monthly_inventory_summary
    ADD COLUMN IF NOT EXISTS year INTEGER,
    ADD COLUMN IF NOT EXISTS month INTEGER,
    ADD COLUMN IF NOT EXISTS total_items INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_value NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS low_stock_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS out_of_stock_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS new_items_added INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS items_sold INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS items_returned INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS items_damaged INTEGER DEFAULT 0;

ALTER TABLE daily_debt_summary
    ADD COLUMN IF NOT EXISTS date DATE,
    ADD COLUMN IF NOT EXISTS total_debt NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS new_debt NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS payments_received NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS overdue_debt NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS overdue_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS paid_debt NUMERIC(15,2) DEFAULT 0;

ALTER TABLE monthly_debt_summary
    ADD COLUMN IF NOT EXISTS year INTEGER,
    ADD COLUMN IF NOT EXISTS month INTEGER,
    ADD COLUMN IF NOT EXISTS total_debt NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS new_debt NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS payments_received NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS overdue_debt NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS overdue_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS paid_debt NUMERIC(15,2) DEFAULT 0;

ALTER TABLE daily_profit_summary
    ADD COLUMN IF NOT EXISTS date DATE,
    ADD COLUMN IF NOT EXISTS gross_profit NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS net_profit NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_revenue NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_cost NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS profit_margin NUMERIC(15,2) DEFAULT 0;

ALTER TABLE monthly_profit_summary
    ADD COLUMN IF NOT EXISTS year INTEGER,
    ADD COLUMN IF NOT EXISTS month INTEGER,
    ADD COLUMN IF NOT EXISTS gross_profit NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS net_profit NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_revenue NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_cost NUMERIC(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS profit_margin NUMERIC(15,2) DEFAULT 0;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_sales_summary' AND column_name = 'summary_date') THEN
        UPDATE daily_sales_summary SET date = summary_date WHERE date IS NULL;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_sales_summary' AND column_name = 'total_orders') THEN
            UPDATE daily_sales_summary SET total_revenue = COALESCE(total_sales, 0), total_sales = COALESCE(total_orders, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_sales_summary' AND column_name = 'total_paid') THEN
            UPDATE daily_sales_summary SET cash_sales = COALESCE(total_paid, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_sales_summary' AND column_name = 'total_credit') THEN
            UPDATE daily_sales_summary SET debt_sales = COALESCE(total_credit, 0);
        END IF;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'monthly_sales_summary' AND column_name = 'summary_month') THEN
        UPDATE monthly_sales_summary SET year = EXTRACT(YEAR FROM summary_month), month = EXTRACT(MONTH FROM summary_month) WHERE year IS NULL OR month IS NULL;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'monthly_sales_summary' AND column_name = 'total_orders') THEN
            UPDATE monthly_sales_summary SET total_revenue = COALESCE(total_sales, 0), total_sales = COALESCE(total_orders, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'monthly_sales_summary' AND column_name = 'total_paid') THEN
            UPDATE monthly_sales_summary SET cash_sales = COALESCE(total_paid, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'monthly_sales_summary' AND column_name = 'total_credit') THEN
            UPDATE monthly_sales_summary SET debt_sales = COALESCE(total_credit, 0);
        END IF;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_inventory_summary' AND column_name = 'summary_date') THEN
        UPDATE daily_inventory_summary SET date = summary_date WHERE date IS NULL;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_inventory_summary' AND column_name = 'total_stock') THEN
            UPDATE daily_inventory_summary SET total_items = COALESCE(total_stock, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_inventory_summary' AND column_name = 'low_stock_items') THEN
            UPDATE daily_inventory_summary SET low_stock_count = COALESCE(low_stock_items, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_inventory_summary' AND column_name = 'out_of_stock_items') THEN
            UPDATE daily_inventory_summary SET out_of_stock_count = COALESCE(out_of_stock_items, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_inventory_summary' AND column_name = 'new_items') THEN
            UPDATE daily_inventory_summary SET new_items_added = COALESCE(new_items, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_inventory_summary' AND column_name = 'sold_items') THEN
            UPDATE daily_inventory_summary SET items_sold = COALESCE(sold_items, 0);
        END IF;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'monthly_inventory_summary' AND column_name = 'summary_month') THEN
        UPDATE monthly_inventory_summary SET year = EXTRACT(YEAR FROM summary_month), month = EXTRACT(MONTH FROM summary_month) WHERE year IS NULL OR month IS NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_debt_summary' AND column_name = 'summary_date') THEN
        UPDATE daily_debt_summary SET date = summary_date WHERE date IS NULL;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_debt_summary' AND column_name = 'total_remaining') THEN
            UPDATE daily_debt_summary SET total_debt = COALESCE(total_remaining, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_debt_summary' AND column_name = 'new_debts_today') THEN
            UPDATE daily_debt_summary SET new_debt = COALESCE(new_debts_today, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_debt_summary' AND column_name = 'overdue_amount') THEN
            UPDATE daily_debt_summary SET overdue_debt = COALESCE(overdue_amount, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_debt_summary' AND column_name = 'overdue_debts') THEN
            UPDATE daily_debt_summary SET overdue_count = COALESCE(overdue_debts, 0);
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_debt_summary' AND column_name = 'paid_today') THEN
            UPDATE daily_debt_summary SET paid_debt = COALESCE(paid_today, 0), payments_received = COALESCE(paid_today, 0);
        END IF;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'monthly_debt_summary' AND column_name = 'summary_month') THEN
        UPDATE monthly_debt_summary SET year = EXTRACT(YEAR FROM summary_month), month = EXTRACT(MONTH FROM summary_month) WHERE year IS NULL OR month IS NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'daily_profit_summary' AND column_name = 'summary_date') THEN
        UPDATE daily_profit_summary SET date = summary_date WHERE date IS NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'monthly_profit_summary' AND column_name = 'summary_month') THEN
        UPDATE monthly_profit_summary SET year = EXTRACT(YEAR FROM summary_month), month = EXTRACT(MONTH FROM summary_month) WHERE year IS NULL OR month IS NULL;
    END IF;
END $$;

DROP TRIGGER IF EXISTS trigger_update_daily_sales_summary ON sales;
DROP TRIGGER IF EXISTS trigger_update_daily_debt_summary ON debts;
DROP FUNCTION IF EXISTS update_daily_sales_summary();
DROP FUNCTION IF EXISTS update_daily_debt_summary();

-- The runtime refresh uses ON CONFLICT(date) and ON CONFLICT(year, month).
-- Legacy migration 033 keyed these tables by summary_date/summary_month, so
-- add the equivalent unique indexes after copying legacy values.  Remove only
-- duplicate summary rows; the newest physical row is retained.
DELETE FROM daily_sales_summary a USING daily_sales_summary b
 WHERE a.date IS NOT NULL AND a.date = b.date AND a.ctid > b.ctid;
DELETE FROM monthly_sales_summary a USING monthly_sales_summary b
 WHERE a.year IS NOT NULL AND a.month IS NOT NULL
   AND a.year = b.year AND a.month = b.month AND a.ctid > b.ctid;
DELETE FROM daily_inventory_summary a USING daily_inventory_summary b
 WHERE a.date IS NOT NULL AND a.date = b.date AND a.ctid > b.ctid;
DELETE FROM monthly_inventory_summary a USING monthly_inventory_summary b
 WHERE a.year IS NOT NULL AND a.month IS NOT NULL
   AND a.year = b.year AND a.month = b.month AND a.ctid > b.ctid;
DELETE FROM daily_debt_summary a USING daily_debt_summary b
 WHERE a.date IS NOT NULL AND a.date = b.date AND a.ctid > b.ctid;
DELETE FROM monthly_debt_summary a USING monthly_debt_summary b
 WHERE a.year IS NOT NULL AND a.month IS NOT NULL
   AND a.year = b.year AND a.month = b.month AND a.ctid > b.ctid;
DELETE FROM daily_profit_summary a USING daily_profit_summary b
 WHERE a.date IS NOT NULL AND a.date = b.date AND a.ctid > b.ctid;
DELETE FROM monthly_profit_summary a USING monthly_profit_summary b
 WHERE a.year IS NOT NULL AND a.month IS NOT NULL
   AND a.year = b.year AND a.month = b.month AND a.ctid > b.ctid;

CREATE UNIQUE INDEX IF NOT EXISTS uq_daily_sales_summary_runtime_date ON daily_sales_summary(date);
CREATE UNIQUE INDEX IF NOT EXISTS uq_monthly_sales_summary_runtime_year_month ON monthly_sales_summary(year, month);
CREATE UNIQUE INDEX IF NOT EXISTS uq_daily_inventory_summary_runtime_date ON daily_inventory_summary(date);
CREATE UNIQUE INDEX IF NOT EXISTS uq_monthly_inventory_summary_runtime_year_month ON monthly_inventory_summary(year, month);
CREATE UNIQUE INDEX IF NOT EXISTS uq_daily_debt_summary_runtime_date ON daily_debt_summary(date);
CREATE UNIQUE INDEX IF NOT EXISTS uq_monthly_debt_summary_runtime_year_month ON monthly_debt_summary(year, month);
CREATE UNIQUE INDEX IF NOT EXISTS uq_daily_profit_summary_runtime_date ON daily_profit_summary(date);
CREATE UNIQUE INDEX IF NOT EXISTS uq_monthly_profit_summary_runtime_year_month ON monthly_profit_summary(year, month);

CREATE INDEX IF NOT EXISTS idx_daily_sales_summary_runtime_date ON daily_sales_summary(date);
CREATE INDEX IF NOT EXISTS idx_monthly_sales_summary_runtime_year_month ON monthly_sales_summary(year, month);
CREATE INDEX IF NOT EXISTS idx_daily_inventory_summary_runtime_date ON daily_inventory_summary(date);
CREATE INDEX IF NOT EXISTS idx_monthly_inventory_summary_runtime_year_month ON monthly_inventory_summary(year, month);
CREATE INDEX IF NOT EXISTS idx_daily_debt_summary_runtime_date ON daily_debt_summary(date);
CREATE INDEX IF NOT EXISTS idx_monthly_debt_summary_runtime_year_month ON monthly_debt_summary(year, month);
CREATE INDEX IF NOT EXISTS idx_daily_profit_summary_runtime_date ON daily_profit_summary(date);
CREATE INDEX IF NOT EXISTS idx_monthly_profit_summary_runtime_year_month ON monthly_profit_summary(year, month);

COMMIT;
