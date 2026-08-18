package warranty

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles warranty data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new warranty repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateWarrantyClaim creates a new warranty claim
func (r *Repository) CreateWarrantyClaim(ctx context.Context, claim *WarrantyClaim) error {
	query := `
		INSERT INTO warranty_claims (id, organization_id, claim_number, product_id, serial_number, 
			customer_id, sale_id, claim_date, claim_type, reason, description, status, priority,
			estimated_cost, actual_cost, resolution, resolved_at, assigned_to, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		claim.ID, claim.OrganizationID, claim.ClaimNumber, claim.ProductID, claim.SerialNumber,
		claim.CustomerID, claim.SaleID, claim.ClaimDate, claim.ClaimType, claim.Reason, claim.Description,
		claim.Status, claim.Priority, claim.EstimatedCost, claim.ActualCost, claim.Resolution,
		claim.ResolvedAt, claim.AssignedTo, claim.Notes, claim.CreatedBy, claim.CreatedAt, claim.UpdatedAt,
	).Scan(&claim.ID, &claim.CreatedAt, &claim.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create warranty claim: %w", err)
	}
	return nil
}

// GetWarrantyClaimByID retrieves a warranty claim by ID
func (r *Repository) GetWarrantyClaimByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*WarrantyClaim, error) {
	var claim WarrantyClaim
	query := `
		SELECT id, organization_id, claim_number, product_id, serial_number, customer_id, sale_id,
			claim_date, claim_type, reason, description, status, priority, estimated_cost, actual_cost,
			resolution, resolved_at, assigned_to, notes, created_by, created_at, updated_at
		FROM warranty_claims
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &claim, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWarrantyClaimNotFound
		}
		return nil, fmt.Errorf("failed to get warranty claim: %w", err)
	}
	return &claim, nil
}

// ListWarrantyClaims retrieves warranty claims with pagination and filters
func (r *Repository) ListWarrantyClaims(ctx context.Context, organizationID uuid.UUID, req WarrantyClaimListRequest) ([]WarrantyClaim, int, error) {
	var claims []WarrantyClaim
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, claim_number, product_id, serial_number, customer_id, sale_id,
			claim_date, claim_type, reason, description, status, priority, estimated_cost, actual_cost,
			resolution, resolved_at, assigned_to, notes, created_by, created_at, updated_at
		FROM warranty_claims
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM warranty_claims WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if req.ProductID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND product_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND product_id = $%d", argCount)
		args = append(args, *req.ProductID)
	}
	
	if req.CustomerID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, *req.CustomerID)
	}
	
	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}
	
	if req.Priority != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND priority = $%d", argCount)
		countQuery += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, req.Priority)
	}
	
	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND claim_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND claim_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	
	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND claim_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND claim_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (claim_number ILIKE $%d OR serial_number ILIKE $%d OR reason ILIKE $%d)", argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (claim_number ILIKE $%d OR serial_number ILIKE $%d OR reason ILIKE $%d)", argCount, argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count warranty claims: %w", err)
	}
	
	// Add sorting
	sortBy := "claim_date"
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
	
	err = r.db.SelectContext(ctx, &claims, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list warranty claims: %w", err)
	}
	
	return claims, count, nil
}

// UpdateWarrantyClaim updates a warranty claim
func (r *Repository) UpdateWarrantyClaim(ctx context.Context, claim *WarrantyClaim) error {
	query := `
		UPDATE warranty_claims
		SET status = $2, priority = $3, estimated_cost = $4, actual_cost = $5, 
			resolution = $6, resolved_at = $7, assigned_to = $8, notes = $9, updated_at = $10
		WHERE id = $1 AND organization_id = $11
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		claim.ID, claim.Status, claim.Priority, claim.EstimatedCost, claim.ActualCost,
		claim.Resolution, claim.ResolvedAt, claim.AssignedTo, claim.Notes, claim.UpdatedAt, claim.OrganizationID,
	).Scan(&claim.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrWarrantyClaimNotFound
		}
		return fmt.Errorf("failed to update warranty claim: %w", err)
	}
	return nil
}

// CreateWarrantyClaimItem creates a new warranty claim item
func (r *Repository) CreateWarrantyClaimItem(ctx context.Context, item *WarrantyClaimItem) error {
	query := `
		INSERT INTO warranty_claim_items (id, claim_id, product_id, serial_number, quantity, 
			condition, defect_type, defect_description, is_repaired, repaired_at, repair_cost, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.ClaimID, item.ProductID, item.SerialNumber, item.Quantity,
		item.Condition, item.DefectType, item.DefectDescription, item.IsRepaired, item.RepairedAt,
		item.RepairCost, item.Notes, item.CreatedAt, item.UpdatedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create warranty claim item: %w", err)
	}
	return nil
}

