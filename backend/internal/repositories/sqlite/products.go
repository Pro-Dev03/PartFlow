package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Product represents a product entity
type Product struct {
	ID            string    `db:"id"`
	SKU           string    `db:"sku"`
	Name          string    `db:"name"`
	Description   string    `db:"description"`
	CategoryID    *string   `db:"category_id"`
	PurchasePrice float64   `db:"purchase_price"`
	SellingPrice  float64   `db:"selling_price"`
	Currency      string    `db:"currency"`
	MinStockLevel int       `db:"min_stock_level"`
	IsActive      int       `db:"is_active"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// ProductRepository implements repositories.ProductRepository for SQLite
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new SQLite product repository
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create inserts a new product
func (r *ProductRepository) Create(ctx context.Context, product interface{}) (string, error) {
	p, ok := product.(*Product)
	if !ok {
		return "", fmt.Errorf("invalid product type")
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO products (id, sku, name, description, category_id, purchase_price, selling_price, currency, min_stock_level, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		id, p.SKU, p.Name, p.Description, p.CategoryID, p.PurchasePrice, p.SellingPrice, p.Currency, p.MinStockLevel, now, now)

	if err != nil {
		return "", fmt.Errorf("insert product: %w", err)
	}

	return id, nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id string) (interface{}, error) {
	query := `
		SELECT id, sku, name, description, category_id, purchase_price, selling_price, currency, min_stock_level, is_active, created_at, updated_at
		FROM products
		WHERE id = ? AND is_active = 1
	`

	var product Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.SKU, &product.Name, &product.Description, &product.CategoryID,
		&product.PurchasePrice, &product.SellingPrice, &product.Currency, &product.MinStockLevel,
		&product.IsActive, &product.CreatedAt, &product.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}

	return &product, nil
}

// GetBySKU retrieves a product by SKU
func (r *ProductRepository) GetBySKU(ctx context.Context, sku string) (interface{}, error) {
	query := `
		SELECT id, sku, name, description, category_id, purchase_price, selling_price, currency, min_stock_level, is_active, created_at, updated_at
		FROM products
		WHERE sku = ? AND is_active = 1
	`

	var product Product
	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&product.ID, &product.SKU, &product.Name, &product.Description, &product.CategoryID,
		&product.PurchasePrice, &product.SellingPrice, &product.Currency, &product.MinStockLevel,
		&product.IsActive, &product.CreatedAt, &product.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query product by sku: %w", err)
	}

	return &product, nil
}

// List retrieves all active products
func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]interface{}, error) {
	query := `
		SELECT id, sku, name, description, category_id, purchase_price, selling_price, currency, min_stock_level, is_active, created_at, updated_at
		FROM products
		WHERE is_active = 1
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []interface{}
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Description, &product.CategoryID,
			&product.PurchasePrice, &product.SellingPrice, &product.Currency, &product.MinStockLevel,
			&product.IsActive, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, rows.Err()
}

// Update modifies an existing product
func (r *ProductRepository) Update(ctx context.Context, id string, product interface{}) error {
	p, ok := product.(*Product)
	if !ok {
		return fmt.Errorf("invalid product type")
	}

	query := `
		UPDATE products
		SET sku = ?, name = ?, description = ?, category_id = ?, purchase_price = ?, selling_price = ?, currency = ?, min_stock_level = ?, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.ExecContext(ctx, query,
		p.SKU, p.Name, p.Description, p.CategoryID, p.PurchasePrice, p.SellingPrice, p.Currency, p.MinStockLevel, time.Now(), id)

	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Delete soft-deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE products
		SET is_active = 0, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Search searches products by name or SKU
func (r *ProductRepository) Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error) {
	sql := `
		SELECT id, sku, name, description, category_id, purchase_price, selling_price, currency, min_stock_level, is_active, created_at, updated_at
		FROM products
		WHERE is_active = 1 AND (name LIKE ? OR sku LIKE ?)
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sql, searchTerm, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	var products []interface{}
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Description, &product.CategoryID,
			&product.PurchasePrice, &product.SellingPrice, &product.Currency, &product.MinStockLevel,
			&product.IsActive, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, rows.Err()
}

// Count returns total number of active products
func (r *ProductRepository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM products WHERE is_active = 1"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count products: %w", err)
	}
	return count, nil
}
