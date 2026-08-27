-- Indexes for the most common list, stock, and notification filters.
CREATE INDEX IF NOT EXISTS idx_purchases_status_date
    ON purchases (status, purchase_date DESC);

CREATE INDEX IF NOT EXISTS idx_sales_status_date
    ON sales (status, sale_date DESC);

CREATE INDEX IF NOT EXISTS idx_inventory_items_product_status_created
    ON inventory_items (product_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_inventory_items_available_product
    ON inventory_items (product_id)
    WHERE status = 'AVAILABLE';

CREATE INDEX IF NOT EXISTS idx_debts_customer_status_due_date
    ON debts (customer_id, status, due_date);

CREATE INDEX IF NOT EXISTS idx_notifications_user_read_date
    ON notifications (user_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_date
    ON audit_logs (user_id, created_at DESC);
