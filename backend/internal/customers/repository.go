package customers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
		INSERT INTO customers (id, organization_id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.OrganizationID, customer.Code, customer.Name,
		customer.Email, customer.Phone, customer.Address, customer.City,
		customer.Country, customer.TaxID, customer.CreditLimit, customer.CurrentBalance,
		customer.Notes, customer.IsActive, customer.CreatedAt, customer.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	return nil
}

// GetByID retrieves a customer by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Customer, error) {
	query := `
		SELECT id, organization_id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE id = $1 AND organization_id = $2
	`
	var customer Customer
	err := r.db.GetContext(ctx, &customer, query, id, organizationID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}
	return &customer, nil
}

// GetByCode retrieves a customer by code
func (r *Repository) GetByCode(ctx context.Context, code string, organizationID uuid.UUID) (*Customer, error) {
	query := `
		SELECT id, organization_id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE code = $1 AND organization_id = $2
	`
	var customer Customer
	err := r.db.GetContext(ctx, &customer, query, code, organizationID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}
	return &customer, nil
}

// List retrieves customers with pagination and filters
func (r *Repository) List(ctx context.Context, organizationID uuid.UUID, page, perPage int, search string, isActive *bool) ([]Customer, int, error) {
	offset := (page - 1) * perPage

	query := `
		SELECT id, organization_id, code, name, email, phone, address, city, country, tax_id, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE organization_id = $1
	`
	args := []interface{}{organizationID}
	argCount := 1

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)", argCount, argCount, argCount, argCount)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
		argCount += 3
	}

	if isActive != nil {
		argCount++
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *isActive)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM customers WHERE organization_id = $1"
	countArgs := []interface{}{organizationID}
	countArgCount := 1

	if search != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)", countArgCount, countArgCount, countArgCount, countArgCount)
		searchPattern := "%" + search + "%"
		countArgs = append(countArgs, searchPattern, searchPattern, searchPattern, searchPattern)
		countArgCount += 3
	}

	if isActive != nil {
		countArgCount++
		countQuery += fmt.Sprintf(" AND is_active = $%d", countArgCount)
		countArgs = append(countArgs, *isActive)
	}

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, perPage, offset)

	var customers []Customer
	err = r.db.SelectContext(ctx, &customers, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}

	return customers, total, nil
}

// Update updates a customer
func (r *Repository) Update(ctx context.Context, customer *Customer) error {
	query := `
		UPDATE customers
		SET code = $2, name = $3, email = $4, phone = $5, address = $6, city = $7, country = $8, tax_id = $9, credit_limit = $10, notes = $11, is_active = $12, updated_at = $13
		WHERE id = $1 AND organization_id = $14
	`
	result, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.Code, customer.Name, customer.Email, customer.Phone,
		customer.Address, customer.City, customer.Country, customer.TaxID,
		customer.CreditLimit, customer.Notes, customer.IsActive, customer.UpdatedAt,
		customer.OrganizationID,
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
func (r *Repository) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM customers WHERE id = $1 AND organization_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

// UpdateBalance updates customer balance
func (r *Repository) UpdateBalance(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID, amount float64) error {
	query := `
		UPDATE customers
		SET current_balance = current_balance + $1, updated_at = NOW()
		WHERE id = $2 AND organization_id = $3
	`
	result, err := r.db.ExecContext(ctx, query, amount, customerID, organizationID)
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
func (r *Repository) GetCustomerLedger(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID) ([]LedgerEntry, float64, float64, float64, error) {
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

	return nil
}
