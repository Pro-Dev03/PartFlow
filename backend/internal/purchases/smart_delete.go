package purchases

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// SmartDeleteResult represents the result of a smart delete operation (PRODUCT-PHILOSOPHY.md)
type SmartDeleteResult struct {
	Action     string              `json:"action"`      // deleted, reversed, blocked
	Message    string              `json:"message"`     // User-friendly message
	CanProceed bool                `json:"can_proceed"` // Whether the operation succeeded
	Details    *SmartDeleteDetails `json:"details,omitempty"`
}

// SmartDeleteDetails provides additional information when deletion is blocked
type SmartDeleteDetails struct {
	Reason          string         `json:"reason"` // Why deletion was blocked
	UsedItems       []UsedItemInfo `json:"used_items,omitempty"`
	SuggestedAction string         `json:"suggested_action"` // What the user should do instead
}

// SmartDeleteService handles intelligent deletion based on business rules (PRODUCT-PHILOSOPHY.md)
type SmartDeleteService struct {
	db *sqlx.DB
}

// NewSmartDeleteService creates a new smart delete service
func NewSmartDeleteService(db *sqlx.DB) *SmartDeleteService {
	return &SmartDeleteService{db: db}
}

// SmartDelete performs intelligent deletion based on purchase state and dependencies
func (s *SmartDeleteService) SmartDelete(ctx context.Context, purchaseID uuid.UUID, userID uuid.UUID) (*SmartDeleteResult, error) {
	// Get purchase details
	var purchase struct {
		ID          uuid.UUID  `db:"id"`
		Status      string     `db:"status"`
		ReversedAt  *time.Time `db:"reversed_at"`
		TotalAmount float64    `db:"total_amount"`
		PaidAmount  float64    `db:"paid_amount"`
	}

	purchaseQuery := "SELECT id, status, reversed_at, total_amount, COALESCE(paid_amount, 0) AS paid_amount FROM purchases WHERE id = $1"
	if dbutil.IsSQLite(s.db) {
		purchaseQuery = "SELECT id, status, total_amount, COALESCE(paid_amount, 0) AS paid_amount FROM purchases WHERE id = $1"
	}
	err := s.db.GetContext(ctx, &purchase, purchaseQuery, purchaseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("purchase not found")
		}
		return nil, fmt.Errorf("failed to get purchase: %w", err)
	}

	// Check if already reversed
	if purchase.ReversedAt != nil {
		return &SmartDeleteResult{
			Action:     "blocked",
			Message:    "تم بالفعل إلغاء هذه العملية",
			CanProceed: false,
			Details: &SmartDeleteDetails{
				Reason:          "تم بالفعل عكس العملية",
				SuggestedAction: "لا حاجة لأي إجراء",
			},
		}, nil
	}

	// Check purchase status and dependencies
	switch purchase.Status {
	case "draft", "pending":
		// Pending purchases can be deleted, including their recorded payments.
		return s.deleteDraftPurchase(ctx, purchaseID, userID)

	case "received":
		// Check if items have been used in sales or other operations
		return s.handleReceivedPurchase(ctx, purchaseID, userID, purchase)

	case "cancelled":
		return &SmartDeleteResult{
			Action:     "blocked",
			Message:    "تم بالفعل إلغاء هذه العملية",
			CanProceed: false,
			Details: &SmartDeleteDetails{
				Reason:          "العملية ملغاة بالفعل",
				SuggestedAction: "لا حاجة لأي إجراء",
			},
		}, nil

	default:
		return &SmartDeleteResult{
			Action:     "blocked",
			Message:    "لا يمكن حذف هذه العملية",
			CanProceed: false,
			Details: &SmartDeleteDetails{
				Reason:          "حالة العملية غير معروفة",
				SuggestedAction: "تواصل مع الدعم الفني",
			},
		}, nil
	}
}

