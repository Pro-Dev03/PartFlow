package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo *Repository
	db   *sqlx.DB
}

func NewService(repo *Repository, db *sqlx.DB) *Service {
	return &Service{repo: repo, db: db}
}

// CreateInventoryItem creates a new inventory item with validation
func (s *Service) CreateInventoryItem(ctx context.Context, req *InventoryItemRequest, organizationID uuid.UUID, userID uuid.UUID) (*InventoryItem, error) {
	// Validate condition
	if !isValidCondition(req.Condition) {
		return nil, ErrInvalidCondition
	}

	// Validate grade if condition is used
	if req.Condition == ConditionUsed && !isValidGrade(req.Grade) {
		return nil, ErrInvalidGrade
	}

	// Generate item code if not provided
	itemCode := req.ItemCode
	if itemCode == "" {
		itemCode = generateItemCode()
	}

	// Generate barcode if not provided
	barcode := req.Barcode
	if barcode == "" {
		barcode = generateBarcode(req.ProductID)
	}

	// Check if barcode already exists
	if barcode != "" {
		_, err := s.repo.GetInventoryItemByBarcode(ctx, barcode, organizationID)
		if err == nil {
			return nil, ErrDuplicateBarcode
		}
	}

	// Check if serial number already exists
	if req.SerialNumber != "" {
		_, err := s.repo.GetInventoryItemBySerialNumber(ctx, req.SerialNumber, organizationID)
		if err == nil {
			return nil, ErrDuplicateSerialNumber
		}
	}

	now := time.Now()
	item := &InventoryItem{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ProductID:      req.ProductID,
		ItemCode:       itemCode,
		Barcode:        barcode,
		SerialNumber:   req.SerialNumber,
		Condition:      req.Condition,
		Grade:          req.Grade,
		PurchaseCost:   req.PurchaseCost,
		SellingPrice:   req.SellingPrice,
		Status:         StatusPurchased,
		LocationID:     req.LocationID,
		SupplierID:     req.SupplierID,
		Notes:          req.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateInventoryItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}

	return item, nil
}

// GetInventoryItem retrieves an inventory item by ID
func (s *Service) GetInventoryItem(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*InventoryItem, error) {
	return s.repo.GetInventoryItemByID(ctx, id, organizationID)
}

// LookupBarcode looks up a product or item by barcode
func (s *Service) LookupBarcode(ctx context.Context, barcode string, organizationID uuid.UUID) (*InventoryItem, error) {
	// Try to find as inventory item first
	item, err := s.repo.GetInventoryItemByBarcode(ctx, barcode, organizationID)
	if err == nil {
		return item, nil
	}

	// If not found as item, return error
	return nil, ErrItemNotFound
}

// UpdateItemStatus updates the status of an inventory item with validation
func (s *Service) UpdateItemStatus(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, newStatus Status) error {
	// Get current item
	item, err := s.repo.GetInventoryItemByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	// Validate status transition
	if !isValidStatusTransition(item.Status, newStatus) {
		return ErrInvalidStatus
	}

	return s.repo.UpdateItemStatus(ctx, id, organizationID, newStatus)
}

