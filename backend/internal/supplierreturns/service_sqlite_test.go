package supplierreturns

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestServiceListHandlesSQLiteTextTimestamps(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE supplier_returns (
			id TEXT PRIMARY KEY,
			customer_return_id TEXT,
			sale_id TEXT,
			purchase_id TEXT NOT NULL,
			supplier_id TEXT NOT NULL,
			return_number TEXT NOT NULL,
			status TEXT NOT NULL,
			reason TEXT NOT NULL,
			refund_amount REAL NOT NULL DEFAULT 0,
			notes TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE supplier_return_items (
			id TEXT PRIMARY KEY,
			supplier_return_id TEXT NOT NULL,
			inventory_item_id TEXT,
			barcode TEXT,
			serial_number TEXT,
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create supplier return items table: %v", err)
	}

	id := uuid.New()
	purchaseID := uuid.New()
	supplierID := uuid.New()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	_, err = db.Exec(`
		INSERT INTO supplier_returns (id, purchase_id, supplier_id, return_number, status, reason, refund_amount, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id.String(), purchaseID.String(), supplierID.String(), "SRET-001", "PENDING", "DEFECTIVE", 25.5, "bad item", createdAt, createdAt)
	if err != nil {
		t.Fatalf("insert row: %v", err)
	}

	service := NewService(db)
	rows, err := service.List(context.Background(), "")
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("List() len = %d, want 1", len(rows))
	}
	if rows[0].CreatedAt.IsZero() {
		t.Fatal("CreatedAt is zero after SQLite text timestamp scan")
	}
}

func TestServiceListDirectSupplierReturnDoesNotLookLikeCustomerReturn(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE supplier_returns (
			id TEXT PRIMARY KEY,
			customer_return_id TEXT,
			sale_id TEXT,
			purchase_id TEXT NOT NULL,
			supplier_id TEXT NOT NULL,
			return_number TEXT NOT NULL,
			status TEXT NOT NULL,
			source_status TEXT,
			reason TEXT NOT NULL,
			refund_amount REAL NOT NULL DEFAULT 0,
			notes TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE supplier_return_items (
			id TEXT PRIMARY KEY,
			supplier_return_id TEXT NOT NULL,
			inventory_item_id TEXT,
			barcode TEXT,
			serial_number TEXT,
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create supplier return items table: %v", err)
	}

	id := uuid.New()
	purchaseID := uuid.New()
	supplierID := uuid.New()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	_, err = db.Exec(`
		INSERT INTO supplier_returns (id, customer_return_id, sale_id, purchase_id, supplier_id, return_number, status, source_status, reason, refund_amount, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id.String(), "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000", purchaseID.String(), supplierID.String(), "SRET-001", "PENDING", "RESOLVED", "Damaged item", 25.5, "direct supplier request", createdAt, createdAt)
	if err != nil {
		t.Fatalf("insert direct supplier return row: %v", err)
	}

	service := NewService(db)
	rows, err := service.List(context.Background(), "")
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("List() len = %d, want 1", len(rows))
	}
	if rows[0].CustomerReturnID != uuid.Nil {
		t.Fatalf("CustomerReturnID = %s, want zero UUID for direct supplier return", rows[0].CustomerReturnID)
	}
	if rows[0].Source != "Supplier Return" {
		t.Fatalf("Source = %q, want %q", rows[0].Source, "Supplier Return")
	}
}
