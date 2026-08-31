package expenses

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles expense data operations
type Repository struct {
	db *sqlx.DB
}

type localExpenseRow struct {
	ID              string         `db:"id"`
	CategoryID      sql.NullString `db:"category_id"`
	Title           string         `db:"title"`
	Description     sql.NullString `db:"description"`
	Amount          float64        `db:"amount"`
	Currency        string         `db:"currency"`
	ExpenseDate     string         `db:"expense_date"`
	PaymentMethod   sql.NullString `db:"payment_method"`
	Reference       sql.NullString `db:"reference"`
	ReceiptURL      sql.NullString `db:"receipt_url"`
	IsRecurring     int            `db:"is_recurring"`
	RecurringPeriod sql.NullString `db:"recurring_period"`
	ApprovedBy      sql.NullString `db:"approved_by"`
	Status          string         `db:"status"`
	CreatedBy       sql.NullString `db:"created_by"`
	CreatedAt       string         `db:"created_at"`
	UpdatedAt       string         `db:"updated_at"`
}

type localExpenseCategoryRow struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Color       sql.NullString `db:"color"`
	Icon        sql.NullString `db:"icon"`
	Budget      float64        `db:"budget"`
	IsActive    int            `db:"is_active"`
	CreatedAt   string         `db:"created_at"`
	UpdatedAt   string         `db:"updated_at"`
}

func localExpenseCategoryFromRow(row localExpenseCategoryRow) (ExpenseCategory, error) {
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return ExpenseCategory{}, fmt.Errorf("parse expense category id: %w", err)
	}
	createdAt, err := dbutil.ParseTimestamp(row.CreatedAt)
	if err != nil {
		return ExpenseCategory{}, fmt.Errorf("parse category created_at: %w", err)
	}
	updatedAt, err := dbutil.ParseTimestamp(row.UpdatedAt)
	if err != nil {
		return ExpenseCategory{}, fmt.Errorf("parse category updated_at: %w", err)
	}
	category := ExpenseCategory{ID: id, Name: row.Name, Budget: row.Budget, IsActive: row.IsActive != 0, CreatedAt: createdAt, UpdatedAt: updatedAt}
	if row.Description.Valid {
		category.Description = row.Description.String
	}
	if row.Color.Valid {
		category.Color = row.Color.String
	}
	if row.Icon.Valid {
		category.Icon = row.Icon.String
	}
	return category, nil
}

func localExpenseFromRow(row localExpenseRow) (Expense, error) {
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return Expense{}, fmt.Errorf("parse expense id: %w", err)
	}
	categoryID := uuid.Nil
	if row.CategoryID.Valid && row.CategoryID.String != "" {
		categoryID, err = uuid.Parse(row.CategoryID.String)
		if err != nil {
			return Expense{}, fmt.Errorf("parse category id: %w", err)
		}
	}
	expenseDate, err := dbutil.ParseTimestamp(row.ExpenseDate)
	if err != nil {
		return Expense{}, fmt.Errorf("parse expense date: %w", err)
	}
	createdAt, err := dbutil.ParseTimestamp(row.CreatedAt)
	if err != nil {
		return Expense{}, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := dbutil.ParseTimestamp(row.UpdatedAt)
	if err != nil {
		return Expense{}, fmt.Errorf("parse updated_at: %w", err)
	}
	expense := Expense{ID: id, CategoryID: categoryID, Title: row.Title, Amount: row.Amount, Currency: row.Currency, ExpenseDate: expenseDate, IsRecurring: row.IsRecurring != 0, Status: row.Status, CreatedAt: createdAt, UpdatedAt: updatedAt}
	if row.Description.Valid {
		expense.Description = row.Description.String
	}
	if row.PaymentMethod.Valid {
		expense.PaymentMethod = row.PaymentMethod.String
	}
	if row.Reference.Valid {
		expense.Reference = row.Reference.String
	}
	if row.ReceiptURL.Valid {
		expense.ReceiptURL = row.ReceiptURL.String
	}
	if row.RecurringPeriod.Valid {
		expense.RecurringPeriod = row.RecurringPeriod.String
	}
	if row.ApprovedBy.Valid && row.ApprovedBy.String != "" {
		if id, e := uuid.Parse(row.ApprovedBy.String); e == nil {
			expense.ApprovedBy = &id
		}
	}
	if row.CreatedBy.Valid && row.CreatedBy.String != "" {
		if id, e := uuid.Parse(row.CreatedBy.String); e == nil {
			expense.CreatedBy = id
		}
	}
	return expense, nil
}

