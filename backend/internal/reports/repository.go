package reports

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles report data operations
type Repository struct {
	db *sqlx.DB
}

func (r *Repository) returnDateExpression(alias string) string {
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "accounting_returns", "refund_date", "updated_at") {
		return fmt.Sprintf("date(%s.return_date)", alias)
	}
	return fmt.Sprintf("date(COALESCE(%s.refund_date, %s.updated_at, %s.return_date))", alias, alias, alias)
}

func (r *Repository) salesDateExpression(alias string) string {
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "sales", "created_at") {
		if alias == "" {
			return "date(sale_date)"
		}
		return fmt.Sprintf("date(%s.sale_date)", alias)
	}
	if alias == "" {
		return "date(sale_date)"
	}
	return fmt.Sprintf("date(%s.sale_date)", alias)
}

func reportsSQLiteHasColumns(db *sqlx.DB, table string, required ...string) bool {
	if !dbutil.IsSQLite(db) {
		return true
	}
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()
	found := map[string]bool{}
	for rows.Next() {
		var cid, notNull, pk int
		var name, dataType string
		var defaultValue any
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

// reportTimestamp accepts both PostgreSQL timestamps and the TEXT timestamps
// used by the local SQLite store.  SQLite returns TEXT values for date
// expressions (DATE/strftime), so scanning directly into time.Time silently
// dropped rows from several reports.
type reportTimestamp struct{ time.Time }

func (t *reportTimestamp) Scan(value any) error {
	parsed, err := dbutil.ParseTimestamp(value)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// NewRepository creates a new report repository
func NewRepository(db *sqlx.DB) *Repository {
	// Keep lightweight SQLite report fixtures and pre-ledger test databases
	// compatible. Initialized PartFlow databases create permanent union views
	// through the schema migration, so this fallback is only a direct alias.
	if db != nil && dbutil.IsSQLite(db) {
		ensureSQLiteReportAliasView(db, "accounting_returns", "returns")
		ensureSQLiteReportAliasView(db, "accounting_return_items", "return_items")
	}
	return &Repository{db: db}
}

func ensureSQLiteReportAliasView(db *sqlx.DB, viewName, sourceName string) {
	var viewCount int
	if err := db.Get(&viewCount, `SELECT COUNT(*) FROM sqlite_master WHERE type='view' AND name=?`, viewName); err != nil || viewCount > 0 {
		return
	}
	var tableCount int
	if err := db.Get(&tableCount, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, sourceName); err != nil || tableCount == 0 {
		return
	}
	_, _ = db.Exec(`CREATE TEMP VIEW ` + viewName + ` AS SELECT * FROM ` + sourceName)
}

// historicalCOGSTotalSQL returns one COGS value per sale. A stored sale cost
// is authoritative; line costs are only used for legacy rows without it.
func (r *Repository) historicalCOGSTotalSQL(startPlaceholder, endPlaceholder string) string {
	saleDateExpr := r.salesDateExpression("s")
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "sales", "cost_amount") {
		return fmt.Sprintf(`
			SELECT COALESCE(SUM(CASE WHEN sale_cost IS NOT NULL THEN sale_cost ELSE line_cost END), 0)
			FROM (
				SELECT s.id, s.cost_amount AS sale_cost,
					COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0) AS line_cost
				FROM sales s
				LEFT JOIN sale_items si ON si.sale_id = s.id
				WHERE %s >= date(%s)
				  AND %s < date(%s)
				  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
				GROUP BY s.id, s.cost_amount
			) costs`, saleDateExpr, startPlaceholder, saleDateExpr, endPlaceholder)
	}
	return fmt.Sprintf(`
		SELECT COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0)
		FROM sale_items si JOIN sales s ON s.id = si.sale_id
		WHERE %s >= date(%s)
		  AND %s < date(%s)
		  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, saleDateExpr, startPlaceholder, saleDateExpr, endPlaceholder)
}

// CreateReport creates a new report
func (r *Repository) CreateReport(ctx context.Context, report *Report) error {
	if strings.EqualFold(r.db.DriverName(), "sqlite") {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO reports (id, type, title, description, parameters,
				data, status, generated_by, generated_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			report.ID, report.Type, report.Title, report.Description, report.Parameters,
			report.Data, report.Status, report.GeneratedBy, report.GeneratedAt, report.CreatedAt, report.UpdatedAt)
		return err
	}
	query := `
		INSERT INTO reports (type, title, description, parameters,
			data, status, generated_by, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		report.Type, report.Title, report.Description, report.Parameters,
		report.Data, report.Status, report.GeneratedBy, report.GeneratedAt,
		report.CreatedAt, report.UpdatedAt,
	).Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}
	return nil
}

// GetReportByID retrieves a report by ID
func (r *Repository) GetReportByID(ctx context.Context, id uuid.UUID) (*Report, error) {
	var report Report
	query := `
		SELECT id, type, title, description, parameters,
			data, status, generated_by, generated_at, created_at, updated_at
		FROM reports
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &report, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReportNotFound
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	return &report, nil
}

// ListReports retrieves reports with pagination and filters
func (r *Repository) ListReports(ctx context.Context, req ReportListRequest) ([]Report, int, error) {
	var reports []Report
	var count int

	// Build base query
	baseQuery := `
		SELECT id, type, title, description, parameters,
			data, status, generated_by, generated_at, created_at, updated_at
		FROM reports
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM reports
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if req.Type != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, req.Type)
	}

	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}

	if req.GeneratedBy != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND generated_by = $%d", argCount)
		countQuery += fmt.Sprintf(" AND generated_by = $%d", argCount)
		args = append(args, *req.GeneratedBy)
	}

	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND generated_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND generated_at >= $%d", argCount)
		args = append(args, *req.StartDate)
	}

	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND generated_at <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND generated_at <= $%d", argCount)
		args = append(args, *req.EndDate)
	}

	if req.Search != "" {
		argCount++
		like := "ILIKE"
		if dbutil.IsSQLite(r.db) {
			like = "LIKE"
		}
		baseQuery += fmt.Sprintf(" AND (title %s $%d OR description %s $%d)", like, argCount, like, argCount)
		countQuery += fmt.Sprintf(" AND (title %s $%d OR description %s $%d)", like, argCount, like, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern)
	}

	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	// Add sorting
	sortColumns := map[string]string{
		"generated_at": "generated_at",
		"created_at":   "created_at",
		"updated_at":   "updated_at",
		"title":        "title",
		"type":         "type",
		"status":       "status",
	}
	sortBy := sortColumns[req.SortBy]
	if sortBy == "" {
		sortBy = "generated_at"
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.SortOrder, "ASC") {
		sortOrder = "ASC"
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)

	err = r.db.SelectContext(ctx, &reports, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	return reports, count, nil
}

// UpdateReport updates a report
func (r *Repository) UpdateReport(ctx context.Context, report *Report) error {
	if strings.EqualFold(r.db.DriverName(), "sqlite") {
		result, err := r.db.ExecContext(ctx, `UPDATE reports SET data = $2, status = $3, updated_at = $4 WHERE id = $1`, report.ID, report.Data, report.Status, report.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to update report: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return ErrReportNotFound
		}
		return nil
	}
	query := `
		UPDATE reports
		SET data = $2, status = $3, updated_at = $4
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		report.ID, report.Data, report.Status, report.UpdatedAt,
	).Scan(&report.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReportNotFound
		}
		return fmt.Errorf("failed to update report: %w", err)
	}
	return nil
}

// DeleteReport deletes a report
func (r *Repository) DeleteReport(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reports WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrReportNotFound
	}

	return nil
}

// GetUserName retrieves user name by ID
func (r *Repository) GetUserName(ctx context.Context, userID uuid.UUID) (string, error) {
	var name string
	query := `SELECT first_name || ' ' || last_name as name FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &name, query, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user name: %w", err)
	}
	return name, nil
}

