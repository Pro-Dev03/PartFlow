package dashboard

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	_ "modernc.org/sqlite"
)

func TestFetchTodayMetricsIsDateScoped(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL, tax_amount REAL, status TEXT, payment_method TEXT, paid_amount REAL, sale_date TEXT, created_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, inventory_item_id TEXT, product_id TEXT, quantity INTEGER, unit_cost REAL, total_amount REAL, tax_amount REAL);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, purchase_cost REAL);
		CREATE TABLE products (id TEXT PRIMARY KEY, purchase_price REAL, cost_price REAL);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, status TEXT, expense_date TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, return_date TEXT, status TEXT, reference_number TEXT, total_refund_amount REAL);
		CREATE TABLE return_items (id TEXT PRIMARY KEY, return_id TEXT, sale_item_id TEXT, quantity_returned INTEGER, total_refund_amount REAL, original_cost REAL);
		CREATE TABLE payments (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, payment_date TEXT, created_at TEXT);
		CREATE TABLE customer_payments (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, payment_date TEXT);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	now := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)
	today, err := accounting.StoreDate(now)
	if err != nil {
		t.Fatal(err)
	}
	yesterday, err := accounting.StoreDate(now.AddDate(0, 0, -1))
	if err != nil {
		t.Fatal(err)
	}
	todayTimestamp := today + "T12:00:00Z"
	yesterdayTimestamp := yesterday + "T12:00:00Z"

	_, err = db.Exec(`INSERT INTO sales (id, total_amount, tax_amount, status, payment_method, paid_amount, sale_date, created_at) VALUES
		('today', 100, 0, 'completed', 'cash', 100, ?, ?),
		('today-taxed', 115, 15, 'completed', 'cash', 115, ?, ?),
		('yesterday', 1000, 0, 'completed', 'cash', 1000, ?, ?),
		('cancelled-today', 500, 0, 'cancelled', 'cash', 500, ?, ?)`,
		todayTimestamp, todayTimestamp, todayTimestamp, todayTimestamp,
		yesterdayTimestamp, yesterdayTimestamp, todayTimestamp, todayTimestamp)
	if err != nil {
		t.Fatalf("insert sales: %v", err)
	}
	_, err = db.Exec(`INSERT INTO inventory_items (id, purchase_cost) VALUES ('item-today', 200), ('item-taxed', 30), ('item-yesterday', 30);
		INSERT INTO products (id, purchase_price, cost_price) VALUES ('product-today', 25, 25), ('product-taxed', 30, 30), ('product-yesterday', 35, 35);
		INSERT INTO sale_items (id, sale_id, inventory_item_id, product_id, quantity, unit_cost, total_amount, tax_amount) VALUES
			('line-today', 'today', 'item-today', 'product-today', 2, 20, 40, 0),
			('line-taxed', 'today-taxed', 'item-taxed', 'product-taxed', 1, 30, 115, 15),
			('line-yesterday', 'yesterday', 'item-yesterday', 'product-yesterday', 2, 30, 60, 0)`)
	if err != nil {
		t.Fatalf("insert sale items: %v", err)
	}
	_, err = db.Exec(`INSERT INTO expenses (id, amount, status, expense_date) VALUES
		('expense-today', 10, 'approved', ?), ('expense-yesterday', 500, 'approved', ?)`, today, yesterday)
	if err != nil {
		t.Fatalf("insert expenses: %v", err)
	}
	_, err = db.Exec(`INSERT INTO returns (id, return_date, status, total_refund_amount) VALUES
		('return-today', ?, 'COMPLETED', 145)`, today)
	if err != nil {
		t.Fatalf("insert returns: %v", err)
	}
	_, err = db.Exec(`INSERT INTO return_items (id, return_id, sale_item_id, quantity_returned, total_refund_amount, original_cost) VALUES
		('return-item-today', 'return-today', 'line-today', 1, 30, 20),
		('return-item-taxed-today', 'return-today', 'line-taxed', 1, 115, 30);
		INSERT INTO payments (id, customer_id, amount, created_at) VALUES
			('debt-today', 'customer-1', 50, ?), ('debt-yesterday', 'customer-1', 500, ?)`, todayTimestamp, yesterdayTimestamp)
	if err != nil {
		t.Fatalf("insert return items and payments: %v", err)
	}
	_, err = db.Exec(`INSERT INTO customer_payments (id, customer_id, amount, payment_date) VALUES ('legacy-copy', 'customer-1', 50, ?)`, today)
	if err != nil {
		t.Fatalf("insert legacy payment fixture: %v", err)
	}
	metrics, err := fetchTodayMetrics(context.Background(), sqlx.NewDb(db, "sqlite"), now)
	if err != nil {
		t.Fatalf("fetchTodayMetrics: %v", err)
	}
	if metrics.Sales != 70 {
		t.Fatalf("today sales = %v, want 70 after the completed return", metrics.Sales)
	}
	// 200 revenue - (2*20 + 30) cost - 10 approved expense - (30 + 100) tax-exclusive refunds + (20 + 30) returned cost = 40.
	if metrics.Profit != 40 {
		t.Fatalf("today profit = %v, want 40", metrics.Profit)
	}
	if metrics.DebtCollected != 50 {
		t.Fatalf("today debt collected = %v, want 50", metrics.DebtCollected)
	}
	if metrics.Collected != 265 {
		t.Fatalf("today collected = %v, want 265 including both cash sales and the debt payment", metrics.Collected)
	}
}

