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
		INSERT INTO inventory_items (product_id, part_type_id, item_code, barcode, serial_number,
		 condition, grade, purchase_cost, selling_price, status, location_id,
		 supplier_id, purchase_date, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		item.ProductID, item.PartTypeID, item.ItemCode, item.Barcode, item.SerialNumber,
		item.Condition, item.Grade, item.PurchaseCost, item.SellingPrice,
		item.Status, item.LocationID, item.SupplierID, item.PurchaseDate, item.Notes,
		item.CreatedAt, item.UpdatedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inventory item: %w", err)
	}

	return nil
}

// GetInventoryItemByID retrieves an inventory item by ID
func (r *Repository) GetInventoryItemByID(ctx context.Context, id uuid.UUID) (*InventoryItem, error) {
	query := `
		SELECT id, product_id, part_type_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii2
			   WHERE ii2.product_id = inventory_items.product_id
			   ), 0) AS current_quantity,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii3
			   WHERE ii3.product_id = inventory_items.product_id
			   AND ii3.status = 'AVAILABLE'
			   ), 0) AS available_quantity
		FROM inventory_items
		WHERE id = $1
	`

	var item InventoryItem
	err := r.db.GetContext(ctx, &item, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	return &item, nil
}

// GetInventoryItemByBarcode retrieves an inventory item by barcode
func (r *Repository) GetInventoryItemByBarcode(ctx context.Context, barcode string) (*InventoryItem, error) {
	query := `
		SELECT id, product_id, part_type_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii2
			   WHERE ii2.product_id = inventory_items.product_id
			   ), 0) AS current_quantity,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii3
			   WHERE ii3.product_id = inventory_items.product_id
			   AND ii3.status = 'AVAILABLE'
			   ), 0) AS available_quantity
		FROM inventory_items
		WHERE barcode = $1
	`

	var item InventoryItem
	err := r.db.GetContext(ctx, &item, query, barcode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get inventory item by barcode: %w", err)
	}

	return &item, nil
}

// GetInventoryItemBySerialNumber retrieves an inventory item by serial number
func (r *Repository) GetInventoryItemBySerialNumber(ctx context.Context, serialNumber string) (*InventoryItem, error) {
	query := `
		SELECT id, product_id, part_type_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii2
			   WHERE ii2.product_id = inventory_items.product_id
			   ), 0) AS current_quantity,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii3
			   WHERE ii3.product_id = inventory_items.product_id
			   AND ii3.status = 'AVAILABLE'
			   ), 0) AS available_quantity
		FROM inventory_items
		WHERE serial_number = $1
	`

	var item InventoryItem
	err := r.db.GetContext(ctx, &item, query, serialNumber)
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
		SET product_id = $2, part_type_id = $3, item_code = $4, barcode = $5, serial_number = $6,
		    condition = $7, grade = $8, purchase_cost = $9, selling_price = $10,
		    status = $11, location_id = $12, supplier_id = $13, purchase_date = $14,
		    sold_at = $15, notes = $16, updated_at = $17
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProductID, item.PartTypeID, item.ItemCode, item.Barcode, item.SerialNumber,
		item.Condition, item.Grade, item.PurchaseCost, item.SellingPrice,
		item.Status, item.LocationID, item.SupplierID, item.PurchaseDate, item.SoldAt,
		item.Notes, item.UpdatedAt,
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
func (r *Repository) UpdateItemStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE inventory_items
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, status)
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
func (r *Repository) ListInventoryItems(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*InventoryItem, int64, error) {
	// Build base query
	baseQuery := `
		SELECT id, product_id, part_type_id, item_code, barcode, serial_number,
		       condition, grade, purchase_cost, selling_price, status, location_id,
		       supplier_id, purchase_date, sold_at, notes, created_at, updated_at,
			   COALESCE((
		       SELECT COUNT(*)
		       FROM inventory_items ii2
		       WHERE ii2.product_id = inventory_items.product_id
			   ), 0) AS current_quantity,
			   COALESCE((
		       SELECT COUNT(*)
		       FROM inventory_items ii3
		       WHERE ii3.product_id = inventory_items.product_id
		       AND ii3.status = 'AVAILABLE'
			   ), 0) AS available_quantity
		FROM inventory_items
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM inventory_items WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	// Add filters if provided
	if condition, ok := filters["condition"].(string); ok && condition != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND condition = ` + param
		countQuery += ` AND condition = ` + param
		args = append(args, condition)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND status = ` + param
		countQuery += ` AND status = ` + param
		args = append(args, status)
	}

	if productID, ok := filters["product_id"].(uuid.UUID); ok && productID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND product_id = ` + param
		countQuery += ` AND product_id = ` + param
		args = append(args, productID)
	}

	if partTypeID, ok := filters["part_type_id"].(uuid.UUID); ok && partTypeID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND part_type_id = ` + param
		countQuery += ` AND part_type_id = ` + param
		args = append(args, partTypeID)
	}

	if locationID, ok := filters["location_id"].(uuid.UUID); ok && locationID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND location_id = ` + param
		countQuery += ` AND location_id = ` + param
		args = append(args, locationID)
	}

	// Get total count
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory items: %w", err)
	}

	// Add pagination
	argCount++
	param := fmt.Sprintf("$%d", argCount)
	baseQuery += ` ORDER BY created_at DESC LIMIT ` + param
	args = append(args, limit)

	argCount++
	param = fmt.Sprintf("$%d", argCount)
	baseQuery += ` OFFSET ` + param
	args = append(args, offset)

	// Execute query
	var items []*InventoryItem
	err = r.db.SelectContext(ctx, &items, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory items: %w", err)
	}

	return items, total, nil
}

