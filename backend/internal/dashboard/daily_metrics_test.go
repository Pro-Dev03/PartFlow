package dashboard

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
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
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL, tax_amount REAL, status TEXT, created_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, inventory_item_id TEXT, product_id TEXT, quantity INTEGER, unit_cost REAL, total_amount REAL, tax_amount REAL);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, purchase_cost REAL);
		CREATE TABLE products (id TEXT PRIMARY KEY, purchase_price REAL, cost_price REAL);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, status TEXT, expense_date TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, return_date TEXT, status TEXT, total_refund_amount REAL);
		CREATE TABLE return_items (id TEXT PRIMARY KEY, return_id TEXT, sale_item_id TEXT, quantity_returned INTEGER, original_cost REAL);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	now := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	todayTimestamp := today + "T12:00:00Z"
	yesterdayTimestamp := yesterday + "T12:00:00Z"

	_, err = db.Exec(`
		INSERT INTO sales (id, total_amount, tax_amount, status, created_at) VALUES
			('today', 100, 0, 'completed', ?),
			('yesterday', 1000, 0, 'completed', ?),
			('cancelled-today', 500, 0, 'cancelled', ?);
		INSERT INTO inventory_items (id, purchase_cost) VALUES ('item-today', 20), ('item-yesterday', 30);
		INSERT INTO products (id, purchase_price, cost_price) VALUES ('product-today', 25, 25), ('product-yesterday', 35, 35);
		INSERT INTO sale_items (id, sale_id, inventory_item_id, product_id, quantity, unit_cost, total_amount, tax_amount) VALUES
			('line-today', 'today', 'item-today', 'product-today', 2, 20, 40, 0),
			('line-yesterday', 'yesterday', 'item-yesterday', 'product-yesterday', 2, 30, 60, 0);
		INSERT INTO expenses (id, amount, status, expense_date) VALUES
			('expense-today', 10, 'approved', ?),
			('expense-yesterday', 500, 'approved', ?);
		INSERT INTO returns (id, return_date, status, total_refund_amount) VALUES
			('return-today', ?, 'COMPLETED', 30);
		INSERT INTO return_items (id, return_id, sale_item_id, quantity_returned, original_cost) VALUES
			('return-item-today', 'return-today', 'line-today', 1, 20);
	`, todayTimestamp, yesterdayTimestamp, todayTimestamp, today, yesterday, today)
	if err != nil {
		t.Fatalf("insert fixtures: %v", err)
	}

	metrics, err := fetchTodayMetrics(context.Background(), sqlx.NewDb(db, "sqlite"), now)
	if err != nil {
		t.Fatalf("fetchTodayMetrics: %v", err)
	}
	if metrics.Sales != 100 {
		t.Fatalf("today sales = %v, want 100", metrics.Sales)
	}
	// 100 revenue - (2 * 20 cost) - 10 approved expense - 30 return refund + (1 * 20 returned cost) = 40.
	if metrics.Profit != 40 {
		t.Fatalf("today profit = %v, want 40", metrics.Profit)
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
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, inventory_item_id TEXT, product_id TEXT, quantity INTEGER);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	now := time.Now().UTC()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
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
	stats, err := service.fetchFromDatabase(context.Background())
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
