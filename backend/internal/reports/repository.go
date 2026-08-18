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
		INSERT INTO reports (id, organization_id, type, title, description, parameters, 
			data, status, generated_by, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		report.ID, report.OrganizationID, report.Type, report.Title, report.Description,
		report.Parameters, report.Data, report.Status, report.GeneratedBy, report.GeneratedAt,
		report.CreatedAt, report.UpdatedAt,
	).Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}
	return nil
}

// GetReportByID retrieves a report by ID
func (r *Repository) GetReportByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Report, error) {
	var report Report
	query := `
		SELECT id, organization_id, type, title, description, parameters, 
			data, status, generated_by, generated_at, created_at, updated_at
		FROM reports
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &report, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReportNotFound
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	return &report, nil
}

// ListReports retrieves reports with pagination and filters
func (r *Repository) ListReports(ctx context.Context, organizationID uuid.UUID, req ReportListRequest) ([]Report, int, error) {
	var reports []Report
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, type, title, description, parameters, 
			data, status, generated_by, generated_at, created_at, updated_at
		FROM reports
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM reports WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
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
		WHERE id = $1 AND organization_id = $5
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		report.ID, report.Data, report.Status, report.UpdatedAt, report.OrganizationID,
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
func (r *Repository) DeleteReport(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM reports WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
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
func (r *Repository) GetSalesData(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) (*SalesReport, error) {
	var report SalesReport
	report.StartDate = startDate
	report.EndDate = endDate
	
	// Total sales and revenue
	err := r.db.GetContext(ctx, &report.TotalSales, 
		`SELECT COUNT(*) FROM sales WHERE organization_id = $1 AND sale_date BETWEEN $2 AND $3`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sales: %w", err)
	}
	
	err = r.db.GetContext(ctx, &report.TotalRevenue, 
		`SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE organization_id = $1 AND sale_date BETWEEN $2 AND $3`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}
	
	// Calculate COGS (this would be more complex in real implementation)
	err = r.db.GetContext(ctx, &report.TotalCOGS, 
		`SELECT COALESCE(SUM(si.quantity * si.unit_cost), 0) 
		 FROM sale_items si 
		 JOIN sales s ON si.sale_id = s.id 
		 WHERE s.organization_id = $1 AND s.sale_date BETWEEN $2 AND $3`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total COGS: %w", err)
	}
	
	report.GrossProfit = report.TotalRevenue - report.TotalCOGS
	report.ProfitMargin = report.CalculateProfitMargin()
	
	// Get daily sales
	rows, err := r.db.QueryContext(ctx, 
		`SELECT DATE(sale_date) as date, COUNT(*) as sales, COALESCE(SUM(total_amount), 0) as revenue
		 FROM sales 
		 WHERE organization_id = $1 AND sale_date BETWEEN $2 AND $3
		 GROUP BY DATE(sale_date)
		 ORDER BY date`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily sales: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var daily DailySales
		if err := rows.Scan(&daily.Date, &daily.Sales, &daily.Revenue); err != nil {
			continue
		}
		report.ByDay = append(report.ByDay, daily)
	}
	
	// Get top products
	rows, err = r.db.QueryContext(ctx, 
		`SELECT p.id, p.name, SUM(si.quantity) as quantity, SUM(si.quantity * si.unit_price) as revenue
		 FROM sale_items si
		 JOIN sales s ON si.sale_id = s.id
		 JOIN products p ON si.product_id = p.id
		 WHERE s.organization_id = $1 AND s.sale_date BETWEEN $2 AND $3
		 GROUP BY p.id, p.name
		 ORDER BY revenue DESC
		 LIMIT 10`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var product ProductSales
		if err := rows.Scan(&product.ProductID, &product.ProductName, &product.Quantity, &product.Revenue); err != nil {
			continue
		}
		product.Profit = product.Revenue * 0.3 // Simplified profit calculation
		report.TopProducts = append(report.TopProducts, product)
	}
	
	// Get payment method breakdown
	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx, 
		`SELECT payment_method, COALESCE(SUM(amount), 0) as total
		 FROM payments
		 JOIN sales s ON payments.sale_id = s.id
		 WHERE s.organization_id = $1 AND s.sale_date BETWEEN $2 AND $3
		 GROUP BY payment_method`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment methods: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var method string
		var total float64
		if err := rows.Scan(&method, &total); err != nil {
			continue
		}
		report.ByPaymentMethod[method] = total
	}
	
	return &report, nil
}

// GetInventoryData retrieves inventory data for report
func (r *Repository) GetInventoryData(ctx context.Context, organizationID uuid.UUID) (*InventoryReport, error) {
	var report InventoryReport
	
	// Get total items and value
	err := r.db.GetContext(ctx, &report.TotalItems, 
		`SELECT COUNT(*) FROM inventory_items WHERE organization_id = $1`,
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total items: %w", err)
	}
	
	err = r.db.GetContext(ctx, &report.TotalValue, 
		`SELECT COALESCE(SUM(purchase_cost), 0) FROM inventory_items WHERE organization_id = $1`,
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total value: %w", err)
	}
	
	// Get items by condition
	report.ByCondition = make(map[string]int)
	rows, err := r.db.QueryContext(ctx, 
		`SELECT condition, COUNT(*) FROM inventory_items WHERE organization_id = $1 GROUP BY condition`,
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by condition: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var condition string
		var count int
		if err := rows.Scan(&condition, &count); err != nil {
			continue
		}
		report.ByCondition[condition] = count
	}
	
	// Get low stock items
	rows, err = r.db.QueryContext(ctx, 
		`SELECT p.id, p.name, COUNT(ii.id) as current_stock, p.min_stock_level
		 FROM inventory_items ii
		 JOIN products p ON ii.product_id = p.id
		 WHERE ii.organization_id = $1 AND ii.status = 'available'
		 GROUP BY p.id, p.name, p.min_stock_level
		 HAVING COUNT(ii.id) < p.min_stock_level`,
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock items: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var item LowStockItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &item.MinStock); err != nil {
			continue
		}
		item.ReorderLevel = item.MinStock * 2 // Simplified calculation
		report.LowStockItems = append(report.LowStockItems, item)
	}
	
	return &report, nil
}

// GetExpensesData retrieves expenses data for report
func (r *Repository) GetExpensesData(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) (*ExpensesReport, error) {
	var report ExpensesReport
	report.StartDate = startDate
	report.EndDate = endDate
	
	// Total expenses
	err := r.db.GetContext(ctx, &report.TotalExpenses, 
		`SELECT COALESCE(SUM(amount), 0) FROM expenses 
		 WHERE organization_id = $1 AND expense_date BETWEEN $2 AND $3 AND status = 'approved'`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total expenses: %w", err)
	}
	
	// Get expenses by category
	report.ByCategory = make(map[string]float64)
	rows, err := r.db.QueryContext(ctx, 
		`SELECT ec.name, COALESCE(SUM(e.amount), 0) as total
		 FROM expenses e
		 JOIN expense_categories ec ON e.category_id = ec.id
		 WHERE e.organization_id = $1 AND e.expense_date BETWEEN $2 AND $3 AND e.status = 'approved'
		 GROUP BY ec.name`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses by category: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			continue
		}
		report.ByCategory[category] = total
	}
	
	// Get expenses by payment method
	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx, 
		`SELECT payment_method, COALESCE(SUM(amount), 0) as total
		 FROM expenses 
		 WHERE organization_id = $1 AND expense_date BETWEEN $2 AND $3 AND status = 'approved'
		 GROUP BY payment_method`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses by payment method: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var method string
		var total float64
		if err := rows.Scan(&method, &total); err != nil {
			continue
		}
		report.ByPaymentMethod[method] = total
	}
	
	return &report, nil
}

// GetProfitsData retrieves profits data for report
func (r *Repository) GetProfitsData(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) (*ProfitsReport, error) {
	var report ProfitsReport
	report.StartDate = startDate
	report.EndDate = endDate
	
	// Get revenue
	err := r.db.GetContext(ctx, &report.TotalRevenue, 
		`SELECT COALESCE(SUM(total_amount), 0) FROM sales 
		 WHERE organization_id = $1 AND sale_date BETWEEN $2 AND $3`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}
	
	// Get COGS
	err = r.db.GetContext(ctx, &report.TotalCOGS, 
		`SELECT COALESCE(SUM(si.quantity * si.unit_cost), 0) 
		 FROM sale_items si 
		 JOIN sales s ON si.sale_id = s.id 
		 WHERE s.organization_id = $1 AND s.sale_date BETWEEN $2 AND $3`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total COGS: %w", err)
	}
	
	report.GrossProfit = report.TotalRevenue - report.TotalCOGS
	
	// Get expenses
	err = r.db.GetContext(ctx, &report.TotalExpenses, 
		`SELECT COALESCE(SUM(amount), 0) FROM expenses 
		 WHERE organization_id = $1 AND expense_date BETWEEN $2 AND $3 AND status = 'approved'`,
		organizationID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total expenses: %w", err)
	}
	
	report.NetProfit = report.GrossProfit - report.TotalExpenses
	report.ProfitMargin = report.CalculateProfitMargin()
	
	return &report, nil
}

// GetDebtsData retrieves debts data for report
func (r *Repository) GetDebtsData(ctx context.Context, organizationID uuid.UUID) (*DebtsReport, error) {
	var report DebtsReport
	
	// Get total debt
	err := r.db.GetContext(ctx, &report.TotalDebt, 
		`SELECT COALESCE(SUM(amount), 0) FROM customer_ledger 
		 WHERE organization_id = $1 AND transaction_type = 'sale'`,
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total debt: %w", err)
	}
	
	// Get total paid
	err = r.db.GetContext(ctx, &report.TotalPaid, 
		`SELECT COALESCE(SUM(ABS(amount)), 0) FROM customer_ledger 
		 WHERE organization_id = $1 AND transaction_type = 'payment'`,
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total paid: %w", err)
	}
	
	report.Outstanding = report.TotalDebt - report.TotalPaid
	
	// Get overdue debt
	err = r.db.GetContext(ctx, &report.OverdueDebt, 
		`SELECT COALESCE(SUM(balance), 0) FROM customers 
		 WHERE organization_id = $1 AND balance > 0 AND due_date < CURRENT_DATE`,
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue debt: %w", err)
	}
	
	// Get debts by customer
	rows, err := r.db.QueryContext(ctx, 
		`SELECT id, name, total_purchases, total_paid, balance, due_date 
		 FROM customers 
		 WHERE organization_id = $1 AND balance > 0
		 ORDER BY balance DESC`,
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debts by customer: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var customer CustomerDebt
		var dueDate time.Time
		if err := rows.Scan(&customer.CustomerID, &customer.CustomerName, &customer.TotalDebt, 
			&customer.PaidAmount, &customer.Outstanding, &dueDate); err != nil {
			continue
		}
		
		if dueDate.Before(time.Now()) {
			customer.OverdueAmount = customer.Outstanding
		}
		
		report.ByCustomer = append(report.ByCustomer, customer)
	}
	
	return &report, nil
}
