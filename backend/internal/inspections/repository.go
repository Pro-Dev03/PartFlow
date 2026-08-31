package inspections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
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
	result, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE acquisition_items
		SET inspection_id = $1, inventory_item_id = COALESCE($2, inventory_item_id),
		    inspection_status = $3,
		    item_status = CASE
		        WHEN $3 = 'passed' THEN 'available'
		        WHEN $3 = 'failed' THEN 'rejected'
		        ELSE 'inspection'
		    END,
		    updated_at = %s
		WHERE id = $4
	`, dbutil.NowSQL(r.db)), inspectionID, inventoryItemID, status, itemID)
	if err != nil {
		return fmt.Errorf("failed to link acquisition item inspection: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("acquisition item not found")
	}
	return nil
}

func updateAcquisitionItemInspectionStatusTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, itemID, inspectionID uuid.UUID, inventoryItemID *uuid.UUID, status string) error {
	itemStatus := "inspection"
	if status == "passed" {
		itemStatus = "available"
	} else if status == "failed" {
		itemStatus = "rejected"
	}
	nowSQL := dbutil.NowSQL(db)
	result, err := tx.ExecContext(ctx, fmt.Sprintf(`UPDATE acquisition_items SET inspection_id = $1, inventory_item_id = COALESCE($2, inventory_item_id), inspection_status = $3, item_status = $4, updated_at = %s WHERE id = $5`, nowSQL), inspectionID, inventoryItemID, status, itemStatus, itemID)
	if err != nil {
		return fmt.Errorf("failed to link acquisition item inspection: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("acquisition item not found")
	}
	if inventoryItemID != nil {
		inventoryStatus := "INSPECTION"
		if status == "passed" {
			inventoryStatus = "AVAILABLE"
		} else if status == "failed" {
			inventoryStatus = "DAMAGED"
		}
		if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE inventory_items SET status = $1, updated_at = %s WHERE id = $2`, nowSQL), inventoryStatus, *inventoryItemID); err != nil {
			return fmt.Errorf("failed to update inventory item status: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE acquisitions SET status = CASE WHEN EXISTS (SELECT 1 FROM acquisition_items WHERE acquisition_id = acquisitions.id AND inspection_status = 'failed') THEN 'rejected' WHEN NOT EXISTS (SELECT 1 FROM acquisition_items WHERE acquisition_id = acquisitions.id AND inspection_status IN ('pending', 'needs_repair')) THEN 'approved' ELSE 'inspection' END, updated_at = %s WHERE id = (SELECT acquisition_id FROM acquisition_items WHERE id = $1)`, nowSQL), itemID); err != nil {
		return fmt.Errorf("failed to update acquisition status: %w", err)
	}
	return nil
}

// UpdateAcquisitionItemInspectionStatus keeps acquisition workflow state in sync.
func (r *Repository) UpdateAcquisitionItemInspectionStatus(ctx context.Context, itemID, inspectionID uuid.UUID, status string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin inspection workflow transaction: %w", err)
	}
	if err := r.updateAcquisitionItemInspectionStatus(ctx, tx, itemID, inspectionID, status); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inspection workflow transaction: %w", err)
	}
	return nil
}