// ListInventoryItemsWithSupplierInfo retrieves inventory items with supplier information
func (r *Repository) ListInventoryItemsWithSupplierInfo(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*InventoryItemWithSupplier, int64, error) {
	// Build base query with JOIN to suppliers
	baseQuery := `
		SELECT 
			ii.id, ii.product_id, ii.part_type_id, ii.item_code, ii.barcode, ii.serial_number,
			ii.condition, ii.grade, ii.purchase_cost, ii.selling_price, ii.status, ii.location_id,
			ii.supplier_id, ii.purchase_date, ii.sold_at, ii.notes, ii.created_at, ii.updated_at,
			COALESCE((
				SELECT COUNT(*)
				FROM inventory_items ii2
				WHERE ii2.product_id = ii.product_id
			), 0) AS current_quantity,
			COALESCE((
				SELECT COUNT(*)
				FROM inventory_items ii3
				WHERE ii3.product_id = ii.product_id
				AND ii3.status = 'AVAILABLE'
			), 0) AS available_quantity,
			p.name as product_name,
			s.name as supplier_name,
			s.phone as supplier_phone
		FROM inventory_items ii
		LEFT JOIN products p ON ii.product_id = p.id
		LEFT JOIN suppliers s ON ii.supplier_id = s.id
		WHERE 1=1
	`
	countQuery := `
		SELECT COUNT(*) FROM inventory_items ii
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add filters if provided
	if condition, ok := filters["condition"].(string); ok && condition != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.condition = ` + param
		countQuery += ` AND ii.condition = ` + param
		args = append(args, condition)
	}
	if excludeCondition, ok := filters["exclude_condition"].(string); ok && excludeCondition != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.condition <> ` + param
		countQuery += ` AND ii.condition <> ` + param
		args = append(args, excludeCondition)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.status = ` + param
		countQuery += ` AND ii.status = ` + param
		args = append(args, status)
	}

	if productID, ok := filters["product_id"].(uuid.UUID); ok && productID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.product_id = ` + param
		countQuery += ` AND ii.product_id = ` + param
		args = append(args, productID)
	}

	if partTypeID, ok := filters["part_type_id"].(uuid.UUID); ok && partTypeID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.part_type_id = ` + param
		countQuery += ` AND ii.part_type_id = ` + param
		args = append(args, partTypeID)
	}

	if locationID, ok := filters["location_id"].(uuid.UUID); ok && locationID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.location_id = ` + param
		countQuery += ` AND ii.location_id = ` + param
		args = append(args, locationID)
	}

	// New filters for supplier and purchase
	if supplierID, ok := filters["supplier_id"].(uuid.UUID); ok && supplierID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.supplier_id = ` + param
		countQuery += ` AND ii.supplier_id = ` + param
		args = append(args, supplierID)
	}

	if purchaseDateFrom, ok := filters["purchase_date_from"].(string); ok && purchaseDateFrom != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.purchase_date >= ` + param
		countQuery += ` AND ii.purchase_date >= ` + param
		args = append(args, purchaseDateFrom)
	}

	if purchaseDateTo, ok := filters["purchase_date_to"].(string); ok && purchaseDateTo != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.purchase_date <= ` + param
		countQuery += ` AND ii.purchase_date <= ` + param
		args = append(args, purchaseDateTo)
	}

	if minPurchaseCost, ok := filters["min_purchase_cost"].(float64); ok && minPurchaseCost > 0 {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.purchase_cost >= ` + param
		countQuery += ` AND ii.purchase_cost >= ` + param
		args = append(args, minPurchaseCost)
	}

	if maxPurchaseCost, ok := filters["max_purchase_cost"].(float64); ok && maxPurchaseCost > 0 {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ii.purchase_cost <= ` + param
		countQuery += ` AND ii.purchase_cost <= ` + param
		args = append(args, maxPurchaseCost)
	}

	// Get total count
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory items: %w", err)
	}

	// Add pagination
	argCount++
	param := fmt.Sprintf("$%d", argCount)
	baseQuery += ` ORDER BY ii.created_at DESC LIMIT ` + param
	args = append(args, limit)

	argCount++
	param = fmt.Sprintf("$%d", argCount)
	baseQuery += ` OFFSET ` + param
	args = append(args, offset)

	// Execute query
	var items []*InventoryItemWithSupplier
	err = r.db.SelectContext(ctx, &items, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory items with supplier info: %w", err)
	}

	return items, total, nil
}

