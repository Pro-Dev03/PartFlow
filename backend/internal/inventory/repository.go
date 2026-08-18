package inventory

import (
	"context"
	"database/sql"
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

// CreateInventoryItem creates a new inventory item
func (r *Repository) CreateInventoryItem(ctx context.Context, item *InventoryItem) error {
	query := `
		INSERT INTO inventory_items 
		(id, organization_id, product_id, item_code, barcode, serial_number, 
		 condition, grade, purchase_cost, selling_price, status, location_id, 
		 supplier_id, purchase_date, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.OrganizationID, item.ProductID, item.ItemCode, item.Barcode,
		item.SerialNumber, item.Condition, item.Grade, item.PurchaseCost, item.SellingPrice,
		item.Status, item.LocationID, item.SupplierID, item.PurchaseDate, item.Notes,
		item.CreatedAt, item.UpdatedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inventory item: %w", err)
	}

	return nil
}

// GetInventoryItemByID retrieves an inventory item by ID
func (r *Repository) GetInventoryItemByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*InventoryItem, error) {
	query := `
		SELECT id, organization_id, product_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		WHERE id = $1 AND organization_id = $2
	`

	var item InventoryItem
	err := r.db.GetContext(ctx, &item, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	return &item, nil
}

// GetInventoryItemByBarcode retrieves an inventory item by barcode
func (r *Repository) GetInventoryItemByBarcode(ctx context.Context, barcode string, organizationID uuid.UUID) (*InventoryItem, error) {
	query := `
		SELECT id, organization_id, product_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		WHERE barcode = $1 AND organization_id = $2
	`

	var item InventoryItem
	err := r.db.GetContext(ctx, &item, query, barcode, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get inventory item by barcode: %w", err)
	}

	return &item, nil
}

// GetInventoryItemBySerialNumber retrieves an inventory item by serial number
func (r *Repository) GetInventoryItemBySerialNumber(ctx context.Context, serialNumber string, organizationID uuid.UUID) (*InventoryItem, error) {
	query := `
		SELECT id, organization_id, product_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		WHERE serial_number = $1 AND organization_id = $2
	`

	var item InventoryItem
	err := r.db.GetContext(ctx, &item, query, serialNumber, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get inventory item by serial number: %w", err)
	}

	return &item, nil
}

// UpdateInventoryItem updates an inventory item
func (r *Repository) UpdateInventoryItem(ctx context.Context, item *InventoryItem) error {
	query := `
		UPDATE inventory_items
		SET product_id = $2, item_code = $3, barcode = $4, serial_number = $5,
		    condition = $6, grade = $7, purchase_cost = $8, selling_price = $9,
		    status = $10, location_id = $11, supplier_id = $12, purchase_date = $13,
		    sold_at = $14, notes = $15, updated_at = $16
		WHERE id = $1 AND organization_id = $17
	`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProductID, item.ItemCode, item.Barcode, item.SerialNumber,
		item.Condition, item.Grade, item.PurchaseCost, item.SellingPrice,
		item.Status, item.LocationID, item.SupplierID, item.PurchaseDate,
		item.SoldAt, item.Notes, item.UpdatedAt, item.OrganizationID,
	)

	if err != nil {
		return fmt.Errorf("failed to update inventory item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrItemNotFound
	}

	return nil
}

// UpdateItemStatus updates the status of an inventory item
func (r *Repository) UpdateItemStatus(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, status Status) error {
	query := `
		UPDATE inventory_items
		SET status = $3, updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, organizationID, status)
	if err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrItemNotFound
	}

	return nil
}

// ListInventoryItems retrieves a list of inventory items with pagination
func (r *Repository) ListInventoryItems(ctx context.Context, organizationID uuid.UUID, limit, offset int, filters map[string]interface{}) ([]*InventoryItem, int64, error) {
	// Build base query
	baseQuery := `
		SELECT id, organization_id, product_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		WHERE organization_id = $1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM inventory_items
		WHERE organization_id = $1
	`

	args := []interface{}{organizationID}
	argCount := 1

	// Add filters
	if status, ok := filters["status"]; ok {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if condition, ok := filters["condition"]; ok {
		argCount++
		baseQuery += fmt.Sprintf(" AND condition = $%d", argCount)
		countQuery += fmt.Sprintf(" AND condition = $%d", argCount)
		args = append(args, condition)
	}

	if locationID, ok := filters["location_id"]; ok {
		argCount++
		baseQuery += fmt.Sprintf(" AND location_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND location_id = $%d", argCount)
		args = append(args, locationID)
	}

	if productID, ok := filters["product_id"]; ok {
		argCount++
		baseQuery += fmt.Sprintf(" AND product_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND product_id = $%d", argCount)
		args = append(args, productID)
	}

	// Get total count
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory items: %w", err)
	}

	// Add pagination
	argCount++
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	// Execute query
	var items []*InventoryItem
	err = r.db.SelectContext(ctx, &items, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory items: %w", err)
	}

	return items, total, nil
}

