package api

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/aggregations"
	_ "modernc.org/sqlite"
)

func TestDailySalesSummaryQueryUsesSQLiteSafeSyntax(t *testing.T) {
	query := buildDailySalesSummaryQuery()
	if strings.Contains(query, "::date") {
		t.Fatalf("daily sales query should not use PostgreSQL-only cast syntax: %s", query)
	}
	if strings.Contains(query, "NOW()") {
		t.Fatalf("daily sales query should not use PostgreSQL NOW(): %s", query)
	}
	if !strings.Contains(query, "LEFT JOIN daily_sales_summary s ON s.date = dates.summary_date") {
		t.Fatalf("daily sales query should join the real daily summary table by date: %s", query)
	}
}

func TestMonthlySalesSummaryQueryUsesSQLiteSafeSyntax(t *testing.T) {
	query := buildMonthlySalesSummaryQuery()
	if strings.Contains(query, "::int") {
		t.Fatalf("monthly sales query should not use PostgreSQL-only cast syntax: %s", query)
	}
	if strings.Contains(query, "NOW()") {
		t.Fatalf("monthly sales query should not use PostgreSQL NOW(): %s", query)
	}
	if !strings.Contains(query, "LEFT JOIN monthly_sales_summary s ON s.year = dates.summary_year AND s.month = dates.summary_month") {
		t.Fatalf("monthly sales query should join the real monthly summary table by year/month: %s", query)
	}
}

func TestAggregationQueriesExecuteOnSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE daily_sales_summary (
		date TEXT PRIMARY KEY,
		total_sales INTEGER DEFAULT 0,
		total_revenue REAL DEFAULT 0,
		total_profit REAL DEFAULT 0,
		total_customers INTEGER DEFAULT 0,
		average_order_value REAL DEFAULT 0,
		total_items_sold INTEGER DEFAULT 0,
		cash_sales REAL DEFAULT 0,
		card_sales REAL DEFAULT 0,
		debt_sales REAL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create daily_sales_summary: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE monthly_sales_summary (
		year INTEGER NOT NULL,
		month INTEGER NOT NULL,
		total_sales INTEGER DEFAULT 0,
		total_revenue REAL DEFAULT 0,
		total_profit REAL DEFAULT 0,
		total_customers INTEGER DEFAULT 0,
		average_order_value REAL DEFAULT 0,
		total_items_sold INTEGER DEFAULT 0,
		cash_sales REAL DEFAULT 0,
		card_sales REAL DEFAULT 0,
		debt_sales REAL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (year, month)
	)`); err != nil {
		t.Fatalf("create monthly_sales_summary: %v", err)
	}

	if err := db.QueryRow(buildDailySalesSummaryQuery(), "2026-08-30", "2026-08-30").Scan(new(string), new(int), new(float64), new(float64), new(int), new(float64), new(int), new(float64), new(float64), new(float64), new(string)); err != nil {
		t.Fatalf("daily aggregation query failed on sqlite: %v", err)
	}
	if err := db.QueryRow(buildMonthlySalesSummaryQuery(), 2026, 8, 2026, 8).Scan(new(int), new(int), new(int), new(float64), new(float64), new(int), new(float64), new(int), new(float64), new(float64), new(float64), new(string)); err != nil {
		t.Fatalf("monthly aggregation query failed on sqlite: %v", err)
	}
}

func TestDailySalesSummaryStructScansOnSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE daily_sales_summary (
		date TEXT PRIMARY KEY,
		total_sales INTEGER DEFAULT 0,
		total_revenue REAL DEFAULT 0,
		total_profit REAL DEFAULT 0,
		total_customers INTEGER DEFAULT 0,
		average_order_value REAL DEFAULT 0,
		total_items_sold INTEGER DEFAULT 0,
		cash_sales REAL DEFAULT 0,
		card_sales REAL DEFAULT 0,
		debt_sales REAL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create daily_sales_summary: %v", err)
	}

	xdb := sqlx.NewDb(db, "sqlite")
	var summary aggregations.DailySalesSummary
	err = xdb.Get(&summary, buildDailySalesSummaryQuery(), "2026-08-30", "2026-08-30")
	if err != nil {
		t.Fatalf("daily sales summary should scan on sqlite: %v", err)
	}
	if summary.Date == "" {
		t.Fatalf("daily sales summary date should be populated")
	}
}

func TestRefreshSQLiteSummariesUsesOperationalData(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	statements := []string{
		`CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, purchase_price REAL DEFAULT 0, min_stock_level INTEGER DEFAULT 0, is_active INTEGER DEFAULT 1)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, purchase_cost REAL DEFAULT 0, status TEXT, created_at TEXT)`,
		`CREATE TABLE sales (id TEXT PRIMARY KEY, customer_id TEXT, total_amount REAL, payment_method TEXT, status TEXT, created_at TEXT)`,
		`CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER)`,
		`CREATE TABLE returns (id TEXT PRIMARY KEY, return_date TEXT, created_at TEXT, status TEXT)`,
		`CREATE TABLE return_items (id TEXT PRIMARY KEY, return_id TEXT, quantity INTEGER)`,
		`CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, movement_type TEXT, created_at TEXT)`,
		`CREATE TABLE debts (id TEXT PRIMARY KEY, amount REAL, remaining_amount REAL, due_date TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE payments (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, created_at TEXT)`,
		`CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, expense_date TEXT, status TEXT)`,
		`CREATE TABLE daily_sales_summary (date TEXT PRIMARY KEY, total_sales INTEGER, total_revenue REAL, total_profit REAL, total_customers INTEGER, average_order_value REAL, total_items_sold INTEGER, cash_sales REAL, card_sales REAL, debt_sales REAL, updated_at TEXT)`,
		`CREATE TABLE monthly_sales_summary (year INTEGER, month INTEGER, total_sales INTEGER, total_revenue REAL, total_profit REAL, total_customers INTEGER, average_order_value REAL, total_items_sold INTEGER, cash_sales REAL, card_sales REAL, debt_sales REAL, updated_at TEXT, PRIMARY KEY(year, month))`,
		`CREATE TABLE daily_inventory_summary (date TEXT PRIMARY KEY, total_items INTEGER, total_value REAL, low_stock_count INTEGER, out_of_stock_count INTEGER, new_items_added INTEGER, items_sold INTEGER, items_returned INTEGER, items_damaged INTEGER, updated_at TEXT)`,
		`CREATE TABLE monthly_inventory_summary (year INTEGER, month INTEGER, total_items INTEGER, total_value REAL, low_stock_count INTEGER, out_of_stock_count INTEGER, new_items_added INTEGER, items_sold INTEGER, items_returned INTEGER, items_damaged INTEGER, updated_at TEXT, PRIMARY KEY(year, month))`,
		`CREATE TABLE daily_debt_summary (date TEXT PRIMARY KEY, total_debt REAL, new_debt REAL, payments_received REAL, overdue_debt REAL, overdue_count INTEGER, paid_debt REAL, updated_at TEXT)`,
		`CREATE TABLE monthly_debt_summary (year INTEGER, month INTEGER, total_debt REAL, new_debt REAL, payments_received REAL, overdue_debt REAL, overdue_count INTEGER, paid_debt REAL, updated_at TEXT, PRIMARY KEY(year, month))`,
		`CREATE TABLE daily_profit_summary (date TEXT PRIMARY KEY, gross_profit REAL, net_profit REAL, total_revenue REAL, total_cost REAL, profit_margin REAL, updated_at TEXT)`,
		`CREATE TABLE monthly_profit_summary (year INTEGER, month INTEGER, gross_profit REAL, net_profit REAL, total_revenue REAL, total_cost REAL, profit_margin REAL, updated_at TEXT, PRIMARY KEY(year, month))`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("create test table: %v", err)
		}
	}
	if _, err := db.Exec(`INSERT INTO products (id, name, purchase_price, min_stock_level, is_active) VALUES ('p1', 'Brake pad', 20, 2, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, purchase_cost, status, created_at) VALUES ('i1', 'p1', 20, 'AVAILABLE', '2026-08-30 09:00:00')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, customer_id, total_amount, payment_method, status, created_at) VALUES ('s1', 'c1', 100, 'cash', 'completed', '2026-08-30 10:00:00')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id, sale_id, product_id, inventory_item_id, quantity) VALUES ('si1', 's1', 'p1', 'i1', 2)`); err != nil {
		t.Fatal(err)
	}

	xdb := sqlx.NewDb(db, "sqlite")
	h := NewAggregationHandler(xdb)
	day := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	if err := h.refreshSQLiteDay(context.Background(), day); err != nil {
		t.Fatalf("refresh summaries: %v", err)
	}

	var salesRow struct {
		TotalSales     int     `db:"total_sales"`
		TotalRevenue   float64 `db:"total_revenue"`
		TotalProfit    float64 `db:"total_profit"`
		TotalItemsSold int     `db:"total_items_sold"`
	}
	if err := xdb.Get(&salesRow, `SELECT total_sales, total_revenue, total_profit, total_items_sold FROM daily_sales_summary WHERE date = ?`, "2026-08-30"); err != nil {
		t.Fatalf("read daily sales summary: %v", err)
	}
	if salesRow.TotalSales != 1 || salesRow.TotalRevenue != 100 || salesRow.TotalProfit != 60 || salesRow.TotalItemsSold != 2 {
		t.Fatalf("unexpected sales summary: %+v", salesRow)
	}
	row, err := h.getDailyInventorySummary(context.Background(), "2026-08-30")
	if err != nil {
		t.Fatalf("read daily inventory summary: %v", err)
	}
	if row.TotalItems != 1 || row.TotalValue != 20 || row.LowStockCount != 1 || row.ItemsSold != 2 {
		t.Fatalf("unexpected inventory summary: %+v", row)
	}
}