// CreateLocation creates a new location
func (r *Repository) CreateLocation(ctx context.Context, location *Location) error {
	query := `
		INSERT INTO locations (name, type, parent_id, warehouse_id, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		location.Name, location.Type, location.ParentID, location.WarehouseID, location.Description,
		location.IsActive, location.CreatedAt, location.UpdatedAt,
	).Scan(&location.ID, &location.CreatedAt, &location.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create location: %w", err)
	}

	return nil
}

// GetLocationByID retrieves a location by ID
func (r *Repository) GetLocationByID(ctx context.Context, id uuid.UUID) (*Location, error) {
	query := `
		SELECT id, name, type, parent_id, warehouse_id, description, is_active, created_at, updated_at
		FROM locations
		WHERE id = $1
	`

	var location Location
	err := r.db.GetContext(ctx, &location, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLocationNotFound
		}
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	return &location, nil
}

// ListLocations retrieves all locations
func (r *Repository) ListLocations(ctx context.Context) ([]*Location, error) {
	query := `
		SELECT id, name, type, parent_id, warehouse_id, description, is_active, created_at, updated_at
		FROM locations
		ORDER BY name ASC
	`

	var locations []*Location
	err := r.db.SelectContext(ctx, &locations, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	return locations, nil
}

// CreateMovement creates a new inventory movement
func (r *Repository) CreateMovement(ctx context.Context, movement *InventoryMovement) error {
	query := `
		INSERT INTO inventory_movements (item_id, movement_type, quantity,
		 before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		movement.ItemID, movement.MovementType, movement.Quantity, movement.BeforeQuantity,
		movement.AfterQuantity, movement.ReferenceType, movement.ReferenceID,
		movement.Reason, movement.CreatedBy, movement.CreatedAt,
	).Scan(&movement.ID, &movement.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	return nil
}

// GetMovementsByItem retrieves movements for a specific item
func (r *Repository) GetMovementsByItem(ctx context.Context, itemID uuid.UUID, limit, offset int) ([]*InventoryMovement, int64, error) {
	query := `
		SELECT id, item_id, movement_type, quantity,
		       before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at
		FROM inventory_movements
		WHERE item_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*)
		FROM inventory_movements
		WHERE item_id = $1
	`

	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, itemID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count movements: %w", err)
	}

	var movements []*InventoryMovement
	err = r.db.SelectContext(ctx, &movements, query, itemID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get movements: %w", err)
	}

	return movements, total, nil
}

// CreateReservation creates a new reservation
func (r *Repository) CreateReservation(ctx context.Context, reservation *Reservation) error {
	query := `
		INSERT INTO reservations (item_id, customer_id, user_id, reserved_at, expires_at, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		reservation.ItemID, reservation.CustomerID, reservation.UserID, reservation.ReservedAt,
		reservation.ExpiresAt, reservation.Status, reservation.Notes,
		reservation.CreatedAt, reservation.UpdatedAt,
	).Scan(&reservation.ID, &reservation.CreatedAt, &reservation.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	return nil
}

// GetActiveReservationByItem retrieves active reservation for an item
func (r *Repository) GetActiveReservationByItem(ctx context.Context, itemID uuid.UUID) (*Reservation, error) {
	query := `
		SELECT id, item_id, customer_id, user_id, reserved_at,
		       expires_at, status, notes, created_at, updated_at
		FROM reservations
		WHERE item_id = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var reservation Reservation
	err := r.db.GetContext(ctx, &reservation, query, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active reservation
		}
		return nil, fmt.Errorf("failed to get active reservation: %w", err)
	}

	return &reservation, nil
}

// UpdateReservationStatus updates reservation status
func (r *Repository) UpdateReservationStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE reservations
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, status)
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
