package expenses

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles expense data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new expense repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateExpense creates a new expense
func (r *Repository) CreateExpense(ctx context.Context, expense *Expense) error {
	query := `
		INSERT INTO expenses (id, organization_id, category_id, title, description, 
			amount, currency, expense_date, payment_method, reference, receipt_url, 
			is_recurring, recurring_period, approved_by, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		expense.ID, expense.OrganizationID, expense.CategoryID, expense.Title, expense.Description,
		expense.Amount, expense.Currency, expense.ExpenseDate, expense.PaymentMethod, expense.Reference,
		expense.ReceiptURL, expense.IsRecurring, expense.RecurringPeriod, expense.ApprovedBy,
		expense.Status, expense.CreatedBy, expense.CreatedAt, expense.UpdatedAt,
	).Scan(&expense.ID, &expense.CreatedAt, &expense.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create expense: %w", err)
	}
	return nil
}

// GetExpenseByID retrieves an expense by ID
func (r *Repository) GetExpenseByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Expense, error) {
	var expense Expense
	query := `
		SELECT id, organization_id, category_id, title, description, amount, currency, 
			expense_date, payment_method, reference, receipt_url, is_recurring, 
			recurring_period, approved_by, status, created_by, created_at, updated_at
		FROM expenses
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &expense, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrExpenseNotFound
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}
	return &expense, nil
}

// ListExpenses retrieves expenses with pagination and filters
func (r *Repository) ListExpenses(ctx context.Context, organizationID uuid.UUID, req ExpenseListRequest) ([]Expense, int, error) {
	var expenses []Expense
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, category_id, title, description, amount, currency, 
			expense_date, payment_method, reference, receipt_url, is_recurring, 
			recurring_period, approved_by, status, created_by, created_at, updated_at
		FROM expenses
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM expenses WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if req.CategoryID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND category_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND category_id = $%d", argCount)
		args = append(args, *req.CategoryID)
	}
	
	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}
	
	if req.PaymentMethod != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND payment_method = $%d", argCount)
		countQuery += fmt.Sprintf(" AND payment_method = $%d", argCount)
		args = append(args, req.PaymentMethod)
	}
	
	if req.IsRecurring != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND is_recurring = $%d", argCount)
		countQuery += fmt.Sprintf(" AND is_recurring = $%d", argCount)
		args = append(args, *req.IsRecurring)
	}
	
	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND expense_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND expense_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	
	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND expense_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND expense_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR reference ILIKE $%d)", argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR reference ILIKE $%d)", argCount, argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count expenses: %w", err)
	}
	
	// Add sorting
	sortBy := "expense_date"
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
	
	err = r.db.SelectContext(ctx, &expenses, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list expenses: %w", err)
	}
	
	return expenses, count, nil
}

// UpdateExpense updates an expense
func (r *Repository) UpdateExpense(ctx context.Context, expense *Expense) error {
	query := `
		UPDATE expenses
		SET category_id = $2, title = $3, description = $4, amount = $5, currency = $6,
			expense_date = $7, payment_method = $8, reference = $9, receipt_url = $10,
			is_recurring = $11, recurring_period = $12, approved_by = $13, status = $14, updated_at = $15
		WHERE id = $1 AND organization_id = $16
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		expense.ID, expense.CategoryID, expense.Title, expense.Description, expense.Amount,
		expense.Currency, expense.ExpenseDate, expense.PaymentMethod, expense.Reference,
		expense.ReceiptURL, expense.IsRecurring, expense.RecurringPeriod, expense.ApprovedBy,
		expense.Status, expense.UpdatedAt, expense.OrganizationID,
	).Scan(&expense.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrExpenseNotFound
		}
		return fmt.Errorf("failed to update expense: %w", err)
	}
	return nil
}

// DeleteExpense deletes an expense
func (r *Repository) DeleteExpense(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM expenses WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrExpenseNotFound
	}
	
	return nil
}

// CreateExpenseCategory creates a new expense category
func (r *Repository) CreateExpenseCategory(ctx context.Context, category *ExpenseCategory) error {
	query := `
		INSERT INTO expense_categories (id, organization_id, name, description, color, icon, budget, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		category.ID, category.OrganizationID, category.Name, category.Description,
		category.Color, category.Icon, category.Budget, category.IsActive,
		category.CreatedAt, category.UpdatedAt,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create expense category: %w", err)
	}
	return nil
}

