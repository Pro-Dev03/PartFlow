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

func TestStoreDateUsesConfiguredTimezoneAcrossMidnight(t *testing.T) {
	original := CurrentStoreTimezone()
	t.Cleanup(func() { _ = ConfigureStoreTimezone(original) })
	if err := ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		value time.Time
		want  string
	}{
		{name: "23:59:59", value: time.Date(2026, 9, 17, 3, 59, 59, 0, time.UTC), want: "2026-09-16"},
		{name: "00:00:00", value: time.Date(2026, 9, 17, 4, 0, 0, 0, time.UTC), want: "2026-09-17"},
		{name: "00:01", value: time.Date(2026, 9, 17, 4, 1, 0, 0, time.UTC), want: "2026-09-17"},
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

func TestStoreDateRangeIncludesExactlyRequestedCalendarDays(t *testing.T) {
	original := CurrentStoreTimezone()
	t.Cleanup(func() { _ = ConfigureStoreTimezone(original) })
	if err := ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}

	start, end, err := StoreDateRange(time.Date(2026, 9, 17, 3, 30, 0, 0, time.UTC), 30)
	if err != nil {
		t.Fatal(err)
	}
	if start != "2026-08-18" || end != "2026-09-17" {
		t.Fatalf("30-day range = [%s, %s), want 2026-08-18 through 2026-09-16 in New York", start, end)
	}
}

func TestStoreDayBoundsFollowConfiguredTimezoneDST(t *testing.T) {
	original := CurrentStoreTimezone()
	t.Cleanup(func() { _ = ConfigureStoreTimezone(original) })
	if err := ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	startOfDSTDay := time.Date(2026, time.March, 8, 12, 0, 0, 0, time.UTC)
	start, end, err := StoreDayBounds(startOfDSTDay)
	if err != nil {
		t.Fatal(err)
	}
	if duration := end.Sub(start); duration != 23*time.Hour {
		t.Fatalf("DST start day duration = %s, want 23h from the New York clock change", duration)
	}
	fallBackDay := time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC)
	start, end, err = StoreDayBounds(fallBackDay)
	if err != nil {
		t.Fatal(err)
	}
	if duration := end.Sub(start); duration != 25*time.Hour {
		t.Fatalf("DST end day duration = %s, want 25h from the New York clock change", duration)
	}
}

func TestChangingStoreTimezoneChangesDateAndMonthBounds(t *testing.T) {
	original := CurrentStoreTimezone()
	t.Cleanup(func() { _ = ConfigureStoreTimezone(original) })
	instant := time.Date(2026, 9, 16, 23, 30, 0, 0, time.UTC)
	if err := ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	date, err := StoreDate(instant)
	if err != nil {
		t.Fatal(err)
	}
	if date != "2026-09-16" {
		t.Fatalf("StoreDate() = %s, want New York date 2026-09-16", date)
	}
	if err := ConfigureStoreTimezone("Asia/Tokyo"); err != nil {
		t.Fatal(err)
	}
	date, err = StoreDate(instant)
	if err != nil {
		t.Fatal(err)
	}
	if date != "2026-09-17" {
		t.Fatalf("StoreDate() after timezone change = %s, want Tokyo date 2026-09-17", date)
	}
	monthStart, monthEnd, err := StoreMonthBounds(time.Date(2026, 9, 1, 0, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if got := monthStart.Format(time.RFC3339); got != "2026-08-31T15:00:00Z" {
		t.Fatalf("Tokyo month start = %s, want 2026-08-31T15:00:00Z", got)
	}
	if got := monthEnd.Format(time.RFC3339); got != "2026-09-30T15:00:00Z" {
		t.Fatalf("Tokyo month end = %s, want 2026-09-30T15:00:00Z", got)
	}
}

func TestConfigureStoreTimezoneRejectsInvalidValueWithoutChangingCurrentZone(t *testing.T) {
	original := CurrentStoreTimezone()
	if err := ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ConfigureStoreTimezone(original) })
	if err := ConfigureStoreTimezone("Local"); err == nil {
		t.Fatal("expected machine-local timezone to be rejected")
	}
	if got := CurrentStoreTimezone(); got != "America/New_York" {
		t.Fatalf("timezone changed after invalid value: got %s", got)
	}
}

func TestPostgresStoreDateExpressionUsesConfiguredTimezone(t *testing.T) {
	original := CurrentStoreTimezone()
	t.Cleanup(func() { _ = ConfigureStoreTimezone(original) })
	if err := ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	got := PostgresStoreDateExpression("created_at")
	want := "(created_at AT TIME ZONE 'America/New_York')::date"
	if got != want {
		t.Fatalf("PostgresStoreDateExpression() = %s, want %s", got, want)
	}
}

func TestSQLiteStoreDateUsesConfiguredTimezoneAndPreservesDateColumns(t *testing.T) {
	original := CurrentStoreTimezone()
	t.Cleanup(func() { _ = ConfigureStoreTimezone(original) })
	if err := ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tests := []struct {
		value string
		want  string
	}{
		{value: "2026-09-17T03:59:59Z", want: "2026-09-16"},
		{value: "2026-09-17T04:00:00Z", want: "2026-09-17"},
		{value: "2026-09-16", want: "2026-09-16"},
	}
	for _, test := range tests {
		var got string
		if err := db.QueryRow(`SELECT store_date(?)`, test.value).Scan(&got); err != nil {
			t.Fatalf("store_date(%q): %v", test.value, err)
		}
		if got != test.want {
			t.Errorf("store_date(%q) = %s, want %s", test.value, got, test.want)
		}
	}
}

func TestLoadStoreTimezonePreservesConfiguredValue(t *testing.T) {
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
	var stored string
	if err := db.QueryRow(`SELECT value FROM settings WHERE key='store_timezone'`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != "America/New_York" {
		t.Fatalf("stored timezone = %s, want America/New_York", stored)
	}
	if got := CurrentStoreTimezone(); got != "America/New_York" {
		t.Fatalf("active timezone = %s, want America/New_York", got)
	}
	date, err := StoreDate(time.Date(2026, 9, 16, 1, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if date != "2026-09-15" {
		t.Fatalf("StoreDate() = %s, want 2026-09-15 in New York", date)
	}
}
