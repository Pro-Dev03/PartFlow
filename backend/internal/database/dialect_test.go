package database

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestSQLiteDialect(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if !IsSQLite(db) {
		t.Fatal("expected SQLite to be detected")
	}
	if NowSQL(db) != "CURRENT_TIMESTAMP" {
		t.Fatalf("NowSQL() = %q", NowSQL(db))
	}
	if Placeholder(db, 1) != "?" {
		t.Fatalf("Placeholder() = %q", Placeholder(db, 1))
	}
}
