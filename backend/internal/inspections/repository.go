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
	"github.com/partflow/smart-store/internal/accounting"
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
		if err := captureInspectionWorkflowSnapshotTx(ctx, tx, inspection.ID, *inspection.AcquisitionItemID, inspection.InventoryItemID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("capture inspection workflow before-state: %w", err)
		}
		if err := updateAcquisitionItemInspectionStatusTx(ctx, tx, r.db, *inspection.AcquisitionItemID, inspection.ID, inspection.InventoryItemID, inspection.Status); err != nil {
			_ = tx.Rollback()
			return err
		}
	} else if inspection.InventoryItemID != nil {
		if err := captureStandaloneInventoryInspectionSnapshotTx(ctx, tx, inspection.ID, *inspection.InventoryItemID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("capture standalone inventory inspection before-state: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inspection creation transaction: %w", err)
	}
	return nil
}

// captureInspectionWorkflowSnapshotTx records the state that the inspection
// workflow is about to replace. Keeping this at the first write makes a later
// hard delete able to reverse the inspection without guessing what inventory
// and acquisition state existed before it.
func captureInspectionWorkflowSnapshotTx(ctx context.Context, tx *sqlx.Tx, inspectionID, acquisitionItemID uuid.UUID, inspectionInventoryItemID *uuid.UUID) error {
	var inventoryItemID, inventoryStatus sql.NullString
	var priorInventoryItemID, priorInspectionID, priorInspectionStatus, priorItemStatus sql.NullString
	var acquisitionID, acquisitionStatus sql.NullString
	query := `SELECT ai.inventory_item_id, ii.status, ai.inventory_item_id, ai.inspection_id,
		ai.inspection_status, ai.item_status, ai.acquisition_id, a.status
		FROM acquisition_items ai
		LEFT JOIN acquisitions a ON a.id = ai.acquisition_id
		LEFT JOIN inventory_items ii ON ii.id = COALESCE($1, ai.inventory_item_id)
		WHERE ai.id = $2`
	if err := tx.QueryRowxContext(ctx, query, inspectionInventoryItemID, acquisitionItemID).Scan(
		&inventoryItemID, &inventoryStatus, &priorInventoryItemID, &priorInspectionID,
		&priorInspectionStatus, &priorItemStatus, &acquisitionID, &acquisitionStatus,
	); err != nil {
		return fmt.Errorf("read acquisition workflow before-state: %w", err)
	}
	if inspectionInventoryItemID != nil {
		inventoryItemID = sql.NullString{String: inspectionInventoryItemID.String(), Valid: true}
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO inspection_workflow_snapshots
		(inspection_id, inventory_item_id, inventory_status_before, acquisition_item_id,
		 acquisition_inventory_item_id_before, acquisition_inspection_id_before,
		 acquisition_inspection_status_before, acquisition_item_status_before,
		 acquisition_id, acquisition_status_before)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (inspection_id) DO NOTHING`, inspectionID, nullStringValue(inventoryItemID), nullStringValue(inventoryStatus), acquisitionItemID,
		nullStringValue(priorInventoryItemID), nullStringValue(priorInspectionID), nullStringValue(priorInspectionStatus), nullStringValue(priorItemStatus),
		nullStringValue(acquisitionID), nullStringValue(acquisitionStatus))
	return err
}

func captureStandaloneInventoryInspectionSnapshotTx(ctx context.Context, tx *sqlx.Tx, inspectionID, inventoryItemID uuid.UUID) error {
	var status string
	if err := tx.GetContext(ctx, &status, `SELECT status FROM inventory_items WHERE id = $1`, inventoryItemID); err != nil {
		return fmt.Errorf("read standalone inventory item before-state: %w", err)
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO inspection_workflow_snapshots
		(inspection_id, inventory_item_id, inventory_status_before)
		VALUES ($1, $2, $3) ON CONFLICT (inspection_id) DO NOTHING`, inspectionID, inventoryItemID, status)
	return err
}

