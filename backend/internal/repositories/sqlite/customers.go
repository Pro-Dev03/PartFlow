package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Customer represents a customer entity
type Customer struct {
	ID             string    `db:"id"`
	Code           string    `db:"code"`
	Name           string    `db:"name"`
	Email          *string   `db:"email"`
	Phone          *string   `db:"phone"`
	Address        *string   `db:"address"`
	City           *string   `db:"city"`
	Country        *string   `db:"country"`
	CreditLimit    float64   `db:"credit_limit"`
	CurrentBalance float64   `db:"current_balance"`
	Notes          *string   `db:"notes"`
	IsActive       int       `db:"is_active"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// CustomerRepository implements repositories.CustomerRepository for SQLite
type CustomerRepository struct {
	db *sql.DB
}

// NewCustomerRepository creates a new SQLite customer repository
func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Create inserts a new customer
func (r *CustomerRepository) Create(ctx context.Context, customer interface{}) (string, error) {
	c, ok := customer.(*Customer)
	if !ok {
		return "", fmt.Errorf("invalid customer type")
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO customers (id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		id, c.Code, c.Name, c.Email, c.Phone, c.Address, c.City, c.Country, c.CreditLimit, c.CurrentBalance, c.Notes, now, now)

	if err != nil {
		return "", fmt.Errorf("insert customer: %w", err)
	}

	return id, nil
}

// GetByID retrieves a customer by ID
func (r *CustomerRepository) GetByID(ctx context.Context, id string) (interface{}, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE id = ? AND is_active = 1
	`

	var customer Customer
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID, &customer.Code, &customer.Name, &customer.Email, &customer.Phone, &customer.Address, &customer.City, &customer.Country,
		&customer.CreditLimit, &customer.CurrentBalance, &customer.Notes, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query customer: %w", err)
	}

	return &customer, nil
}

// GetByCode retrieves a customer by code
func (r *CustomerRepository) GetByCode(ctx context.Context, code string) (interface{}, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE code = ? AND is_active = 1
	`

	var customer Customer
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&customer.ID, &customer.Code, &customer.Name, &customer.Email, &customer.Phone, &customer.Address, &customer.City, &customer.Country,
		&customer.CreditLimit, &customer.CurrentBalance, &customer.Notes, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query customer by code: %w", err)
	}

	return &customer, nil
}

// List retrieves all active customers
func (r *CustomerRepository) List(ctx context.Context, limit, offset int) ([]interface{}, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE is_active = 1
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query customers: %w", err)
	}
	defer rows.Close()

	var customers []interface{}
	for rows.Next() {
		var customer Customer
		if err := rows.Scan(&customer.ID, &customer.Code, &customer.Name, &customer.Email, &customer.Phone, &customer.Address, &customer.City, &customer.Country,
			&customer.CreditLimit, &customer.CurrentBalance, &customer.Notes, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	return customers, rows.Err()
}

// Update modifies an existing customer
func (r *CustomerRepository) Update(ctx context.Context, id string, customer interface{}) error {
	c, ok := customer.(*Customer)
	if !ok {
		return fmt.Errorf("invalid customer type")
	}

	query := `
		UPDATE customers
		SET code = ?, name = ?, email = ?, phone = ?, address = ?, city = ?, country = ?, credit_limit = ?, current_balance = ?, notes = ?, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.ExecContext(ctx, query,
		c.Code, c.Name, c.Email, c.Phone, c.Address, c.City, c.Country, c.CreditLimit, c.CurrentBalance, c.Notes, time.Now(), id)

	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

// Delete soft-deletes a customer
func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE customers
		SET is_active = 0, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

// Search searches customers by name or code
func (r *CustomerRepository) Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error) {
	sql := `
		SELECT id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM customers
		WHERE is_active = 1 AND (name LIKE ? OR code LIKE ? OR phone LIKE ? OR email LIKE ?)
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sql, searchTerm, searchTerm, searchTerm, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search customers: %w", err)
	}
	defer rows.Close()

	var customers []interface{}
	for rows.Next() {
		var customer Customer
		if err := rows.Scan(&customer.ID, &customer.Code, &customer.Name, &customer.Email, &customer.Phone, &customer.Address, &customer.City, &customer.Country,
			&customer.CreditLimit, &customer.CurrentBalance, &customer.Notes, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	return customers, rows.Err()
}

// Count returns total number of active customers
func (r *CustomerRepository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM customers WHERE is_active = 1"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count customers: %w", err)
	}
	return count, nil
}
