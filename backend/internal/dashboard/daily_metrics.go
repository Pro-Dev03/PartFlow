package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// todayMetrics contains values for the current business day only. It is kept
// separate from the dashboard's lifetime totals so the frontend cannot mistake
// cumulative purchases for today's cost of goods sold.
type todayMetrics struct {
	Sales           float64 `db:"today_sales"`
	Profit          float64 `db:"today_profit"`
	SupplierReturns float64 `db:"today_supplier_returns"`
}

func fetchTodayMetrics(ctx context.Context, db *sqlx.DB, now time.Time) (todayMetrics, error) {
	if db == nil {
		return todayMetrics{}, fmt.Errorf("dashboard database is nil")
	}

	now = now.UTC()
	date := now.Format("2006-01-02")
	query := `
		WITH sale_costs AS (
			SELECT s.id, s.total_amount, COALESCE(s.tax_amount, 0) AS tax_amount,
				COALESCE(SUM(si.quantity * COALESCE(ii.purchase_cost, p.cost_price, 0)), 0) AS total_cost
			FROM sales s
			LEFT JOIN sale_items si ON si.sale_id = s.id
			LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
			LEFT JOIN products p ON p.id = si.product_id
			WHERE COALESCE(s.sale_date::date, s.created_at::date) = $1::date
			  AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
			GROUP BY s.id, s.total_amount, s.tax_amount
		), totals AS (
			SELECT COALESCE(SUM(total_amount), 0) AS gross_revenue,
			       COALESCE(SUM(total_amount - tax_amount), 0) AS revenue,
			       COALESCE(SUM(total_cost), 0) AS cost
			FROM sale_costs
		), expenses_total AS (
			SELECT COALESCE(SUM(amount), 0) AS amount
			FROM expenses
			WHERE expense_date::date = $1::date
			  AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed')
		), returns_total AS (
			SELECT COALESCE(SUM(r.total_refund_amount), 0) AS refunded,
			       COALESCE(SUM(ri.quantity_returned * COALESCE(ri.original_cost, si.unit_cost, p.cost_price, 0)), 0) AS returned_cost
			FROM returns r
			JOIN return_items ri ON ri.return_id = r.id
			LEFT JOIN sale_items si ON si.id = ri.sale_item_id
			LEFT JOIN products p ON p.id = ri.product_id
			WHERE r.return_date::date = $1::date
			  AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'
		)
		SELECT totals.gross_revenue - returns_total.refunded AS today_sales,
		       totals.revenue - totals.cost - expenses_total.amount - returns_total.refunded + returns_total.returned_cost AS today_profit
		FROM totals, expenses_total, returns_total
	`
	args := []any{date}

	if isSQLiteDriver(db.DriverName()) {
		productJoin := "LEFT JOIN products p ON p.id = ri.product_id"
		if !sqliteHasColumns(db, "return_items", "product_id") {
			productJoin = "LEFT JOIN products p ON p.id = si.product_id"
		}
		query = fmt.Sprintf(`
			WITH sale_costs AS (
				SELECT s.id, s.total_amount, COALESCE(s.tax_amount, 0) AS tax_amount,
					COALESCE(SUM(si.quantity * COALESCE(ii.purchase_cost, p.purchase_price, p.cost_price, 0)), 0) AS total_cost
				FROM sales s
				LEFT JOIN sale_items si ON si.sale_id = s.id
				LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
				LEFT JOIN products p ON p.id = si.product_id
				WHERE date(s.created_at) = ?
				  AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
				GROUP BY s.id, s.total_amount, s.tax_amount
			), totals AS (
				SELECT COALESCE(SUM(total_amount), 0) AS gross_revenue,
				       COALESCE(SUM(total_amount - tax_amount), 0) AS revenue,
				       COALESCE(SUM(total_cost), 0) AS cost
				FROM sale_costs
			), expenses_total AS (
				SELECT COALESCE(SUM(amount), 0) AS amount
				FROM expenses
				WHERE date(expense_date) = ?
				  AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed')
			), returns_total AS (
				SELECT COALESCE(SUM(r.total_refund_amount), 0) AS refunded,
				       COALESCE(SUM(ri.quantity_returned * COALESCE(ri.original_cost, si.unit_cost, p.cost_price, 0)), 0) AS returned_cost
				FROM returns r
				JOIN return_items ri ON ri.return_id = r.id
				LEFT JOIN sale_items si ON si.id = ri.sale_item_id
				%s
				WHERE date(r.return_date) = ?
				  AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'
			)
			SELECT totals.gross_revenue AS today_sales,
			       totals.revenue - totals.cost - expenses_total.amount - returns_total.refunded + returns_total.returned_cost AS today_profit
			FROM totals, expenses_total, returns_total
		`, productJoin)
		args = []any{date, date, date}
		// Older local databases (and lightweight unit-test schemas) may not
		// have the returns tables yet. Keep the dashboard usable there while
		// using the return-aware calculation on the current schema.
		if !sqliteHasColumns(db, "returns", "total_refund_amount", "return_date", "status") || !sqliteHasColumns(db, "return_items", "quantity_returned", "sale_item_id", "original_cost") {
			productCostRef := "p.cost_price"
			if !sqliteHasColumns(db, "products", "cost_price") {
				productCostRef = "p.purchase_price"
			}
			query = fmt.Sprintf(`
				WITH sale_costs AS (
					SELECT s.id, s.total_amount, COALESCE(s.tax_amount, 0) AS tax_amount,
						COALESCE(SUM(si.quantity * COALESCE(ii.purchase_cost, %s, 0)), 0) AS total_cost
					FROM sales s
					LEFT JOIN sale_items si ON si.sale_id = s.id
					LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
					LEFT JOIN products p ON p.id = si.product_id
					WHERE date(s.created_at) = ?
					  AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
					GROUP BY s.id, s.total_amount, s.tax_amount
				), totals AS (
					SELECT COALESCE(SUM(total_amount), 0) AS gross_revenue,
					       COALESCE(SUM(total_amount - tax_amount), 0) AS revenue,
					       COALESCE(SUM(total_cost), 0) AS cost
					FROM sale_costs
				), expenses_total AS (
					SELECT COALESCE(SUM(amount), 0) AS amount
					FROM expenses
					WHERE date(expense_date) = ?
					  AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed')
				)
				SELECT totals.gross_revenue AS today_sales,
				       totals.revenue - totals.cost - expenses_total.amount AS today_profit
				FROM totals, expenses_total
			`, productCostRef)
			args = []any{date, date}
		}
	}
	if isSQLiteDriver(db.DriverName()) {
		if sqliteHasColumns(db, "returns", "refund_date", "updated_at") {
			query = strings.ReplaceAll(query, "date(r.return_date) = ?", "date(COALESCE(r.refund_date, r.updated_at, r.return_date)) = ?")
		}
	} else {
		query = strings.ReplaceAll(query, "r.return_date::date = $1::date", "COALESCE(r.refund_date::date, r.updated_at::date, r.return_date::date) = $1::date")
	}
	if isSQLiteDriver(db.DriverName()) && !sqliteHasColumns(db, "sales", "tax_amount") {
		query = strings.ReplaceAll(query, "COALESCE(s.tax_amount, 0)", "0")
		query = strings.ReplaceAll(query, "s.tax_amount", "0")
		query = strings.ReplaceAll(query, "GROUP BY s.id, s.total_amount, 0", "GROUP BY s.id, s.total_amount")
	}

	var metrics todayMetrics
	if err := db.GetContext(ctx, &metrics, query, args...); err != nil {
		return todayMetrics{}, fmt.Errorf("calculate today's dashboard metrics: %w", err)
	}
	if isSQLiteDriver(db.DriverName()) {
		if sqliteHasColumns(db, "supplier_returns", "refund_amount", "status", "updated_at", "created_at") {
			_ = db.GetContext(ctx, &metrics.SupplierReturns, `
				SELECT COALESCE(SUM(refund_amount), 0)
				FROM supplier_returns
				WHERE UPPER(COALESCE(status, '')) = 'COMPLETED'
				  AND date(COALESCE(updated_at, created_at)) = ?`, date)
		}
	} else {
		_ = db.GetContext(ctx, &metrics.SupplierReturns, `
			SELECT COALESCE(SUM(refund_amount), 0)
			FROM supplier_returns
			WHERE UPPER(COALESCE(status, '')) = 'COMPLETED'
			  AND updated_at::date = $1::date`, date)
	}
	return metrics, nil
}

func isSQLiteDriver(driver string) bool {
	driver = strings.ToLower(strings.TrimSpace(driver))
	return driver == "sqlite" || driver == "sqlite3"
}

func sqliteHasColumns(db *sqlx.DB, table string, required ...string) bool {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()

	found := make(map[string]bool, len(required))
	for rows.Next() {
		var cid, notNull, pk int
		var name, dataType string
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false
		}
		found[strings.ToLower(name)] = true
	}
	if err := rows.Err(); err != nil {
		return false
	}

	for _, column := range required {
		if !found[strings.ToLower(column)] {
			return false
		}
	}
	return true
}
