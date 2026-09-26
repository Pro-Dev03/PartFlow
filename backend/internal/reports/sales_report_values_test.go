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

func TestReportsPreferCapturedSaleCostAndRecoverZeroLegacyCosts(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY, sale_date TEXT, total_amount REAL, tax_amount REAL,
			payment_method TEXT, status TEXT, cost_amount REAL
		);
		CREATE TABLE sale_items (
			id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, quantity INTEGER,
			unit_cost REAL, total_amount REAL, tax_amount REAL
		);
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, category_id TEXT, cost_price REAL);
		CREATE TABLE categories (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, expense_date TEXT, status TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO products (id, name, cost_price) VALUES ('00000000-0000-0000-0000-000000000001', 'Part 1', 20), ('00000000-0000-0000-0000-000000000002', 'Part 2', 10);
		INSERT INTO sales (id, sale_date, total_amount, tax_amount, payment_method, status, cost_amount) VALUES
			('00000000-0000-0000-0000-000000000111', '2026-09-16T10:00:00Z', 100, 0, 'cash', 'completed', 60),
			('00000000-0000-0000-0000-000000000112', '2026-09-16T11:00:00Z', 200, 0, 'cash', 'completed', 120),
			('00000000-0000-0000-0000-000000000113', '2026-09-16T12:00:00Z', 25, 0, 'cash', 'completed', 0);
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_cost, total_amount, tax_amount) VALUES
			('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000111', '00000000-0000-0000-0000-000000000001', 1, 0, 100, 0),
			('00000000-0000-0000-0000-000000000102', '00000000-0000-0000-0000-000000000112', '00000000-0000-0000-0000-000000000001', 1, 0, 100, 0),
			('00000000-0000-0000-0000-000000000103', '00000000-0000-0000-0000-000000000112', '00000000-0000-0000-0000-000000000002', 1, 0, 100, 0),
			('00000000-0000-0000-0000-000000000104', '00000000-0000-0000-0000-000000000113', '00000000-0000-0000-0000-000000000001', 1, 0, 25, 0);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	start := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	salesReport, err := repo.GetSalesData(context.Background(), start, end)
	if err != nil {
		t.Fatal(err)
	}
	if salesReport.TotalRevenue != 325 || salesReport.TotalCOGS != 200 || salesReport.GrossProfit != 125 {
		t.Fatalf("sales report failed cost reconciliation: revenue=%v cogs=%v gross=%v", salesReport.TotalRevenue, salesReport.TotalCOGS, salesReport.GrossProfit)
	}

	profitReport, err := repo.GetProfitsData(context.Background(), start, end)
	if err != nil {
		t.Fatal(err)
	}
	if profitReport.TotalRevenue != 325 || profitReport.TotalCOGS != 200 || profitReport.NetProfit != 125 {
		t.Fatalf("profit report failed cost reconciliation: revenue=%v cogs=%v net=%v", profitReport.TotalRevenue, profitReport.TotalCOGS, profitReport.NetProfit)
	}
	if len(profitReport.ByDay) != 1 || profitReport.ByDay[0].COGS != 200 || profitReport.ByDay[0].NetProfit != 125 {
		t.Fatalf("daily profit trend failed cost reconciliation: %#v", profitReport.ByDay)
	}
	if len(profitReport.ByMonth) != 1 || profitReport.ByMonth[0].COGS != 200 || profitReport.ByMonth[0].NetProfit != 125 {
		t.Fatalf("monthly profit trend failed cost reconciliation: %#v", profitReport.ByMonth)
	}
}