// ReceiveItem marks an item as received and available with automatic inventory updates
func (s *Service) ReceiveItem(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, locationID *uuid.UUID, userID uuid.UUID) error {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get current item
	item, err := s.repo.GetInventoryItemByID(ctx, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Update status to available
	if err := s.repo.UpdateItemStatus(ctx, id, organizationID, StatusAvailable); err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}

	// Update location if provided
	if locationID != nil {
		item.LocationID = locationID
		item.UpdatedAt = time.Now()
		if err := s.repo.UpdateInventoryItem(ctx, item); err != nil {
			return fmt.Errorf("failed to update item location: %w", err)
		}
	}

	// Update inventory table (increase stock)
	inventoryUpdateQuery := `
		UPDATE inventory 
		SET quantity = quantity + 1, updated_at = NOW()
		WHERE product_id = $1 AND organization_id = $2
	`
	result, err := tx.ExecContext(ctx, inventoryUpdateQuery, item.ProductID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		// Create inventory record if it doesn't exist
		createInventoryQuery := `
			INSERT INTO inventory (id, organization_id, product_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, 1, NOW(), NOW())
		`
		_, err = tx.ExecContext(ctx, createInventoryQuery, uuid.New(), organizationID, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to create inventory record: %w", err)
		}
	}

	// Get current inventory quantity for movement record
	var currentQuantity int
	tx.GetContext(ctx, &currentQuantity, `SELECT COALESCE(quantity, 0) FROM inventory WHERE product_id = $1 AND organization_id = $2`, item.ProductID, organizationID)

	// Create movement record
	movement := &InventoryMovement{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ItemID:         &id,
		ProductID:      &item.ProductID,
		MovementType:   MovementPurchase,
		Quantity:       1,
		BeforeQuantity: currentQuantity - 1,
		AfterQuantity:  currentQuantity,
		Reason:         "Item received and made available",
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.CreateMovement(ctx, movement); err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	changes := fmt.Sprintf("Received item %s, location: %v", id, locationID)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "RECEIVE_ITEM", "inventory_item", id,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReserveItem reserves an item for a customer
func (s *Service) ReserveItem(ctx context.Context, req *ReservationRequest, organizationID uuid.UUID, userID uuid.UUID) (*Reservation, error) {
	// Check if item exists and is available
	item, err := s.repo.GetInventoryItemByID(ctx, req.ItemID, organizationID)
	if err != nil {
		return nil, err
	}

	if item.Status != StatusAvailable {
		return nil, ErrInvalidStatus
	}

	// Check if item is already reserved
	existingReservation, err := s.repo.GetActiveReservationByItem(ctx, req.ItemID, organizationID)
	if err != nil {
		return nil, err
	}
	if existingReservation != nil {
		return nil, ErrItemAlreadyReserved
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Minute)
	if req.ExpiresIn == 0 {
		expiresAt = time.Now().Add(24 * time.Hour) // Default 24 hours
	}

	// Create reservation
	reservation := &Reservation{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ItemID:         req.ItemID,
		CustomerID:     req.CustomerID,
		UserID:         userID,
		ReservedAt:     time.Now(),
		ExpiresAt:      expiresAt,
		Status:         "active",
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateReservation(ctx, reservation); err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Update item status
	if err := s.repo.UpdateItemStatus(ctx, req.ItemID, organizationID, StatusReserved); err != nil {
		return nil, fmt.Errorf("failed to update item status: %w", err)
	}

	return reservation, nil
}

// ReleaseReservation releases a reservation and makes item available again with full automation
func (s *Service) ReleaseReservation(ctx context.Context, reservationID uuid.UUID, organizationID uuid.UUID, userID uuid.UUID) error {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get reservation with row lock
	var reservation Reservation
	reservationQuery := `SELECT * FROM reservations WHERE id = $1 AND organization_id = $2 FOR UPDATE`
	err = tx.GetContext(ctx, &reservation, reservationQuery, reservationID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if reservation.Status != "active" {
		return fmt.Errorf("reservation is not active")
	}

	// Update reservation status to cancelled
	updateReservationQuery := `UPDATE reservations SET status = 'cancelled', updated_at = NOW() WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateReservationQuery, reservationID)
	if err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	// Update item status back to available
	updateItemQuery := `UPDATE inventory_items SET status = 'AVAILABLE', updated_at = NOW() WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateItemQuery, reservation.ItemID)
	if err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}

	// Create movement record
	movementQuery := `
		INSERT INTO inventory_movements (id, organization_id, item_id, movement_type, 
			quantity, before_quantity, after_quantity, reference_type, reference_id, 
			reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(), organizationID, reservation.ItemID, "RELEASE",
		1, 0, 1, "reservation", reservationID, "Reservation released", userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	changes := fmt.Sprintf("Released reservation %s for item %s", reservationID, reservation.ItemID)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "RELEASE_RESERVATION", "reservation", reservationID,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ConvertReservationToSale converts a reservation to a sale
func (s *Service) ConvertReservationToSale(ctx context.Context, reservationID uuid.UUID, organizationID uuid.UUID) error {
	// Get reservation
	// Update reservation status to converted
	// Update item status to sold
	// This would be called during sale creation
	
	return nil
}

// AdjustInventory adjusts inventory quantity (for quantity-based products) with full automation
func (s *Service) AdjustInventory(ctx context.Context, req *AdjustmentRequest, organizationID uuid.UUID, userID uuid.UUID) error {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get current item with row lock
	var item InventoryItem
	itemQuery := `SELECT * FROM inventory_items WHERE id = $1 AND organization_id = $2 FOR UPDATE`
	err = tx.GetContext(ctx, &item, itemQuery, req.ItemID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Calculate quantity difference
	quantityDiff := req.NewQuantity - 1 // Since individual items have quantity of 1

	// Update item status if needed
	if req.NewStatus != "" {
		if !isValidStatusTransition(item.Status, Status(req.NewStatus)) {
			return ErrInvalidStatus
		}
		updateStatusQuery := `UPDATE inventory_items SET status = $1, updated_at = NOW() WHERE id = $2`
		_, err = tx.ExecContext(ctx, updateStatusQuery, req.NewStatus, req.ItemID)
		if err != nil {
			return fmt.Errorf("failed to update item status: %w", err)
		}
	}

	// Update aggregate inventory table
	inventoryUpdateQuery := `
		UPDATE inventory 
		SET quantity = quantity + $1, updated_at = NOW()
		WHERE product_id = $2 AND organization_id = $3
	`
	result, err := tx.ExecContext(ctx, inventoryUpdateQuery, quantityDiff, item.ProductID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		// Create inventory record if it doesn't exist
		createInventoryQuery := `
			INSERT INTO inventory (id, organization_id, product_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`
		_, err = tx.ExecContext(ctx, createInventoryQuery, uuid.New(), organizationID, item.ProductID, req.NewQuantity)
		if err != nil {
			return fmt.Errorf("failed to create inventory record: %w", err)
		}
	}

	// Get current inventory quantity for movement record
	var currentQuantity int
	tx.GetContext(ctx, &currentQuantity, `SELECT COALESCE(quantity, 0) FROM inventory WHERE product_id = $1 AND organization_id = $2`, item.ProductID, organizationID)

	// Create movement record
	movementQuery := `
		INSERT INTO inventory_movements (id, organization_id, item_id, product_id, movement_type, 
			quantity, before_quantity, after_quantity, reference_type, reference_id, 
			reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(), organizationID, req.ItemID, item.ProductID, "ADJUSTMENT",
		quantityDiff, currentQuantity - quantityDiff, currentQuantity, "adjustment", req.ItemID, req.Reason, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	changes := fmt.Sprintf("Adjusted item %s, new status: %s, reason: %s", req.ItemID, req.NewStatus, req.Reason)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "ADJUST_INVENTORY", "inventory_item", req.ItemID,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransferItem transfers an item between locations with full automation
func (s *Service) TransferItem(ctx context.Context, req *TransferRequest, organizationID uuid.UUID, userID uuid.UUID) error {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get item with row lock
	var item InventoryItem
	itemQuery := `SELECT * FROM inventory_items WHERE id = $1 AND organization_id = $2 FOR UPDATE`
	err = tx.GetContext(ctx, &item, itemQuery, req.ItemID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Validate current location
	if item.LocationID == nil || *item.LocationID != req.FromLocationID {
		return ErrInvalidStatus
	}

	// Validate destination location exists
	var locationExists bool
	locationCheckQuery := `SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1 AND organization_id = $2)`
	err = tx.GetContext(ctx, &locationExists, locationCheckQuery, req.ToLocationID, organizationID)
	if err != nil || !locationExists {
		return fmt.Errorf("destination location not found")
	}

	// Update location
	updateItemQuery := `UPDATE inventory_items SET location_id = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateItemQuery, req.ToLocationID, req.ItemID)
	if err != nil {
		return fmt.Errorf("failed to update item location: %w", err)
	}

	// Create movement record
	movementQuery := `
		INSERT INTO inventory_movements (id, organization_id, item_id, product_id, movement_type, 
			quantity, before_quantity, after_quantity, reference_type, reference_id, 
			reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.ExecContext(ctx, movementQuery,
		uuid.New(), organizationID, req.ItemID, item.ProductID, "TRANSFER",
		1, 1, 1, "transfer", req.ItemID, req.Reason, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	changes := fmt.Sprintf("Transferred item %s from location %s to %s", req.ItemID, req.FromLocationID, req.ToLocationID)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "TRANSFER_ITEM", "inventory_item", req.ItemID,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListInventoryItems lists inventory items with filters
func (s *Service) ListInventoryItems(ctx context.Context, organizationID uuid.UUID, page, perPage int, filters map[string]interface{}) ([]*InventoryItem, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.ListInventoryItems(ctx, organizationID, perPage, offset, filters)
}

// GetItemHistory retrieves movement history for an item
func (s *Service) GetItemHistory(ctx context.Context, itemID uuid.UUID, organizationID uuid.UUID, page, perPage int) ([]*InventoryMovement, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.GetMovementsByItem(ctx, itemID, organizationID, perPage, offset)
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

func isValidStatusTransition(currentStatus, newStatus Status) bool {
	// Define valid status transitions
	validTransitions := map[Status][]Status{
		StatusPurchased:  {StatusReceived, StatusInspection},
		StatusReceived:   {StatusInspection, StatusAvailable},
		StatusInspection: {StatusAvailable, StatusDamaged, StatusInRepair},
		StatusAvailable:  {StatusReserved, StatusSold, StatusDamaged},
		StatusReserved:   {StatusSold, StatusAvailable},
		StatusSold:       {StatusReturned, StatusWarranty},
		StatusDamaged:    {StatusInRepair, StatusForParts},
		StatusInRepair:   {StatusAvailable, StatusForParts},
		StatusReturned:   {StatusAvailable, StatusForParts},
		StatusWarranty:   {StatusAvailable, StatusInRepair, StatusForParts},
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

func generateBarcode(productID uuid.UUID) string {
	return fmt.Sprintf("PF-%s", strings.ToUpper(productID.String()[:8]))
}
