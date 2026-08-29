-- Data retention structures. Archival is explicit and never runs automatically.
CREATE TABLE IF NOT EXISTS archived_sales (
    sale_id UUID PRIMARY KEY,
    archived_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sale_date TIMESTAMPTZ,
    sale_data JSONB NOT NULL,
    items_data JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS archived_purchases (
    purchase_id UUID PRIMARY KEY,
    archived_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    purchase_date TIMESTAMPTZ,
    purchase_data JSONB NOT NULL,
    items_data JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS annual_financial_summaries (
    year SMALLINT PRIMARY KEY,
    sales_count BIGINT NOT NULL DEFAULT 0,
    sales_revenue NUMERIC(14, 2) NOT NULL DEFAULT 0,
    purchases_count BIGINT NOT NULL DEFAULT 0,
    purchases_amount NUMERIC(14, 2) NOT NULL DEFAULT 0,
    expenses_amount NUMERIC(14, 2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_archived_sales_date ON archived_sales (sale_date);
CREATE INDEX IF NOT EXISTS idx_archived_purchases_date ON archived_purchases (purchase_date);

CREATE OR REPLACE FUNCTION archive_closed_records_before(cutoff_date DATE)
RETURNS TABLE (archived_sales_count BIGINT, archived_purchases_count BIGINT)
LANGUAGE plpgsql
AS $$
DECLARE
    sales_count BIGINT := 0;
    purchases_count BIGINT := 0;
BEGIN
    INSERT INTO annual_financial_summaries (year, sales_count, sales_revenue, purchases_count, purchases_amount, updated_at)
    SELECT EXTRACT(YEAR FROM d.year_start)::SMALLINT,
           COALESCE(s.sales_count, 0), COALESCE(s.sales_revenue, 0),
           COALESCE(p.purchases_count, 0), COALESCE(p.purchases_amount, 0), NOW()
    FROM generate_series(
        (SELECT MIN(DATE_TRUNC('year', sale_date)) FROM sales WHERE sale_date < cutoff_date),
        (SELECT MAX(DATE_TRUNC('year', purchase_date)) FROM purchases WHERE purchase_date < cutoff_date),
        INTERVAL '1 year'
    ) AS d(year_start)
    LEFT JOIN (
        SELECT DATE_TRUNC('year', sale_date) AS year_start, COUNT(*) AS sales_count,
               SUM(total_amount) AS sales_revenue
        FROM sales
        WHERE sale_date < cutoff_date AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
        GROUP BY 1
    ) s ON s.year_start = d.year_start
    LEFT JOIN (
        SELECT DATE_TRUNC('year', purchase_date) AS year_start, COUNT(*) AS purchases_count,
               SUM(total_amount) AS purchases_amount
        FROM purchases
        WHERE purchase_date < cutoff_date AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
        GROUP BY 1
    ) p ON p.year_start = d.year_start
    ON CONFLICT (year) DO UPDATE SET
        sales_count = EXCLUDED.sales_count,
        sales_revenue = EXCLUDED.sales_revenue,
        purchases_count = EXCLUDED.purchases_count,
        purchases_amount = EXCLUDED.purchases_amount,
        updated_at = NOW();

    WITH candidates AS (
        SELECT s.id, s.sale_date, to_jsonb(s) AS sale_data,
               COALESCE((SELECT jsonb_agg(to_jsonb(si)) FROM sale_items si WHERE si.sale_id = s.id), '[]'::jsonb) AS items_data
        FROM sales s
        WHERE s.sale_date < cutoff_date AND LOWER(COALESCE(s.status, 'completed')) IN ('cancelled', 'canceled', 'reversed')
    )
    INSERT INTO archived_sales (sale_id, sale_date, sale_data, items_data)
    SELECT id, sale_date, sale_data, items_data FROM candidates
    ON CONFLICT (sale_id) DO NOTHING;
    GET DIAGNOSTICS sales_count = ROW_COUNT;

    DELETE FROM sale_items WHERE sale_id IN (SELECT sale_id FROM archived_sales WHERE sale_date < cutoff_date);
    DELETE FROM sales WHERE id IN (SELECT sale_id FROM archived_sales WHERE sale_date < cutoff_date);

    WITH candidates AS (
        SELECT p.id, p.purchase_date, to_jsonb(p) AS purchase_data,
               COALESCE((SELECT jsonb_agg(to_jsonb(pi)) FROM purchase_items pi WHERE pi.purchase_id = p.id), '[]'::jsonb) AS items_data
        FROM purchases p
        WHERE p.purchase_date < cutoff_date AND LOWER(COALESCE(p.status, 'completed')) IN ('cancelled', 'canceled', 'reversed')
    )
    INSERT INTO archived_purchases (purchase_id, purchase_date, purchase_data, items_data)
    SELECT id, purchase_date, purchase_data, items_data FROM candidates
    ON CONFLICT (purchase_id) DO NOTHING;
    GET DIAGNOSTICS purchases_count = ROW_COUNT;

    DELETE FROM purchase_items WHERE purchase_id IN (SELECT purchase_id FROM archived_purchases WHERE purchase_date < cutoff_date);
    DELETE FROM purchases WHERE id IN (SELECT purchase_id FROM archived_purchases WHERE purchase_date < cutoff_date);

    RETURN QUERY SELECT sales_count, purchases_count;
END;
$$;
