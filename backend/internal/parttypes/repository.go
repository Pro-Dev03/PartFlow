package parttypes

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Part Types CRUD
func parsePartTypeMap(record map[string]any) (PartType, error) {
	var partType PartType
	if raw, ok := record["id"]; ok && raw != nil {
		parsed, err := uuid.Parse(raw.(string))
		if err != nil {
			return partType, fmt.Errorf("parse part type id: %w", err)
		}
		partType.ID = parsed
	}
	if raw, ok := record["name_ar"]; ok && raw != nil {
		partType.NameAr = raw.(string)
	}
	if raw, ok := record["name_en"]; ok && raw != nil {
		partType.NameEn = raw.(string)
	}
	if raw, ok := record["icon"]; ok && raw != nil {
		partType.Icon = raw.(string)
	}
	if raw, ok := record["color"]; ok && raw != nil {
		partType.Color = raw.(string)
	}
	if raw, ok := record["is_active"]; ok && raw != nil {
		switch v := raw.(type) {
		case bool:
			partType.IsActive = v
		case int:
			partType.IsActive = v != 0
		case int64:
			partType.IsActive = v != 0
		case string:
			partType.IsActive = v == "1" || v == "true" || v == "TRUE"
		}
	}
	if raw, ok := record["sort_order"]; ok && raw != nil {
		switch v := raw.(type) {
		case int64:
			partType.SortOrder = int(v)
		case int32:
			partType.SortOrder = int(v)
		case float64:
			partType.SortOrder = int(v)
		case int:
			partType.SortOrder = v
		}
	}
	if raw, ok := record["created_at"]; ok && raw != nil {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return partType, fmt.Errorf("parse created_at: %w", err)
		}
		partType.CreatedAt = parsed
	}
	if raw, ok := record["updated_at"]; ok && raw != nil {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return partType, fmt.Errorf("parse updated_at: %w", err)
		}
		partType.UpdatedAt = parsed
	}
	return partType, nil
}

func (r *Repository) ListPartTypes(ctx context.Context) ([]PartType, error) {
	query := `
		SELECT id, name_ar, name_en, icon, color, is_active, sort_order, created_at, updated_at
		FROM part_types
		ORDER BY sort_order ASC, name_ar ASC
	`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list part types: %w", err)
	}
	defer rows.Close()

	var types []PartType
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, fmt.Errorf("failed to scan part type: %w", err)
		}
		partType, err := parsePartTypeMap(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse part type: %w", err)
		}
		types = append(types, partType)
	}
	return types, rows.Err()
}

func (r *Repository) GetPartType(ctx context.Context, id uuid.UUID) (*PartType, error) {
	query := `
		SELECT id, name_ar, name_en, icon, color, is_active, sort_order, created_at, updated_at
		FROM part_types
		WHERE id = $1
	`

	row := r.db.QueryRowxContext(ctx, query, id)
	if row == nil {
		return nil, fmt.Errorf("failed to get part type: %w", sql.ErrNoRows)
	}
	record := map[string]any{}
	if err := row.MapScan(record); err != nil {
		return nil, fmt.Errorf("failed to get part type: %w", err)
	}
	partType, err := parsePartTypeMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse part type: %w", err)
	}
	return &partType, nil
}

func (r *Repository) CreatePartType(ctx context.Context, partType *PartType) error {
	query := `
		INSERT INTO part_types (id, name_ar, name_en, icon, color, is_active, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		partType.ID, partType.NameAr, partType.NameEn, partType.Icon, partType.Color, partType.IsActive, partType.SortOrder,
	).Scan(&partType.CreatedAt, &partType.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create part type: %w", err)
	}

	return nil
}

func (r *Repository) UpdatePartType(ctx context.Context, partType *PartType) error {
	query := `
		UPDATE part_types
		SET name_ar = COALESCE($2, name_ar),
		    name_en = COALESCE($3, name_en),
		    icon = COALESCE($4, icon),
		    color = COALESCE($5, color),
		    is_active = COALESCE($6, is_active),
		    sort_order = COALESCE($7, sort_order),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		partType.ID, partType.NameAr, partType.NameEn, partType.Icon,
		partType.Color, partType.IsActive, partType.SortOrder,
	).Scan(&partType.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to update part type: %w", err)
	}
	
	return nil
}

