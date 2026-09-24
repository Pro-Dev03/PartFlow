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
	"github.com/partflow/smart-store/internal/accounting"
	dbutil "github.com/partflow/smart-store/internal/database"
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
	return r.CreateWithOpeningDebt(ctx, customer, 0)
}

// CreateWithOpeningDebt inserts a customer and its opening balance as one
// transaction so a failed opening-debt write cannot leave a partial account.
func (r *Repository) CreateWithOpeningDebt(ctx context.Context, customer *Customer, openingDebt float64) error {
	if openingDebt < 0 {
		return ErrPaymentAmountInvalid
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin customer creation: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO customers (id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, tx.Rebind(query), customer.ID, customer.Code, customer.Name, customer.Email, customer.Phone, customer.Address, customer.City,
		customer.Country, customer.TaxID, customer.CreditLimit, customer.CurrentBalance, customer.Notes, customer.IsActive, customer.CreatedAt, customer.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	if openingDebt > 0 {
		now := time.Now().UTC()
		debtID := uuid.New()
		if dbutil.IsSQLite(r.db) {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO debts (id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, notes, created_at, updated_at) VALUES (?, ?, NULL, ?, 0, ?, ?, 'pending', ?, ?, ?)`), debtID.String(), customer.ID.String(), openingDebt, openingDebt, now, "opening_debt", now, now); err != nil {
				return fmt.Errorf("create local opening debt entry: %w", err)
			}
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_ledger (id, customer_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT ?, ?, 'debit', 'SALE', ?, COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id = ?), 0) + ?, ?, NULL, ?`), uuid.New().String(), customer.ID.String(), openingDebt, customer.ID.String(), openingDebt, "\u062F\u064A\u0646 \u0633\u0627\u0628\u0642 \u0642\u0628\u0644 \u0627\u0633\u062A\u062E\u062F\u0627\u0645 \u0627\u0644\u0646\u0638\u0627\u0645", now); err != nil {
				return fmt.Errorf("create local opening debt ledger entry: %w", err)
			}
		} else {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_debts (id, customer_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at) VALUES (?, ?, ?, NULL, 'opening_debt', ?, FALSE, 0, ?)`), debtID, customer.ID, openingDebt, now, now); err != nil {
				return fmt.Errorf("create opening debt entry: %w", err)
			}
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at) SELECT ?, ?, 'debit', ?, COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id = ?), 0) + ?, ?, NULL, ?`), uuid.New(), customer.ID, openingDebt, customer.ID, openingDebt, "\u062F\u064A\u0646 \u0633\u0627\u0628\u0642 \u0642\u0628\u0644 \u0627\u0633\u062A\u062E\u062F\u0627\u0645 \u0627\u0644\u0646\u0638\u0627\u0645", now); err != nil {
				return fmt.Errorf("create opening debt ledger entry: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE customers SET current_balance = ? WHERE id = ?`), openingDebt, customer.ID); err != nil {
			return fmt.Errorf("set opening customer balance: %w", err)
		}
		customer.CurrentBalance = openingDebt
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit customer creation: %w", err)
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
		// database/sql may serialize a Go time.Time using time.String(),
		// which includes a monotonic-clock suffix ("m=+â€¦"). That suffix is
		// not part of the persisted wall-clock value and must be removed before
		// parsing SQLite rows written by runtime updates.
		v = strings.TrimSpace(v)
		if monotonic := strings.Index(v, " m="); monotonic >= 0 {
			v = strings.TrimSpace(v[:monotonic])
		}
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05.999999999 -0700 MST",
			"2006-01-02 15:04:05 -0700 MST",
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

func nullableUUID(value uuid.UUID) interface{} {
	if value == uuid.Nil {
		return nil
	}
	return value
}

// sqliteTableExists keeps list queries compatible with the small SQLite
// schemas used by migrations and unit tests. The production local database
// has both tables and receives the calculated financial fields below.
func sqliteTableExists(db *sqlx.DB, table string) bool {
	var exists int
	if err := db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = $1)`, table); err != nil {
		return false
	}
	return exists == 1
}

// GetByID retrieves a customer by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE id = $1
	`

	var (
		idValue        string
		code, name     string
		email, phone   sql.NullString
		address, city  sql.NullString
		country, taxID sql.NullString
		notes          sql.NullString
		creditLimit    float64
		currentBalance float64
		isActive       bool
		createdAtRaw   any
		updatedAtRaw   any
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

	// Customer cards display transaction totals. Calculate them from the
	// source tables so they cannot become stale when a sale or debt changes.
	// Keep a fallback for minimal/legacy SQLite schemas used by migrations and
	// unit tests where those source tables are not present yet.
	withFinancialSummary := !dbutil.IsSQLite(r.db) ||
		(sqliteTableExists(r.db, "sales") && sqliteTableExists(r.db, "debts"))
	selectColumns := `id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at`
	if withFinancialSummary {
		selectColumns += `,
			COALESCE((SELECT SUM(s.total_amount) FROM sales s
				WHERE s.customer_id = customers.id
				  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 0) AS total_purchases,
			COALESCE((SELECT SUM(COALESCE(s.paid_amount, 0)) FROM sales s
				WHERE s.customer_id = customers.id
				  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')), 0) AS paid_amount,
			COALESCE((SELECT SUM(CASE WHEN d.remaining_amount > 0 THEN d.remaining_amount ELSE 0 END)
				FROM debts d WHERE d.customer_id = customers.id
				  AND LOWER(COALESCE(d.status, 'pending')) NOT IN ('paid', 'completed', 'settled', 'cancelled', 'canceled', 'reversed')), 0) AS outstanding,
			(SELECT MAX(s.sale_date) FROM sales s
				WHERE s.customer_id = customers.id
				  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) AS last_purchase`
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM customers
		WHERE 1=1
	`, selectColumns)
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
		storeDate, err := accounting.StoreDate(accounting.StoreNow())
		if err != nil {
			return nil, 0, fmt.Errorf("failed to calculate overdue filter date: %w", err)
		}
		dueDateParam := len(args) + 1
		query += ` AND id IN (
			SELECT DISTINCT customer_id FROM debts
			WHERE remaining_amount > 0
				AND due_date < $` + fmt.Sprint(dueDateParam) + `
			AND status = 'pending'
		)`
		countQuery += ` AND EXISTS (
			SELECT 1 FROM debts d
			WHERE d.customer_id = customers.id
			AND d.remaining_amount > 0
				AND d.due_date < $` + fmt.Sprint(dueDateParam) + `
			AND d.status = 'pending'
		)`
		args = append(args, storeDate)
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
			idValue         string
			code, name      string
			email, phone    sql.NullString
			address, city   sql.NullString
			country, taxID  sql.NullString
			notes           sql.NullString
			creditLimit     float64
			currentBalance  float64
			isActive        bool
			createdAtRaw    any
			updatedAtRaw    any
			totalPurchases  float64
			paidAmount      float64
			outstanding     float64
			lastPurchaseRaw any
		)

		scanArgs := []any{
			&idValue, &code, &name,
			&email, &phone,
			&address, &city, &country,
			&taxID, &creditLimit, &currentBalance,
			&notes, &isActive,
			&createdAtRaw, &updatedAtRaw,
		}
		if withFinancialSummary {
			scanArgs = append(scanArgs, &totalPurchases, &paidAmount, &outstanding, &lastPurchaseRaw)
		}
		if err := rows.Scan(scanArgs...); err != nil {
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
		var lastPurchase *time.Time
		if withFinancialSummary && lastPurchaseRaw != nil {
			parsed, parseErr := parseDatabaseTimestamp(lastPurchaseRaw)
			if parseErr != nil {
				return nil, 0, fmt.Errorf("failed to parse last_purchase: %w", parseErr)
			}
			if !parsed.IsZero() {
				lastPurchase = &parsed
			}
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
			TotalPurchases: totalPurchases,
			PaidAmount:     paidAmount,
			Outstanding:    outstanding,
			LastPurchase:   lastPurchase,
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

// Delete archives a customer while preserving the financial and operational
// history that references it. Historical sales, payments, debts, returns, and
// ledger entries must remain available for reconciliation and audit reports.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`UPDATE customers SET is_active = FALSE, updated_at = %s WHERE id = $1`, dbutil.NowSQL(r.db))
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to archive customer: %w", err)
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
			AND created_at > %s
		)
	`
	if dbutil.IsSQLite(r.db) {
		query = fmt.Sprintf(query, "datetime('now', '-1 year')")
	} else {
		query = fmt.Sprintf(query, "NOW() - INTERVAL '1 year'")
	}

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
	if dbutil.IsSQLite(r.db) {
		// The local schema keeps warranty claims (rather than the optional
		// cloud warranties view). Treat pending/in-progress claims as active.
		var active bool
		err := r.db.GetContext(ctx, &active, `SELECT EXISTS(SELECT 1 FROM warranty_claims WHERE customer_id = $1 AND status IN ('pending', 'in_progress', 'approved'))`, customerID)
		if err == nil {
			return active, nil
		}
		return false, nil
	}
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
	query := fmt.Sprintf(`
		UPDATE customers
		SET current_balance = current_balance + $1, updated_at = %s
		WHERE id = $2
	`, dbutil.NowSQL(r.db))
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
		SELECT id, customer_id, type, amount,
			SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END)
				OVER (PARTITION BY customer_id ORDER BY created_at ASC, id ASC ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS balance,
			description, reference_id, created_at
		FROM customer_ledger
		WHERE customer_id = $1
		ORDER BY created_at ASC
	`
	var entries []LedgerEntry
	var err error
	if dbutil.IsSQLite(r.db) {
		var rows []struct {
			ID              string         `db:"id"`
			CustomerID      string         `db:"customer_id"`
			Type            string         `db:"type"`
			TransactionType sql.NullString `db:"transaction_type"`
			Amount          float64        `db:"amount"`
			Balance         float64        `db:"balance"`
			Description     sql.NullString `db:"description"`
			ReferenceID     sql.NullString `db:"reference_id"`
			CreatedAt       string         `db:"created_at"`
		}
		err = r.db.SelectContext(ctx, &rows, `SELECT id, customer_id, COALESCE(type, '') AS type, transaction_type, amount, SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) OVER (PARTITION BY customer_id ORDER BY created_at ASC, id ASC ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS balance, description, reference_id, created_at FROM customer_ledger WHERE customer_id = $1 ORDER BY created_at ASC, id ASC`, customerID)
		if err == nil {
			entries = make([]LedgerEntry, 0, len(rows))
			for _, row := range rows {
				entryID, parseErr := uuid.Parse(row.ID)
				if parseErr != nil {
					return nil, 0, 0, 0, parseErr
				}
				entryCustomer, parseErr := uuid.Parse(row.CustomerID)
				if parseErr != nil {
					return nil, 0, 0, 0, parseErr
				}
				created, parseErr := dbutil.ParseTimestamp(row.CreatedAt)
				if parseErr != nil {
					return nil, 0, 0, 0, parseErr
				}
				typ := row.Type
				if typ == "" {
					typ = strings.ToLower(row.TransactionType.String)
				}
				entry := LedgerEntry{ID: entryID, CustomerID: entryCustomer, Type: typ, Amount: row.Amount, Balance: row.Balance, CreatedAt: created}
				if row.Description.Valid {
					entry.Description = row.Description.String
				}
				if row.ReferenceID.Valid && row.ReferenceID.String != "" {
					if ref, e := uuid.Parse(row.ReferenceID.String); e == nil {
						entry.ReferenceID = &ref
					}
				}
				entries = append(entries, entry)
			}
		}
	} else {
		err = r.db.SelectContext(ctx, &entries, query, customerID)
	}
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
	if dbutil.IsSQLite(r.db) {
		entries, _, _, _, err := r.GetCustomerLedger(ctx, customerID)
		return entries, err
	}
	query := `
		SELECT id, customer_id, type, amount,
			SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END)
				OVER (PARTITION BY customer_id ORDER BY created_at ASC, id ASC ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS balance,
			description, reference_id, created_at
		FROM customer_ledger
		WHERE customer_id = $1
		ORDER BY created_at ASC, id ASC
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
	return r.RecordPaymentTransaction(ctx, payment, false)
}

// RecordPaymentTransaction commits the payment, customer ledger, debt allocation,
// and denormalized customer balance as one operation. strictDebtCoverage is used
// by explicit debt collection so concurrent collections cannot over-allocate.
func (r *Repository) RecordPaymentTransaction(ctx context.Context, payment *PaymentResponse, strictDebtCoverage bool) error {
	if payment.Amount <= 0 {
		return ErrPaymentAmountInvalid
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin customer payment transaction: %w", err)
	}
	defer tx.Rollback()

	lock := ""
	if !dbutil.IsSQLite(r.db) {
		lock = " FOR UPDATE"
	}
	var currentBalance float64
	if err := tx.GetContext(ctx, &currentBalance, tx.Rebind(`SELECT COALESCE(current_balance, 0) FROM customers WHERE id = ?`)+lock, payment.CustomerID.String()); err != nil {
		if err == sql.ErrNoRows {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("lock customer balance: %w", err)
	}
	if payment.Amount > currentBalance+0.000001 {
		return ErrPaymentExceedsBalance
	}

	if payment.Reference != nil && strings.TrimSpace(*payment.Reference) != "" {
		column := "reference_number"
		if dbutil.IsSQLite(r.db) {
			column = "reference"
		}
		var duplicate bool
		if err := tx.GetContext(ctx, &duplicate, tx.Rebind(`SELECT EXISTS (SELECT 1 FROM payments WHERE customer_id = ? AND `+column+` = ?)`), payment.CustomerID.String(), strings.TrimSpace(*payment.Reference)); err != nil {
			return fmt.Errorf("check duplicate customer payment reference: %w", err)
		}
		if duplicate {
			return ErrPaymentDuplicate
		}
	}

	type debtRow struct {
		ID     string  `db:"id"`
		Amount float64 `db:"amount"`
		Paid   float64 `db:"paid_amount"`
	}
	var debts []debtRow
	usesSalesDebts := dbutil.IsSQLite(r.db) || !strictDebtCoverage
	debtQuery := `SELECT id, amount, COALESCE(paid_amount, 0) AS paid_amount FROM customer_debts
		WHERE customer_id = ? AND COALESCE(is_paid, FALSE) = FALSE AND amount > COALESCE(paid_amount, 0)
		ORDER BY due_date, created_at, id`
	if dbutil.IsSQLite(r.db) {
		debtQuery = `SELECT id, amount, COALESCE(paid_amount, 0) AS paid_amount FROM debts
			WHERE customer_id = ? AND COALESCE(status, 'pending') IN ('pending', 'partial', 'overdue')
			AND amount > COALESCE(paid_amount, 0)
			ORDER BY due_date, created_at, id`
	} else if !strictDebtCoverage {
		// Ordinary account payments have historically been applied to POS sale
		// debts, which live in `debts`; explicit debt collection uses
		// `customer_debts` on PostgreSQL.
		debtQuery = `SELECT id, amount, COALESCE(paid_amount, 0) AS paid_amount FROM debts
			WHERE customer_id = ? AND COALESCE(status, 'pending') IN ('pending', 'partial', 'overdue')
			AND amount > COALESCE(paid_amount, 0)
			ORDER BY due_date, created_at, id`
	}
	if !dbutil.IsSQLite(r.db) {
		debtQuery += ` FOR UPDATE`
	}
	if err := tx.SelectContext(ctx, &debts, tx.Rebind(debtQuery), payment.CustomerID.String()); err != nil {
		return fmt.Errorf("load customer debts for payment: %w", err)
	}
	var outstanding float64
	for _, debt := range debts {
		if remaining := debt.Amount - debt.Paid; remaining > 0 {
			outstanding += remaining
		}
	}
	if strictDebtCoverage && payment.Amount > outstanding+0.000001 {
		return ErrPaymentExceedsBalance
	}

	if err := r.insertPaymentAndLedgerTx(ctx, tx, payment); err != nil {
		return err
	}
	remainingPayment := payment.Amount
	for _, debt := range debts {
		debtRemaining := debt.Amount - debt.Paid
		if debtRemaining <= 0 || remainingPayment <= 0.000001 {
			continue
		}
		applied := debtRemaining
		if applied > remainingPayment {
			applied = remainingPayment
		}
		paid := debt.Paid + applied
		remaining := debt.Amount - paid
		var update string
		if usesSalesDebts {
			status := "partial"
			if remaining <= 0.000001 {
				status = "paid"
				remaining = 0
				paid = debt.Amount
			}
			update = `UPDATE debts SET paid_amount = ?, remaining_amount = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND customer_id = ?`
			result, err := tx.ExecContext(ctx, tx.Rebind(update), paid, remaining, status, debt.ID, payment.CustomerID.String())
			if err != nil {
				return fmt.Errorf("allocate customer payment to debt: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected != 1 {
				return fmt.Errorf("customer debt changed during payment allocation")
			}
		} else {
			isPaid := remaining <= 0.000001
			if isPaid {
				remaining = 0
				paid = debt.Amount
			}
			update = `UPDATE customer_debts SET paid_amount = ?, is_paid = ? WHERE id = ? AND customer_id = ?`
			result, err := tx.ExecContext(ctx, tx.Rebind(update), paid, isPaid, debt.ID, payment.CustomerID.String())
			if err != nil {
				return fmt.Errorf("allocate customer payment to debt: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected != 1 {
				return fmt.Errorf("customer debt changed during payment allocation")
			}
		}
		remainingPayment -= applied
	}

	updateBalance := fmt.Sprintf(`UPDATE customers SET current_balance = COALESCE(current_balance, 0) - ?, updated_at = %s WHERE id = ? AND COALESCE(current_balance, 0) >= ?`, dbutil.NowSQL(r.db))
	result, err := tx.ExecContext(ctx, tx.Rebind(updateBalance), payment.Amount, payment.CustomerID.String(), payment.Amount)
	if err != nil {
		return fmt.Errorf("update customer balance for payment: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrPaymentExceedsBalance
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit customer payment transaction: %w", err)
	}
	return nil
}

func (r *Repository) insertPaymentAndLedgerTx(ctx context.Context, tx *sqlx.Tx, payment *PaymentResponse) error {
	paymentNumber := "PAY-" + payment.ID.String()[:8]
	if dbutil.IsSQLite(r.db) {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO payments (id, transaction_number, customer_id, amount, payment_method, reference, notes, payment_date, created_at, updated_at, payment_status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'completed')`), payment.ID.String(), paymentNumber, payment.CustomerID.String(), payment.Amount, payment.Method, payment.Reference, payment.Notes, payment.PaymentDate, payment.CreatedAt, payment.CreatedAt); err != nil {
			return fmt.Errorf("failed to add payment: %w", err)
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_ledger (id, customer_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT ?, ?, 'credit', 'PAYMENT', ?, COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id = ?), 0) - ?, ?, ?, ?`), uuid.New().String(), payment.CustomerID.String(), payment.Amount, payment.CustomerID.String(), payment.Amount, "Payment: "+payment.Method, payment.ID.String(), payment.CreatedAt); err != nil {
			return fmt.Errorf("failed to add customer ledger entry: %w", err)
		}
		return nil
	}
	referenceNumber := paymentNumber
	if payment.Reference != nil && strings.TrimSpace(*payment.Reference) != "" {
		referenceNumber = strings.TrimSpace(*payment.Reference)
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO payments (id, reference_number, customer_id, amount, payment_method, payment_date, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`), payment.ID, referenceNumber, payment.CustomerID, payment.Amount, payment.Method, payment.PaymentDate, payment.Notes, payment.CreatedAt, payment.CreatedAt); err != nil {
		return fmt.Errorf("failed to add payment: %w", err)
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at) SELECT ?, ?, 'credit', ?, COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id = ?), 0) - ?, ?, ?, ?`), uuid.New(), payment.CustomerID, payment.Amount, payment.CustomerID, payment.Amount, "Payment: "+payment.Method, payment.ID, payment.CreatedAt); err != nil {
		return fmt.Errorf("failed to add customer ledger entry: %w", err)
	}
	return nil
}

func (r *Repository) ApplyPaymentToOldestDebt(ctx context.Context, customerID uuid.UUID, amount float64) error {
	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, `UPDATE debts SET paid_amount = MIN(amount, paid_amount + $1), remaining_amount = MAX(0, remaining_amount - $1), status = CASE WHEN remaining_amount - $1 <= 0 THEN 'paid' ELSE 'partial' END, updated_at = CURRENT_TIMESTAMP WHERE id = (SELECT id FROM debts WHERE customer_id = $2 AND remaining_amount > 0 AND status IN ('pending','partial','overdue') ORDER BY due_date ASC, created_at ASC LIMIT 1)`, amount, customerID)
		return err
	}
	_, err := r.db.ExecContext(ctx, `WITH ordered_debts AS (SELECT id FROM debts WHERE customer_id = $1 AND status IN ('pending', 'partial', 'overdue') AND remaining_amount > 0 ORDER BY due_date ASC LIMIT 1) UPDATE debts SET remaining_amount = GREATEST(0, remaining_amount - $2), paid_amount = LEAST(amount, paid_amount + $2), updated_at = NOW(), status = CASE WHEN remaining_amount - $2 <= 0 THEN 'paid' ELSE status END WHERE id = (SELECT id FROM ordered_debts)`, customerID, amount)
	return err
}

func (r *Repository) HasPaymentReference(ctx context.Context, customerID uuid.UUID, reference string) (bool, error) {
	if strings.TrimSpace(reference) == "" {
		return false, nil
	}
	column := "reference_number"
	if dbutil.IsSQLite(r.db) {
		column = "reference"
	}
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM payments WHERE customer_id = $1 AND %s = $2)`, column)
	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, customerID, reference); err != nil {
		return false, fmt.Errorf("failed to check customer payment reference: %w", err)
	}
	return exists, nil
}

// AddLedgerEntry adds a ledger entry
func (r *Repository) AddLedgerEntry(ctx context.Context, customerID uuid.UUID, entryType string, amount float64, description string, referenceID uuid.UUID) error {
	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, `INSERT INTO customer_ledger (id, customer_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT $1, $2, $3, CASE WHEN $3 = 'debit' THEN 'SALE' ELSE 'ADJUSTMENT' END, $4, COALESCE((SELECT balance FROM customer_ledger WHERE customer_id = $2 ORDER BY created_at DESC LIMIT 1), 0) + CASE WHEN $3 = 'debit' THEN $4 ELSE -$4 END, $5, $6, CURRENT_TIMESTAMP`, uuid.New(), customerID, entryType, amount, description, referenceID)
		if err != nil {
			return fmt.Errorf("failed to add ledger entry: %w", err)
		}
		return nil
	}
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
	if dbutil.IsSQLite(r.db) {
		status := "pending"
		if debt.IsPaid {
			status = "paid"
		}
		_, err := r.db.ExecContext(ctx, `INSERT INTO debts (id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)`, debt.ID, debt.CustomerID, nullableUUID(debt.ReferenceID), debt.Amount, debt.PaidAmount, debt.Amount-debt.PaidAmount, debt.DueDate, status, debt.ReferenceType, debt.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to create debt entry: %w", err)
		}
		return nil
	}
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

// RecordDebtEntryTransaction writes the debt entry, customer ledger debit, and
// customer balance together so a failed step cannot leave accounting drift.
func (r *Repository) RecordDebtEntryTransaction(ctx context.Context, debt *DebtEntry, description string, enforceCreditLimit bool) error {
	if debt.Amount <= 0 {
		return ErrPaymentAmountInvalid
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin customer debt transaction: %w", err)
	}
	defer tx.Rollback()

	lock := ""
	if !dbutil.IsSQLite(r.db) {
		lock = " FOR UPDATE"
	}
	var balance, creditLimit float64
	if err := tx.QueryRowxContext(ctx, tx.Rebind(`SELECT COALESCE(current_balance, 0), COALESCE(credit_limit, 0) FROM customers WHERE id = ?`)+lock, debt.CustomerID.String()).Scan(&balance, &creditLimit); err != nil {
		if err == sql.ErrNoRows {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("lock customer debt balance: %w", err)
	}
	if enforceCreditLimit && balance+debt.Amount > creditLimit+0.000001 {
		return ErrCreditLimitExceeded
	}

	if dbutil.IsSQLite(r.db) {
		status := "pending"
		if debt.IsPaid {
			status = "paid"
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO debts (id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`), debt.ID.String(), debt.CustomerID.String(), nullableUUID(debt.ReferenceID), debt.Amount, debt.PaidAmount, debt.Amount-debt.PaidAmount, debt.DueDate, status, debt.ReferenceType, debt.CreatedAt, debt.CreatedAt); err != nil {
			return fmt.Errorf("create local customer debt entry: %w", err)
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_ledger (id, customer_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT ?, ?, 'debit', 'SALE', ?, COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id = ?), 0) + ?, ?, ?, ?`), uuid.New().String(), debt.CustomerID.String(), debt.Amount, debt.CustomerID.String(), debt.Amount, description, nullableUUID(debt.ReferenceID), debt.CreatedAt); err != nil {
			return fmt.Errorf("create customer debt ledger entry: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_debts (id, customer_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`), debt.ID, debt.CustomerID, debt.Amount, nullableUUID(debt.ReferenceID), debt.ReferenceType, debt.DueDate, debt.IsPaid, debt.PaidAmount, debt.CreatedAt); err != nil {
			return fmt.Errorf("create customer debt entry: %w", err)
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at) SELECT ?, ?, 'debit', ?, COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id = ?), 0) + ?, ?, ?, ?`), uuid.New(), debt.CustomerID, debt.Amount, debt.CustomerID, debt.Amount, description, nullableUUID(debt.ReferenceID), debt.CreatedAt); err != nil {
			return fmt.Errorf("create customer debt ledger entry: %w", err)
		}
	}

	updateBalance := fmt.Sprintf(`UPDATE customers SET current_balance = COALESCE(current_balance, 0) + ?, updated_at = %s WHERE id = ?`, dbutil.NowSQL(r.db))
	result, err := tx.ExecContext(ctx, tx.Rebind(updateBalance), debt.Amount, debt.CustomerID.String())
	if err != nil {
		return fmt.Errorf("update customer balance for debt: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrCustomerNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit customer debt transaction: %w", err)
	}
	return nil
}

// RecordAccountDebtTransaction atomically records an account debit and balance
// increase for debt paths that intentionally do not create a debt-row record.
func (r *Repository) RecordAccountDebtTransaction(ctx context.Context, customerID uuid.UUID, amount float64, description string, referenceID uuid.UUID) error {
	if amount <= 0 {
		return ErrPaymentAmountInvalid
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin customer account-debt transaction: %w", err)
	}
	defer tx.Rollback()

	lock := ""
	if !dbutil.IsSQLite(r.db) {
		lock = " FOR UPDATE"
	}
	var balance, creditLimit float64
	if err := tx.QueryRowxContext(ctx, tx.Rebind(`SELECT COALESCE(current_balance, 0), COALESCE(credit_limit, 0) FROM customers WHERE id = ?`)+lock, customerID.String()).Scan(&balance, &creditLimit); err != nil {
		if err == sql.ErrNoRows {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("lock customer account-debt balance: %w", err)
	}
	if balance+amount > creditLimit+0.000001 {
		return ErrCreditLimitExceeded
	}
	insert := `INSERT INTO customer_ledger (id, customer_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT ?, ?, 'debit', 'SALE', ?, COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id = ?), 0) + ?, ?, ?, ?`
	if dbutil.IsSQLite(r.db) {
		if _, err := tx.ExecContext(ctx, tx.Rebind(insert), uuid.New().String(), customerID.String(), amount, customerID.String(), amount, description, nullableUUID(referenceID), time.Now().UTC()); err != nil {
			return fmt.Errorf("record customer account-debt ledger: %w", err)
		}
	} else {
		insert = `INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at) SELECT ?, ?, 'debit', ?, COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id = ?), 0) + ?, ?, ?, ?`
		if _, err := tx.ExecContext(ctx, tx.Rebind(insert), uuid.New(), customerID, amount, customerID, amount, description, nullableUUID(referenceID), time.Now().UTC()); err != nil {
			return fmt.Errorf("record customer account-debt ledger: %w", err)
		}
	}

	update := fmt.Sprintf(`UPDATE customers SET current_balance = COALESCE(current_balance, 0) + ?, updated_at = %s WHERE id = ?`, dbutil.NowSQL(r.db))
	result, err := tx.ExecContext(ctx, tx.Rebind(update), amount, customerID.String())
	if err != nil {
		return fmt.Errorf("update customer account-debt balance: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrCustomerNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit customer account-debt transaction: %w", err)
	}
	return nil
}

// GetDebtEntries retrieves debt entries for a customer
func (r *Repository) GetDebtEntries(ctx context.Context, customerID uuid.UUID) ([]DebtEntry, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []struct {
			ID         string         `db:"id"`
			CustomerID string         `db:"customer_id"`
			SaleID     sql.NullString `db:"sale_id"`
			Amount     float64        `db:"amount"`
			Paid       float64        `db:"paid_amount"`
			Remaining  float64        `db:"remaining_amount"`
			DueDate    string         `db:"due_date"`
			Status     string         `db:"status"`
			CreatedAt  string         `db:"created_at"`
		}
		if err := r.db.SelectContext(ctx, &rows, `SELECT id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, created_at FROM debts WHERE customer_id = $1 ORDER BY due_date ASC`, customerID); err != nil {
			return nil, fmt.Errorf("failed to get debt entries: %w", err)
		}
		out := make([]DebtEntry, 0, len(rows))
		for _, row := range rows {
			id, err := uuid.Parse(row.ID)
			if err != nil {
				return nil, err
			}
			cid, err := uuid.Parse(row.CustomerID)
			if err != nil {
				return nil, err
			}
			due, err := dbutil.ParseTimestamp(row.DueDate)
			if err != nil {
				return nil, err
			}
			created, err := dbutil.ParseTimestamp(row.CreatedAt)
			if err != nil {
				return nil, err
			}
			ref := uuid.Nil
			if row.SaleID.Valid && row.SaleID.String != "" {
				ref, _ = uuid.Parse(row.SaleID.String)
			}
			out = append(out, DebtEntry{ID: id, CustomerID: cid, Amount: row.Amount, PaidAmount: row.Paid, ReferenceID: ref, ReferenceType: "sale", DueDate: due, IsPaid: row.Status == "paid" || row.Remaining <= 0, CreatedAt: created})
		}
		return out, nil
	}
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
	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, `UPDATE debts SET paid_amount = MIN(amount, paid_amount + $1), remaining_amount = MAX(0, remaining_amount - $1), status = CASE WHEN remaining_amount - $1 <= 0 THEN 'paid' ELSE 'partial' END, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, paymentAmount, debtID)
		if err != nil {
			return fmt.Errorf("failed to update debt payment: %w", err)
		}
		return nil
	}
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
	query := fmt.Sprintf(`
		SELECT dc.id, dc.customer_id, dc.type, dc.status, dc.notes, dc.scheduled_date, dc.completed_date, dc.created_at
		FROM debt_collections dc
		JOIN customers c ON dc.customer_id = c.id
		WHERE dc.status = 'pending'
		AND dc.scheduled_date <= %s
		ORDER BY dc.scheduled_date ASC
	`, dbutil.NowSQL(r.db))
	var collections []DebtCollection
	err := r.db.SelectContext(ctx, &collections, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending debt collections: %w", err)
	}
	return collections, nil
}
