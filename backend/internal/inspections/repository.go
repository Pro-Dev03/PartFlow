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

func mustMarshalTestResults(results TestResults) []byte {
	data, err := json.Marshal(results)
	if err != nil {
		return []byte(`{}`)
	}
	return data
}

// LinkAcquisitionItemInspection associates an inspection with its acquisition item.
func (r *Repository) LinkAcquisitionItemInspection(ctx context.Context, itemID, inspectionID uuid.UUID, inventoryItemID *uuid.UUID, status string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE acquisition_items
		SET inspection_id = $1, inventory_item_id = COALESCE($2, inventory_item_id),
		    inspection_status = $3,
		    item_status = CASE
		        WHEN $3 = 'passed' THEN 'available'
		        WHEN $3 = 'failed' THEN 'rejected'
		        ELSE 'inspection'
		    END,
		    updated_at = NOW()
		WHERE id = $4
	`, inspectionID, inventoryItemID, status, itemID)
	if err != nil {
		return fmt.Errorf("failed to link acquisition item inspection: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("acquisition item not found")
	}
	return nil
}

// UpdateAcquisitionItemInspectionStatus keeps acquisition workflow state in sync.
func (r *Repository) UpdateAcquisitionItemInspectionStatus(ctx context.Context, itemID, inspectionID uuid.UUID, status string) error {
	itemStatus := "inspection"
	if status == "passed" {
		itemStatus = "available"
	} else if status == "failed" {
		itemStatus = "rejected"
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE acquisition_items
		SET inspection_id = $1,
		    inventory_item_id = COALESCE(inventory_item_id,
		        (SELECT inventory_item_id FROM inspections WHERE id = $1)),
		    inspection_status = $2, item_status = $3, updated_at = NOW()
		WHERE id = $4
	`, inspectionID, status, itemStatus, itemID)
	if err != nil {
		return fmt.Errorf("failed to update acquisition item inspection status: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("acquisition item not found")
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE inventory_items
		SET status = CASE WHEN $2 = 'passed' THEN 'AVAILABLE' ELSE 'DAMAGED' END,
		    updated_at = NOW()
		WHERE id = (
		    SELECT COALESCE(i.inventory_item_id, ai.inventory_item_id)
		    FROM inspections i
		    LEFT JOIN acquisition_items ai ON ai.id = i.acquisition_item_id
		    WHERE i.id = $1
		)
	`, inspectionID, status)
	if err != nil {
		return fmt.Errorf("failed to update inventory item status: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE acquisitions
		SET status = CASE
			WHEN EXISTS (
				SELECT 1 FROM acquisition_items
				WHERE acquisition_id = acquisitions.id AND inspection_status = 'failed'
			) THEN 'rejected'
			WHEN NOT EXISTS (
				SELECT 1 FROM acquisition_items
				WHERE acquisition_id = acquisitions.id
				  AND inspection_status IN ('pending', 'needs_repair')
			) THEN 'approved'
			ELSE 'inspection'
		END,
		updated_at = NOW()
		WHERE id = (SELECT acquisition_id FROM acquisition_items WHERE id = $1)
	`, itemID)
	if err != nil {
		return fmt.Errorf("failed to update acquisition status: %w", err)
	}
	return nil
}

