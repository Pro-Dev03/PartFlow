package accounting

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestAccountingExpensesForPeriodSupportsLegacyGoTimestamps(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, expense_date TEXT, status TEXT);
		INSERT INTO expenses (id, amount, expense_date, status) VALUES
			('approved-50', 50, '2026-09-16 12:00:00 +0000 UTC', 'approved'),
			('approved-1', 1, '2026-09-16 23:59:00 +0300 EEST', 'approved'),
			('pending', 100, '2026-09-16 13:00:00 +0000 UTC', 'pending'),
			('next-day', 7, '2026-09-17 00:00:00 +0300 EEST', 'approved');
	`); err != nil {
		t.Fatal(err)
	}

	location, err := StoreLocation()
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 9, 16, 12, 0, 0, 0, location)
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, location)
	total, err := AccountingExpensesForPeriod(context.Background(), sqlx.NewDb(db, "sqlite"), start, start.AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if total != 51 {
		t.Fatalf("expenses = %v, want 51", total)
	}
}

func TestAccountingExpensesForPeriodReturnsQueryErrors(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	start := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	_, err = AccountingExpensesForPeriod(context.Background(), db, start, start.AddDate(0, 0, 1))
	if err == nil {
		t.Fatal("expected missing expenses table error")
	}
}