// GetExpenseCategoryByID retrieves an expense category by ID
func (r *Repository) GetExpenseCategoryByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*ExpenseCategory, error) {
	var category ExpenseCategory
	query := `
		SELECT id, organization_id, name, description, color, icon, budget, is_active, created_at, updated_at
		FROM expense_categories
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &category, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrExpenseCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get expense category: %w", err)
	}
	return &category, nil
}

// ListExpenseCategories retrieves expense categories with pagination and filters
func (r *Repository) ListExpenseCategories(ctx context.Context, organizationID uuid.UUID, req ExpenseCategoryListRequest) ([]ExpenseCategory, int, error) {
	var categories []ExpenseCategory
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, name, description, color, icon, budget, is_active, created_at, updated_at
		FROM expense_categories
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM expense_categories WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if req.IsActive != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND is_active = $%d", argCount)
		countQuery += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *req.IsActive)
	}
	
	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count expense categories: %w", err)
	}
	
	// Add sorting
	sortBy := "name"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "ASC"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	
	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)
	
	err = r.db.SelectContext(ctx, &categories, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list expense categories: %w", err)
	}
	
	return categories, count, nil
}

// UpdateExpenseCategory updates an expense category
func (r *Repository) UpdateExpenseCategory(ctx context.Context, category *ExpenseCategory) error {
	query := `
		UPDATE expense_categories
		SET name = $2, description = $3, color = $4, icon = $5, budget = $6, is_active = $7, updated_at = $8
		WHERE id = $1 AND organization_id = $9
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		category.ID, category.Name, category.Description, category.Color,
		category.Icon, category.Budget, category.IsActive, category.UpdatedAt,
		category.OrganizationID,
	).Scan(&category.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrExpenseCategoryNotFound
		}
		return fmt.Errorf("failed to update expense category: %w", err)
	}
	return nil
}

// DeleteExpenseCategory deletes an expense category
func (r *Repository) DeleteExpenseCategory(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM expense_categories WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete expense category: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrExpenseCategoryNotFound
	}
	
	return nil
}

// GetExpenseCategoryByName retrieves an expense category by name
func (r *Repository) GetExpenseCategoryByName(ctx context.Context, name string, organizationID uuid.UUID) (*ExpenseCategory, error) {
	var category ExpenseCategory
	query := `
		SELECT id, organization_id, name, description, color, icon, budget, is_active, created_at, updated_at
		FROM expense_categories
		WHERE name = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &category, query, name, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrExpenseCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get expense category by name: %w", err)
	}
	return &category, nil
}

// GetExpenseSummary retrieves expense summary statistics
func (r *Repository) GetExpenseSummary(ctx context.Context, organizationID uuid.UUID) (*ExpenseSummary, error) {
	var summary ExpenseSummary
	
	// Total expenses
	err := r.db.GetContext(ctx, &summary.TotalExpenses, 
		`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE organization_id = $1 AND status = 'approved'`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total expenses: %w", err)
	}
	
	// Pending expenses
	err = r.db.GetContext(ctx, &summary.PendingExpenses, 
		`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE organization_id = $1 AND status = 'pending'`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending expenses: %w", err)
	}
	
	// Approved expenses
	err = r.db.GetContext(ctx, &summary.ApprovedExpenses, 
		`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE organization_id = $1 AND status = 'approved'`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved expenses: %w", err)
	}
	
	// This month expenses
	err = r.db.GetContext(ctx, &summary.ThisMonth, 
		`SELECT COALESCE(SUM(amount), 0) FROM expenses 
		 WHERE organization_id = $1 AND status = 'approved' 
		 AND DATE_TRUNC('month', expense_date) = DATE_TRUNC('month', CURRENT_DATE)`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month expenses: %w", err)
	}
	
	// Last month expenses
	err = r.db.GetContext(ctx, &summary.LastMonth, 
		`SELECT COALESCE(SUM(amount), 0) FROM expenses 
		 WHERE organization_id = $1 AND status = 'approved' 
		 AND DATE_TRUNC('month', expense_date) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last month expenses: %w", err)
	}
	
	// By category
	summary.ByCategory = make(map[string]float64)
	rows, err := r.db.QueryContext(ctx, 
		`SELECT ec.name, COALESCE(SUM(e.amount), 0) 
		 FROM expense_categories ec 
		 LEFT JOIN expenses e ON ec.id = e.category_id AND e.organization_id = $1 AND e.status = 'approved'
		 WHERE ec.organization_id = $1 
		 GROUP BY ec.name`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses by category: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var name string
		var amount float64
		if err := rows.Scan(&name, &amount); err != nil {
			continue
		}
		summary.ByCategory[name] = amount
	}
	
	// By payment method
	summary.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx, 
		`SELECT payment_method, COALESCE(SUM(amount), 0) 
		 FROM expenses 
		 WHERE organization_id = $1 AND status = 'approved'
		 GROUP BY payment_method`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses by payment method: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var method string
		var amount float64
		if err := rows.Scan(&method, &amount); err != nil {
			continue
		}
		summary.ByPaymentMethod[method] = amount
	}
	
	return &summary, nil
}
