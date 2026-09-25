package database

import (
	"testing"
	"time"

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

func TestParseTimestampAcceptsTimezoneLessISOText(t *testing.T) {
	parsed, err := ParseTimestamp("2026-09-18T19:54:37")
	if err != nil {
		t.Fatalf("ParseTimestamp() error = %v", err)
	}
	if got := parsed.Format("2006-01-02T15:04:05"); got != "2026-09-18T19:54:37" {
		t.Fatalf("ParseTimestamp() = %q", got)
	}
}

func TestParseTimestampAcceptsPostgresSpaceSeparatedUTCText(t *testing.T) {
	parsed, err := ParseTimestamp("2026-09-25 00:21:21.0201Z")
	if err != nil {
		t.Fatalf("ParseTimestamp() error = %v", err)
	}
	if got := parsed.UTC().Format(time.RFC3339Nano); got != "2026-09-25T00:21:21.0201Z" {
		t.Fatalf("ParseTimestamp() = %q", got)
	}
}
