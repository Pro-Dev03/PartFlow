package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type Service struct {
	repo *Repository
	db   *sqlx.DB
}

func NewService(repo *Repository, db *sqlx.DB) *Service {
	return &Service{repo: repo, db: db}
}

// CreateInventoryItem creates a new inventory item with validation
func (s *Service) CreateInventoryItem(ctx context.Context, req *InventoryItemRequest, userID uuid.UUID) (*InventoryItem, error) {
	// Validate condition
	if !isValidCondition(req.Condition) {
		return nil, ErrInvalidCondition
	}

	// Validate grade if condition is used
	if req.Condition == ConditionUsed && req.Grade != nil && !isValidGrade(*req.Grade) {
		return nil, ErrInvalidGrade
	}

	// Generate item code if not provided
	itemCode := req.ItemCode
	if itemCode == nil || *itemCode == "" {
		generatedCode := generateItemCode()
		itemCode = &generatedCode
	}

	// Generate barcode if not provided
	barcode := req.Barcode
	if barcode == nil || *barcode == "" {
		generatedBarcode := generateBarcode(req.ProductID, req.PartTypeID)
		barcode = &generatedBarcode
	}

	// Check if barcode already exists
	if barcode != nil && *barcode != "" {
		_, err := s.repo.GetInventoryItemByBarcode(ctx, *barcode)
		if err == nil {
			return nil, ErrDuplicateBarcode
		}
	}

	// Check if serial number already exists
	if req.SerialNumber != nil && *req.SerialNumber != "" {
		_, err := s.repo.GetInventoryItemBySerialNumber(ctx, *req.SerialNumber)
		if err == nil {
			return nil, ErrDuplicateSerialNumber
		}
	}

	now := time.Now()

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = StatusPurchased
	}

	// Convert Grade pointer
	var gradePtr *string
	if req.Grade != nil {
		gradeStr := string(*req.Grade)
		gradePtr = &gradeStr
	}

	item := &InventoryItem{
		ID:           uuid.New(),
		ProductID:    req.ProductID,
		PartTypeID:   req.PartTypeID,
		ItemCode:     itemCode,
		Barcode:      barcode,
		SerialNumber: req.SerialNumber,
		Condition:    string(req.Condition),
		Grade:        gradePtr,
		PurchaseCost: req.PurchaseCost,
		SellingPrice: req.SellingPrice,
		Status:       string(status),
		LocationID:   req.LocationID,
		SupplierID:   req.SupplierID,
		Notes:        req.Notes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.CreateInventoryItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}

	return item, nil
}

// GetInventoryItem retrieves an inventory item by ID
func (s *Service) GetInventoryItem(ctx context.Context, id uuid.UUID) (*InventoryItem, error) {
	return s.repo.GetInventoryItemByID(ctx, id)
}

