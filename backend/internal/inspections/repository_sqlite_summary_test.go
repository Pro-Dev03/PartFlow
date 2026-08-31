package inspections

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestGetInspectionSummarySQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE inspections (id TEXT PRIMARY KEY, result TEXT, inspection_date TEXT, condition TEXT, grade TEXT)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inspections (id, result, inspection_date, condition, grade) VALUES
		('1', 'passed', date('now'), 'good', 'B'),
		('2', 'pending', date('now'), 'fair', 'C')`)
	if err != nil {
		t.Fatal(err)
	}
	summary, err := NewRepository(db).GetInspectionSummary(context.Background())
	if err != nil {
		t.Fatalf("SQLite summary failed: %v", err)
	}
	if summary.TotalInspections != 2 || summary.PassedInspections != 1 || summary.PendingInspections != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.ThisWeek != 2 || summary.ThisMonth != 2 {
		t.Fatalf("expected current rows in week/month: %+v", summary)
	}
}