// NewRepository creates a new expense repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateExpense creates a new expense
func (r *Repository) CreateExpense(ctx context.Context, expense *Expense) error {
	var categoryName string
	if err := r.db.GetContext(ctx, &categoryName, `SELECT name FROM expense_categories WHERE id = $1`, expense.CategoryID); err != nil {
		return fmt.Errorf("failed to resolve expense category: %w", err)
	}
	query := `
		INSERT INTO expenses (id, reference_number, category, category_id, title, description,
		amount, currency, expense_date, payment_method, reference, receipt_url, 
		is_recurring, recurring_period, approved_by, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		expense.ID, expense.Reference, categoryName, expense.CategoryID, expense.Title, expense.Description,
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
func (r *Repository) GetExpenseByID(ctx context.Context, id uuid.UUID) (*Expense, error) {
	if dbutil.IsSQLite(r.db) {
		var row localExpenseRow
		if err := r.db.GetContext(ctx, &row, `SELECT id, category_id, title, description, amount, currency, expense_date, payment_method, reference, receipt_url, is_recurring, recurring_period, approved_by, status, created_by, created_at, updated_at FROM expenses WHERE id = $1`, id); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrExpenseNotFound
			}
			return nil, fmt.Errorf("failed to get expense: %w", err)
		}
		expense, err := localExpenseFromRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expense: %w", err)
		}
		return &expense, nil
	}
	var expense Expense
	query := `
		SELECT id, category_id, title, COALESCE(description, ''),
			amount, currency, expense_date, COALESCE(payment_method, ''), COALESCE(reference, ''), COALESCE(receipt_url, ''), is_recurring,
			COALESCE(recurring_period, ''), COALESCE(approved_by, ''), status, COALESCE(created_by, ''), created_at, updated_at
		FROM expenses
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &expense, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrExpenseNotFound
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}
	return &expense, nil
}

