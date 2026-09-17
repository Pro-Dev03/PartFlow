package accounting

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestTotalExpensesForStoreDayUsesOnePolicyAndTimezoneBoundaries(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, expense_date TEXT, status TEXT);
		INSERT INTO expenses (id, amount, expense_date, status) VALUES
			('before', 2, '2026-09-15 23:59:59 +0300 EEST', 'approved'),
			('at-start', 50, '2026-09-16 12:00:00 +0000 UTC', 'approved'),
			('local-late', 1, '2026-09-16 23:59:00 +0300 EEST', 'approved'),
			('next-start', 7, '2026-09-17 00:00:00 +0300 EEST', 'approved'),
			('pending', 100, '2026-09-16 13:00:00 +0000 UTC', 'pending');
	`); err != nil {
		t.Fatal(err)
	}

	location, err := StoreLocation()
	if err != nil {
		t.Fatal(err)
	}
	localDay := time.Date(2026, 9, 16, 12, 0, 0, 0, location)
	start, end, err := StoreDayBounds(localDay)
	if err != nil {
		t.Fatal(err)
	}

	total, err := TotalExpensesForPeriod(context.Background(), sqlx.NewDb(db, "sqlite"), start, end)
	if err != nil {
		t.Fatal(err)
	}
	if total != 51 {
		t.Fatalf("expenses for store day = %v, want 51", total)
	}
}

func TestTotalExpensesForPeriodReturnsQueryErrors(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	start := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	_, err = TotalExpensesForPeriod(context.Background(), db, start, end)
	if err == nil {
		t.Fatal("expected missing expenses table error")
	}
}

func TestStoreDateUsesAsiaJerusalemBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		value time.Time
		want  string
	}{
		{name: "23:59", value: time.Date(2026, 9, 16, 20, 59, 0, 0, time.UTC), want: "2026-09-16"},
		{name: "00:00", value: time.Date(2026, 9, 16, 21, 0, 0, 0, time.UTC), want: "2026-09-17"},
		{name: "00:01", value: time.Date(2026, 9, 16, 21, 1, 0, 0, time.UTC), want: "2026-09-17"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := StoreDate(test.value)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("StoreDate() = %s, want %s", got, test.want)
			}
		})
	}
}

func TestStoreDateUsesConfiguredTimezoneForSameInstant(t *testing.T) {
	instant := time.Date(2026, 9, 16, 23, 30, 0, 0, time.UTC)
	original := CurrentStoreTimezone()
	defer ConfigureStoreTimezone(original)

	if err := ConfigureStoreTimezone("Asia/Jerusalem"); err != nil {
		t.Fatal(err)
	}
	jerusalemDate, err := StoreDate(instant)
	if err != nil {
		t.Fatal(err)
	}
	if err := ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	newYorkDate, err := StoreDate(instant)
	if err != nil {
		t.Fatal(err)
	}
	if jerusalemDate == newYorkDate {
		t.Fatalf("same instant should be interpreted by each configured timezone: both were %s", jerusalemDate)
	}
}

func TestLoadStoreTimezoneUsesPersistedIANAValue(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT); INSERT INTO settings (key, value) VALUES ('store_timezone', 'America/New_York');`); err != nil {
		t.Fatal(err)
	}

	original := CurrentStoreTimezone()
	defer ConfigureStoreTimezone(original)
	if err := LoadStoreTimezone(context.Background(), sqlx.NewDb(db, "sqlite")); err != nil {
		t.Fatal(err)
	}
	date, err := StoreDate(time.Date(2026, 9, 16, 1, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if date != "2026-09-15" {
		t.Fatalf("StoreDate() = %s, want 2026-09-15", date)
	}
}
