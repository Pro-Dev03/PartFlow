package purchases

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestSQLitePurchaseReadsUseOfficialBusinessDate(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "purchase-business-date.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlite")

	for _, statement := range []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT NOT NULL)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, selling_price REAL, cost_price REAL)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT NOT NULL UNIQUE, supplier_id TEXT, purchase_date TEXT NOT NULL, tax_amount REAL DEFAULT 0, total_amount REAL NOT NULL, paid_amount REAL DEFAULT 0, status TEXT NOT NULL, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, item_code TEXT, status TEXT, created_at TEXT)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT, status TEXT)`,
		`CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT, purchase_item_id TEXT, quantity INTEGER)`,
	} {
		if _, err := xdb.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}

	supplierID := uuid.New()
	purchaseID := uuid.New()
	productID := uuid.New()
	businessDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	eventDate := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	if _, err := xdb.Exec(`INSERT INTO suppliers (id, name) VALUES (?, ?)`, supplierID, "Business Date Supplier"); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO products (id, name, selling_price, cost_price) VALUES (?, ?, ?, ?)`, productID, "Business Date Product", 20, 10); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, purchase_date, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "DATE-001", supplierID, businessDate.Format(time.RFC3339), 10, 0, "pending", eventDate.Format(time.RFC3339), eventDate.Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, uuid.New(), purchaseID, productID, 1, 10, 10, eventDate.Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(xdb)
	purchase, err := repo.GetByID(context.Background(), purchaseID)
	if err != nil {
		t.Fatal(err)
	}
	if !purchase.PurchaseDate.Equal(businessDate) {
		t.Fatalf("GetByID purchase_date = %s, want %s", purchase.PurchaseDate, businessDate)
	}

	rows, _, err := repo.ListSummaries(context.Background(), PurchaseListRequest{Page: 1, PerPage: 10, StartDate: &businessDate, EndDate: &businessDate})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListSummaries rows = %d, want 1 when filtering by purchase_date", len(rows))
	}
	if !rows[0].PurchaseDate.Equal(businessDate) {
		t.Fatalf("ListSummaries purchase_date = %s, want %s", rows[0].PurchaseDate, businessDate)
	}
}