// ListExpenses retrieves expenses with pagination and filters
func (r *Repository) ListExpenses(ctx context.Context, req ExpenseListRequest) ([]Expense, int, error) {
	if dbutil.IsSQLite(r.db) {
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PerPage <= 0 || req.PerPage > 100 {
			req.PerPage = 20
		}
		query := `SELECT id, category_id, title, description, amount, currency, expense_date, payment_method, reference, receipt_url, is_recurring, recurring_period, approved_by, status, created_by, created_at, updated_at FROM expenses WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM expenses WHERE 1=1`
		args := make([]interface{}, 0, 8)
		argCount := 0
		add := func(condition string, value interface{}) {
			argCount++
			query += fmt.Sprintf(" AND %s $%d", condition, argCount)
			countQuery += fmt.Sprintf(" AND %s $%d", condition, argCount)
			args = append(args, value)
		}
		if req.CategoryID != nil {
			add("category_id =", *req.CategoryID)
		}
		if req.Status != "" {
			add("status =", req.Status)
		}
		if req.PaymentMethod != "" {
			add("payment_method =", req.PaymentMethod)
		}
		if req.IsRecurring != nil {
			value := 0
			if *req.IsRecurring {
				value = 1
			}
			add("is_recurring =", value)
		}
		if req.StartDate != nil {
			add("expense_date >=", *req.StartDate)
		}
		if req.EndDate != nil {
			add("expense_date <=", *req.EndDate)
		}
		if req.Search != "" {
			argCount++
			condition := fmt.Sprintf("(LOWER(COALESCE(title, '')) LIKE LOWER($%d) OR LOWER(COALESCE(description, '')) LIKE LOWER($%d) OR LOWER(COALESCE(reference, '')) LIKE LOWER($%d))", argCount, argCount, argCount)
			query += " AND " + condition
			countQuery += " AND " + condition
			args = append(args, "%"+req.Search+"%")
		}
		var count int
		if err := r.db.GetContext(ctx, &count, countQuery, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to count expenses: %w", err)
		}
		sortColumns := map[string]string{"expense_date": "expense_date", "created_at": "created_at", "amount": "amount", "status": "status"}
		sortBy := sortColumns[req.SortBy]
		if sortBy == "" {
			sortBy = "expense_date"
		}
		sortOrder := "DESC"
		if strings.EqualFold(req.SortOrder, "asc") {
			sortOrder = "ASC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortBy, sortOrder, argCount+1, argCount+2)
		args = append(args, req.PerPage, (req.Page-1)*req.PerPage)
		var rows []localExpenseRow
		if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to list expenses: %w", err)
		}
		result := make([]Expense, 0, len(rows))
		for _, row := range rows {
			expense, err := localExpenseFromRow(row)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to parse expense: %w", err)
			}
			result = append(result, expense)
		}
		return result, count, nil
	}
	var expenses []Expense
	var count int

	// Build base query
	baseQuery := `
		SELECT id, category_id, title, COALESCE(description, ''),
			amount, currency, expense_date, COALESCE(payment_method, ''), COALESCE(reference, ''), COALESCE(receipt_url, ''), is_recurring,
			COALESCE(recurring_period, ''), COALESCE(approved_by, ''), status, COALESCE(created_by, ''), created_at, updated_at
		FROM expenses
	`

	countQuery := `
		SELECT COUNT(*) FROM expenses
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if req.CategoryID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" WHERE category_id = $%d", argCount)
		countQuery += fmt.Sprintf(" WHERE category_id = $%d", argCount)
		args = append(args, *req.CategoryID)
	} else {
		baseQuery += " WHERE 1=1"
		countQuery += " WHERE 1=1"
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
		baseQuery += fmt.Sprintf(" AND (LOWER(title) LIKE LOWER($%d) OR LOWER(description) LIKE LOWER($%d) OR LOWER(reference) LIKE LOWER($%d))", argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (LOWER(title) LIKE LOWER($%d) OR LOWER(description) LIKE LOWER($%d) OR LOWER(reference) LIKE LOWER($%d))", argCount, argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		// If table doesn't exist, return empty results
		if err.Error() == `pq: relation "expenses" does not exist` {
			return []Expense{}, 0, nil
		}
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
		// If table doesn't exist, return empty results
		if err.Error() == `pq: relation "expenses" does not exist` {
			return []Expense{}, 0, nil
		}
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
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		expense.ID, expense.CategoryID, expense.Title, expense.Description, expense.Amount,
		expense.Currency, expense.ExpenseDate, expense.PaymentMethod, expense.Reference,
		expense.ReceiptURL, expense.IsRecurring, expense.RecurringPeriod, expense.ApprovedBy,
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
func (r *Repository) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM expenses WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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
		INSERT INTO expense_categories (id, name, description, color, icon, budget, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		category.ID, category.Name, category.Description, category.Color, category.Icon,
		category.Budget, category.IsActive, category.CreatedAt, category.UpdatedAt,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create expense category: %w", err)
	}
	return nil
}

// GetExpenseCategoryByID retrieves an expense category by ID
func (r *Repository) GetExpenseCategoryByID(ctx context.Context, id uuid.UUID) (*ExpenseCategory, error) {
	if dbutil.IsSQLite(r.db) {
		var row localExpenseCategoryRow
		if err := r.db.GetContext(ctx, &row, `SELECT id, name, description, color, icon, budget, is_active, created_at, updated_at FROM expense_categories WHERE id = $1`, id); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrExpenseCategoryNotFound
			}
			return nil, fmt.Errorf("failed to get expense category: %w", err)
		}
		category, err := localExpenseCategoryFromRow(row)
		if err != nil {
			return nil, err
		}
		return &category, nil
	}
	var category ExpenseCategory
	query := `
		SELECT id, name, description, color, icon, budget, is_active, created_at, updated_at
		FROM expense_categories
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &category, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrExpenseCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get expense category: %w", err)
	}
	return &category, nil
}