// GetSalesData retrieves sales data for report
func (r *Repository) GetSalesData(ctx context.Context, startDate, endDate time.Time) (*SalesReport, error) {
	var report SalesReport
	report.StartDate = startDate
	report.EndDate = endDate

	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize sales report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize sales report end date: %w", err)
	}

	var totals struct {
		TotalSales     int     `db:"total_sales"`
		TotalRevenue   float64 `db:"total_revenue"`
		TotalTax       float64 `db:"total_tax"`
		TotalCOGS      float64 `db:"total_cogs"`
		TotalItemsSold int     `db:"total_items_sold"`
		CashRevenue    float64 `db:"cash_revenue"`
		CreditRevenue  float64 `db:"credit_revenue"`
		TotalPaid      float64 `db:"total_paid"`
		CashReceived   float64 `db:"cash_received"`
		ChangeAmount   float64 `db:"change_amount"`
	}

	if dbutil.IsSQLite(r.db) {
		saleDateExpr := r.salesDateExpression("")
		paidAmountExpr := "0"
		cashReceivedExpr := "0"
		changeAmountExpr := "0"
		if reportsSQLiteHasColumns(r.db, "sales", "paid_amount") {
			paidAmountExpr = "COALESCE(paid_amount, 0)"
		}
		if reportsSQLiteHasColumns(r.db, "sales", "cash_received") {
			cashReceivedExpr = "COALESCE(cash_received, 0)"
		}
		if reportsSQLiteHasColumns(r.db, "sales", "change_amount") {
			changeAmountExpr = "COALESCE(change_amount, 0)"
		}
		err = r.db.GetContext(ctx, &totals, fmt.Sprintf(`
			SELECT COUNT(*) AS total_sales,
				COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) AS total_revenue,
				COALESCE(SUM(COALESCE(tax_amount, 0)), 0) AS total_tax,
				COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('cash', 'card') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS cash_revenue,
				COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('credit', 'debt') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS credit_revenue,
				COALESCE(SUM(%s), 0) AS total_paid,
				COALESCE(SUM(%s), 0) AS cash_received,
				COALESCE(SUM(%s), 0) AS change_amount
			FROM sales
			WHERE %s >= date(?) AND %s < date(?)
				AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, paidAmountExpr, cashReceivedExpr, changeAmountExpr, saleDateExpr, saleDateExpr), startDateKey, endDateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve sales totals: %w", err)
		}

		err = r.db.GetContext(ctx, &totals.TotalCOGS, r.historicalCOGSTotalSQL("?", "?"), startDateKey, endDateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate sales COGS: %w", err)
		}

		err = r.db.GetContext(ctx, &totals.TotalItemsSold, fmt.Sprintf(`
			SELECT COALESCE(SUM(si.quantity), 0)
			FROM sale_items si
			JOIN sales s ON s.id = si.sale_id
			WHERE %s >= date(?) AND %s < date(?)
			  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, r.salesDateExpression("s"), r.salesDateExpression("s")), startDateKey, endDateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to count sold items: %w", err)
		}
	} else {
		itemCOGSQuery := r.historicalCOGSTotalSQL("$1", "$2")
		err = r.db.GetContext(ctx, &totals, fmt.Sprintf(`
			SELECT st.total_sales, st.total_revenue, st.total_tax, st.total_paid, st.cash_received, st.change_amount, it.total_cogs, it.total_items_sold,
				st.cash_revenue, st.credit_revenue
			FROM (
				SELECT COUNT(*) AS total_sales,
					COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) AS total_revenue,
					COALESCE(SUM(COALESCE(tax_amount, 0)), 0) AS total_tax,
					COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('cash', 'card') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS cash_revenue,
					COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('credit', 'debt') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS credit_revenue,
					COALESCE(SUM(COALESCE(paid_amount, 0)), 0) AS total_paid,
					COALESCE(SUM(COALESCE(cash_received, 0)), 0) AS cash_received,
					COALESCE(SUM(COALESCE(change_amount, 0)), 0) AS change_amount
				FROM sales
				WHERE date(COALESCE(sale_date, created_at)) >= date($1) AND date(COALESCE(sale_date, created_at)) < date($2)
					AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			) st
			CROSS JOIN (
				SELECT (%s) AS total_cogs,
					(SELECT COALESCE(SUM(si.quantity), 0)
						 FROM sale_items si JOIN sales s ON s.id = si.sale_id
						 WHERE date(COALESCE(s.sale_date, s.created_at)) >= date($3) AND date(COALESCE(s.sale_date, s.created_at)) < date($4)
						   AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) AS total_items_sold
			) it`,
			itemCOGSQuery),
			startDateKey, endDateKey, startDateKey, endDateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve sales totals: %w", err)
		}
	}

	// Completed customer returns reduce report revenue and sold-item cost by
	// the refunded amount and the cost of the returned units.
	var refundedAmount, returnedCost float64
	var returnedQuantity int
	returnQuantityColumn := ""
	if dbutil.IsSQLite(r.db) {
		if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity_returned") {
			returnQuantityColumn = "ri.quantity_returned"
		} else if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity") {
			returnQuantityColumn = "ri.quantity"
		}
	} else {
		returnQuantityColumn = "ri.quantity_returned"
	}
	if returnQuantityColumn != "" && reportsSQLiteHasColumns(r.db, "accounting_returns", "return_date", "status", "total_refund_amount") && reportsSQLiteHasColumns(r.db, "accounting_return_items", "total_refund_amount") {
		returnQuery := fmt.Sprintf(`
			SELECT COALESCE(SUM(r.total_refund_amount), 0),
				COALESCE(SUM(COALESCE(%s, 0) * COALESCE(ri.original_cost, si.unit_cost, p.cost_price, 0)), 0),
				COALESCE(SUM(COALESCE(%s, 0)), 0)
			FROM accounting_returns r
			JOIN accounting_return_items ri ON ri.return_id = r.id
			LEFT JOIN sale_items si ON si.id = ri.sale_item_id
			LEFT JOIN products p ON p.id = COALESCE(ri.product_id, si.product_id)
			WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
				AND COALESCE(r.total_refund_amount, 0) >= 0
				AND date(COALESCE(r.return_date, r.created_at)) >= date(?)
				AND date(COALESCE(r.return_date, r.created_at)) < date(?)`, returnQuantityColumn, returnQuantityColumn)
		if dbutil.IsSQLite(r.db) {
			_ = r.db.QueryRowContext(ctx, returnQuery, startDateKey, endDateKey).Scan(&refundedAmount, &returnedCost, &returnedQuantity)
		} else {
			_ = r.db.QueryRowContext(ctx, returnQuery, startDate, endDate).Scan(&refundedAmount, &returnedCost, &returnedQuantity)
		}
	}
	report.TotalSales = totals.TotalSales
	report.TotalRevenue = totals.TotalRevenue - refundedAmount
	report.TotalTax = totals.TotalTax
	report.TotalCOGS = totals.TotalCOGS - returnedCost
	report.TotalItemsSold = totals.TotalItemsSold - returnedQuantity
	if report.TotalRevenue < 0 {
		report.TotalRevenue = 0
	}
	if report.TotalCOGS < 0 {
		report.TotalCOGS = 0
	}
	if report.TotalItemsSold < 0 {
		report.TotalItemsSold = 0
	}
	report.CashRevenue = totals.CashRevenue
	report.CreditRevenue = totals.CreditRevenue
	report.TotalPaid = totals.TotalPaid
	report.CashReceived = totals.CashReceived
	report.ChangeAmount = totals.ChangeAmount
	report.GrossProfit = report.TotalRevenue - report.TotalCOGS
	if report.TotalRevenue > 0 {
		report.ProfitMargin = (report.GrossProfit / report.TotalRevenue) * 100
	} else {
		report.ProfitMargin = 0
	}

	dailySalesQuery := fmt.Sprintf(`SELECT DATE(sale_date) as date, COUNT(*) as sales, COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) as revenue
		 FROM sales 
		 WHERE %s >= date(?) AND %s < date(?)
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY DATE(sale_date)
		 ORDER BY date`, r.salesDateExpression(""), r.salesDateExpression(""))
	rows, err := r.db.QueryContext(ctx, r.db.Rebind(dailySalesQuery), startDateKey, endDateKey)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var daily DailySales
			var date reportTimestamp
			if err := rows.Scan(&date, &daily.Sales, &daily.Revenue); err != nil {
				continue
			}
			daily.Date = date.Time
			report.ByDay = append(report.ByDay, daily)
		}
		_ = rows.Err()
	}

	report.TopProducts = []ProductSales{}
	returnedJoin := ""
	returnedQuantityExpr := "0"
	returnedRefundExpr := "0"
	returnedCostExpr := "0"
	if !dbutil.IsSQLite(r.db) {
		returnedJoin = `LEFT JOIN (
			SELECT COALESCE(ri.product_id, si2.product_id, ii2.product_id) AS product_id,
				SUM(COALESCE(ri.quantity_returned, 0)) AS quantity,
				SUM(COALESCE(ri.total_refund_amount, 0)) AS refund_amount,
				SUM(COALESCE(ri.quantity_returned, 0) * COALESCE(ri.original_cost, si2.unit_cost, 0)) AS returned_cost
			FROM accounting_return_items ri
			JOIN accounting_returns r ON r.id = ri.return_id
			LEFT JOIN sale_items si2 ON si2.id = ri.sale_item_id
			LEFT JOIN inventory_items ii2 ON ii2.id = ri.inventory_item_id
			WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
				AND COALESCE(ri.total_refund_amount, 0) >= 0
				AND date(COALESCE(r.return_date, r.created_at)) >= date(?)
				AND date(COALESCE(r.return_date, r.created_at)) < date(?)
			GROUP BY ri.product_id
		) returned ON returned.product_id = si.product_id`
		returnedQuantityExpr = "COALESCE(MAX(returned.quantity), 0)"
		returnedRefundExpr = "COALESCE(MAX(returned.refund_amount), 0)"
		returnedCostExpr = "COALESCE(MAX(returned.returned_cost), 0)"
	} else {
		var returnItemsTableCount int
		_ = r.db.Get(&returnItemsTableCount, `SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table','view') AND name = 'accounting_return_items'`)
		returnQuantityColumn := ""
		if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity_returned") {
			returnQuantityColumn = "ri.quantity_returned"
		} else if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity") {
			returnQuantityColumn = "ri.quantity"
		}
		returnCostColumn := "0"
		if reportsSQLiteHasColumns(r.db, "accounting_return_items", "original_cost") {
			returnCostColumn = "ri.original_cost"
		}
		if returnItemsTableCount > 0 && returnQuantityColumn != "" && reportsSQLiteHasColumns(r.db, "accounting_return_items", "product_id", "total_refund_amount") && reportsSQLiteHasColumns(r.db, "accounting_returns", "return_date", "status", "total_refund_amount") {
			returnedJoin = fmt.Sprintf(`LEFT JOIN (
				SELECT COALESCE(ri.product_id, si2.product_id, ii2.product_id) AS product_id,
					SUM(COALESCE(%s, 0)) AS quantity,
					SUM(COALESCE(ri.total_refund_amount, 0)) AS refund_amount,
					SUM(COALESCE(%s, 0) * COALESCE(%s, si2.unit_cost, 0)) AS returned_cost
				FROM accounting_return_items ri
				JOIN accounting_returns r ON r.id = ri.return_id
				LEFT JOIN sale_items si2 ON si2.id = ri.sale_item_id
				LEFT JOIN inventory_items ii2 ON ii2.id = ri.inventory_item_id
				WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
					AND COALESCE(ri.total_refund_amount, 0) >= 0
					AND date(r.return_date) >= date(?)
					AND date(r.return_date) < date(?)
				GROUP BY ri.product_id
			) returned ON returned.product_id = si.product_id`, returnQuantityColumn, returnQuantityColumn, returnCostColumn)
			returnedQuantityExpr = "COALESCE(MAX(returned.quantity), 0)"
			returnedRefundExpr = "COALESCE(MAX(returned.refund_amount), 0)"
			returnedCostExpr = "COALESCE(MAX(returned.returned_cost), 0)"
		}
	}
	topProductsQuery := fmt.Sprintf(`
		SELECT p.id, p.name,
			COALESCE(SUM(si.quantity), 0) - %s,
			COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0) - %s,
			COALESCE(SUM((COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) - (si.quantity * COALESCE(si.unit_cost, 0))), 0)
				- %s + %s
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		JOIN products p ON p.id = si.product_id
		%s
		WHERE %s >= date(?) AND %s < date(?)
			AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		GROUP BY p.id, p.name
		ORDER BY (COALESCE(SUM((COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) - (si.quantity * COALESCE(si.unit_cost, 0))), 0)
			- %s + %s) DESC
		LIMIT 10`, returnedQuantityExpr, returnedRefundExpr, returnedRefundExpr, returnedCostExpr, returnedJoin, r.salesDateExpression("s"), r.salesDateExpression("s"), returnedRefundExpr, returnedCostExpr)
	returnArgs := []interface{}{startDateKey, endDateKey}
	if returnedJoin == "" {
		returnArgs = nil
	}
	returnArgs = append(returnArgs, startDateKey, endDateKey)
	rows, err = r.db.QueryContext(ctx, r.db.Rebind(topProductsQuery), returnArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve top products: %w", err)
	}
	for rows.Next() {
		var product ProductSales
		if err := rows.Scan(&product.ProductID, &product.ProductName, &product.Quantity, &product.Revenue, &product.Profit); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan top product: %w", err)
		}
		report.TopProducts = append(report.TopProducts, product)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read top products: %w", err)
	}
	rows.Close()

	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(payment_method, 'غير محدد'), COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) as total
		 FROM sales
		 WHERE date(sale_date) >= date(?) AND date(sale_date) < date(?)
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY payment_method`,
		startDateKey, endDateKey)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var method string
			var total float64
			if err := rows.Scan(&method, &total); err != nil {
				continue
			}
			report.ByPaymentMethod[method] = total
		}
		_ = rows.Err()
	}

	// Keep payment distribution consistent with net revenue by subtracting
	// completed customer refunds from the payment method of the original sale.
	refundByPaymentQuery := fmt.Sprintf(`
		SELECT COALESCE(s.payment_method, 'غير محدد'), COALESCE(SUM(r.total_refund_amount), 0)
		FROM accounting_returns r
		JOIN sales s ON s.id = r.sale_id
		WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
		  AND date(COALESCE(r.return_date, r.created_at)) >= date(?)
		  AND date(COALESCE(r.return_date, r.created_at)) < date(?)
		GROUP BY s.payment_method`)
	refundRows, refundErr := r.db.QueryContext(ctx, r.db.Rebind(refundByPaymentQuery), startDateKey, endDateKey)
	if refundErr == nil {
		defer refundRows.Close()
		for refundRows.Next() {
			var method string
			var refund float64
			if err := refundRows.Scan(&method, &refund); err != nil {
				continue
			}
			report.ByPaymentMethod[method] -= refund
			if report.ByPaymentMethod[method] <= 0 {
				delete(report.ByPaymentMethod, method)
			}
		}
	}

	return &report, nil
}

// GetInventoryData retrieves inventory data for report
func (r *Repository) GetInventoryData(ctx context.Context) (*InventoryReport, error) {
	var report InventoryReport
	storeNow := accounting.StoreNow()
	storeDate, err := accounting.StoreDate(storeNow)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate inventory report date: %w", err)
	}
	overstockStart, _, err := accounting.StoreDateRange(storeNow, 90)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate overstock report period: %w", err)
	}

	report.ByCondition = make(map[string]int)
	report.ByCategory = make(map[string]int)
	report.Items = []InventoryItem{}
	report.LowStockItems = []LowStockItem{}
	report.OverstockItems = []OverstockItem{}
	report.StagnantItems = []StagnantItem{}
	report.Valuation.ByCondition = make(map[string]float64)

	err = r.db.GetContext(ctx, &report.TotalItems,
		`SELECT COALESCE(SUM(stock), 0) FROM (
			SELECT p.id, COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) AS stock
			FROM products p
			LEFT JOIN inventory inv ON inv.product_id = p.id
			LEFT JOIN inventory_items ii ON ii.product_id = p.id
			WHERE p.deleted_at IS NULL
			GROUP BY p.id, inv.quantity
		) stock_totals WHERE stock > 0`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory totals: %w", err)
	}

	err = r.db.GetContext(ctx, &report.TotalValue,
		`SELECT COALESCE(SUM(stock * p.cost_price), 0) FROM (
			SELECT p.id, COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) AS stock
			FROM products p
			LEFT JOIN inventory inv ON inv.product_id = p.id
			LEFT JOIN inventory_items ii ON ii.product_id = p.id
			WHERE p.deleted_at IS NULL
			GROUP BY p.id, inv.quantity
		) stock_totals JOIN products p ON p.id = stock_totals.id WHERE stock > 0`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory value: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`WITH available_items AS (
		     SELECT product_id, COUNT(*) AS quantity
		     FROM inventory_items
		     WHERE status = 'AVAILABLE' AND UPPER(COALESCE(condition, '')) <> 'USED'
		     GROUP BY product_id
		 )
		 SELECT stock.condition, COALESCE(SUM(stock.quantity), 0),
		        COALESCE(SUM(stock.quantity * stock.unit_cost), 0)
		 FROM (
		     SELECT COALESCE(NULLIF(UPPER(ii.condition), ''), 'NEW') AS condition,
		            1 AS quantity, COALESCE(p.cost_price, 0) AS unit_cost
		     FROM inventory_items ii
		     JOIN products p ON p.id = ii.product_id
		     LEFT JOIN inventory inv ON inv.product_id = p.id
		     LEFT JOIN available_items ai ON ai.product_id = p.id
		     WHERE ii.status = 'AVAILABLE'
		       AND UPPER(COALESCE(ii.condition, '')) <> 'USED'
		       AND p.deleted_at IS NULL
		       AND (inv.quantity IS NULL OR ai.quantity <= inv.quantity)
		     UNION ALL
		     SELECT 'NEW' AS condition,
		            CASE WHEN COALESCE(ai.quantity, 0) <= inv.quantity
		                 THEN inv.quantity - COALESCE(ai.quantity, 0)
		                 ELSE inv.quantity END AS quantity,
		            COALESCE(p.cost_price, 0) AS unit_cost
		     FROM inventory inv
		     JOIN products p ON p.id = inv.product_id
		     LEFT JOIN available_items ai ON ai.product_id = p.id
		     WHERE inv.quantity > 0 AND p.deleted_at IS NULL
		 ) stock
		 WHERE stock.quantity > 0
		 GROUP BY stock.condition`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory conditions: %w", err)
	}
	for rows.Next() {
		var condition string
		var count int
		var value float64
		if err := rows.Scan(&condition, &count, &value); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan inventory condition: %w", err)
		}
		report.ByCondition[condition] = count
		report.Valuation.ByCondition[condition] = value
	}
	_ = rows.Err()
	rows.Close()
	report.Valuation.TotalCost = report.TotalValue
	err = r.db.GetContext(ctx, &report.Valuation.TotalRetail,
		`SELECT COALESCE(SUM(stock * p.selling_price), 0) FROM (
			SELECT p.id, COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) AS stock
			FROM products p
			LEFT JOIN inventory inv ON inv.product_id = p.id
			LEFT JOIN inventory_items ii ON ii.product_id = p.id
			WHERE p.deleted_at IS NULL
			GROUP BY p.id, inv.quantity
		) stock_totals JOIN products p ON p.id = stock_totals.id WHERE stock > 0`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve retail inventory value: %w", err)
	}
	var taxRate float64
	if err := r.db.GetContext(ctx, &taxRate, `SELECT COALESCE(CAST(value AS FLOAT), 0) FROM settings WHERE key = 'tax_rate'`); err != nil {
		taxRate = 0
	}
	report.Valuation.TotalRetailWithTax = report.Valuation.TotalRetail * (1 + taxRate/100)
	report.Valuation.PotentialProfit = report.Valuation.TotalRetail - report.Valuation.TotalCost

	rows, err = r.db.QueryContext(ctx,
		`SELECT p.id, p.name,
				COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0),
				p.min_stock_level
		 FROM products p
		 LEFT JOIN inventory inv ON inv.product_id = p.id
		 LEFT JOIN inventory_items ii ON ii.product_id = p.id
		 WHERE p.is_active = true AND p.deleted_at IS NULL
		 GROUP BY p.id, p.name, p.min_stock_level, inv.quantity
		 HAVING COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) > 0
		 ORDER BY p.name`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory items: %w", err)
	}
	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &item.MinStock); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		report.Items = append(report.Items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read inventory items: %w", err)
	}
	rows.Close()

	rows, err = r.db.QueryContext(ctx,
		`SELECT p.id, p.name,
					 COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0),
		        p.min_stock_level, p.min_stock_level
		 FROM products p
		 LEFT JOIN inventory inv ON inv.product_id = p.id
		 LEFT JOIN inventory_items ii ON ii.product_id = p.id
		 WHERE p.is_active = true AND p.deleted_at IS NULL AND p.min_stock_level > 0
		 GROUP BY p.id, p.name, p.min_stock_level, inv.quantity
			 HAVING COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) <= p.min_stock_level
			 ORDER BY (p.min_stock_level - COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0)) DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve low stock items: %w", err)
	}
	for rows.Next() {
		var item LowStockItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &item.MinStock, &item.ReorderLevel); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan low stock item: %w", err)
		}
		report.LowStockItems = append(report.LowStockItems, item)
	}
	_ = rows.Err()
	rows.Close()
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(c.name, 'غير مصنف'), COUNT(*)
		 FROM inventory_items ii
		 JOIN products p ON p.id = ii.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
			 WHERE ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' AND p.deleted_at IS NULL
		 GROUP BY c.name ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory categories: %w", err)
	}
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan inventory category: %w", err)
		}
		report.ByCategory[category] = count
	}
	_ = rows.Err()
	rows.Close()

	overstockQuery := `SELECT p.id, p.name,
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED'),
		        (SELECT COALESCE(SUM(si.quantity), 0) / 3 FROM sale_items si
		         JOIN sales s ON s.id = si.sale_id
		         WHERE si.product_id = p.id AND s.sale_date >= CURRENT_DATE - INTERVAL '90 days'
		           AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')),
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED') /
		        GREATEST((SELECT COALESCE(SUM(si.quantity), 0) / 3 FROM sale_items si
		                  JOIN sales s ON s.id = si.sale_id
		                  WHERE si.product_id = p.id AND s.sale_date >= CURRENT_DATE - INTERVAL '90 days'
		                    AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 1)
		 FROM products p
		 WHERE p.is_active = true AND p.deleted_at IS NULL
		   AND (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED') > 0
		   AND ((SELECT COALESCE(SUM(si.quantity), 0) FROM sale_items si
		         JOIN sales s ON s.id = si.sale_id
		         WHERE si.product_id = p.id AND s.sale_date >= CURRENT_DATE - INTERVAL '90 days'
		           AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) = 0 OR
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED') /
		        GREATEST((SELECT COALESCE(SUM(si.quantity), 0) / 3 FROM sale_items si
		                  JOIN sales s ON s.id = si.sale_id
		                  WHERE si.product_id = p.id AND s.sale_date >= CURRENT_DATE - INTERVAL '90 days'
		                    AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 1) >= 6)
		 ORDER BY 3 DESC`
	if dbutil.IsSQLite(r.db) {
		overstockQuery = `SELECT p.id, p.name,
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED'),
		        (SELECT COALESCE(SUM(si.quantity), 0) / 3.0 FROM sale_items si
		         JOIN sales s ON s.id = si.sale_id
				 WHERE si.product_id = p.id AND date(s.sale_date) >= date(?)
		           AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')),
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED') /
		        MAX((SELECT COALESCE(SUM(si.quantity), 0) / 3.0 FROM sale_items si
		                  JOIN sales s ON s.id = si.sale_id
				          WHERE si.product_id = p.id AND date(s.sale_date) >= date(?)
		                    AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 1)
		 FROM products p
		 WHERE p.is_active = 1 AND p.deleted_at IS NULL
		   AND (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED') > 0
		   AND ((SELECT COALESCE(SUM(si.quantity), 0) FROM sale_items si
		         JOIN sales s ON s.id = si.sale_id
			         WHERE si.product_id = p.id AND date(s.sale_date) >= date(?)
		           AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) = 0 OR
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED') /
		        MAX((SELECT COALESCE(SUM(si.quantity), 0) / 3.0 FROM sale_items si
		                  JOIN sales s ON s.id = si.sale_id
			          WHERE si.product_id = p.id AND date(s.sale_date) >= date(?)
		                    AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 1) >= 6)
		 ORDER BY 3 DESC`
	} else {
		overstockQuery = strings.ReplaceAll(overstockQuery, "s.sale_date >= CURRENT_DATE - INTERVAL '90 days'", "s.sale_date >= $1::date")
	}
	overstockArgs := []any{overstockStart}
	if dbutil.IsSQLite(r.db) {
		overstockArgs = []any{overstockStart, overstockStart, overstockStart, overstockStart}
	}
	rows, err = r.db.QueryContext(ctx, overstockQuery, overstockArgs...)
	if err == nil {
		for rows.Next() {
			var item OverstockItem
			if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &item.AvgMonthlySales, &item.MonthsOfSupply); err == nil {
				report.OverstockItems = append(report.OverstockItems, item)
			}
		}
		_ = rows.Err()
		rows.Close()
	}

	stagnantQuery := `SELECT p.id, p.name, COUNT(ii.id),
		        CASE WHEN MAX(latest_sale.last_sale_date) IS NULL OR MAX(latest_sale.last_sale_date) < MAX(ii.created_at)::date
		             THEN MAX(ii.created_at)::date ELSE MAX(latest_sale.last_sale_date) END,
		        CURRENT_DATE - CASE WHEN MAX(latest_sale.last_sale_date) IS NULL OR MAX(latest_sale.last_sale_date) < MAX(ii.created_at)::date
		             THEN MAX(ii.created_at)::date ELSE MAX(latest_sale.last_sale_date) END,
		        COALESCE(SUM(ii.purchase_cost), 0)
		 FROM products p
		 JOIN inventory_items ii ON ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED'
		 LEFT JOIN (
			SELECT si.product_id, MAX(s.sale_date) AS last_sale_date
			FROM sale_items si
			JOIN sales s ON s.id = si.sale_id
			WHERE LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY si.product_id
		 ) latest_sale ON latest_sale.product_id = p.id
		 WHERE p.is_active = true AND p.deleted_at IS NULL
		 GROUP BY p.id, p.name
		 HAVING CURRENT_DATE - CASE WHEN MAX(latest_sale.last_sale_date) IS NULL OR MAX(latest_sale.last_sale_date) < MAX(ii.created_at)::date
		             THEN MAX(ii.created_at)::date ELSE MAX(latest_sale.last_sale_date) END >= 30
		 ORDER BY CURRENT_DATE - CASE WHEN MAX(latest_sale.last_sale_date) IS NULL OR MAX(latest_sale.last_sale_date) < MAX(ii.created_at)::date
		             THEN MAX(ii.created_at)::date ELSE MAX(latest_sale.last_sale_date) END DESC`
	if dbutil.IsSQLite(r.db) {
		stagnantQuery = `SELECT p.id, p.name,
		        COUNT(ii.id),
		        CASE WHEN MAX(latest_sale.last_sale_date) IS NULL OR MAX(latest_sale.last_sale_date) < MAX(ii.created_at)
		             THEN MAX(ii.created_at) ELSE MAX(latest_sale.last_sale_date) END,
		        CAST(julianday(?) - julianday(CASE WHEN MAX(latest_sale.last_sale_date) IS NULL OR MAX(latest_sale.last_sale_date) < MAX(ii.created_at)
		             THEN MAX(ii.created_at) ELSE MAX(latest_sale.last_sale_date) END) AS INTEGER),
		        COALESCE(SUM(ii.purchase_cost), 0)
		 FROM products p
		 JOIN inventory_items ii ON ii.product_id = p.id AND ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED'
		 LEFT JOIN (
			SELECT si.product_id, MAX(s.sale_date) AS last_sale_date
			FROM sale_items si
			JOIN sales s ON s.id = si.sale_id
			WHERE LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY si.product_id
		 ) latest_sale ON latest_sale.product_id = p.id
		 WHERE p.is_active = 1 AND p.deleted_at IS NULL
		 GROUP BY p.id, p.name
		 HAVING julianday(?) - julianday(CASE WHEN MAX(latest_sale.last_sale_date) IS NULL OR MAX(latest_sale.last_sale_date) < MAX(ii.created_at)
		             THEN MAX(ii.created_at) ELSE MAX(latest_sale.last_sale_date) END) >= 30
		 ORDER BY julianday(?) - julianday(CASE WHEN MAX(latest_sale.last_sale_date) IS NULL OR MAX(latest_sale.last_sale_date) < MAX(ii.created_at)
		             THEN MAX(ii.created_at) ELSE MAX(latest_sale.last_sale_date) END) DESC`
	} else {
		stagnantQuery = strings.ReplaceAll(stagnantQuery, "CURRENT_DATE", "$1::date")
	}
	stagnantArgs := []any{storeDate}
	if dbutil.IsSQLite(r.db) {
		stagnantArgs = []any{storeDate, storeDate, storeDate}
	}
	rows, err = r.db.QueryContext(ctx, stagnantQuery, stagnantArgs...)
	if err == nil {
		for rows.Next() {
			var item StagnantItem
			var lastSale reportTimestamp
			if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &lastSale, &item.DaysSinceSale, &item.Value); err == nil {
				item.LastSaleDate = lastSale.Time
				report.StagnantItems = append(report.StagnantItems, item)
			}
		}
		_ = rows.Err()
		rows.Close()
	}

	return &report, nil
}

