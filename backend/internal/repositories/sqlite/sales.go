package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Sale represents a sale entity
type Sale struct {
	ID              string    `db:"id"`
	SaleNumber      string    `db:"sale_number"`
	CustomerID      *string   `db:"customer_id"`
	TotalAmount     float64   `db:"total_amount"`
	TaxAmount       float64   `db:"tax_amount"`
	DiscountAmount  float64   `db:"discount_amount"`
	PaidAmount      float64   `db:"paid_amount"`
	RemainingAmount float64   `db:"remaining_amount"`
	PaymentMethod   *string   `db:"payment_method"`
	Status          string    `db:"status"`
	Notes           *string   `db:"notes"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// SaleItem represents a sale item entity
type SaleItem struct {
	ID              string    `db:"id"`
	SaleID          string    `db:"sale_id"`
	InventoryItemID *string   `db:"inventory_item_id"`
	ProductID       string    `db:"product_id"`
	Quantity        int       `db:"quantity"`
	UnitPrice       float64   `db:"unit_price"`
	ItemTotal       float64   `db:"item_total"`
	CreatedAt       time.Time `db:"created_at"`
}

// SalesRepository implements repositories.SalesRepository for SQLite
type SalesRepository struct {
	db *sql.DB
}

// NewSalesRepository creates a new SQLite sales repository
func NewSalesRepository(db *sql.DB) *SalesRepository {
	return &SalesRepository{db: db}
}

// Create inserts a new sale
func (r *SalesRepository) Create(ctx context.Context, sale interface{}) (string, error) {
	s, ok := sale.(*Sale)
	if !ok {
		return "", fmt.Errorf("invalid sale type")
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO sales (id, sale_number, customer_id, total_amount, tax_amount, discount_amount, paid_amount, remaining_amount, payment_method, status, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		id, s.SaleNumber, s.CustomerID, s.TotalAmount, s.TaxAmount, s.DiscountAmount, s.PaidAmount, s.RemainingAmount, s.PaymentMethod, s.Status, s.Notes, now, now)

	if err != nil {
		return "", fmt.Errorf("insert sale: %w", err)
	}

	return id, nil
}

// GetByID retrieves a sale by ID
func (r *SalesRepository) GetByID(ctx context.Context, id string) (interface{}, error) {
	query := `
		SELECT id, sale_number, customer_id, total_amount, tax_amount, discount_amount, paid_amount, remaining_amount, payment_method, status, notes, created_at, updated_at
		FROM sales
		WHERE id = ?
	`

	var sale Sale
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sale.ID, &sale.SaleNumber, &sale.CustomerID, &sale.TotalAmount, &sale.TaxAmount, &sale.DiscountAmount,
		&sale.PaidAmount, &sale.RemainingAmount, &sale.PaymentMethod, &sale.Status, &sale.Notes, &sale.CreatedAt, &sale.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query sale: %w", err)
	}

	return &sale, nil
}

// List retrieves all sales
func (r *SalesRepository) List(ctx context.Context, limit, offset int) ([]interface{}, error) {
	query := `
		SELECT id, sale_number, customer_id, total_amount, tax_amount, discount_amount, paid_amount, remaining_amount, payment_method, status, notes, created_at, updated_at
		FROM sales
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query sales: %w", err)
	}
	defer rows.Close()

	var sales []interface{}
	for rows.Next() {
		var sale Sale
		if err := rows.Scan(&sale.ID, &sale.SaleNumber, &sale.CustomerID, &sale.TotalAmount, &sale.TaxAmount, &sale.DiscountAmount,
			&sale.PaidAmount, &sale.RemainingAmount, &sale.PaymentMethod, &sale.Status, &sale.Notes, &sale.CreatedAt, &sale.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan sale: %w", err)
		}
		sales = append(sales, &sale)
	}

	return sales, rows.Err()
}

// ListByCustomer retrieves sales for a specific customer
func (r *SalesRepository) ListByCustomer(ctx context.Context, customerID string, limit, offset int) ([]interface{}, error) {
	query := `
		SELECT id, sale_number, customer_id, total_amount, tax_amount, discount_amount, paid_amount, remaining_amount, payment_method, status, notes, created_at, updated_at
		FROM sales
		WHERE customer_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query sales by customer: %w", err)
	}
	defer rows.Close()

	var sales []interface{}
	for rows.Next() {
		var sale Sale
		if err := rows.Scan(&sale.ID, &sale.SaleNumber, &sale.CustomerID, &sale.TotalAmount, &sale.TaxAmount, &sale.DiscountAmount,
			&sale.PaidAmount, &sale.RemainingAmount, &sale.PaymentMethod, &sale.Status, &sale.Notes, &sale.CreatedAt, &sale.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan sale: %w", err)
		}
		sales = append(sales, &sale)
	}

	return sales, rows.Err()
}

// Update modifies an existing sale
func (r *SalesRepository) Update(ctx context.Context, id string, sale interface{}) error {
	s, ok := sale.(*Sale)
	if !ok {
		return fmt.Errorf("invalid sale type")
	}

	query := `
		UPDATE sales
		SET sale_number = ?, customer_id = ?, total_amount = ?, tax_amount = ?, discount_amount = ?, paid_amount = ?, remaining_amount = ?, payment_method = ?, status = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		s.SaleNumber, s.CustomerID, s.TotalAmount, s.TaxAmount, s.DiscountAmount, s.PaidAmount, s.RemainingAmount, s.PaymentMethod, s.Status, s.Notes, time.Now(), id)

	if err != nil {
		return fmt.Errorf("update sale: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("sale not found")
	}

	return nil
}

// Delete removes a sale
func (r *SalesRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sales WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete sale: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("sale not found")
	}

	return nil
}

// Count returns total number of sales
func (r *SalesRepository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM sales"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count sales: %w", err)
	}
	return count, nil
}

// GetTotal returns total sales amount
func (r *SalesRepository) GetTotal(ctx context.Context) (float64, error) {
	query := "SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE status = 'completed'"
	var total float64
	err := r.db.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get sales total: %w", err)
	}
	return total, nil
}
