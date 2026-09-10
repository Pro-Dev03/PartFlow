package localdb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type Database struct {
	DB   *sql.DB
	Path string
}

func Open() (*Database, error) {
	path, err := databasePath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create local database directory: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, fmt.Errorf("open local database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping local database: %w", err)
	}

	if err := initializeSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := ensureDefaultOwnerUser(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Database{DB: db, Path: path}, nil
}

func ensureDefaultOwnerUser(db *sql.DB) error {
	var userCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		return fmt.Errorf("count local users: %w", err)
	}
	if userCount > 0 {
		return nil
	}

	bootstrapPassword := strings.TrimSpace(os.Getenv("PARTFLOW_BOOTSTRAP_OWNER_PASSWORD"))
	if bootstrapPassword == "" {
		// Production/Desktop builds must receive users from the authenticated cloud
		// snapshot; never create a predictable local account automatically.
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(bootstrapPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash default owner password: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	expires := time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)
	uid := uuid.NewString()

	_, err = db.Exec(`
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, phone, role,
			is_active, is_verified, subscription_status, subscription_expires_at,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 1, 1, 'active', ?, ?, ?)
	`, uid, "owner@partflow.com", string(hash), "Owner", "Admin", "+970599000000", "owner", expires, now, now)
	if err != nil {
		return fmt.Errorf("create default owner user: %w", err)
	}

	return nil
}

func databasePath() (string, error) {
	if path := os.Getenv("PARTFLOW_LOCAL_DB_PATH"); path != "" {
		return path, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve local database directory: %w", err)
	}

	return filepath.Join(configDir, "PartFlow", "data", "partflow.db"), nil
}

