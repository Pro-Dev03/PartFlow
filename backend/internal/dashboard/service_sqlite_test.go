package dashboard

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestGetDashboardStatsWorksWithSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	schema := `
        CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, is_active INTEGER, min_stock_level INTEGER);
        CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL);
        CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT);
        CREATE TABLE sales (id TEXT PRIMARY KEY, status TEXT, total_amount REAL);
        CREATE TABLE purchases (id TEXT PRIMARY KEY, status TEXT, total_amount REAL);
        CREATE TABLE expenses (id TEXT PRIMARY KEY, status TEXT, amount REAL);
        CREATE TABLE returns (id TEXT PRIMARY KEY, status TEXT, refund_amount REAL);
        CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, remaining_amount REAL, due_date TEXT, status TEXT);
        CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT);
    `
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO products (id, name, is_active, min_stock_level) VALUES ('p1','Widget',1,5)`); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO customers (id, name, current_balance) VALUES ('c1','Alice',120)`); err != nil {
		t.Fatalf("insert customer: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, status, total_amount) VALUES ('s1','completed',250)`); err != nil {
		t.Fatalf("insert sale: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO purchases (id, status, total_amount) VALUES ('p1','received',80)`); err != nil {
		t.Fatalf("insert purchase: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO expenses (id, status, amount) VALUES ('e1','approved',20)`); err != nil {
		t.Fatalf("insert expense: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO returns (id, status, refund_amount) VALUES ('r1','completed',10)`); err != nil {
		t.Fatalf("insert return: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, remaining_amount, due_date, status) VALUES ('d1','c1',100,'2024-01-01','pending')`); err != nil {
		t.Fatalf("insert debt: %v", err)
	}

	svc := NewService(sqlx.NewDb(db, "sqlite"))
	stats, err := svc.GetDashboardStats(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardStats should work with SQLite: %v", err)
	}
	if stats.TotalProducts != 1 {
		t.Fatalf("expected 1 product, got %d", stats.TotalProducts)
	}
	if stats.TotalSales != 250 {
		t.Fatalf("expected total sales 250, got %v", stats.TotalSales)
	}
}
