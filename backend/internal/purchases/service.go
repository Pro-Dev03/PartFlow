package purchases

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
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
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
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

	// Calculate the supplier invoice from the tax rate captured at creation time.
	var subtotal float64
	for _, item := range req.Items {
		subtotal += float64(item.Quantity) * item.UnitCost
	}
	var taxRate float64
	if req.TaxRate != nil {
		taxRate = *req.TaxRate
	} else if err := tx.GetContext(ctx, &taxRate, `SELECT COALESCE(CAST(value AS DOUBLE PRECISION), 0) FROM settings WHERE key = 'tax_rate'`); err != nil {
		taxRate = 0
	}
	if taxRate < 0 || taxRate > 100 {
		taxRate = 0
	}
	taxAmount := subtotal * taxRate / 100
	totalAmount := subtotal + taxAmount

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
		Subtotal:             subtotal,
		TaxAmount:            taxAmount,
		TotalAmount:          totalAmount,
		PaidAmount:           0,
		Status:               StatusPending,
		Notes:                strPtr(req.Notes),
		UserID:               userIDPtr, // Pass as pointer (nil if userID is Nil)
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.repo.CreateTx(ctx, tx, purchase); err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	// Create purchase items
	var items []PurchaseItem
	for _, itemReq := range req.Items {
		item := CreatePurchaseItem(purchase.ID, itemReq)
		item.Barcode = strings.TrimSpace(itemReq.Barcode)
		if err := s.repo.CreatePurchaseItemTx(ctx, tx, item); err != nil {
			return nil, fmt.Errorf("failed to create purchase item: %w", err)
		}
		items = append(items, *item)
	}

	if err := s.recordSupplierPurchaseLedger(ctx, tx, req.SupplierID, totalAmount, purchase.ID, req.InvoiceNumber, userID); err != nil {
		return nil, err
	}
	if err := s.recordPurchaseAudit(ctx, tx, purchase, items, userID); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true
	dashboard.InvalidateDashboardCacheWithReason("purchase_created")

	// Get supplier info
	supplier, err := s.repo.GetSupplierInfo(ctx, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier info: %w", err)
	}

	return purchase.ToPurchaseResponse(items, supplier), nil
}