// deleteDraftPurchase handles deletion of draft/pending purchases
func (s *SmartDeleteService) deleteDraftPurchase(ctx context.Context, purchaseID uuid.UUID, userID uuid.UUID) (*SmartDeleteResult, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin purchase deletion: %w", err)
	}
	defer tx.Rollback()

	var supplierID uuid.UUID
	if err = tx.GetContext(ctx, &supplierID, "SELECT supplier_id FROM purchases WHERE id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to find purchase supplier: %w", err)
	}

	// Payments do not cascade on purchase deletion, so remove dependent records explicitly.
	if _, err = tx.ExecContext(ctx, "DELETE FROM payments WHERE purchase_id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to delete purchase payments: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM supplier_ledger WHERE reference_id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to delete supplier ledger entries: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM purchase_items WHERE purchase_id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to delete purchase items: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM purchases WHERE id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to delete purchase: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE suppliers
		SET current_balance = COALESCE((
			SELECT SUM(CASE
				WHEN type = 'debit' OR transaction_type = 'PURCHASE' THEN amount
				ELSE -amount
			END)
			FROM supplier_ledger
			WHERE supplier_id = $1
		), 0), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`, supplierID); err != nil {
		return nil, fmt.Errorf("failed to refresh supplier balance: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit purchase deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("purchase_deleted")

	return &SmartDeleteResult{
		Action:     "deleted",
		Message:    "تم حذف طلب الشراء",
		CanProceed: true,
	}, nil
}

// handleReceivedPurchase handles deletion of received purchases with dependency checking
func (s *SmartDeleteService) handleReceivedPurchase(ctx context.Context, purchaseID uuid.UUID, userID uuid.UUID, purchase interface{}) (*SmartDeleteResult, error) {
	// Check for dependent operations
	dependencies, err := s.checkDependencies(ctx, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to check dependencies: %w", err)
	}

	// If there are sales or other blocking operations, block deletion
	if len(dependencies.UsedItems) > 0 {
		return &SmartDeleteResult{
			Action:     "blocked",
			Message:    "لا يمكن حذف هذه العملية لأن بعض المنتجات تم بيعها بالفعل",
			CanProceed: false,
			Details: &SmartDeleteDetails{
				Reason:          "المنتجات المرتبطة بهذه العملية دخلت في عمليات بيع",
				UsedItems:       dependencies.UsedItems,
				SuggestedAction: "يمكنك مراجعة عمليات البيع أو معالجة الإرجاع حسب الحالة",
			},
		}, nil
	}

	// If no blocking dependencies, we can reverse the operation
	return s.reversePurchase(ctx, purchaseID, userID, dependencies)
}

// DependencyCheck represents the result of dependency checking
type DependencyCheck struct {
	HasSales    bool           `json:"has_sales"`
	HasReturns  bool           `json:"has_returns"`
	HasPayments bool           `json:"has_payments"`
	UsedItems   []UsedItemInfo `json:"used_items"`
}

// checkDependencies checks if a purchase has dependent operations
func (s *SmartDeleteService) checkDependencies(ctx context.Context, purchaseID uuid.UUID) (*DependencyCheck, error) {
	check := &DependencyCheck{}

	// Check for sales using items from this purchase
	purchasePrefixExpr := "LEFT(pi.purchase_id::text, 8)"
	if dbutil.IsSQLite(s.db) {
		purchasePrefixExpr = "substr(replace(pi.purchase_id, '-', ''), 1, 8)"
	}
	var query string
	if dbutil.IsSQLite(s.db) {
		query = fmt.Sprintf(`
			SELECT pi.id AS item_id, pi.product_id, p.name AS product_name,
				pi.quantity AS original_quantity,
				COALESCE(SUM(CASE WHEN ii.status = 'SOLD' THEN 1 ELSE 0 END), 0) AS sold_quantity,
				0 AS transferred_quantity, 0 AS damaged_quantity, 0 AS repair_quantity
			FROM purchase_items pi
			LEFT JOIN products p ON pi.product_id = p.id
			LEFT JOIN inventory_items ii ON ii.product_id = pi.product_id
				AND ii.item_code LIKE 'ITM-' || %s || '-%%'
			WHERE pi.purchase_id = $1
			GROUP BY pi.id, pi.product_id, p.name, pi.quantity
			HAVING COALESCE(SUM(CASE WHEN ii.status = 'SOLD' THEN 1 ELSE 0 END), 0) > 0
		`, purchasePrefixExpr)
	} else {
		query = fmt.Sprintf(`
		SELECT 
			pi.id as item_id,
			pi.product_id,
			p.name as product_name,
			pi.quantity as original_quantity,
			COALESCE(SUM(CASE WHEN im.movement_type = 'SALE' THEN im.quantity ELSE 0 END), 0) as sold_quantity,
			COALESCE(SUM(CASE WHEN im.movement_type = 'TRANSFER' THEN im.quantity ELSE 0 END), 0) as transferred_quantity,
			COALESCE(SUM(CASE WHEN im.movement_type = 'DAMAGE' THEN im.quantity ELSE 0 END), 0) as damaged_quantity,
			COALESCE(SUM(CASE WHEN im.movement_type = 'REPAIR' THEN im.quantity ELSE 0 END), 0) as repair_quantity
		FROM purchase_items pi
		LEFT JOIN products p ON pi.product_id = p.id
		LEFT JOIN inventory_items ii ON ii.product_id = pi.product_id
			AND ii.item_code LIKE 'ITM-' || %s || '-%%'
		LEFT JOIN inventory_movements im ON im.item_id = ii.id
			AND im.movement_type IN ('SALE', 'TRANSFER', 'DAMAGE', 'REPAIR')
		WHERE pi.purchase_id = $1
		GROUP BY pi.id, pi.product_id, p.name, pi.quantity
		HAVING COALESCE(SUM(CASE WHEN im.movement_type = 'SALE' THEN im.quantity ELSE 0 END), 0) > 0
		`, purchasePrefixExpr)
	}

	rows, err := s.db.QueryContext(ctx, query, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to check sales dependencies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item UsedItemInfo
		err := rows.Scan(
			&item.ItemID, &item.ProductID, &item.ProductName,
			&item.OriginalQuantity, &item.SoldQuantity,
			&item.TransferredQuantity, &item.DamagedQuantity, &item.RepairQuantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan used item: %w", err)
		}
		check.UsedItems = append(check.UsedItems, item)
		check.HasSales = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate used items: %w", err)
	}

	// Check for returns
	var returnCount int
	err = s.db.GetContext(ctx, &returnCount,
		"SELECT COUNT(*) FROM returns WHERE purchase_id = $1", purchaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to check returns: %w", err)
	}
	check.HasReturns = returnCount > 0

	// Check recorded purchase payments without depending on the optional ledger table.
	var paymentCount int
	err = s.db.GetContext(ctx, &paymentCount,
		"SELECT COUNT(*) FROM purchases WHERE id = $1 AND COALESCE(paid_amount, 0) > 0", purchaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to check payments: %w", err)
	}
	check.HasPayments = paymentCount > 0

	return check, nil
}

// reversePurchase reverses a received purchase without blocking dependencies
func (s *SmartDeleteService) reversePurchase(ctx context.Context, purchaseID uuid.UUID, userID uuid.UUID, dependencies *DependencyCheck) (*SmartDeleteResult, error) {
	if _, err := NewService(NewRepository(s.db), s.db).ReversePurchase(ctx, purchaseID, userID, "User requested deletion"); err != nil {
		return nil, fmt.Errorf("failed to reverse purchase: %w", err)
	}

	message := "تم إلغاء العملية وإزالة تأثيرها من المخزون"
	if dependencies.HasPayments {
		message = "تم إلغاء العملية. يرجى مراجعة حساب المورد"
	}

	return &SmartDeleteResult{
		Action:     "reversed",
		Message:    message,
		CanProceed: true,
	}, nil
}

// GetUsedItemsInfo returns detailed information about used items
func (s *SmartDeleteService) GetUsedItemsInfo(ctx context.Context, purchaseID uuid.UUID) ([]UsedItemInfo, error) {
	dependencies, err := s.checkDependencies(ctx, purchaseID)
	if err != nil {
		return nil, err
	}
	return dependencies.UsedItems, nil
}
