package debts

import (
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestDebtSummaryTracksPartialAndClosedDebt(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE debts (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			amount REAL NOT NULL,
			remaining_amount REAL NOT NULL,
			status TEXT NOT NULL
		);
		INSERT INTO debts (id, customer_id, amount, remaining_amount, status)
		VALUES ('debt-1', 'customer-1', 250, 250, 'pending');
	`)
	if err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(sqlx.NewDb(db, "sqlite"))
	summary, err := handler.loadDebtSummary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalDebt != 250 || summary.PaidAmount != 0 || summary.RemainingAmount != 250 || summary.CustomerCount != 1 {
		t.Fatalf("initial summary = %+v, want total 250, paid 0, remaining 250, customers 1", summary)
	}

	_, err = db.Exec(`UPDATE debts SET remaining_amount = 150, status = 'partial' WHERE id = 'debt-1'`)
	if err != nil {
		t.Fatal(err)
	}
	summary, err = handler.loadDebtSummary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalDebt != 250 || summary.PaidAmount != 100 || summary.RemainingAmount != 150 || summary.CustomerCount != 1 {
		t.Fatalf("partial summary = %+v, want total 250, paid 100, remaining 150, customers 1", summary)
	}

	_, err = db.Exec(`UPDATE debts SET remaining_amount = 0, status = 'paid' WHERE id = 'debt-1'`)
	if err != nil {
		t.Fatal(err)
	}
	summary, err = handler.loadDebtSummary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalDebt != 250 || summary.PaidAmount != 250 || summary.RemainingAmount != 0 || summary.CustomerCount != 0 {
		t.Fatalf("closed summary = %+v, want total 250, paid 250, remaining 0, customers 0", summary)
	}
}