func TestFetchTodayMetricsExcludesSalePaymentFromDebtCollections(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL, tax_amount REAL, status TEXT, created_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, inventory_item_id TEXT, product_id TEXT, quantity INTEGER, unit_cost REAL);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, purchase_cost REAL);
		CREATE TABLE products (id TEXT PRIMARY KEY, purchase_price REAL, cost_price REAL);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, status TEXT, expense_date TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, sale_id TEXT, customer_id TEXT, amount REAL, payment_date TEXT, created_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, sale_id TEXT, customer_id TEXT, amount REAL, remaining_amount REAL, due_date TEXT, status TEXT);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	now := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)
	_, err = db.Exec(`
		INSERT INTO sales (id, total_amount, tax_amount, status, created_at) VALUES ('sale-1', 100, 0, 'completed', ?);
		INSERT INTO debts (id, sale_id, customer_id, amount, remaining_amount, due_date, status) VALUES ('debt-1', 'sale-1', 'customer-1', 100, 0, ?, 'paid');
		INSERT INTO payments (id, sale_id, customer_id, amount, payment_date, created_at) VALUES ('payment-1', 'sale-1', 'customer-1', 100, NULL, ?);
	`, now.Format(time.RFC3339), now.AddDate(0, 0, 30).Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert fixtures: %v", err)
	}

	metrics, err := fetchTodayMetrics(context.Background(), sqlx.NewDb(db, "sqlite"), now)
	if err != nil {
		t.Fatalf("fetchTodayMetrics: %v", err)
	}
	if metrics.DebtCollected != 0 {
		t.Fatalf("today debt collected = %v, want 0 for a sale-linked payment", metrics.DebtCollected)
	}
}

func TestCachedDashboardStatsUseDateScopedTodayProfit(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, is_active INTEGER, deleted_at TEXT, min_stock_level INTEGER, purchase_price REAL, cost_price REAL);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL);
		CREATE TABLE suppliers (id TEXT PRIMARY KEY);
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL, status TEXT, created_at TEXT);
		CREATE TABLE purchases (id TEXT PRIMARY KEY, total_amount REAL, status TEXT, created_at TEXT);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, status TEXT, expense_date TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, status TEXT, refund_amount REAL);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, remaining_amount REAL, due_date TEXT, status TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, condition TEXT, purchase_cost REAL, selling_price REAL);
		CREATE TABLE inventory (product_id TEXT, quantity INTEGER, reserved_quantity INTEGER);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, inventory_item_id TEXT, product_id TEXT, quantity INTEGER);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	now := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)
	today, err := accounting.StoreDate(now)
	if err != nil {
		t.Fatal(err)
	}
	yesterday, err := accounting.StoreDate(now.AddDate(0, 0, -1))
	if err != nil {
		t.Fatal(err)
	}
	todayTimestamp := today + "T12:00:00Z"
	yesterdayTimestamp := yesterday + "T12:00:00Z"
	_, err = db.Exec(`
		INSERT INTO products (id, name, is_active, min_stock_level, purchase_price, cost_price) VALUES ('product', 'Widget', 1, 0, 25, 25);
		INSERT INTO sales (id, total_amount, status, created_at) VALUES ('today-sale', 100, 'completed', ?), ('old-sale', 1000, 'completed', ?);
		INSERT INTO purchases (id, total_amount, status, created_at) VALUES ('old-purchase', 500, 'received', ?);
		INSERT INTO inventory_items (id, product_id, status, condition, purchase_cost, selling_price) VALUES ('item', 'product', 'SOLD', 'NEW', 20, 100);
		INSERT INTO sale_items (id, sale_id, inventory_item_id, product_id, quantity) VALUES ('line', 'today-sale', 'item', 'product', 2);
		INSERT INTO expenses (id, amount, status, expense_date) VALUES ('today-expense', 10, 'approved', ?);
	`, todayTimestamp, yesterdayTimestamp, yesterdayTimestamp, today)
	if err != nil {
		t.Fatalf("insert fixtures: %v", err)
	}

	service := NewCachedService(sqlx.NewDb(db, "sqlite"))
	stats, err := service.fetchFromDatabaseAt(context.Background(), now)
	if err != nil {
		t.Fatalf("fetch dashboard stats: %v", err)
	}
	if stats.TodaySales != 100 {
		t.Fatalf("today sales = %v, want 100", stats.TodaySales)
	}
	if stats.TodayProfit != 50 {
		t.Fatalf("today profit = %v, want 50", stats.TodayProfit)
	}
}

func TestTodayMetricsFallsBackToProductCostWhenStoredSaleCostIsZero(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	now := time.Date(2026, time.September, 26, 12, 0, 0, 0, time.UTC)
	storeDate, err := accounting.StoreDate(now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY, total_amount REAL, tax_amount REAL, cost_amount REAL,
			status TEXT, sale_date TEXT, created_at TEXT
		);
		CREATE TABLE products (id TEXT PRIMARY KEY, cost_price REAL);
		CREATE TABLE sale_items (
			id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT,
			quantity REAL, unit_cost REAL
		);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, purchase_cost REAL);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, status TEXT, expense_date TEXT);
		INSERT INTO products (id, cost_price) VALUES ('product', 20);
		INSERT INTO sales (id, total_amount, tax_amount, cost_amount, status, sale_date, created_at)
		VALUES ('sale', 25, 0, 0, 'completed', ?, ?);
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_cost)
		VALUES ('line', 'sale', 'product', 1, 0);
	`, storeDate, now.Format(time.RFC3339)); err != nil {
		t.Fatalf("seed dashboard profit fixture: %v", err)
	}

	metrics, err := fetchTodayMetrics(context.Background(), sqlx.NewDb(db, "sqlite"), now)
	if err != nil {
		t.Fatalf("fetch today's metrics: %v", err)
	}
	if metrics.Sales != 25 || metrics.Profit != 5 {
		t.Fatalf("today sales/profit = %.2f/%.2f, want 25/5 with product cost fallback", metrics.Sales, metrics.Profit)
	}
}
