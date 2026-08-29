package purchases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPurchaseCannotReverse   = errors.New("purchase cannot be reversed - wrong status")
	ErrInvalidReversalReason   = errors.New("reversal reason is required")
	ErrItemsAlreadyUsed        = errors.New("cannot reverse - some items were already used in subsequent operations")
	ErrPartialReversalRequired = errors.New("cannot fully reverse - partial reversal required as some items were used")
)

// ReversalService handles purchase reversals (ARCHITECTURE-PRINCIPLES.md)
type ReversalService struct {
	db *sql.DB
}

// NewReversalService creates a new reversal service
func NewReversalService(db *sql.DB) *ReversalService {
	return &ReversalService{db: db}
}

// CanReverse checks if a purchase can be reversed
// Returns (canReverse, partialReversalNeeded, error)
func (s *ReversalService) CanReverse(ctx context.Context, purchaseID uuid.UUID) (bool, bool, error) {
	var status string
	err := s.db.QueryRowContext(ctx,
		"SELECT status FROM purchases WHERE id = $1",
		purchaseID,
	).Scan(&status)

	if err == sql.ErrNoRows {
		return false, false, ErrPurchaseNotFound
	}
	if err != nil {
		return false, false, fmt.Errorf("failed to check purchase status: %w", err)
	}

	// Only received purchases can be reversed
	// Draft purchases should be deleted, not reversed
	if status != StatusReceived && status != StatusPartiallyReceived {
		return false, false, nil
	}

	// Check if any items from this purchase were used in subsequent operations
	// (sold, transferred, damaged, etc.)
	var usedItemsCount int
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT im.item_id)
		 FROM inventory_movements im
		 JOIN purchase_items pi ON im.item_id = pi.id
		 WHERE pi.purchase_id = $1
		 AND im.movement_type IN ('SALE', 'TRANSFER', 'DAMAGE', 'REPAIR')
		 AND im.created_at > (
		 	SELECT MAX(created_at) FROM inventory_movements
		 	WHERE movement_type = 'PURCHASE' AND reference_id = $1
		 )`,
		purchaseID,
	).Scan(&usedItemsCount)

	if err != nil {
		return false, false, fmt.Errorf("failed to check item usage: %w", err)
	}

	// If items were used, partial reversal is needed
	partialReversalNeeded := usedItemsCount > 0
	canReverse := true

	return canReverse, partialReversalNeeded, nil
}

// ReversePurchase reverses a purchase (ARCHITECTURE-PRINCIPLES.md)
// This creates a reversal record and updates the purchase status
// Returns error if items were already used in subsequent operations
func (s *ReversalService) ReversePurchase(ctx context.Context, purchaseID uuid.UUID, req PurchaseReversalRequest, userID uuid.UUID) (*PurchaseReversal, error) {
	// Validate request
	if req.Reason == "" {
		return nil, ErrInvalidReversalReason
	}

	// Check if purchase can be reversed and if partial reversal is needed
	canReverse, partialReversalNeeded, err := s.CanReverse(ctx, purchaseID)
	if err != nil {
		return nil, err
	}
	if !canReverse {
		return nil, ErrPurchaseCannotReverse
	}
	if partialReversalNeeded {
		return nil, ErrPartialReversalRequired
	}

	// Get purchase details
	var purchase Purchase
	err = s.db.QueryRowContext(ctx,
		`SELECT id, supplier_id, invoice_number, total_amount, status
		 FROM purchases WHERE id = $1`,
		purchaseID,
	).Scan(&purchase.ID, &purchase.SupplierID, &purchase.InvoiceNumber, &purchase.TotalAmount, &purchase.Status)

	if err != nil {
		return nil, fmt.Errorf("failed to get purchase: %w", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create reversal record
	reversalID := uuid.New()
	now := time.Now()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO purchase_reversals (id, purchase_id, reason, reversed_by, reversed_at, original_total, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		reversalID, purchaseID, req.Reason, userID, now, purchase.TotalAmount, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reversal record: %w", err)
	}

	// Update purchase status and add reversal tracking
	_, err = tx.ExecContext(ctx,
		`UPDATE purchases
		 SET status = $1, reversed_at = $2, reversed_by = $3, reversal_reason = $4, updated_at = $5
		 WHERE id = $6`,
		StatusReversed, now, userID, req.Reason, now, purchaseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase status: %w", err)
	}

	// Create reverse inventory movements for all items in this purchase
	// This will trigger the automatic current state update via the trigger
	_, err = tx.ExecContext(ctx,
		`INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at, is_reversed, reversed_by, reversed_at, reversal_reason)
		 SELECT
				gen_random_uuid(),
				pi.id as item_id,
				pi.product_id,
				'REVERSE_PURCHASE' as movement_type,
				-pi.quantity as quantity,
				ii.current_quantity + pi.quantity as before_quantity,
				ii.current_quantity as after_quantity,
				'purchase' as reference_type,
				$1 as reference_id,
				$2 as reason,
				$3 as created_by,
				$4 as created_at,
				TRUE as is_reversed,
				$3 as reversed_by,
				$4 as reversed_at,
				$2 as reversal_reason
		 FROM purchase_items pi
		 LEFT JOIN inventory_items ii ON pi.product_id = ii.product_id
		 WHERE pi.purchase_id = $5`,
		purchaseID, req.Reason, userID, now, purchaseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reverse inventory movements: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the reversal record
	reversal := &PurchaseReversal{
		ID:            reversalID,
		PurchaseID:    purchaseID,
		Reason:        req.Reason,
		ReversedBy:    userID,
		ReversedAt:    now,
		OriginalTotal: purchase.TotalAmount,
		CreatedAt:     now,
	}

	return reversal, nil
}

// GetUsedItemsInfo returns information about which items from a purchase were used in subsequent operations
func (s *ReversalService) GetUsedItemsInfo(ctx context.Context, purchaseID uuid.UUID) ([]UsedItemInfo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT
			pi.id as item_id,
			pi.product_id,
			pi.quantity as original_quantity,
			COALESCE(SUM(CASE WHEN im.movement_type = 'SALE' THEN im.quantity ELSE 0 END), 0) as sold_quantity,
			COALESCE(SUM(CASE WHEN im.movement_type = 'TRANSFER' THEN im.quantity ELSE 0 END), 0) as transferred_quantity,
			COALESCE(SUM(CASE WHEN im.movement_type = 'DAMAGE' THEN im.quantity ELSE 0 END), 0) as damaged_quantity,
			COALESCE(SUM(CASE WHEN im.movement_type = 'REPAIR' THEN im.quantity ELSE 0 END), 0) as repair_quantity
		 FROM purchase_items pi
		 LEFT JOIN inventory_items ii ON ii.product_id = pi.product_id
		 	AND ii.item_code LIKE 'ITM-' || LEFT(pi.purchase_id::text, 8) || '-%'
		 LEFT JOIN inventory_movements im ON im.item_id = ii.id
		 	AND im.movement_type IN ('SALE', 'TRANSFER', 'DAMAGE', 'REPAIR')
		 WHERE pi.purchase_id = $1
		 GROUP BY pi.id, pi.product_id, pi.quantity`,
		purchaseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get used items info: %w", err)
	}
	defer rows.Close()

	var usedItems []UsedItemInfo
	for rows.Next() {
		var info UsedItemInfo
		err := rows.Scan(
			&info.ItemID,
			&info.ProductID,
			&info.OriginalQuantity,
			&info.SoldQuantity,
			&info.TransferredQuantity,
			&info.DamagedQuantity,
			&info.RepairQuantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan used item info: %w", err)
		}
		usedItems = append(usedItems, info)
	}

	return usedItems, nil
}

