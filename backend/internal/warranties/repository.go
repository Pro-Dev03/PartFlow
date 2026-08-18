package warranties

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

// CreateWarranty creates a new warranty
func (r *Repository) CreateWarranty(ctx context.Context, warranty *Warranty) error {
	query := `
		INSERT INTO warranties (id, organization_id, product_id, serial_number, warranty_number, 
			warranty_type, warranty_period, start_date, end_date, status, customer_id, sale_id, 
			terms, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		warranty.ID, warranty.OrganizationID, warranty.ProductID, warranty.SerialNumber,
		warranty.WarrantyNumber, warranty.WarrantyType, warranty.WarrantyPeriod, warranty.StartDate,
		warranty.EndDate, warranty.Status, warranty.CustomerID, warranty.SaleID, warranty.Terms,
		warranty.Notes, warranty.CreatedBy, warranty.CreatedAt, warranty.UpdatedAt,
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
		SELECT id, organization_id, product_id, serial_number, warranty_number, 
			warranty_type, warranty_period, start_date, end_date, status, customer_id, sale_id, 
			terms, notes, created_by, created_at, updated_at
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

// ListWarranties retrieves warranties with pagination and filters
func (r *Repository) ListWarranties(ctx context.Context, organizationID uuid.UUID, req WarrantyListRequest) ([]Warranty, int, error) {
	var warranties []Warranty
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, product_id, serial_number, warranty_number, 
			warranty_type, warranty_period, start_date, end_date, status, customer_id, sale_id, 
			terms, notes, created_by, created_at, updated_at
		FROM warranties
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM warranties WHERE organization_id = $1`
	
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
	
	if req.WarrantyType != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND warranty_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND warranty_type = $%d", argCount)
		args = append(args, req.WarrantyType)
	}
	
	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND start_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND start_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	
	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND end_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND end_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (warranty_number ILIKE $%d OR serial_number ILIKE $%d OR notes ILIKE $%d)", argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (warranty_number ILIKE $%d OR serial_number ILIKE $%d OR notes ILIKE $%d)", argCount, argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count warranties: %w", err)
	}
	
	// Add sorting
	sortBy := "end_date"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "ASC"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	
	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)
	
	err = r.db.SelectContext(ctx, &warranties, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list warranties: %w", err)
	}
	
	return warranties, count, nil
}

// UpdateWarranty updates a warranty
func (r *Repository) UpdateWarranty(ctx context.Context, warranty *Warranty) error {
	query := `
		UPDATE warranties
		SET serial_number = $2, warranty_type = $3, warranty_period = $4, start_date = $5,
			end_date = $6, status = $7, customer_id = $8, terms = $9, notes = $10, updated_at = $11
		WHERE id = $1 AND organization_id = $12
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		warranty.ID, warranty.SerialNumber, warranty.WarrantyType, warranty.WarrantyPeriod,
		warranty.StartDate, warranty.EndDate, warranty.Status, warranty.CustomerID,
		warranty.Terms, warranty.Notes, warranty.UpdatedAt, warranty.OrganizationID,
	).Scan(&warranty.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrWarrantyNotFound
		}
		return fmt.Errorf("failed to update warranty: %w", err)
	}
	return nil
}

// DeleteWarranty deletes a warranty
func (r *Repository) DeleteWarranty(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM warranties WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete warranty: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrWarrantyNotFound
	}
	
	return nil
}

// CreateWarrantyClaim creates a new warranty claim
func (r *Repository) CreateWarrantyClaim(ctx context.Context, claim *WarrantyClaim) error {
	query := `
		INSERT INTO warranty_claims (id, organization_id, warranty_id, claim_number, claim_date, 
			customer_id, issue_description, status, resolution, resolution_date, approved_by, 
			approved_date, rejected_by, rejected_date, completed_by, completed_date, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		claim.ID, claim.OrganizationID, claim.WarrantyID, claim.ClaimNumber, claim.ClaimDate,
		claim.CustomerID, claim.IssueDescription, claim.Status, claim.Resolution, claim.ResolutionDate,
		claim.ApprovedBy, claim.ApprovedDate, claim.RejectedBy, claim.RejectedDate,
		claim.CompletedBy, claim.CompletedDate, claim.Notes, claim.CreatedAt, claim.UpdatedAt,
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
		SELECT id, organization_id, warranty_id, claim_number, claim_date, customer_id, 
			issue_description, status, resolution, resolution_date, approved_by, approved_date, 
			rejected_by, rejected_date, completed_by, completed_date, notes, created_at, updated_at
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
		SELECT id, organization_id, warranty_id, claim_number, claim_date, customer_id, 
			issue_description, status, resolution, resolution_date, approved_by, approved_date, 
			rejected_by, rejected_date, completed_by, completed_date, notes, created_at, updated_at
		FROM warranty_claims
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM warranty_claims WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if req.WarrantyID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND warranty_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND warranty_id = $%d", argCount)
		args = append(args, *req.WarrantyID)
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
		baseQuery += fmt.Sprintf(" AND (claim_number ILIKE $%d OR issue_description ILIKE $%d OR notes ILIKE $%d)", argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (claim_number ILIKE $%d OR issue_description ILIKE $%d OR notes ILIKE $%d)", argCount, argCount, argCount)
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
		SET issue_description = $2, status = $3, resolution = $4, resolution_date = $5,
			approved_by = $6, approved_date = $7, rejected_by = $8, rejected_date = $9,
			completed_by = $10, completed_date = $11, notes = $12, updated_at = $13
		WHERE id = $1 AND organization_id = $14
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		claim.ID, claim.IssueDescription, claim.Status, claim.Resolution, claim.ResolutionDate,
		claim.ApprovedBy, claim.ApprovedDate, claim.RejectedBy, claim.RejectedDate,
		claim.CompletedBy, claim.CompletedDate, claim.Notes, claim.UpdatedAt, claim.OrganizationID,
	).Scan(&claim.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrWarrantyClaimNotFound
		}
		return fmt.Errorf("failed to update warranty claim: %w", err)
	}
	return nil
}

