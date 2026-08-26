package reports

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles report data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new report repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateReport creates a new report
func (r *Repository) CreateReport(ctx context.Context, report *Report) error {
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
		baseQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}
	
	// Add sorting
	sortBy := "generated_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "DESC"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
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
	
	// Total sales and revenue - simplified query
	err := r.db.GetContext(ctx, &report.TotalSales,
		`SELECT COUNT(*) FROM sales WHERE sale_date >= $1 AND sale_date <= $2`,
		startDate, endDate)
	if err != nil {
		// If sales table doesn't exist or query fails, return empty report
		report.TotalSales = 0
		report.TotalRevenue = 0
		report.TotalCOGS = 0
		report.GrossProfit = 0
		report.ProfitMargin = 0
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalRevenue,
		`SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE sale_date >= $1 AND sale_date <= $2`,
		startDate, endDate)
	if err != nil {
		report.TotalRevenue = 0
	}

	// Simplified COGS calculation
	report.TotalCOGS = report.TotalRevenue * 0.7 // Assume 70% of revenue is COGS
	
	report.GrossProfit = report.TotalRevenue - report.TotalCOGS
	if report.TotalRevenue > 0 {
		report.ProfitMargin = (report.GrossProfit / report.TotalRevenue) * 100
	} else {
		report.ProfitMargin = 0
	}
	
	// Get daily sales - simplified
	rows, err := r.db.QueryContext(ctx, 
		`SELECT DATE(sale_date) as date, COUNT(*) as sales, COALESCE(SUM(total_amount), 0) as revenue
		 FROM sales 
		 WHERE sale_date >= $1 AND sale_date <= $2
		 GROUP BY DATE(sale_date)
		 ORDER BY date`,
		startDate, endDate)
	if err == nil {
		defer rows.Close()
		
		for rows.Next() {
			var daily DailySales
			if err := rows.Scan(&daily.Date, &daily.Sales, &daily.Revenue); err != nil {
				continue
			}
			report.ByDay = append(report.ByDay, daily)
		}
	}
	
	// Simplified top products - just get product names if we can
	report.TopProducts = []ProductSales{}
	
	// Get payment method breakdown - simplified
	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx, 
		`SELECT payment_method, COALESCE(SUM(total_amount), 0) as total
		 FROM sales
		 WHERE sale_date >= $1 AND sale_date <= $2
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
	}
	
	return &report, nil
}

// GetInventoryData retrieves inventory data for report
func (r *Repository) GetInventoryData(ctx context.Context) (*InventoryReport, error) {
	var report InventoryReport
	
	// Get total items and value - simplified
	err := r.db.GetContext(ctx, &report.TotalItems,
		`SELECT COUNT(*) FROM inventory_items WHERE status = 'AVAILABLE'`)
	if err != nil {
		// If table doesn't exist, return empty report
		report.TotalItems = 0
		report.TotalValue = 0
		report.ByCondition = make(map[string]int)
		report.LowStockItems = []LowStockItem{}
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalValue,
		`SELECT COALESCE(SUM(purchase_cost), 0) FROM inventory_items WHERE status = 'AVAILABLE'`)
	if err != nil {
		report.TotalValue = 0
	}

	// Get items by condition
	report.ByCondition = make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		`SELECT condition, COUNT(*) FROM inventory_items GROUP BY condition`)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var condition string
			var count int
			if err := rows.Scan(&condition, &count); err != nil {
				continue
			}
			report.ByCondition[condition] = count
		}
	}

	// Simplified low stock items - empty for now
	report.LowStockItems = []LowStockItem{}
	
	return &report, nil
}

// GetExpensesData retrieves expenses data for report
func (r *Repository) GetExpensesData(ctx context.Context, startDate, endDate time.Time) (*ExpensesReport, error) {
	var report ExpensesReport
	report.StartDate = startDate
	report.EndDate = endDate
	
	// Total expenses - simplified
	err := r.db.GetContext(ctx, &report.TotalExpenses,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE expense_date >= $1 AND expense_date <= $2`,
		startDate, endDate)
	if err != nil {
		// If expenses table doesn't exist, return empty report
		report.TotalExpenses = 0
		report.ByCategory = make(map[string]float64)
		report.ByPaymentMethod = make(map[string]float64)
		return &report, nil
	}

	// Get expenses by category - simplified
	report.ByCategory = make(map[string]float64)
	rows, err := r.db.QueryContext(ctx,
		`SELECT category, COALESCE(SUM(amount), 0) as total
		 FROM expenses
		 WHERE expense_date >= $1 AND expense_date <= $2
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
	}

	// Get expenses by payment method - simplified
	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT payment_method, COALESCE(SUM(amount), 0) as total
		 FROM expenses
		 WHERE expense_date >= $1 AND expense_date <= $2
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
	}
	
	return &report, nil
}

// GetProfitsData retrieves profits data for report
func (r *Repository) GetProfitsData(ctx context.Context, startDate, endDate time.Time) (*ProfitsReport, error) {
	var report ProfitsReport
	report.StartDate = startDate
	report.EndDate = endDate
	
	// Get revenue - simplified
	err := r.db.GetContext(ctx, &report.TotalRevenue,
		`SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE sale_date >= $1 AND sale_date <= $2`,
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

	// Simplified COGS calculation
	report.TotalCOGS = report.TotalRevenue * 0.7 // Assume 70% of revenue is COGS

	report.GrossProfit = report.TotalRevenue - report.TotalCOGS

	// Get expenses - simplified
	err = r.db.GetContext(ctx, &report.TotalExpenses,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE expense_date >= $1 AND expense_date <= $2`,
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
	
	return &report, nil
}

