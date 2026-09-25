package accounting

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

const DefaultStoreTimezone = "Asia/Jerusalem"

var (
	storeTimezoneMu sync.RWMutex
	storeTimezone   = DefaultStoreTimezone
)

var accountingExpenseStatuses = "LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')"

// StoreLocation returns the calendar timezone used by store business dates.
func StoreLocation() (*time.Location, error) {
	storeTimezoneMu.RLock()
	timezone := storeTimezone
	storeTimezoneMu.RUnlock()
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("load store timezone %s: %w", timezone, err)
	}
	return location, nil
}

func CurrentStoreTimezone() string {
	storeTimezoneMu.RLock()
	defer storeTimezoneMu.RUnlock()
	return storeTimezone
}

// StoreNow returns the current instant. Business dates must interpret it with
// StoreLocation rather than the host or browser timezone.
func StoreNow() time.Time {
	return time.Now().UTC()
}

func ConfigureStoreTimezone(timezone string) error {
	if strings.TrimSpace(timezone) != DefaultStoreTimezone {
		return fmt.Errorf("PartFlow store timezone is fixed to %s", DefaultStoreTimezone)
	}
	if _, err := time.LoadLocation(DefaultStoreTimezone); err != nil {
		return fmt.Errorf("load store timezone %s: %w", DefaultStoreTimezone, err)
	}
	storeTimezoneMu.Lock()
	storeTimezone = DefaultStoreTimezone
	storeTimezoneMu.Unlock()
	return nil
}

// LoadStoreTimezone loads the persisted store timezone into the central
// accounting timezone state. Older databases may not have a settings table
// or a stored value, so the configured default remains in effect.
func LoadStoreTimezone(ctx context.Context, db *sqlx.DB) error {
	var timezone string
	query := db.Rebind(`SELECT value FROM settings WHERE key = ? LIMIT 1`)
	if err := db.GetContext(ctx, &timezone, query, "store_timezone"); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "no such table") || strings.Contains(message, "relation \"settings\" does not exist") {
			return nil
		}
		return fmt.Errorf("load persisted store timezone: %w", err)
	}
	if strings.TrimSpace(timezone) != DefaultStoreTimezone {
		query := db.Rebind(`UPDATE settings SET value = ? WHERE key = ?`)
		if _, err := db.ExecContext(ctx, query, DefaultStoreTimezone, "store_timezone"); err != nil {
			return fmt.Errorf("normalize persisted store timezone: %w", err)
		}
	}
	return ConfigureStoreTimezone(DefaultStoreTimezone)
}

// StoreDayBounds returns UTC instants for the local calendar day containing now.
func StoreDayBounds(now time.Time) (time.Time, time.Time, error) {
	location, err := StoreLocation()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	local := now.In(location)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	return start.UTC(), start.AddDate(0, 0, 1).UTC(), nil
}

// StoreDate returns the calendar date used by the store for a timestamp.
// Callers must use this instead of formatting UTC directly for day-scoped SQL.
func StoreDate(value time.Time) (string, error) {
	location, err := StoreLocation()
	if err != nil {
		return "", err
	}
	return value.In(location).Format("2006-01-02"), nil
}

// StoreDateBounds returns the UTC instants enclosing a store-local calendar date.
func StoreDateBounds(date string) (time.Time, time.Time, error) {
	location, err := StoreLocation()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	local, err := time.ParseInLocation("2006-01-02", date, location)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse store date %q: %w", date, err)
	}
	return local.UTC(), local.AddDate(0, 0, 1).UTC(), nil
}

// StoreMonthBounds returns UTC boundaries for the store-local month containing
// the supplied instant.
func StoreMonthBounds(now time.Time) (time.Time, time.Time, error) {
	location, err := StoreLocation()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	local := now.In(location)
	start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, location)
	return start.UTC(), start.AddDate(0, 1, 0).UTC(), nil
}

// StorePreviousMonthBounds returns UTC boundaries for the previous store-local month.
func StorePreviousMonthBounds(now time.Time) (time.Time, time.Time, error) {
	monthStart, _, err := StoreMonthBounds(now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	location, err := StoreLocation()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	previousStart := monthStart.In(location).AddDate(0, -1, 0)
	return previousStart.UTC(), monthStart, nil
}

// StoreWeekBounds returns UTC boundaries for the Monday-based store-local week.
func StoreWeekBounds(now time.Time) (time.Time, time.Time, error) {
	location, err := StoreLocation()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	local := now.In(location)
	daysSinceMonday := (int(local.Weekday()) + 6) % 7
	weekStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -daysSinceMonday)
	return weekStart.UTC(), weekStart.AddDate(0, 0, 7).UTC(), nil
}

// StoreDateRange returns inclusive-start/exclusive-end calendar dates for a
// rolling store-local period. It is used when SQL must not consult the host
// database timezone through CURRENT_DATE or date('now').
func StoreDateRange(now time.Time, days int) (string, string, error) {
	if days <= 0 {
		return "", "", fmt.Errorf("rolling store period must be positive")
	}
	location, err := StoreLocation()
	if err != nil {
		return "", "", err
	}
	local := now.In(location)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -days)
	return start.Format("2006-01-02"), start.AddDate(0, 0, days+1).Format("2006-01-02"), nil
}

// AccountingExpenseStatusSQL is the single accounting policy for expenses.
// Approved expenses affect accrual profit. There is no payment_status field yet,
// so cash-flow code must not claim that approval proves cash was paid.
func AccountingExpenseStatusSQL() string {
	return accountingExpenseStatuses
}

// IsAccountingExpenseStatus reports whether an expense status contributes to
// accrual profit and must therefore be treated as financially committed.
func IsAccountingExpenseStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved", "paid", "completed", "archived":
		return true
	default:
		return false
	}
}

// TotalExpensesForPeriod returns approved accounting expenses in [start, end).
// The range is supplied as UTC instants representing store-local boundaries.
func TotalExpensesForPeriod(ctx context.Context, db *sqlx.DB, start, end time.Time) (float64, error) {
	return AccountingExpensesForPeriod(ctx, db, start, end)
}
