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

	status := strings.ToLower(strings.TrimSpace(purchase.Status))
	var supplierReturnCount int
	if err := s.db.GetContext(ctx, &supplierReturnCount, s.db.Rebind(`SELECT COUNT(*) FROM supplier_returns WHERE purchase_id=?`), purchaseID.String()); err != nil {
		return nil, fmt.Errorf("failed to check supplier return dependencies: %w", err)
	}
	if supplierReturnCount > 0 {
		return &SmartDeleteResult{Action: "blocked", Message: "لا يمكن حذف شراء مرتبط بإرجاع للمورد", CanProceed: false, Details: &SmartDeleteDetails{Reason: "يوجد إرجاع مورد مرتبط بالشراء", SuggestedAction: "اعكس إرجاع المورد أولاً"}}, nil
	}
	if status == StatusReceived || status == StatusPartiallyReceived || status == "completed" {
		dependencies, err := s.checkDependencies(ctx, purchaseID)
		if err != nil {
			return nil, fmt.Errorf("failed to check purchase dependencies: %w", err)
		}
		if dependencies.HasSales || len(dependencies.UsedItems) > 0 {
			return &SmartDeleteResult{
				Action: "blocked", Message: "لا يمكن حذف الشراء لأن بعض عناصره دخلت في عمليات لاحقة", CanProceed: false,
				Details: &SmartDeleteDetails{Reason: "توجد مبيعات أو استخدام لاحق لعناصر الشراء", UsedItems: dependencies.UsedItems, SuggestedAction: "عالج العمليات اللاحقة أولاً"},
			}, nil
		}
		if dependencies.HasReturns {
			return &SmartDeleteResult{Action: "blocked", Message: "لا يمكن حذف شراء مرتبط بمرتجعات", CanProceed: false, Details: &SmartDeleteDetails{Reason: "توجد مرتجعات مرتبطة بالشراء", SuggestedAction: "احذف أو اعكس المرتجعات أولاً"}}, nil
		}
		return s.deleteReceivedPurchase(ctx, purchaseID, userID)
	}
	return s.deleteDraftPurchase(ctx, purchaseID, userID)
}

func (s *SmartDeleteService) tableExists(ctx context.Context, tx *sqlx.Tx, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1)`
	if dbutil.IsSQLite(s.db) {
		query = `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=$1)`
	}
	if err := tx.GetContext(ctx, &exists, query, name); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *SmartDeleteService) writeDeletionAudit(ctx context.Context, tx *sqlx.Tx, purchaseID, userID uuid.UUID, invoice string, total float64) error {
	if exists, err := s.tableExists(ctx, tx, "audit_logs"); err != nil || !exists {
		return err
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM audit_logs WHERE entity_id=? AND LOWER(entity_type) IN ('purchase','purchases')`), purchaseID.String()); err != nil {
		return fmt.Errorf("remove old purchase audit logs: %w", err)
	}
	snapshot, err := json.Marshal(map[string]interface{}{"invoice_number": invoice, "total_amount": total, "deleted_at": time.Now().UTC()})
	if err != nil {
		return err
	}
	var userArg interface{} = userID.String()
	if userID == uuid.Nil {
		userArg = nil
	}
	query := `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES (?, ?, 'DELETE','purchase',?,?,CURRENT_TIMESTAMP)`
	if _, err := tx.ExecContext(ctx, tx.Rebind(query), uuid.New().String(), userArg, purchaseID.String(), string(snapshot)); err != nil {
		return fmt.Errorf("write purchase deletion audit: %w", err)
	}
	return nil
}

