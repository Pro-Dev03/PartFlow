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

const (
	OpeningStockSource         = "OPENING_STOCK"
	OpeningStockModeQuantity   = "quantity"
	OpeningStockModeIndividual = "individual"
)

type OpeningStockResult struct {
	SourceType   string         `json:"source_type"`
	BusinessDate string         `json:"business_date"`
	Quantity     int            `json:"quantity"`
	Item         *InventoryItem `json:"item,omitempty"`
}

type BulkUsedStockResult struct {
	Created int              `json:"created"`
	Items   []*InventoryItem `json:"items"`
}

func (s *Service) CreateBulkUsedStock(ctx context.Context, req *BulkUsedStockRequest, userID uuid.UUID) (*BulkUsedStockResult, error) {
	if req == nil || req.ProductID == nil || req.PartTypeID == nil || len(req.Barcodes) == 0 || len(req.Barcodes) > 1000 {
		return nil, fmt.Errorf("product_id, part_type_id and 1-1000 barcodes are required")
	}
	businessDate, err := time.Parse("2006-01-02", req.BusinessDate)
	if err != nil {
		return nil, ErrInvalidBusinessDate
	}
	seen := make(map[string]struct{}, len(req.Barcodes))
	for index, barcode := range req.Barcodes {
		barcode = strings.TrimSpace(barcode)
		if barcode == "" {
			return nil, fmt.Errorf("barcode at row %d is empty", index+1)
		}
		key := strings.ToLower(barcode)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate barcode in batch: %s", barcode)
		}
		seen[key] = struct{}{}
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin bulk used stock: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	before, err := openingStockQuantity(ctx, tx, *req.ProductID)
	if err != nil {
		return nil, err
	}
	if err := updateOpeningStockQuantity(ctx, tx, *req.ProductID, len(req.Barcodes)); err != nil {
		return nil, err
	}
	result := &BulkUsedStockResult{Items: make([]*InventoryItem, 0, len(req.Barcodes))}
	condition := ConditionUsed
	for index, barcode := range req.Barcodes {
		barcodeCopy := strings.TrimSpace(barcode)
		itemReq := &OpeningStockRequest{
			ProductID: req.ProductID, Mode: OpeningStockModeIndividual, Quantity: 1,
			BusinessDate: req.BusinessDate, Barcode: &barcodeCopy, PartTypeID: req.PartTypeID,
			Condition: condition, Grade: req.Grade, PurchaseCost: req.PurchaseCost,
			SellingPrice: req.SellingPrice, Notes: req.Notes,
		}
		item, err := insertOpeningStockItem(ctx, tx, itemReq, condition, index)
		if err != nil {
			return nil, fmt.Errorf("failed at barcode %s: %w", barcodeCopy, err)
		}
		if err := insertOpeningStockMovement(ctx, tx, req.ProductID, item, before+index, 1, businessDate, userID); err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit bulk used stock: %w", err)
	}
	committed = true
	result.Created = len(result.Items)
	return result, nil
}

// CreateOpeningStock records an initial balance or item identity in one transaction.
func (s *Service) CreateOpeningStock(ctx context.Context, req *OpeningStockRequest, userID uuid.UUID) (*OpeningStockResult, error) {
	if req == nil || req.ProductID == nil {
		return nil, fmt.Errorf("product_id is required")
	}
	if req.Quantity <= 0 || req.Quantity > 10000 {
		return nil, ErrInvalidQuantity
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode != OpeningStockModeQuantity && mode != OpeningStockModeIndividual {
		return nil, ErrInvalidOpeningStockMode
	}
	businessDate, err := time.Parse("2006-01-02", req.BusinessDate)
	if err != nil {
		return nil, ErrInvalidBusinessDate
	}
	condition := req.Condition
	if condition == "" {
		condition = ConditionNew
	}
	if mode == OpeningStockModeIndividual && !isValidCondition(condition) {
		return nil, ErrInvalidCondition
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin opening stock: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	before, err := openingStockQuantity(ctx, tx, *req.ProductID)
	if err != nil {
		return nil, err
	}
	if err := updateOpeningStockQuantity(ctx, tx, *req.ProductID, req.Quantity); err != nil {
		return nil, err
	}

	result := &OpeningStockResult{SourceType: OpeningStockSource, BusinessDate: req.BusinessDate, Quantity: req.Quantity}
	if mode == OpeningStockModeQuantity {
		if err := insertOpeningStockMovement(ctx, tx, req.ProductID, nil, before, req.Quantity, businessDate, userID); err != nil {
			return nil, err
		}
	}
	for index := 0; index < req.Quantity; index++ {
		if mode != OpeningStockModeIndividual {
			break
		}
		item, itemErr := insertOpeningStockItem(ctx, tx, req, condition, index)
		if itemErr != nil {
			return nil, itemErr
		}
		if result.Item == nil {
			result.Item = item
		}
		if err := insertOpeningStockMovement(ctx, tx, req.ProductID, item, before+index, 1, businessDate, userID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit opening stock: %w", err)
	}
	committed = true
	return result, nil
}

func openingStockQuantity(ctx context.Context, tx *sqlx.Tx, productID uuid.UUID) (int, error) {
	var quantity int
	if err := tx.GetContext(ctx, &quantity, `SELECT COALESCE(quantity, 0) FROM inventory WHERE product_id = $1`, productID); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read opening stock quantity: %w", err)
	}
	return quantity, nil
}

func updateOpeningStockQuantity(ctx context.Context, tx *sqlx.Tx, productID uuid.UUID, quantity int) error {
	result, err := tx.ExecContext(ctx, `UPDATE inventory SET quantity = quantity + $1, updated_at = CURRENT_TIMESTAMP WHERE product_id = $2`, quantity, productID)
	if err != nil {
		return fmt.Errorf("failed to update opening stock quantity: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read opening stock update: %w", err)
	}
	if affected == 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, uuid.New(), productID, quantity); err != nil {
			return fmt.Errorf("failed to create opening stock quantity: %w", err)
		}
	}
	return nil
}

