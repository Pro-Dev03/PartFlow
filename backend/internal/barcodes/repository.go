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
func (r *Repository) GetBarcodeByCode(ctx context.Context, code string, organizationID uuid.UUID) (*Barcode, error) {
	query := `
		SELECT id, organization_id, code, type, product_id, inventory_item_id, is_active, created_at, updated_at
		FROM barcodes
		WHERE code = $1 AND organization_id = $2 AND is_active = true
	`

	var barcode Barcode
	err := r.db.GetContext(ctx, &barcode, query, code, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode: %w", err)
	}

	return &barcode, nil
}

// GetBarcodeByID retrieves a barcode by its ID
func (r *Repository) GetBarcodeByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Barcode, error) {
	query := `
		SELECT id, organization_id, code, type, product_id, inventory_item_id, is_active, created_at, updated_at
		FROM barcodes
		WHERE id = $1 AND organization_id = $2 AND is_active = true
	`

	var barcode Barcode
	err := r.db.GetContext(ctx, &barcode, query, id, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode: %w", err)
	}

	return &barcode, nil
}

// CreateBarcode creates a new barcode
func (r *Repository) CreateBarcode(ctx context.Context, barcode *Barcode) error {
	query := `
		INSERT INTO barcodes (id, organization_id, code, type, product_id, inventory_item_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		barcode.ID, barcode.OrganizationID, barcode.Code, barcode.Type,
		barcode.ProductID, barcode.InventoryItemID, barcode.IsActive,
		barcode.CreatedAt, barcode.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create barcode: %w", err)
	}

	return nil
}

// ListBarcodes lists barcodes for an organization
func (r *Repository) ListBarcodes(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]*Barcode, int64, error) {
	query := `
		SELECT id, organization_id, code, type, product_id, inventory_item_id, is_active, created_at, updated_at
		FROM barcodes
		WHERE organization_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var barcodes []*Barcode
	err := r.db.SelectContext(ctx, &barcodes, query, organizationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list barcodes: %w", err)
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM barcodes WHERE organization_id = $1 AND is_active = true`
	err = r.db.GetContext(ctx, &total, countQuery, organizationID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count barcodes: %w", err)
	}

	return barcodes, total, nil
}

// DeleteBarcode soft deletes a barcode
func (r *Repository) DeleteBarcode(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `
		UPDATE barcodes
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete barcode: %w", err)
	}

	return nil
}