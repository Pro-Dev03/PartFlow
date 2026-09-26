package accounting

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
	"modernc.org/sqlite"
)

const DefaultStoreTimezone = "Asia/Jerusalem"

var (
	storeTimezoneMu sync.RWMutex
	storeTimezone   = DefaultStoreTimezone
)

var accountingExpenseStatuses = "LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')"

func init() {
	sqlite.MustRegisterScalarFunction("store_date", 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		if len(args) != 1 || args[0] == nil {
			return nil, nil
		}
		var raw string
		switch value := args[0].(type) {
		case string:
			raw = strings.TrimSpace(value)
		case []byte:
			raw = strings.TrimSpace(string(value))
		case time.Time:
			date, err := StoreDate(value)
			return date, err
		default:
			return nil, fmt.Errorf("unsupported SQLite store date value %T", args[0])
		}
		if len(raw) == len("2006-01-02") {
			if _, err := time.Parse("2006-01-02", raw); err == nil {
				return raw, nil
			}
		}
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return nil, err
		}
		date, err := StoreDate(parsed)
		return date, err
	})
}

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

// PostgresStoreDateExpression converts a PostgreSQL timestamp expression to
// the store's calendar date. Callers must pass a trusted SQL expression, not
// user supplied text.
func PostgresStoreDateExpression(timestampExpression string) string {
	timezone := strings.ReplaceAll(CurrentStoreTimezone(), "'", "''")
	return fmt.Sprintf("(%s AT TIME ZONE '%s')::date", timestampExpression, timezone)
}

// StoreNow returns the current instant. Business dates must interpret it with
// StoreLocation rather than the host or browser timezone.
func StoreNow() time.Time {
	return time.Now().UTC()
}

// ValidateStoreTimezone accepts an IANA timezone that can be used by both
// application date calculations and PostgreSQL AT TIME ZONE expressions.
func ValidateStoreTimezone(timezone string) (string, error) {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" || timezone == "Local" {
		return "", fmt.Errorf("store timezone must be a valid IANA timezone")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return "", fmt.Errorf("load store timezone %s: %w", timezone, err)
	}
	return timezone, nil
}

func ConfigureStoreTimezone(timezone string) error {
	validated, err := ValidateStoreTimezone(timezone)
	if err != nil {
		return err
	}
	storeTimezoneMu.Lock()
	storeTimezone = validated
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
	return ConfigureStoreTimezone(timezone)
}

// StoreDayBounds returns UTC instants for the local calendar day containing now.
func StoreDayBounds(now time.Time) (time.Time, time.Time, error) {
	date, err := StoreDate(now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return StoreDateBounds(date)
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
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse store date %q: %w", date, err)
	}
	start, err := firstInstantOfStoreDate(parsed, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	nextDate := parsed.AddDate(0, 0, 1)
	end, err := firstInstantOfStoreDate(nextDate, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start.UTC(), end.UTC(), nil
}

// firstInstantOfStoreDate finds the first representable instant whose local
// calendar date is date. This handles zones that move clocks at midnight and
// avoids assuming every business day is exactly 24 hours.
func firstInstantOfStoreDate(date time.Time, location *time.Location) (time.Time, error) {
	target := date.UTC().Format("2006-01-02")
	center := date.UTC()
	low := center.Add(-36 * time.Hour).Unix()
	high := center.Add(36 * time.Hour).Unix()
	for low < high {
		mid := low + (high-low)/2
		localDate := time.Unix(mid, 0).In(location).Format("2006-01-02")
		if localDate >= target {
			high = mid
		} else {
			low = mid + 1
		}
	}
	start := time.Unix(low, 0).In(location)
	if start.Format("2006-01-02") != target {
		return time.Time{}, fmt.Errorf("store date %s does not exist in timezone %s", target, location.String())
	}
	return start, nil
}

// StoreMonthBounds returns UTC boundaries for the store-local month containing
// the supplied instant.
func StoreMonthBounds(now time.Time) (time.Time, time.Time, error) {
	date, err := StoreDate(now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse store date %q: %w", date, err)
	}
	startDate := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonthDate := startDate.AddDate(0, 1, 0)
	start, _, err := StoreDateBounds(startDate.Format("2006-01-02"))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, _, err := StoreDateBounds(nextMonthDate.Format("2006-01-02"))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

// StorePreviousMonthBounds returns UTC boundaries for the previous store-local month.
func StorePreviousMonthBounds(now time.Time) (time.Time, time.Time, error) {
	date, err := StoreDate(now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	monthStart := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
	previousMonth := monthStart.AddDate(0, -1, 0)
	previousStart, _, err := StoreDateBounds(previousMonth.Format("2006-01-02"))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	currentStart, _, err := StoreDateBounds(monthStart.Format("2006-01-02"))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return previousStart, currentStart, nil
}

// StoreWeekBounds returns UTC boundaries for the Monday-based store-local week.
func StoreWeekBounds(now time.Time) (time.Time, time.Time, error) {
	date, err := StoreDate(now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse store date %q: %w", date, err)
	}
	daysSinceMonday := (int(parsed.Weekday()) + 6) % 7
	weekStart := parsed.AddDate(0, 0, -daysSinceMonday)
	return StoreDateBounds(weekStart.Format("2006-01-02"))
}

// StoreDateRangeKeys converts instants into a pair of store-local calendar
// keys using one timezone snapshot, so a concurrent settings update cannot
// split the range across two timezone configurations.
func StoreDateRangeKeys(start, end time.Time) (string, string, error) {
	location, err := StoreLocation()
	if err != nil {
		return "", "", err
	}
	return start.In(location).Format("2006-01-02"), end.In(location).Format("2006-01-02"), nil
}

// StoreDateRange returns inclusive-start/exclusive-end calendar dates for a
// rolling store-local period. It is used when SQL must not consult the host
// database timezone through CURRENT_DATE or date('now').
func StoreDateRange(now time.Time, days int) (string, string, error) {
	if days <= 0 {
		return "", "", fmt.Errorf("rolling store period must be positive")
	}
	date, err := StoreDate(now)
	if err != nil {
		return "", "", err
	}
	localDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", "", fmt.Errorf("parse store date %q: %w", date, err)
	}
	start := localDate.AddDate(0, 0, -(days - 1))
	return start.Format("2006-01-02"), localDate.AddDate(0, 0, 1).Format("2006-01-02"), nil
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
