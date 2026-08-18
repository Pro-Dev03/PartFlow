package inspections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles inspection data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new inspection repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateInspection creates a new inspection
func (r *Repository) CreateInspection(ctx context.Context, inspection *Inspection) error {
	testResultsJSON, err := json.Marshal(inspection.TestResults)
	if err != nil {
		return fmt.Errorf("failed to marshal test results: %w", err)
	}

	photosJSON, err := json.Marshal(inspection.Photos)
	if err != nil {
		return fmt.Errorf("failed to marshal photos: %w", err)
	}

	query := `
		INSERT INTO inspections (id, organization_id, product_id, serial_number, inspection_date, 
			inspected_by, status, condition, grade, notes, photos, test_results, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	
	err = r.db.QueryRowContext(ctx, query,
		inspection.ID, inspection.OrganizationID, inspection.ProductID, inspection.SerialNumber,
		inspection.InspectionDate, inspection.InspectedBy, inspection.Status, inspection.Condition,
		inspection.Grade, inspection.Notes, photosJSON, testResultsJSON,
		inspection.CreatedAt, inspection.UpdatedAt,
	).Scan(&inspection.ID, &inspection.CreatedAt, &inspection.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create inspection: %w", err)
	}
	return nil
}

// GetInspectionByID retrieves an inspection by ID
func (r *Repository) GetInspectionByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Inspection, error) {
	var inspection Inspection
	var testResultsJSON, photosJSON []byte
	
	query := `
		SELECT id, organization_id, product_id, serial_number, inspection_date, 
			inspected_by, status, condition, grade, notes, photos, test_results, created_at, updated_at
		FROM inspections
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &inspection, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInspectionNotFound
		}
		return nil, fmt.Errorf("failed to get inspection: %w", err)
	}
	
	// Parse JSON fields
	if len(testResultsJSON) > 0 {
		if err := json.Unmarshal(testResultsJSON, &inspection.TestResults); err != nil {
			return nil, fmt.Errorf("failed to unmarshal test results: %w", err)
		}
	}
	
	if len(photosJSON) > 0 {
		if err := json.Unmarshal(photosJSON, &inspection.Photos); err != nil {
			return nil, fmt.Errorf("failed to unmarshal photos: %w", err)
		}
	}
	
	return &inspection, nil
}

// ListInspections retrieves inspections with pagination and filters
func (r *Repository) ListInspections(ctx context.Context, organizationID uuid.UUID, req InspectionListRequest) ([]Inspection, int, error) {
	var inspections []Inspection
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, product_id, serial_number, inspection_date, 
			inspected_by, status, condition, grade, notes, photos, test_results, created_at, updated_at
		FROM inspections
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM inspections WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if req.ProductID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND product_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND product_id = $%d", argCount)
		args = append(args, *req.ProductID)
	}
	
	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}
	
	if req.Condition != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND condition = $%d", argCount)
		countQuery += fmt.Sprintf(" AND condition = $%d", argCount)
		args = append(args, req.Condition)
	}
	
	if req.Grade != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND grade = $%d", argCount)
		countQuery += fmt.Sprintf(" AND grade = $%d", argCount)
		args = append(args, req.Grade)
	}
	
	if req.InspectedBy != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND inspected_by = $%d", argCount)
		countQuery += fmt.Sprintf(" AND inspected_by = $%d", argCount)
		args = append(args, *req.InspectedBy)
	}
	
	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND inspection_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND inspection_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	
	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND inspection_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND inspection_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (serial_number ILIKE $%d OR notes ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (serial_number ILIKE $%d OR notes ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inspections: %w", err)
	}
	
	// Add sorting
	sortBy := "inspection_date"
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
	
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list inspections: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var inspection Inspection
		var testResultsJSON, photosJSON []byte
		
		err := rows.Scan(
			&inspection.ID, &inspection.OrganizationID, &inspection.ProductID, &inspection.SerialNumber,
			&inspection.InspectionDate, &inspection.InspectedBy, &inspection.Status, &inspection.Condition,
			&inspection.Grade, &inspection.Notes, &photosJSON, &testResultsJSON,
			&inspection.CreatedAt, &inspection.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		// Parse JSON fields
		if len(testResultsJSON) > 0 {
			json.Unmarshal(testResultsJSON, &inspection.TestResults)
		}
		if len(photosJSON) > 0 {
			json.Unmarshal(photosJSON, &inspection.Photos)
		}
		
		inspections = append(inspections, inspection)
	}
	
	return inspections, count, nil
}

