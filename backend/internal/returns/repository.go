package returns

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles return data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new return repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateReturn creates a new return
func (r *Repository) CreateReturn(ctx context.Context, returnRecord *Return) error {
	query := `
		INSERT INTO returns (id, organization_id, sale_id, customer_id, return_number, 
			return_date, reason, condition, status, refund_amount, refund_method, refund_date, 
			notes, processed_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		returnRecord.ID, returnRecord.OrganizationID, returnRecord.SaleID, returnRecord.CustomerID,
		returnRecord.ReturnNumber, returnRecord.ReturnDate, returnRecord.Reason, returnRecord.Condition,
		returnRecord.Status, returnRecord.RefundAmount, returnRecord.RefundMethod, returnRecord.RefundDate,
		returnRecord.Notes, returnRecord.ProcessedBy, returnRecord.CreatedAt, returnRecord.UpdatedAt,
	).Scan(&returnRecord.ID, &returnRecord.CreatedAt, &returnRecord.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create return: %w", err)
	}
	return nil
}

// GetReturnByID retrieves a return by ID
func (r *Repository) GetReturnByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Return, error) {
	var returnRecord Return
	query := `
		SELECT id, organization_id, sale_id, customer_id, return_number, return_date, 
			reason, condition, status, refund_amount, refund_method, refund_date, 
			notes, processed_by, created_at, updated_at
		FROM returns
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &returnRecord, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReturnNotFound
		}
		return nil, fmt.Errorf("failed to get return: %w", err)
	}
	return &returnRecord, nil
}

// ListReturns retrieves returns with pagination and filters
func (r *Repository) ListReturns(ctx context.Context, organizationID uuid.UUID, req ReturnListRequest) ([]Return, int, error) {
	var returns []Return
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, sale_id, customer_id, return_number, return_date, 
			reason, condition, status, refund_amount, refund_method, refund_date, 
			notes, processed_by, created_at, updated_at
		FROM returns
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM returns WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if req.CustomerID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, *req.CustomerID)
	}
	
	if req.SaleID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND sale_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND sale_id = $%d", argCount)
		args = append(args, *req.SaleID)
	}
	
	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}
	
	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND return_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND return_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	
	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND return_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND return_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (return_number ILIKE $%d OR reason ILIKE $%d OR notes ILIKE $%d)", argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (return_number ILIKE $%d OR reason ILIKE $%d OR notes ILIKE $%d)", argCount, argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count returns: %w", err)
	}
	
	// Add sorting
	sortBy := "return_date"
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
	
	err = r.db.SelectContext(ctx, &returns, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list returns: %w", err)
	}
	
	return returns, count, nil
}

// UpdateReturn updates a return
func (r *Repository) UpdateReturn(ctx context.Context, returnRecord *Return) error {
	query := `
		UPDATE returns
		SET return_date = $2, reason = $3, condition = $4, status = $5, 
			refund_amount = $6, refund_method = $7, refund_date = $8, notes = $9, updated_at = $10
		WHERE id = $1 AND organization_id = $11
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		returnRecord.ID, returnRecord.ReturnDate, returnRecord.Reason, returnRecord.Condition,
		returnRecord.Status, returnRecord.RefundAmount, returnRecord.RefundMethod, returnRecord.RefundDate,
		returnRecord.Notes, returnRecord.UpdatedAt, returnRecord.OrganizationID,
	).Scan(&returnRecord.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReturnNotFound
		}
		return fmt.Errorf("failed to update return: %w", err)
	}
	return nil
}

// DeleteReturn deletes a return
func (r *Repository) DeleteReturn(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM returns WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete return: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrReturnNotFound
	}
	
	return nil
}

// CreateReturnItem creates a new return item
func (r *Repository) CreateReturnItem(ctx context.Context, item *ReturnItem) error {
	query := `
		INSERT INTO return_items (id, return_id, sale_item_id, product_id, quantity, 
			unit_price, total_price, reason, condition, is_resellable, location_id, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.ReturnID, item.SaleItemID, item.ProductID, item.Quantity,
		item.UnitPrice, item.TotalPrice, item.Reason, item.Condition, item.IsResellable,
		item.LocationID, item.Notes, item.CreatedAt, item.UpdatedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create return item: %w", err)
	}
	return nil
}