// recordSupplierPurchaseLedger records the purchase as a debit. Older cloud
// databases use transaction_type while newer migrations use type; the helper
// supports both schemas and keeps the write in the purchase transaction.
func (s *Service) recordSupplierPurchaseLedger(ctx context.Context, tx *sqlx.Tx, supplierID uuid.UUID, amount float64, purchaseID uuid.UUID, invoice string, userID uuid.UUID) error {
	hasType, hasTransactionType, hasReferenceType, hasCreatedBy, err := supplierLedgerSchema(ctx, tx, s.db)
	if err != nil {
		return fmt.Errorf("failed to inspect supplier ledger schema: %w", err)
	}
	var balance float64
	if hasType && hasTransactionType {
		if err := tx.GetContext(ctx, &balance, `SELECT COALESCE(SUM(CASE WHEN type = 'debit' OR transaction_type = 'PURCHASE' THEN amount ELSE -amount END), 0) FROM supplier_ledger WHERE supplier_id = $1`, supplierID); err != nil {
			return fmt.Errorf("failed to read supplier balance: %w", err)
		}
		query := `INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at)
			VALUES ($1, $2, 'debit', 'PURCHASE', $3, $4, $5, $6, $7)`
		args := []interface{}{uuid.New(), supplierID, amount, balance + amount, "Purchase: " + invoice, purchaseID, time.Now()}
		if hasReferenceType && hasCreatedBy {
			query = `INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, reference_type, reference_id, description, created_by, created_at)
				VALUES ($1, $2, 'debit', 'PURCHASE', $3, $4, 'purchase', $5, $6, $7, $8)`
			args = []interface{}{uuid.New(), supplierID, amount, balance + amount, purchaseID, "Purchase: " + invoice, nullableUUID(userID), time.Now()}
		}
		_, err = tx.ExecContext(ctx, query, args...)
	} else if hasType {
		if err := tx.GetContext(ctx, &balance, `SELECT COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END), 0) FROM supplier_ledger WHERE supplier_id = $1`, supplierID); err != nil {
			return fmt.Errorf("failed to read supplier balance: %w", err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO supplier_ledger (id, supplier_id, type, amount, balance, description, reference_id, created_at) VALUES ($1, $2, 'debit', $3, $4, $5, $6, $7)`, uuid.New(), supplierID, amount, balance+amount, "Purchase: "+invoice, purchaseID, time.Now())
	} else if hasTransactionType {
		if err := tx.GetContext(ctx, &balance, `SELECT COALESCE(SUM(CASE WHEN transaction_type IN ('PAYMENT', 'RETURN') THEN -amount ELSE amount END), 0) FROM supplier_ledger WHERE supplier_id = $1`, supplierID); err != nil {
			return fmt.Errorf("failed to read supplier balance: %w", err)
		}
		query := `INSERT INTO supplier_ledger (id, supplier_id, transaction_type, amount, balance, description, reference_id, created_at) VALUES ($1, $2, 'PURCHASE', $3, $4, $5, $6, $7)`
		args := []interface{}{uuid.New(), supplierID, amount, balance + amount, "Purchase: " + invoice, purchaseID, time.Now()}
		if hasReferenceType && hasCreatedBy {
			query = `INSERT INTO supplier_ledger (id, supplier_id, transaction_type, amount, balance, reference_type, reference_id, description, created_by, created_at) VALUES ($1, $2, 'PURCHASE', $3, $4, 'purchase', $5, $6, $7, $8)`
			args = []interface{}{uuid.New(), supplierID, amount, balance + amount, purchaseID, "Purchase: " + invoice, nullableUUID(userID), time.Now()}
		}
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		return fmt.Errorf("supplier_ledger has neither type nor transaction_type")
	}
	if err != nil {
		return fmt.Errorf("failed to update supplier ledger: %w", err)
	}
	if err := updateSupplierBalanceTx(ctx, tx, s.db, supplierID, amount); err != nil {
		return err
	}
	return nil
}

func (s *Service) recordPurchaseAudit(ctx context.Context, tx *sqlx.Tx, purchase *Purchase, items []PurchaseItem, userID uuid.UUID) error {
	changes, err := json.Marshal(map[string]interface{}{
		"purchase_id":    purchase.ID,
		"supplier_id":    purchase.SupplierID,
		"invoice_number": purchase.InvoiceNumber,
		"purchase_date":  purchase.PurchaseDate,
		"subtotal":       purchase.Subtotal,
		"tax":            purchase.TaxAmount,
		"total_amount":   purchase.TotalAmount,
		"paid_amount":    purchase.PaidAmount,
		"status":         purchase.Status,
		"items_count":    len(items),
		"items":          items,
	})
	if err != nil {
		return fmt.Errorf("failed to encode purchase audit: %w", err)
	}
	var auditUser interface{}
	if userID != uuid.Nil {
		var exists bool
		if err := tx.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, userID); err == nil && exists {
			auditUser = userID
		}
	}
	query := `INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := tx.ExecContext(ctx, query, uuid.New(), auditUser, "CREATE_PURCHASE", "purchase", purchase.ID, string(changes), time.Now()); err != nil {
		return fmt.Errorf("failed to create purchase audit log: %w", err)
	}
	return nil
}

func nullableUUID(id uuid.UUID) interface{} {
	if id == uuid.Nil {
		return nil
	}
	return id
}

func supplierLedgerSchema(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB) (hasType, hasTransactionType, hasReferenceType, hasCreatedBy bool, err error) {
	if dbutil.IsSQLite(db) {
		row := struct {
			Type            bool `db:"has_type"`
			TransactionType bool `db:"has_transaction_type"`
			ReferenceType   bool `db:"has_reference_type"`
			CreatedBy       bool `db:"has_created_by"`
		}{}
		err = tx.GetContext(ctx, &row, `SELECT
			EXISTS (SELECT 1 FROM pragma_table_info('supplier_ledger') WHERE name = 'type') AS has_type,
			EXISTS (SELECT 1 FROM pragma_table_info('supplier_ledger') WHERE name = 'transaction_type') AS has_transaction_type,
			EXISTS (SELECT 1 FROM pragma_table_info('supplier_ledger') WHERE name = 'reference_type') AS has_reference_type,
			EXISTS (SELECT 1 FROM pragma_table_info('supplier_ledger') WHERE name = 'created_by') AS has_created_by`)
		return row.Type, row.TransactionType, row.ReferenceType, row.CreatedBy, err
	}
	row := struct {
		Type            bool `db:"has_type"`
		TransactionType bool `db:"has_transaction_type"`
		ReferenceType   bool `db:"has_reference_type"`
		CreatedBy       bool `db:"has_created_by"`
	}{}
	err = tx.GetContext(ctx, &row, `SELECT
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'supplier_ledger' AND column_name = 'type') AS has_type,
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'supplier_ledger' AND column_name = 'transaction_type') AS has_transaction_type,
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'supplier_ledger' AND column_name = 'reference_type') AS has_reference_type,
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'supplier_ledger' AND column_name = 'created_by') AS has_created_by`)
	return row.Type, row.TransactionType, row.ReferenceType, row.CreatedBy, err
}

func updateSupplierBalanceTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, supplierID uuid.UUID, amount float64) error {
	var hasBalance bool
	if dbutil.IsSQLite(db) {
		if err := tx.GetContext(ctx, &hasBalance, `SELECT EXISTS (SELECT 1 FROM pragma_table_info('suppliers') WHERE name = 'current_balance')`); err != nil {
			return fmt.Errorf("failed to inspect supplier balance column: %w", err)
		}
	} else if err := tx.GetContext(ctx, &hasBalance, `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'suppliers' AND column_name = 'current_balance')`); err != nil {
		return fmt.Errorf("failed to inspect supplier balance column: %w", err)
	}
	if !hasBalance {
		return nil
	}
	query := fmt.Sprintf(`UPDATE suppliers SET current_balance = COALESCE(current_balance, 0) + $1, updated_at = %s WHERE id = $2`, dbutil.NowSQL(db))
	if _, err := tx.ExecContext(ctx, query, amount, supplierID); err != nil {
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}
	return nil
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
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin purchase update: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if purchase.Status == StatusReversed || purchase.Status == StatusCancelled {
		return nil, ErrInvalidPurchaseStatus
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

	if len(req.Items) > 0 {
		if err := s.updatePurchaseItemsAndInventory(ctx, tx, purchase, req.Items); err != nil {
			return nil, err
		}
		purchase.TotalAmount = 0
		for _, item := range req.Items {
			purchase.TotalAmount += float64(item.Quantity) * item.UnitCost
		}
	}
	purchaseDate, err := purchaseDateForStorage(purchase.PurchaseDate)
	if err != nil {
		return nil, err
	}
	updateQuery := `UPDATE purchases SET invoice_number = $1, purchase_date = $2, status = $3, notes = $4, total_amount = $5, remaining_amount = CASE WHEN $5 - paid_amount > 0 THEN $5 - paid_amount ELSE 0 END, updated_at = $6 WHERE id = $7`
	if _, err := tx.ExecContext(ctx, updateQuery, purchase.InvoiceNumber, purchaseDate, purchase.Status, purchase.Notes, purchase.TotalAmount, purchase.UpdatedAt, purchase.ID); err != nil {
		return nil, fmt.Errorf("failed to update purchase: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit purchase update: %w", err)
	}
	committed = true

	return s.GetPurchase(ctx, id)
}

// updatePurchaseItemsAndInventory keeps a received purchase and its unsold
// inventory items aligned when purchase costs or quantities are edited.
func (s *Service) updatePurchaseItemsAndInventory(ctx context.Context, tx *sqlx.Tx, purchase *Purchase, requested []PurchaseItemRequest) error {
	items, err := s.repo.GetPurchaseItems(ctx, purchase.ID)
	if err != nil {
		return fmt.Errorf("failed to load purchase items: %w", err)
	}
	prefix := purchase.ID.String()[:8]
	for _, item := range requested {
		if item.Quantity <= 0 || item.UnitCost < 0 {
			return fmt.Errorf("invalid purchase item quantity or cost")
		}
		var currentID uuid.UUID
		var currentQuantity int
		found := false
		for _, existing := range items {
			if existing.ProductID == item.ProductID {
				currentID = existing.ID
				currentQuantity = existing.Quantity
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("purchase item not found")
		}
		var usedCount int
		pattern := fmt.Sprintf("ITM-%s-%%", prefix)
		if err := tx.GetContext(ctx, &usedCount, `SELECT COUNT(*) FROM inventory_items WHERE item_code LIKE $1 AND product_id = $2 AND status <> 'AVAILABLE'`, pattern, item.ProductID); err != nil {
			return fmt.Errorf("failed to check purchase inventory usage: %w", err)
		}
		if item.Quantity < usedCount {
			return fmt.Errorf("cannot reduce quantity below used inventory count")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE purchase_items SET quantity = $1, unit_price = $2, item_total = $1 * $2 WHERE id = $3`, item.Quantity, item.UnitCost, currentID); err != nil {
			return fmt.Errorf("failed to update purchase item: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE inventory_items SET purchase_cost = $1, selling_price = $2, condition = $3, updated_at = CURRENT_TIMESTAMP WHERE item_code LIKE $4 AND product_id = $5 AND status = 'AVAILABLE'`, item.UnitCost, item.SellingPrice, strings.ToUpper(item.Condition), pattern, item.ProductID); err != nil {
			return fmt.Errorf("failed to update linked inventory: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE inventory_items SET category_id = $1, updated_at = CURRENT_TIMESTAMP WHERE product_id = $2`, item.CategoryID, item.ProductID.String()); err != nil {
			return fmt.Errorf("failed to update linked inventory category: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE products SET category_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, item.CategoryID, item.ProductID); err != nil {
			return fmt.Errorf("failed to update product category: %w", err)
		}
		if item.Quantity > currentQuantity {
			return fmt.Errorf("increasing received quantity requires receiving new inventory items")
		}
	}
	return nil
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

// DeletePurchase deletes a purchase and its dependent rows when the caller is authorized.
func (s *Service) DeletePurchase(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.deleteDraftPurchase(ctx, id)
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
	for _, query := range []string{
		`DELETE FROM payments WHERE purchase_id = $1`,
		`DELETE FROM supplier_ledger WHERE reference_id = $1`,
		`DELETE FROM supplier_return_items WHERE supplier_return_id IN (SELECT id FROM supplier_returns WHERE purchase_id = $1)`,
		`DELETE FROM supplier_returns WHERE purchase_id = $1`,
		`DELETE FROM inventory_movements WHERE reference_type = 'purchase' AND reference_id = $1`,
		`DELETE FROM item_history WHERE reference_type = 'purchase' AND reference_id = $1`,
	} {
		if _, err = tx.ExecContext(ctx, query, id); err != nil {
			return fmt.Errorf("failed to delete purchase dependent rows: %w", err)
		}
	}

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
	dashboard.InvalidateDashboardCacheWithReason("purchase_deleted")

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
	purchaseDate, err := purchaseDateForStorage(purchase.PurchaseDate)
	if err != nil {
		return nil, err
	}

	if purchase.Status == "cancelled" {
		return nil, ErrPurchaseCancelled
	}

	if purchase.Status != "received" && purchase.Status != "completed" && purchase.PaidAmount <= 0 {
		return nil, ErrPurchasePaymentRequired
	}

	// Get purchase items
	items, err := s.repo.GetPurchaseItems(ctx, purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase items: %w", err)
	}

	purchasePrefix := purchase.ID.String()[:8]
	itemPattern := fmt.Sprintf("ITM-%s-%%", purchasePrefix)

	var existingRows []struct {
		ItemCode     string         `db:"item_code"`
		ProductID    string         `db:"product_id"`
		Barcode      string         `db:"barcode"`
		SerialNumber sql.NullString `db:"serial_number"`
	}
	if err = tx.SelectContext(ctx, &existingRows, `SELECT item_code, product_id, barcode, serial_number FROM inventory_items WHERE item_code LIKE $1`, itemPattern); err != nil {
		return nil, fmt.Errorf("failed to inspect existing inventory items: %w", err)
	}

	existingItemCodes := make(map[string]struct{}, len(existingRows))
	existingItemProducts := make(map[string]string, len(existingRows))
	existingSerialNumbers := make(map[string]struct{}, len(existingRows))
	for _, row := range existingRows {
		existingItemCodes[row.ItemCode] = struct{}{}
		existingItemProducts[row.ItemCode] = row.ProductID
		if row.SerialNumber.Valid {
			trimmed := strings.TrimSpace(row.SerialNumber.String)
			if trimmed != "" {
				existingSerialNumbers[trimmed] = struct{}{}
			}
		}
	}

	// Create only missing inventory items so repeated receive requests are safe.
	nextItemNumber := 1
	for _, item := range items {
		var productSellingPrice float64
		productQuery := `SELECT selling_price FROM products WHERE id = $1`
		err = tx.GetContext(ctx, &productSellingPrice, productQuery, item.ProductID)
		if err != nil {
			productSellingPrice = item.UnitCost * 1.2
		}

		createdCount := 0
		for i := 0; i < item.Quantity; i++ {
			itemCode := ""
			itemNumber := 0
			existingItem := false
			for {
				itemNumber = nextItemNumber
				candidate := fmt.Sprintf("ITM-%s-%03d", purchasePrefix, itemNumber)
				nextItemNumber++
				if _, exists := existingItemCodes[candidate]; !exists {
					itemCode = candidate
					break
				}
				if existingItemProducts[candidate] == item.ProductID.String() {
					itemCode = candidate
					existingItem = true
					break
				}
			}
			barcode := ""
			if item.Quantity == 1 && strings.TrimSpace(item.Barcode) != "" {
				barcode = strings.TrimSpace(item.Barcode)
			} else {
				barcode = fmt.Sprintf("BC-%s-%03d", purchasePrefix, itemNumber)
			}

			if existingItem {
				continue
			}

			condition := ConditionNew
			if item.Condition == "used" {
				condition = ConditionUsed
			} else if item.Condition == "refurbished" {
				condition = ConditionRefurbished
			}

			sellingPrice := productSellingPrice
			serialNumber := strings.TrimSpace(item.SerialNumber)
			var serialNumberValue interface{}
			if serialNumber != "" {
				candidate := serialNumber
				candidateSuffix := 2
				for {
					if _, exists := existingSerialNumbers[candidate]; !exists {
						break
					}
					candidate = fmt.Sprintf("%s-%d", serialNumber, candidateSuffix)
					candidateSuffix++
				}
				serialNumber = candidate
				existingSerialNumbers[serialNumber] = struct{}{}
				serialNumberValue = serialNumber
			}

			createItemQuery := `
				INSERT INTO inventory_items (
					id, product_id, category_id, item_code, barcode, serial_number, condition, grade, purchase_cost, selling_price, status,
					supplier_id, purchase_date, notes, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			`

			now := time.Now().UTC()
			_, err = tx.ExecContext(ctx, createItemQuery,
				uuid.New(), item.ProductID, item.CategoryID, itemCode, barcode, serialNumberValue,
				condition, item.Grade, item.UnitCost, sellingPrice, "AVAILABLE",
				&purchase.SupplierID, purchaseDate, item.Notes, now, now)
			if err != nil {
				return nil, fmt.Errorf("failed to create inventory item: %w", err)
			}

			existingItemCodes[itemCode] = struct{}{}
			createdCount++
		}

		if err = s.incrementInventoryAggregate(ctx, tx, item.ProductID, createdCount); err != nil {
			return nil, err
		}
	}

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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("purchase_reversed")

	return s.GetPurchase(ctx, id)
}

func (s *Service) incrementInventoryAggregate(ctx context.Context, tx *sqlx.Tx, productID uuid.UUID, amount int) error {
	if amount <= 0 {
		return nil
	}

	var tableExists bool
	if dbutil.IsSQLite(s.db) {
		if err := tx.GetContext(ctx, &tableExists, `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'inventory')`); err != nil {
			return fmt.Errorf("failed to inspect inventory table: %w", err)
		}
	} else if err := tx.GetContext(ctx, &tableExists, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'inventory')`); err != nil {
		return fmt.Errorf("failed to inspect inventory table: %w", err)
	}
	if !tableExists {
		return nil
	}

	updatedAt := dbutil.NowSQL(s.db)
	result, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE inventory
		SET quantity = COALESCE(quantity, 0) + $1, updated_at = %s
		WHERE product_id = $2
	`, updatedAt), amount, productID)
	if err != nil {
		return fmt.Errorf("failed to update inventory quantity: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to inspect inventory update: %w", err)
	}
	if rowsAffected > 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO inventory (id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, %s, %s)
	`, updatedAt, updatedAt), uuid.New(), productID, amount); err != nil {
		return fmt.Errorf("failed to create inventory aggregate: %w", err)
	}
	return nil
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
		updateItemQuery := fmt.Sprintf(`
			UPDATE inventory_items
			SET status = 'REVERSED', updated_at = %s
			WHERE id = $1
		`, dbutil.NowSQL(s.db))
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
			-1, currentQuantity, maxInt(currentQuantity-1, 0), "purchase", purchase.ID,
			reverseReason, userID, time.Now()); err != nil {
			return nil, fmt.Errorf("failed to create reverse movement: %w", err)
		}

		inventoryUpdateQuery := fmt.Sprintf(`
			UPDATE inventory
			SET quantity = CASE WHEN quantity > 0 THEN quantity - 1 ELSE 0 END, updated_at = %s
			WHERE product_id = $1
		`, dbutil.NowSQL(s.db))
		result, execErr := tx.ExecContext(ctx, inventoryUpdateQuery, item.ProductID)
		if execErr != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", execErr)
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
			createInventoryQuery := fmt.Sprintf(`
				INSERT INTO inventory (id, product_id, quantity, created_at, updated_at)
				VALUES ($1, $2, 0, %s, %s)
			`, dbutil.NowSQL(s.db), dbutil.NowSQL(s.db))
			if _, execErr = tx.ExecContext(ctx, createInventoryQuery, uuid.New(), item.ProductID); execErr != nil {
				return nil, fmt.Errorf("failed to create inventory record: %w", execErr)
			}
		}
	}

	updatePurchaseQuery := fmt.Sprintf(`
		UPDATE purchases
		SET status = $1, updated_at = %s
		WHERE id = $2
	`, dbutil.NowSQL(s.db))
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

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
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
	if dbutil.IsSQLite(s.db) {
		purchaseDateColumn := "created_at"
		var hasPurchaseDate bool
		if schemaErr := tx.GetContext(ctx, &hasPurchaseDate, `SELECT EXISTS (SELECT 1 FROM pragma_table_info('purchases') WHERE name = 'purchase_date')`); schemaErr == nil && hasPurchaseDate {
			purchaseDateColumn = "purchase_date"
		}
		var row struct {
			SupplierID    string         `db:"supplier_id"`
			InvoiceNumber string         `db:"invoice_number"`
			PurchaseDate  string         `db:"purchase_date"`
			TotalAmount   float64        `db:"total_amount"`
			PaidAmount    float64        `db:"paid_amount"`
			Status        string         `db:"status"`
			Notes         sql.NullString `db:"notes"`
			CreatedAt     string         `db:"created_at"`
			UpdatedAt     string         `db:"updated_at"`
		}
		err = tx.GetContext(ctx, &row, fmt.Sprintf(`
			SELECT supplier_id, purchase_number AS invoice_number, COALESCE(%s, created_at, datetime('now')) AS purchase_date,
				total_amount, paid_amount, status, notes, created_at, updated_at
			FROM purchases WHERE id = $1
		`, purchaseDateColumn), id)
		if err == nil {
			purchase.ID = id
			purchase.SupplierID, err = uuid.Parse(row.SupplierID)
			if err == nil {
				purchase.PurchaseDate, err = dbutil.ParseTimestamp(row.PurchaseDate)
			}
			if err == nil {
				purchase.CreatedAt, err = dbutil.ParseTimestamp(row.CreatedAt)
			}
			if err == nil {
				purchase.UpdatedAt, err = dbutil.ParseTimestamp(row.UpdatedAt)
			}
			purchase.InvoiceNumber = row.InvoiceNumber
			purchase.TotalAmount = row.TotalAmount
			purchase.PaidAmount = row.PaidAmount
			purchase.Status = row.Status
			if row.Notes.Valid {
				purchase.Notes = &row.Notes.String
			}
		}
	} else {
		purchaseQuery := `
			SELECT id, supplier_id, invoice_number, purchase_date, total_amount, paid_amount, status, notes, user_id, created_at, updated_at
			FROM purchases WHERE id = $1 FOR UPDATE
		`
		err = tx.GetContext(ctx, &purchase, purchaseQuery, id)
	}
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
	updatePaidQuery := fmt.Sprintf(`UPDATE purchases SET paid_amount = $1, remaining_amount = CASE WHEN total_amount - $1 > 0 THEN total_amount - $1 ELSE 0 END, updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db))
	_, err = tx.ExecContext(ctx, updatePaidQuery, newPaidAmount, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update paid amount: %w", err)
	}

	// Create payment record
	paymentID := uuid.New()
	paymentQuery := `
		INSERT INTO payments (id, transaction_number, purchase_id, supplier_id, amount, payment_method, payment_status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.ExecContext(ctx, paymentQuery,
		paymentID, "PAY-"+paymentID.String()[:8], purchase.ID, purchase.SupplierID, amount,
		paymentMethod, "completed", userID, time.Now(), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update supplier ledger.  The runtime migration exposes both legacy and
	// canonical columns, while this fallback keeps older installations usable.
	hasType, hasTransactionType, hasReferenceType, hasCreatedBy, schemaErr := supplierLedgerSchema(ctx, tx, s.db)
	if schemaErr != nil {
		return nil, fmt.Errorf("failed to inspect supplier ledger schema: %w", schemaErr)
	}
	var currentBalance float64
	if hasType && hasTransactionType {
		err = tx.GetContext(ctx, &currentBalance, `SELECT COALESCE(SUM(CASE WHEN type = 'debit' OR transaction_type = 'PURCHASE' THEN amount ELSE -amount END), 0) FROM supplier_ledger WHERE supplier_id = $1`, purchase.SupplierID)
	} else if hasType {
		err = tx.GetContext(ctx, &currentBalance, `SELECT COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END), 0) FROM supplier_ledger WHERE supplier_id = $1`, purchase.SupplierID)
	} else if hasTransactionType {
		err = tx.GetContext(ctx, &currentBalance, `SELECT COALESCE(SUM(CASE WHEN transaction_type IN ('PAYMENT', 'RETURN', 'SUPPLIER_RETURN') THEN -amount ELSE amount END), 0) FROM supplier_ledger WHERE supplier_id = $1`, purchase.SupplierID)
	} else {
		return nil, fmt.Errorf("supplier_ledger has neither type nor transaction_type")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read supplier balance: %w", err)
	}

	// Calculate new balance (payment reduces debt)
	newBalance := currentBalance - amount

	ledgerQuery := `INSERT INTO supplier_ledger (id, supplier_id, type, amount, balance, description, reference_id, created_at) VALUES ($1, $2, 'credit', $3, $4, $5, $6, $7)`
	ledgerArgs := []interface{}{uuid.New(), purchase.SupplierID, amount, newBalance, "Payment for purchase " + purchase.InvoiceNumber, purchase.ID, time.Now()}
	if hasType && hasTransactionType {
		ledgerQuery = `INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at) VALUES ($1, $2, 'credit', 'PAYMENT', $3, $4, $5, $6, $7)`
		ledgerArgs = []interface{}{uuid.New(), purchase.SupplierID, amount, newBalance, "Payment for purchase " + purchase.InvoiceNumber, purchase.ID, time.Now()}
	}
	if hasTransactionType && !hasType {
		ledgerQuery = `INSERT INTO supplier_ledger (id, supplier_id, transaction_type, amount, balance, description, reference_id, created_at) VALUES ($1, $2, 'PAYMENT', $3, $4, $5, $6, $7)`
		ledgerArgs = []interface{}{uuid.New(), purchase.SupplierID, amount, newBalance, "Payment for purchase " + purchase.InvoiceNumber, purchase.ID, time.Now()}
	}
	if hasReferenceType && hasCreatedBy {
		if hasType && hasTransactionType {
			ledgerQuery = `INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, reference_type, reference_id, description, created_by, created_at) VALUES ($1, $2, 'credit', 'PAYMENT', $3, $4, 'payment', $5, $6, $7, $8)`
		} else if hasTransactionType {
			ledgerQuery = `INSERT INTO supplier_ledger (id, supplier_id, transaction_type, amount, balance, reference_type, reference_id, description, created_by, created_at) VALUES ($1, $2, 'PAYMENT', $3, $4, 'payment', $5, $6, $7, $8)`
		} else {
			ledgerQuery = `INSERT INTO supplier_ledger (id, supplier_id, type, amount, balance, reference_type, reference_id, description, created_by, created_at) VALUES ($1, $2, 'credit', $3, $4, 'payment', $5, $6, $7, $8)`
		}
		ledgerArgs = []interface{}{uuid.New(), purchase.SupplierID, amount, newBalance, purchase.ID, "Payment for purchase " + purchase.InvoiceNumber, nullableUUID(userID), time.Now()}
	}
	_, err = tx.ExecContext(ctx, ledgerQuery, ledgerArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to update supplier ledger: %w", err)
	}
	if err := updateSupplierBalanceTx(ctx, tx, s.db, purchase.SupplierID, -amount); err != nil {
		return nil, err
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
	dashboard.InvalidateDashboardCacheWithReason("supplier_payment_recorded")

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
