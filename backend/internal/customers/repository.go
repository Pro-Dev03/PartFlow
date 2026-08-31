package customers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Repository handles customer data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new customer repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new customer
func (r *Repository) Create(ctx context.Context, customer *Customer) error {
	query := `
		INSERT INTO customers (id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.Code, customer.Name, customer.Email, customer.Phone, customer.Address, customer.City,
		customer.Country, customer.TaxID, customer.CreditLimit, customer.CurrentBalance,
		customer.Notes, customer.IsActive, customer.CreatedAt, customer.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	return nil
}

func parseDatabaseTimestamp(value any) (time.Time, error) {
	switch v := value.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return v, nil
	case string:
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05-07:00",
			"2006-01-02 15:04:05",
			"2006-01-02",
		} {
			if parsed, err := time.Parse(layout, v); err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, fmt.Errorf("unsupported timestamp format: %q", v)
	case []byte:
		return parseDatabaseTimestamp(string(v))
	default:
		return time.Time{}, fmt.Errorf("unsupported timestamp type %T", value)
	}
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	out := value.String
	return &out
}

// GetByID retrieves a customer by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE id = $1
	`

	var (
		idValue            string
		code, name         string
		email, phone       sql.NullString
		address, city      sql.NullString
		country, taxID     sql.NullString
		notes              sql.NullString
		creditLimit        float64
		currentBalance     float64
		isActive           bool
		createdAtRaw       any
		updatedAtRaw       any
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&idValue, &code, &name,
		&email, &phone,
		&address, &city, &country,
		&taxID, &creditLimit, &currentBalance,
		&notes, &isActive,
		&createdAtRaw, &updatedAtRaw,
	)
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	createdAt, err := parseDatabaseTimestamp(createdAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	updatedAt, err := parseDatabaseTimestamp(updatedAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	parsedID, err := uuid.Parse(idValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse customer id: %w", err)
	}

	customer := &Customer{
		ID:             parsedID,
		Code:           code,
		Name:           name,
		Email:          nullableString(email),
		Phone:          nullableString(phone),
		Address:        nullableString(address),
		City:           nullableString(city),
		Country:        nullableString(country),
		TaxID:          nullableString(taxID),
		CreditLimit:    creditLimit,
		CurrentBalance: currentBalance,
		Notes:          nullableString(notes),
		IsActive:       isActive,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
	return customer, nil
}

// GetByCode retrieves a customer by code
func (r *Repository) GetByCode(ctx context.Context, code string) (*Customer, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE code = $1
	`

	var (
		idValue            string
		customerCode, name string
		email, phone       sql.NullString
		address, city      sql.NullString
		country, taxID     sql.NullString
		notes              sql.NullString
		creditLimit        float64
		currentBalance     float64
		isActive           bool
		createdAtRaw       any
		updatedAtRaw       any
	)

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&idValue, &customerCode, &name,
		&email, &phone,
		&address, &city, &country,
		&taxID, &creditLimit, &currentBalance,
		&notes, &isActive,
		&createdAtRaw, &updatedAtRaw,
	)
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	createdAt, err := parseDatabaseTimestamp(createdAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	updatedAt, err := parseDatabaseTimestamp(updatedAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	parsedID, err := uuid.Parse(idValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse customer id: %w", err)
	}

	customer := &Customer{
		ID:             parsedID,
		Code:           customerCode,
		Name:           name,
		Email:          nullableString(email),
		Phone:          nullableString(phone),
		Address:        nullableString(address),
		City:           nullableString(city),
		Country:        nullableString(country),
		TaxID:          nullableString(taxID),
		CreditLimit:    creditLimit,
		CurrentBalance: currentBalance,
		Notes:          nullableString(notes),
		IsActive:       isActive,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
	return customer, nil
}

// List retrieves customers with pagination and filters
func (r *Repository) List(ctx context.Context, req *CustomerListRequest) ([]Customer, int, error) {
	page := req.Page
	perPage := req.PerPage
	search := req.Search
	isActive := req.IsActive
	offset := (page - 1) * perPage

	query := `
		SELECT id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM customers WHERE 1=1`

	args := []interface{}{}

	if search != "" {
		// Enhanced search with partial words support - split search into words
		searchWords := []string{}
		currentWord := ""
		for _, char := range search {
			if char == ' ' {
				if currentWord != "" {
					searchWords = append(searchWords, currentWord)
					currentWord = ""
				}
			} else {
				currentWord += string(char)
			}
		}
		if currentWord != "" {
			searchWords = append(searchWords, currentWord)
		}

		// Build search conditions for each word
		searchConditions := []string{}
		for _, word := range searchWords {
			if len(word) >= 2 { // Only search for words with 2+ characters
				searchConditions = append(searchConditions, `LOWER(name) LIKE '%`+strings.ToLower(word)+`%'`)
				searchConditions = append(searchConditions, `LOWER(code) LIKE '%`+strings.ToLower(word)+`%'`)
				searchConditions = append(searchConditions, `LOWER(email) LIKE '%`+strings.ToLower(word)+`%'`)
				searchConditions = append(searchConditions, `LOWER(phone) LIKE '%`+strings.ToLower(word)+`%'`)
				searchConditions = append(searchConditions, `LOWER(address) LIKE '%`+strings.ToLower(word)+`%'`)
				searchConditions = append(searchConditions, `LOWER(city) LIKE '%`+strings.ToLower(word)+`%'`)
			}
		}

		if len(searchConditions) > 0 {
			searchCondition := "(" + searchConditions[0]
			for i := 1; i < len(searchConditions); i++ {
				searchCondition += " OR " + searchConditions[i]
			}
			searchCondition += ")"
			query += " AND " + searchCondition
			countQuery += " AND " + searchCondition
		}
	}

	if isActive != nil {
		// After search with raw SQL, we need to use parameter position correctly
		// Since search uses raw SQL, we start parameters from $1 for isActive
		paramNum := 1
		if len(args) > 0 {
			paramNum = len(args) + 1
		}
		query += fmt.Sprintf(" AND is_active = $%d", paramNum)
		countQuery += fmt.Sprintf(" AND is_active = $%d", paramNum)
		args = append(args, *isActive)
	}

	// Advanced filtering: by city
	if req.City != "" {
		paramNum := len(args) + 1
		query += fmt.Sprintf(" AND city = $%d", paramNum)
		countQuery += fmt.Sprintf(" AND city = $%d", paramNum)
		args = append(args, req.City)
	}

	// Advanced filtering: has debt
	if req.HasDebt != nil {
		if *req.HasDebt {
			query += " AND current_balance > 0"
			countQuery += " AND current_balance > 0"
		} else {
			query += " AND current_balance = 0"
			countQuery += " AND current_balance = 0"
		}
	}

	// Advanced filtering: is overdue
	if req.IsOverdue != nil && *req.IsOverdue {
		query += ` AND id IN (
			SELECT DISTINCT customer_id FROM debts
			WHERE remaining_amount > 0
				AND due_date < date('now')
			AND status = 'pending'
		)`
		countQuery += ` AND EXISTS (
			SELECT 1 FROM debts d
			WHERE d.customer_id = customers.id
			AND d.remaining_amount > 0
				AND d.due_date < date('now')
			AND d.status = 'pending'
		)`
	}

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Add pagination
	paramNum := len(args) + 1
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramNum, paramNum+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var (
			idValue            string
			code, name         string
			email, phone       sql.NullString
			address, city      sql.NullString
			country, taxID     sql.NullString
			notes              sql.NullString
			creditLimit        float64
			currentBalance     float64
			isActive           bool
			createdAtRaw       any
			updatedAtRaw       any
		)

		if err := rows.Scan(
			&idValue, &code, &name,
			&email, &phone,
			&address, &city, &country,
			&taxID, &creditLimit, &currentBalance,
			&notes, &isActive,
			&createdAtRaw, &updatedAtRaw,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer row: %w", err)
		}

		createdAt, err := parseDatabaseTimestamp(createdAtRaw)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse created_at: %w", err)
		}
		updatedAt, err := parseDatabaseTimestamp(updatedAtRaw)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		parsedID, err := uuid.Parse(idValue)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse customer id: %w", err)
		}

		customers = append(customers, Customer{
			ID:             parsedID,
			Code:           code,
			Name:           name,
			Email:          nullableString(email),
			Phone:          nullableString(phone),
			Address:        nullableString(address),
			City:           nullableString(city),
			Country:        nullableString(country),
			TaxID:          nullableString(taxID),
			CreditLimit:    creditLimit,
			CurrentBalance: currentBalance,
			Notes:          nullableString(notes),
			IsActive:       isActive,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate customer rows: %w", err)
	}

	return customers, total, nil
}

// Update updates a customer
func (r *Repository) Update(ctx context.Context, customer *Customer) error {
	query := `
		UPDATE customers
		SET code = $2, name = $3, email = $4, phone = $5, address = $6, city = $7, country = $8, tax_id = $9, credit_limit = $10, notes = $11, is_active = $12, updated_at = $13
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.Code, customer.Name, customer.Email, customer.Phone,
		customer.Address, customer.City, customer.Country, customer.TaxID,
		customer.CreditLimit, customer.Notes, customer.IsActive, customer.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

// Delete deletes a customer
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM customers WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

// HasActiveTransactions checks if customer has active sales/transactions
func (r *Repository) HasActiveTransactions(ctx context.Context, customerID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM sales 
			WHERE customer_id = $1 
			AND status != 'cancelled'
			AND created_at > NOW() - INTERVAL '1 year'
		)
	`

	var hasTransactions bool
	err := r.db.GetContext(ctx, &hasTransactions, query, customerID)
	if err != nil {
		// If table doesn't exist, treat as no transactions
		if err.Error() == `pq: relation "sales" does not exist` {
			return false, nil
		}
		return false, fmt.Errorf("failed to check active transactions: %w", err)
	}

	return hasTransactions, nil
}

// HasActiveWarranties checks if customer has active warranties
func (r *Repository) HasActiveWarranties(ctx context.Context, customerID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM warranties 
			WHERE customer_id = $1 
			AND is_active = true 
			AND expires_at > NOW()
		)
	`

	var hasWarranties bool
	err := r.db.GetContext(ctx, &hasWarranties, query, customerID)
	if err != nil {
		// If table doesn't exist, treat as no warranties
		if err.Error() == `pq: relation "warranties" does not exist` {
			return false, nil
		}
		return false, fmt.Errorf("failed to check active warranties: %w", err)
	}

	return hasWarranties, nil
}

// UpdateBalance updates customer balance
func (r *Repository) UpdateBalance(ctx context.Context, customerID uuid.UUID, amount float64) error {
	query := `
		UPDATE customers
		SET current_balance = current_balance + $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, amount, customerID)
	if err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

// GetCustomerLedger retrieves customer ledger entries
func (r *Repository) GetCustomerLedger(ctx context.Context, customerID uuid.UUID) ([]LedgerEntry, float64, float64, float64, error) {
	// Get ledger entries
	query := `
		SELECT id, customer_id, type, amount, balance, description, reference_id, created_at
		FROM customer_ledger
		WHERE customer_id = $1
		ORDER BY created_at ASC
	`
	var entries []LedgerEntry
	err := r.db.SelectContext(ctx, &entries, query, customerID)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("failed to get customer ledger: %w", err)
	}

	// Get totals
	var totalPurchases, totalPayments, currentBalance float64
	query = `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE 0 END), 0) as total_purchases,
			COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0) as total_payments,
			COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END), 0) as current_balance
		FROM customer_ledger
		WHERE customer_id = $1
	`
	err = r.db.QueryRowContext(ctx, query, customerID).Scan(&totalPurchases, &totalPayments, &currentBalance)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("failed to get customer totals: %w", err)
	}

	return entries, totalPurchases, totalPayments, currentBalance, nil
}

// GetFinancialTimeline retrieves comprehensive financial timeline including sales, payments, returns, etc.
func (r *Repository) GetFinancialTimeline(ctx context.Context, customerID uuid.UUID) ([]LedgerEntry, error) {
	query := `
		SELECT id, customer_id, type, amount, balance, description, reference_id, created_at
		FROM customer_ledger
		WHERE customer_id = $1
		ORDER BY created_at ASC
	`
	var entries []LedgerEntry
	err := r.db.SelectContext(ctx, &entries, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial timeline: %w", err)
	}

	return entries, nil
}

// AddPayment adds a payment to customer ledger
func (r *Repository) AddPayment(ctx context.Context, payment *PaymentResponse) error {
	query := `
		INSERT INTO customer_payments (id, customer_id, amount, payment_date, method, reference, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.CustomerID, payment.Amount, payment.PaymentDate,
		payment.Method, payment.Reference, payment.Notes, payment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add payment: %w", err)
	}

	// Add to ledger
	ledgerQuery := `
		INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at)
		SELECT $1, $2, 'credit', $3, 
			(SELECT COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END), 0) FROM customer_ledger WHERE customer_id = $2) - $3,
			$4, $5, $6
	`
	_, err = r.db.ExecContext(ctx, ledgerQuery,
		uuid.New(), payment.CustomerID, payment.Amount,
		"Payment: "+payment.Method, payment.ID, payment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add ledger entry: %w", err)
	}

	// Update debts table - reduce remaining_amount for unpaid debts
	// Start with the oldest debt first
	updateDebtsQuery := `
		WITH ordered_debts AS (
			SELECT id, remaining_amount 
			FROM debts 
			WHERE customer_id = $1 
			AND status IN ('pending', 'partial', 'overdue')
			AND remaining_amount > 0
			ORDER BY due_date ASC
		)
		UPDATE debts 
		SET remaining_amount = GREATEST(0, remaining_amount - $2),
		    updated_at = NOW()
		WHERE id = (SELECT id FROM ordered_debts LIMIT 1)
		RETURNING remaining_amount
	`
	var remainingAmount float64
	err = r.db.GetContext(ctx, &remainingAmount, updateDebtsQuery, payment.CustomerID, payment.Amount)
	if err != nil {
		// Log but don't fail if debts update fails
		fmt.Printf("Warning: failed to update debts: %v\n", err)
	}

	// Update debt status based on remaining amount
	if remainingAmount == 0 {
		updateStatusQuery := `
			UPDATE debts 
			SET status = 'paid',
			    updated_at = NOW()
			WHERE customer_id = $1 
			AND remaining_amount = 0
			AND status IN ('pending', 'partial', 'overdue')
		`
		_, err = r.db.ExecContext(ctx, updateStatusQuery, payment.CustomerID)
		if err != nil {
			fmt.Printf("Warning: failed to update debt status: %v\n", err)
		}
	}

	return nil
}

