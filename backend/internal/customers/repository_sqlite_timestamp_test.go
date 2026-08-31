package customers

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestRepositoryListHandlesSQLiteTextTimestamps(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE customers (
			id TEXT PRIMARY KEY,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			email TEXT,
			phone TEXT,
			address TEXT,
			city TEXT,
			country TEXT,
			tax_id TEXT,
			credit_limit REAL DEFAULT 0,
			current_balance REAL DEFAULT 0,
			notes TEXT,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	created := time.Now().UTC().Format(time.RFC3339)
	updated := time.Now().UTC().Add(time.Minute).Format(time.RFC3339)
	customerID := "123e4567-e89b-12d3-a456-426614174000"

	_, err = db.Exec(`
		INSERT INTO customers (
			id, code, name, email, phone, address, city, country, tax_id,
			credit_limit, current_balance, notes, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, customerID, "C-001", "Ahmad", "a@example.com", "+966500000000", "Main St", "Riyadh", "SA", "TAX-1", 0.0, 0.0, "note", 1, created, updated)
	if err != nil {
		t.Fatalf("insert customer: %v", err)
	}

	repo := NewRepository(db)
	customers, total, err := repo.List(context.Background(), &CustomerListRequest{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("List() total = %d, want 1", total)
	}
	if len(customers) != 1 {
		t.Fatalf("List() length = %d, want 1", len(customers))
	}
	if customers[0].CreatedAt.IsZero() {
		t.Fatal("CreatedAt is zero after SQLite text timestamp scan")
	}
	if customers[0].UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt is zero after SQLite text timestamp scan")
	}
}
