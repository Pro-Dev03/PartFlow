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
