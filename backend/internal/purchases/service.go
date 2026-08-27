package purchases

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Import InventoryItem from inventory package
// We'll define a local struct for the fields we need
type InventoryItem struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ProductID    *uuid.UUID `json:"product_id" db:"product_id"`
	ItemCode     *string    `json:"item_code" db:"item_code"`
	Barcode      *string    `json:"barcode" db:"barcode"`
	Condition    string     `json:"condition" db:"condition"`
	Grade        *string    `json:"grade" db:"grade"`
	PurchaseCost float64    `json:"purchase_cost" db:"purchase_cost"`
	SellingPrice float64    `json:"selling_price" db:"selling_price"`
	Status       string     `json:"status" db:"status"`
}

func strPtr(s string) *string {
	return &s
}

// Condition represents the condition of an inventory item
type Condition string

const (
	ConditionNew         Condition = "NEW"
	ConditionUsed        Condition = "USED"
	ConditionRefurbished Condition = "REFURBISHED"
	ConditionDamaged     Condition = "DAMAGED"
	ConditionForParts    Condition = "FOR_PARTS"
)

// Service handles purchase business logic
type Service struct {
	repo *Repository
	db   *sqlx.DB
}

// NewService creates a new purchase service
func NewService(repo *Repository, db *sqlx.DB) *Service {
	return &Service{repo: repo, db: db}
}

// CreatePurchase creates a new purchase with items and full automation
func (s *Service) CreatePurchase(ctx context.Context, userID uuid.UUID, req *PurchaseRequest) (*PurchaseResponse, error) {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Validate request
	if err := ValidatePurchaseRequest(req); err != nil {
		return nil, err
	}

	// Check if invoice number already exists
	existing, err := s.repo.GetPurchaseByInvoiceNumber(ctx, req.InvoiceNumber)
	if err == nil && existing != nil {
		return nil, ErrPurchaseExists
	}

	// Calculate total amount
	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += float64(item.Quantity) * item.UnitCost
	}

	// Create purchase
	var userIDPtr *uuid.UUID
	if userID != uuid.Nil {
		userIDPtr = &userID
	}

	purchase := &Purchase{
		ID:                   uuid.New(),
		SupplierID:           req.SupplierID,
		InvoiceNumber:        req.InvoiceNumber,
		PurchaseDate:         req.PurchaseDate,
		ExpectedDeliveryDate: req.ExpectedDeliveryDate,
		TotalAmount:          totalAmount,
		PaidAmount:           0,
		Status:               StatusPending,
		Notes:                strPtr(req.Notes),
		UserID:               userIDPtr, // Pass as pointer (nil if userID is Nil)
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.repo.Create(ctx, purchase); err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	// Create purchase items
	var items []PurchaseItem
	for _, itemReq := range req.Items {
		item := CreatePurchaseItem(purchase.ID, itemReq)
		if err := s.repo.CreatePurchaseItem(ctx, item); err != nil {
			return nil, fmt.Errorf("failed to create purchase item: %w", err)
		}
		items = append(items, *item)
	}

	// Update supplier ledger (temporarily disabled - table doesn't exist yet)
	// TODO: Create supplier_ledger table and enable this logic
	/*
		var currentBalance float64
		balanceQuery := `
			SELECT COALESCE(SUM(amount), 0)
			FROM supplier_ledger
			WHERE supplier_id = $1
		`
		err = tx.GetContext(ctx, &currentBalance, balanceQuery, req.SupplierID)
		if err != nil {
			currentBalance = 0
		}

		// Calculate new balance (purchase increases debt)
		newBalance := currentBalance + totalAmount

		ledgerQuery := `
			INSERT INTO supplier_ledger (id, supplier_id, transaction_type, amount, balance, reference_type, reference_id, description, user_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err = tx.ExecContext(ctx, ledgerQuery,
			uuid.New(), req.SupplierID, "PURCHASE",
			totalAmount, newBalance, "purchase", purchase.ID, "Purchase: "+req.InvoiceNumber, userID, time.Now())
		if err != nil {
			return nil, fmt.Errorf("failed to update supplier ledger: %w", err)
		}
	*/

	// Create audit log (temporarily disabled)
	// TODO: Check audit_logs schema and enable if needed
	/*
		auditQuery := `
			INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		changes := fmt.Sprintf("Created purchase %s with %d items, total: %.2f", req.InvoiceNumber, len(items), totalAmount)
		_, err = tx.ExecContext(ctx, auditQuery,
			uuid.New(), userID, "CREATE_PURCHASE", "purchase", purchase.ID,
			changes, time.Now())
		if err != nil {
			fmt.Printf("Warning: failed to create audit log: %v\n", err)
		}
	*/

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get supplier info
	supplier, err := s.repo.GetSupplierInfo(ctx, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier info: %w", err)
	}

	return purchase.ToPurchaseResponse(items, supplier), nil
}

