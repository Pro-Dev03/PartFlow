package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/database"
)

// AccountingExpensesForPeriod is the canonical accounting expense query.
// Expense dates are store calendar dates, so both SQL backends compare the
// normalized store date keys instead of converting them through the database
// server timezone.
func AccountingExpensesForPeriod(ctx context.Context, db *sqlx.DB, start, end time.Time) (float64, error) {
	if db == nil {
		return 0, fmt.Errorf("accounting database is nil")
	}
	if !end.After(start) {
		return 0, fmt.Errorf("invalid accounting expense range: end must be after start")
	}

	startDate, endDate, err := StoreDateRangeKeys(start, end)
	if err != nil {
		return 0, err
	}
	query := fmt.Sprintf(
		"SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE expense_date >= $1::date AND expense_date < $2::date AND %s",
		AccountingExpenseStatusSQL(),
	)
	args := []any{startDate, endDate}
	if database.IsSQLite(db) {
		query = fmt.Sprintf(
			"SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE date(substr(expense_date, 1, 10)) >= date(?) AND date(substr(expense_date, 1, 10)) < date(?) AND %s",
			AccountingExpenseStatusSQL(),
		)
	}

	var total float64
	if err := db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, fmt.Errorf("retrieve accounting expenses for %s to %s: %w", start.Format(time.RFC3339), end.Format(time.RFC3339), err)
	}
	return total, nil
}
