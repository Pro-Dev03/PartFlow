package reports

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestGetSalesDataUsesHistoricalCostAndExcludesTaxFromProfit(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY, sale_date TEXT, total_amount REAL, tax_amount REAL,
			payment_method TEXT, status TEXT
		);
		CREATE TABLE sale_items (
			id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, quantity INTEGER,
			unit_cost REAL, total_amount REAL, tax_amount REAL
		);
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO products (id, name) VALUES ('00000000-0000-0000-0000-000000000001', 'Part');
		INSERT INTO sales (id, sale_date, total_amount, tax_amount, payment_method, status) VALUES
			('00000000-0000-0000-0000-000000000101', '2026-09-02T10:00:00Z', 1150, 150, 'cash', 'completed'),
			('00000000-0000-0000-0000-000000000102', '2026-09-02T11:00:00Z', 1000, 0, 'cash', 'completed'),
			('00000000-0000-0000-0000-000000000103', '2026-09-02T12:00:00Z', 1000, 0, 'cash', 'completed');
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_cost, total_amount, tax_amount) VALUES
			('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000001', 1, 400, 1150, 150),
			('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000102', '00000000-0000-0000-0000-000000000001', 1, 1000, 1000, 0),
			('00000000-0000-0000-0000-000000000203', '00000000-0000-0000-0000-000000000103', '00000000-0000-0000-0000-000000000001', 1, 1000, 1000, 0);
	`)
	if err != nil {
		t.Fatal(err)
	}

	report, err := NewRepository(db).GetSalesData(context.Background(),
		time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	if report.TotalSales != 3 || report.TotalItemsSold != 3 {
		t.Fatalf("unexpected counts: sales=%d items=%d", report.TotalSales, report.TotalItemsSold)
	}
	if report.TotalRevenue != 3000 || report.TotalCOGS != 2400 || report.GrossProfit != 600 {
		t.Fatalf("unexpected financial values: revenue=%v cost=%v profit=%v", report.TotalRevenue, report.TotalCOGS, report.GrossProfit)
	}
	if report.CashRevenue != 3000 {
		t.Fatalf("cash revenue = %v, want 3000", report.CashRevenue)
	}
}