// AddLedgerEntry adds a ledger entry
func (r *Repository) AddLedgerEntry(ctx context.Context, customerID uuid.UUID, entryType string, amount float64, description string, referenceID uuid.UUID) error {
	ledgerQuery := `
		INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at)
		SELECT $1, $2, $3, $4, 
			(SELECT COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END), 0) FROM customer_ledger WHERE customer_id = $2) + 
			CASE WHEN $3 = 'debit' THEN $4 ELSE -$4 END,
			$5, $6, $7
	`
	_, err := r.db.ExecContext(ctx, ledgerQuery,
		uuid.New(), customerID, entryType, amount, description, referenceID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to add ledger entry: %w", err)
	}

	return nil
}

// CreateDebtEntry creates a new debt entry
func (r *Repository) CreateDebtEntry(ctx context.Context, debt *DebtEntry) error {
	query := `
		INSERT INTO customer_debts (id, customer_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		debt.ID, debt.CustomerID, debt.Amount, debt.ReferenceID, debt.ReferenceType,
		debt.DueDate, debt.IsPaid, debt.PaidAmount, debt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create debt entry: %w", err)
	}
	return nil
}

// GetDebtEntries retrieves debt entries for a customer
func (r *Repository) GetDebtEntries(ctx context.Context, customerID uuid.UUID) ([]DebtEntry, error) {
	query := `
		SELECT id, customer_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at
		FROM customer_debts
		WHERE customer_id = $1
		ORDER BY due_date ASC
	`
	var debts []DebtEntry
	err := r.db.SelectContext(ctx, &debts, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt entries: %w", err)
	}
	return debts, nil
}

// UpdateDebtPayment updates payment for a debt entry
func (r *Repository) UpdateDebtPayment(ctx context.Context, debtID uuid.UUID, paymentAmount float64) error {
	query := `
		UPDATE customer_debts
		SET paid_amount = paid_amount + $1,
			is_paid = (paid_amount + $1) >= amount
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, paymentAmount, debtID)
	if err != nil {
		return fmt.Errorf("failed to update debt payment: %w", err)
	}
	return nil
}

