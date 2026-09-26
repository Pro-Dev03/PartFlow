package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
)

// todayMetrics contains values for the current business day only. It is kept
// separate from the dashboard's lifetime totals so the frontend cannot mistake
// cumulative purchases for today's cost of goods sold.
type todayMetrics struct {
	Sales           float64 `db:"today_sales"`
	Profit          float64 `db:"today_profit"`
	SupplierReturns float64 `db:"today_supplier_returns"`
	Collected       float64 `db:"today_collected"`
	DebtCollected   float64 `db:"today_debt_collected"`
	SupplierPaid    float64 `db:"today_supplier_paid"`
	Expenses        float64 `db:"today_expenses"`
}

func fetchTodayMetrics(ctx context.Context, db *sqlx.DB, now time.Time) (todayMetrics, error) {
	if db == nil {
		return todayMetrics{}, fmt.Errorf("dashboard database is nil")
	}
	if db.DriverName() == "sqlite" {
		ensureSQLiteAccountingReturnViews(db.DB)
	}

	date, err := accounting.StoreDate(now)
	if err != nil {
		return todayMetrics{}, fmt.Errorf("calculate store date: %w", err)
	}
	returnReferenceFilter := ""
	if isSQLiteDriver(db.DriverName()) {
		if sqliteHasColumns(db, "returns", "reference_number") {
			returnReferenceFilter = " AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'"
		}
	} else {
		returnReferenceFilter = " AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'"
	}
	query := `
		WITH sale_costs AS (
			SELECT s.id, s.total_amount, COALESCE(s.tax_amount, 0) AS tax_amount,
				CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount
				     ELSE COALESCE(SUM(si.quantity * COALESCE(NULLIF(si.unit_cost, 0), NULLIF(ii.purchase_cost, 0), p.cost_price, 0)), 0)
				END AS total_cost
			FROM sales s
			LEFT JOIN sale_items si ON si.sale_id = s.id
			LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
			LEFT JOIN products p ON p.id = si.product_id
			WHERE COALESCE(s.sale_date::date, s.created_at::date) = $1::date
			  AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
			GROUP BY s.id, s.total_amount, s.tax_amount, s.cost_amount
		), totals AS (
			SELECT COALESCE(SUM(total_amount), 0) AS gross_revenue,
			       COALESCE(SUM(total_amount - tax_amount), 0) AS revenue,
			       COALESCE(SUM(total_cost), 0) AS cost
			FROM sale_costs
		), expenses_total AS (SELECT 0 AS amount), return_lines AS (
			SELECT ri.return_id,
			       COALESCE(SUM(COALESCE(ri.total_refund_amount, 0)), 0) AS line_gross,
			       COALESCE(SUM(COALESCE(ri.total_refund_amount, 0) * CASE
			           WHEN COALESCE(si.total_amount, 0) > 0 THEN
			               CASE WHEN COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0) > 0
			                    THEN (COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) / si.total_amount ELSE 0 END
			           ELSE 1 END), 0) AS line_net,
		       COALESCE(SUM(ri.quantity_returned * COALESCE(NULLIF(ri.original_cost, 0), NULLIF(si.unit_cost, 0), p.cost_price, 0)), 0) AS returned_cost
			FROM accounting_return_items ri
			LEFT JOIN sale_items si ON si.id = ri.sale_item_id
			LEFT JOIN products p ON p.id = ri.product_id
			GROUP BY ri.return_id
		), returns_total AS (
			SELECT COALESCE(SUM(CASE WHEN COALESCE(lines.line_gross, 0) > 0
			                    THEN r.total_refund_amount * lines.line_net / lines.line_gross
			                    ELSE r.total_refund_amount END), 0) AS refunded,
			       COALESCE(SUM(COALESCE(lines.returned_cost, 0)), 0) AS returned_cost
			FROM accounting_returns r
			LEFT JOIN return_lines lines ON lines.return_id = r.id
			WHERE r.return_date::date = $1::date
			  AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'` + returnReferenceFilter + `
		)
		SELECT totals.revenue - returns_total.refunded AS today_sales,
		       totals.revenue - totals.cost - expenses_total.amount - returns_total.refunded + returns_total.returned_cost AS today_profit
		FROM totals, expenses_total, returns_total
	`
	args := []any{date}

	if isSQLiteDriver(db.DriverName()) {
		productCostRef := "p.cost_price"
		if !sqliteHasColumns(db, "products", "cost_price") {
			productCostRef = "p.purchase_price"
		}
		returnQuantityColumn := ""
		if sqliteHasColumns(db, "return_items", "quantity_returned") {
			returnQuantityColumn = "ri.quantity_returned"
		} else if sqliteHasColumns(db, "return_items", "quantity") {
			returnQuantityColumn = "ri.quantity"
		}
		productJoin := "LEFT JOIN products p ON p.id = ri.product_id"
		if !sqliteHasColumns(db, "return_items", "product_id") {
			productJoin = "LEFT JOIN products p ON p.id = si.product_id"
		}
		returnItemRefundAmount := "0"
		if sqliteHasColumns(db, "return_items", "total_refund_amount") {
			returnItemRefundAmount = "ri.total_refund_amount"
		}
		returnTaxRatio := "1"
		if sqliteHasColumns(db, "sale_items", "total_amount", "tax_amount") {
			returnTaxRatio = `CASE WHEN COALESCE(si.total_amount, 0) > 0 THEN
				CASE WHEN COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0) > 0
					THEN (COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) / si.total_amount ELSE 0 END
				ELSE 1 END`
		}
		query = fmt.Sprintf(`
			WITH sale_costs AS (
				SELECT s.id, s.total_amount, COALESCE(s.tax_amount, 0) AS tax_amount,
					COALESCE(SUM(si.quantity * COALESCE(NULLIF(si.unit_cost, 0), NULLIF(ii.purchase_cost, 0), p.cost_price, 0)), 0) AS total_cost
				FROM sales s
				LEFT JOIN sale_items si ON si.sale_id = s.id
				LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
				LEFT JOIN products p ON p.id = si.product_id
				WHERE store_date(s.created_at) = ?
				  AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
				GROUP BY s.id, s.total_amount, s.tax_amount
			), totals AS (
				SELECT COALESCE(SUM(total_amount), 0) AS gross_revenue,
				       COALESCE(SUM(total_amount - tax_amount), 0) AS revenue,
				       COALESCE(SUM(total_cost), 0) AS cost
				FROM sale_costs
			), expenses_total AS (SELECT 0 AS amount), return_lines AS (
				SELECT ri.return_id,
				       COALESCE(SUM(COALESCE(%s, 0)), 0) AS line_gross,
				       COALESCE(SUM(COALESCE(%s, 0) * (%s)), 0) AS line_net,
				       COALESCE(SUM(%s * COALESCE(NULLIF(ri.original_cost, 0), NULLIF(si.unit_cost, 0), p.cost_price, 0)), 0) AS returned_cost
				FROM accounting_return_items ri
				LEFT JOIN sale_items si ON si.id = ri.sale_item_id
				%s
				GROUP BY ri.return_id
			), returns_total AS (
				SELECT COALESCE(SUM(CASE WHEN COALESCE(lines.line_gross, 0) > 0
				                    THEN r.total_refund_amount * lines.line_net / lines.line_gross
				                    ELSE r.total_refund_amount END), 0) AS refunded,
				       COALESCE(SUM(COALESCE(lines.returned_cost, 0)), 0) AS returned_cost
				FROM accounting_returns r
				LEFT JOIN return_lines lines ON lines.return_id = r.id
				WHERE store_date(r.return_date) = ?
			  AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'`+returnReferenceFilter+`
			)
			SELECT totals.revenue - returns_total.refunded AS today_sales,
			       totals.revenue - totals.cost - expenses_total.amount - returns_total.refunded + returns_total.returned_cost AS today_profit
			FROM totals, expenses_total, returns_total
		`, returnItemRefundAmount, returnItemRefundAmount, returnTaxRatio, returnQuantityColumn, productJoin)
		args = []any{date, date, date}
		// Older local databases (and lightweight unit-test schemas) may not
		// have the returns tables yet. Keep the dashboard usable there while
		// using the return-aware calculation on the current schema.
		if returnQuantityColumn == "" || !sqliteHasColumns(db, "returns", "total_refund_amount", "return_date", "status") || !sqliteHasColumns(db, "return_items", "sale_item_id", "original_cost") {
			query = fmt.Sprintf(`
				WITH sale_costs AS (
					SELECT s.id, s.total_amount, COALESCE(s.tax_amount, 0) AS tax_amount,
					COALESCE(SUM(si.quantity * COALESCE(NULLIF(ii.purchase_cost, 0), %s, 0)), 0) AS total_cost
					FROM sales s
					LEFT JOIN sale_items si ON si.sale_id = s.id
					LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
					LEFT JOIN products p ON p.id = si.product_id
					WHERE store_date(s.created_at) = ?
					  AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
					GROUP BY s.id, s.total_amount, s.tax_amount
				), totals AS (
					SELECT COALESCE(SUM(total_amount), 0) AS gross_revenue,
					       COALESCE(SUM(total_amount - tax_amount), 0) AS revenue,
					       COALESCE(SUM(total_cost), 0) AS cost
					FROM sale_costs
				), expenses_total AS (SELECT 0 AS amount)
				SELECT totals.revenue AS today_sales,
				       totals.revenue - totals.cost - expenses_total.amount AS today_profit
				FROM totals, expenses_total
			`, productCostRef)
			args = []any{date, date}
		}
		if sqliteHasColumns(db, "sales", "sale_date") {
			query = strings.ReplaceAll(query, "store_date(s.created_at) = ?", "store_date(COALESCE(s.sale_date, s.created_at)) = ?")
		}
		if sqliteHasColumns(db, "sales", "cost_amount") {
			query = strings.ReplaceAll(query,
				"COALESCE(SUM(si.quantity * COALESCE(NULLIF(si.unit_cost, 0), NULLIF(ii.purchase_cost, 0), p.cost_price, 0)), 0) AS total_cost",
				"CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount ELSE COALESCE(SUM(si.quantity * COALESCE(NULLIF(si.unit_cost, 0), NULLIF(ii.purchase_cost, 0), p.cost_price, 0)), 0) END AS total_cost")
			query = strings.ReplaceAll(query,
				fmt.Sprintf("COALESCE(SUM(si.quantity * COALESCE(ii.purchase_cost, %s, 0)), 0) AS total_cost", productCostRef),
				fmt.Sprintf("CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount ELSE COALESCE(SUM(si.quantity * COALESCE(NULLIF(ii.purchase_cost, 0), %s, 0)), 0) END AS total_cost", productCostRef))
			query = strings.ReplaceAll(query, "GROUP BY s.id, s.total_amount, s.tax_amount", "GROUP BY s.id, s.total_amount, s.tax_amount, s.cost_amount")
		}
	}
	if isSQLiteDriver(db.DriverName()) {
		if sqliteHasColumns(db, "returns", "refund_date", "updated_at") {
			query = strings.ReplaceAll(query, "store_date(r.return_date) = ?", "store_date(COALESCE(r.refund_date, r.return_date, r.updated_at, r.created_at)) = ?")
		}
	} else {
		query = strings.ReplaceAll(query, "s.created_at::date", accounting.PostgresStoreDateExpression("s.created_at"))
		query = strings.ReplaceAll(query, "r.return_date::date = $1::date", "COALESCE(r.refund_date::date, r.updated_at::date, r.return_date::date) = $1::date")
		query = strings.ReplaceAll(query, "r.updated_at::date", accounting.PostgresStoreDateExpression("r.updated_at"))
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
	expenseStart, expenseEnd, err := accounting.StoreDateBounds(date)
	if err != nil {
		return todayMetrics{}, fmt.Errorf("calculate today's accounting period: %w", err)
	}
	metrics.Expenses, err = accounting.AccountingExpensesForPeriod(ctx, db, expenseStart, expenseEnd)
	if err != nil {
		return todayMetrics{}, err
	}
	metrics.Profit -= metrics.Expenses
	if isSQLiteDriver(db.DriverName()) {
		if sqliteHasColumns(db, "supplier_returns", "refund_amount", "status", "updated_at", "created_at") {
			_ = db.GetContext(ctx, &metrics.SupplierReturns, `
				SELECT COALESCE(SUM(refund_amount), 0)
				FROM supplier_returns
				WHERE UPPER(COALESCE(status, '')) = 'COMPLETED'
				  AND store_date(COALESCE(updated_at, created_at)) = ?`, date)
		}
	} else {
		_ = db.GetContext(ctx, &metrics.SupplierReturns, fmt.Sprintf(`
			SELECT COALESCE(SUM(refund_amount), 0)
			FROM supplier_returns
			WHERE UPPER(COALESCE(status, '')) = 'COMPLETED'
			  AND %s = $1::date`, accounting.PostgresStoreDateExpression("updated_at")), date)
	}

	// Payments are the reliable source for cash movement. These are best-effort
	// so older local databases without the payments table remain usable.
	paymentDateColumn := "created_at"
	debtPaymentFilter := "1 = 1"
	customerPaymentFilter := "customer_id IS NOT NULL"
	collectedPaymentFilter := customerPaymentFilter
	supplierPaymentFilter := "supplier_id IS NOT NULL"
	paymentStatusFilter := "1 = 1"
	if sqliteHasColumns(db, "payments", "payment_date") {
		// SQLite may receive Go's time.Time string, which can include a
		// monotonic suffix ("m=+"). Keep the parseable date-time prefix.
		paymentDateColumn = "COALESCE(payment_date, created_at)"
	}
	if isSQLiteDriver(db.DriverName()) && paymentDateColumn == "created_at" {
		paymentDateColumn = "created_at"
	}
	if sqliteHasColumns(db, "payments", "reference_id", "type") {
		customerPaymentFilter = "(customer_id IS NOT NULL OR (LOWER(COALESCE(type, '')) = 'customer' AND reference_id IS NOT NULL))"
		supplierPaymentFilter = "(supplier_id IS NOT NULL OR (LOWER(COALESCE(type, '')) = 'supplier' AND reference_id IS NOT NULL))"
	}
	collectedPaymentFilter = customerPaymentFilter
	hasSaleID := sqliteHasColumns(db, "payments", "sale_id")
	if !isSQLiteDriver(db.DriverName()) {
		hasSaleID = true
	}
	if hasSaleID {
		// A sale-linked payment is the payment made at checkout, even when the
		// sale creates a debt. Debt collections are recorded separately with no
		// sale_id and must be the only values shown in today's debt collection.
		debtPaymentFilter = "(sale_id IS NULL OR TRIM(sale_id) = '') AND " + customerPaymentFilter
		collectedPaymentFilter = "((sale_id IS NOT NULL AND TRIM(sale_id) <> '') OR " + customerPaymentFilter + ")"
	}
	if isSQLiteDriver(db.DriverName()) && sqliteHasColumns(db, "payments", "payment_status") {
		paymentStatusFilter = "LOWER(COALESCE(payment_status, 'completed')) IN ('completed', 'paid')"
	} else if !isSQLiteDriver(db.DriverName()) {
		paymentStatusFilter = "LOWER(COALESCE(status, payment_status, 'completed')) IN ('completed', 'paid')"
	}
	if isSQLiteDriver(db.DriverName()) {
		_ = db.GetContext(ctx, &metrics.Collected, fmt.Sprintf(`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE %s AND %s AND store_date(%s) = ?`, collectedPaymentFilter, paymentStatusFilter, paymentDateColumn), date)
		_ = db.GetContext(ctx, &metrics.DebtCollected, fmt.Sprintf(`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE %s AND %s AND store_date(%s) = ?`, debtPaymentFilter, paymentStatusFilter, paymentDateColumn), date)
		_ = db.GetContext(ctx, &metrics.SupplierPaid, fmt.Sprintf(`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE %s AND %s AND store_date(%s) = ?`, supplierPaymentFilter, paymentStatusFilter, paymentDateColumn), date)
		if !hasSaleID && sqliteHasColumns(db, "sales", "payment_method", "paid_amount") {
			var salesCashIn float64
			dateColumn := "created_at"
			if sqliteHasColumns(db, "sales", "sale_date") {
				dateColumn = "COALESCE(sale_date, created_at)"
			}
			_ = db.GetContext(ctx, &salesCashIn, fmt.Sprintf(`
				SELECT COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('cash', 'cash_payment', 'card', 'credit_card')
					THEN COALESCE(NULLIF(paid_amount, 0), total_amount) ELSE 0 END), 0)
				FROM sales
				WHERE LOWER(COALESCE(status, 'completed')) = 'completed' AND store_date(%s) = ?`, dateColumn), date)
			metrics.Collected += salesCashIn
		}
	} else {
		paymentStoreDate := "COALESCE(payment_date::date, " + accounting.PostgresStoreDateExpression("created_at") + ")"
		_ = db.GetContext(ctx, &metrics.Collected, fmt.Sprintf(`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE (sale_id IS NOT NULL OR customer_id IS NOT NULL OR (LOWER(COALESCE(type, '')) = 'customer' AND reference_id IS NOT NULL)) AND LOWER(COALESCE(status, payment_status, 'completed')) IN ('completed', 'paid') AND %s = $1::date`, paymentStoreDate), date)
		_ = db.GetContext(ctx, &metrics.DebtCollected, fmt.Sprintf(`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE sale_id IS NULL AND (customer_id IS NOT NULL OR (LOWER(COALESCE(type, '')) = 'customer' AND reference_id IS NOT NULL)) AND LOWER(COALESCE(status, payment_status, 'completed')) IN ('completed', 'paid') AND %s = $1::date`, paymentStoreDate), date)
		_ = db.GetContext(ctx, &metrics.SupplierPaid, fmt.Sprintf(`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE (supplier_id IS NOT NULL OR (LOWER(COALESCE(type, '')) = 'supplier' AND reference_id IS NOT NULL)) AND LOWER(COALESCE(status, payment_status, 'completed')) IN ('completed', 'paid') AND %s = $1::date`, paymentStoreDate), date)
		_ = db.GetContext(ctx, &metrics.Expenses, `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE expense_date::date = $1::date AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')`, date)
	}
	return metrics, nil
}

func isSQLiteDriver(driver string) bool {
	driver = strings.ToLower(strings.TrimSpace(driver))
	return driver == "sqlite" || driver == "sqlite3"
}

func sqliteHasColumns(db *sqlx.DB, table string, required ...string) bool {
	var rows *sqlx.Rows
	var err error
	if isSQLiteDriver(db.DriverName()) {
		rows, err = db.Queryx("PRAGMA table_info(" + table + ")")
	} else {
		rows, err = db.Queryx(`SELECT column_name AS name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1`, table)
	}
	if err != nil {
		return false
	}
	defer rows.Close()

	found := make(map[string]bool, len(required))
	if !isSQLiteDriver(db.DriverName()) {
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
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