// GetReturnItems retrieves items for a return
func (r *Repository) GetReturnItems(ctx context.Context, returnID uuid.UUID) ([]ReturnItem, error) {
	var items []ReturnItem
	query := `
		SELECT id, return_id, sale_item_id, product_id, quantity, unit_price, total_price, 
			reason, condition, is_resellable, location_id, notes, created_at, updated_at
		FROM return_items
		WHERE return_id = $1
		ORDER BY created_at
	`
	
	err := r.db.SelectContext(ctx, &items, query, returnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get return items: %w", err)
	}
	return items, nil
}

// UpdateReturnItem updates a return item
func (r *Repository) UpdateReturnItem(ctx context.Context, item *ReturnItem) error {
	query := `
		UPDATE return_items
		SET quantity = $2, reason = $3, condition = $4, is_resellable = $5, 
			location_id = $6, notes = $7, updated_at = $8
		WHERE id = $1
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.Quantity, item.Reason, item.Condition, item.IsResellable,
		item.LocationID, item.Notes, item.UpdatedAt,
	).Scan(&item.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReturnItemNotFound
		}
		return fmt.Errorf("failed to update return item: %w", err)
	}
	return nil
}

// DeleteReturnItem deletes a return item
func (r *Repository) DeleteReturnItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM return_items WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete return item: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrReturnItemNotFound
	}
	
	return nil
}

// GetCustomerInfo retrieves customer information
func (r *Repository) GetCustomerInfo(ctx context.Context, customerID uuid.UUID) (*CustomerInfo, error) {
	var customer CustomerInfo
	query := `SELECT id, name, phone, email FROM customers WHERE id = $1`
	
	err := r.db.GetContext(ctx, &customer, query, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer info: %w", err)
	}
	return &customer, nil
}

// GetSaleInfo retrieves sale information
func (r *Repository) GetSaleInfo(ctx context.Context, saleID uuid.UUID) (*SaleInfo, error) {
	var sale SaleInfo
	query := `SELECT id, invoice_number, sale_date, total_amount FROM sales WHERE id = $1`
	
	err := r.db.GetContext(ctx, &sale, query, saleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSaleNotFound
		}
		return nil, fmt.Errorf("failed to get sale info: %w", err)
	}
	return &sale, nil
}

// GetSaleItemInfo retrieves sale item information
func (r *Repository) GetSaleItemInfo(ctx context.Context, saleItemID uuid.UUID) (struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	UnitPrice float64
}, error) {
	var item struct {
		ID        uuid.UUID
		ProductID uuid.UUID
		Quantity  int
		UnitPrice float64
	}
	
	query := `SELECT id, product_id, quantity, unit_price FROM sale_items WHERE id = $1`
	
	err := r.db.GetContext(ctx, &item, query, saleItemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return item, ErrSaleItemNotFound
		}
		return item, fmt.Errorf("failed to get sale item info: %w", err)
	}
	return item, nil
}

// GetReturnByReturnNumber retrieves a return by return number
func (r *Repository) GetReturnByReturnNumber(ctx context.Context, returnNumber string, organizationID uuid.UUID) (*Return, error) {
	var returnRecord Return
	query := `
		SELECT id, organization_id, sale_id, customer_id, return_number, return_date, 
			reason, condition, status, refund_amount, refund_method, refund_date, 
			notes, processed_by, created_at, updated_at
		FROM returns
		WHERE return_number = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &returnRecord, query, returnNumber, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReturnNotFound
		}
		return nil, fmt.Errorf("failed to get return by return number: %w", err)
	}
	return &returnRecord, nil
}