// CreateDebtCollection creates a new debt collection action
func (r *Repository) CreateDebtCollection(ctx context.Context, collection *DebtCollection) error {
	query := `
		INSERT INTO debt_collections (id, customer_id, type, status, notes, scheduled_date, completed_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		collection.ID, collection.CustomerID, collection.Type, collection.Status,
		collection.Notes, collection.ScheduledDate, collection.CompletedDate, collection.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create debt collection: %w", err)
	}
	return nil
}

// GetDebtCollections retrieves debt collection actions for a customer
func (r *Repository) GetDebtCollections(ctx context.Context, customerID uuid.UUID) ([]DebtCollection, error) {
	query := `
		SELECT id, customer_id, type, status, notes, scheduled_date, completed_date, created_at
		FROM debt_collections
		WHERE customer_id = $1
		ORDER BY scheduled_date DESC
	`
	var collections []DebtCollection
	err := r.db.SelectContext(ctx, &collections, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt collections: %w", err)
	}
	return collections, nil
}

// GetPendingDebtCollections retrieves pending debt collection actions
func (r *Repository) GetPendingDebtCollections(ctx context.Context) ([]DebtCollection, error) {
	query := `
		SELECT dc.id, dc.customer_id, dc.type, dc.status, dc.notes, dc.scheduled_date, dc.completed_date, dc.created_at
		FROM debt_collections dc
		JOIN customers c ON dc.customer_id = c.id
		WHERE dc.status = 'pending'
		AND dc.scheduled_date <= NOW()
		ORDER BY dc.scheduled_date ASC
	`
	var collections []DebtCollection
	err := r.db.SelectContext(ctx, &collections, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending debt collections: %w", err)
	}
	return collections, nil
}