// GetDebtsData retrieves debts data for report
func (r *Repository) GetDebtsData(ctx context.Context) (*DebtsReport, error) {
	var report DebtsReport

	// Get total debt - simplified
	err := r.db.GetContext(ctx, &report.TotalDebt,
		`SELECT COALESCE(SUM(current_balance), 0) FROM customers WHERE current_balance > 0`)
	if err != nil {
		// If customers table doesn't exist, return empty report
		report.TotalDebt = 0
		report.TotalPaid = 0
		report.Outstanding = 0
		report.OverdueDebt = 0
		report.ByCustomer = []CustomerDebt{}
		return &report, nil
	}

	// Get total paid - simplified
	err = r.db.GetContext(ctx, &report.TotalPaid,
		`SELECT COALESCE(SUM(current_balance), 0) FROM customers WHERE current_balance < 0`)
	if err != nil {
		report.TotalPaid = 0
	}

	report.Outstanding = report.TotalDebt

	// Get overdue debt - simplified (assume 0 for now)
	report.OverdueDebt = 0

	// Get debts by customer - simplified
	report.ByCustomer = []CustomerDebt{}

	return &report, nil
}

// GetPurchasesData retrieves purchases data for report
func (r *Repository) GetPurchasesData(ctx context.Context, startDate, endDate time.Time) (*PurchasesReport, error) {
	var report PurchasesReport
	report.StartDate = startDate
	report.EndDate = endDate

	// Get total purchases and cost - simplified
	err := r.db.GetContext(ctx, &report.TotalPurchases,
		`SELECT COUNT(*) FROM purchases WHERE purchase_date >= $1 AND purchase_date <= $2`,
		startDate, endDate)
	if err != nil {
		// If purchases table doesn't exist, return empty report
		report.TotalPurchases = 0
		report.TotalCost = 0
		report.BySupplier = []SupplierPurchases{}
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalCost,
		`SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE purchase_date >= $1 AND purchase_date <= $2`,
		startDate, endDate)
	if err != nil {
		report.TotalCost = 0
	}

	// Simplified purchases by supplier - empty for now
	report.BySupplier = []SupplierPurchases{}

	return &report, nil
}

