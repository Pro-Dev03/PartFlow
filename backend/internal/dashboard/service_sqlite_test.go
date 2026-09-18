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
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, is_active INTEGER, deleted_at TEXT, min_stock_level INTEGER);
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL);
        CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, customer_id TEXT, status TEXT, total_amount REAL);
        CREATE TABLE purchases (id TEXT PRIMARY KEY, status TEXT, total_amount REAL);
        CREATE TABLE expenses (id TEXT PRIMARY KEY, status TEXT, amount REAL);
        CREATE TABLE returns (id TEXT PRIMARY KEY, status TEXT, refund_amount REAL);
        CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, remaining_amount REAL, due_date TEXT, status TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, condition TEXT);
		CREATE TABLE inventory (product_id TEXT, quantity INTEGER);
    `
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO products (id, name, is_active, deleted_at, min_stock_level) VALUES ('p1','Widget',1,NULL,5)`); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO customers (id, name, current_balance) VALUES ('c1','Alice',120)`); err != nil {
		t.Fatalf("insert customer: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO customers (id, name, current_balance) VALUES ('c2','Paid customer',0)`); err != nil {
		t.Fatalf("insert paid customer: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, customer_id, status, total_amount) VALUES ('s1','c1','completed',250)`); err != nil {
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

	cachedStats, err := NewCachedService(sqlx.NewDb(db, "sqlite")).fetchFromDatabase(context.Background())
	if err != nil {
		t.Fatalf("cached dashboard stats should work with SQLite: %v", err)
	}
	if cachedStats.ActiveCustomers != 1 {
		t.Fatalf("expected 1 customer who purchased, got %d", cachedStats.ActiveCustomers)
	}

}

func TestGetLowStockItemsIncludesOutOfStockProducts(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE products (
			id TEXT PRIMARY KEY,
			name TEXT,
			is_active INTEGER,
			deleted_at TEXT,
			min_stock_level INTEGER,
			cost_price REAL,
			selling_price REAL,
			preferred_supplier_id TEXT
		);
		CREATE TABLE inventory_items (
			id TEXT PRIMARY KEY,
			product_id TEXT,
			status TEXT,
			condition TEXT
		);
		CREATE TABLE inventory (product_id TEXT, quantity INTEGER);
		INSERT INTO products (id, name, is_active, deleted_at, min_stock_level, cost_price, selling_price) VALUES
			('p1', 'Out of stock', 1, NULL, 2, 10, 20),
			('p2', 'Used only', 1, NULL, 2, 10, 20),
			('p3', 'Deleted product', 1, '2026-09-16T00:00:00Z', 50, 10, 20);
		INSERT INTO inventory_items (id, product_id, status, condition) VALUES ('i1', 'p2', 'AVAILABLE', 'USED');
	`); err != nil {
		t.Fatalf("create schema and fixture: %v", err)
	}

	items, err := NewCachedService(sqlx.NewDb(db, "sqlite")).GetLowStockItems(context.Background())
	if err != nil {
		t.Fatalf("GetLowStockItems should include zero stock: %v", err)
	}
	if len(items) != 1 || items[0].ID != "p1" || items[0].Quantity != 0 {
		t.Fatalf("expected only the out-of-stock general product, got %+v", items)
	}
}

func TestDeletedProductsAreExcludedFromDashboardStats(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	schema := `
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, is_active INTEGER, deleted_at TEXT, min_stock_level INTEGER);
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL);
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, customer_id TEXT, status TEXT, total_amount REAL);
		CREATE TABLE purchases (id TEXT PRIMARY KEY, status TEXT, total_amount REAL);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, status TEXT, amount REAL);
		CREATE TABLE returns (id TEXT PRIMARY KEY, status TEXT, refund_amount REAL);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, remaining_amount REAL, due_date TEXT, status TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, condition TEXT);
		CREATE TABLE inventory (product_id TEXT, quantity INTEGER);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO products (id, name, is_active, deleted_at, min_stock_level) VALUES
			('active', 'Active Product', 1, NULL, 2),
			('deleted', 'Deleted Product', 1, '2026-09-16T00:00:00Z', 2),
			('debug', 'test1', 1, '2026-09-16T00:00:00Z', 0),
			('debug2', '123123', 1, '2026-09-16T00:00:00Z', 0);
		INSERT INTO inventory (product_id, quantity) VALUES ('active', 1), ('deleted', 0), ('debug', 0), ('debug2', 0);
	`); err != nil {
		t.Fatalf("insert products: %v", err)
	}

	svc := NewService(sqlx.NewDb(db, "sqlite"))
	stats, err := svc.GetDashboardStats(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardStats should work: %v", err)
	}
	if stats.TotalProducts != 1 {
		t.Fatalf("deleted and stale test rows must not be counted; got %d", stats.TotalProducts)
	}
	if stats.LowStockCount != 1 {
		t.Fatalf("deleted rows must not appear as low stock; got %d", stats.LowStockCount)
	}
}