func (r *Repository) updateAcquisitionItemInspectionStatus(ctx context.Context, exec sqlx.ExtContext, itemID, inspectionID uuid.UUID, status string) error {
	itemStatus := "inspection"
	if status == "passed" {
		itemStatus = "available"
	} else if status == "failed" {
		itemStatus = "rejected"
	}
	nowSQL := dbutil.NowSQL(r.db)
	result, err := exec.ExecContext(ctx, fmt.Sprintf(`
		UPDATE acquisition_items
		SET inspection_id = $1,
		    inventory_item_id = COALESCE(inventory_item_id,
		        (SELECT inventory_item_id FROM inspections WHERE id = $1)),
		    inspection_status = $2, item_status = $3, updated_at = %s
		WHERE id = $4
	`, nowSQL), inspectionID, status, itemStatus, itemID)
	if err != nil {
		return fmt.Errorf("failed to update acquisition item inspection status: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("acquisition item not found")
	}
	_, err = exec.ExecContext(ctx, fmt.Sprintf(`
		UPDATE inventory_items
		SET status = CASE WHEN $2 = 'passed' THEN 'AVAILABLE' ELSE 'DAMAGED' END,
		    updated_at = %s
		WHERE id = (
		    SELECT COALESCE(i.inventory_item_id, ai.inventory_item_id)
		    FROM inspections i
		    LEFT JOIN acquisition_items ai ON ai.id = i.acquisition_item_id
		    WHERE i.id = $1
		)
	`, nowSQL), inspectionID, status)
	if err != nil {
		return fmt.Errorf("failed to update inventory item status: %w", err)
	}
	_, err = exec.ExecContext(ctx, fmt.Sprintf(`
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
		updated_at = %s
		WHERE id = (SELECT acquisition_id FROM acquisition_items WHERE id = $1)
	`, nowSQL), itemID)
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
			(id, product_id, inventory_item_id, inspector_id, inspection_date, result, condition, grade,
			 notes, images, test_results, acquisition_item_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
	`

	_, err = r.db.ExecContext(ctx, query,
		inspection.ID, inspection.ProductID, inspection.InventoryItemID, inspection.InspectedBy, inspection.InspectionDate,
		inspection.Status, inspection.Condition, inspection.Grade, inspection.Notes, photosJSON,
		mustMarshalTestResults(inspection.TestResults), inspection.AcquisitionItemID, inspection.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create inspection: %w", err)
	}
	return nil
}

// CreateInspectionWithWorkflow persists an inspection and links its
// acquisition item atomically. It is used by the API so a failed link cannot
// leave an orphan inspection record.
func (r *Repository) CreateInspectionWithWorkflow(ctx context.Context, inspection *Inspection) error {
	photosJSON, err := json.Marshal(inspection.Photos)
	if err != nil {
		return fmt.Errorf("failed to marshal photos: %w", err)
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin inspection creation transaction: %w", err)
	}
	query := `INSERT INTO inspections (id, product_id, inventory_item_id, inspector_id, inspection_date, result, condition, grade, notes, images, test_results, acquisition_item_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)`
	if _, err = tx.ExecContext(ctx, query, inspection.ID, inspection.ProductID, inspection.InventoryItemID, inspection.InspectedBy, inspection.InspectionDate, inspection.Status, inspection.Condition, inspection.Grade, inspection.Notes, photosJSON, mustMarshalTestResults(inspection.TestResults), inspection.AcquisitionItemID, inspection.CreatedAt); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to create inspection: %w", err)
	}
	if inspection.AcquisitionItemID != nil {
		if err := updateAcquisitionItemInspectionStatusTx(ctx, tx, r.db, *inspection.AcquisitionItemID, inspection.ID, inspection.InventoryItemID, inspection.Status); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inspection creation transaction: %w", err)
	}
	return nil
}

// GetInspectionByID retrieves an inspection by ID
func (r *Repository) GetInspectionByID(ctx context.Context, id uuid.UUID) (*Inspection, error) {
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
	record := map[string]any{}
	err := r.db.QueryRowxContext(ctx, query, id).MapScan(record)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInspectionNotFound
		}
		return nil, fmt.Errorf("failed to get inspection: %w", err)
	}
	inspection, err := inspectionFromMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inspection: %w", err)
	}
	return inspection, nil
}

func inspectionFromMap(record map[string]any) (*Inspection, error) {
	inspection := &Inspection{
		ID:                parseInspectionUUID(record["id"]),
		ProductID:         parseInspectionNullableUUID(record["product_id"]),
		InventoryItemID:   parseInspectionNullableUUID(record["inventory_item_id"]),
		AcquisitionItemID: parseInspectionNullableUUID(record["acquisition_item_id"]),
		SerialNumber:      strings.TrimSpace(fmt.Sprint(record["serial_number"])),
		Status:            strings.TrimSpace(fmt.Sprint(record["status"])),
		Condition:         strings.TrimSpace(fmt.Sprint(record["condition"])),
		Grade:             strings.TrimSpace(fmt.Sprint(record["grade"])),
		Notes:             strings.TrimSpace(fmt.Sprint(record["notes"])),
	}
	inspection.InspectedBy = parseInspectionUUID(record["inspected_by"])
	for field, destination := range map[string]*time.Time{"inspection_date": &inspection.InspectionDate, "created_at": &inspection.CreatedAt, "updated_at": &inspection.UpdatedAt} {
		if value, err := dbutil.ParseTimestamp(record[field]); err == nil {
			*destination = value
		}
	}
	if raw := inspectionJSONBytes(record["test_results"]); len(raw) > 0 {
		if err := json.Unmarshal(raw, &inspection.TestResults); err != nil {
			return nil, fmt.Errorf("unmarshal test results: %w", err)
		}
	}
	if raw := inspectionJSONBytes(record["images"]); len(raw) > 0 {
		if err := json.Unmarshal(raw, &inspection.Photos); err != nil {
			return nil, fmt.Errorf("unmarshal images: %w", err)
		}
	}
	return inspection, nil
}

func parseInspectionUUID(value any) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	if parsed, ok := value.(uuid.UUID); ok {
		return parsed
	}
	if bytes, ok := value.([]byte); ok {
		if len(bytes) == 16 {
			var parsed uuid.UUID
			copy(parsed[:], bytes)
			return parsed
		}
		value = string(bytes)
	}
	parsed, _ := uuid.Parse(strings.TrimSpace(fmt.Sprint(value)))
	return parsed
}

func parseInspectionNullableUUID(value any) *uuid.UUID {
	parsed := parseInspectionUUID(value)
	if parsed == uuid.Nil {
		return nil
	}
	return &parsed
}

func inspectionJSONBytes(value any) []byte {
	switch raw := value.(type) {
	case []byte:
		return raw
	case string:
		return []byte(raw)
	default:
		return nil
	}
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
		baseQuery += fmt.Sprintf(" AND (LOWER(COALESCE(ai.serial_number, ii.serial_number, '')) LIKE LOWER($%d) OR LOWER(i.notes) LIKE LOWER($%d))", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (LOWER(COALESCE(ai.serial_number, ii.serial_number, '')) LIKE LOWER($%d) OR LOWER(i.notes) LIKE LOWER($%d))", argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern)
	}

	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inspections: %w", err)
	}

	// Add sorting from a fixed allow-list; sort_by is user input.
	sortColumns := map[string]string{
		"inspection_date": "i.inspection_date",
		"created_at":      "i.created_at",
		"status":          "i.result",
		"condition":       "i.condition",
		"grade":           "i.grade",
	}
	sortBy := sortColumns[req.SortBy]
	if sortBy == "" {
		sortBy = "i.inspection_date"
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.SortOrder, "asc") {
		sortOrder = "ASC"
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)

	rows, err := r.db.QueryxContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list inspections: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, 0, fmt.Errorf("failed to scan inspection: %w", err)
		}
		inspection, err := inspectionFromMap(record)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse inspection: %w", err)
		}
		inspections = append(inspections, *inspection)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate inspections: %w", err)
	}

	return inspections, count, nil
}

// UpdateInspection updates an inspection
func (r *Repository) UpdateInspection(ctx context.Context, inspection *Inspection) error {
	if err := r.updateInspection(ctx, r.db, inspection); err != nil {
		return err
	}
	return nil
}

// UpdateInspectionWithWorkflow updates the inspection and all acquisition/
// inventory state in one transaction so a partial inspection can never leave
// the item and its acquisition out of sync.
func (r *Repository) UpdateInspectionWithWorkflow(ctx context.Context, inspection *Inspection) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin inspection update transaction: %w", err)
	}
	if err := r.updateInspection(ctx, tx, inspection); err != nil {
		_ = tx.Rollback()
		return err
	}
	if inspection.AcquisitionItemID != nil {
		if err := r.updateAcquisitionItemInspectionStatus(ctx, tx, *inspection.AcquisitionItemID, inspection.ID, inspection.Status); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inspection update transaction: %w", err)
	}
	return nil
}

func (r *Repository) updateInspection(ctx context.Context, exec sqlx.ExtContext, inspection *Inspection) error {
	photosJSON, err := json.Marshal(inspection.Photos)
	if err != nil {
		return fmt.Errorf("failed to marshal photos: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE inspections
		SET inspection_date = $2, result = $3, condition = $4, grade = $5, notes = $6,
		    images = $7, test_results = $8, updated_at = %s
		WHERE id = $1
	`, dbutil.NowSQL(r.db))
	result, err := exec.ExecContext(ctx, query,
		inspection.ID, inspection.InspectionDate, inspection.Status, inspection.Condition,
		inspection.Grade, inspection.Notes, photosJSON, mustMarshalTestResults(inspection.TestResults),
	)

	if err != nil {
		return fmt.Errorf("failed to update inspection: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrInspectionNotFound
	}
	inspection.UpdatedAt = time.Now().UTC()
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
		`SELECT COUNT(*) FROM inspections WHERE LOWER(COALESCE(result, '')) = 'passed'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get passed inspections: %w", err)
	}

	// Failed inspections
	err = r.db.GetContext(ctx, &summary.FailedInspections,
		`SELECT COUNT(*) FROM inspections WHERE LOWER(COALESCE(result, '')) = 'failed'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed inspections: %w", err)
	}

	// Pending inspections
	err = r.db.GetContext(ctx, &summary.PendingInspections,
		`SELECT COUNT(*) FROM inspections WHERE LOWER(COALESCE(result, '')) = 'pending'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending inspections: %w", err)
	}

	// This week inspections
	weekQuery := `SELECT COUNT(*) FROM inspections WHERE inspection_date >= DATE_TRUNC('week', CURRENT_DATE)`
	monthQuery := `SELECT COUNT(*) FROM inspections WHERE DATE_TRUNC('month', inspection_date) = DATE_TRUNC('month', CURRENT_DATE)`
	if dbutil.IsSQLite(r.db) {
		// SQLite stores local timestamps as TEXT.  Using a half-open date range
		// keeps the query index-friendly and avoids PostgreSQL-only DATE_TRUNC.
		weekQuery = `SELECT COUNT(*) FROM inspections WHERE date(inspection_date) >= date('now', 'weekday 1', '-7 days')`
		monthQuery = `SELECT COUNT(*) FROM inspections WHERE strftime('%Y-%m', inspection_date) = strftime('%Y-%m', 'now')`
	}
	err = r.db.GetContext(ctx, &summary.ThisWeek, weekQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get this week inspections: %w", err)
	}

	// This month inspections
	err = r.db.GetContext(ctx, &summary.ThisMonth, monthQuery)
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