// UpdateInspection updates an inspection
func (r *Repository) UpdateInspection(ctx context.Context, inspection *Inspection) error {
	testResultsJSON, err := json.Marshal(inspection.TestResults)
	if err != nil {
		return fmt.Errorf("failed to marshal test results: %w", err)
	}

	photosJSON, err := json.Marshal(inspection.Photos)
	if err != nil {
		return fmt.Errorf("failed to marshal photos: %w", err)
	}

	query := `
		UPDATE inspections
		SET inspection_date = $2, status = $3, condition = $4, grade = $5, notes = $6,
			photos = $7, test_results = $8, updated_at = $9
		WHERE id = $1 AND organization_id = $10
		RETURNING updated_at
	`
	
	err = r.db.QueryRowContext(ctx, query,
		inspection.ID, inspection.InspectionDate, inspection.Status, inspection.Condition,
		inspection.Grade, inspection.Notes, photosJSON, testResultsJSON,
		inspection.UpdatedAt, inspection.OrganizationID,
	).Scan(&inspection.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInspectionNotFound
		}
		return fmt.Errorf("failed to update inspection: %w", err)
	}
	return nil
}

// DeleteInspection deletes an inspection
func (r *Repository) DeleteInspection(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM inspections WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete inspection: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrInspectionNotFound
	}
	
	return nil
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

// GetUserInfo retrieves user information
func (r *Repository) GetUserInfo(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	var user UserInfo
	query := `SELECT id, first_name, last_name, email FROM users WHERE id = $1`
	
	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInspectionNotFound
		}
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	return &user, nil
}

// GetInspectionSummary retrieves inspection summary statistics
func (r *Repository) GetInspectionSummary(ctx context.Context, organizationID uuid.UUID) (*InspectionSummary, error) {
	var summary InspectionSummary
	
	// Total inspections
	err := r.db.GetContext(ctx, &summary.TotalInspections, 
		`SELECT COUNT(*) FROM inspections WHERE organization_id = $1`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total inspections: %w", err)
	}
	
	// Passed inspections
	err = r.db.GetContext(ctx, &summary.PassedInspections, 
		`SELECT COUNT(*) FROM inspections WHERE organization_id = $1 AND status = 'passed'`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get passed inspections: %w", err)
	}
	
	// Failed inspections
	err = r.db.GetContext(ctx, &summary.FailedInspections, 
		`SELECT COUNT(*) FROM inspections WHERE organization_id = $1 AND status = 'failed'`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed inspections: %w", err)
	}
	
	// Pending inspections
	err = r.db.GetContext(ctx, &summary.PendingInspections, 
		`SELECT COUNT(*) FROM inspections WHERE organization_id = $1 AND status = 'pending'`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending inspections: %w", err)
	}
	
	// This week inspections
	err = r.db.GetContext(ctx, &summary.ThisWeek, 
		`SELECT COUNT(*) FROM inspections 
		 WHERE organization_id = $1 
		 AND inspection_date >= DATE_TRUNC('week', CURRENT_DATE)`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get this week inspections: %w", err)
	}
	
	// This month inspections
	err = r.db.GetContext(ctx, &summary.ThisMonth, 
		`SELECT COUNT(*) FROM inspections 
		 WHERE organization_id = $1 
		 AND DATE_TRUNC('month', inspection_date) = DATE_TRUNC('month', CURRENT_DATE)`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month inspections: %w", err)
	}
	
	// By condition
	summary.ByCondition = make(map[string]int)
	rows, err := r.db.QueryContext(ctx, 
		`SELECT condition, COUNT(*) FROM inspections WHERE organization_id = $1 GROUP BY condition`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspections by condition: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var condition string
		var count int
		if err := rows.Scan(&condition, &count); err != nil {
			continue
		}
		summary.ByCondition[condition] = count
	}
	
	// By grade
	summary.ByGrade = make(map[string]int)
	rows, err = r.db.QueryContext(ctx, 
		`SELECT grade, COUNT(*) FROM inspections WHERE organization_id = $1 GROUP BY grade`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspections by grade: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var grade string
		var count int
		if err := rows.Scan(&grade, &count); err != nil {
			continue
		}
		summary.ByGrade[grade] = count
	}
	
	return &summary, nil
}
