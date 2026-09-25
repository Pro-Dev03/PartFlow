package reports

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestProfitReportIncludesApprovedExpensesFromStoreDay(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_date TEXT, total_amount REAL, tax_amount REAL, status TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, quantity INTEGER, unit_cost REAL);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, expense_date TEXT, status TEXT);
		INSERT INTO sales (id, sale_date, total_amount, tax_amount, status)
		VALUES ('sale-1', '2026-09-16T10:00:00Z', 2348, 0, 'completed');
		INSERT INTO sale_items (id, sale_id, quantity, unit_cost)
		VALUES ('line-1', 'sale-1', 1, 800);
		INSERT INTO expenses (id, amount, expense_date, status) VALUES
			('expense-50', 50, '2026-09-16T12:00:00Z', 'approved'),
			('expense-1', 1, '2026-09-16T23:59:00+03:00', 'approved'),
			('expense-archived', 2, '2026-09-16T14:00:00Z', 'archived'),
			('expense-pending', 100, '2026-09-16T13:00:00Z', 'pending');
	`)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Date(2026, 9, 16, 0, 0, 0, 0, time.FixedZone("Asia/Jerusalem", 3*60*60))
	end := start.AddDate(0, 0, 1)
	report, err := NewRepository(db).GetProfitsData(context.Background(), start, end)
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalExpenses != 53 || report.NetProfit != 1495 {
		t.Fatalf("profit report = expenses %v, net profit %v; want 53 and 1495", report.TotalExpenses, report.NetProfit)
	}
}

func TestProfitReportReturnsQueryErrorWhenSalesTableIsMissing(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	start := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	_, err = NewRepository(db).GetProfitsData(context.Background(), start, start.AddDate(0, 0, 1))
	if err == nil {
		t.Fatal("expected missing sales table query error")
	}
}