// NewRepository creates a new inspection repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateInspection creates a new inspection
func (r *Repository) CreateInspection(ctx context.Context, inspection *Inspection) error {
	photosJSON, err := json.Marshal(inspection.Photos)
	if err != nil {
		return fmt.Errorf("failed to marshal photos: %w", err)
	}

	query := `
		INSERT INTO inspections
			(product_id, inventory_item_id, inspector_id, inspection_date, result, condition, grade,
			 notes, images, test_results, acquisition_item_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		inspection.ProductID, inspection.InventoryItemID, inspection.InspectedBy, inspection.InspectionDate,
		inspection.Status, inspection.Condition, inspection.Grade, inspection.Notes, photosJSON,
		mustMarshalTestResults(inspection.TestResults), inspection.AcquisitionItemID, inspection.CreatedAt,
	).Scan(&inspection.ID, &inspection.CreatedAt, &inspection.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inspection: %w", err)
	}
	return nil
}

// GetInspectionByID retrieves an inspection by ID
func (r *Repository) GetInspectionByID(ctx context.Context, id uuid.UUID) (*Inspection, error) {
	var inspection Inspection
	var testResultsJSON, photosJSON []byte
	var condition, grade, notes sql.NullString

	query := `
		SELECT i.id, i.product_id, i.inventory_item_id, i.inspection_date,
		       i.inspector_id AS inspected_by,
		       COALESCE(ai.serial_number, ii.serial_number, '') AS serial_number,
		       i.result AS status, i.condition, i.grade, i.notes, i.images,
		       i.test_results, i.acquisition_item_id, i.created_at, i.updated_at
		FROM inspections i
		LEFT JOIN acquisition_items ai ON ai.id = i.acquisition_item_id
		LEFT JOIN inventory_items ii ON ii.id = i.inventory_item_id
		WHERE i.id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inspection.ID, &inspection.ProductID, &inspection.InventoryItemID,
		&inspection.InspectionDate, &inspection.InspectedBy, &inspection.SerialNumber,
		&inspection.Status,
		&condition, &grade, &notes, &photosJSON, &testResultsJSON,
		&inspection.AcquisitionItemID, &inspection.CreatedAt, &inspection.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInspectionNotFound
		}
		return nil, fmt.Errorf("failed to get inspection: %w", err)
	}

	inspection.Condition = condition.String
	inspection.Grade = grade.String
	inspection.Notes = notes.String
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
func (r *Repository) ListInspections(ctx context.Context, req InspectionListRequest) ([]Inspection, int, error) {
	var inspections []Inspection
	var count int

	// Build base query
	baseQuery := `
		SELECT i.id, i.product_id, i.inventory_item_id, i.inspection_date,
		       i.inspector_id AS inspected_by,
		       COALESCE(ai.serial_number, ii.serial_number, '') AS serial_number,
		       i.result AS status, i.condition, i.grade, i.notes, i.images,
		       i.test_results, i.acquisition_item_id, i.created_at, i.updated_at
		FROM inspections i
		LEFT JOIN acquisition_items ai ON ai.id = i.acquisition_item_id
		LEFT JOIN inventory_items ii ON ii.id = i.inventory_item_id
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*) FROM inspections i
		LEFT JOIN acquisition_items ai ON ai.id = i.acquisition_item_id
		LEFT JOIN inventory_items ii ON ii.id = i.inventory_item_id
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if req.ProductID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.product_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND i.product_id = $%d", argCount)
		args = append(args, *req.ProductID)
	}

	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.result = $%d", argCount)
		countQuery += fmt.Sprintf(" AND i.result = $%d", argCount)
		args = append(args, req.Status)
	}

	if req.InspectedBy != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.inspector_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND i.inspector_id = $%d", argCount)
		args = append(args, *req.InspectedBy)
	}

	if req.Condition != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.condition = $%d", argCount)
		countQuery += fmt.Sprintf(" AND i.condition = $%d", argCount)
		args = append(args, req.Condition)
	}

	if req.Grade != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.grade = $%d", argCount)
		countQuery += fmt.Sprintf(" AND i.grade = $%d", argCount)
		args = append(args, req.Grade)
	}

	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.inspection_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND i.inspection_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}

	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND i.inspection_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND i.inspection_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}

	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (COALESCE(ai.serial_number, ii.serial_number, '') ILIKE $%d OR i.notes ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (COALESCE(ai.serial_number, ii.serial_number, '') ILIKE $%d OR i.notes ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern)
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
		var condition, grade, notes sql.NullString

		err := rows.Scan(
			&inspection.ID, &inspection.ProductID, &inspection.InventoryItemID,
			&inspection.InspectionDate, &inspection.InspectedBy, &inspection.SerialNumber,
			&inspection.Status,
			&condition, &grade, &notes, &photosJSON, &testResultsJSON,
			&inspection.AcquisitionItemID, &inspection.CreatedAt, &inspection.UpdatedAt,
		)
		if err != nil {
			continue
		}
		inspection.Condition = condition.String
		inspection.Grade = grade.String
		inspection.Notes = notes.String

		// Parse JSON fields
		if len(testResultsJSON) > 0 {
			json.Unmarshal(testResultsJSON, &inspection.TestResults)
		}
		if len(photosJSON) > 0 {
			json.Unmarshal(photosJSON, &inspection.Photos)
		}

		inspections = append(inspections, inspection)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate inspections: %w", err)
	}

	return inspections, count, nil
}

// UpdateInspection updates an inspection
func (r *Repository) UpdateInspection(ctx context.Context, inspection *Inspection) error {
	photosJSON, err := json.Marshal(inspection.Photos)
	if err != nil {
		return fmt.Errorf("failed to marshal photos: %w", err)
	}

	query := `
		UPDATE inspections
		SET inspection_date = $2, result = $3, condition = $4, grade = $5, notes = $6,
		    images = $7, test_results = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		inspection.ID, inspection.InspectionDate, inspection.Status, inspection.Condition,
		inspection.Grade, inspection.Notes, photosJSON, mustMarshalTestResults(inspection.TestResults),
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
func (r *Repository) DeleteInspection(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM inspections WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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
	query := `SELECT id, name, COALESCE(model, '') AS model, COALESCE(sku, '') AS sku,
		COALESCE(barcode, '') AS barcode FROM products WHERE id = $1`

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
	query := `SELECT id, COALESCE(first_name, '') AS first_name, COALESCE(last_name, '') AS last_name,
		COALESCE(email, '') AS email FROM users WHERE id = $1`

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
func (r *Repository) GetInspectionSummary(ctx context.Context) (*InspectionSummary, error) {
	var summary InspectionSummary

	// Total inspections
	err := r.db.GetContext(ctx, &summary.TotalInspections,
		`SELECT COUNT(*) FROM inspections`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total inspections: %w", err)
	}

	// Passed inspections
	err = r.db.GetContext(ctx, &summary.PassedInspections,
		`SELECT COUNT(*) FROM inspections WHERE result = 'passed'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get passed inspections: %w", err)
	}

	// Failed inspections
	err = r.db.GetContext(ctx, &summary.FailedInspections,
		`SELECT COUNT(*) FROM inspections WHERE result = 'failed'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed inspections: %w", err)
	}

	// Pending inspections
	err = r.db.GetContext(ctx, &summary.PendingInspections,
		`SELECT COUNT(*) FROM inspections WHERE result = 'pending'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending inspections: %w", err)
	}

	// This week inspections
	err = r.db.GetContext(ctx, &summary.ThisWeek,
		`SELECT COUNT(*) FROM inspections
		 WHERE inspection_date >= DATE_TRUNC('week', CURRENT_DATE)`)
	if err != nil {
		return nil, fmt.Errorf("failed to get this week inspections: %w", err)
	}

	// This month inspections
	err = r.db.GetContext(ctx, &summary.ThisMonth,
		`SELECT COUNT(*) FROM inspections
		 WHERE DATE_TRUNC('month', inspection_date) = DATE_TRUNC('month', CURRENT_DATE)`)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month inspections: %w", err)
	}

	// By condition
	summary.ByCondition = make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		`SELECT condition, COUNT(*) FROM inspections GROUP BY condition`)
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
		`SELECT grade, COUNT(*) FROM inspections GROUP BY grade`)
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