// GetPurchase retrieves a purchase by ID
func (s *Service) GetPurchase(ctx context.Context, id uuid.UUID) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetPurchaseItems(ctx, purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase items: %w", err)
	}

	supplier, err := s.repo.GetSupplierInfo(ctx, purchase.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier info: %w", err)
	}

	return purchase.ToPurchaseResponse(items, supplier), nil
}

// ListPurchases retrieves purchases with pagination and filters
func (s *Service) ListPurchases(ctx context.Context, req PurchaseListRequest) ([]PurchaseListItem, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	return s.repo.ListSummaries(ctx, req)
}

// UpdatePurchase updates a purchase with status transition validation
func (s *Service) UpdatePurchase(ctx context.Context, id uuid.UUID, req *PurchaseUpdateRequest) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.InvoiceNumber != "" {
		purchase.InvoiceNumber = req.InvoiceNumber
	}
	if !req.PurchaseDate.IsZero() {
		purchase.PurchaseDate = req.PurchaseDate
	}
	if req.Status != "" {
		// Validate status transition
		if !isValidStatusTransition(purchase.Status, req.Status) {
			return nil, ErrInvalidStatusTransition
		}
		purchase.Status = req.Status
	}
	if req.Notes != "" {
		purchase.Notes = strPtr(req.Notes)
	}

	purchase.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, purchase); err != nil {
		return nil, err
	}

	return s.GetPurchase(ctx, id)
}