// DeleteWarrantyClaim deletes a warranty claim
func (r *Repository) DeleteWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM warranty_claims WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete warranty claim: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrWarrantyClaimNotFound
	}
	
	return nil
}

// GetWarrantyClaimsByWarrantyID retrieves claims for a warranty
func (r *Repository) GetWarrantyClaimsByWarrantyID(ctx context.Context, warrantyID uuid.UUID) ([]WarrantyClaim, error) {
	var claims []WarrantyClaim
	query := `
		SELECT id, organization_id, warranty_id, claim_number, claim_date, customer_id, 
			issue_description, status, resolution, resolution_date, approved_by, approved_date, 
			rejected_by, rejected_date, completed_by, completed_date, notes, created_at, updated_at
		FROM warranty_claims
		WHERE warranty_id = $1
		ORDER BY claim_date DESC
	`
	
	err := r.db.SelectContext(ctx, &claims, query, warrantyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claims: %w", err)
	}
	return claims, nil
}

// GetProductInfo retrieves product information
func (r *Repository) GetProductInfo(ctx context.Context, productID uuid.UUID) (*ProductInfo, error) {
	var product ProductInfo
	query := `SELECT id, name, model, sku, barcode FROM products WHERE id = $1`
	
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

// GetWarrantiesExpiringSoon retrieves warranties that will expire soon
func (r *Repository) GetWarrantiesExpiringSoon(ctx context.Context, organizationID uuid.UUID, days int) ([]WarrantyExpiringSoon, error) {
	var warranties []WarrantyExpiringSoon
	query := `
		SELECT w.id, w.product_id, p.name as product_name, w.serial_number, c.name as customer_name, 
			w.end_date, EXTRACT(DAY FROM (w.end_date - CURRENT_DATE)) as days_remaining
		FROM warranties w
		LEFT JOIN products p ON w.product_id = p.id
		LEFT JOIN customers c ON w.customer_id = c.id
		WHERE w.organization_id = $1 
			AND w.status = 'active'
			AND w.end_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '1 day' * $2
		ORDER BY w.end_date ASC
	`
	
	err := r.db.SelectContext(ctx, &warranties, query, organizationID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring warranties: %w", err)
	}
	return warranties, nil
}

// UpdateWarrantyStatus updates warranty status
func (r *Repository) UpdateWarrantyStatus(ctx context.Context, warrantyID uuid.UUID, status string) error {
	query := `
		UPDATE warranties
		SET status = $2, updated_at = $3
		WHERE id = $1
		RETURNING updated_at
	`
	
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, warrantyID, status, time.Now()).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update warranty status: %w", err)
	}
	return nil
}