// UpdateItemClassification updates the condition and part type of an inventory item.
func (s *Service) UpdateItemClassification(ctx context.Context, id uuid.UUID, condition Condition, partTypeID *uuid.UUID) (*InventoryItem, error) {
	if !isValidCondition(condition) {
		return nil, ErrInvalidCondition
	}

	item, err := s.repo.GetInventoryItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	item.Condition = string(condition)
	item.PartTypeID = partTypeID
	item.UpdatedAt = time.Now()
	if err := s.repo.UpdateInventoryItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// LookupBarcode looks up a product or item by barcode
func (s *Service) LookupBarcode(ctx context.Context, barcode string) (*InventoryItem, error) {
	// Try to find as inventory item first
	item, err := s.repo.GetInventoryItemByBarcode(ctx, barcode)
	if err == nil {
		return item, nil
	}

	// If not found as item, return error
	return nil, ErrItemNotFound
}

// UpdateItemStatus updates the status of an inventory item with validation
func (s *Service) UpdateItemStatus(ctx context.Context, id uuid.UUID, newStatus string) error {
	newStatus = strings.ToUpper(strings.TrimSpace(newStatus))

	// Get current item
	item, err := s.repo.GetInventoryItemByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate status transition
	if !isValidStatusTransition(item.Status, newStatus) {
		return ErrInvalidStatus
	}

	return s.repo.UpdateItemStatus(ctx, id, newStatus)
}

// ReceiveItem marks an item as received and available with automatic inventory updates
func (s *Service) ReceiveItem(ctx context.Context, id uuid.UUID, locationID *uuid.UUID, userID uuid.UUID) error {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Read and lock the item inside the same transaction used for every update.
	itemQuery := `SELECT id, product_id, part_type_id, item_code, barcode, serial_number, condition, grade, purchase_cost, selling_price, status, location_id, supplier_id, purchase_date, sold_at, notes, created_at, updated_at FROM inventory_items WHERE id = $1`
	if !dbutil.IsSQLite(s.db) {
		itemQuery += " FOR UPDATE"
	}
	var item *InventoryItem
	item, err = getInventoryItemTx(ctx, tx, itemQuery, id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	var currentQuantity int
	if err = tx.GetContext(ctx, &currentQuantity, `SELECT COALESCE(quantity, 0) FROM inventory WHERE product_id = $1 LIMIT 1`, item.ProductID); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to read inventory quantity: %w", err)
	}

	// Update status to available
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE inventory_items SET status = 'AVAILABLE', updated_at = %s WHERE id = $1`, dbutil.NowSQL(s.db)), id); err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}

	// Update location if provided
	if locationID != nil {
		item.LocationID = locationID
		item.UpdatedAt = time.Now()
		if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE inventory_items SET location_id = $1, updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db)), locationID, id); err != nil {
			return fmt.Errorf("failed to update item location: %w", err)
		}
	}

	// Update inventory table (increase stock)
	inventoryUpdateQuery := fmt.Sprintf(`
		UPDATE inventory
		SET quantity = quantity + 1, updated_at = %s
		WHERE product_id = $1
	`, dbutil.NowSQL(s.db))
	result, err := tx.ExecContext(ctx, inventoryUpdateQuery, item.ProductID)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		// Create inventory record if it doesn't exist
		createInventoryQuery := fmt.Sprintf(`
			INSERT INTO inventory (id, product_id, quantity, created_at, updated_at)
			VALUES ($1, $2, 1, %s, %s)
		`, dbutil.NowSQL(s.db), dbutil.NowSQL(s.db))
		_, err = tx.ExecContext(ctx, createInventoryQuery, uuid.New(), item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to create inventory record: %w", err)
		}
	}

	// Create movement record
	reason := "Item received and made available"
	movement := &InventoryMovement{
		ID:           uuid.New(),
		ItemID:       &id,
		ProductID:    item.ProductID,
		MovementType: MovementPurchase,
		Quantity:     1,
		Reason:       &reason,
		CreatedBy:    userID,
	}

	movementQuery := fmt.Sprintf(`INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reason, created_by, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, %s)`, dbutil.NowSQL(s.db))
	if _, err = tx.ExecContext(ctx, movementQuery, movement.ID, movement.ItemID, movement.ProductID, movement.MovementType, movement.Quantity, currentQuantity, currentQuantity+1, movement.Reason, movement.CreatedBy); err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	// Create audit log
	auditQuery := fmt.Sprintf(`
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, %s)
	`, dbutil.NowSQL(s.db))
	changes := fmt.Sprintf("Received item %s, location: %v", id, locationID)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), userID, "RECEIVE_ITEM", "inventory_item", id,
		changes, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	return nil
}

// ReserveItem reserves an item for a customer
func (s *Service) ReserveItem(ctx context.Context, req *ReservationRequest, userID uuid.UUID) (*Reservation, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin reservation: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	var status string
	itemQuery := `SELECT status FROM inventory_items WHERE id = $1`
	if !dbutil.IsSQLite(s.db) {
		itemQuery += " FOR UPDATE"
	}
	if err = tx.GetContext(ctx, &status, itemQuery, req.ItemID); err != nil {
		return nil, ErrItemNotFound
	}
	if status != string(StatusAvailable) {
		return nil, ErrInvalidStatus
	}
	var existing int
	if err = tx.GetContext(ctx, &existing, `SELECT COUNT(*) FROM reservations WHERE item_id = $1 AND status = 'active'`, req.ItemID); err != nil {
		return nil, fmt.Errorf("failed to check reservation: %w", err)
	}
	if existing > 0 {
		return nil, ErrItemAlreadyReserved
	}
	if req.ExpiresIn < 0 {
		return nil, fmt.Errorf("reservation expiration cannot be negative")
	}
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)
	if req.ExpiresIn > 0 {
		expiresAt = now.Add(time.Duration(req.ExpiresIn) * time.Minute)
	}
	reservation := &Reservation{ID: uuid.New(), ItemID: req.ItemID, CustomerID: req.CustomerID, UserID: userID, ReservedAt: now, ExpiresAt: expiresAt, Status: "active", Notes: req.Notes, CreatedAt: now, UpdatedAt: now}
	insert := fmt.Sprintf(`INSERT INTO reservations (id, item_id, customer_id, user_id, reserved_at, expires_at, status, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, 'active', $7, $8, %s)`, dbutil.NowSQL(s.db))
	if _, err = tx.ExecContext(ctx, insert, reservation.ID, reservation.ItemID, reservation.CustomerID, reservation.UserID, reservation.ReservedAt, reservation.ExpiresAt, reservation.Notes, reservation.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE inventory_items SET status = 'RESERVED', updated_at = %s WHERE id = $1`, dbutil.NowSQL(s.db)), req.ItemID); err != nil {
		return nil, fmt.Errorf("failed to reserve item: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE inventory SET reserved_quantity = COALESCE(reserved_quantity, 0) + 1 WHERE product_id = (SELECT product_id FROM inventory_items WHERE id = $1)`, req.ItemID); err != nil {
		return nil, fmt.Errorf("failed to update reserved inventory quantity: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit reservation: %w", err)
	}
	committed = true
	return reservation, nil
}