// GetWarrantyClaimItems retrieves items for a warranty claim
func (r *Repository) GetWarrantyClaimItems(ctx context.Context, claimID uuid.UUID) ([]WarrantyClaimItem, error) {
	var items []WarrantyClaimItem
	query := `
		SELECT id, claim_id, product_id, serial_number, quantity, condition, defect_type, 
			defect_description, is_repaired, repaired_at, repair_cost, notes, created_at, updated_at
		FROM warranty_claim_items
		WHERE claim_id = $1
		ORDER BY created_at
	`
	
	err := r.db.SelectContext(ctx, &items, query, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claim items: %w", err)
	}
	return items, nil
}

// CreateWarranty creates a new warranty
func (r *Repository) CreateWarranty(ctx context.Context, warranty *Warranty) error {
	query := `
		INSERT INTO warranties (id, organization_id, product_id, warranty_type, duration_days, 
			start_date, end_date, terms, coverage, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		warranty.ID, warranty.OrganizationID, warranty.ProductID, warranty.WarrantyType,
		warranty.DurationDays, warranty.StartDate, warranty.EndDate, warranty.Terms,
		warranty.Coverage, warranty.IsActive, warranty.CreatedAt, warranty.UpdatedAt,
	).Scan(&warranty.ID, &warranty.CreatedAt, &warranty.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create warranty: %w", err)
	}
	return nil
}

// GetWarrantyByID retrieves a warranty by ID
func (r *Repository) GetWarrantyByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Warranty, error) {
	var warranty Warranty
	query := `
		SELECT id, organization_id, product_id, warranty_type, duration_days, start_date, end_date,
			terms, coverage, is_active, created_at, updated_at
		FROM warranties
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &warranty, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWarrantyNotFound
		}
		return nil, fmt.Errorf("failed to get warranty: %w", err)
	}
	return &warranty, nil
}

// GetActiveWarranty retrieves the active warranty for a product
func (r *Repository) GetActiveWarranty(ctx context.Context, productID uuid.UUID) (*Warranty, error) {
	var warranty Warranty
	query := `
		SELECT id, organization_id, product_id, warranty_type, duration_days, start_date, end_date,
			terms, coverage, is_active, created_at, updated_at
		FROM warranties
		WHERE product_id = $1 AND is_active = true AND end_date > NOW()
		ORDER BY end_date DESC
		LIMIT 1
	`
	
	err := r.db.GetContext(ctx, &warranty, query, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWarrantyNotFound
		}
		return nil, fmt.Errorf("failed to get active warranty: %w", err)
	}
	return &warranty, nil
}

// GetProductInfo retrieves product information
func (r *Repository) GetProductInfo(ctx context.Context, productID uuid.UUID) (*ProductInfo, error) {
	var product ProductInfo
	query := `SELECT id, name, model, sku FROM products WHERE id = $1`
	
	err := r.db.GetContext(ctx, &product, query, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product info: %w", err)
	}
	return &product, nil
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

// GetUserInfo retrieves user information
func (r *Repository) GetUserInfo(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	var user UserInfo
	query := `SELECT id, first_name, last_name, email FROM users WHERE id = $1`
	
	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	return &user, nil
}

// GetWarrantyClaimsSummary retrieves warranty claims summary statistics
func (r *Repository) GetWarrantyClaimsSummary(ctx context.Context, organizationID uuid.UUID) (*WarrantyClaimsSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_claims,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_claims,
			COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_claims,
			COUNT(CASE WHEN status = 'rejected' THEN 1 END) as rejected_claims,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_claims,
			COALESCE(SUM(actual_cost), 0) as total_cost,
			COALESCE(AVG(EXTRACT(DAY FROM (resolved_at - claim_date))), 0) as average_resolution_time_days
		FROM warranty_claims
		WHERE organization_id = $1
	`
	var summary WarrantyClaimsSummary
	err := r.db.GetContext(ctx, &summary, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claims summary: %w", err)
	}
	return &summary, nil
}