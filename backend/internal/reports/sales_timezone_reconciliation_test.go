package reports

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
)

func TestSalesReportBreakdownsUseStoreCalendarDate(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() {
		if err := accounting.ConfigureStoreTimezone(previousTimezone); err != nil {
			t.Errorf("restore store timezone: %v", err)
		}
	})
	if err := accounting.ConfigureStoreTimezone("America/Los_Angeles"); err != nil {
		t.Fatal(err)
	}

	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY, sale_date TEXT, created_at TEXT, total_amount REAL,
			tax_amount REAL, subtotal REAL, discount_amount REAL, payment_method TEXT,
			status TEXT, cost_amount REAL, paid_amount REAL, cash_received REAL, change_amount REAL
		);
		CREATE TABLE sale_items (
			id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, quantity INTEGER,
			unit_cost REAL, total_amount REAL, tax_amount REAL
		);
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, category_id TEXT, cost_price REAL, is_active INTEGER);
		CREATE TABLE categories (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT);
		CREATE TABLE accounting_returns (
			id TEXT PRIMARY KEY, sale_id TEXT, refund_date TEXT, return_date TEXT, created_at TEXT,
			status TEXT, total_refund_amount REAL, reference_number TEXT, reason TEXT
		);
		CREATE TABLE accounting_return_items (
			id TEXT PRIMARY KEY, return_id TEXT, product_id TEXT, sale_item_id TEXT, inventory_item_id TEXT,
			quantity_returned INTEGER, total_refund_amount REAL, original_cost REAL
		);
		INSERT INTO products (id, name, category_id, cost_price, is_active)
		VALUES ('00000000-0000-0000-0000-000000000001', 'Part One', NULL, 50, 1);
		INSERT INTO sales (id, sale_date, created_at, total_amount, tax_amount, subtotal, discount_amount,
			payment_method, status, cost_amount, paid_amount, cash_received, change_amount)
		VALUES ('sale-1', '2026-09-17T06:30:00Z', '2026-09-17T06:30:00Z', 110, 10, 100, 0,
			'cash', 'completed', 50, 110, 110, 0);
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_cost, total_amount, tax_amount)
		VALUES ('line-1', 'sale-1', '00000000-0000-0000-0000-000000000001', 1, 50, 110, 10);
	`)
	if err != nil {
		t.Fatal(err)
	}

	start, end, err := accounting.StoreDateBounds("2026-09-16")
	if err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	ctx := context.Background()

	sales, err := repo.GetSalesData(ctx, start, end)
	if err != nil {
		t.Fatalf("GetSalesData: %v", err)
	}
	if sales.TotalSales != 1 || !closeReportAmount(sales.TotalRevenue, 100) {
		t.Fatalf("sales totals = count %d, revenue %.2f; want 1 and 100", sales.TotalSales, sales.TotalRevenue)
	}
	if len(sales.ByDay) != 1 || sales.ByDay[0].Date.Format("2006-01-02") != "2026-09-16" {
		t.Fatalf("sales daily breakdown = %#v; want sale on store date 2026-09-16", sales.ByDay)
	}
	if !closeReportAmount(sales.ByPaymentMethod["cash"], 100) {
		t.Fatalf("sales payment method breakdown = %#v; want cash 100", sales.ByPaymentMethod)
	}

	netSales, err := repo.GetNetSalesData(ctx, start, end)
	if err != nil {
		t.Fatalf("GetNetSalesData: %v", err)
	}
	if netSales.GrossSales != 1 || !closeReportAmount(netSales.GrossRevenue, 100) ||
		len(netSales.ByDay) != 1 || netSales.ByDay[0].Date.Format("2006-01-02") != "2026-09-16" ||
		!closeReportAmount(netSales.ByPaymentMethod["cash"], 100) {
		t.Fatalf("net sales does not reconcile on store day: %#v", netSales)
	}

	tax, err := repo.GetTaxData(ctx, start, end)
	if err != nil {
		t.Fatalf("GetTaxData: %v", err)
	}
	if !closeReportAmount(tax.SalesTotal, 110) || !closeReportAmount(tax.TaxCollected, 10) ||
		!closeReportAmount(tax.NetTaxCollected, 10) {
		t.Fatalf("tax totals do not include the store-day sale: %#v", tax)
	}
}
