package payments

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func openPaymentDeleteDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE payments (id TEXT PRIMARY KEY, payment_status TEXT)`); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestDeletePaymentOnlyDeletesPendingRowAtomicallySQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	pendingID := uuid.New()
	completedID := uuid.New()
	if _, err := db.Exec(`INSERT INTO payments (id, payment_status) VALUES (?, 'pending'), (?, 'completed')`, pendingID, completedID); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)

	if err := repo.Delete(context.Background(), completedID); err != ErrPaymentCannotBeCancelled {
		t.Fatalf("Delete(completed) error = %v, want ErrPaymentCannotBeCancelled", err)
	}
	var completedCount int
	if err := db.Get(&completedCount, `SELECT COUNT(*) FROM payments WHERE id = ?`, completedID); err != nil {
		t.Fatal(err)
	}
	if completedCount != 1 {
		t.Fatalf("completed payment rows = %d, want 1", completedCount)
	}

	if err := repo.Delete(context.Background(), pendingID); err != nil {
		t.Fatalf("Delete(pending): %v", err)
	}
	var pendingCount int
	if err := db.Get(&pendingCount, `SELECT COUNT(*) FROM payments WHERE id = ?`, pendingID); err != nil {
		t.Fatal(err)
	}
	if pendingCount != 0 {
		t.Fatalf("pending payment rows = %d, want 0", pendingCount)
	}
}

func TestDeletePaymentMissingRowReturnsNotFoundSQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	if err := NewRepository(db).Delete(context.Background(), uuid.New()); err != ErrPaymentNotFound {
		t.Fatalf("Delete(missing) error = %v, want ErrPaymentNotFound", err)
	}
}
