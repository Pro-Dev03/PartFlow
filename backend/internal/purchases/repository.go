package purchases

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles purchase data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new purchase repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new purchase
func (r *Repository) Create(ctx context.Context, purchase *Purchase) error {
	query := `
		INSERT INTO purchases (id, organization_id, supplier_id, invoice_number, purchase_date, 
			total_amount, paid_amount, status, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		purchase.ID, purchase.OrganizationID, purchase.SupplierID, purchase.InvoiceNumber,
		purchase.PurchaseDate, purchase.TotalAmount, purchase.PaidAmount, purchase.Status,
		purchase.Notes, purchase.CreatedBy, purchase.CreatedAt, purchase.UpdatedAt,
	).Scan(&purchase.ID, &purchase.CreatedAt, &purchase.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create purchase: %w", err)
	}
	return nil
}

// GetByID retrieves a purchase by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Purchase, error) {
	var purchase Purchase
	query := `
		SELECT id, organization_id, supplier_id, invoice_number, purchase_date, 
			total_amount, paid_amount, status, notes, created_by, created_at, updated_at
		FROM purchases
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &purchase, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPurchaseNotFound
		}
		return nil, fmt.Errorf("failed to get purchase: %w", err)
	}
	return &purchase, nil
}

// List retrieves purchases with pagination and filters
func (r *Repository) List(ctx context.Context, organizationID uuid.UUID, req PurchaseListRequest) ([]Purchase, int, error) {
	var purchases []Purchase
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, supplier_id, invoice_number, purchase_date, 
			total_amount, paid_amount, status, notes, created_by, created_at, updated_at
		FROM purchases
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM purchases WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if req.SupplierID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND supplier_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND supplier_id = $%d", argCount)
		args = append(args, *req.SupplierID)
	}
	
	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}
	
	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND purchase_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND purchase_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	
	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND purchase_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND purchase_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (invoice_number ILIKE $%d OR notes ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (invoice_number ILIKE $%d OR notes ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count purchases: %w", err)
	}
	
	// Add sorting
	sortBy := "purchase_date"
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
	
	err = r.db.SelectContext(ctx, &purchases, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list purchases: %w", err)
	}
	
	return purchases, count, nil
}

// Update updates a purchase
func (r *Repository) Update(ctx context.Context, purchase *Purchase) error {
	query := `
		UPDATE purchases
		SET invoice_number = $2, purchase_date = $3, status = $4, notes = $5, updated_at = $6
		WHERE id = $1 AND organization_id = $7
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		purchase.ID, purchase.InvoiceNumber, purchase.PurchaseDate, purchase.Status,
		purchase.Notes, purchase.UpdatedAt, purchase.OrganizationID,
	).Scan(&purchase.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPurchaseNotFound
		}
		return fmt.Errorf("failed to update purchase: %w", err)
	}
	return nil
}

// Delete deletes a purchase
func (r *Repository) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM purchases WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete purchase: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrPurchaseNotFound
	}
	
	return nil
}

// CreatePurchaseItem creates a new purchase item
func (r *Repository) CreatePurchaseItem(ctx context.Context, item *PurchaseItem) error {
	query := `
		INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_cost, 
			total_cost, serial_number, condition, location_id, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.PurchaseID, item.ProductID, item.Quantity, item.UnitCost,
		item.TotalCost, item.SerialNumber, item.Condition, item.LocationID,
		item.Notes, item.CreatedAt, item.UpdatedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create purchase item: %w", err)
	}
	return nil
}

// GetPurchaseItems retrieves items for a purchase
func (r *Repository) GetPurchaseItems(ctx context.Context, purchaseID uuid.UUID) ([]PurchaseItem, error) {
	var items []PurchaseItem
	query := `
		SELECT id, purchase_id, product_id, quantity, unit_cost, total_cost, 
			serial_number, condition, location_id, notes, created_at, updated_at
		FROM purchase_items
		WHERE purchase_id = $1
		ORDER BY created_at
	`
	
	err := r.db.SelectContext(ctx, &items, query, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase items: %w", err)
	}
	return items, nil
}

// UpdatePurchaseItem updates a purchase item
func (r *Repository) UpdatePurchaseItem(ctx context.Context, item *PurchaseItem) error {
	query := `
		UPDATE purchase_items
		SET quantity = $2, unit_cost = $3, total_cost = $4, serial_number = $5, 
			condition = $6, location_id = $7, notes = $8, updated_at = $9
		WHERE id = $1
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.Quantity, item.UnitCost, item.TotalCost, item.SerialNumber,
		item.Condition, item.LocationID, item.Notes, item.UpdatedAt,
	).Scan(&item.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPurchaseItemNotFound
		}
		return fmt.Errorf("failed to update purchase item: %w", err)
	}
	return nil
}

// DeletePurchaseItem deletes a purchase item
func (r *Repository) DeletePurchaseItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM purchase_items WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete purchase item: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrPurchaseItemNotFound
	}
	
	return nil
}

// GetSupplierInfo retrieves supplier information
func (r *Repository) GetSupplierInfo(ctx context.Context, supplierID uuid.UUID) (*SupplierInfo, error) {
	var supplier SupplierInfo
	query := `SELECT id, name, phone FROM suppliers WHERE id = $1`
	
	err := r.db.GetContext(ctx, &supplier, query, supplierID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier info: %w", err)
	}
	return &supplier, nil
}

// UpdatePaidAmount updates the paid amount for a purchase
func (r *Repository) UpdatePaidAmount(ctx context.Context, purchaseID uuid.UUID, amount float64) error {
	query := `
		UPDATE purchases
		SET paid_amount = paid_amount + $2, updated_at = $3
		WHERE id = $1
		RETURNING updated_at
	`
	
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, purchaseID, amount, time.Now()).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update paid amount: %w", err)
	}
	return nil
}

// GetPurchaseByInvoiceNumber retrieves a purchase by invoice number
func (r *Repository) GetPurchaseByInvoiceNumber(ctx context.Context, invoiceNumber string, organizationID uuid.UUID) (*Purchase, error) {
	var purchase Purchase
	query := `
		SELECT id, organization_id, supplier_id, invoice_number, purchase_date, 
			total_amount, paid_amount, status, notes, created_by, created_at, updated_at
		FROM purchases
		WHERE invoice_number = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &purchase, query, invoiceNumber, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPurchaseNotFound
		}
		return nil, fmt.Errorf("failed to get purchase by invoice number: %w", err)
	}
	return &purchase, nil
}
