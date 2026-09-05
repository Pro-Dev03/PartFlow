package reports

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestGetTaxDataExcludesTaxExemptSales(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY,
			sale_date TEXT,
			subtotal REAL,
			discount_amount REAL,
			tax_amount REAL,
			total_amount REAL,
			status TEXT
		);
		CREATE TABLE returns (
			id TEXT PRIMARY KEY,
			return_date TEXT,
			total_refund_amount REAL,
			status TEXT
		);
	`)
	if err != nil {
		t.Fatalf("create tables: %v", err)
	}

	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)

	_, err = db.Exec(`
		INSERT INTO sales (id, sale_date, subtotal, discount_amount, tax_amount, total_amount, status)
		VALUES
			('taxed-1', '2026-09-01T10:00:00Z', 100, 0, 15, 115, 'completed'),
			('tax-free-1', '2026-09-01T12:00:00Z', 200, 0, 0, 200, 'completed'),
			('taxed-2', '2026-09-02T09:00:00Z', 50, 10, 6, 46, 'completed');
		INSERT INTO returns (id, return_date, total_refund_amount, status)
		VALUES ('r1', '2026-09-02T11:00:00Z', 10, 'COMPLETED');
	`)
	if err != nil {
		t.Fatalf("insert sales: %v", err)
	}

	repo := NewRepository(db)
	report, err := repo.GetTaxData(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetTaxData failed: %v", err)
	}

	if got, want := report.GrossSales, 150.0; got != want {
		t.Fatalf("GrossSales = %v, want %v", got, want)
	}
	if got, want := report.Discounts, 10.0; got != want {
		t.Fatalf("Discounts = %v, want %v", got, want)
	}
	if got, want := report.TaxableSales, 140.0; got != want {
		t.Fatalf("TaxableSales = %v, want %v", got, want)
	}
	if got, want := report.TaxCollected, 21.0; got != want {
		t.Fatalf("TaxCollected = %v, want %v", got, want)
	}
	if got, want := report.SalesTotal, 161.0; got != want {
		t.Fatalf("SalesTotal = %v, want %v", got, want)
	}
	if got, want := report.ReturnsTotal, 10.0; got != want {
		t.Fatalf("ReturnsTotal = %v, want %v", got, want)
	}
	if got, want := report.NetSalesTotal, 151.0; got != want {
		t.Fatalf("NetSalesTotal = %v, want %v", got, want)
	}
}