// CreateLocation creates a new location
func (r *Repository) CreateLocation(ctx context.Context, location *Location) error {
	query := `
		INSERT INTO locations 
		(id, organization_id, name, type, parent_id, warehouse_id, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		location.ID, location.OrganizationID, location.Name, location.Type,
		location.ParentID, location.WarehouseID, location.Description,
		location.IsActive, location.CreatedAt, location.UpdatedAt,
	).Scan(&location.ID, &location.CreatedAt, &location.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create location: %w", err)
	}

	return nil
}

// GetLocationByID retrieves a location by ID
func (r *Repository) GetLocationByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Location, error) {
	query := `
		SELECT id, organization_id, name, type, parent_id, warehouse_id, 
		       description, is_active, created_at, updated_at
		FROM locations
		WHERE id = $1 AND organization_id = $2
	`

	var location Location
	err := r.db.GetContext(ctx, &location, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLocationNotFound
		}
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	return &location, nil
}

// ListLocations retrieves all locations for an organization
func (r *Repository) ListLocations(ctx context.Context, organizationID uuid.UUID) ([]*Location, error) {
	query := `
		SELECT id, organization_id, name, type, parent_id, warehouse_id,
		       description, is_active, created_at, updated_at
		FROM locations
		WHERE organization_id = $1
		ORDER BY name ASC
	`

	var locations []*Location
	err := r.db.SelectContext(ctx, &locations, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	return locations, nil
}

// CreateMovement creates a new inventory movement
func (r *Repository) CreateMovement(ctx context.Context, movement *InventoryMovement) error {
	query := `
		INSERT INTO inventory_movements 
		(id, organization_id, item_id, product_id, movement_type, quantity,
		 before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		movement.ID, movement.OrganizationID, movement.ItemID, movement.ProductID,
		movement.MovementType, movement.Quantity, movement.BeforeQuantity,
		movement.AfterQuantity, movement.ReferenceType, movement.ReferenceID,
		movement.Reason, movement.CreatedBy, movement.CreatedAt,
	).Scan(&movement.ID, &movement.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	return nil
}

// GetMovementsByItem retrieves movements for a specific item
func (r *Repository) GetMovementsByItem(ctx context.Context, itemID uuid.UUID, organizationID uuid.UUID, limit, offset int) ([]*InventoryMovement, int64, error) {
	query := `
		SELECT id, organization_id, item_id, product_id, movement_type, quantity,
		       before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at
		FROM inventory_movements
		WHERE item_id = $1 AND organization_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	countQuery := `
		SELECT COUNT(*)
		FROM inventory_movements
		WHERE item_id = $1 AND organization_id = $2
	`

	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, itemID, organizationID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count movements: %w", err)
	}

	var movements []*InventoryMovement
	err = r.db.SelectContext(ctx, &movements, query, itemID, organizationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get movements: %w", err)
	}

	return movements, total, nil
}

// CreateReservation creates a new reservation
func (r *Repository) CreateReservation(ctx context.Context, reservation *Reservation) error {
	query := `
		INSERT INTO reservations 
		(id, organization_id, item_id, customer_id, user_id, reserved_at, expires_at, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		reservation.ID, reservation.OrganizationID, reservation.ItemID,
		reservation.CustomerID, reservation.UserID, reservation.ReservedAt,
		reservation.ExpiresAt, reservation.Status, reservation.Notes,
		reservation.CreatedAt, reservation.UpdatedAt,
	).Scan(&reservation.ID, &reservation.CreatedAt, &reservation.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	return nil
}

// GetActiveReservationByItem retrieves active reservation for an item
func (r *Repository) GetActiveReservationByItem(ctx context.Context, itemID uuid.UUID, organizationID uuid.UUID) (*Reservation, error) {
	query := `
		SELECT id, organization_id, item_id, customer_id, user_id, reserved_at, 
		       expires_at, status, notes, created_at, updated_at
		FROM reservations
		WHERE item_id = $1 AND organization_id = $2 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var reservation Reservation
	err := r.db.GetContext(ctx, &reservation, query, itemID, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active reservation
		}
		return nil, fmt.Errorf("failed to get active reservation: %w", err)
	}

	return &reservation, nil
}

// UpdateReservationStatus updates reservation status
func (r *Repository) UpdateReservationStatus(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, status string) error {
	query := `
		UPDATE reservations
		SET status = $3, updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, organizationID, status)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrItemNotFound
	}

	return nil
}
