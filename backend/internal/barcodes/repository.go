package barcodes

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetBarcodeByCode retrieves a barcode by its code
func (r *Repository) GetBarcodeByCode(ctx context.Context, code string) (*Barcode, error) {
	query := `
		SELECT id, code, product_id, type, generated_at, created_at, updated_at
		FROM barcodes
		WHERE code = $1
	`

	var barcode Barcode
	err := r.db.GetContext(ctx, &barcode, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode: %w", err)
	}

	return &barcode, nil
}

// GetProductByBarcode retrieves product information by barcode code
func (r *Repository) GetProductByBarcode(ctx context.Context, code string) (*ProductInfo, error) {
	query := `
		SELECT id, name, sku, barcode, selling_price, cost_price, stock, condition, category
		FROM products
		WHERE barcode = $1 OR sku = $1
		LIMIT 1
	`

	var product ProductInfo
	err := r.db.GetContext(ctx, &product, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by barcode: %w", err)
	}

	return &product, nil
}

// GetProductBySKU retrieves product information by SKU
func (r *Repository) GetProductBySKU(ctx context.Context, sku string) (*ProductInfo, error) {
	query := `
		SELECT id, name, sku, barcode, selling_price, cost_price, stock, condition, category
		FROM products
		WHERE sku = $1
		LIMIT 1
	`

	var product ProductInfo
	err := r.db.GetContext(ctx, &product, query, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	return &product, nil
}

// GetBarcodeByID retrieves a barcode by its ID
func (r *Repository) GetBarcodeByID(ctx context.Context, id uuid.UUID) (*Barcode, error) {
	query := `
		SELECT id, code, product_id, type, generated_at, created_at, updated_at
		FROM barcodes
		WHERE id = $1
	`

	var barcode Barcode
	err := r.db.GetContext(ctx, &barcode, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode: %w", err)
	}

	return &barcode, nil
}

// CreateBarcode creates a new barcode
func (r *Repository) CreateBarcode(ctx context.Context, barcode *Barcode) error {
	query := `
		INSERT INTO barcodes (id, code, product_id, inventory_item_id, type, is_active, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		barcode.ID, barcode.Code, barcode.ProductID, barcode.InventoryItemID, barcode.Type, barcode.IsActive,
		barcode.GeneratedAt, barcode.CreatedAt, barcode.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create barcode: %w", err)
	}

	return nil
}

// ListBarcodes lists barcodes
func (r *Repository) ListBarcodes(ctx context.Context, limit, offset int) ([]*Barcode, int64, error) {
	query := `
		SELECT id, code, product_id, inventory_item_id, type, is_active, generated_at, created_at, updated_at
		FROM barcodes
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var barcodes []*Barcode
	err := r.db.SelectContext(ctx, &barcodes, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list barcodes: %w", err)
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM barcodes`
	err = r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count barcodes: %w", err)
	}

	return barcodes, total, nil
}

// DeleteBarcode soft deletes a barcode
func (r *Repository) DeleteBarcode(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE barcodes
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete barcode: %w", err)
	}

	return nil
}