// GetReturnsData retrieves returns data for report
func (r *Repository) GetReturnsData(ctx context.Context, startDate, endDate time.Time) (*ReturnsReport, error) {
	var report ReturnsReport
	report.StartDate = startDate
	report.EndDate = endDate

	// Get total returns and refunds - simplified
	err := r.db.GetContext(ctx, &report.TotalReturns,
		`SELECT COUNT(*) FROM returns WHERE return_date >= $1 AND return_date <= $2`,
		startDate, endDate)
	if err != nil {
		// If returns table doesn't exist, return empty report
		report.TotalReturns = 0
		report.TotalRefunded = 0
		report.ByReason = make(map[string]int)
		return &report, nil
	}

	err = r.db.GetContext(ctx, &report.TotalRefunded,
		`SELECT COALESCE(SUM(refund_amount), 0) FROM returns WHERE return_date >= $1 AND return_date <= $2`,
		startDate, endDate)
	if err != nil {
		report.TotalRefunded = 0
	}

	// Get returns by reason - simplified
	report.ByReason = make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		`SELECT reason, COUNT(*) FROM returns WHERE return_date >= $1 AND return_date <= $2 GROUP BY reason`,
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
		`SELECT COUNT(*) FROM sales WHERE sale_date >= $1 AND sale_date <= $2`,
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
		`SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE sale_date >= $1 AND sale_date <= $2`,
		startDate, endDate)
	if err != nil {
		grossRevenue = 0
	}
	report.GrossRevenue = grossRevenue

	// Get returns data
	var totalReturns, totalRefunded float64
	err = r.db.GetContext(ctx, &totalReturns,
		`SELECT COUNT(*) FROM returns WHERE return_date >= $1 AND return_date <= $2`,
		startDate, endDate)
	if err != nil {
		totalReturns = 0
	}
	report.TotalReturns = int(totalReturns)

	err = r.db.GetContext(ctx, &totalRefunded,
		`SELECT COALESCE(SUM(refund_amount), 0) FROM returns WHERE return_date >= $1 AND return_date <= $2`,
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
			COALESCE(SUM(total_amount), 0) as gross_revenue,
			COALESCE((SELECT COUNT(*) FROM returns r WHERE DATE(r.return_date) = DATE(s.sale_date) AND r.return_date >= $1 AND r.return_date <= $2), 0) as returns,
			COALESCE((SELECT SUM(refund_amount) FROM returns r WHERE DATE(r.return_date) = DATE(s.sale_date) AND r.return_date >= $1 AND r.return_date <= $2), 0) as refunded
		 FROM sales s
		 WHERE s.sale_date >= $1 AND s.sale_date <= $2
		 GROUP BY DATE(s.sale_date)
		 ORDER BY date`,
		startDate, endDate)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var daily DailyNetSales
			if err := rows.Scan(&daily.Date, &daily.GrossSales, &daily.GrossRevenue, &daily.Returns, &daily.Refunded); err != nil {
				continue
			}
			daily.NetSales = daily.GrossSales - daily.Returns
			daily.NetRevenue = daily.GrossRevenue - daily.Refunded
			report.ByDay = append(report.ByDay, daily)
		}
	}

	// Get payment method breakdown for gross sales
	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT payment_method, COALESCE(SUM(total_amount), 0) as total
		 FROM sales
		 WHERE sale_date >= $1 AND sale_date <= $2
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
	}

	// Get top returned products with net sales analysis
	report.TopReturnedProducts = []ProductNetSales{}
	rows, err = r.db.QueryContext(ctx,
		`SELECT 
			p.id as product_id,
			p.name as product_name,
			COALESCE(SUM(si.quantity), 0) as gross_quantity,
			COALESCE(SUM(si.quantity * si.unit_price), 0) as gross_revenue,
			COALESCE((SELECT SUM(ri.quantity) FROM return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM returns WHERE return_date >= $1 AND return_date <= $2)), 0) as returned_quantity,
			COALESCE((SELECT SUM(ri.quantity * ri.refund_amount) FROM return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM returns WHERE return_date >= $1 AND return_date <= $2)), 0) as refunded_amount
		 FROM products p
		 LEFT JOIN sale_items si ON p.id = si.product_id AND si.sale_id IN (SELECT id FROM sales WHERE sale_date >= $1 AND sale_date <= $2)
		 WHERE p.is_active = true
		 GROUP BY p.id, p.name
		 HAVING COALESCE(SUM(si.quantity), 0) > 0 OR COALESCE((SELECT SUM(ri.quantity) FROM return_items ri WHERE ri.product_id = p.id AND ri.return_id IN (SELECT id FROM returns WHERE return_date >= $1 AND return_date <= $2)), 0) > 0
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
	}

	return &report, nil
}
