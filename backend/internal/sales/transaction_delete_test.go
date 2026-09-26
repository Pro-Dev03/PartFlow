package sales

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestDeleteStandaloneFinancialTransactionHardDeletesSourceRowSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	xdb := sqlx.NewDb(db, "sqlite")
	if _, err := db.Exec(`CREATE TABLE financial_transactions (id TEXT PRIMARY KEY,sale_id TEXT,type TEXT NOT NULL,amount REAL NOT NULL,reference TEXT,created_at TEXT,updated_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	if _, err := db.Exec(`INSERT INTO financial_transactions (id,sale_id,type,amount,reference) VALUES (?,NULL,'income',75,'manual income')`, id); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(xdb), xdb)
	if err := service.DeleteTransaction(ctx, id, uuid.Nil); err != nil {
		t.Fatalf("delete standalone transaction: %v", err)
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM financial_transactions WHERE id=?`, id).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("financial transaction count=%d, want 0", remaining)
	}
	if err := service.DeleteTransaction(ctx, id, uuid.Nil); err != ErrFinancialTransactionNotFound {
		t.Fatalf("delete missing transaction error=%v, want not-found", err)
	}
}