// deleteReceivedPurchase reverses only purchase-created stock that is still
// available (or was already reversed), removes the transaction rows, and
// writes one deletion audit entry in the same transaction.
func (s *SmartDeleteService) deleteReceivedPurchase(ctx context.Context, purchaseID uuid.UUID, userID uuid.UUID) (*SmartDeleteResult, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin received purchase deletion: %w", err)
	}
	defer tx.Rollback()
	var purchase struct {
		SupplierID uuid.UUID `db:"supplier_id"`
		Invoice    string    `db:"invoice_number"`
		Total      float64   `db:"total_amount"`
	}
	if err := tx.GetContext(ctx, &purchase, tx.Rebind(`SELECT supplier_id, COALESCE(invoice_number,''), total_amount FROM purchases WHERE id=?`), purchaseID.String()); err != nil {
		return nil, fmt.Errorf("load received purchase: %w", err)
	}
	if returnsExist, err := s.tableExists(ctx, tx, "supplier_returns"); err != nil {
		return nil, err
	} else if returnsExist {
		var count int
		if err := tx.GetContext(ctx, &count, tx.Rebind(`SELECT COUNT(*) FROM supplier_returns WHERE purchase_id=?`), purchaseID.String()); err != nil {
			return nil, err
		}
		if count > 0 {
			return &SmartDeleteResult{Action: "blocked", Message: "لا يمكن حذف شراء مرتبط بإرجاع للمورد", CanProceed: false, Details: &SmartDeleteDetails{Reason: "يوجد إرجاع مورد مرتبط بالشراء", SuggestedAction: "اعكس إرجاع المورد أولاً"}}, nil
		}
	}

	prefix := purchaseID.String()[:8]
	pattern := fmt.Sprintf("ITM-%s-%%", prefix)
	var items []struct {
		ID        string `db:"id"`
		ProductID string `db:"product_id"`
		Status    string `db:"status"`
	}
	itemQuery := `SELECT id, product_id, UPPER(TRIM(COALESCE(status,''))) AS status FROM inventory_items WHERE item_code LIKE ?`
	if !dbutil.IsSQLite(s.db) {
		itemQuery += ` FOR UPDATE`
	}
	if err := tx.SelectContext(ctx, &items, tx.Rebind(itemQuery), pattern); err != nil {
		return nil, fmt.Errorf("load purchase inventory items: %w", err)
	}
	productCounts := map[string]int{}
	for _, item := range items {
		if item.Status != "AVAILABLE" && item.Status != "REVERSED" {
			return &SmartDeleteResult{Action: "blocked", Message: "لا يمكن حذف الشراء لأن عناصره لم تعد متاحة في المخزون", CanProceed: false, Details: &SmartDeleteDetails{Reason: "إحدى القطع بيعت أو تغيرت حالتها بعد الشراء", SuggestedAction: "اعكس العملية اللاحقة أولاً"}}, nil
		}
		if item.Status == "AVAILABLE" {
			productCounts[item.ProductID]++
		}
	}
	if movementExists, err := s.tableExists(ctx, tx, "inventory_movements"); err != nil {
		return nil, err
	} else if movementExists {
		var linkedChanges int
		if err := tx.GetContext(ctx, &linkedChanges, tx.Rebind(`SELECT COUNT(*) FROM inventory_movements WHERE item_id IN (SELECT id FROM inventory_items WHERE item_code LIKE ?) AND UPPER(movement_type) NOT IN ('PURCHASE','PURCHASE_REVERSAL')`), pattern); err != nil {
			return nil, fmt.Errorf("check later stock movements: %w", err)
		}
		if linkedChanges > 0 {
			return &SmartDeleteResult{Action: "blocked", Message: "لا يمكن حذف الشراء لوجود حركات لاحقة على عناصره", CanProceed: false, Details: &SmartDeleteDetails{Reason: "توجد حركات مخزون بعد الاستلام", SuggestedAction: "اعكس الحركات اللاحقة أولاً"}}, nil
		}
	}

	if inventoryExists, err := s.tableExists(ctx, tx, "inventory"); err != nil {
		return nil, err
	} else if inventoryExists {
		for productID, count := range productCounts {
			var quantity int
			if err := tx.GetContext(ctx, &quantity, tx.Rebind(`SELECT COALESCE(quantity,0) FROM inventory WHERE product_id=?`), productID); err != nil && err != sql.ErrNoRows {
				return nil, fmt.Errorf("read aggregate inventory before purchase deletion: %w", err)
			}
			if quantity < count {
				return &SmartDeleteResult{Action: "blocked", Message: "لا يمكن عكس الشراء لأن كمية المخزون الإجمالية أقل من عناصره المتاحة", CanProceed: false, Details: &SmartDeleteDetails{Reason: "الرصيد الإجمالي لا يطابق عناصر الشراء", SuggestedAction: "راجع التسويات وحركات المخزون أولاً"}}, nil
			}
		}
	}

	for _, item := range items {
		for _, dependent := range []struct{ table, query string }{
			{"inventory_movements", `DELETE FROM inventory_movements WHERE item_id=?`},
			{"item_history", `DELETE FROM item_history WHERE inventory_item_id=?`},
			{"barcodes", `DELETE FROM barcodes WHERE inventory_item_id=?`},
			{"reservations", `DELETE FROM reservations WHERE item_id=?`},
			{"acquisition_items", `DELETE FROM acquisition_items WHERE inventory_item_id=?`},
			{"inspection_items", `DELETE FROM inspection_items WHERE item_id=?`},
			{"item_repair_costs", `DELETE FROM item_repair_costs WHERE inventory_item_id=?`},
		} {
			exists, err := s.tableExists(ctx, tx, dependent.table)
			if err != nil {
				return nil, err
			}
			if exists {
				if _, err := tx.ExecContext(ctx, tx.Rebind(dependent.query), item.ID); err != nil {
					return nil, fmt.Errorf("delete purchase item dependency %s: %w", dependent.table, err)
				}
			}
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM inventory_items WHERE id=?`), item.ID); err != nil {
			return nil, fmt.Errorf("delete purchase inventory item: %w", err)
		}
	}
	if inventoryExists, _ := s.tableExists(ctx, tx, "inventory"); inventoryExists {
		for productID, count := range productCounts {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=quantity-?, updated_at=CURRENT_TIMESTAMP WHERE product_id=?`), count, productID); err != nil {
				return nil, fmt.Errorf("reverse purchase aggregate quantity: %w", err)
			}
		}
	}

	if paymentsExist, err := s.tableExists(ctx, tx, "payments"); err != nil {
		return nil, err
	} else if paymentsExist {
		if ledgerExists, err := s.tableExists(ctx, tx, "supplier_ledger"); err != nil {
			return nil, err
		} else if ledgerExists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM supplier_ledger WHERE reference_id=? OR reference_id IN (SELECT id FROM payments WHERE purchase_id=?)`), purchaseID.String(), purchaseID.String()); err != nil {
				return nil, fmt.Errorf("remove supplier payment ledger rows: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM payments WHERE purchase_id=?`), purchaseID.String()); err != nil {
			return nil, fmt.Errorf("delete purchase payments: %w", err)
		}
	} else if ledgerExists, err := s.tableExists(ctx, tx, "supplier_ledger"); err != nil {
		return nil, err
	} else if ledgerExists {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM supplier_ledger WHERE reference_id=?`), purchaseID.String()); err != nil {
			return nil, err
		}
	}

	for _, dependent := range []struct{ table, query string }{
		{"supplier_return_items", `DELETE FROM supplier_return_items WHERE supplier_return_id IN (SELECT id FROM supplier_returns WHERE purchase_id=?)`},
		{"inventory_movements", `DELETE FROM inventory_movements WHERE reference_type='purchase' AND reference_id=?`},
		{"item_history", `DELETE FROM item_history WHERE reference_type='purchase' AND reference_id=?`},
		{"purchase_items", `DELETE FROM purchase_items WHERE purchase_id=?`},
	} {
		exists, err := s.tableExists(ctx, tx, dependent.table)
		if err != nil {
			return nil, err
		}
		if exists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(dependent.query), purchaseID.String()); err != nil {
				return nil, fmt.Errorf("delete purchase dependent %s: %w", dependent.table, err)
			}
		}
	}
	if err := s.writeDeletionAudit(ctx, tx, purchaseID, userID, purchase.Invoice, purchase.Total); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM purchases WHERE id=?`), purchaseID.String()); err != nil {
		return nil, fmt.Errorf("delete received purchase: %w", err)
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE suppliers SET current_balance=COALESCE((SELECT SUM(CASE WHEN type='debit' OR transaction_type='PURCHASE' THEN amount ELSE -amount END) FROM supplier_ledger WHERE supplier_id=?),0), updated_at=CURRENT_TIMESTAMP WHERE id=?`), purchase.SupplierID.String(), purchase.SupplierID.String()); err != nil {
		return nil, fmt.Errorf("recalculate supplier balance: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit received purchase deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("purchase_deleted")
	return &SmartDeleteResult{Action: "deleted", Message: "تم حذف الشراء وعكس أثره على المخزون والرصيد", CanProceed: true}, nil
}

// deleteDraftPurchase handles deletion of draft/pending purchases
func (s *SmartDeleteService) deleteDraftPurchase(ctx context.Context, purchaseID uuid.UUID, userID uuid.UUID) (*SmartDeleteResult, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin purchase deletion: %w", err)
	}
	defer tx.Rollback()

	var purchase struct {
		SupplierID uuid.UUID `db:"supplier_id"`
		Invoice    string    `db:"invoice_number"`
		Total      float64   `db:"total_amount"`
	}
	if err = tx.GetContext(ctx, &purchase, "SELECT supplier_id, COALESCE(invoice_number,''), total_amount FROM purchases WHERE id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to find purchase supplier: %w", err)
	}

	// Payments and historical source rows do not all cascade on purchase deletion.
	if _, err = tx.ExecContext(ctx, "DELETE FROM payments WHERE purchase_id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to delete purchase payments: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM supplier_ledger WHERE reference_id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to delete supplier ledger entries: %w", err)
	}
	for _, query := range []string{
		"DELETE FROM supplier_return_items WHERE supplier_return_id IN (SELECT id FROM supplier_returns WHERE purchase_id = $1)",
		"DELETE FROM supplier_returns WHERE purchase_id = $1",
		"DELETE FROM inventory_movements WHERE reference_type = 'purchase' AND reference_id = $1",
		"DELETE FROM item_history WHERE reference_type = 'purchase' AND reference_id = $1",
	} {
		if _, err = tx.ExecContext(ctx, query, purchaseID); err != nil {
			return nil, fmt.Errorf("failed to delete purchase dependent rows: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM purchase_items WHERE purchase_id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to delete purchase items: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM purchases WHERE id = $1", purchaseID); err != nil {
		return nil, fmt.Errorf("failed to delete purchase: %w", err)
	}
	if err := s.writeDeletionAudit(ctx, tx, purchaseID, userID, purchase.Invoice, purchase.Total); err != nil {
		return nil, err
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
		WHERE id = $1`, purchase.SupplierID); err != nil {
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