func initializeSchema(db *sql.DB) error {
	const schema = `
-- Metadata tables for sync and settings
CREATE TABLE IF NOT EXISTS local_metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sync_queue (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    payload TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    next_retry_at TEXT,
    synced_at TEXT
);

CREATE TABLE IF NOT EXISTS sync_conflicts (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    entity_table TEXT NOT NULL,
    operation TEXT NOT NULL,
    conflict_type TEXT NOT NULL DEFAULT 'generic',
    local_updated_at TEXT,
    remote_updated_at TEXT,
    conflict_reason TEXT NOT NULL,
    payload TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS local_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    email TEXT NOT NULL,
    display_name TEXT,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    phone TEXT,
    avatar_url TEXT,
    role TEXT DEFAULT 'owner',
    is_active INTEGER NOT NULL DEFAULT 1,
    is_verified INTEGER NOT NULL DEFAULT 0,
    last_login_at TEXT,
    subscription_status TEXT DEFAULT 'active',
    subscription_expires_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    used INTEGER NOT NULL DEFAULT 0,
    used_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_sync_queue_pending
    ON sync_queue (synced_at, next_retry_at, created_at);

CREATE INDEX IF NOT EXISTS idx_sync_conflicts_entity
    ON sync_conflicts (entity_type, entity_id, created_at);

CREATE INDEX IF NOT EXISTS idx_local_sessions_user
    ON local_sessions (user_id, email);

-- Business entity tables
CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    parent_id TEXT,
    icon TEXT,
    color TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS brands (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    logo_url TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    category_id TEXT,
    brand_id TEXT,
    preferred_supplier_id TEXT,
    model TEXT,
    barcode TEXT,
    purchase_price REAL DEFAULT 0,
    cost_price REAL DEFAULT 0,
    selling_price REAL DEFAULT 0,
    currency TEXT DEFAULT 'ILS',
    min_stock_level INTEGER DEFAULT 0,
    max_stock_level INTEGER DEFAULT 0,
    warranty_days INTEGER DEFAULT 0,
    track_serial INTEGER NOT NULL DEFAULT 0,
    track_individual INTEGER NOT NULL DEFAULT 0,
    is_active INTEGER NOT NULL DEFAULT 1,
    deleted_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id),
    FOREIGN KEY (brand_id) REFERENCES brands(id),
    FOREIGN KEY (preferred_supplier_id) REFERENCES suppliers(id)
);

CREATE TABLE IF NOT EXISTS customers (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    address TEXT,
    city TEXT,
    country TEXT,
    tax_id TEXT,
    credit_limit REAL DEFAULT 0,
    current_balance REAL DEFAULT 0,
    notes TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS suppliers (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    address TEXT,
    city TEXT,
    country TEXT,
    credit_limit REAL DEFAULT 0,
    current_balance REAL DEFAULT 0,
    notes TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS inventory_items (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    item_code TEXT NOT NULL UNIQUE,
    barcode TEXT UNIQUE,
    serial_number TEXT UNIQUE,
    condition TEXT,
    purchase_cost REAL DEFAULT 0,
    selling_price REAL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'AVAILABLE',
    supplier_id TEXT,
    purchase_date TEXT,
    sold_at TEXT,
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
);

CREATE TABLE IF NOT EXISTS sales (
    id TEXT PRIMARY KEY,
    sale_number TEXT NOT NULL UNIQUE,
    customer_id TEXT,
    total_amount REAL NOT NULL,
    tax_amount REAL DEFAULT 0,
    discount_amount REAL DEFAULT 0,
    paid_amount REAL DEFAULT 0,
    remaining_amount REAL DEFAULT 0,
    payment_method TEXT,
    status TEXT NOT NULL DEFAULT 'completed',
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);

CREATE TABLE IF NOT EXISTS sale_items (
    id TEXT PRIMARY KEY,
    sale_id TEXT NOT NULL,
    inventory_item_id TEXT,
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price REAL NOT NULL,
    item_total REAL NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (sale_id) REFERENCES sales(id),
    FOREIGN KEY (inventory_item_id) REFERENCES inventory_items(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE TABLE IF NOT EXISTS purchases (
    id TEXT PRIMARY KEY,
    purchase_number TEXT NOT NULL UNIQUE,
	supplier_id TEXT,
    total_amount REAL NOT NULL,
    tax_amount REAL DEFAULT 0,
    paid_amount REAL DEFAULT 0,
    remaining_amount REAL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'completed',
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
);

CREATE TABLE IF NOT EXISTS purchase_items (
    id TEXT PRIMARY KEY,
    purchase_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price REAL NOT NULL,
    item_total REAL NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (purchase_id) REFERENCES purchases(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    transaction_number TEXT NOT NULL UNIQUE,
    customer_id TEXT,
    supplier_id TEXT,
    amount REAL NOT NULL,
    payment_method TEXT,
    reference TEXT,
    notes TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
);

CREATE TABLE IF NOT EXISTS debts (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    sale_id TEXT,
    amount REAL NOT NULL,
    paid_amount REAL DEFAULT 0,
    remaining_amount REAL DEFAULT 0,
    due_date TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    FOREIGN KEY (sale_id) REFERENCES sales(id)
);

CREATE TABLE IF NOT EXISTS expense_categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT,
    icon TEXT,
    budget REAL DEFAULT 0,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS expenses (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    category_id TEXT,
    amount REAL NOT NULL DEFAULT 0,
    currency TEXT DEFAULT 'ILS',
    reference TEXT,
    notes TEXT,
    expense_date TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'approved',
    is_recurring INTEGER NOT NULL DEFAULT 0,
    recurring_period TEXT,
    approved_by TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (category_id) REFERENCES expense_categories(id)
);

CREATE TABLE IF NOT EXISTS returns (
    id TEXT PRIMARY KEY,
    return_number TEXT NOT NULL UNIQUE,
    customer_id TEXT,
    sale_id TEXT,
    purchase_id TEXT,
    total_refund_amount REAL NOT NULL DEFAULT 0,
    refund_status TEXT NOT NULL DEFAULT 'pending',
    status TEXT NOT NULL DEFAULT 'pending',
    reason TEXT,
    notes TEXT,
    return_date TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    FOREIGN KEY (sale_id) REFERENCES sales(id),
    FOREIGN KEY (purchase_id) REFERENCES purchases(id)
);

CREATE TABLE IF NOT EXISTS return_items (
    id TEXT PRIMARY KEY,
    return_id TEXT NOT NULL,
    product_id TEXT,
    quantity INTEGER NOT NULL DEFAULT 0,
    unit_price REAL NOT NULL DEFAULT 0,
    total_refund_amount REAL NOT NULL DEFAULT 0,
    reason TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (return_id) REFERENCES returns(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE TABLE IF NOT EXISTS warranty_claims (
    id TEXT PRIMARY KEY,
    claim_number TEXT NOT NULL UNIQUE,
    customer_id TEXT,
    product_id TEXT,
    serial_number TEXT,
    issue_description TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE TABLE IF NOT EXISTS supplier_returns (
    id TEXT PRIMARY KEY,
    purchase_id TEXT NOT NULL,
    supplier_id TEXT NOT NULL,
    return_number TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'PENDING',
    reason TEXT NOT NULL,
    refund_amount REAL NOT NULL DEFAULT 0,
    notes TEXT,
    created_by TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (purchase_id) REFERENCES purchases(id),
    FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
);

CREATE TABLE IF NOT EXISTS supplier_return_items (
    id TEXT PRIMARY KEY,
    supplier_return_id TEXT NOT NULL,
    purchase_item_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_cost REAL NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (supplier_return_id) REFERENCES supplier_returns(id),
    FOREIGN KEY (purchase_item_id) REFERENCES purchase_items(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE VIEW IF NOT EXISTS monthly_returns_analysis AS
SELECT
    date(r.return_date, 'start of month') AS month,
    COUNT(DISTINCT r.id) AS total_returns,
    COUNT(DISTINCT r.customer_id) AS unique_customers,
    COALESCE(SUM(r.total_refund_amount), 0) AS total_refund_amount,
    COALESCE(AVG(r.total_refund_amount), 0) AS avg_refund_amount,
    COUNT(CASE WHEN r.return_type = 'FULL' THEN 1 END) AS full_returns,
    COUNT(CASE WHEN r.return_type = 'PARTIAL' THEN 1 END) AS partial_returns,
    COUNT(CASE WHEN r.return_type = 'QUANTITY_PARTIAL' THEN 1 END) AS quantity_partial_returns,
    COUNT(CASE WHEN r.reason = 'DEFECTIVE' THEN 1 END) AS defective_returns,
    COUNT(CASE WHEN r.reason = 'WARRANTY' THEN 1 END) AS warranty_returns,
    COUNT(CASE WHEN r.is_warranty_claim = 1 THEN 1 END) AS warranty_claims,
    SUM(CASE WHEN r.item_condition_after_return = 'SELLABLE' THEN 1 ELSE 0 END) AS sellable_items,
    SUM(CASE WHEN r.item_condition_after_return = 'NEEDS_REPAIR' THEN 1 ELSE 0 END) AS repair_needed,
    SUM(CASE WHEN r.item_condition_after_return = 'WRITE_OFF' THEN 1 ELSE 0 END) AS written_off
FROM returns r
WHERE r.status = 'COMPLETED'
GROUP BY date(r.return_date, 'start of month')
ORDER BY month DESC;

CREATE VIEW IF NOT EXISTS sales_returns_analysis AS
SELECT
    date(s.sale_date, 'start of month') AS month,
    COUNT(DISTINCT s.id) AS total_sales,
    COALESCE(SUM(s.total_amount), 0) AS gross_sales,
    COALESCE(SUM(s.cost_amount), 0) AS total_cost,
    COALESCE(SUM(s.gross_profit), 0) AS gross_profit,
    COALESCE(SUM(r.total_refund_amount), 0) AS returns_amount,
    COALESCE(COUNT(r.id), 0) AS return_count,
    COALESCE(SUM(s.total_amount), 0) - COALESCE(SUM(r.total_refund_amount), 0) AS net_sales
FROM sales s
LEFT JOIN returns r ON r.sale_id = s.id
    AND r.status = 'COMPLETED'
    AND date(r.return_date, 'start of month') = date(s.sale_date, 'start of month')
WHERE s.status = 'completed'
GROUP BY date(s.sale_date, 'start of month')
ORDER BY month DESC;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_inventory_product ON inventory_items(product_id);
CREATE INDEX IF NOT EXISTS idx_inventory_barcode ON inventory_items(barcode);
CREATE INDEX IF NOT EXISTS idx_sales_customer ON sales(customer_id);
CREATE INDEX IF NOT EXISTS idx_sale_items_sale ON sale_items(sale_id);
CREATE INDEX IF NOT EXISTS idx_purchases_supplier ON purchases(supplier_id);
CREATE INDEX IF NOT EXISTS idx_payments_customer ON payments(customer_id);
CREATE INDEX IF NOT EXISTS idx_payments_supplier ON payments(supplier_id);
CREATE INDEX IF NOT EXISTS idx_debts_customer ON debts(customer_id);
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category_id);
CREATE INDEX IF NOT EXISTS idx_expenses_status ON expenses(status);
CREATE INDEX IF NOT EXISTS idx_returns_customer ON returns(customer_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_status ON warranty_claims(status);
`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("initialize local database schema: %w", err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS inventory (
		id TEXT PRIMARY KEY,
		product_id TEXT NOT NULL,
		quantity INTEGER NOT NULL DEFAULT 0,
		reserved_quantity INTEGER NOT NULL DEFAULT 0,
		location TEXT,
		warehouse_id TEXT,
		last_restocked_at TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (product_id) REFERENCES products(id)
	)`); err != nil {
		return fmt.Errorf("initialize inventory compatibility table: %w", err)
	}

	compatibilitySchema := []string{
		`CREATE TABLE IF NOT EXISTS settings (id TEXT PRIMARY KEY, key TEXT NOT NULL UNIQUE, value TEXT, value_type TEXT DEFAULT 'string', category TEXT DEFAULT 'general', description TEXT, is_public INTEGER DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS barcodes (id TEXT PRIMARY KEY, code TEXT NOT NULL UNIQUE, product_id TEXT, inventory_item_id TEXT, type TEXT NOT NULL, is_active INTEGER DEFAULT 1, generated_at TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		// Operational support tables used by the inventory and customer/supplier
		// ledger endpoints. They are kept in SQLite as well so a local build does
		// not lose history just because it cannot reach PostgreSQL.
		`CREATE TABLE IF NOT EXISTS locations (id TEXT PRIMARY KEY, name TEXT NOT NULL, type TEXT NOT NULL, parent_id TEXT, warehouse_id TEXT, description TEXT, is_active INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT NOT NULL, quantity INTEGER NOT NULL DEFAULT 0, before_quantity INTEGER NOT NULL DEFAULT 0, after_quantity INTEGER NOT NULL DEFAULT 0, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT NOT NULL, is_reversed INTEGER NOT NULL DEFAULT 0, reversed_by TEXT, reversed_at TEXT, reversal_reason TEXT)`,
		`CREATE TABLE IF NOT EXISTS reservations (id TEXT PRIMARY KEY, item_id TEXT NOT NULL, customer_id TEXT, user_id TEXT NOT NULL, reserved_at TEXT NOT NULL, expires_at TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'active', notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, type TEXT, transaction_type TEXT, amount REAL NOT NULL DEFAULT 0, balance REAL NOT NULL DEFAULT 0, description TEXT, reference_id TEXT, reference_type TEXT, created_by TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS customer_payments (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, amount REAL NOT NULL DEFAULT 0, payment_date TEXT NOT NULL, method TEXT NOT NULL, reference TEXT, notes TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS customer_debts (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, amount REAL NOT NULL DEFAULT 0, reference_id TEXT, reference_type TEXT, due_date TEXT, is_paid INTEGER NOT NULL DEFAULT 0, paid_amount REAL NOT NULL DEFAULT 0, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS debt_collections (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, type TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending', notes TEXT, scheduled_date TEXT NOT NULL, completed_date TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT NOT NULL, type TEXT, transaction_type TEXT, amount REAL NOT NULL DEFAULT 0, balance REAL NOT NULL DEFAULT 0, description TEXT, reference_id TEXT, reference_type TEXT, created_by TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS supplier_payments (id TEXT PRIMARY KEY, supplier_id TEXT NOT NULL, amount REAL NOT NULL DEFAULT 0, payment_date TEXT NOT NULL, method TEXT NOT NULL, reference TEXT, notes TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT NOT NULL, entity_type TEXT NOT NULL, entity_id TEXT NOT NULL, old_values TEXT, new_values TEXT, ip_address TEXT, user_agent TEXT, request_id TEXT, changes TEXT, description TEXT, status TEXT DEFAULT 'success', error_message TEXT, metadata TEXT DEFAULT '{}', created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS ledger_entries (id TEXT PRIMARY KEY, ledger_type TEXT NOT NULL, entity_id TEXT NOT NULL, transaction_type TEXT NOT NULL, reference_id TEXT, reference_type TEXT, amount REAL NOT NULL DEFAULT 0, balance REAL NOT NULL DEFAULT 0, previous_balance REAL DEFAULT 0, description TEXT, metadata TEXT DEFAULT '{}', created_by TEXT, created_at TEXT NOT NULL, cost_before REAL DEFAULT 0, cost_after REAL DEFAULT 0, value_before REAL DEFAULT 0, value_after REAL DEFAULT 0, is_reversed INTEGER NOT NULL DEFAULT 0, reversed_by TEXT, reversed_at TEXT, reversal_reason TEXT, product_id TEXT)`,
		`CREATE TABLE IF NOT EXISTS inspection_items (id TEXT PRIMARY KEY, inspection_id TEXT NOT NULL, item_id TEXT, checkpoint_name TEXT NOT NULL, status TEXT NOT NULL, notes TEXT, images TEXT DEFAULT '[]', created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS item_repair_costs (id TEXT PRIMARY KEY, inventory_item_id TEXT NOT NULL, acquisition_item_id TEXT, repair_date TEXT NOT NULL, repair_type TEXT NOT NULL, cost REAL NOT NULL DEFAULT 0, description TEXT, performed_by TEXT, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT NOT NULL, event_type TEXT NOT NULL, event_date TEXT NOT NULL, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT DEFAULT '{}', created_by TEXT, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS held_sales (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, items TEXT NOT NULL, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS notifications (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, type TEXT NOT NULL, title TEXT NOT NULL, message TEXT NOT NULL, data TEXT DEFAULT '{}', priority TEXT DEFAULT 'medium', status TEXT DEFAULT 'unread', action_url TEXT, action_text TEXT, expires_at TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, read_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS notification_preferences (id TEXT PRIMARY KEY, user_id TEXT NOT NULL UNIQUE, email_enabled INTEGER DEFAULT 1, push_enabled INTEGER DEFAULT 1, low_stock INTEGER DEFAULT 1, debt_overdue INTEGER DEFAULT 1, return_requests INTEGER DEFAULT 1, expense_approval INTEGER DEFAULT 1, sales_updates INTEGER DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS inspections (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, inventory_item_id TEXT, inspector_id TEXT NOT NULL, inspection_date TEXT NOT NULL, result TEXT NOT NULL, condition TEXT, grade TEXT, notes TEXT, images TEXT DEFAULT '[]', test_results TEXT DEFAULT '{}', acquisition_item_id TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS reports (id TEXT PRIMARY KEY, type TEXT NOT NULL, title TEXT NOT NULL, description TEXT, parameters TEXT, data TEXT, status TEXT DEFAULT 'completed', generated_by TEXT, generated_at TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS acquisitions (id TEXT PRIMARY KEY, type TEXT NOT NULL, acquisition_date TEXT NOT NULL, supplier_id TEXT, customer_id TEXT, total_cost REAL DEFAULT 0, paid_amount REAL DEFAULT 0, payment_status TEXT DEFAULT 'payable', status TEXT DEFAULT 'draft', notes TEXT, user_id TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, reversed_at TEXT, reversed_by TEXT, reversal_reason TEXT)`,
		`CREATE TABLE IF NOT EXISTS acquisition_items (id TEXT PRIMARY KEY, acquisition_id TEXT, product_id TEXT, inventory_item_id TEXT, item_code TEXT, serial_number TEXT, condition TEXT, grade TEXT, unit_cost REAL DEFAULT 0, total_cost REAL DEFAULT 0, item_status TEXT DEFAULT 'available', notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		// A historical trade-in may be recorded before an inventory item is
		// materialized. Keep the cloud schema's nullable relationship so one
		// legacy row cannot abort the entire cloud-to-local snapshot.
		`CREATE TABLE IF NOT EXISTS trade_ins (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, inventory_item_id TEXT, purchase_price REAL NOT NULL DEFAULT 0, purchase_date TEXT NOT NULL, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS item_specification_values (id TEXT PRIMARY KEY, inventory_item_id TEXT NOT NULL, specification_id TEXT NOT NULL, value_text TEXT, value_number REAL, value_boolean INTEGER, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(inventory_item_id, specification_id))`,
		`CREATE TABLE IF NOT EXISTS part_types (id TEXT PRIMARY KEY, name_ar TEXT NOT NULL UNIQUE, name_en TEXT NOT NULL, icon TEXT, color TEXT, is_active INTEGER NOT NULL DEFAULT 1, sort_order INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS part_specifications (id TEXT PRIMARY KEY, name_ar TEXT NOT NULL UNIQUE, name_en TEXT NOT NULL, data_type TEXT NOT NULL, options TEXT DEFAULT '[]', is_required INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS type_specifications (id TEXT PRIMARY KEY, part_type_id TEXT NOT NULL, specification_id TEXT NOT NULL, sort_order INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, UNIQUE(part_type_id, specification_id))`,
		`CREATE TABLE IF NOT EXISTS seller_payments (id TEXT PRIMARY KEY, acquisition_id TEXT, customer_id TEXT, amount REAL DEFAULT 0, payment_method TEXT, payment_date TEXT, notes TEXT, user_id TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS daily_sales_summary (date TEXT PRIMARY KEY, total_sales INTEGER DEFAULT 0, total_revenue REAL DEFAULT 0, total_profit REAL DEFAULT 0, total_customers INTEGER DEFAULT 0, average_order_value REAL DEFAULT 0, total_items_sold INTEGER DEFAULT 0, cash_sales REAL DEFAULT 0, card_sales REAL DEFAULT 0, debt_sales REAL DEFAULT 0, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS monthly_sales_summary (year INTEGER NOT NULL, month INTEGER NOT NULL, total_sales INTEGER DEFAULT 0, total_revenue REAL DEFAULT 0, total_profit REAL DEFAULT 0, total_customers INTEGER DEFAULT 0, average_order_value REAL DEFAULT 0, total_items_sold INTEGER DEFAULT 0, cash_sales REAL DEFAULT 0, card_sales REAL DEFAULT 0, debt_sales REAL DEFAULT 0, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (year, month))`,
		`CREATE TABLE IF NOT EXISTS daily_inventory_summary (date TEXT PRIMARY KEY, total_items INTEGER DEFAULT 0, total_value REAL DEFAULT 0, low_stock_count INTEGER DEFAULT 0, out_of_stock_count INTEGER DEFAULT 0, new_items_added INTEGER DEFAULT 0, items_sold INTEGER DEFAULT 0, items_returned INTEGER DEFAULT 0, items_damaged INTEGER DEFAULT 0, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS monthly_inventory_summary (year INTEGER NOT NULL, month INTEGER NOT NULL, total_items INTEGER DEFAULT 0, total_value REAL DEFAULT 0, low_stock_count INTEGER DEFAULT 0, out_of_stock_count INTEGER DEFAULT 0, new_items_added INTEGER DEFAULT 0, items_sold INTEGER DEFAULT 0, items_returned INTEGER DEFAULT 0, items_damaged INTEGER DEFAULT 0, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (year, month))`,
		`CREATE TABLE IF NOT EXISTS daily_debt_summary (date TEXT PRIMARY KEY, total_debt REAL DEFAULT 0, new_debt REAL DEFAULT 0, payments_received REAL DEFAULT 0, overdue_debt REAL DEFAULT 0, overdue_count INTEGER DEFAULT 0, paid_debt REAL DEFAULT 0, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS monthly_debt_summary (year INTEGER NOT NULL, month INTEGER NOT NULL, total_debt REAL DEFAULT 0, new_debt REAL DEFAULT 0, payments_received REAL DEFAULT 0, overdue_debt REAL DEFAULT 0, overdue_count INTEGER DEFAULT 0, paid_debt REAL DEFAULT 0, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (year, month))`,
		`CREATE TABLE IF NOT EXISTS daily_profit_summary (date TEXT PRIMARY KEY, gross_profit REAL DEFAULT 0, net_profit REAL DEFAULT 0, total_revenue REAL DEFAULT 0, total_cost REAL DEFAULT 0, profit_margin REAL DEFAULT 0, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS monthly_profit_summary (year INTEGER NOT NULL, month INTEGER NOT NULL, gross_profit REAL DEFAULT 0, net_profit REAL DEFAULT 0, total_revenue REAL DEFAULT 0, total_cost REAL DEFAULT 0, profit_margin REAL DEFAULT 0, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (year, month))`,
	}
	for _, statement := range compatibilitySchema {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("initialize compatibility schema: %w", err)
		}
	}
	if err := ensureTradeInsInventoryItemNullable(db); err != nil {
		return fmt.Errorf("migrate trade-ins schema: %w", err)
	}
	// Inspection, repair, item-history, and aging workflows were removed from
	// the product. Keep acquisitions and inventory intact, but remove their
	// dedicated local storage so old installations converge to the clean schema.
	for _, statement := range []string{
		`DROP VIEW IF EXISTS used_parts_aging`,
		`DROP TABLE IF EXISTS inspection_items`,
		`DROP TABLE IF EXISTS item_repair_costs`,
		`DROP TABLE IF EXISTS item_history`,
		`DROP TABLE IF EXISTS inspections`,
	} {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("clean removed used-parts workflow schema: %w", err)
		}
	}
	if err := dropRetiredAcquisitionColumns(db); err != nil {
		return fmt.Errorf("clean removed acquisition columns: %w", err)
	}
	/* if _, err := db.Exec(`CREATE VIEW IF NOT EXISTS used_parts_aging AS
		SELECT ii.id AS item_id, ai.acquisition_id, a.acquisition_date,
		CAST(julianday('now') - julianday(a.acquisition_date) AS INTEGER) AS days_in_stock,
		ii.status, ii.condition, ii.purchase_cost AS cost, ii.selling_price AS current_price,
		CASE WHEN julianday('now') - julianday(a.acquisition_date) <= 30 THEN 'fresh' WHEN julianday('now') - julianday(a.acquisition_date) <= 60 THEN 'normal' WHEN julianday('now') - julianday(a.acquisition_date) <= 90 THEN 'aged' ELSE 'long_aged' END AS aging_category,
		CASE WHEN julianday('now') - julianday(a.acquisition_date) > 90 THEN 'critical' WHEN julianday('now') - julianday(a.acquisition_date) > 60 THEN 'warning' ELSE 'none' END AS alert_level
		FROM inventory_items ii JOIN acquisition_items ai ON ii.id = ai.inventory_item_id JOIN acquisitions a ON ai.acquisition_id = a.id
		WHERE a.type = 'CUSTOMER' AND ii.status IN ('AVAILABLE', 'RESERVED')`); err != nil {
		return fmt.Errorf("initialize local aging view: %w", err)
	} */

	if err := migrateLegacySchema(db); err != nil {
		return fmt.Errorf("migrate local database schema: %w", err)
	}
	if err := ensureReturnItemsProductNullable(db); err != nil {
		return fmt.Errorf("migrate return-items schema: %w", err)
	}

	if _, err := db.Exec(`INSERT OR IGNORE INTO settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
		VALUES ('setting-tax-rate', 'tax_rate', '0', 'number', 'financial', 'Tax rate', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		       ('setting-max-discount-rate', 'max_discount_rate', '15', 'number', 'financial', 'Maximum percentage discount', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		       ('setting-store-name', 'store_name', 'PartFlow Store', 'string', 'general', 'Store name', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		       ('setting-currency', 'currency', 'ILS', 'string', 'general', 'Currency', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		return fmt.Errorf("seed local settings: %w", err)
	}

	if _, err := db.Exec(`CREATE VIEW IF NOT EXISTS returns_summary AS
		SELECT r.id, r.return_number, COALESCE(r.reference_number, '') AS reference_number,
			r.sale_id, s.invoice_number AS sale_invoice, r.customer_id, c.name AS customer_name,
			r.return_date, r.return_type, r.status, r.total_refund_amount, r.refund_method,
			r.reason, r.item_condition_after_return, COUNT(ri.id) AS total_items,
			COALESCE(SUM(ri.quantity_returned), 0) AS total_quantity_returned,
			r.created_at, r.updated_at
		FROM returns r
		LEFT JOIN sales s ON r.sale_id = s.id
		LEFT JOIN customers c ON r.customer_id = c.id
		LEFT JOIN return_items ri ON r.id = ri.return_id
		GROUP BY r.id, r.return_number, r.reference_number, r.sale_id, s.invoice_number,
			r.customer_id, c.name, r.return_date, r.return_type, r.status,
			r.total_refund_amount, r.refund_method, r.reason, r.item_condition_after_return,
			r.created_at, r.updated_at`); err != nil {
		return fmt.Errorf("create local returns summary: %w", err)
	}

	return nil
}

// ensureTradeInsInventoryItemNullable upgrades databases created by older
// builds where inventory_item_id was incorrectly declared NOT NULL. SQLite
// cannot alter a column constraint in place, so rebuild this standalone table
// while preserving all existing rows.
func ensureTradeInsInventoryItemNullable(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info('trade_ins')`)
	if err != nil {
		return fmt.Errorf("inspect trade_ins schema: %w", err)
	}
	defer rows.Close()

	needsMigration := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("read trade_ins schema: %w", err)
		}
		if name == "inventory_item_id" && notNull == 1 {
			needsMigration = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read trade_ins schema: %w", err)
	}
	if !needsMigration {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin trade_ins migration: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.Exec(`ALTER TABLE trade_ins RENAME TO trade_ins_legacy`); err != nil {
		return fmt.Errorf("rename legacy trade_ins: %w", err)
	}
	if _, err := tx.Exec(`CREATE TABLE trade_ins (
		id TEXT PRIMARY KEY,
		customer_id TEXT NOT NULL,
		inventory_item_id TEXT,
		purchase_price REAL NOT NULL DEFAULT 0,
		purchase_date TEXT NOT NULL,
		notes TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("create migrated trade_ins: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO trade_ins (id, customer_id, inventory_item_id, purchase_price, purchase_date, notes, created_at, updated_at)
		SELECT id, customer_id, inventory_item_id, purchase_price, purchase_date, notes, created_at, updated_at
		FROM trade_ins_legacy`); err != nil {
		return fmt.Errorf("copy legacy trade_ins: %w", err)
	}
	if _, err := tx.Exec(`DROP TABLE trade_ins_legacy`); err != nil {
		return fmt.Errorf("drop legacy trade_ins: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit trade_ins migration: %w", err)
	}
	committed = true
	return nil
}

// ensureReturnItemsProductNullable upgrades databases created by older builds
// where product_id was incorrectly declared NOT NULL. A cloud return line may
// retain only its sale_item_id, so one legacy row must not abort a full sync.
func ensureReturnItemsProductNullable(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info('return_items')`)
	if err != nil {
		return fmt.Errorf("inspect return_items schema: %w", err)
	}
	defer rows.Close()

	needsMigration := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("read return_items schema: %w", err)
		}
		if name == "product_id" && notNull == 1 {
			needsMigration = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read return_items schema: %w", err)
	}
	if !needsMigration {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin return_items migration: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.Exec(`PRAGMA legacy_alter_table = ON`); err != nil {
		return fmt.Errorf("prepare return_items migration: %w", err)
	}
	if _, err := tx.Exec(`ALTER TABLE return_items RENAME TO return_items_legacy`); err != nil {
		return fmt.Errorf("rename legacy return_items: %w", err)
	}
	if _, err := tx.Exec(`CREATE TABLE return_items (
		id TEXT PRIMARY KEY,
		return_id TEXT NOT NULL,
		product_id TEXT,
		quantity INTEGER NOT NULL DEFAULT 0,
		unit_price REAL NOT NULL DEFAULT 0,
		total_refund_amount REAL NOT NULL DEFAULT 0,
		reason TEXT,
		created_at TEXT NOT NULL,
		sale_item_id TEXT,
		inventory_item_id TEXT,
		quantity_returned INTEGER DEFAULT 0,
		original_quantity INTEGER,
		serial_number TEXT,
		barcode TEXT,
		original_condition TEXT,
		returned_condition TEXT,
		condition_notes TEXT,
		resolution TEXT,
		inventory_status TEXT,
		inspection_required INTEGER DEFAULT 0,
		inspection_date TEXT,
		inspection_result TEXT,
		inspection_notes TEXT,
		original_cost REAL,
		repair_cost REAL DEFAULT 0,
		updated_at TEXT
	)`); err != nil {
		return fmt.Errorf("create migrated return_items: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO return_items (
		id, return_id, product_id, quantity, unit_price, total_refund_amount,
		reason, created_at, sale_item_id, inventory_item_id, quantity_returned,
		original_quantity, serial_number, barcode, original_condition,
		returned_condition, condition_notes, resolution, inventory_status,
		inspection_required, inspection_date, inspection_result, inspection_notes,
		original_cost, repair_cost, updated_at
	) SELECT id, return_id, product_id, quantity, unit_price, total_refund_amount,
		reason, created_at, sale_item_id, inventory_item_id, quantity_returned,
		original_quantity, serial_number, barcode, original_condition,
		returned_condition, condition_notes, resolution, inventory_status,
		inspection_required, inspection_date, inspection_result, inspection_notes,
		original_cost, repair_cost, updated_at
		FROM return_items_legacy`); err != nil {
		return fmt.Errorf("copy legacy return_items: %w", err)
	}
	if _, err := tx.Exec(`DROP TABLE return_items_legacy`); err != nil {
		return fmt.Errorf("drop legacy return_items: %w", err)
	}
	if _, err := tx.Exec(`PRAGMA legacy_alter_table = OFF`); err != nil {
		return fmt.Errorf("finish return_items migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit return_items migration: %w", err)
	}
	committed = true
	return nil
}

func ensureColumnExists(db *sql.DB, tableName, columnName, columnDefinition string) error {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = '%s'", tableName, columnName)
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return fmt.Errorf("check column %s.%s: %w", tableName, columnName, err)
	}
	if count > 0 {
		return nil
	}

	alter := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnDefinition)
	if _, err := db.Exec(alter); err != nil {
		return fmt.Errorf("add column %s.%s: %w", tableName, columnName, err)
	}
	return nil
}

func ensureIndexExists(db *sql.DB, indexName, tableName, columnName string) error {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = '%s'", indexName)
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return fmt.Errorf("check index %s: %w", indexName, err)
	}
	if count > 0 {
		return nil
	}

	createSQL := fmt.Sprintf("CREATE INDEX %s ON %s(%s)", indexName, tableName, columnName)
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("create index %s: %w", indexName, err)
	}
	return nil
}

func dropRetiredAcquisitionColumns(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(acquisition_items)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	columns := map[string]bool{}
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, column := range []string{"inspection_id", "inspection_status"} {
		if columns[column] {
			if _, err := db.Exec(`ALTER TABLE acquisition_items DROP COLUMN ` + column); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateLegacySchema(db *sql.DB) error {
	migrations := []struct {
		tableName  string
		columnName string
		columnDef  string
	}{
		{tableName: "products", columnName: "preferred_supplier_id", columnDef: "preferred_supplier_id TEXT"},
		{tableName: "products", columnName: "brand_id", columnDef: "brand_id TEXT"},
		{tableName: "products", columnName: "model", columnDef: "model TEXT"},
		{tableName: "products", columnName: "barcode", columnDef: "barcode TEXT"},
		{tableName: "products", columnName: "cost_price", columnDef: "cost_price REAL DEFAULT 0"},
		{tableName: "products", columnName: "max_stock_level", columnDef: "max_stock_level INTEGER DEFAULT 0"},
		{tableName: "products", columnName: "warranty_days", columnDef: "warranty_days INTEGER DEFAULT 0"},
		{tableName: "products", columnName: "track_serial", columnDef: "track_serial INTEGER NOT NULL DEFAULT 0"},
		{tableName: "products", columnName: "track_individual", columnDef: "track_individual INTEGER NOT NULL DEFAULT 0"},
		{tableName: "products", columnName: "deleted_at", columnDef: "deleted_at TEXT"},
		{tableName: "customers", columnName: "tax_id", columnDef: "tax_id TEXT"},
		{tableName: "suppliers", columnName: "tax_id", columnDef: "tax_id TEXT"},
		{tableName: "suppliers", columnName: "payment_terms", columnDef: "payment_terms TEXT"},
		{tableName: "inventory_items", columnName: "location_id", columnDef: "location_id TEXT"},
		{tableName: "inventory_items", columnName: "grade", columnDef: "grade TEXT"},
		{tableName: "inventory_items", columnName: "part_type_id", columnDef: "part_type_id TEXT"},
		{tableName: "inventory_items", columnName: "sold_at", columnDef: "sold_at TEXT"},
		{tableName: "inventory_items", columnName: "notes", columnDef: "notes TEXT"},
		{tableName: "acquisition_items", columnName: "condition", columnDef: "condition TEXT"},
		{tableName: "acquisition_items", columnName: "grade", columnDef: "grade TEXT"},
		{tableName: "acquisition_items", columnName: "unit_cost", columnDef: "unit_cost REAL DEFAULT 0"},
		{tableName: "acquisition_items", columnName: "total_cost", columnDef: "total_cost REAL DEFAULT 0"},
		{tableName: "acquisition_items", columnName: "notes", columnDef: "notes TEXT"},
		{tableName: "purchases", columnName: "discount_amount", columnDef: "discount_amount REAL DEFAULT 0"},
		{tableName: "purchases", columnName: "invoice_number", columnDef: "invoice_number TEXT"},
		{tableName: "purchases", columnName: "purchase_date", columnDef: "purchase_date TEXT"},
		{tableName: "purchases", columnName: "expected_delivery_date", columnDef: "expected_delivery_date TEXT"},
		{tableName: "purchases", columnName: "user_id", columnDef: "user_id TEXT"},
		{tableName: "sales", columnName: "invoice_number", columnDef: "invoice_number TEXT"},
		{tableName: "sales", columnName: "sale_number", columnDef: "sale_number TEXT"},
		{tableName: "sales", columnName: "user_id", columnDef: "user_id TEXT"},
		{tableName: "sales", columnName: "sale_date", columnDef: "sale_date TEXT"},
		{tableName: "sales", columnName: "subtotal", columnDef: "subtotal REAL DEFAULT 0"},
		{tableName: "sales", columnName: "cost_amount", columnDef: "cost_amount REAL DEFAULT 0"},
		{tableName: "sales", columnName: "gross_profit", columnDef: "gross_profit REAL DEFAULT 0"},
		{tableName: "sales", columnName: "net_profit", columnDef: "net_profit REAL DEFAULT 0"},
		{tableName: "sales", columnName: "payment_status", columnDef: "payment_status TEXT DEFAULT 'paid'"},
		{tableName: "payments", columnName: "payment_date", columnDef: "payment_date TEXT"},
		{tableName: "sale_items", columnName: "total_amount", columnDef: "total_amount REAL DEFAULT 0"},
		{tableName: "sale_items", columnName: "unit_cost", columnDef: "unit_cost REAL DEFAULT 0"},
		{tableName: "sale_items", columnName: "discount_amount", columnDef: "discount_amount REAL DEFAULT 0"},
		{tableName: "sale_items", columnName: "tax_amount", columnDef: "tax_amount REAL DEFAULT 0"},
		{tableName: "sale_items", columnName: "supplier_id", columnDef: "supplier_id TEXT"},
		{tableName: "payments", columnName: "sale_id", columnDef: "sale_id TEXT"},
		{tableName: "payments", columnName: "purchase_id", columnDef: "purchase_id TEXT"},
		{tableName: "payments", columnName: "payment_status", columnDef: "payment_status TEXT DEFAULT 'completed'"},
		{tableName: "payments", columnName: "created_by", columnDef: "created_by TEXT"},
		{tableName: "payments", columnName: "updated_at", columnDef: "updated_at TEXT"},
		{tableName: "payments", columnName: "payment_date", columnDef: "payment_date TEXT"},
		{tableName: "return_items", columnName: "sale_item_id", columnDef: "sale_item_id TEXT"},
		{tableName: "return_items", columnName: "inventory_item_id", columnDef: "inventory_item_id TEXT"},
		{tableName: "return_items", columnName: "quantity_returned", columnDef: "quantity_returned INTEGER DEFAULT 0"},
		{tableName: "return_items", columnName: "original_quantity", columnDef: "original_quantity INTEGER"},
		{tableName: "return_items", columnName: "serial_number", columnDef: "serial_number TEXT"},
		{tableName: "return_items", columnName: "barcode", columnDef: "barcode TEXT"},
		{tableName: "return_items", columnName: "original_condition", columnDef: "original_condition TEXT"},
		{tableName: "return_items", columnName: "returned_condition", columnDef: "returned_condition TEXT"},
		{tableName: "return_items", columnName: "condition_notes", columnDef: "condition_notes TEXT"},
		{tableName: "return_items", columnName: "resolution", columnDef: "resolution TEXT"},
		{tableName: "return_items", columnName: "inventory_status", columnDef: "inventory_status TEXT"},
		{tableName: "return_items", columnName: "inspection_required", columnDef: "inspection_required INTEGER DEFAULT 0"},
		{tableName: "return_items", columnName: "inspection_date", columnDef: "inspection_date TEXT"},
		{tableName: "return_items", columnName: "inspection_result", columnDef: "inspection_result TEXT"},
		{tableName: "return_items", columnName: "inspection_notes", columnDef: "inspection_notes TEXT"},
		{tableName: "return_items", columnName: "original_cost", columnDef: "original_cost REAL"},
		{tableName: "return_items", columnName: "repair_cost", columnDef: "repair_cost REAL DEFAULT 0"},
		{tableName: "return_items", columnName: "updated_at", columnDef: "updated_at TEXT"},
		{tableName: "inventory_items", columnName: "part_type_id", columnDef: "part_type_id TEXT"},
		{tableName: "expenses", columnName: "title", columnDef: "title TEXT DEFAULT 'Expense'"},
		{tableName: "expenses", columnName: "description", columnDef: "description TEXT"},
		{tableName: "expenses", columnName: "reference_number", columnDef: "reference_number TEXT"},
		{tableName: "expenses", columnName: "category", columnDef: "category TEXT"},
		{tableName: "expenses", columnName: "payment_method", columnDef: "payment_method TEXT"},
		{tableName: "expenses", columnName: "receipt_url", columnDef: "receipt_url TEXT"},
		{tableName: "expenses", columnName: "created_by", columnDef: "created_by TEXT"},
		{tableName: "expenses", columnName: "category_id", columnDef: "category_id TEXT"},
		{tableName: "expenses", columnName: "currency", columnDef: "currency TEXT DEFAULT 'ILS'"},
		{tableName: "expenses", columnName: "reference", columnDef: "reference TEXT"},
		{tableName: "expenses", columnName: "is_recurring", columnDef: "is_recurring INTEGER DEFAULT 0"},
		{tableName: "expenses", columnName: "recurring_period", columnDef: "recurring_period TEXT"},
		{tableName: "expenses", columnName: "approved_by", columnDef: "approved_by TEXT"},
		{tableName: "returns", columnName: "refund_status", columnDef: "refund_status TEXT DEFAULT 'pending'"},
		{tableName: "returns", columnName: "return_number", columnDef: "return_number TEXT DEFAULT ''"},
		{tableName: "returns", columnName: "total_refund_amount", columnDef: "total_refund_amount REAL DEFAULT 0"},
		{tableName: "warranty_claims", columnName: "claim_number", columnDef: "claim_number TEXT DEFAULT ''"},
		{tableName: "warranty_claims", columnName: "serial_number", columnDef: "serial_number TEXT"},
		{tableName: "returns", columnName: "reference_number", columnDef: "reference_number TEXT"},
		{tableName: "returns", columnName: "return_type", columnDef: "return_type TEXT"},
		{tableName: "returns", columnName: "refund_method", columnDef: "refund_method TEXT"},
		{tableName: "returns", columnName: "refund_date", columnDef: "refund_date TEXT"},
		{tableName: "returns", columnName: "refund_reference", columnDef: "refund_reference TEXT"},
		{tableName: "returns", columnName: "debt_id", columnDef: "debt_id TEXT"},
		{tableName: "returns", columnName: "debt_adjustment", columnDef: "debt_adjustment REAL DEFAULT 0"},
		{tableName: "returns", columnName: "customer_credit", columnDef: "customer_credit REAL DEFAULT 0"},
		{tableName: "returns", columnName: "reason_detail", columnDef: "reason_detail TEXT"},
		{tableName: "returns", columnName: "item_condition_after_return", columnDef: "item_condition_after_return TEXT"},
		{tableName: "returns", columnName: "is_warranty_claim", columnDef: "is_warranty_claim INTEGER DEFAULT 0"},
		{tableName: "returns", columnName: "warranty_id", columnDef: "warranty_id TEXT"},
		{tableName: "returns", columnName: "warranty_valid_until", columnDef: "warranty_valid_until TEXT"},
		{tableName: "returns", columnName: "created_by", columnDef: "created_by TEXT"},
		{tableName: "returns", columnName: "processed_by", columnDef: "processed_by TEXT"},
		{tableName: "returns", columnName: "approved_by", columnDef: "approved_by TEXT"},
		{tableName: "returns", columnName: "approved_at", columnDef: "approved_at TEXT"},
		{tableName: "returns", columnName: "internal_notes", columnDef: "internal_notes TEXT"},
		{tableName: "audit_logs", columnName: "request_id", columnDef: "request_id TEXT"},
		{tableName: "audit_logs", columnName: "changes", columnDef: "changes TEXT"},
		{tableName: "audit_logs", columnName: "description", columnDef: "description TEXT"},
		{tableName: "audit_logs", columnName: "status", columnDef: "status TEXT DEFAULT 'success'"},
		{tableName: "audit_logs", columnName: "error_message", columnDef: "error_message TEXT"},
		{tableName: "audit_logs", columnName: "metadata", columnDef: "metadata TEXT DEFAULT '{}'"},
	}

	for _, migration := range migrations {
		if err := ensureColumnExists(db, migration.tableName, migration.columnName, migration.columnDef); err != nil {
			return err
		}
	}

	if err := ensureIndexExists(db, "idx_products_preferred_supplier", "products", "preferred_supplier_id"); err != nil {
		return err
	}
	if err := ensureIndexExists(db, "idx_products_barcode", "products", "barcode"); err != nil {
		return err
	}

	legacyTableStatements := []string{
		`CREATE TABLE IF NOT EXISTS brands (id TEXT PRIMARY KEY, name TEXT NOT NULL UNIQUE, description TEXT, logo_url TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS expense_categories (id TEXT PRIMARY KEY, name TEXT NOT NULL UNIQUE, description TEXT, color TEXT, icon TEXT, budget REAL DEFAULT 0, is_active INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS expenses (id TEXT PRIMARY KEY, title TEXT NOT NULL, category_id TEXT, amount REAL NOT NULL DEFAULT 0, currency TEXT DEFAULT 'ILS', reference TEXT, notes TEXT, expense_date TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'approved', is_recurring INTEGER NOT NULL DEFAULT 0, recurring_period TEXT, approved_by TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS returns (id TEXT PRIMARY KEY, return_number TEXT NOT NULL UNIQUE, customer_id TEXT, sale_id TEXT, purchase_id TEXT, total_refund_amount REAL NOT NULL DEFAULT 0, refund_status TEXT NOT NULL DEFAULT 'pending', status TEXT NOT NULL DEFAULT 'pending', reason TEXT, notes TEXT, return_date TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS return_items (id TEXT PRIMARY KEY, return_id TEXT NOT NULL, product_id TEXT, quantity INTEGER NOT NULL DEFAULT 0, unit_price REAL NOT NULL DEFAULT 0, total_refund_amount REAL NOT NULL DEFAULT 0, reason TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS warranty_claims (id TEXT PRIMARY KEY, claim_number TEXT NOT NULL UNIQUE, customer_id TEXT, product_id TEXT, serial_number TEXT, issue_description TEXT, status TEXT NOT NULL DEFAULT 'pending', created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
	}

	for _, stmt := range legacyTableStatements {
		if _, err := db.Exec(stmt); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}
	return nil
}

func SeedLocalSnapshot(db *sql.DB, snapshot map[string]any) error {
	if db == nil {
		return fmt.Errorf("local database is nil")
	}
	if len(snapshot) == 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin local snapshot transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	tableOrder := []struct {
		key   string
		table string
	}{
		{key: "categories", table: "categories"},
		{key: "brands", table: "brands"},
		{key: "suppliers", table: "suppliers"},
		{key: "customers", table: "customers"},
		{key: "customer_ledger", table: "customer_ledger"},
		{key: "supplier_ledger", table: "supplier_ledger"},
		{key: "ledger_entries", table: "ledger_entries"},
		{key: "products", table: "products"},
		{key: "inventory", table: "inventory"},
		{key: "locations", table: "locations"},
		{key: "sales", table: "sales"},
		{key: "sale_items", table: "sale_items"},
		{key: "purchases", table: "purchases"},
		{key: "purchase_items", table: "purchase_items"},
		{key: "payments", table: "payments"},
		{key: "debts", table: "debts"},
		{key: "expense_categories", table: "expense_categories"},
		// expenses.category_id references expense_categories(id), so seed the
		// parent rows first while foreign-key enforcement is enabled.
		{key: "expenses", table: "expenses"},
		{key: "seller_payments", table: "seller_payments"},
		{key: "supplier_returns", table: "supplier_returns"},
		{key: "supplier_return_items", table: "supplier_return_items"},
		{key: "inspections", table: "inspections"},
		{key: "inspection_items", table: "inspection_items"},
		{key: "part_types", table: "part_types"},
		{key: "part_specifications", table: "part_specifications"},
		{key: "type_specifications", table: "type_specifications"},
		{key: "acquisitions", table: "acquisitions"},
		{key: "acquisition_items", table: "acquisition_items"},
		{key: "trade_ins", table: "trade_ins"},
		{key: "item_specification_values", table: "item_specification_values"},
		{key: "returns", table: "returns"},
		{key: "return_items", table: "return_items"},
		{key: "used_parts", table: "used_parts"},
		{key: "inventory_items", table: "inventory_items"},
		{key: "inventory_movements", table: "inventory_movements"},
		{key: "reservations", table: "reservations"},
		{key: "barcodes", table: "barcodes"},
		{key: "notifications", table: "notifications"},
		{key: "notification_preferences", table: "notification_preferences"},
		{key: "reports", table: "reports"},
		{key: "settings", table: "settings"},
		{key: "held_sales", table: "held_sales"},
	}

	for _, item := range tableOrder {
		raw, ok := snapshot[item.key]
		if !ok || raw == nil {
			continue
		}
		rows, err := normalizeSnapshotRows(raw)
		if err != nil {
			return fmt.Errorf("normalize %s snapshot: %w", item.key, err)
		}
		if len(rows) == 0 {
			continue
		}
		if err := upsertSnapshotRows(tx, db, item.table, rows); err != nil {
			return fmt.Errorf("seed %s: %w", item.table, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit local snapshot transaction: %w", err)
	}
	committed = true

	return nil
}

func normalizeSnapshotRows(raw any) ([]map[string]any, error) {
	switch rows := raw.(type) {
	case []map[string]any:
		return rows, nil
	case []any:
		out := make([]map[string]any, 0, len(rows))
		for i, row := range rows {
			mapped, ok := row.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("row %d is %T, expected map", i, row)
			}
			out = append(out, mapped)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported snapshot type %T", raw)
	}
}

type snapshotExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func upsertSnapshotRows(executor snapshotExecutor, db *sql.DB, tableName string, rows []map[string]any) error {
	allowedFields, err := localSnapshotFields(db, tableName)
	if err != nil {
		return err
	}
	for _, row := range rows {
		normalized := normalizeSnapshotRow(row)
		if tableName == "sales" && isBlankSnapshotValue(normalized["sale_number"]) {
			normalized["sale_number"] = normalized["invoice_number"]
		}
		if tableName == "purchases" && isBlankSnapshotValue(normalized["purchase_number"]) {
			normalized["purchase_number"] = normalized["invoice_number"]
		}
		if tableName == "sale_items" || tableName == "purchase_items" {
			if _, ok := normalized["item_total"]; !ok {
				normalized["item_total"] = normalized["total_amount"]
			}
		}
		if tableName == "payments" && isBlankSnapshotValue(normalized["transaction_number"]) {
			normalized["transaction_number"] = normalized["reference_number"]
			if normalized["transaction_number"] == nil || normalized["transaction_number"] == "" {
				normalized["transaction_number"] = normalized["id"]
			}
		}
		if tableName == "customer_ledger" || tableName == "supplier_ledger" {
			if isBlankSnapshotValue(normalized["type"]) {
				normalized["type"] = ledgerTypeAlias(normalized["transaction_type"])
			}
		}
		if tableName == "expenses" {
			// The cloud schema calls this field description/category, while the
			// local schema requires a non-null title.
			if value, ok := normalized["title"]; !ok || strings.TrimSpace(fmt.Sprint(value)) == "" {
				for _, fallback := range []string{"description", "category", "reference_number"} {
					if value, ok := normalized[fallback]; ok && strings.TrimSpace(fmt.Sprint(value)) != "" {
						normalized["title"] = value
						break
					}
				}
			}
			if value, ok := normalized["title"]; !ok || strings.TrimSpace(fmt.Sprint(value)) == "" {
				normalized["title"] = "Expense"
			}
		}
		if tableName == "return_items" {
			// The enhanced cloud schema uses quantity_returned; SQLite keeps the
			// legacy quantity column for the list/repository queries.
			if value, ok := normalized["quantity"]; !ok || strings.TrimSpace(fmt.Sprint(value)) == "" {
				normalized["quantity"] = normalized["quantity_returned"]
			}
		}
		if len(normalized) == 0 {
			continue
		}
		fields := make([]string, 0, len(normalized))
		values := make([]any, 0, len(normalized))
		for _, field := range allowedFields {
			value, ok := normalized[field]
			if !ok {
				continue
			}
			normalizedValue := normalizeSnapshotValue(value)
			if normalizedValue == nil {
				continue
			}
			fields = append(fields, field)
			values = append(values, normalizedValue)
		}
		if len(fields) == 0 {
			continue
		}
		placeholders := make([]string, len(fields))
		updates := make([]string, 0, len(fields))
		conflictTarget := "id"
		if tableName == "settings" {
			// Settings are identified by their stable key. Local bootstrap
			// settings can have a different generated id than the cloud row.
			conflictTarget = "key"
		}
		for i, field := range fields {
			placeholders[i] = "?"
			if field == "id" {
				continue
			}
			updates = append(updates, fmt.Sprintf("%s = excluded.%s", field, field))
		}
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(fields, ", "), strings.Join(placeholders, ", "))
		if len(updates) > 0 {
			query += " ON CONFLICT(" + conflictTarget + ") DO UPDATE SET " + strings.Join(updates, ", ")
		} else {
			query += " ON CONFLICT(" + conflictTarget + ") DO NOTHING"
		}
		if _, err := executor.Exec(query, values...); err != nil {
			return fmt.Errorf("insert row into %s: %w", tableName, err)
		}
	}
	return nil
}

func localSnapshotFields(db *sql.DB, tableName string) ([]string, error) {
	available := make(map[string]struct{})
	quotedTable := strings.ReplaceAll(tableName, "'", "''")
	rows, err := db.Query("PRAGMA table_info('" + quotedTable + "')")
	if err != nil {
		return nil, fmt.Errorf("inspect local table %s: %w", tableName, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return nil, fmt.Errorf("inspect local table %s columns: %w", tableName, err)
		}
		available[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("inspect local table %s columns: %w", tableName, err)
	}
	if len(available) == 0 {
		return nil, fmt.Errorf("local snapshot table %s does not exist", tableName)
	}

	fields := make([]string, 0, len(available))
	for _, field := range fieldMapForTable(tableName) {
		if _, ok := available[field]; ok {
			fields = append(fields, field)
		}
	}
	return fields, nil
}

func isBlankSnapshotValue(value any) bool {
	if value == nil {
		return true
	}
	return strings.TrimSpace(fmt.Sprint(value)) == ""
}

func ledgerTypeAlias(value any) string {
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
	case "sale", "payment", "return", "refund", "adjustment", "purchase":
		if strings.EqualFold(strings.TrimSpace(fmt.Sprint(value)), "sale") || strings.EqualFold(strings.TrimSpace(fmt.Sprint(value)), "purchase") {
			return "debit"
		}
		return "credit"
	default:
		return ""
	}
}

func fieldMapForTable(tableName string) []string {
	switch tableName {
	case "categories":
		return []string{"id", "name", "description", "parent_id", "icon", "color", "is_active", "created_at", "updated_at"}
	case "brands":
		return []string{"id", "name", "description", "logo_url", "created_at", "updated_at"}
	case "products":
		return []string{"id", "sku", "name", "description", "category_id", "brand_id", "preferred_supplier_id", "model", "barcode", "purchase_price", "cost_price", "selling_price", "currency", "min_stock_level", "max_stock_level", "warranty_days", "track_serial", "track_individual", "is_active", "deleted_at", "created_at", "updated_at"}
	case "customers":
		return []string{"id", "code", "name", "email", "phone", "address", "city", "country", "tax_id", "credit_limit", "current_balance", "notes", "is_active", "created_at", "updated_at"}
	case "suppliers":
		return []string{"id", "code", "name", "email", "phone", "address", "city", "country", "tax_id", "credit_limit", "payment_terms", "current_balance", "notes", "is_active", "created_at", "updated_at"}
	case "customer_ledger":
		return []string{"id", "customer_id", "type", "transaction_type", "amount", "balance", "description", "reference_id", "reference_type", "created_by", "created_at"}
	case "supplier_ledger":
		return []string{"id", "supplier_id", "type", "transaction_type", "amount", "balance", "description", "reference_id", "reference_type", "created_by", "created_at"}
	case "ledger_entries":
		return []string{"id", "ledger_type", "entity_id", "transaction_type", "reference_id", "reference_type", "amount", "balance", "previous_balance", "description", "metadata", "created_by", "created_at", "cost_before", "cost_after", "value_before", "value_after", "is_reversed", "reversed_by", "reversed_at", "reversal_reason", "product_id"}
	case "inventory":
		return []string{"id", "product_id", "quantity", "reserved_quantity", "location", "warehouse_id", "current_quantity", "available_quantity", "current_cost", "current_value", "last_movement_id", "last_restocked_at", "created_at", "updated_at"}
	case "sales":
		return []string{"id", "sale_number", "invoice_number", "customer_id", "sale_date", "subtotal", "tax_amount", "discount_amount", "total_amount", "paid_amount", "remaining_amount", "payment_method", "payment_status", "status", "notes", "created_at", "updated_at"}
	case "sale_items":
		return []string{"id", "sale_id", "inventory_item_id", "product_id", "quantity", "unit_price", "item_total", "total_amount", "unit_cost", "created_at"}
	case "purchases":
		return []string{"id", "purchase_number", "invoice_number", "supplier_id", "user_id", "purchase_date", "expected_delivery_date", "discount_amount", "tax_amount", "total_amount", "paid_amount", "remaining_amount", "status", "notes", "created_at", "updated_at"}
	case "purchase_items":
		return []string{"id", "purchase_id", "product_id", "quantity", "unit_price", "item_total", "total_amount", "unit_cost", "created_at"}
	case "payments":
		return []string{"id", "transaction_number", "customer_id", "supplier_id", "amount", "payment_method", "reference", "notes", "created_at", "payment_date"}
	case "debts":
		return []string{"id", "customer_id", "sale_id", "amount", "paid_amount", "remaining_amount", "due_date", "status", "notes", "created_at", "updated_at"}
	case "expenses":
		return []string{"id", "title", "category_id", "reference_number", "category", "amount", "description", "expense_date", "payment_method", "receipt_url", "created_by", "currency", "reference", "notes", "status", "is_recurring", "recurring_period", "approved_by", "created_at", "updated_at"}
	case "expense_categories":
		return []string{"id", "name", "description", "color", "icon", "budget", "is_active", "created_at", "updated_at"}
	case "locations":
		return []string{"id", "name", "type", "parent_id", "warehouse_id", "description", "is_active", "created_at", "updated_at"}
	case "inventory_movements":
		return []string{"id", "item_id", "product_id", "movement_type", "quantity", "before_quantity", "after_quantity", "reference_type", "reference_id", "reason", "created_by", "created_at", "is_reversed", "reversed_by", "reversed_at", "reversal_reason"}
	case "reservations":
		return []string{"id", "item_id", "customer_id", "user_id", "reserved_at", "expires_at", "status", "notes", "created_at", "updated_at"}
	case "barcodes":
		return []string{"id", "code", "product_id", "inventory_item_id", "type", "is_active", "generated_at", "created_at", "updated_at"}
	case "inspections":
		return []string{"id", "product_id", "inventory_item_id", "inspector_id", "inspection_date", "result", "condition", "grade", "notes", "images", "created_at", "updated_at"}
	case "inspection_items":
		return []string{"id", "inspection_id", "item_id", "checkpoint_name", "status", "notes", "images", "created_at"}
	case "notifications":
		return []string{"id", "user_id", "type", "title", "message", "data", "priority", "status", "action_url", "action_text", "expires_at", "created_at", "updated_at", "read_at"}
	case "notification_preferences":
		return []string{"id", "user_id", "email_enabled", "push_enabled", "low_stock", "debt_overdue", "return_requests", "expense_approval", "sales_updates", "created_at", "updated_at"}
	case "reports":
		return []string{"id", "type", "title", "description", "parameters", "data", "status", "generated_by", "generated_at", "created_at", "updated_at"}
	case "settings":
		return []string{"id", "key", "value", "value_type", "category", "description", "is_public", "created_at", "updated_at"}
	case "held_sales":
		return []string{"id", "user_id", "items", "created_at"}
	case "part_types":
		return []string{"id", "name_ar", "name_en", "icon", "color", "is_active", "sort_order", "created_at", "updated_at"}
	case "part_specifications":
		return []string{"id", "name_ar", "name_en", "data_type", "options", "is_required", "created_at"}
	case "type_specifications":
		return []string{"id", "part_type_id", "specification_id", "sort_order", "created_at"}
	case "acquisitions":
		return []string{"id", "type", "acquisition_date", "supplier_id", "customer_id", "total_cost", "paid_amount", "payment_status", "status", "notes", "user_id", "created_at", "updated_at", "reversed_at", "reversed_by", "reversal_reason"}
	case "acquisition_items":
		return []string{"id", "acquisition_id", "product_id", "inventory_item_id", "item_code", "serial_number", "condition", "grade", "unit_cost", "total_cost", "item_status", "notes", "created_at", "updated_at"}
	case "trade_ins":
		return []string{"id", "customer_id", "inventory_item_id", "purchase_price", "purchase_date", "notes", "created_at", "updated_at"}
	case "item_specification_values":
		return []string{"id", "inventory_item_id", "specification_id", "value_text", "value_number", "value_boolean", "created_at", "updated_at"}
	case "seller_payments":
		return []string{"id", "acquisition_id", "customer_id", "amount", "payment_method", "payment_date", "notes", "user_id", "created_at"}
	case "supplier_returns":
		return []string{"id", "purchase_id", "supplier_id", "return_number", "status", "reason", "refund_amount", "notes", "created_by", "created_at", "updated_at"}
	case "supplier_return_items":
		return []string{"id", "supplier_return_id", "purchase_item_id", "product_id", "quantity", "unit_cost", "created_at"}
	case "returns":
		return []string{"id", "return_number", "reference_number", "sale_id", "purchase_id", "customer_id", "return_date", "return_type", "status", "total_refund_amount", "refund_method", "refund_date", "refund_reference", "debt_id", "debt_adjustment", "customer_credit", "reason", "reason_detail", "item_condition_after_return", "is_warranty_claim", "warranty_id", "warranty_valid_until", "created_by", "processed_by", "approved_by", "approved_at", "notes", "internal_notes", "refund_status", "created_at", "updated_at"}
	case "return_items":
		return []string{
			"id", "return_id", "sale_item_id", "product_id", "inventory_item_id",
			"quantity", "quantity_returned", "original_quantity", "serial_number", "barcode",
			"unit_price", "total_refund_amount", "reason", "original_condition", "returned_condition",
			"condition_notes", "resolution", "inventory_status", "inspection_required", "inspection_date",
			"inspection_result", "inspection_notes", "original_cost", "repair_cost", "created_at", "updated_at",
		}
	case "used_parts":
		return []string{"id", "product_id", "quantity", "status", "created_at", "updated_at"}
	case "inventory_items":
		return []string{"id", "product_id", "part_type_id", "item_code", "barcode", "serial_number", "condition", "grade", "purchase_cost", "selling_price", "status", "location_id", "supplier_id", "purchase_date", "sold_at", "notes", "created_at", "updated_at"}
	default:
		return nil
	}
}

func normalizeSnapshotRow(row map[string]any) map[string]any {
	out := make(map[string]any, len(row))
	for key, value := range row {
		normalizedKey := normalizeSnapshotColumn(key)
		if normalizedKey == "" {
			continue
		}
		if value == nil {
			out[normalizedKey] = ""
			continue
		}
		out[normalizedKey] = value
	}
	return out
}

func normalizeSnapshotColumn(key string) string {
	if key == "" {
		return ""
	}
	var builder strings.Builder
	for i, r := range key {
		if r == '-' || r == ' ' || r == '.' {
			if i > 0 && builder.Len() > 0 && builder.String()[builder.Len()-1] != '_' {
				builder.WriteRune('_')
			}
			continue
		}
		if unicode.IsUpper(r) {
			if i > 0 && builder.Len() > 0 && builder.String()[builder.Len()-1] != '_' {
				builder.WriteRune('_')
			}
			r = unicode.ToLower(r)
		}
		builder.WriteRune(r)
	}
	result := strings.Trim(builder.String(), "_")
	return result
}

func normalizeSnapshotValue(value any) any {
	switch v := value.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case []byte:
		return string(v)
	case map[string]any, []any:
		payload, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(payload)
	case nil:
		return nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return v
	default:
		return v
	}
}

func SetMetadata(db *sql.DB, key, value string) error {
	_, err := db.Exec(`
		INSERT INTO local_metadata (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, key, value)
	if err != nil {
		return fmt.Errorf("set local metadata: %w", err)
	}
	return nil
}

func GetMetadata(db *sql.DB, key string) (string, error) {
	var value string
	if err := db.QueryRow(`
		SELECT value FROM local_metadata WHERE key = ? LIMIT 1
	`, key).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("read local metadata: %w", err)
	}
	return strings.TrimSpace(value), nil
}

func EnqueueSyncOperation(db *sql.DB, entityType, entityID, operation, payload string) (string, error) {
	idempotencyKey := fmt.Sprintf("%s:%s:%s:%d", entityType, entityID, operation, time.Now().UnixNano())
	if _, err := db.Exec(`
		INSERT INTO sync_queue (id, entity_type, entity_id, operation, payload, idempotency_key, created_at, attempts, last_error, next_retry_at, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 0, NULL, CURRENT_TIMESTAMP, NULL)
		ON CONFLICT(idempotency_key) DO NOTHING
	`, idempotencyKey, entityType, entityID, operation, payload, idempotencyKey); err != nil {
		return "", fmt.Errorf("enqueue sync operation: %w", err)
	}
	return idempotencyKey, nil
}

func GetPendingSyncCount(db *sql.DB) (int, error) {
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM sync_queue
		WHERE synced_at IS NULL
		  AND (next_retry_at IS NULL OR next_retry_at <= CURRENT_TIMESTAMP)
	`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count pending sync entries: %w", err)
	}
	return count, nil
}

type SyncQueueEntry struct {
	ID          string
	EntityType  string
	EntityID    string
	Operation   string
	Payload     string
	Idempotency string
	CreatedAt   string
	Attempts    int
	LastError   string
	NextRetryAt string
}

func ListPendingSyncOperations(db *sql.DB, limit int) ([]SyncQueueEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.Query(`
		SELECT id, entity_type, entity_id, operation, payload, idempotency_key, created_at, attempts, last_error, next_retry_at
		FROM sync_queue
		WHERE synced_at IS NULL
		  AND (next_retry_at IS NULL OR next_retry_at <= CURRENT_TIMESTAMP)
		ORDER BY next_retry_at ASC, created_at ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending sync operations: %w", err)
	}
	defer rows.Close()

	var entries []SyncQueueEntry
	for rows.Next() {
		var item SyncQueueEntry
		if err := rows.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Operation, &item.Payload, &item.Idempotency, &item.CreatedAt, &item.Attempts, &item.LastError, &item.NextRetryAt); err != nil {
			return nil, fmt.Errorf("scan pending sync operation: %w", err)
		}
		entries = append(entries, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending sync operations: %w", err)
	}
	return entries, nil
}

func MarkSyncOperationSucceeded(db *sql.DB, queueID string) error {
	_, err := db.Exec(`
		UPDATE sync_queue
		SET synced_at = CURRENT_TIMESTAMP,
		    last_error = NULL,
		    next_retry_at = NULL,
		    attempts = attempts + 1
		WHERE id = ?
	`, queueID)
	if err != nil {
		return fmt.Errorf("mark sync operation succeeded: %w", err)
	}
	return nil
}

func MarkSyncOperationFailed(db *sql.DB, queueID, errMessage string) error {
	var attempts int
	if err := db.QueryRow(`SELECT attempts FROM sync_queue WHERE id = ?`, queueID).Scan(&attempts); err != nil {
		return fmt.Errorf("load attempt count for retry backoff: %w", err)
	}

	retryDelaySeconds := nextRetryDelaySeconds(attempts)
	_, err := db.Exec(`
		UPDATE sync_queue
		SET attempts = attempts + 1,
		    last_error = ?,
		    next_retry_at = datetime(CURRENT_TIMESTAMP, '+' || ? || ' seconds'),
		    synced_at = CASE WHEN synced_at IS NOT NULL THEN synced_at ELSE NULL END
		WHERE id = ?
	`, errMessage, retryDelaySeconds, queueID)
	if err != nil {
		return fmt.Errorf("mark sync operation failed: %w", err)
	}
	return nil
}

func nextRetryDelaySeconds(attempts int) int {
	if attempts < 0 {
		attempts = 0
	}
	if attempts == 0 {
		return 30
	}

	delay := 30 * (1 << smallestInt(5, attempts))
	if delay > 900 {
		return 900
	}
	return delay
}

func smallestInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func SaveLocalSession(db *sql.DB, userID, email, displayName, accessToken, refreshToken, expiresAt string) error {
	_, err := db.Exec(`
		INSERT INTO local_sessions (id, user_id, email, display_name, access_token, refresh_token, expires_at, created_at, updated_at)
		VALUES ('current', ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
		    user_id = excluded.user_id,
		    email = excluded.email,
		    display_name = excluded.display_name,
		    access_token = excluded.access_token,
		    refresh_token = excluded.refresh_token,
		    expires_at = excluded.expires_at,
		    updated_at = CURRENT_TIMESTAMP
	`, userID, email, displayName, accessToken, refreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("save local session: %w", err)
	}
	return nil
}

func GetLocalSession(db *sql.DB) (map[string]string, error) {
	row := db.QueryRow(`
		SELECT user_id, email, display_name, access_token, refresh_token, expires_at
		FROM local_sessions
		WHERE id = 'current'
	`)

	var userID, email, displayName, accessToken, refreshToken, expiresAt string
	if err := row.Scan(&userID, &email, &displayName, &accessToken, &refreshToken, &expiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("read local session: %w", err)
	}

	return map[string]string{
		"user_id":       userID,
		"email":         email,
		"display_name":  displayName,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt,
	}, nil
}

func ClearLocalSession(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM local_sessions WHERE id = 'current'`)
	if err != nil {
		return fmt.Errorf("clear local session: %w", err)
	}
	return nil
}

func RecordSyncConflict(db *sql.DB, entityType, entityID, entityTable, operation, localUpdatedAt, remoteUpdatedAt, reason, payload string) error {
	conflictType := classifySyncConflict(entityTable, entityType, payload)
	id := fmt.Sprintf("%s:%s:%s:%d", entityType, entityID, operation, time.Now().UnixNano())
	_, err := db.Exec(`
		INSERT INTO sync_conflicts (id, entity_type, entity_id, entity_table, operation, conflict_type, local_updated_at, remote_updated_at, conflict_reason, payload, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, id, entityType, entityID, entityTable, operation, conflictType, localUpdatedAt, remoteUpdatedAt, reason, payload)
	if err != nil {
		return fmt.Errorf("record sync conflict: %w", err)
	}
	return nil
}

func ListSyncConflicts(db *sql.DB, limit int) ([]map[string]string, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := db.Query(`
		SELECT id, entity_type, entity_id, entity_table, operation, conflict_type, local_updated_at, remote_updated_at, conflict_reason, payload, created_at
		FROM sync_conflicts
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list sync conflicts: %w", err)
	}
	defer rows.Close()

	entries := make([]map[string]string, 0)
	for rows.Next() {
		var id, entityType, entityID, entityTable, operation, conflictType, localUpdatedAt, remoteUpdatedAt, conflictReason, payload, createdAt string
		if err := rows.Scan(&id, &entityType, &entityID, &entityTable, &operation, &conflictType, &localUpdatedAt, &remoteUpdatedAt, &conflictReason, &payload, &createdAt); err != nil {
			return nil, fmt.Errorf("scan sync conflict: %w", err)
		}
		entries = append(entries, map[string]string{
			"id":                id,
			"entity_type":       entityType,
			"entity_id":         entityID,
			"entity_table":      entityTable,
			"operation":         operation,
			"conflict_type":     conflictType,
			"local_updated_at":  localUpdatedAt,
			"remote_updated_at": remoteUpdatedAt,
			"conflict_reason":   conflictReason,
			"payload":           payload,
			"created_at":        createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sync conflicts: %w", err)
	}
	return entries, nil
}

// GetSyncConflict returns one pending conflict, including the original local
// payload. Keeping this lookup in localdb prevents handlers from deleting a
// conflict before its selected policy has actually been applied.
func GetSyncConflict(db *sql.DB, conflictID string) (map[string]string, error) {
	var conflict = make(map[string]string)
	var id, entityType, entityID, entityTable, operation, conflictType, localUpdatedAt, remoteUpdatedAt, reason, payload, createdAt string
	err := db.QueryRow(`SELECT id, entity_type, entity_id, entity_table, operation, conflict_type, local_updated_at, remote_updated_at, conflict_reason, payload, created_at FROM sync_conflicts WHERE id = ?`, conflictID).
		Scan(&id, &entityType, &entityID, &entityTable, &operation, &conflictType, &localUpdatedAt, &remoteUpdatedAt, &reason, &payload, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sync conflict: %w", err)
	}
	conflict["id"] = id
	conflict["entity_type"] = entityType
	conflict["entity_id"] = entityID
	conflict["entity_table"] = entityTable
	conflict["operation"] = operation
	conflict["conflict_type"] = conflictType
	conflict["local_updated_at"] = localUpdatedAt
	conflict["remote_updated_at"] = remoteUpdatedAt
	conflict["conflict_reason"] = reason
	conflict["payload"] = payload
	conflict["created_at"] = createdAt
	return conflict, nil
}

func ClearSyncConflicts(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM sync_conflicts`)
	if err != nil {
		return fmt.Errorf("clear sync conflicts: %w", err)
	}
	return nil
}

func DeleteSyncConflict(db *sql.DB, conflictID string) error {
	_, err := db.Exec(`DELETE FROM sync_conflicts WHERE id = ?`, conflictID)
	if err != nil {
		return fmt.Errorf("delete sync conflict: %w", err)
	}
	return nil
}

func classifySyncConflict(entityTable, entityType, payload string) string {
	lowerTable := strings.ToLower(entityTable)
	lowerType := strings.ToLower(entityType)
	lowerPayload := strings.ToLower(payload)

	switch {
	case lowerTable == "inventory_items" || strings.Contains(lowerTable, "inventory") || strings.Contains(lowerPayload, "quantity") || strings.Contains(lowerPayload, "stock"):
		return "stock"
	case lowerTable == "payments" || (strings.Contains(lowerPayload, "amount") && strings.Contains(lowerType, "payment")) || strings.Contains(lowerPayload, "payment_method"):
		return "payment"
	case lowerTable == "debts" || strings.Contains(lowerPayload, "remaining_amount") || strings.Contains(lowerPayload, "paid_amount") || strings.Contains(lowerPayload, "due_date"):
		return "debt"
	default:
		return "generic"
	}
}