func (r *Repository) DeletePartType(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM part_types WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete part type: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// Specifications CRUD
func (r *Repository) ListSpecifications(ctx context.Context) ([]PartSpecification, error) {
	query := `
		SELECT id, name_ar, name_en, data_type, options, is_required, created_at
		FROM part_specifications
		ORDER BY name_ar ASC
	`
	
	var specs []PartSpecification
	err := r.db.SelectContext(ctx, &specs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list specifications: %w", err)
	}
	
	return specs, nil
}

func (r *Repository) CreateSpecification(ctx context.Context, spec *PartSpecification) error {
	query := `
		INSERT INTO part_specifications (id, name_ar, name_en, data_type, options, is_required, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		spec.ID, spec.NameAr, spec.NameEn, spec.DataType, spec.Options, spec.IsRequired,
	).Scan(&spec.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create specification: %w", err)
	}
	
	return nil
}

// Type Specifications Linking
func (r *Repository) GetTypeSpecifications(ctx context.Context, partTypeID uuid.UUID) ([]PartSpecification, error) {
	query := `
		SELECT ps.id, ps.name_ar, ps.name_en, ps.data_type, ps.options, ps.is_required, ps.created_at
		FROM part_specifications ps
		INNER JOIN type_specifications ts ON ps.id = ts.specification_id
		WHERE ts.part_type_id = $1
		ORDER BY ts.sort_order ASC, ps.name_ar ASC
	`
	
	var specs []PartSpecification
	err := r.db.SelectContext(ctx, &specs, query, partTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get type specifications: %w", err)
	}
	
	return specs, nil
}

func (r *Repository) LinkSpecification(ctx context.Context, link *TypeSpecification) error {
	query := `
		INSERT INTO type_specifications (id, part_type_id, specification_id, sort_order, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (part_type_id, specification_id) 
		DO UPDATE SET sort_order = EXCLUDED.sort_order
		RETURNING created_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		link.ID, link.PartTypeID, link.SpecificationID, link.SortOrder,
	).Scan(&link.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to link specification: %w", err)
	}
	
	return nil
}

func (r *Repository) UnlinkSpecification(ctx context.Context, partTypeID, specificationID uuid.UUID) error {
	query := `DELETE FROM type_specifications WHERE part_type_id = $1 AND specification_id = $2`
	
	_, err := r.db.ExecContext(ctx, query, partTypeID, specificationID)
	if err != nil {
		return fmt.Errorf("failed to unlink specification: %w", err)
	}
	
	return nil
}

// Item Specification Values
func (r *Repository) GetItemSpecifications(ctx context.Context, inventoryItemID uuid.UUID) ([]ItemSpecificationValue, error) {
	query := `
		SELECT id, inventory_item_id, specification_id, value_text, value_number, value_boolean, created_at, updated_at
		FROM item_specification_values
		WHERE inventory_item_id = $1
	`
	
	var values []ItemSpecificationValue
	err := r.db.SelectContext(ctx, &values, query, inventoryItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item specifications: %w", err)
	}
	
	return values, nil
}

func (r *Repository) SetItemSpecification(ctx context.Context, value *ItemSpecificationValue) error {
	query := `
		INSERT INTO item_specification_values (id, inventory_item_id, specification_id, value_text, value_number, value_boolean, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (inventory_item_id, specification_id)
		DO UPDATE SET 
			value_text = EXCLUDED.value_text,
			value_number = EXCLUDED.value_number,
			value_boolean = EXCLUDED.value_boolean,
			updated_at = NOW()
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		value.ID, value.InventoryItemID, value.SpecificationID,
		value.ValueText, value.ValueNumber, value.ValueBoolean,
	).Scan(&value.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to set item specification: %w", err)
	}
	
	return nil
}

func (r *Repository) DeleteItemSpecifications(ctx context.Context, inventoryItemID uuid.UUID) error {
	query := `DELETE FROM item_specification_values WHERE inventory_item_id = $1`
	
	_, err := r.db.ExecContext(ctx, query, inventoryItemID)
	if err != nil {
		return fmt.Errorf("failed to delete item specifications: %w", err)
	}
	
	return nil
}

// Get part type with specifications
func (r *Repository) GetPartTypeWithSpecs(ctx context.Context, id uuid.UUID) (*PartTypeWithSpecs, error) {
	partType, err := r.GetPartType(ctx, id)
	if err != nil {
		return nil, err
	}
	
	specs, err := r.GetTypeSpecifications(ctx, id)
	if err != nil {
		return nil, err
	}
	
	return &PartTypeWithSpecs{
		PartType:       *partType,
		Specifications: specs,
	}, nil
}