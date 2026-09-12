package expenses

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestEnsureRecurringExpensesMaterializesDueDatesOnce(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE expense_categories (id TEXT PRIMARY KEY, name TEXT NOT NULL);
		CREATE TABLE expenses (
			id TEXT PRIMARY KEY, title TEXT NOT NULL, category_id TEXT, amount REAL NOT NULL,
			currency TEXT, reference TEXT, notes TEXT, expense_date TEXT NOT NULL,
			status TEXT NOT NULL, is_recurring INTEGER NOT NULL DEFAULT 0, recurring_period TEXT,
			approved_by TEXT, created_by TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
			description TEXT, payment_method TEXT, receipt_url TEXT
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	categoryID := uuid.New()
	templateID := uuid.New()
	stamp := "2026-09-01T12:00:00Z"
	if _, err = db.Exec(`INSERT INTO expense_categories (id, name) VALUES (?, 'Rent')`, categoryID.String()); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		INSERT INTO expenses (id, title, category_id, amount, currency, expense_date, status, is_recurring, recurring_period, created_at, updated_at, description, payment_method)
		VALUES (?, ?, ?, 100, 'ILS', ?, 'approved', 1, 'monthly', ?, ?, 'Monthly rent', 'cash')
	`, templateID.String(), "Monthly rent", categoryID.String(), stamp, stamp, stamp)
	if err != nil {
		t.Fatal(err)
	}

	repository := NewRepository(sqlx.NewDb(db, "sqlite"))
	service := NewService(repository)
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	if err := service.EnsureRecurringExpenses(context.Background(), now); err != nil {
		t.Fatalf("first generation: %v", err)
	}
	if err := service.EnsureRecurringExpenses(context.Background(), now); err != nil {
		t.Fatalf("second generation: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM expenses`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("expense count = %d, want template plus two monthly instances", count)
	}

	var generated int
	if err := db.QueryRow(`SELECT COUNT(*) FROM expenses WHERE is_recurring = 0 AND reference LIKE 'recurring:%'`).Scan(&generated); err != nil {
		t.Fatal(err)
	}
	if generated != 2 {
		t.Fatalf("generated recurring expenses = %d, want 2", generated)
	}
}
