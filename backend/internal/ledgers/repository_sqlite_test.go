package ledgers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestLedgerRepositorySQLiteRoundTrip(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range []string{
		`CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, credit_limit REAL DEFAULT 0, current_balance REAL DEFAULT 0)`,
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL DEFAULT 0)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, sku TEXT, barcode TEXT)`,
		`CREATE TABLE ledger_entries (id TEXT PRIMARY KEY, ledger_type TEXT NOT NULL, entity_id TEXT NOT NULL, transaction_type TEXT NOT NULL, reference_id TEXT, reference_type TEXT, amount REAL NOT NULL, balance REAL NOT NULL, previous_balance REAL DEFAULT 0, description TEXT, metadata TEXT DEFAULT '{}', created_by TEXT, created_at TEXT NOT NULL)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	customerID, userID, saleID := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO customers (id,name,credit_limit,current_balance) VALUES (?,?,?,?)`, customerID.String(), "Customer", 500, 40); err != nil {
		t.Fatal(err)
	}
	xdb := sqlx.NewDb(db, "sqlite")
	repo := NewRepository(xdb)
	now := time.Now().UTC()
	entry := &LedgerEntry{ID: uuid.New(), LedgerType: LedgerTypeCustomer, EntityID: customerID, TransactionType: TransactionSale, ReferenceID: &saleID, Amount: 40, Balance: 40, Description: "Sale", Metadata: map[string]interface{}{"source": "test"}, CreatedBy: userID, CreatedAt: now}
	if err := repo.CreateLedgerEntry(context.Background(), entry); err != nil {
		t.Fatalf("create ledger entry: %v", err)
	}
	entries, total, err := repo.GetLedgerEntries(context.Background(), LedgerTypeCustomer, customerID, 1, 20)
	if err != nil {
		t.Fatalf("list ledger entries: %v", err)
	}
	if total != 1 || len(entries) != 1 || entries[0].Balance != 40 || entries[0].Metadata["source"] != "test" {
		t.Fatalf("unexpected ledger result total=%d entries=%+v", total, entries)
	}
	summary, err := repo.GetCustomerLedgerSummary(context.Background(), customerID)
	if err != nil {
		t.Fatalf("get customer ledger summary: %v", err)
	}
	if summary.TotalPurchases != 40 || summary.CurrentBalance != 40 {
		t.Fatalf("unexpected customer summary: %+v", summary)
	}
}