// ListExpenseCategories retrieves expense categories with pagination and filters
func (r *Repository) ListExpenseCategories(ctx context.Context, req ExpenseCategoryListRequest) ([]ExpenseCategory, int, error) {
	if dbutil.IsSQLite(r.db) {
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PerPage <= 0 || req.PerPage > 100 {
			req.PerPage = 20
		}
		query := `SELECT id, name, description, color, icon, budget, is_active, created_at, updated_at FROM expense_categories WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM expense_categories WHERE 1=1`
		args := make([]interface{}, 0, 3)
		argCount := 0
		if req.IsActive != nil {
			argCount++
			query += fmt.Sprintf(" AND is_active = $%d", argCount)
			countQuery += fmt.Sprintf(" AND is_active = $%d", argCount)
			value := 0
			if *req.IsActive {
				value = 1
			}
			args = append(args, value)
		}
		if req.Search != "" {
			argCount++
			query += fmt.Sprintf(" AND (LOWER(COALESCE(name, '')) LIKE LOWER($%d) OR LOWER(COALESCE(description, '')) LIKE LOWER($%d))", argCount, argCount)
			countQuery += fmt.Sprintf(" AND (LOWER(COALESCE(name, '')) LIKE LOWER($%d) OR LOWER(COALESCE(description, '')) LIKE LOWER($%d))", argCount, argCount)
			args = append(args, "%"+req.Search+"%")
		}
		var count int
		if err := r.db.GetContext(ctx, &count, countQuery, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to count expense categories: %w", err)
		}
		sortBy := "name"
		if req.SortBy == "created_at" {
			sortBy = "created_at"
		}
		sortOrder := "ASC"
		if strings.EqualFold(req.SortOrder, "desc") {
			sortOrder = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortBy, sortOrder, argCount+1, argCount+2)
		args = append(args, req.PerPage, (req.Page-1)*req.PerPage)
		var rows []localExpenseCategoryRow
		if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to list expense categories: %w", err)
		}
		result := make([]ExpenseCategory, 0, len(rows))
		for _, row := range rows {
			category, err := localExpenseCategoryFromRow(row)
			if err != nil {
				return nil, 0, err
			}
			result = append(result, category)
		}
		return result, count, nil
	}
	var categories []ExpenseCategory
	var count int

	// Build base query
	baseQuery := `
		SELECT id, name, description, color, icon, budget, is_active, created_at, updated_at
		FROM expense_categories
	`

	countQuery := `
		SELECT COUNT(*) FROM expense_categories
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	baseQuery += " WHERE 1=1"
	countQuery += " WHERE 1=1"

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
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		category.ID, category.Name, category.Description, category.Color,
		category.Icon, category.Budget, category.IsActive, category.UpdatedAt,
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
func (r *Repository) DeleteExpenseCategory(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM expense_categories WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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
func (r *Repository) GetExpenseCategoryByName(ctx context.Context, name string) (*ExpenseCategory, error) {
	var category ExpenseCategory
	query := `
		SELECT id, name, description, color, icon, budget, is_active, created_at, updated_at
		FROM expense_categories
		WHERE name = $1
	`

	err := r.db.GetContext(ctx, &category, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrExpenseCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get expense category by name: %w", err)
	}
	return &category, nil
}

// GetExpenseSummary retrieves expense summary statistics
func (r *Repository) GetExpenseSummary(ctx context.Context) (*ExpenseSummary, error) {
	var summary ExpenseSummary

	// Total expenses
	err := r.db.GetContext(ctx, &summary.TotalExpenses,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses`)
	if err != nil {
		// If table doesn't exist, return empty summary
		if err.Error() == `pq: relation "expenses" does not exist` {
			return &ExpenseSummary{
				TotalExpenses:    0,
				PendingExpenses:  0,
				ApprovedExpenses: 0,
				ThisMonth:        0,
				LastMonth:        0,
				ByCategory:       make(map[string]float64),
				ByPaymentMethod:  make(map[string]float64),
			}, nil
		}
		return nil, fmt.Errorf("failed to get total expenses: %w", err)
	}

	// Pending expenses
	err = r.db.GetContext(ctx, &summary.PendingExpenses,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE status = 'pending'`)
	if err != nil {
		if err.Error() == `pq: relation "expenses" does not exist` {
			summary.PendingExpenses = 0
		} else {
			return nil, fmt.Errorf("failed to get pending expenses: %w", err)
		}
	}

	// Approved expenses
	err = r.db.GetContext(ctx, &summary.ApprovedExpenses,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE status = 'approved'`)
	if err != nil {
		if err.Error() == `pq: relation "expenses" does not exist` {
			summary.ApprovedExpenses = 0
		} else {
			return nil, fmt.Errorf("failed to get approved expenses: %w", err)
		}
	}

	// This month expenses
	err = r.db.GetContext(ctx, &summary.ThisMonth,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses 
		 WHERE DATE_TRUNC('month', expense_date) = DATE_TRUNC('month', CURRENT_DATE)`)
	if err != nil {
		if err.Error() == `pq: relation "expenses" does not exist` {
			summary.ThisMonth = 0
		} else {
			return nil, fmt.Errorf("failed to get this month expenses: %w", err)
		}
	}

	// Last month expenses
	err = r.db.GetContext(ctx, &summary.LastMonth,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses 
		 WHERE DATE_TRUNC('month', expense_date) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')`)
	if err != nil {
		if err.Error() == `pq: relation "expenses" does not exist` {
			summary.LastMonth = 0
		} else {
			return nil, fmt.Errorf("failed to get last month expenses: %w", err)
		}
	}

	// By category
	summary.ByCategory = make(map[string]float64)
	rows, err := r.db.QueryContext(ctx,
		`SELECT ec.name, COALESCE(SUM(e.amount), 0) 
		 FROM expense_categories ec 
		 LEFT JOIN expenses e ON ec.id = e.category_id
		 GROUP BY ec.name`)
	if err != nil {
		// If table doesn't exist, skip this part
		if err.Error() != `pq: relation "expense_categories" does not exist` {
			return nil, fmt.Errorf("failed to get expenses by category: %w", err)
		}
	} else {
		defer rows.Close()

		for rows.Next() {
			var name string
			var amount float64
			if err := rows.Scan(&name, &amount); err != nil {
				continue
			}
			summary.ByCategory[name] = amount
		}
	}

	// By payment method
	summary.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT payment_method, COALESCE(SUM(amount), 0) 
		 FROM expenses 
		 GROUP BY payment_method`)
	if err != nil {
		// If table doesn't exist, skip this part
		if err.Error() != `pq: relation "expenses" does not exist` {
			return nil, fmt.Errorf("failed to get expenses by payment method: %w", err)
		}
	} else {
		defer rows.Close()

		for rows.Next() {
			var method string
			var amount float64
			if err := rows.Scan(&method, &amount); err != nil {
				continue
			}
			summary.ByPaymentMethod[method] = amount
		}
	}

	return &summary, nil
}