// ReleaseReservation releases a reservation and makes item available again with full automation
func (s *Service) ReleaseReservation(ctx context.Context, reservationID uuid.UUID, userID uuid.UUID) error {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Get reservation with row lock
	var reservation struct {
		ItemID    uuid.UUID `db:"item_id"`
		ProductID uuid.UUID `db:"product_id"`
		Status    string    `db:"status"`
	}
	reservationQuery := `SELECT r.item_id, ii.product_id, r.status FROM reservations r JOIN inventory_items ii ON ii.id = r.item_id WHERE r.id = $1`
	if !dbutil.IsSQLite(s.db) {
		reservationQuery += " FOR UPDATE"
	}
	err = tx.GetContext(ctx, &reservation, reservationQuery, reservationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if reservation.Status != "active" {
		return fmt.Errorf("reservation is not active")
	}
	var beforeAvailable int
	if err = tx.GetContext(ctx, &beforeAvailable, `SELECT COUNT(*) FROM inventory_items WHERE product_id = $1 AND status = 'AVAILABLE'`, reservation.ProductID); err != nil {
		return fmt.Errorf("failed to read available inventory quantity: %w", err)
	}

	// Update reservation status to cancelled
	updateReservationQuery := fmt.Sprintf(`UPDATE reservations SET status = 'cancelled', updated_at = %s WHERE id = $1`, dbutil.NowSQL(s.db))
	_, err = tx.ExecContext(ctx, updateReservationQuery, reservationID)
	if err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	// Update item status back to available
	updateItemQuery := fmt.Sprintf(`UPDATE inventory_items SET status = 'AVAILABLE', updated_at = %s WHERE id = $1`, dbutil.NowSQL(s.db))
	_, err = tx.ExecContext(ctx, updateItemQuery, reservation.ItemID)
	if err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE inventory SET reserved_quantity = CASE WHEN COALESCE(reserved_quantity, 0) > 0 THEN reserved_quantity - 1 ELSE 0 END WHERE product_id = (SELECT product_id FROM inventory_items WHERE id = $1)`, reservation.ItemID); err != nil {
		return fmt.Errorf("failed to release reserved inventory quantity: %w", err)
	}

	// Create movement record
	releaseReason := "Reservation released"
	movementQuery := `
		INSERT INTO inventory_movements (id, item_id, movement_type,
			quantity, before_quantity, after_quantity, reference_type, reference_id,
			reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(), reservation.ItemID, "RELEASE",
		1, beforeAvailable, beforeAvailable+1, "reservation", reservationID, releaseReason, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	// Create audit log
	auditQuery := fmt.Sprintf(`
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, %s)
	`, dbutil.NowSQL(s.db))
	changes := fmt.Sprintf("Released reservation %s for item %s", reservationID, reservation.ItemID)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), userID, "RELEASE_RESERVATION", "reservation", reservationID,
		changes, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	return nil
}

// ConvertReservationToSale converts a reservation to a sale
func (s *Service) ConvertReservationToSale(ctx context.Context, reservationID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin reservation conversion: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var reservation struct {
		ItemID uuid.UUID `db:"item_id"`
		UserID uuid.UUID `db:"user_id"`
		Status string    `db:"status"`
	}
	query := `SELECT item_id, user_id, status FROM reservations WHERE id = $1`
	if !dbutil.IsSQLite(s.db) {
		query += " FOR UPDATE"
	}
	if err = tx.GetContext(ctx, &reservation, query, reservationID); err != nil {
		return fmt.Errorf("reservation not found: %w", err)
	}
	if reservation.Status != "active" {
		return fmt.Errorf("reservation is not active")
	}
	nowSQL := dbutil.NowSQL(s.db)
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE reservations SET status = 'converted', updated_at = %s WHERE id = $1`, nowSQL), reservationID); err != nil {
		return fmt.Errorf("failed to convert reservation: %w", err)
	}
	var productID uuid.UUID
	if err = tx.GetContext(ctx, &productID, `SELECT product_id FROM inventory_items WHERE id = $1`, reservation.ItemID); err != nil {
		return fmt.Errorf("failed to get reserved item product: %w", err)
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE inventory_items SET status = 'SOLD', sold_at = %s, updated_at = %s WHERE id = $1`, nowSQL, nowSQL), reservation.ItemID); err != nil {
		return fmt.Errorf("failed to mark reserved item sold: %w", err)
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE inventory SET quantity = CASE WHEN COALESCE(quantity, 0) > 0 THEN quantity - 1 ELSE 0 END, reserved_quantity = CASE WHEN COALESCE(reserved_quantity, 0) > 0 THEN reserved_quantity - 1 ELSE 0 END, updated_at = %s WHERE product_id = $1`, nowSQL), productID); err != nil {
		return fmt.Errorf("failed to update reserved item aggregate: %w", err)
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`INSERT INTO inventory_movements (id, item_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at) VALUES ($1, $2, 'SALE', 1, 1, 0, 'reservation', $3, 'Reservation converted to sale', $4, %s)`, nowSQL), uuid.New(), reservation.ItemID, reservationID, reservation.UserID); err != nil {
		return fmt.Errorf("failed to record reservation sale movement: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit reservation conversion: %w", err)
	}
	return nil
}

// AdjustInventory adjusts inventory quantity (for quantity-based products) with full automation
func (s *Service) AdjustInventory(ctx context.Context, req *AdjustmentRequest, userID uuid.UUID) error {
	if req.NewQuantity < 0 {
		return fmt.Errorf("new quantity cannot be negative")
	}
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Get current item with row lock
	itemQuery := `SELECT id, product_id, part_type_id, item_code, barcode, serial_number, condition, grade, purchase_cost, selling_price, status, location_id, supplier_id, purchase_date, sold_at, notes, created_at, updated_at FROM inventory_items WHERE id = $1`
	if !dbutil.IsSQLite(s.db) {
		itemQuery += " FOR UPDATE"
	}
	item, err := getInventoryItemTx(ctx, tx, itemQuery, req.ItemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Calculate the difference from the aggregate quantity, not from the
	// individual row's implicit quantity of one.
	var currentQuantity int
	if err = tx.GetContext(ctx, &currentQuantity, `SELECT COALESCE(quantity, 0) FROM inventory WHERE product_id = $1 LIMIT 1`, item.ProductID); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to read current inventory quantity: %w", err)
	}
	quantityDiff := req.NewQuantity - currentQuantity

	// Update item status if needed
	if req.NewStatus != nil && *req.NewStatus != "" {
		if !isValidStatusTransition(item.Status, *req.NewStatus) {
			return ErrInvalidStatus
		}
		newStatus := strings.ToUpper(strings.TrimSpace(*req.NewStatus))
		updateStatusQuery := fmt.Sprintf(`UPDATE inventory_items SET status = $1, updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db))
		_, err = tx.ExecContext(ctx, updateStatusQuery, newStatus, req.ItemID)
		if err != nil {
			return fmt.Errorf("failed to update item status: %w", err)
		}
	}

	// Update aggregate inventory table
	inventoryUpdateQuery := fmt.Sprintf(`
		UPDATE inventory
		SET quantity = quantity + $1, updated_at = %s
		WHERE product_id = $2
	`, dbutil.NowSQL(s.db))
	result, err := tx.ExecContext(ctx, inventoryUpdateQuery, quantityDiff, item.ProductID)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		// Create inventory record if it doesn't exist
		createInventoryQuery := fmt.Sprintf(`
			INSERT INTO inventory (id, product_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, %s, %s)
		`, dbutil.NowSQL(s.db), dbutil.NowSQL(s.db))
		_, err = tx.ExecContext(ctx, createInventoryQuery, uuid.New(), item.ProductID, req.NewQuantity)
		if err != nil {
			return fmt.Errorf("failed to create inventory record: %w", err)
		}
	}

	// Create movement record
	adjustmentReason := req.Reason
	if adjustmentReason == nil {
		defaultReason := "Inventory adjustment"
		adjustmentReason = &defaultReason
	}
	movementQuery := `
		INSERT INTO inventory_movements (id, item_id, movement_type,
			quantity, before_quantity, after_quantity, reference_type, reference_id,
			reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(), req.ItemID, "ADJUSTMENT",
		quantityDiff, currentQuantity, req.NewQuantity, "adjustment", req.ItemID, adjustmentReason, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	// Create audit log
	auditQuery := fmt.Sprintf(`
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, %s)
	`, dbutil.NowSQL(s.db))
	var statusStr string
	if req.NewStatus != nil {
		statusStr = *req.NewStatus
	} else {
		statusStr = "unchanged"
	}
	changes := fmt.Sprintf("Adjusted item %s, new status: %s, reason: %s", req.ItemID, statusStr, *adjustmentReason)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), userID, "ADJUST_INVENTORY", "inventory_item", req.ItemID,
		changes, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	return nil
}

// TransferItem transfers an item between locations with full automation
func (s *Service) TransferItem(ctx context.Context, req *TransferRequest, userID uuid.UUID) error {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Get item with row lock
	itemQuery := `SELECT id, product_id, part_type_id, item_code, barcode, serial_number, condition, grade, purchase_cost, selling_price, status, location_id, supplier_id, purchase_date, sold_at, notes, created_at, updated_at FROM inventory_items WHERE id = $1`
	if !dbutil.IsSQLite(s.db) {
		itemQuery += " FOR UPDATE"
	}
	item, err := getInventoryItemTx(ctx, tx, itemQuery, req.ItemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Validate current location
	if item.LocationID == nil || *item.LocationID != req.FromLocationID {
		return ErrInvalidStatus
	}

	// Validate destination location exists
	var locationExists bool
	locationCheckQuery := `SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1)`
	err = tx.GetContext(ctx, &locationExists, locationCheckQuery, req.ToLocationID)
	if err != nil || !locationExists {
		return fmt.Errorf("destination location not found")
	}

	// Update location
	updateItemQuery := fmt.Sprintf(`UPDATE inventory_items SET location_id = $1, updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db))
	_, err = tx.ExecContext(ctx, updateItemQuery, req.ToLocationID, req.ItemID)
	if err != nil {
		return fmt.Errorf("failed to update item location: %w", err)
	}

	// Create movement record
	transferReason := req.Reason
	if transferReason == nil {
		defaultReason := "Item transfer"
		transferReason = &defaultReason
	}
	movementQuery := `
		INSERT INTO inventory_movements (id, item_id, movement_type,
			quantity, before_quantity, after_quantity, reference_type, reference_id,
			reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	if req.Quantity <= 0 {
		return fmt.Errorf("transfer quantity must be positive")
	}
	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(), req.ItemID, "TRANSFER",
		req.Quantity, req.Quantity, req.Quantity, "transfer", req.ItemID, transferReason, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	// Create audit log
	auditQuery := fmt.Sprintf(`
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, %s)
	`, dbutil.NowSQL(s.db))
	var reasonStr string
	if transferReason != nil {
		reasonStr = *transferReason
	} else {
		reasonStr = "no reason"
	}
	changes := fmt.Sprintf("Transferred item %s from location %s to %s, reason: %s", req.ItemID, req.FromLocationID, req.ToLocationID, reasonStr)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), userID, "TRANSFER_ITEM", "inventory_item", req.ItemID,
		changes, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	return nil
}

func getInventoryItemTx(ctx context.Context, tx *sqlx.Tx, query string, id uuid.UUID) (*InventoryItem, error) {
	record := map[string]any{}
	if err := tx.QueryRowxContext(ctx, query, id).MapScan(record); err != nil {
		return nil, err
	}
	return inventoryItemFromMap(record)
}

// ListInventoryItems lists inventory items with filters
func (s *Service) ListInventoryItems(ctx context.Context, page, perPage int, filters map[string]interface{}) ([]*InventoryItem, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.ListInventoryItems(ctx, perPage, offset, filters)
}

func (s *Service) DeleteInventoryItem(ctx context.Context, itemID, userID uuid.UUID) error {
	item, err := s.repo.GetInventoryItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	protected, err := s.repo.HasProtectedHistory(ctx, itemID)
	if err != nil {
		return err
	}
	if !protected {
		return s.repo.DeleteInventoryItem(ctx, itemID)
	}
	if strings.EqualFold(string(item.Status), string(StatusSold)) {
		return ErrCannotDeleteSoldItem
	}
	if err := s.repo.UpdateItemStatus(ctx, itemID, string(StatusArchived)); err != nil {
		return err
	}
	reason := "Inventory item removed from active inventory"
	return s.repo.CreateMovement(ctx, &InventoryMovement{
		ID: uuid.New(), ItemID: &itemID, ProductID: item.ProductID,
		MovementType: MovementAdjustment, Quantity: -1, BeforeQuantity: 1, AfterQuantity: 0,
		ReferenceType: "inventory_removal", ReferenceID: &itemID, Reason: &reason,
		CreatedBy: userID, CreatedAt: time.Now(),
	})
}

// ListInventoryItemsWithSupplierInfo lists inventory items with supplier information
func (s *Service) ListInventoryItemsWithSupplierInfo(ctx context.Context, page, perPage int, filters map[string]interface{}) ([]*InventoryItemWithSupplier, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.ListInventoryItemsWithSupplierInfo(ctx, perPage, offset, filters)
}

func (s *Service) CreateLocation(ctx context.Context, req *LocationRequest) (*Location, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Type) == "" {
		return nil, fmt.Errorf("location name and type are required")
	}
	now := time.Now().UTC()
	location := &Location{ID: uuid.New(), Name: strings.TrimSpace(req.Name), Type: strings.TrimSpace(req.Type), ParentID: req.ParentID, WarehouseID: req.WarehouseID, Description: req.Description, IsActive: true, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateLocation(ctx, location); err != nil {
		return nil, err
	}
	return location, nil
}

func (s *Service) GetLocation(ctx context.Context, id uuid.UUID) (*Location, error) {
	return s.repo.GetLocationByID(ctx, id)
}

func (s *Service) ListLocations(ctx context.Context) ([]*Location, error) {
	return s.repo.ListLocations(ctx)
}

// GetItemHistory retrieves movement history for an item
func (s *Service) GetItemHistory(ctx context.Context, itemID uuid.UUID, page, perPage int) ([]*InventoryMovement, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.GetMovementsByItem(ctx, itemID, perPage, offset)
}

// Helper functions

func isValidCondition(condition Condition) bool {
	switch condition {
	case ConditionNew, ConditionUsed, ConditionRefurbished, ConditionDamaged, ConditionForParts:
		return true
	default:
		return false
	}
}

func isValidGrade(grade Grade) bool {
	switch grade {
	case GradeExcellent, GradeVeryGood, GradeGood, GradeFair, GradePoor:
		return true
	default:
		return false
	}
}

func isValidStatusTransition(currentStatus, newStatus string) bool {
	currentStatus = strings.ToUpper(strings.TrimSpace(currentStatus))
	newStatus = strings.ToUpper(strings.TrimSpace(newStatus))

	if currentStatus == newStatus {
		return true
	}

	// Define valid status transitions
	validTransitions := map[string][]string{
		string(StatusPurchased): {string(StatusReceived), string(StatusAvailable)},
		string(StatusReceived):  {string(StatusAvailable)},
		string(StatusAvailable): {string(StatusReserved), string(StatusSold), string(StatusDamaged)},
		string(StatusReserved):  {string(StatusSold), string(StatusAvailable)},
		string(StatusSold):      {string(StatusReturned)},
		string(StatusDamaged):   {string(StatusAvailable), string(StatusInRepair), string(StatusForParts)},
		string(StatusInRepair):  {string(StatusAvailable), string(StatusForParts)},
		string(StatusReturned):  {string(StatusAvailable), string(StatusForParts)},
	}

	allowedStatuses, ok := validTransitions[currentStatus]
	if !ok {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

func generateItemCode() string {
	return fmt.Sprintf("ITEM-%d", time.Now().UnixNano())
}

func generateBarcode(productID *uuid.UUID, partTypeID *uuid.UUID) string {
	var id string
	if productID != nil {
		id = productID.String()[:8]
	} else if partTypeID != nil {
		id = partTypeID.String()[:8]
	} else {
		id = uuid.New().String()[:8]
	}
	return fmt.Sprintf("PF-%s-%s", strings.ToUpper(id), strings.ToUpper(uuid.New().String()[:8]))
}
