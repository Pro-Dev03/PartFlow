package reports

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles report data operations
type Repository struct {
	db *sqlx.DB
}

func (r *Repository) returnDateExpression(alias string) string {
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "returns", "refund_date", "updated_at") {
		return fmt.Sprintf("date(%s.return_date)", alias)
	}
	return fmt.Sprintf("date(COALESCE(%s.refund_date, %s.updated_at, %s.return_date))", alias, alias, alias)
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
	return &Repository{db: db}
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

	var totals struct {
		TotalSales     int     `db:"total_sales"`
		TotalRevenue   float64 `db:"total_revenue"`
		TotalTax       float64 `db:"total_tax"`
		TotalCOGS      float64 `db:"total_cogs"`
		TotalItemsSold int     `db:"total_items_sold"`
		CashRevenue    float64 `db:"cash_revenue"`
		CreditRevenue  float64 `db:"credit_revenue"`
	}
	err := r.db.GetContext(ctx, &totals, `
		SELECT st.total_sales, st.total_revenue, st.total_tax, it.total_cogs, it.total_items_sold,
			st.cash_revenue, st.credit_revenue
		FROM (
			SELECT COUNT(*) AS total_sales,
				COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) AS total_revenue,
				COALESCE(SUM(COALESCE(tax_amount, 0)), 0) AS total_tax,
				COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('cash', 'card') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS cash_revenue,
				COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('credit', 'debt') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS credit_revenue
			FROM sales
			WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) < date(substr($2, 1, 10))
				AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		) st
		CROSS JOIN (
			SELECT COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0) AS total_cogs,
				COALESCE(SUM(si.quantity), 0) AS total_items_sold
			FROM sale_items si
			JOIN sales s ON s.id = si.sale_id
			WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
				AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		) it`,
		startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sales totals: %w", err)
	}
	report.TotalSales = totals.TotalSales
	report.TotalRevenue = totals.TotalRevenue
	report.TotalTax = totals.TotalTax
	report.TotalCOGS = totals.TotalCOGS
	report.TotalItemsSold = totals.TotalItemsSold
	report.CashRevenue = totals.CashRevenue
	report.CreditRevenue = totals.CreditRevenue
	report.GrossProfit = report.TotalRevenue - report.TotalCOGS
	if report.TotalRevenue > 0 {
		report.ProfitMargin = (report.GrossProfit / report.TotalRevenue) * 100
	} else {
		report.ProfitMargin = 0
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT DATE(sale_date) as date, COUNT(*) as sales, COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) as revenue
		 FROM sales 
		 WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY DATE(sale_date)
		 ORDER BY date`,
		startDate, endDate)
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
	rows, err = r.db.QueryContext(ctx, `
		SELECT p.id, p.name, COALESCE(SUM(si.quantity), 0),
			COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0),
			COALESCE(SUM((COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) - (si.quantity * COALESCE(si.unit_cost, 0))), 0)
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		JOIN products p ON p.id = si.product_id
		WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
			AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		GROUP BY p.id, p.name
		ORDER BY SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) DESC
		LIMIT 10`, startDate, endDate)
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
		`SELECT COALESCE(payment_method, 'ØºÙŠØ± Ù…Ø­Ø¯Ø¯'), COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) as total
		 FROM sales
		 WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
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

	return &report, nil
}