func insertOpeningStockItem(ctx context.Context, tx *sqlx.Tx, req *OpeningStockRequest, condition Condition, index int) (*InventoryItem, error) {
	itemCode := req.ItemCode
	if itemCode == nil || strings.TrimSpace(*itemCode) == "" || index > 0 {
		value := generateItemCode()
		itemCode = &value
	}
	barcode := req.Barcode
	if barcode == nil || strings.TrimSpace(*barcode) == "" || index > 0 {
		value := generateBarcode(req.ProductID, nil)
		barcode = &value
	}
	var existing string
	if err := tx.GetContext(ctx, &existing, `SELECT id FROM inventory_items WHERE item_code = $1`, *itemCode); err == nil {
		return nil, fmt.Errorf("duplicate opening stock item code: %s", *itemCode)
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check opening stock item code: %w", err)
	}
	if err := tx.GetContext(ctx, &existing, `SELECT id FROM inventory_items WHERE barcode = $1`, *barcode); err == nil {
		return nil, ErrDuplicateBarcode
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check opening stock barcode: %w", err)
	}
	if req.SerialNumber != nil && strings.TrimSpace(*req.SerialNumber) != "" {
		if err := tx.GetContext(ctx, &existing, `SELECT id FROM inventory_items WHERE serial_number = $1`, *req.SerialNumber); err == nil {
			return nil, ErrDuplicateSerialNumber
		} else if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check opening stock serial number: %w", err)
		}
	}
	item := &InventoryItem{
		ID: uuid.New(), ProductID: req.ProductID, PartTypeID: req.PartTypeID, ItemCode: itemCode, Barcode: barcode,
		SerialNumber: req.SerialNumber, Condition: string(condition), PurchaseCost: req.PurchaseCost,
		SellingPrice: req.SellingPrice, Status: string(StatusAvailable), LocationID: req.LocationID,
		SupplierID: req.SupplierID, CustomerID: req.CustomerID, Notes: req.Notes, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if req.Grade != nil {
		value := string(*req.Grade)
		item.Grade = &value
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO inventory_items (id, product_id, part_type_id, item_code, barcode, serial_number, condition, grade, purchase_cost, selling_price, status, location_id, supplier_id, customer_id, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`, item.ID, item.ProductID, item.PartTypeID, item.ItemCode, item.Barcode, item.SerialNumber, item.Condition, item.Grade, item.PurchaseCost, item.SellingPrice, item.Status, item.LocationID, item.SupplierID, item.CustomerID, item.Notes, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create opening stock item: %w", err)
	}
	return item, nil
}

func insertOpeningStockMovement(ctx context.Context, tx *sqlx.Tx, productID *uuid.UUID, item *InventoryItem, before, quantity int, businessDate time.Time, userID uuid.UUID) error {
	var itemID *uuid.UUID
	if item != nil {
		itemID = &item.ID
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO inventory_movements (id, item_id, product_id, movement_type, source_type, business_date, quantity, before_quantity, after_quantity, reference_type, reason, created_by, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP)`, uuid.New(), itemID, productID, MovementAdjustment, OpeningStockSource, businessDate.Format("2006-01-02"), quantity, before, before+quantity, OpeningStockSource, "Opening stock", userID)
	if err != nil {
		return fmt.Errorf("failed to create opening stock movement: %w", err)
	}
	return nil
}
