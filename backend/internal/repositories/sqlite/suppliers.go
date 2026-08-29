package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Supplier represents a supplier entity
type Supplier struct {
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

// SupplierRepository implements repositories.SupplierRepository for SQLite
type SupplierRepository struct {
	db *sql.DB
}

// NewSupplierRepository creates a new SQLite supplier repository
func NewSupplierRepository(db *sql.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// Create inserts a new supplier
func (r *SupplierRepository) Create(ctx context.Context, supplier interface{}) (string, error) {
	s, ok := supplier.(*Supplier)
	if !ok {
		return "", fmt.Errorf("invalid supplier type")
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO suppliers (id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		id, s.Code, s.Name, s.Email, s.Phone, s.Address, s.City, s.Country, s.CreditLimit, s.CurrentBalance, s.Notes, now, now)

	if err != nil {
		return "", fmt.Errorf("insert supplier: %w", err)
	}

	return id, nil
}

// GetByID retrieves a supplier by ID
func (r *SupplierRepository) GetByID(ctx context.Context, id string) (interface{}, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM suppliers
		WHERE id = ? AND is_active = 1
	`

	var supplier Supplier
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&supplier.ID, &supplier.Code, &supplier.Name, &supplier.Email, &supplier.Phone, &supplier.Address, &supplier.City, &supplier.Country,
		&supplier.CreditLimit, &supplier.CurrentBalance, &supplier.Notes, &supplier.IsActive, &supplier.CreatedAt, &supplier.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query supplier: %w", err)
	}

	return &supplier, nil
}

// GetByCode retrieves a supplier by code
func (r *SupplierRepository) GetByCode(ctx context.Context, code string) (interface{}, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM suppliers
		WHERE code = ? AND is_active = 1
	`

	var supplier Supplier
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&supplier.ID, &supplier.Code, &supplier.Name, &supplier.Email, &supplier.Phone, &supplier.Address, &supplier.City, &supplier.Country,
		&supplier.CreditLimit, &supplier.CurrentBalance, &supplier.Notes, &supplier.IsActive, &supplier.CreatedAt, &supplier.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query supplier by code: %w", err)
	}

	return &supplier, nil
}

// List retrieves all active suppliers
func (r *SupplierRepository) List(ctx context.Context, limit, offset int) ([]interface{}, error) {
	query := `
		SELECT id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM suppliers
		WHERE is_active = 1
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []interface{}
	for rows.Next() {
		var supplier Supplier
		if err := rows.Scan(&supplier.ID, &supplier.Code, &supplier.Name, &supplier.Email, &supplier.Phone, &supplier.Address, &supplier.City, &supplier.Country,
			&supplier.CreditLimit, &supplier.CurrentBalance, &supplier.Notes, &supplier.IsActive, &supplier.CreatedAt, &supplier.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan supplier: %w", err)
		}
		suppliers = append(suppliers, &supplier)
	}

	return suppliers, rows.Err()
}

// Update modifies an existing supplier
func (r *SupplierRepository) Update(ctx context.Context, id string, supplier interface{}) error {
	s, ok := supplier.(*Supplier)
	if !ok {
		return fmt.Errorf("invalid supplier type")
	}

	query := `
		UPDATE suppliers
		SET code = ?, name = ?, email = ?, phone = ?, address = ?, city = ?, country = ?, credit_limit = ?, current_balance = ?, notes = ?, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.ExecContext(ctx, query,
		s.Code, s.Name, s.Email, s.Phone, s.Address, s.City, s.Country, s.CreditLimit, s.CurrentBalance, s.Notes, time.Now(), id)

	if err != nil {
		return fmt.Errorf("update supplier: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("supplier not found")
	}

	return nil
}

// Delete soft-deletes a supplier
func (r *SupplierRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE suppliers
		SET is_active = 0, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("delete supplier: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("supplier not found")
	}

	return nil
}

// Search searches suppliers by name or code
func (r *SupplierRepository) Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error) {
	sql := `
		SELECT id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at
		FROM suppliers
		WHERE is_active = 1 AND (name LIKE ? OR code LIKE ? OR phone LIKE ? OR email LIKE ?)
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sql, searchTerm, searchTerm, searchTerm, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []interface{}
	for rows.Next() {
		var supplier Supplier
		if err := rows.Scan(&supplier.ID, &supplier.Code, &supplier.Name, &supplier.Email, &supplier.Phone, &supplier.Address, &supplier.City, &supplier.Country,
			&supplier.CreditLimit, &supplier.CurrentBalance, &supplier.Notes, &supplier.IsActive, &supplier.CreatedAt, &supplier.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan supplier: %w", err)
		}
		suppliers = append(suppliers, &supplier)
	}

	return suppliers, rows.Err()
}

// Count returns total number of active suppliers
func (r *SupplierRepository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM suppliers WHERE is_active = 1"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count suppliers: %w", err)
	}
	return count, nil
}