// GetExpensesData retrieves expenses data for report
func (r *Repository) GetExpensesData(ctx context.Context, startDate, endDate time.Time) (*ExpensesReport, error) {
	var report ExpensesReport
	report.StartDate = startDate
	report.EndDate = endDate
	report.ByCategory = make(map[string]float64)
	report.ByPaymentMethod = make(map[string]float64)
	report.ByMonth = []MonthlyExpenses{}
	report.TopExpenses = []ExpenseItem{}
	expensePeriod := "date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))"
	if dbutil.IsSQLite(r.db) {
		expensePeriod = "date(substr(expense_date, 1, 10)) >= date(substr($1, 1, 10)) AND date(substr(expense_date, 1, 10)) < date(substr($2, 1, 10))"
	}
	expenseStatus := accounting.AccountingExpenseStatusSQL()

	var err error
	report.TotalExpenses, err = accounting.AccountingExpensesForPeriod(ctx, r.db, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve expenses total: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT category, COALESCE(SUM(amount), 0) as total
		 FROM expenses
		 WHERE %s AND %s
		 GROUP BY category`, expensePeriod, expenseStatus),
		startDate, endDate)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var category string
			var total float64
			if err := rows.Scan(&category, &total); err != nil {
				continue
			}
			report.ByCategory[category] = total
		}
		_ = rows.Err()
	}

	rows, err = r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT payment_method, COALESCE(SUM(amount), 0) as total
		 FROM expenses
		 WHERE %s AND %s
		 GROUP BY payment_method`, expensePeriod, expenseStatus),
		startDate, endDate)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var method string
			var total float64
			if err := rows.Scan(&method, &total); err != nil {
				continue
			}
			report.ByPaymentMethod[method] = total
		}
		_ = rows.Err()
	}

	monthlyExpensesQuery := fmt.Sprintf(`SELECT DATE_TRUNC('month', expense_date), COALESCE(SUM(amount), 0)
		 FROM expenses
		 WHERE %s AND %s
		 GROUP BY DATE_TRUNC('month', expense_date)
		 ORDER BY DATE_TRUNC('month', expense_date)`, expensePeriod, expenseStatus)
	if dbutil.IsSQLite(r.db) {
		monthlyExpensesQuery = fmt.Sprintf(`SELECT strftime('%%Y-%%m-01', substr(expense_date, 1, 10)), COALESCE(SUM(amount), 0)
		 FROM expenses
		 WHERE %s AND %s
		 GROUP BY strftime('%%Y-%%m', substr(expense_date, 1, 10))
		 ORDER BY strftime('%%Y-%%m', substr(expense_date, 1, 10))`, expensePeriod, expenseStatus)
	}
	rows, err = r.db.QueryContext(ctx, monthlyExpensesQuery, startDate, endDate)
	if err == nil {
		for rows.Next() {
			var monthly MonthlyExpenses
			var month reportTimestamp
			if err := rows.Scan(&month, &monthly.Amount); err == nil {
				monthly.Month = month.Time
				report.ByMonth = append(report.ByMonth, monthly)
			}
		}
		_ = rows.Err()
		rows.Close()
	}

	topExpensesQuery := fmt.Sprintf(`SELECT expense_date, amount,
		COALESCE(NULLIF(description, ''), NULLIF(title, ''), 'مصروف') AS description,
		COALESCE(status, 'approved')
		FROM expenses
		WHERE %s AND %s
		ORDER BY date(expense_date) DESC, amount DESC
		LIMIT 20`, expensePeriod, expenseStatus)
	rows, err = r.db.QueryContext(ctx, topExpensesQuery, startDate, endDate)
	if err == nil {
		for rows.Next() {
			var expense ExpenseItem
			var expenseDate reportTimestamp
			if err := rows.Scan(&expenseDate, &expense.Amount, &expense.Description, &expense.Status); err == nil {
				expense.Date = expenseDate.Time
				report.TopExpenses = append(report.TopExpenses, expense)
			}
		}
		_ = rows.Err()
		rows.Close()
	}
	return &report, nil
}