// isValidStatusTransition validates purchase status transitions
func isValidStatusTransition(currentStatus, newStatus string) bool {
	// Define valid status transitions
	validTransitions := map[string][]string{
		StatusDraft:             {StatusPending, StatusCancelled},
		StatusPending:           {StatusReceived, StatusCancelled, StatusPartiallyReceived},
		StatusReceived:          {StatusReversed}, // Can only reverse after receive
		StatusPartiallyReceived: {StatusReceived, StatusReversed, StatusCancelled},
		StatusCancelled:         {}, // Cannot transition from cancelled
		StatusReversed:          {}, // Cannot transition from reversed
	}

	// Allow same status (no change)
	if currentStatus == newStatus {
		return true
	}

	// Check if transition is valid
	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// DeletePurchase deletes a purchase based on its status
// - draft/pending: full deletion
// - received: use ReversePurchase instead
// - cancelled/reversed: cannot delete
func (s *Service) DeletePurchase(ctx context.Context, id uuid.UUID) error {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Status-based deletion logic
	switch purchase.Status {
	case StatusDraft, StatusPending:
		// Allow full deletion for draft and pending purchases
		return s.deleteDraftPurchase(ctx, id)
	case StatusReceived, StatusPartiallyReceived:
		// Received purchases cannot be deleted directly - must be reversed
		return ErrCannotDeleteReceivedPurchase
	case StatusCancelled, StatusReversed:
		// Cancelled and reversed purchases cannot be deleted for audit trail
		return ErrInvalidPurchaseStatus
	default:
		return ErrInvalidPurchaseStatus
	}
}

// deleteDraftPurchase performs full deletion for draft/pending purchases
func (s *Service) deleteDraftPurchase(ctx context.Context, id uuid.UUID) error {
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

	// Delete purchase items
	deleteItemsQuery := `DELETE FROM purchase_items WHERE purchase_id = $1`
	_, err = tx.ExecContext(ctx, deleteItemsQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete purchase items: %w", err)
	}

	// Delete purchase
	deletePurchaseQuery := `DELETE FROM purchases WHERE id = $1`
	result, err := tx.ExecContext(ctx, deletePurchaseQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete purchase: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrPurchaseNotFound
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReceivePurchase marks a purchase as received and creates inventory items
func (s *Service) ReceivePurchase(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*PurchaseResponse, error) {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if purchase.Status == "cancelled" {
		return nil, ErrPurchaseCancelled
	}

	if purchase.Status == "received" {
		return nil, ErrPurchaseAlreadyReceived
	}

	// Get purchase items
	items, err := s.repo.GetPurchaseItems(ctx, purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase items: %w", err)
	}

	// Create inventory items for each purchase item
	for _, item := range items {
		// Get product selling price
		var productSellingPrice float64
		productQuery := `SELECT selling_price FROM products WHERE id = $1`
		err = tx.GetContext(ctx, &productSellingPrice, productQuery, item.ProductID)
		if err != nil {
			// Fallback to markup calculation if product not found
			productSellingPrice = item.UnitCost * 1.2
		}

		// Create inventory items based on quantity
		for i := 0; i < item.Quantity; i++ {
			itemCode := fmt.Sprintf("ITM-%s-%03d", purchase.ID.String()[:8], i+1)
			barcode := fmt.Sprintf("BC-%s-%03d", purchase.ID.String()[:8], i+1)

			// Map condition from purchase item to inventory item
			condition := ConditionNew
			if item.Condition == "used" {
				condition = ConditionUsed
			} else if item.Condition == "refurbished" {
				condition = ConditionRefurbished
			}

			// Use product selling price, fallback to markup if not available
			sellingPrice := productSellingPrice

			// Create inventory item
			createItemQuery := `
				INSERT INTO inventory_items (
					id, product_id, item_code, barcode, condition, grade, purchase_cost, selling_price, status,
					supplier_id, purchase_date, notes, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			`

			_, err = tx.ExecContext(ctx, createItemQuery,
				uuid.New(), item.ProductID, itemCode, barcode,
				condition, item.Grade, item.UnitCost, sellingPrice, "AVAILABLE",
				&purchase.SupplierID, &purchase.UpdatedAt, item.Notes, time.Now(), time.Now())
			if err != nil {
				return nil, fmt.Errorf("failed to create inventory item: %w", err)
			}
		}
	}

	// Update purchase status
	purchase.Status = "received"
	purchase.UpdatedAt = time.Now()

	updatePurchaseQuery := `
		UPDATE purchases
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err = tx.ExecContext(ctx, updatePurchaseQuery, purchase.Status, purchase.UpdatedAt, purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetPurchase(ctx, id)
}

// CancelPurchase cancels a purchase
func (s *Service) CancelPurchase(ctx context.Context, id uuid.UUID) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if purchase.Status == StatusCancelled {
		return nil, ErrPurchaseCancelled
	}

	if purchase.Status == StatusReceived {
		return nil, ErrPurchaseAlreadyReceived
	}

	purchase.Status = StatusCancelled
	purchase.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, purchase); err != nil {
		return nil, err
	}

	return s.GetPurchase(ctx, id)
}

// ReversePurchase reverses a received purchase with full audit trail and inventory reversal
// This is the academic approach: instead of deleting, we reverse the operation
func (s *Service) ReversePurchase(ctx context.Context, id uuid.UUID, userID uuid.UUID, reason string) (*PurchaseResponse, error) {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if purchase.Status == StatusReversed {
		return nil, ErrPurchaseAlreadyReversed
	}

	if purchase.Status != StatusReceived && purchase.Status != StatusPartiallyReceived {
		return nil, ErrInvalidPurchaseStatus
	}

	soldItemsCheck := `
		SELECT COUNT(*)
		FROM inventory_items
		WHERE item_code LIKE $1 AND status = 'SOLD'
	`
	purchaseIDPrefix := purchase.ID.String()[:8]
	itemCodePattern := fmt.Sprintf("ITM-%s-%%", purchaseIDPrefix)

	var soldCount int
	if err = tx.GetContext(ctx, &soldCount, soldItemsCheck, itemCodePattern); err != nil {
		return nil, fmt.Errorf("failed to check sold items: %w", err)
	}
	if soldCount > 0 {
		return nil, ErrItemsAlreadySold
	}

	getItemsQuery := `
		SELECT id, product_id, item_code, barcode, condition, grade, purchase_cost, selling_price, status
		FROM inventory_items
		WHERE item_code LIKE $1
	`
	var items []InventoryItem
	if err = tx.SelectContext(ctx, &items, getItemsQuery, itemCodePattern); err != nil {
		return nil, fmt.Errorf("failed to get inventory items: %w", err)
	}

	for _, item := range items {
		updateItemQuery := `
			UPDATE inventory_items
			SET status = 'RETURNED', updated_at = NOW()
			WHERE id = $1
		`
		if _, err = tx.ExecContext(ctx, updateItemQuery, item.ID); err != nil {
			return nil, fmt.Errorf("failed to update item status: %w", err)
		}

		currentQuantity := 0
		if item.ProductID != nil {
			inventoryQuery := `
				SELECT COALESCE(quantity, 0)
				FROM inventory
				WHERE product_id = $1
			`
			queryErr := tx.GetContext(ctx, &currentQuantity, inventoryQuery, item.ProductID)
			if queryErr != nil && queryErr != sql.ErrNoRows {
				return nil, fmt.Errorf("failed to read current inventory quantity: %w", queryErr)
			}
		}

		reverseMovementQuery := `
			INSERT INTO inventory_movements (id, item_id, product_id, movement_type,
				quantity, before_quantity, after_quantity, reference_type, reference_id,
				reason, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`

		reverseReason := fmt.Sprintf("Purchase reversal: %s", reason)
		if reverseReason == "" {
			reverseReason = "Purchase reversal"
		}

		if _, err = tx.ExecContext(ctx, reverseMovementQuery,
			uuid.New(), item.ID, item.ProductID, "PURCHASE_REVERSAL",
			-1, currentQuantity, currentQuantity-1, "purchase", purchase.ID,
			reverseReason, userID, time.Now()); err != nil {
			return nil, fmt.Errorf("failed to create reverse movement: %w", err)
		}

		inventoryUpdateQuery := `
			UPDATE inventory
			SET quantity = quantity - 1, updated_at = NOW()
			WHERE product_id = $1
		`
		result, execErr := tx.ExecContext(ctx, inventoryUpdateQuery, item.ProductID)
		if execErr != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", execErr)
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
			createInventoryQuery := `
				INSERT INTO inventory (id, product_id, quantity, created_at, updated_at)
				VALUES ($1, $2, 0, NOW(), NOW())
			`
			if _, execErr = tx.ExecContext(ctx, createInventoryQuery, uuid.New(), item.ProductID); execErr != nil {
				return nil, fmt.Errorf("failed to create inventory record: %w", execErr)
			}
		}
	}

	updatePurchaseQuery := `
		UPDATE purchases
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	if _, err = tx.ExecContext(ctx, updatePurchaseQuery, StatusReversed, purchase.ID); err != nil {
		return nil, fmt.Errorf("failed to update purchase status: %w", err)
	}

	var userExists bool
	if userID != uuid.Nil {
		if err = tx.GetContext(ctx, &userExists, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, userID); err != nil {
			userExists = false
		}
	}
	if userExists {
		auditQuery := `
			INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		changesPayload, marshalErr := json.Marshal(map[string]interface{}{
			"invoice_number": purchase.InvoiceNumber,
			"reason":         reason,
			"items_affected": len(items),
		})
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to encode audit log: %w", marshalErr)
		}
		if _, err = tx.ExecContext(ctx, auditQuery,
			uuid.New(), userID, "REVERSE_PURCHASE", "purchase", purchase.ID,
			string(changesPayload), time.Now()); err != nil {
			return nil, fmt.Errorf("failed to create audit log: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetPurchase(ctx, id)
}

// AddPayment adds a payment to a purchase with full automation
func (s *Service) AddPayment(ctx context.Context, id uuid.UUID, userID uuid.UUID, amount float64, paymentMethod string) (*PurchaseResponse, error) {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if amount <= 0 {
		return nil, ErrInvalidCost
	}

	// Get purchase with row lock
	var purchase Purchase
	purchaseQuery := `
		SELECT id, supplier_id, invoice_number, purchase_date, total_amount, paid_amount, status, notes, user_id, created_at, updated_at
		FROM purchases
		WHERE id = $1
		FOR UPDATE
	`
	err = tx.GetContext(ctx, &purchase, purchaseQuery, id)
	if err != nil {
		return nil, err
	}

	if purchase.Status == "cancelled" {
		return nil, ErrPurchaseCancelled
	}

	// Check if payment exceeds total
	if purchase.PaidAmount+amount > purchase.TotalAmount {
		return nil, ErrInvalidCost
	}

	// Update paid amount
	newPaidAmount := purchase.PaidAmount + amount
	updatePaidQuery := `UPDATE purchases SET paid_amount = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updatePaidQuery, newPaidAmount, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update paid amount: %w", err)
	}

	// Create payment record
	paymentQuery := `
		INSERT INTO payments (id, purchase_id, supplier_id, amount, payment_method, payment_status, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.ExecContext(ctx, paymentQuery,
		uuid.New(), purchase.ID, purchase.SupplierID, amount,
		paymentMethod, "completed", userID, time.Now(), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update supplier ledger
	// Get current balance
	var currentBalance float64
	balanceQuery := `
		SELECT COALESCE(SUM(amount), 0)
		FROM supplier_ledger
		WHERE supplier_id = $1
	`
	err = tx.GetContext(ctx, &currentBalance, balanceQuery, purchase.SupplierID)
	if err != nil {
		currentBalance = 0
	}

	// Calculate new balance (payment reduces debt)
	newBalance := currentBalance - amount

	ledgerQuery := `
		INSERT INTO supplier_ledger (id, supplier_id, transaction_type, amount, balance, reference_type, reference_id, description, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.ExecContext(ctx, ledgerQuery,
		uuid.New(), purchase.SupplierID, "PAYMENT",
		-amount, newBalance, "payment", purchase.ID, "Payment for purchase "+purchase.InvoiceNumber, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to update supplier ledger: %w", err)
	}

	// Create audit log
	var userExists bool
	if userID != uuid.Nil {
		if err = tx.GetContext(ctx, &userExists, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, userID); err != nil {
			userExists = false
		}
	}
	if userExists {
		auditQuery := `
			INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		changesPayload, marshalErr := json.Marshal(map[string]interface{}{
			"purchase_id":     purchase.ID,
			"invoice_number":  purchase.InvoiceNumber,
			"payment_amount":  amount,
			"new_paid_amount": newPaidAmount,
		})
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to encode audit log: %w", marshalErr)
		}
		if _, err = tx.ExecContext(ctx, auditQuery,
			uuid.New(), userID, "ADD_PAYMENT", "purchase", purchase.ID,
			string(changesPayload), time.Now()); err != nil {
			return nil, fmt.Errorf("failed to create audit log: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetPurchase(ctx, id)
}

// AddPurchaseItem adds an item to a purchase
func (s *Service) AddPurchaseItem(ctx context.Context, purchaseID uuid.UUID, req *PurchaseItemRequest) (*PurchaseItem, error) {
	purchase, err := s.repo.GetByID(ctx, purchaseID)
	if err != nil {
		return nil, err
	}

	if purchase.Status == "received" || purchase.Status == "cancelled" {
		return nil, ErrInvalidPurchaseStatus
	}

	// Validate item request
	if req.ProductID == uuid.Nil {
		return nil, ErrProductNotFound
	}
	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if req.UnitCost < 0 {
		return nil, ErrInvalidCost
	}
	if req.Condition != "new" && req.Condition != "used" && req.Condition != "refurbished" {
		return nil, ErrInvalidCondition
	}

	item := CreatePurchaseItem(purchaseID, *req)
	if err := s.repo.CreatePurchaseItem(ctx, item); err != nil {
		return nil, err
	}

	// Update purchase total amount
	purchase.TotalAmount += item.TotalCost
	purchase.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, purchase); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdatePurchaseItem updates a purchase item
func (s *Service) UpdatePurchaseItem(ctx context.Context, itemID uuid.UUID, req *PurchaseItemRequest) (*PurchaseItem, error) {
	// Get the item first
	var item PurchaseItem
	// This would require a more complex query to join with purchases table
	// For now, we'll update directly

	totalCost := float64(req.Quantity) * req.UnitCost

	item.ID = itemID
	item.Quantity = req.Quantity
	item.UnitCost = req.UnitCost
	item.TotalCost = totalCost
	item.SerialNumber = req.SerialNumber
	item.Condition = req.Condition
	item.LocationID = req.LocationID
	item.Notes = req.Notes
	item.UpdatedAt = time.Now()

	if err := s.repo.UpdatePurchaseItem(ctx, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

// DeletePurchaseItem deletes a purchase item
func (s *Service) DeletePurchaseItem(ctx context.Context, itemID uuid.UUID) error {
	return s.repo.DeletePurchaseItem(ctx, itemID)
}
