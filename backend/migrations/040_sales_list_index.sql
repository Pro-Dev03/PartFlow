BEGIN;

CREATE INDEX IF NOT EXISTS idx_sales_list_date_id
    ON sales (sale_date DESC, id DESC);

COMMIT;