// UsedItemInfo contains information about items that were used in subsequent operations
type UsedItemInfo struct {
	ItemID              uuid.UUID `json:"item_id"`
	ProductID           uuid.UUID `json:"product_id"`
	ProductName         string    `json:"product_name"`
	OriginalQuantity    int       `json:"original_quantity"`
	SoldQuantity        int       `json:"sold_quantity"`
	TransferredQuantity int       `json:"transferred_quantity"`
	DamagedQuantity     int       `json:"damaged_quantity"`
	RepairQuantity      int       `json:"repair_quantity"`
}

// GetReversalHistory returns the reversal history for a purchase
func (s *ReversalService) GetReversalHistory(ctx context.Context, purchaseID uuid.UUID) ([]PurchaseReversal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, purchase_id, reason, reversed_by, reversed_at, original_total, created_at
		 FROM purchase_reversals
		 WHERE purchase_id = $1
		 ORDER BY reversed_at DESC`,
		purchaseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get reversal history: %w", err)
	}
	defer rows.Close()

	var reversals []PurchaseReversal
	for rows.Next() {
		var reversal PurchaseReversal
		err := rows.Scan(
			&reversal.ID,
			&reversal.PurchaseID,
			&reversal.Reason,
			&reversal.ReversedBy,
			&reversal.ReversedAt,
			&reversal.OriginalTotal,
			&reversal.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reversal: %w", err)
		}
		reversals = append(reversals, reversal)
	}

	return reversals, nil
}