// GetProfitsData retrieves profits data for report
func (r *Repository) GetProfitsData(ctx context.Context, startDate, endDate time.Time) (*ProfitsReport, error) {
	var report ProfitsReport
	report.StartDate = startDate
	report.EndDate = endDate

	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize profit report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize profit report end date: %w", err)
	}

	salesDateExpr := r.salesDateExpression("")
	if dbutil.IsSQLite(r.db) {
		err = r.db.GetContext(ctx, &report.TotalRevenue,
			fmt.Sprintf(`SELECT COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) FROM sales
			 WHERE %s >= date(?) AND %s < date(?)
			   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, salesDateExpr, salesDateExpr),
			startDateKey, endDateKey)
	} else {
		err = r.db.GetContext(ctx, &report.TotalRevenue,
			fmt.Sprintf(`SELECT COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) FROM sales
			 WHERE %s >= date($1) AND %s < date($2)
			   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, salesDateExpr, salesDateExpr),
			startDate, endDate)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve profit revenue: %w", err)
	}

	if dbutil.IsSQLite(r.db) {
		err = r.db.GetContext(ctx, &report.TotalCOGS, r.historicalCOGSTotalSQL("?", "?"), startDateKey, endDateKey)
	} else {
		err = r.db.GetContext(ctx, &report.TotalCOGS, r.historicalCOGSTotalSQL("$1", "$2"), startDate, endDate)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve profit cost: %w", err)
	}
	var refunded, returnedCost float64
	returnQuantityColumn := "ri.quantity_returned"
	if dbutil.IsSQLite(r.db) {
		if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity_returned") {
			returnQuantityColumn = "ri.quantity_returned"
		} else if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity") {
			returnQuantityColumn = "ri.quantity"
		}
	}
	returnAdjustmentQuery := fmt.Sprintf(`SELECT COALESCE(SUM(r.total_refund_amount), 0), COALESCE(SUM(%s * COALESCE(ri.original_cost, si.unit_cost, p.cost_price, 0)), 0) FROM accounting_returns r LEFT JOIN accounting_return_items ri ON ri.return_id = r.id LEFT JOIN sale_items si ON si.id = ri.sale_item_id LEFT JOIN products p ON p.id = ri.product_id WHERE date(COALESCE(r.refund_date, r.return_date, r.created_at)) >= date(substr($1, 1, 10)) AND date(COALESCE(r.refund_date, r.return_date, r.created_at)) < date(substr($2, 1, 10)) AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'`, returnQuantityColumn)
	if dbutil.IsSQLite(r.db) {
		returnAdjustmentQuery = fmt.Sprintf(`SELECT COALESCE(SUM(r.total_refund_amount), 0), COALESCE(SUM(%s * COALESCE(ri.original_cost, si.unit_cost, p.cost_price, 0)), 0) FROM accounting_returns r LEFT JOIN accounting_return_items ri ON ri.return_id = r.id LEFT JOIN sale_items si ON si.id = ri.sale_item_id LEFT JOIN products p ON p.id = ri.product_id WHERE date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)) >= date(substr($1, 1, 10)) AND date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)) < date(substr($2, 1, 10)) AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'`, returnQuantityColumn)
	}
	if dbutil.IsSQLite(r.db) {
		if err := r.db.QueryRowContext(ctx, returnAdjustmentQuery, startDateKey, endDateKey).Scan(&refunded, &returnedCost); err != nil {
			refunded, returnedCost = 0, 0
		}
	} else {
		if err := r.db.QueryRowContext(ctx, returnAdjustmentQuery, startDate, endDate).Scan(&refunded, &returnedCost); err != nil {
			refunded, returnedCost = 0, 0
		}
	}
	report.TotalRevenue -= refunded
	report.TotalCOGS -= returnedCost

	report.GrossProfit = report.TotalRevenue - report.TotalCOGS

	report.TotalExpenses, err = accounting.AccountingExpensesForPeriod(ctx, r.db, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve profit expenses: %w", err)
	}

	report.NetProfit = report.GrossProfit - report.TotalExpenses
	if report.TotalRevenue > 0 {
		report.ProfitMargin = (report.NetProfit / report.TotalRevenue) * 100
	} else {
		report.ProfitMargin = 0
	}
	report.ByMonth = []MonthlyProfit{}
	monthlyProfitsQuery := `WITH sales_by_month AS (
			        SELECT DATE_TRUNC('month', COALESCE(s.sale_date, s.created_at)) AS month,
		               COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
		        FROM sales s
		        WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		          AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		        GROUP BY DATE_TRUNC('month', COALESCE(s.sale_date, s.created_at))
		),
		cogs_by_month AS (
		        SELECT month, COALESCE(SUM(sale_cost), 0) AS cogs
		        FROM (
		                SELECT DATE_TRUNC('month', COALESCE(s.sale_date, s.created_at)) AS month, s.id,
		                       CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount
		                            ELSE COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0)
		                       END AS sale_cost
		                FROM sales s
		                LEFT JOIN sale_items si ON si.sale_id = s.id
		                WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		                  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		                GROUP BY DATE_TRUNC('month', COALESCE(s.sale_date, s.created_at)), s.id, s.cost_amount
		        ) costs
		        GROUP BY month
		),
		expenses_by_month AS (
		        SELECT DATE_TRUNC('month', expense_date) AS month,
		               COALESCE(SUM(amount), 0) AS expenses
		        FROM expenses
			        WHERE date(expense_date) >= date(?) AND date(expense_date) < date(?)
		          AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
		        GROUP BY DATE_TRUNC('month', expense_date)
		)
		SELECT s.month, s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
		FROM sales_by_month s
		LEFT JOIN cogs_by_month c ON c.month = s.month
		LEFT JOIN expenses_by_month e ON e.month = s.month
		ORDER BY s.month`
	if dbutil.IsSQLite(r.db) {
		monthlyProfitsQuery = `WITH sales_by_month AS (
		        SELECT strftime('%Y-%m', substr(COALESCE(s.sale_date, s.created_at), 1, 10)) AS month,
		               COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
		        FROM sales s
		        WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		          AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		        GROUP BY strftime('%Y-%m', substr(COALESCE(s.sale_date, s.created_at), 1, 10))
		),
		cogs_by_month AS (
		        SELECT month, COALESCE(SUM(sale_cost), 0) AS cogs
		        FROM (
		                SELECT strftime('%Y-%m', substr(COALESCE(s.sale_date, s.created_at), 1, 10)) AS month, s.id,
		                       CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount
		                            ELSE COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0)
		                       END AS sale_cost
		                FROM sales s
		                LEFT JOIN sale_items si ON si.sale_id = s.id
		                WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		                  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		                GROUP BY strftime('%Y-%m', substr(COALESCE(s.sale_date, s.created_at), 1, 10)), s.id, s.cost_amount
		        ) costs
		        GROUP BY month
		),
		expenses_by_month AS (
		        SELECT strftime('%Y-%m', expense_date) AS month,
		               COALESCE(SUM(amount), 0) AS expenses
		        FROM expenses
			        WHERE date(expense_date) >= date(?) AND date(expense_date) < date(?)
		          AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
		        GROUP BY strftime('%Y-%m', expense_date)
		)
		SELECT s.month || '-01', s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
		FROM sales_by_month s
		LEFT JOIN cogs_by_month c ON c.month = s.month
		LEFT JOIN expenses_by_month e ON e.month = s.month
		ORDER BY s.month`
		monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery,
			"date(expense_date) >= date(?) AND date(expense_date) < date(?)",
			"date(substr(expense_date, 1, 10)) >= date(?) AND date(substr(expense_date, 1, 10)) < date(?)")
		monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "strftime('%Y-%m', expense_date)", "strftime('%Y-%m', substr(expense_date, 1, 10))")
	}
	rows, err := r.db.QueryContext(ctx, r.db.Rebind(monthlyProfitsQuery),
		startDate, endDate,
		startDate, endDate,
		startDate, endDate)
	if err == nil {
		for rows.Next() {
			var month MonthlyProfit
			var cogs, expenses float64
			var monthDate reportTimestamp
			if err := rows.Scan(&monthDate, &month.Revenue, &cogs, &expenses); err == nil {
				month.Month = monthDate.Time
				month.COGS = cogs
				month.GrossProfit = month.Revenue - cogs
				month.Expenses = expenses
				month.NetProfit = month.GrossProfit - expenses
				report.ByMonth = append(report.ByMonth, month)
			}
		}
		_ = rows.Err()
		rows.Close()
	}
	if len(report.ByMonth) == 0 && report.TotalRevenue != 0 {
		report.ByMonth = append(report.ByMonth, MonthlyProfit{
			Month:       startDate,
			Revenue:     report.TotalRevenue,
			COGS:        report.TotalCOGS,
			GrossProfit: report.GrossProfit,
			Expenses:    report.TotalExpenses,
			NetProfit:   report.NetProfit,
		})
	}
	monthlyReturnQuery := `SELECT DATE_TRUNC('month', COALESCE(r.refund_date, r.return_date, r.created_at)), COALESCE(SUM(r.total_refund_amount), 0), COALESCE(SUM(ri.quantity_returned * COALESCE(ri.original_cost, si.unit_cost, p.cost_price, 0)), 0) FROM accounting_returns r LEFT JOIN accounting_return_items ri ON ri.return_id = r.id LEFT JOIN sale_items si ON si.id = ri.sale_item_id LEFT JOIN products p ON p.id = ri.product_id WHERE COALESCE(r.refund_date, r.return_date, r.created_at)::date >= $1::date AND COALESCE(r.refund_date, r.return_date, r.created_at)::date < $2::date AND UPPER(COALESCE(r.status, '')) = 'COMPLETED' GROUP BY DATE_TRUNC('month', COALESCE(r.refund_date, r.return_date, r.created_at))`
	if dbutil.IsSQLite(r.db) {
		monthlyReturnQuery = `SELECT strftime('%Y-%m-01', substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)), COALESCE(SUM(r.total_refund_amount), 0), COALESCE(SUM(ri.quantity_returned * COALESCE(ri.original_cost, si.unit_cost, p.cost_price, 0)), 0) FROM accounting_returns r LEFT JOIN accounting_return_items ri ON ri.return_id = r.id LEFT JOIN sale_items si ON si.id = ri.sale_item_id LEFT JOIN products p ON p.id = ri.product_id WHERE date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)) >= date(substr($1, 1, 10)) AND date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)) < date(substr($2, 1, 10)) AND UPPER(COALESCE(r.status, '')) = 'COMPLETED' GROUP BY strftime('%Y-%m-01', substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10))`
	}
	rows, err = r.db.QueryContext(ctx, monthlyReturnQuery, startDate, endDate)
	if err == nil {
		adjustments := map[string][2]float64{}
		for rows.Next() {
			var month reportTimestamp
			var monthRefund, monthReturnedCost float64
			if scanErr := rows.Scan(&month, &monthRefund, &monthReturnedCost); scanErr == nil {
				adjustments[month.Time.Format("2006-01")] = [2]float64{monthRefund, monthReturnedCost}
			}
		}
		rows.Close()
		for index := range report.ByMonth {
			adjustment := adjustments[report.ByMonth[index].Month.Format("2006-01")]
			report.ByMonth[index].Revenue -= adjustment[0]
			report.ByMonth[index].COGS -= adjustment[1]
			report.ByMonth[index].GrossProfit = report.ByMonth[index].Revenue - report.ByMonth[index].COGS
			report.ByMonth[index].NetProfit = report.ByMonth[index].GrossProfit - report.ByMonth[index].Expenses
		}
	}
	monthlyNetProfit := 0.0
	for _, month := range report.ByMonth {
		monthlyNetProfit += month.NetProfit
	}
	if len(report.ByMonth) == 0 || monthlyNetProfit-report.NetProfit > 0.01 || report.NetProfit-monthlyNetProfit > 0.01 {
		report.ByMonth = []MonthlyProfit{{
			Month:       startDate,
			Revenue:     report.TotalRevenue,
			COGS:        report.TotalCOGS,
			GrossProfit: report.GrossProfit,
			Expenses:    report.TotalExpenses,
			NetProfit:   report.NetProfit,
		}}
	}

	// Build the daily series from the same revenue, historical COGS, and
	// approved-expense sources used by the period totals.
	report.ByDay = []DailyProfit{}
	dailyProfitsQuery := `WITH sales_by_day AS (
		SELECT DATE(COALESCE(s.sale_date, s.created_at)) AS day,
		       COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
		FROM sales s
		WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		GROUP BY DATE(COALESCE(s.sale_date, s.created_at))
	), cogs_by_day AS (
		SELECT day, COALESCE(SUM(sale_cost), 0) AS cogs
		FROM (
			SELECT DATE(COALESCE(s.sale_date, s.created_at)) AS day, s.id,
			       CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount ELSE COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0) END AS sale_cost
			FROM sales s LEFT JOIN sale_items si ON si.sale_id = s.id
			WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
			  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY DATE(COALESCE(s.sale_date, s.created_at)), s.id, s.cost_amount
		) costs
		GROUP BY day
	), expenses_by_day AS (
		SELECT DATE(expense_date) AS day, COALESCE(SUM(amount), 0) AS expenses
		FROM expenses
		WHERE date(expense_date) >= date(?) AND date(expense_date) < date(?)
		  AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
		GROUP BY DATE(expense_date)
	)
	SELECT s.day, s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
	FROM sales_by_day s
	LEFT JOIN cogs_by_day c ON c.day = s.day
	LEFT JOIN expenses_by_day e ON e.day = s.day
	ORDER BY s.day`
	if dbutil.IsSQLite(r.db) {
		dailyProfitsQuery = `WITH sales_by_day AS (
			SELECT strftime('%Y-%m-%d', substr(COALESCE(s.sale_date, s.created_at), 1, 10)) AS day,
			       COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
			FROM sales s
			WHERE date(substr(COALESCE(s.sale_date, s.created_at), 1, 10)) >= date(?) AND date(substr(COALESCE(s.sale_date, s.created_at), 1, 10)) < date(?)
			  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY strftime('%Y-%m-%d', substr(COALESCE(s.sale_date, s.created_at), 1, 10))
		), cogs_by_day AS (
			SELECT day, COALESCE(SUM(sale_cost), 0) AS cogs
			FROM (
				SELECT strftime('%Y-%m-%d', substr(COALESCE(s.sale_date, s.created_at), 1, 10)) AS day, s.id,
				       CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount ELSE COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0) END AS sale_cost
				FROM sales s LEFT JOIN sale_items si ON si.sale_id = s.id
				WHERE date(substr(COALESCE(s.sale_date, s.created_at), 1, 10)) >= date(?) AND date(substr(COALESCE(s.sale_date, s.created_at), 1, 10)) < date(?)
				  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
				GROUP BY strftime('%Y-%m-%d', substr(COALESCE(s.sale_date, s.created_at), 1, 10)), s.id, s.cost_amount
			) costs
			GROUP BY day
		), expenses_by_day AS (
			SELECT strftime('%Y-%m-%d', substr(expense_date, 1, 10)) AS day, COALESCE(SUM(amount), 0) AS expenses
			FROM expenses
			WHERE date(substr(expense_date, 1, 10)) >= date(?) AND date(substr(expense_date, 1, 10)) < date(?)
			  AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
			GROUP BY strftime('%Y-%m-%d', substr(expense_date, 1, 10))
		)
		SELECT s.day, s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
		FROM sales_by_day s
		LEFT JOIN cogs_by_day c ON c.day = s.day
		LEFT JOIN expenses_by_day e ON e.day = s.day
		ORDER BY s.day`
	}
	dailyStartParam, dailyEndParam := any(startDate), any(endDate)
	if dbutil.IsSQLite(r.db) {
		dailyStartParam, dailyEndParam = startDateKey, endDateKey
	}
	dailyRows, dailyErr := r.db.QueryContext(ctx, r.db.Rebind(dailyProfitsQuery),
		dailyStartParam, dailyEndParam,
		dailyStartParam, dailyEndParam,
		dailyStartParam, dailyEndParam)
	if dailyErr == nil {
		for dailyRows.Next() {
			var day DailyProfit
			var cogs, expenses float64
			var dayDate reportTimestamp
			if scanErr := dailyRows.Scan(&dayDate, &day.Revenue, &cogs, &expenses); scanErr == nil {
				day.Date = dayDate.Time
				day.COGS = cogs
				day.Expenses = expenses
				day.NetProfit = day.Revenue - cogs - expenses
				report.ByDay = append(report.ByDay, day)
			} else {
				return nil, fmt.Errorf("failed to scan daily profit trend: %w", scanErr)
			}
		}
		if rowsErr := dailyRows.Err(); rowsErr != nil {
			return nil, fmt.Errorf("failed to read daily profit trend: %w", rowsErr)
		}
		dailyRows.Close()
	}
	dailyReturnQuery := strings.ReplaceAll(monthlyReturnQuery, "DATE_TRUNC('month',", "DATE(")
	dailyReturnQuery = strings.ReplaceAll(dailyReturnQuery, "strftime('%Y-%m-01',", "strftime('%Y-%m-%d',")
	dailyReturnQuery = strings.ReplaceAll(dailyReturnQuery, "GROUP BY strftime('%Y-%m-01',", "GROUP BY strftime('%Y-%m-%d',")
	if dailyReturnRows, queryErr := r.db.QueryContext(ctx, dailyReturnQuery, startDate, endDate); queryErr == nil {
		dailyAdjustments := map[string][2]float64{}
		for dailyReturnRows.Next() {
			var day reportTimestamp
			var refund, returnedCost float64
			if scanErr := dailyReturnRows.Scan(&day, &refund, &returnedCost); scanErr == nil {
				dailyAdjustments[day.Time.Format("2006-01-02")] = [2]float64{refund, returnedCost}
			}
		}
		dailyReturnRows.Close()
		for index := range report.ByDay {
			adjustment := dailyAdjustments[report.ByDay[index].Date.Format("2006-01-02")]
			report.ByDay[index].Revenue -= adjustment[0]
			report.ByDay[index].COGS -= adjustment[1]
			report.ByDay[index].NetProfit = report.ByDay[index].Revenue - report.ByDay[index].COGS - report.ByDay[index].Expenses
		}
	}
	report.ByCategory = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(NULLIF(TRIM(c.name), ''), p.name, 'غير مصنف'),
		        COALESCE(SUM((COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) - (si.quantity * COALESCE(si.unit_cost, 0))), 0)
		 FROM sale_items si
		 JOIN sales s ON s.id = si.sale_id
		 JOIN products p ON p.id = si.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		   AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY c.name, p.name ORDER BY SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) DESC`,
		startDate, endDate)
	if err == nil {
		for rows.Next() {
			var category string
			var profit float64
			if err := rows.Scan(&category, &profit); err == nil {
				report.ByCategory[category] = profit
			}
		}
		_ = rows.Err()
		rows.Close()
	}

	return &report, nil
}

// GetDebtsData retrieves debts data for report
func (r *Repository) GetDebtsData(ctx context.Context) (*DebtsReport, error) {
	var report DebtsReport
	report.ByCustomer = []CustomerDebt{}
	report.ByAge = make(map[string]int)
	report.PaymentHistory = []PaymentRecord{}
	storeDate, err := accounting.StoreDate(time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to calculate debt report date: %w", err)
	}

	err = r.db.GetContext(ctx, &report.TotalDebt,
		`SELECT COALESCE(SUM(amount), 0) FROM debts`)
	if err != nil {
		// If customers table doesn't exist, return empty report
		report.TotalDebt = 0
		report.TotalPaid = 0
		report.Outstanding = 0
		report.OverdueDebt = 0
		report.ByCustomer = []CustomerDebt{}
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalPaid,
		`SELECT COALESCE(SUM(amount - remaining_amount), 0) FROM debts`)
	if err != nil {
		report.TotalPaid = 0
	}

	err = r.db.GetContext(ctx, &report.Outstanding,
		`SELECT COALESCE(SUM(remaining_amount), 0) FROM debts`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve outstanding debt: %w", err)
	}
	overdueQuery := `SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE date(due_date) < date(?) AND remaining_amount > 0`
	overdueArgs := []any{storeDate}
	if !dbutil.IsSQLite(r.db) {
		overdueQuery = `SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE due_date::date < $1::date AND remaining_amount > 0`
	}
	err = r.db.GetContext(ctx, &report.OverdueDebt, overdueQuery, overdueArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve overdue debt: %w", err)
	}
	customerQuery := `SELECT d.customer_id, c.name, COALESCE(SUM(d.amount), 0),
		        COALESCE(SUM(d.amount - d.remaining_amount), 0),
		        COALESCE(SUM(d.remaining_amount), 0),
		        COALESCE(SUM(CASE WHEN date(d.due_date) < date(?) THEN d.remaining_amount ELSE 0 END), 0),
		        COALESCE(MAX(p.payment_date), '0001-01-01')
		 FROM debts d JOIN customers c ON c.id = d.customer_id
		 LEFT JOIN payments p ON p.customer_id = d.customer_id
		 GROUP BY d.customer_id, c.name ORDER BY SUM(d.remaining_amount) DESC`
	customerArgs := []any{storeDate}
	if !dbutil.IsSQLite(r.db) {
		customerQuery = strings.Replace(customerQuery, "date(d.due_date) < date(?)", "d.due_date::date < $1::date", 1)
	}
	rows, err := r.db.QueryContext(ctx, customerQuery, customerArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve debt customers: %w", err)
	}
	for rows.Next() {
		var item CustomerDebt
		var lastPayment reportTimestamp
		if err := rows.Scan(&item.CustomerID, &item.CustomerName, &item.TotalDebt, &item.PaidAmount, &item.Outstanding, &item.OverdueAmount, &lastPayment); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan debt customer: %w", err)
		}
		item.LastPayment = lastPayment.Time
		report.ByCustomer = append(report.ByCustomer, item)
	}
	_ = rows.Err()
	rows.Close()
	ageQuery := `SELECT CASE
			WHEN date(due_date) >= date(?) THEN 'current'
			WHEN julianday(?) - julianday(due_date) <= 7 THEN 'overdue_1_7'
			WHEN julianday(?) - julianday(due_date) <= 30 THEN 'overdue_8_30'
			ELSE 'overdue_30_plus' END, COUNT(*)
		 FROM debts WHERE remaining_amount > 0 GROUP BY 1`
	ageArgs := []any{storeDate, storeDate, storeDate}
	if dbutil.IsSQLite(r.db) {
		ageQuery = `SELECT CASE
			WHEN date(due_date) >= date(?) THEN 'current'
			WHEN julianday(?) - julianday(due_date) <= 7 THEN 'overdue_1_7'
			WHEN julianday(?) - julianday(due_date) <= 30 THEN 'overdue_8_30'
			ELSE 'overdue_30_plus' END, COUNT(*)
		 FROM debts WHERE remaining_amount > 0 GROUP BY 1`
	} else {
		ageQuery = `SELECT CASE
			WHEN due_date::date >= $1::date THEN 'current'
			WHEN $1::date - due_date::date <= 7 THEN 'overdue_1_7'
			WHEN $1::date - due_date::date <= 30 THEN 'overdue_8_30'
			ELSE 'overdue_30_plus' END, COUNT(*)
		 FROM debts WHERE remaining_amount > 0 GROUP BY 1`
		ageArgs = []any{storeDate}
	}
	rows, err = r.db.QueryContext(ctx, ageQuery, ageArgs...)
	if err == nil {
		for rows.Next() {
			var age string
			var count int
			if err := rows.Scan(&age, &count); err == nil {
				report.ByAge[age] = count
			}
		}
		_ = rows.Err()
		rows.Close()
	}
	rows, err = r.db.QueryContext(ctx,
		`SELECT p.payment_date, p.customer_id, COALESCE(c.name, 'عميل غير معروف'), p.amount
		 FROM payments p LEFT JOIN customers c ON c.id = p.customer_id
		 WHERE p.customer_id IS NOT NULL ORDER BY p.payment_date DESC LIMIT 100`)
	if err == nil {
		for rows.Next() {
			var payment PaymentRecord
			var paymentDate reportTimestamp
			if err := rows.Scan(&paymentDate, &payment.CustomerID, &payment.CustomerName, &payment.Amount, &payment.PaymentMethod, &payment.ReferenceNumber, &payment.Notes); err == nil {
				payment.Date = paymentDate.Time
				report.PaymentHistory = append(report.PaymentHistory, payment)
			}
		}
		_ = rows.Err()
		rows.Close()
	}
	return &report, nil
}

// GetPurchasesData retrieves purchases data for report
func (r *Repository) GetPurchasesData(ctx context.Context, startDate, endDate time.Time) (*PurchasesReport, error) {
	var report PurchasesReport
	report.StartDate = startDate
	report.EndDate = endDate
	_ = r.db.GetContext(ctx, &report.UsedPartPurchases, `
		SELECT
		  (SELECT COUNT(*) FROM acquisitions a
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
		     AND date(a.acquisition_date) <= date(substr($2, 1, 10))),
		  (SELECT COUNT(*) FROM acquisition_items ai JOIN acquisitions a ON a.id = ai.acquisition_id
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
		     AND date(a.acquisition_date) <= date(substr($2, 1, 10))),
		  (SELECT COALESCE(SUM(a.total_cost), 0) FROM acquisitions a
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
		     AND date(a.acquisition_date) <= date(substr($2, 1, 10))),
		  (SELECT COALESCE(SUM(a.paid_amount), 0) FROM acquisitions a
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
		     AND date(a.acquisition_date) <= date(substr($2, 1, 10))),
		  (SELECT COALESCE(SUM(a.total_cost - a.paid_amount), 0) FROM acquisitions a
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
		     AND date(a.acquisition_date) <= date(substr($2, 1, 10))),
		  (SELECT COUNT(*) FROM inventory_items ii JOIN acquisition_items ai ON ai.inventory_item_id = ii.id JOIN acquisitions a ON a.id = ai.acquisition_id
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed') AND ii.status = 'AVAILABLE'
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
		     AND date(a.acquisition_date) <= date(substr($2, 1, 10))),
		  (SELECT COUNT(*) FROM inventory_items ii JOIN acquisition_items ai ON ai.inventory_item_id = ii.id JOIN acquisitions a ON a.id = ai.acquisition_id
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed') AND ii.status = 'SOLD'
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
		     AND date(a.acquisition_date) <= date(substr($2, 1, 10)))`, startDate, endDate)
	itemTotalColumn := "total_amount"
	if dbutil.IsSQLite(r.db) {
		itemTotalColumn = "item_total"
	}

	err := r.db.GetContext(ctx, &report.TotalPurchases,
		`SELECT COUNT(*) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`,
		startDate, endDate)
	if err != nil {
		// If purchases table doesn't exist, return empty report
		report.TotalPurchases = 0
		report.TotalCost = 0
		report.BySupplier = []SupplierPurchases{}
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalCost,
		`SELECT COALESCE(SUM(total_amount), 0) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`,
		startDate, endDate)
	if err != nil {
		report.TotalCost = 0
	}
	_ = r.db.GetContext(ctx, &report.SupplierReturnCredits,
		`SELECT COALESCE(SUM(refund_amount), 0) FROM supplier_returns
		 WHERE date(created_at) >= date(substr($1, 1, 10)) AND date(created_at) < date(substr($2, 1, 10))
		   AND status = 'COMPLETED'`, startDate, endDate)
	report.NetPurchases = report.TotalCost - report.SupplierReturnCredits
	report.TaxAmount = 0
	if err := r.db.GetContext(ctx, &report.TaxAmount,
		fmt.Sprintf(`SELECT COALESCE(SUM(CASE
			WHEN COALESCE(tax_amount, 0) > 0 THEN tax_amount
			WHEN total_amount > COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount) + 0.01
				THEN total_amount - COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount)
			ELSE 0 END), 0) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, itemTotalColumn, itemTotalColumn),
		startDate, endDate); err != nil {
		report.TaxAmount = 0
	}
	var taxedPurchases int
	if err := r.db.GetContext(ctx, &taxedPurchases,
		fmt.Sprintf(`SELECT COUNT(*) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND CASE
				WHEN COALESCE(tax_amount, 0) > 0 THEN tax_amount
				WHEN total_amount > COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount) + 0.01
					THEN total_amount - COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount)
				ELSE 0 END > 0
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, itemTotalColumn, itemTotalColumn),
		startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to calculate taxed purchase count: %w", err)
	}
	var taxedPurchaseCost float64
	if err := r.db.GetContext(ctx, &taxedPurchaseCost,
		fmt.Sprintf(`SELECT COALESCE(SUM(total_amount), 0) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND CASE
				WHEN COALESCE(tax_amount, 0) > 0 THEN tax_amount
				WHEN total_amount > COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount) + 0.01
					THEN total_amount - COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount)
				ELSE 0 END > 0
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, itemTotalColumn, itemTotalColumn),
		startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to calculate taxed purchase cost: %w", err)
	}
	report.UntaxedPurchases = report.TotalPurchases - taxedPurchases
	report.UntaxedPurchaseCost = report.TotalCost - taxedPurchaseCost
	report.Subtotal = report.TotalCost - report.TaxAmount

	report.BySupplier = []SupplierPurchases{}
	rows, err := r.db.QueryContext(ctx,
		`SELECT p.supplier_id, COALESCE(s.name, 'مورد غير معروف'),
		        COALESCE(SUM(p.total_amount), 0), COALESCE(SUM(pi.item_count), 0)
		 FROM purchases p
		 LEFT JOIN suppliers s ON s.id = p.supplier_id
		 LEFT JOIN (
		   SELECT purchase_id, SUM(quantity) AS item_count FROM purchase_items GROUP BY purchase_id
		 ) pi ON pi.purchase_id = p.id
		 WHERE date(p.purchase_date) >= date(substr($1, 1, 10)) AND date(p.purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(p.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY p.supplier_id, s.name ORDER BY SUM(p.total_amount) DESC`,
		startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve purchases by supplier: %w", err)
	}
	for rows.Next() {
		var item SupplierPurchases
		if err := rows.Scan(&item.SupplierID, &item.SupplierName, &item.TotalCost, &item.ItemCount); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan supplier purchases: %w", err)
		}
		report.BySupplier = append(report.BySupplier, item)
	}
	_ = rows.Err()
	rows.Close()

	report.ByCategory = make(map[string]int)
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(c.name, 'غير مصنف'), COALESCE(SUM(pi.quantity), 0)
		 FROM purchase_items pi
		 JOIN purchases p ON p.id = pi.purchase_id
		 JOIN products pr ON pr.id = pi.product_id
		 LEFT JOIN categories c ON c.id = pr.category_id
		 WHERE date(p.purchase_date) >= date(substr($1, 1, 10)) AND date(p.purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(p.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY c.name ORDER BY SUM(pi.quantity) DESC`,
		startDate, endDate)
	if err == nil {
		for rows.Next() {
			var category string
			var count int
			if err := rows.Scan(&category, &count); err == nil {
				report.ByCategory[category] = count
			}
		}
		_ = rows.Err()
		rows.Close()
	}
	report.ByMonth = []MonthlyPurchases{}
	monthlyPurchasesQuery := `SELECT DATE_TRUNC('month', purchase_date), COALESCE(SUM(total_amount), 0), COUNT(*)
			 FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			 GROUP BY DATE_TRUNC('month', purchase_date)
			 ORDER BY DATE_TRUNC('month', purchase_date)`
	if dbutil.IsSQLite(r.db) {
		monthlyPurchasesQuery = `SELECT strftime('%Y-%m-01', purchase_date), COALESCE(SUM(total_amount), 0), COUNT(*)
			 FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY strftime('%Y-%m', purchase_date)
		 ORDER BY strftime('%Y-%m', purchase_date)`
	}
	rows, err = r.db.QueryContext(ctx, monthlyPurchasesQuery, startDate, endDate)
	if err == nil {
		for rows.Next() {
			var month MonthlyPurchases
			var monthDate reportTimestamp
			if err := rows.Scan(&monthDate, &month.Cost, &month.Count); err == nil {
				month.Month = monthDate.Time
				report.ByMonth = append(report.ByMonth, month)
			}
		}
		_ = rows.Err()
		rows.Close()
	}

	return &report, nil
}

// GetReturnsData retrieves returns data for report
func (r *Repository) GetReturnsData(ctx context.Context, startDate, endDate time.Time) (*ReturnsReport, error) {
	var report ReturnsReport
	report.StartDate = startDate
	report.EndDate = endDate

	// Only completed returns affect financial reports.
	err := r.db.GetContext(ctx, &report.TotalReturns,
		`SELECT COUNT(*) FROM accounting_returns
		 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
			   AND status = 'COMPLETED'
			   AND COALESCE(reference_number, '') NOT LIKE 'REV-%'`,
		startDate, endDate)
	if err != nil {
		// If returns table doesn't exist, return empty report
		report.TotalReturns = 0
		report.TotalRefunded = 0
		report.ByReason = make(map[string]int)
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalRefunded,
		`SELECT COALESCE(SUM(total_refund_amount), 0) FROM accounting_returns
		 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
			   AND status = 'COMPLETED'
			   AND COALESCE(reference_number, '') NOT LIKE 'REV-%'`,
		startDate, endDate)
	if err != nil {
		report.TotalRefunded = 0
	}

	report.ByReason = make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		`SELECT reason, COUNT(*) FROM accounting_returns
		 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
			   AND status = 'COMPLETED'
			   AND COALESCE(reference_number, '') NOT LIKE 'REV-%'
		 GROUP BY reason`,
		startDate, endDate)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var reason string
			var count int
			if err := rows.Scan(&reason, &count); err != nil {
				continue
			}
			report.ByReason[reason] = count
		}
		_ = rows.Err()
	}

	report.ByProduct = []ProductReturns{}
	rows, err = r.db.QueryContext(ctx,
		`SELECT ri.product_id, COALESCE(p.name, 'منتج محذوف'),
			        COUNT(DISTINCT r.id), COALESCE(SUM(ri.total_refund_amount), 0)
			 FROM accounting_returns r
			 JOIN accounting_return_items ri ON ri.return_id = r.id
			 LEFT JOIN products p ON p.id = ri.product_id
			 WHERE date(r.return_date) >= date(substr($1, 1, 10)) AND date(r.return_date) <= date(substr($2, 1, 10))
			   AND r.status = 'COMPLETED'
			   AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'
			 GROUP BY ri.product_id, p.name
			 ORDER BY SUM(ri.total_refund_amount) DESC`,
		startDate, endDate)
	if err == nil {
		for rows.Next() {
			var product ProductReturns
			if err := rows.Scan(&product.ProductID, &product.ProductName, &product.ReturnCount, &product.RefundAmount); err == nil {
				report.ByProduct = append(report.ByProduct, product)
			}
		}
		_ = rows.Err()
		rows.Close()
	}

	report.ByMonth = []MonthlyReturns{}
	monthlyReturnsQuery := `SELECT DATE_TRUNC('month', return_date), COUNT(*), COALESCE(SUM(total_refund_amount), 0)
			 FROM accounting_returns
			 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
			   AND status = 'COMPLETED'
			   AND COALESCE(reference_number, '') NOT LIKE 'REV-%'
			 GROUP BY strftime('%Y-%m', return_date)
			 ORDER BY strftime('%Y-%m', return_date)`
	if dbutil.IsSQLite(r.db) {
		monthlyReturnsQuery = `SELECT strftime('%Y-%m-01', return_date), COUNT(*), COALESCE(SUM(total_refund_amount), 0)
			 FROM accounting_returns
			 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
		   AND status = 'COMPLETED'
		   AND COALESCE(reference_number, '') NOT LIKE 'REV-%'
		 GROUP BY strftime('%Y-%m', return_date)
		 ORDER BY strftime('%Y-%m', return_date)`
	}
	rows, err = r.db.QueryContext(ctx, monthlyReturnsQuery, startDate, endDate)
	if err == nil {
		for rows.Next() {
			var monthly MonthlyReturns
			var month reportTimestamp
			if err := rows.Scan(&month, &monthly.Count, &monthly.Amount); err == nil {
				monthly.Month = month.Time
				report.ByMonth = append(report.ByMonth, monthly)
			}
		}
		_ = rows.Err()
		rows.Close()
	}

	return &report, nil
}

// GetNetSalesData retrieves net sales data for report (gross sales minus returns)
func (r *Repository) GetNetSalesData(ctx context.Context, startDate, endDate time.Time) (*NetSalesReport, error) {
	var report NetSalesReport
	report.StartDate = startDate
	report.EndDate = endDate

	// Get gross sales data
	var grossSales, grossRevenue float64
	err := r.db.GetContext(ctx, &grossSales,
		`SELECT COUNT(*) FROM sales WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) < date(substr($2, 1, 10))
			AND LOWER(COALESCE(status, 'completed')) = 'completed'`,
		startDate, endDate)
	if err != nil {
		// If sales table doesn't exist, return empty report
		report.GrossSales = 0
		report.GrossRevenue = 0
		report.TotalReturns = 0
		report.TotalRefunded = 0
		report.NetSales = 0
		report.NetRevenue = 0
		report.ReturnRate = 0
		return &report, nil
	}
	report.GrossSales = int(grossSales)

	err = r.db.GetContext(ctx, &grossRevenue,
		`SELECT COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) FROM sales WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) < date(substr($2, 1, 10))
			AND LOWER(COALESCE(status, 'completed')) = 'completed'`,
		startDate, endDate)
	if err != nil {
		grossRevenue = 0
	}
	report.GrossRevenue = grossRevenue

	// Get returns data using the completion/refund date when the schema supports it.
	returnDate := r.returnDateExpression("r")
	var totalReturns, totalRefunded float64
	err = r.db.GetContext(ctx, &totalReturns,
		fmt.Sprintf(`SELECT COUNT(*) FROM accounting_returns r
		 WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			   AND status = 'COMPLETED'
			   AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'`,
			returnDate, returnDate),
		startDate, endDate)
	if err != nil {
		totalReturns = 0
	}
	report.TotalReturns = int(totalReturns)

	err = r.db.GetContext(ctx, &totalRefunded,
		fmt.Sprintf(`SELECT COALESCE(SUM(total_refund_amount), 0) FROM accounting_returns r
		 WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			   AND status = 'COMPLETED'
			   AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'`,
			returnDate, returnDate),
		startDate, endDate)
	if err != nil {
		totalRefunded = 0
	}
	report.TotalRefunded = totalRefunded

	// Calculate net sales
	report.NetSales = report.GrossSales - report.TotalReturns
	report.NetRevenue = report.GrossRevenue - report.TotalRefunded

	// Calculate return rate
	if report.GrossSales > 0 {
		report.ReturnRate = (float64(report.TotalReturns) / float64(report.GrossSales)) * 100
	} else {
		report.ReturnRate = 0
	}

	// Get daily net sales data
	rows, err := r.db.QueryContext(ctx,
		`SELECT 
			DATE(sale_date) as date,
			COUNT(*) as gross_sales,
			COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) as gross_revenue,
			COALESCE((SELECT COUNT(*) FROM accounting_returns r
				 WHERE DATE(r.return_date) = DATE(s.sale_date)
				  AND date(r.return_date) >= date(substr($1, 1, 10)) AND date(r.return_date) < date(substr($2, 1, 10))
				  AND r.status = 'COMPLETED'
				  AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'), 0) as returns,
			COALESCE((SELECT SUM(total_refund_amount) FROM accounting_returns r
				 WHERE DATE(r.return_date) = DATE(s.sale_date)
				  AND date(r.return_date) >= date(substr($1, 1, 10)) AND date(r.return_date) < date(substr($2, 1, 10))
				  AND r.status = 'COMPLETED'
				  AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'), 0) as refunded
			 FROM sales s
			 WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) <= date(substr($2, 1, 10))
			   AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
			 GROUP BY DATE(s.sale_date)
		 ORDER BY date`,
		startDate, endDate)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var daily DailyNetSales
			var date reportTimestamp
			if err := rows.Scan(&date, &daily.GrossSales, &daily.GrossRevenue, &daily.Returns, &daily.Refunded); err != nil {
				continue
			}
			daily.Date = date.Time
			daily.NetSales = daily.GrossSales - daily.Returns
			daily.NetRevenue = daily.GrossRevenue - daily.Refunded
			report.ByDay = append(report.ByDay, daily)
		}
		_ = rows.Err()
	}

	// Get payment method breakdown for gross sales
	report.ByCategory = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(c.name, p.name, 'غير مصنف'), COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0) AS total
		 FROM sale_items si
		 JOIN sales s ON s.id = si.sale_id
		 JOIN products p ON p.id = si.product_id
			 LEFT JOIN categories c ON c.id = p.category_id
				 WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
			   AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
				 GROUP BY c.name, p.name
		 ORDER BY total DESC`,
		startDate, endDate)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var category string
			var total float64
			if err := rows.Scan(&category, &total); err != nil {
				continue
			}
			report.ByCategory[category] = total
		}
		_ = rows.Err()
	}

	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT payment_method, COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) as total
			 FROM sales
			 WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) <= date(substr($2, 1, 10))
			   AND LOWER(COALESCE(status, 'completed')) = 'completed'
			 GROUP BY payment_method`,
		startDate, endDate)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var method string
			var total float64
			if err := rows.Scan(&method, &total); err != nil {
				continue
			}
			report.ByPaymentMethod[method] = total
		}
		_ = rows.Err()
	}

	// Get top returned products with net sales analysis
	report.TopReturnedProducts = []ProductNetSales{}
	rows, err = r.db.QueryContext(ctx,
		`SELECT 
			p.id as product_id,
			p.name as product_name,
			COALESCE(SUM(si.quantity), 0) as gross_quantity,
			COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0) as gross_revenue,
			COALESCE((SELECT SUM(ri.quantity_returned) FROM accounting_return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM accounting_returns WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10)) AND status = 'COMPLETED')), 0) as returned_quantity,
			COALESCE((SELECT SUM(ri.total_refund_amount) FROM accounting_return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM accounting_returns WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10)) AND status = 'COMPLETED')), 0) as refunded_amount
		 FROM products p
			 LEFT JOIN sale_items si ON p.id = si.product_id AND si.sale_id IN (SELECT id FROM sales WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) <= date(substr($2, 1, 10)) AND LOWER(COALESCE(status, 'completed')) = 'completed')
		 WHERE p.is_active = true
		 GROUP BY p.id, p.name
		 HAVING COALESCE(SUM(si.quantity), 0) > 0 OR COALESCE((SELECT SUM(ri.quantity_returned) FROM accounting_return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM accounting_returns WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10)) AND status = 'COMPLETED')), 0) > 0
		 ORDER BY returned_quantity DESC
		 LIMIT 10`,
		startDate, endDate)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var product ProductNetSales
			if err := rows.Scan(&product.ProductID, &product.ProductName, &product.GrossQuantity, &product.GrossRevenue, &product.ReturnedQuantity, &product.RefundedAmount); err != nil {
				continue
			}
			product.NetQuantity = product.GrossQuantity - product.ReturnedQuantity
			product.NetRevenue = product.GrossRevenue - product.RefundedAmount
			if product.GrossQuantity > 0 {
				product.ReturnRate = (float64(product.ReturnedQuantity) / float64(product.GrossQuantity)) * 100
			} else {
				product.ReturnRate = 0
			}
			report.TopReturnedProducts = append(report.TopReturnedProducts, product)
		}
		_ = rows.Err()
	}

	return &report, nil
}