// GetInventoryData retrieves inventory data for report
func (r *Repository) GetInventoryData(ctx context.Context) (*InventoryReport, error) {
	var report InventoryReport

	report.ByCondition = make(map[string]int)
	report.ByCategory = make(map[string]int)
	report.LowStockItems = []LowStockItem{}
	report.OverstockItems = []OverstockItem{}
	report.StagnantItems = []StagnantItem{}
	report.Valuation.ByCondition = make(map[string]float64)

	err := r.db.GetContext(ctx, &report.TotalItems,
		`SELECT COUNT(*) FROM inventory_items ii
		 JOIN products p ON p.id = ii.product_id
		 WHERE ii.status = 'AVAILABLE' AND p.deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory totals: %w", err)
	}

	err = r.db.GetContext(ctx, &report.TotalValue,
		`SELECT COALESCE(SUM(ii.purchase_cost), 0) FROM inventory_items ii
		 JOIN products p ON p.id = ii.product_id
		 WHERE ii.status = 'AVAILABLE' AND p.deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory value: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT ii.condition, COUNT(*), COALESCE(SUM(ii.purchase_cost), 0)
		 FROM inventory_items ii
		 JOIN products p ON p.id = ii.product_id
		 WHERE ii.status = 'AVAILABLE' AND p.deleted_at IS NULL
		 GROUP BY ii.condition`)
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
		`SELECT COALESCE(SUM(ROUND(p.selling_price)), 0) FROM inventory_items ii
		 JOIN products p ON p.id = ii.product_id
			WHERE ii.status = 'AVAILABLE' AND ii.condition <> 'USED' AND p.deleted_at IS NULL`)
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
				 COALESCE(SUM(CASE WHEN ii.status = 'AVAILABLE' AND ii.condition <> 'USED' THEN 1 ELSE 0 END), 0),
		        p.min_stock_level, p.min_stock_level
		 FROM products p
		 LEFT JOIN inventory_items ii ON ii.product_id = p.id
		 WHERE p.is_active = true AND p.deleted_at IS NULL AND p.min_stock_level > 0
		 GROUP BY p.id, p.name, p.min_stock_level
			 HAVING COALESCE(SUM(CASE WHEN ii.status = 'AVAILABLE' AND ii.condition <> 'USED' THEN 1 ELSE 0 END), 0) < p.min_stock_level
			 ORDER BY (p.min_stock_level - COALESCE(SUM(CASE WHEN ii.status = 'AVAILABLE' AND ii.condition <> 'USED' THEN 1 ELSE 0 END), 0)) DESC`)
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
		`SELECT COALESCE(c.name, 'ØºÙŠØ± Ù…ØµÙ†Ù'), COUNT(*)
		 FROM inventory_items ii
		 JOIN products p ON p.id = ii.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE ii.status = 'AVAILABLE' AND p.deleted_at IS NULL
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
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE'),
		        (SELECT COALESCE(SUM(si.quantity), 0) / 3 FROM sale_items si
		         JOIN sales s ON s.id = si.sale_id
		         WHERE si.product_id = p.id AND s.sale_date >= CURRENT_DATE - INTERVAL '90 days'
		           AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')),
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE') /
		        GREATEST((SELECT COALESCE(SUM(si.quantity), 0) / 3 FROM sale_items si
		                  JOIN sales s ON s.id = si.sale_id
		                  WHERE si.product_id = p.id AND s.sale_date >= CURRENT_DATE - INTERVAL '90 days'
		                    AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 1)
		 FROM products p
		 WHERE p.is_active = true AND p.deleted_at IS NULL
		   AND (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE') > 0
		   AND ((SELECT COALESCE(SUM(si.quantity), 0) FROM sale_items si
		         JOIN sales s ON s.id = si.sale_id
		         WHERE si.product_id = p.id AND s.sale_date >= CURRENT_DATE - INTERVAL '90 days'
		           AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) = 0 OR
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE') /
		        GREATEST((SELECT COALESCE(SUM(si.quantity), 0) / 3 FROM sale_items si
		                  JOIN sales s ON s.id = si.sale_id
		                  WHERE si.product_id = p.id AND s.sale_date >= CURRENT_DATE - INTERVAL '90 days'
		                    AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 1) >= 6)
		 ORDER BY 3 DESC`
	if dbutil.IsSQLite(r.db) {
		overstockQuery = `SELECT p.id, p.name,
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE'),
		        (SELECT COALESCE(SUM(si.quantity), 0) / 3.0 FROM sale_items si
		         JOIN sales s ON s.id = si.sale_id
		         WHERE si.product_id = p.id AND s.sale_date >= datetime('now', '-90 days')
		           AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')),
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE') /
		        MAX((SELECT COALESCE(SUM(si.quantity), 0) / 3.0 FROM sale_items si
		                  JOIN sales s ON s.id = si.sale_id
		                  WHERE si.product_id = p.id AND s.sale_date >= datetime('now', '-90 days')
		                    AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 1)
		 FROM products p
		 WHERE p.is_active = 1 AND p.deleted_at IS NULL
		   AND (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE') > 0
		   AND ((SELECT COALESCE(SUM(si.quantity), 0) FROM sale_items si
		         JOIN sales s ON s.id = si.sale_id
		         WHERE si.product_id = p.id AND s.sale_date >= datetime('now', '-90 days')
		           AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) = 0 OR
		        (SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE') /
		        MAX((SELECT COALESCE(SUM(si.quantity), 0) / 3.0 FROM sale_items si
		                  JOIN sales s ON s.id = si.sale_id
		                  WHERE si.product_id = p.id AND s.sale_date >= datetime('now', '-90 days')
		                    AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 1) >= 6)
		 ORDER BY 3 DESC`
	}
	rows, err = r.db.QueryContext(ctx, overstockQuery)
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

	stagnantQuery := `SELECT p.id, p.name, COUNT(ii.id), COALESCE(MAX(s.sale_date), DATE '0001-01-01'),
		        CURRENT_DATE - COALESCE(MAX(s.sale_date), DATE '0001-01-01'),
		        COALESCE(SUM(ii.purchase_cost), 0)
		 FROM products p
		 JOIN inventory_items ii ON ii.product_id = p.id AND ii.status = 'AVAILABLE'
		 LEFT JOIN sale_items si ON si.product_id = p.id
		 LEFT JOIN sales s ON s.id = si.sale_id
		   AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 WHERE p.is_active = true AND p.deleted_at IS NULL
		 GROUP BY p.id, p.name
		 HAVING CURRENT_DATE - COALESCE(MAX(s.sale_date), DATE '0001-01-01') >= 30
		 ORDER BY CURRENT_DATE - COALESCE(MAX(s.sale_date), DATE '0001-01-01') DESC`
	if dbutil.IsSQLite(r.db) {
		stagnantQuery = `SELECT p.id, p.name, COUNT(ii.id), COALESCE(MAX(s.sale_date), '0001-01-01T00:00:00Z'),
		        CAST(julianday('now') - julianday(COALESCE(MAX(s.sale_date), '0001-01-01T00:00:00Z')) AS INTEGER),
		        COALESCE(SUM(ii.purchase_cost), 0)
		 FROM products p
		 JOIN inventory_items ii ON ii.product_id = p.id AND ii.status = 'AVAILABLE'
		 LEFT JOIN sale_items si ON si.product_id = p.id
		 LEFT JOIN sales s ON s.id = si.sale_id
		   AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 WHERE p.is_active = 1 AND p.deleted_at IS NULL
		 GROUP BY p.id, p.name
		 HAVING julianday('now') - julianday(COALESCE(MAX(s.sale_date), '0001-01-01T00:00:00Z')) >= 30
		 ORDER BY julianday('now') - julianday(COALESCE(MAX(s.sale_date), '0001-01-01T00:00:00Z')) DESC`
	}
	rows, err = r.db.QueryContext(ctx, stagnantQuery)
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

	err := r.db.GetContext(ctx, &report.TotalExpenses,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses
		 WHERE date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')`,
		startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve expenses total: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT category, COALESCE(SUM(amount), 0) as total
		 FROM expenses
		 WHERE date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
		 GROUP BY category`,
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
		`SELECT payment_method, COALESCE(SUM(amount), 0) as total
		 FROM expenses
		 WHERE date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
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

	monthlyExpensesQuery := `SELECT DATE_TRUNC('month', expense_date), COALESCE(SUM(amount), 0)
		 FROM expenses
		 WHERE date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
		 GROUP BY DATE_TRUNC('month', expense_date)
		 ORDER BY DATE_TRUNC('month', expense_date)`
	if dbutil.IsSQLite(r.db) {
		monthlyExpensesQuery = `SELECT strftime('%Y-%m-01', expense_date), COALESCE(SUM(amount), 0)
		 FROM expenses
		 WHERE date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
		 GROUP BY strftime('%Y-%m', expense_date)
		 ORDER BY strftime('%Y-%m', expense_date)`
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
	return &report, nil
}

// GetProfitsData retrieves profits data for report
func (r *Repository) GetProfitsData(ctx context.Context, startDate, endDate time.Time) (*ProfitsReport, error) {
	var report ProfitsReport
	report.StartDate = startDate
	report.EndDate = endDate

	err := r.db.GetContext(ctx, &report.TotalRevenue,
		`SELECT COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) FROM sales
		 WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`,
		startDate, endDate)
	if err != nil {
		// If sales table doesn't exist, return empty report
		report.TotalRevenue = 0
		report.TotalCOGS = 0
		report.GrossProfit = 0
		report.TotalExpenses = 0
		report.NetProfit = 0
		report.ProfitMargin = 0
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalCOGS,
		`SELECT COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0)
		 FROM sale_items si JOIN sales s ON s.id = si.sale_id
		 WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`,
		startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve profit cost: %w", err)
	}

	report.GrossProfit = report.TotalRevenue - report.TotalCOGS

	err = r.db.GetContext(ctx, &report.TotalExpenses,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses
		 WHERE date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')`,
		startDate, endDate)
	if err != nil {
		report.TotalExpenses = 0
	}

	report.NetProfit = report.GrossProfit - report.TotalExpenses
	if report.TotalRevenue > 0 {
		report.ProfitMargin = (report.NetProfit / report.TotalRevenue) * 100
	} else {
		report.ProfitMargin = 0
	}
	report.ByMonth = []MonthlyProfit{}
	monthlyProfitsQuery := `WITH sales_by_month AS (
		        SELECT DATE_TRUNC('month', s.sale_date) AS month,
		               COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
		        FROM sales s
		        WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
		          AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		        GROUP BY DATE_TRUNC('month', s.sale_date)
		),
		cogs_by_month AS (
		        SELECT DATE_TRUNC('month', s.sale_date) AS month,
		               COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0) AS cogs
		        FROM sales s
		        JOIN sale_items si ON si.sale_id = s.id
		        WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
		          AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		        GROUP BY DATE_TRUNC('month', s.sale_date)
		),
		expenses_by_month AS (
		        SELECT DATE_TRUNC('month', expense_date) AS month,
		               COALESCE(SUM(amount), 0) AS expenses
		        FROM expenses
		        WHERE date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))
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
		        SELECT strftime('%Y-%m', s.sale_date) AS month,
		               COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
		        FROM sales s
		        WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
		          AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		        GROUP BY strftime('%Y-%m', s.sale_date)
		),
		cogs_by_month AS (
		        SELECT strftime('%Y-%m', s.sale_date) AS month,
		               COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0) AS cogs
		        FROM sales s JOIN sale_items si ON si.sale_id = s.id
		        WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
		          AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		        GROUP BY strftime('%Y-%m', s.sale_date)
		),
		expenses_by_month AS (
		        SELECT strftime('%Y-%m', expense_date) AS month,
		               COALESCE(SUM(amount), 0) AS expenses
		        FROM expenses
		        WHERE date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))
		          AND LOWER(COALESCE(status, 'approved')) NOT IN ('rejected', 'cancelled', 'canceled')
		        GROUP BY strftime('%Y-%m', expense_date)
		)
		SELECT s.month || '-01', s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
		FROM sales_by_month s
		LEFT JOIN cogs_by_month c ON c.month = s.month
		LEFT JOIN expenses_by_month e ON e.month = s.month
		ORDER BY s.month`
	}
	rows, err := r.db.QueryContext(ctx, monthlyProfitsQuery, startDate, endDate)
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

	report.ByCategory = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(c.name, 'ØºÙŠØ± Ù…ØµÙ†Ù'),
		        COALESCE(SUM((COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) - (si.quantity * COALESCE(si.unit_cost, 0))), 0)
		 FROM sale_items si
		 JOIN sales s ON s.id = si.sale_id
		 JOIN products p ON p.id = si.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY c.name ORDER BY SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) DESC`,
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

	err := r.db.GetContext(ctx, &report.TotalDebt,
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
	err = r.db.GetContext(ctx, &report.OverdueDebt,
		`SELECT COALESCE(SUM(remaining_amount), 0) FROM debts
		 WHERE due_date < CURRENT_DATE AND remaining_amount > 0`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve overdue debt: %w", err)
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT d.customer_id, c.name, COALESCE(SUM(d.amount), 0),
		        COALESCE(SUM(d.amount - d.remaining_amount), 0),
		        COALESCE(SUM(d.remaining_amount), 0),
		        COALESCE(SUM(CASE WHEN d.due_date < CURRENT_DATE THEN d.remaining_amount ELSE 0 END), 0),
		        COALESCE(MAX(p.payment_date), '0001-01-01')
		 FROM debts d JOIN customers c ON c.id = d.customer_id
		 LEFT JOIN payments p ON p.customer_id = d.customer_id
		 GROUP BY d.customer_id, c.name ORDER BY SUM(d.remaining_amount) DESC`,
	)
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
			WHEN due_date >= CURRENT_DATE THEN 'current'
			WHEN CURRENT_DATE - due_date <= 7 THEN 'overdue_1_7'
			WHEN CURRENT_DATE - due_date <= 30 THEN 'overdue_8_30'
			ELSE 'overdue_30_plus' END, COUNT(*)
		 FROM debts WHERE remaining_amount > 0 GROUP BY 1`
	if dbutil.IsSQLite(r.db) {
		ageQuery = `SELECT CASE
			WHEN date(due_date) >= date('now') THEN 'current'
			WHEN julianday('now') - julianday(due_date) <= 7 THEN 'overdue_1_7'
			WHEN julianday('now') - julianday(due_date) <= 30 THEN 'overdue_8_30'
			ELSE 'overdue_30_plus' END, COUNT(*)
		 FROM debts WHERE remaining_amount > 0 GROUP BY 1`
	}
	rows, err = r.db.QueryContext(ctx, ageQuery)
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
		`SELECT p.payment_date, p.customer_id, COALESCE(c.name, 'Ø¹Ù…ÙŠÙ„ ØºÙŠØ± Ù…Ø¹Ø±ÙˆÙ'), p.amount
		 FROM payments p LEFT JOIN customers c ON c.id = p.customer_id
		 WHERE p.customer_id IS NOT NULL ORDER BY p.payment_date DESC LIMIT 100`)
	if err == nil {
		for rows.Next() {
			var payment PaymentRecord
			var paymentDate reportTimestamp
			if err := rows.Scan(&paymentDate, &payment.CustomerID, &payment.CustomerName, &payment.Amount); err == nil {
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
		`SELECT p.supplier_id, COALESCE(s.name, 'Ù…ÙˆØ±Ø¯ ØºÙŠØ± Ù…Ø¹Ø±ÙˆÙ'),
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
		`SELECT COALESCE(c.name, 'ØºÙŠØ± Ù…ØµÙ†Ù'), COALESCE(SUM(pi.quantity), 0)
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
		`SELECT COUNT(*) FROM returns
		 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
		   AND status = 'COMPLETED'`,
		startDate, endDate)
	if err != nil {
		// If returns table doesn't exist, return empty report
		report.TotalReturns = 0
		report.TotalRefunded = 0
		report.ByReason = make(map[string]int)
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalRefunded,
		`SELECT COALESCE(SUM(total_refund_amount), 0) FROM returns
		 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
		   AND status = 'COMPLETED'`,
		startDate, endDate)
	if err != nil {
		report.TotalRefunded = 0
	}

	report.ByReason = make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		`SELECT reason, COUNT(*) FROM returns
		 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
		   AND status = 'COMPLETED'
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
		`SELECT ri.product_id, COALESCE(p.name, 'Ù…Ù†ØªØ¬ Ù…Ø­Ø°ÙˆÙ'),
			        COUNT(DISTINCT r.id), COALESCE(SUM(ri.total_refund_amount), 0)
			 FROM returns r
			 JOIN return_items ri ON ri.return_id = r.id
			 LEFT JOIN products p ON p.id = ri.product_id
			 WHERE date(r.return_date) >= date(substr($1, 1, 10)) AND date(r.return_date) <= date(substr($2, 1, 10))
			   AND r.status = 'COMPLETED'
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
			 FROM returns
			 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
			   AND status = 'COMPLETED'
			 GROUP BY strftime('%Y-%m', return_date)
			 ORDER BY strftime('%Y-%m', return_date)`
	if dbutil.IsSQLite(r.db) {
		monthlyReturnsQuery = `SELECT strftime('%Y-%m-01', return_date), COUNT(*), COALESCE(SUM(total_refund_amount), 0)
			 FROM returns
			 WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10))
		   AND status = 'COMPLETED'
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
		`SELECT COUNT(*) FROM sales WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) <= date(substr($2, 1, 10))
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
		`SELECT COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) FROM sales WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) <= date(substr($2, 1, 10))
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
		fmt.Sprintf(`SELECT COUNT(*) FROM returns r
		 WHERE %s >= date(substr($1, 1, 10)) AND %s <= date(substr($2, 1, 10))
		   AND status = 'COMPLETED'`,
			returnDate, returnDate),
		startDate, endDate)
	if err != nil {
		totalReturns = 0
	}
	report.TotalReturns = int(totalReturns)

	err = r.db.GetContext(ctx, &totalRefunded,
		fmt.Sprintf(`SELECT COALESCE(SUM(total_refund_amount), 0) FROM returns r
		 WHERE %s >= date(substr($1, 1, 10)) AND %s <= date(substr($2, 1, 10))
		   AND status = 'COMPLETED'`,
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
			COALESCE((SELECT COUNT(*) FROM returns r
				WHERE DATE(r.return_date) = DATE(s.sale_date)
				  AND date(r.return_date) >= date(substr($1, 1, 10)) AND date(r.return_date) <= date(substr($2, 1, 10))
				  AND r.status = 'COMPLETED'), 0) as returns,
			COALESCE((SELECT SUM(total_refund_amount) FROM returns r
				WHERE DATE(r.return_date) = DATE(s.sale_date)
				  AND date(r.return_date) >= date(substr($1, 1, 10)) AND date(r.return_date) <= date(substr($2, 1, 10))
				  AND r.status = 'COMPLETED'), 0) as refunded
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
		`SELECT COALESCE(c.name, 'ØºÙŠØ± Ù…ØµÙ†Ù'), COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0) AS total
		 FROM sale_items si
		 JOIN sales s ON s.id = si.sale_id
		 JOIN products p ON p.id = si.product_id
			 LEFT JOIN categories c ON c.id = p.category_id
				 WHERE date(s.sale_date) >= date(substr($1, 1, 10)) AND date(s.sale_date) <= date(substr($2, 1, 10))
			   AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
			 GROUP BY c.name
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
			COALESCE((SELECT SUM(ri.quantity_returned) FROM return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM returns WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10)) AND status = 'COMPLETED')), 0) as returned_quantity,
			COALESCE((SELECT SUM(ri.total_refund_amount) FROM return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM returns WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10)) AND status = 'COMPLETED')), 0) as refunded_amount
		 FROM products p
			 LEFT JOIN sale_items si ON p.id = si.product_id AND si.sale_id IN (SELECT id FROM sales WHERE date(sale_date) >= date(substr($1, 1, 10)) AND date(sale_date) <= date(substr($2, 1, 10)) AND LOWER(COALESCE(status, 'completed')) = 'completed')
		 WHERE p.is_active = true
		 GROUP BY p.id, p.name
		 HAVING COALESCE(SUM(si.quantity), 0) > 0 OR COALESCE((SELECT SUM(ri.quantity_returned) FROM return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM returns WHERE date(return_date) >= date(substr($1, 1, 10)) AND date(return_date) <= date(substr($2, 1, 10)) AND status = 'COMPLETED')), 0) > 0
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
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN total_amount ELSE 0 END), 0)
		FROM sales
		WHERE date(sale_date) >= date(substr($1, 1, 10))
		  AND date(sale_date) <= date(substr($2, 1, 10))
		  AND LOWER(COALESCE(status, 'completed')) = 'completed'
		  AND tax_amount > 0`
	if err := r.db.QueryRowxContext(ctx, query, startDate, endDate).Scan(&report.GrossSales, &report.Discounts, &report.TaxableSales, &report.TaxCollected, &report.SalesTotal); err != nil {
		return nil, err
	}
	returnDate := r.returnDateExpression("r")
	if err := r.db.GetContext(ctx, &report.ReturnsTotal, fmt.Sprintf(`SELECT COALESCE(SUM(total_refund_amount), 0) FROM returns r WHERE %s >= date(substr($1, 1, 10)) AND %s <= date(substr($2, 1, 10)) AND status = 'COMPLETED'`, returnDate, returnDate), startDate, endDate); err != nil {
		report.ReturnsTotal = 0
	}
	report.NetSalesTotal = report.SalesTotal - report.ReturnsTotal
	return report, nil
}