func nullStringValue(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
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
	if inspection.AcquisitionItemID != nil {
		if err := captureInspectionWorkflowSnapshotTx(ctx, tx, inspection.ID, *inspection.AcquisitionItemID, inspection.InventoryItemID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("capture inspection workflow before-state: %w", err)
		}
	} else if inspection.InventoryItemID != nil {
		if err := captureStandaloneInventoryInspectionSnapshotTx(ctx, tx, inspection.ID, *inspection.InventoryItemID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("capture standalone inventory inspection before-state: %w", err)
		}
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
	return r.DeleteInspectionWithWorkflow(ctx, id)
}

type inspectionWorkflowSnapshot struct {
	InventoryItemID                   sql.NullString `db:"inventory_item_id"`
	InventoryStatusBefore             sql.NullString `db:"inventory_status_before"`
	AcquisitionItemID                 sql.NullString `db:"acquisition_item_id"`
	AcquisitionInventoryItemIDBefore  sql.NullString `db:"acquisition_inventory_item_id_before"`
	AcquisitionInspectionIDBefore     sql.NullString `db:"acquisition_inspection_id_before"`
	AcquisitionInspectionStatusBefore sql.NullString `db:"acquisition_inspection_status_before"`
	AcquisitionItemStatusBefore       sql.NullString `db:"acquisition_item_status_before"`
	AcquisitionID                     sql.NullString `db:"acquisition_id"`
	AcquisitionStatusBefore           sql.NullString `db:"acquisition_status_before"`
}

// DeleteInspectionWithWorkflow reverses the state written by an inspection
// and removes it in one transaction. Legacy rows without snapshots use the
// acquisition workflow's known pre-inspection defaults, and only restore an
// item while its current state still matches the state produced by this
// inspection; later sales or manual status changes are preserved.
func (r *Repository) DeleteInspectionWithWorkflow(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin inspection deletion transaction: %w", err)
	}
	defer tx.Rollback()

	var result string
	var inspectionInventoryItemID, acquisitionItemID sql.NullString
	if err := tx.QueryRowxContext(ctx, `SELECT result, inventory_item_id, acquisition_item_id FROM inspections WHERE id = $1`, id).Scan(&result, &inspectionInventoryItemID, &acquisitionItemID); err != nil {
		if err == sql.ErrNoRows {
			return ErrInspectionNotFound
		}
		return fmt.Errorf("load inspection for deletion: %w", err)
	}

	snapshot := inspectionWorkflowSnapshot{}
	hasSnapshot := true
	if err := tx.GetContext(ctx, &snapshot, `SELECT inventory_item_id, inventory_status_before, acquisition_item_id,
		acquisition_inventory_item_id_before, acquisition_inspection_id_before,
		acquisition_inspection_status_before, acquisition_item_status_before, acquisition_id, acquisition_status_before
		FROM inspection_workflow_snapshots WHERE inspection_id = $1`, id); err != nil {
		if err == sql.ErrNoRows {
			hasSnapshot = false
		} else {
			return fmt.Errorf("load inspection workflow snapshot: %w", err)
		}
	}

	if !hasSnapshot && acquisitionItemID.Valid {
		// Prior to snapshots, acquisition items were materialized as available
		// and acquisitions started in draft. Recover that documented baseline.
		snapshot.AcquisitionItemID = acquisitionItemID
		snapshot.AcquisitionInspectionIDBefore = sql.NullString{}
		snapshot.AcquisitionInspectionStatusBefore = sql.NullString{}
		snapshot.AcquisitionItemStatusBefore = sql.NullString{Valid: true, String: "available"}
		if err := tx.QueryRowxContext(ctx, `SELECT inventory_item_id, acquisition_id FROM acquisition_items WHERE id = $1`, acquisitionItemID.String).Scan(&snapshot.AcquisitionInventoryItemIDBefore, &snapshot.AcquisitionID); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("read legacy acquisition link before inspection deletion: %w", err)
		}
		if snapshot.AcquisitionID.Valid {
			snapshot.AcquisitionStatusBefore = sql.NullString{Valid: true, String: "draft"}
		}
		if inspectionInventoryItemID.Valid {
			snapshot.InventoryItemID = inspectionInventoryItemID
			snapshot.InventoryStatusBefore = sql.NullString{Valid: true, String: "AVAILABLE"}
		} else if snapshot.AcquisitionInventoryItemIDBefore.Valid {
			snapshot.InventoryItemID = snapshot.AcquisitionInventoryItemIDBefore
			snapshot.InventoryStatusBefore = sql.NullString{Valid: true, String: "AVAILABLE"}
		}
	}
	if !hasSnapshot && !acquisitionItemID.Valid && inspectionInventoryItemID.Valid {
		// Older standalone inspections did not capture snapshots. Their workflow
		// starts from AVAILABLE; restoration is still guarded below by the
		// inspection-produced status so later item changes are preserved.
		snapshot.InventoryItemID = inspectionInventoryItemID
		snapshot.InventoryStatusBefore = sql.NullString{Valid: true, String: "AVAILABLE"}
	}

	if snapshot.AcquisitionItemID.Valid {
		var currentInspectionID, currentStatus, currentInspectionStatus, currentInventoryItemID sql.NullString
		if err := tx.QueryRowxContext(ctx, `SELECT inspection_id, item_status, inspection_status, inventory_item_id FROM acquisition_items WHERE id = $1`, snapshot.AcquisitionItemID.String).Scan(&currentInspectionID, &currentStatus, &currentInspectionStatus, &currentInventoryItemID); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("read acquisition item during inspection deletion: %w", err)
		} else if err == nil && currentInspectionID.Valid && currentInspectionID.String == id.String() {
			previousInventoryID := nullStringValue(snapshot.AcquisitionInventoryItemIDBefore)
			if snapshot.AcquisitionInventoryItemIDBefore.Valid && currentInventoryItemID.Valid && snapshot.AcquisitionInventoryItemIDBefore.String != currentInventoryItemID.String {
				previousInventoryID = currentInventoryItemID.String
			}
			previousInspectionStatus := nullStringValue(snapshot.AcquisitionInspectionStatusBefore)
			if currentInspectionStatus.Valid && !strings.EqualFold(currentInspectionStatus.String, result) {
				previousInspectionStatus = currentInspectionStatus.String
			}
			previousItemStatus := nullStringValue(snapshot.AcquisitionItemStatusBefore)
			if currentStatus.Valid && !strings.EqualFold(currentStatus.String, inspectionResultAcquisitionStatus(result)) {
				previousItemStatus = currentStatus.String
			}
			if _, err := tx.ExecContext(ctx, `UPDATE acquisition_items SET inventory_item_id = $1, inspection_id = $2,
				inspection_status = $3, item_status = $4, updated_at = $5 WHERE id = $6`,
				previousInventoryID, nullStringValue(snapshot.AcquisitionInspectionIDBefore), previousInspectionStatus,
				previousItemStatus, time.Now().UTC(), snapshot.AcquisitionItemID.String); err != nil {
				return fmt.Errorf("restore acquisition item before inspection: %w", err)
			}
			if snapshot.AcquisitionID.Valid {
				var otherInspectionCount int
				if err := tx.GetContext(ctx, &otherInspectionCount, `SELECT COUNT(*) FROM acquisition_items WHERE acquisition_id = $1 AND inspection_id IS NOT NULL`, snapshot.AcquisitionID.String); err != nil {
					return fmt.Errorf("check related acquisition inspections: %w", err)
				}
				if otherInspectionCount == 0 && snapshot.AcquisitionStatusBefore.Valid {
					if _, err := tx.ExecContext(ctx, `UPDATE acquisitions SET status = $1, updated_at = $2 WHERE id = $3`, snapshot.AcquisitionStatusBefore.String, time.Now().UTC(), snapshot.AcquisitionID.String); err != nil {
						return fmt.Errorf("restore acquisition status before inspection: %w", err)
					}
				} else if err := recomputeAcquisitionInspectionStatusTx(ctx, tx, snapshot.AcquisitionID.String); err != nil {
					return err
				}
			}
		}
	}

	if snapshot.InventoryItemID.Valid && snapshot.InventoryStatusBefore.Valid {
		expectedCurrent := inspectionResultInventoryStatuses(result)
		if len(expectedCurrent) != 0 {
			placeholders := make([]string, 0, len(expectedCurrent))
			args := []any{snapshot.InventoryStatusBefore.String, time.Now().UTC(), snapshot.InventoryItemID.String}
			for index, status := range expectedCurrent {
				placeholders = append(placeholders, fmt.Sprintf("$%d", index+4))
				args = append(args, status)
			}
			query := fmt.Sprintf(`UPDATE inventory_items SET status = $1, updated_at = $2 WHERE id = $3 AND UPPER(TRIM(COALESCE(status,''))) IN (%s)`, strings.Join(placeholders, ","))
			if _, err := tx.ExecContext(ctx, query, args...); err != nil {
				return fmt.Errorf("restore inventory status before inspection: %w", err)
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM inspection_items WHERE inspection_id = $1`, id); err != nil {
		return fmt.Errorf("delete inspection checklist rows: %w", err)
	}
	resultDelete, err := tx.ExecContext(ctx, `DELETE FROM inspections WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete inspection: %w", err)
	}
	if affected, _ := resultDelete.RowsAffected(); affected == 0 {
		return ErrInspectionNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inspection deletion transaction: %w", err)
	}
	return nil
}

func inspectionResultInventoryStatuses(result string) []string {
	switch strings.ToLower(strings.TrimSpace(result)) {
	case "pending", "needs_repair":
		return []string{"INSPECTION", "DAMAGED"}
	case "passed":
		return []string{"AVAILABLE"}
	case "failed":
		return []string{"DAMAGED"}
	default:
		return nil
	}
}

func inspectionResultAcquisitionStatus(result string) string {
	switch strings.ToLower(strings.TrimSpace(result)) {
	case "passed":
		return "available"
	case "failed":
		return "rejected"
	default:
		return "inspection"
	}
}

func recomputeAcquisitionInspectionStatusTx(ctx context.Context, tx *sqlx.Tx, acquisitionID string) error {
	_, err := tx.ExecContext(ctx, `UPDATE acquisitions SET status = CASE
		WHEN EXISTS (SELECT 1 FROM acquisition_items WHERE acquisition_id = $1 AND inspection_status = 'failed') THEN 'rejected'
		WHEN NOT EXISTS (SELECT 1 FROM acquisition_items WHERE acquisition_id = $1 AND inspection_status IN ('pending','needs_repair')) THEN 'approved'
		ELSE 'inspection' END, updated_at = $2 WHERE id = $1`, acquisitionID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("recompute acquisition status after inspection deletion: %w", err)
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
	storeNow := accounting.StoreNow()
	weekStart, _, err := accounting.StoreWeekBounds(storeNow)
	if err != nil {
		return nil, fmt.Errorf("calculate store week boundary: %w", err)
	}
	monthStart, monthEnd, err := accounting.StoreMonthBounds(storeNow)
	if err != nil {
		return nil, fmt.Errorf("calculate store month boundary: %w", err)
	}
	weekStartDate, err := accounting.StoreDate(weekStart)
	if err != nil {
		return nil, fmt.Errorf("calculate store week date: %w", err)
	}
	monthStartDate, err := accounting.StoreDate(monthStart)
	if err != nil {
		return nil, fmt.Errorf("calculate store month date: %w", err)
	}
	monthEndDate, err := accounting.StoreDate(monthEnd)
	if err != nil {
		return nil, fmt.Errorf("calculate next store month date: %w", err)
	}
	weekQuery := `SELECT COUNT(*) FROM inspections WHERE inspection_date >= $1`
	monthQuery := `SELECT COUNT(*) FROM inspections WHERE inspection_date >= $1 AND inspection_date < $2`
	if dbutil.IsSQLite(r.db) {
		weekQuery = `SELECT COUNT(*) FROM inspections WHERE date(inspection_date) >= date(?)`
		monthQuery = `SELECT COUNT(*) FROM inspections WHERE date(inspection_date) >= date(?) AND date(inspection_date) < date(?)`
	}
	err = r.db.GetContext(ctx, &summary.ThisWeek, weekQuery, weekStartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get this week inspections: %w", err)
	}

	// This month inspections
	err = r.db.GetContext(ctx, &summary.ThisMonth, monthQuery, monthStartDate, monthEndDate)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate inspections by condition: %w", err)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate inspections by grade: %w", err)
	}

	return &summary, nil
}