func (r *Repository) GetTaxData(ctx context.Context, startDate, endDate time.Time) (*TaxReport, error) {
	report := &TaxReport{StartDate: startDate, EndDate: endDate, Period: startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02")}
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN subtotal ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN discount_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN subtotal - discount_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN tax_amount ELSE 0 END), 0),
			COALESCE(SUM(total_amount), 0)
		FROM sales
		WHERE date(sale_date) >= date(substr($1, 1, 10))
		  AND date(sale_date) <= date(substr($2, 1, 10))
		  AND LOWER(COALESCE(status, 'completed')) = 'completed'`
	if err := r.db.QueryRowxContext(ctx, query, startDate, endDate).Scan(&report.GrossSales, &report.Discounts, &report.TaxableSales, &report.TaxCollected, &report.SalesTotal); err != nil {
		return nil, err
	}
	report.ExemptSales = report.SalesTotal - report.TaxableSales - report.TaxCollected
	if report.ExemptSales < 0 && report.ExemptSales > -0.01 {
		report.ExemptSales = 0
	}
	returnDate := r.returnDateExpression("r")
	if err := r.db.GetContext(ctx, &report.ReturnsTotal, fmt.Sprintf(`SELECT COALESCE(SUM(total_refund_amount), 0) FROM accounting_returns r WHERE %s >= date(substr($1, 1, 10)) AND %s <= date(substr($2, 1, 10)) AND status = 'COMPLETED'`, returnDate, returnDate), startDate, endDate); err != nil {
		report.ReturnsTotal = 0
	}
	canCalculateReturnedTax := !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_return_items", "return_id", "sale_item_id", "total_refund_amount") && reportsSQLiteHasColumns(r.db, "sale_items", "id", "tax_amount", "total_amount")
	returnedTaxable := 0.0
	if canCalculateReturnedTax {
		var returnedAmounts struct {
			Taxable float64 `db:"taxable"`
			Tax     float64 `db:"tax"`
		}
		if err := r.db.GetContext(ctx, &returnedAmounts, fmt.Sprintf(`
		SELECT
			COALESCE(SUM(CASE WHEN COALESCE(si.tax_amount, 0) > 0 THEN COALESCE(ri.total_refund_amount, 0) * COALESCE((si.total_amount - si.tax_amount) / NULLIF(si.total_amount, 0), 0) ELSE 0 END), 0) AS taxable,
			COALESCE(SUM(CASE WHEN COALESCE(si.tax_amount, 0) > 0 THEN COALESCE(ri.total_refund_amount, 0) * COALESCE(si.tax_amount / NULLIF(si.total_amount, 0), 0) ELSE 0 END), 0) AS tax
		FROM accounting_returns r
		JOIN accounting_return_items ri ON ri.return_id = r.id
		LEFT JOIN sale_items si ON si.id = ri.sale_item_id
		WHERE %s >= date(substr($1, 1, 10))
		  AND %s <= date(substr($2, 1, 10))
		  AND r.status = 'COMPLETED'`, returnDate, returnDate), startDate, endDate); err != nil {
			report.ReturnedTax = 0
		} else {
			report.ReturnedTax = returnedAmounts.Tax
			returnedTaxable = returnedAmounts.Taxable
		}
	}
	report.NetTaxableSales = report.TaxableSales - returnedTaxable
	report.NetTaxCollected = report.TaxCollected - report.ReturnedTax
	report.NetSalesTotal = report.SalesTotal - report.ReturnsTotal
	return report, nil
}
