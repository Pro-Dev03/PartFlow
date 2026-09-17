package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/database"
)

// AccountingExpensesForPeriod is the canonical accounting expense query.
// SQLite legacy stores contain Go's timestamp text (for example, " +0000 UTC")
// that SQLite date functions cannot parse, so their local calendar prefix is
// used. New stores should persist normalized timestamps.
func AccountingExpensesForPeriod(ctx context.Context, db *sqlx.DB, start, end time.Time) (float64, error) {
	if db == nil {
		return 0, fmt.Errorf("accounting database is nil")
	}
	if !end.After(start) {
		return 0, fmt.Errorf("invalid accounting expense range: end must be after start")
	}

	query := fmt.Sprintf(
		"SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE expense_date >= $1 AND expense_date < $2 AND %s",
		AccountingExpenseStatusSQL(),
	)
	args := []any{start, end}
	if database.IsSQLite(db) {
		location, err := StoreLocation()
		if err != nil {
			return 0, err
		}
		query = fmt.Sprintf(
			"SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE date(substr(expense_date, 1, 10)) >= date(?) AND date(substr(expense_date, 1, 10)) < date(?) AND %s",
			AccountingExpenseStatusSQL(),
		)
		args = []any{start.In(location).Format("2006-01-02"), end.In(location).Format("2006-01-02")}
	}

	var total float64
	if err := db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, fmt.Errorf("retrieve accounting expenses for %s to %s: %w", start.Format(time.RFC3339), end.Format(time.RFC3339), err)
	}
	return total, nil
}
