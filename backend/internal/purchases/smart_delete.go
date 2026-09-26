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
	"github.com/partflow/smart-store/internal/payments"
	"github.com/partflow/smart-store/internal/returns"
	"github.com/partflow/smart-store/internal/sales"
	"github.com/partflow/smart-store/internal/supplierreturns"
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

func (s *SmartDeleteService) prepareLinkedPurchasePayments(ctx context.Context, purchaseID, userID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin purchase payment preflight: %w", err)
	}
	ids, err := purchasePaymentIDs(ctx, tx, s, purchaseID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("release purchase payment preflight: %w", err)
	}
	service := payments.NewService(payments.NewRepository(s.db))
	for _, id := range ids {
		if err := service.PrepareDelete(ctx, id, userID); err != nil {
			return fmt.Errorf("reconcile external purchase payment %s: %w", id, err)
		}
	}
	return nil
}

func purchasePaymentIDs(ctx context.Context, tx *sqlx.Tx, s *SmartDeleteService, purchaseID uuid.UUID) ([]uuid.UUID, error) {
	exists, err := s.tableExists(ctx, tx, "payments")
	if err != nil || !exists {
		return nil, err
	}
	hasPurchaseID, err := purchaseColumnExistsTx(ctx, tx, s.db, "payments", "purchase_id")
	if err != nil || !hasPurchaseID {
		return nil, err
	}
	hasID, err := purchaseColumnExistsTx(ctx, tx, s.db, "payments", "id")
	if err != nil || !hasID {
		return nil, ErrPurchasePaymentHistoryInconsistent
	}
	var rawIDs []string
	if err := tx.SelectContext(ctx, &rawIDs, tx.Rebind(`SELECT CAST(id AS TEXT) FROM payments WHERE purchase_id=? ORDER BY id`), purchaseID.String()); err != nil {
		return nil, fmt.Errorf("list purchase payment records: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, ErrPurchasePaymentHistoryInconsistent
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *SmartDeleteService) reversePurchasePaymentsTx(ctx context.Context, tx *sqlx.Tx, purchaseID uuid.UUID) error {
	ids, err := purchasePaymentIDs(ctx, tx, s, purchaseID)
	if err != nil {
		return err
	}
	paymentRepo := payments.NewRepository(s.db)
	for _, id := range ids {
		if err := paymentRepo.DeleteInTx(ctx, tx, id); err != nil {
			return fmt.Errorf("reverse purchase payment %s: %w", id, err)
		}
		if exists, err := s.tableExists(ctx, tx, "audit_logs"); err != nil {
			return err
		} else if exists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM audit_logs WHERE entity_id=?`), id.String()); err != nil {
				return fmt.Errorf("remove deleted payment audit rows: %w", err)
			}
		}
	}
	var purchase struct {
		PaidAmount  float64 `db:"paid_amount"`
		TotalAmount float64 `db:"total_amount"`
	}
	if err := tx.GetContext(ctx, &purchase, tx.Rebind(`SELECT COALESCE(paid_amount,0) AS paid_amount, COALESCE(total_amount,0) AS total_amount FROM purchases WHERE id=?`), purchaseID.String()); err != nil {
		return fmt.Errorf("verify purchase payment reversal: %w", err)
	}
	if purchase.PaidAmount > 0.000001 {
		// Older databases can contain a denormalized paid_amount without a
		// payments row. The parent deletion removes all supplier-ledger rows
		// linked to this purchase in the same transaction, so clear this stale
		// summary instead of making a valid hard delete impossible.
		setters := []string{"paid_amount=0"}
		if hasColumn, err := purchaseColumnExistsTx(ctx, tx, s.db, "purchases", "remaining_amount"); err != nil {
			return err
		} else if hasColumn {
			setters = append(setters, "remaining_amount=COALESCE(total_amount,0)")
		}
		if hasColumn, err := purchaseColumnExistsTx(ctx, tx, s.db, "purchases", "updated_at"); err != nil {
			return err
		} else if hasColumn {
			setters = append(setters, "updated_at=CURRENT_TIMESTAMP")
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE purchases SET `+strings.Join(setters, ",")+` WHERE id=?`), purchaseID.String()); err != nil {
			return fmt.Errorf("clear legacy purchase payment summary before hard delete: %w", err)
		}
	}
	return nil
}

func (s *SmartDeleteService) reversePurchaseReturnsTx(ctx context.Context, tx *sqlx.Tx, purchaseID, userID uuid.UUID) error {
	if exists, err := s.tableExists(ctx, tx, "returns"); err != nil {
		return err
	} else if exists {
		var ids []string
		if err := tx.SelectContext(ctx, &ids, tx.Rebind(`SELECT id FROM returns WHERE purchase_id=? ORDER BY id`), purchaseID.String()); err != nil {
			return fmt.Errorf("load purchase-linked customer returns: %w", err)
		}
		repository := returns.NewRepository(s.db)
		for _, rawID := range ids {
			id, err := uuid.Parse(rawID)
			if err != nil {
				return fmt.Errorf("parse purchase-linked return %q: %w", rawID, err)
			}
			if err := repository.DeleteReturnTx(ctx, tx, id); err != nil {
				return fmt.Errorf("reverse purchase-linked customer return %s: %w", id, err)
			}
		}
	}
	if exists, err := s.tableExists(ctx, tx, "supplier_returns"); err != nil {
		return err
	} else if exists {
		var ids []string
		if err := tx.SelectContext(ctx, &ids, tx.Rebind(`SELECT id FROM supplier_returns WHERE purchase_id=? ORDER BY id`), purchaseID.String()); err != nil {
			return fmt.Errorf("load purchase-linked supplier returns: %w", err)
		}
		service := supplierreturns.NewService(s.db)
		for _, rawID := range ids {
			id, err := uuid.Parse(rawID)
			if err != nil {
				return fmt.Errorf("parse purchase-linked supplier return %q: %w", rawID, err)
			}
			if err := service.DeleteTx(ctx, tx, id, false, userID); err != nil {
				return fmt.Errorf("reverse purchase-linked supplier return %s: %w", id, err)
			}
		}
	}
	return nil
}

func (s *SmartDeleteService) prepareLinkedPurchaseSales(ctx context.Context, purchaseID, userID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin linked sale preflight: %w", err)
	}
	ids, err := purchaseLinkedSales(ctx, tx, s, purchaseID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("release linked sale preflight: %w", err)
	}
	service := sales.NewSmartDeleteService(s.db)
	for _, id := range ids {
		if err := service.PrepareDelete(ctx, id, userID); err != nil {
			return fmt.Errorf("prepare dependent sale %s: %w", id, err)
		}
	}
	return nil
}

func purchaseLinkedSales(ctx context.Context, tx *sqlx.Tx, s *SmartDeleteService, purchaseID uuid.UUID) ([]uuid.UUID, error) {
	itemTable, err := s.tableExists(ctx, tx, "inventory_items")
	if err != nil {
		return nil, err
	}
	hasItemCode := false
	if itemTable {
		hasItemCode, err = purchaseColumnExistsTx(ctx, tx, s.db, "inventory_items", "item_code")
		if err != nil {
			return nil, err
		}
	}
	pattern := fmt.Sprintf("ITM-%s-%%", purchaseID.String()[:8])
	queries := make([]string, 0, 4)
	args := make([]any, 0, 4)
	if exists, err := s.tableExists(ctx, tx, "sale_items"); err != nil {
		return nil, err
	} else if exists {
		hasSaleID, err := purchaseColumnExistsTx(ctx, tx, s.db, "sale_items", "sale_id")
		if err != nil {
			return nil, err
		}
		hasInventoryItemID, err := purchaseColumnExistsTx(ctx, tx, s.db, "sale_items", "inventory_item_id")
		if err != nil {
			return nil, err
		}
		if hasSaleID && hasInventoryItemID && itemTable && hasItemCode {
			queries = append(queries, `SELECT CAST(item.sale_id AS TEXT) FROM sale_items item JOIN inventory_items inventory_item ON inventory_item.id=item.inventory_item_id WHERE inventory_item.item_code LIKE ?`)
			args = append(args, pattern)
		}
		if hasSaleID && hasInventoryItemID {
			if purchaseItemsExist, err := s.tableExists(ctx, tx, "purchase_items"); err != nil {
				return nil, err
			} else if purchaseItemsExist {
				purchaseHasOwner, err := purchaseColumnExistsTx(ctx, tx, s.db, "purchase_items", "purchase_id")
				if err != nil {
					return nil, err
				}
				purchaseHasItem, err := purchaseColumnExistsTx(ctx, tx, s.db, "purchase_items", "inventory_item_id")
				if err != nil {
					return nil, err
				}
				if purchaseHasOwner && purchaseHasItem {
					queries = append(queries, `SELECT CAST(sale_item.sale_id AS TEXT) FROM purchase_items purchase_item JOIN sale_items sale_item ON sale_item.inventory_item_id=purchase_item.inventory_item_id WHERE purchase_item.purchase_id=? AND purchase_item.inventory_item_id IS NOT NULL`)
					args = append(args, purchaseID.String())
				}
			}
		}
	}
	if exists, err := s.tableExists(ctx, tx, "inventory_movements"); err != nil {
		return nil, err
	} else if exists {
		movementColumns := []string{"item_id", "movement_type", "reference_type", "reference_id"}
		complete := true
		for _, column := range movementColumns {
			hasColumn, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", column)
			if err != nil {
				return nil, err
			}
			complete = complete && hasColumn
		}
		if complete && itemTable && hasItemCode {
			queries = append(queries, `SELECT CAST(movement.reference_id AS TEXT) FROM inventory_movements movement JOIN inventory_items inventory_item ON inventory_item.id=movement.item_id WHERE inventory_item.item_code LIKE ? AND UPPER(movement.movement_type)='SALE' AND LOWER(COALESCE(movement.reference_type,''))='sale'`)
			args = append(args, pattern)
		}
		if complete {
			if purchaseItemsExist, err := s.tableExists(ctx, tx, "purchase_items"); err != nil {
				return nil, err
			} else if purchaseItemsExist {
				purchaseHasOwner, err := purchaseColumnExistsTx(ctx, tx, s.db, "purchase_items", "purchase_id")
				if err != nil {
					return nil, err
				}
				purchaseHasItem, err := purchaseColumnExistsTx(ctx, tx, s.db, "purchase_items", "inventory_item_id")
				if err != nil {
					return nil, err
				}
				if purchaseHasOwner && purchaseHasItem {
					queries = append(queries, `SELECT CAST(movement.reference_id AS TEXT) FROM purchase_items purchase_item JOIN inventory_movements movement ON movement.item_id=purchase_item.inventory_item_id WHERE purchase_item.purchase_id=? AND UPPER(movement.movement_type)='SALE' AND LOWER(COALESCE(movement.reference_type,''))='sale'`)
					args = append(args, purchaseID.String())
				}
			}
		}
	}
	if len(queries) == 0 {
		return nil, nil
	}
	query := strings.Join(queries, ` UNION `)
	var rawIDs []string
	if err := tx.SelectContext(ctx, &rawIDs, tx.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query dependent sales for purchase %s: %w", purchaseID, err)
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("parse dependent sale id %q: %w", raw, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// SmartDelete performs intelligent deletion based on purchase state and dependencies
func (s *SmartDeleteService) SmartDelete(ctx context.Context, purchaseID uuid.UUID, userID uuid.UUID) (*SmartDeleteResult, error) {
	if err := s.PrepareDelete(ctx, purchaseID, userID); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin purchase deletion: %w", err)
	}
	defer tx.Rollback()
	result, err := s.SmartDeleteTx(ctx, tx, purchaseID, userID)
	if err != nil || result.Action != "deleted" {
		return result, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit purchase deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("purchase_deleted")
	return result, nil
}

// PrepareDelete handles external provider effects before a caller-owned
// transaction begins.
func (s *SmartDeleteService) PrepareDelete(ctx context.Context, purchaseID, userID uuid.UUID) error {
	if err := s.prepareLinkedPurchasePayments(ctx, purchaseID, userID); err != nil {
		return fmt.Errorf("prepare linked purchase payment reversals: %w", err)
	}
	if err := s.prepareLinkedPurchaseSales(ctx, purchaseID, userID); err != nil {
		return fmt.Errorf("prepare linked sale reversals: %w", err)
	}
	return nil
}

// SmartDeleteTx reverses and hard-deletes a purchase in the caller's
// transaction. Callers must run PrepareDelete before starting the transaction.
func (s *SmartDeleteService) SmartDeleteTx(ctx context.Context, tx *sqlx.Tx, purchaseID, userID uuid.UUID) (*SmartDeleteResult, error) {
	lock := ""
	if !dbutil.IsSQLite(s.db) {
		lock = " FOR UPDATE"
	}
	var status string
	if err := tx.GetContext(ctx, &status, tx.Rebind(`SELECT LOWER(TRIM(COALESCE(status,''))) FROM purchases WHERE id=?`)+lock, purchaseID.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("purchase not found")
		}
		return nil, fmt.Errorf("load purchase for deletion: %w", err)
	}
	if status == StatusDraft || status == StatusPending {
		return s.deleteDraftPurchaseTx(ctx, tx, purchaseID, userID)
	}
	if status == StatusCancelled {
		return s.deleteDraftPurchaseTx(ctx, tx, purchaseID, userID)
	}
	// Unknown and legacy statuses still use the complete reversal workflow.
	// Status labels alone must not prevent deletion when underlying stock,
	// payment, return, and ledger records can be reversed transactionally.
	return s.deleteReceivedPurchaseTx(ctx, tx, purchaseID, userID)
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

// reversePurchaseItemStateTx reverses stock changes that are owned by an
// individual item before its source purchase is removed. Sales and returns
// have already been reversed by the caller. Transfers only change the deleted
// item's location; item adjustments change the aggregate quantity and must be
// subtracted from it before removing the item's original unit.
func (s *SmartDeleteService) reversePurchaseItemStateTx(ctx context.Context, tx *sqlx.Tx, itemID, productID, currentStatus string) (string, error) {
	status := strings.ToUpper(strings.TrimSpace(currentStatus))
	if reservationsExist, err := s.tableExists(ctx, tx, "reservations"); err != nil {
		return status, err
	} else if reservationsExist {
		hasItemID, err := purchaseColumnExistsTx(ctx, tx, s.db, "reservations", "item_id")
		if err != nil {
			return status, err
		}
		if !hasItemID {
			return status, fmt.Errorf("reservation records do not expose item_id; cannot reverse reservation effects")
		}
		hasStatus, err := purchaseColumnExistsTx(ctx, tx, s.db, "reservations", "status")
		if err != nil {
			return status, err
		}
		activeReservations := 0
		if hasStatus {
			if err := tx.GetContext(ctx, &activeReservations, tx.Rebind(`SELECT COUNT(*) FROM reservations WHERE item_id=? AND LOWER(TRIM(COALESCE(status,'')))='active'`), itemID); err != nil {
				return status, fmt.Errorf("count active item reservations: %w", err)
			}
		} else if status == "RESERVED" {
			if err := tx.GetContext(ctx, &activeReservations, tx.Rebind(`SELECT COUNT(*) FROM reservations WHERE item_id=?`), itemID); err != nil {
				return status, fmt.Errorf("count legacy item reservations: %w", err)
			}
		}
		if activeReservations > 1 {
			return status, fmt.Errorf("item has %d active reservations; expected at most one", activeReservations)
		}
		if activeReservations > 0 {
			hasReservedQuantity, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory", "reserved_quantity")
			if err != nil {
				return status, err
			}
			if hasReservedQuantity {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET reserved_quantity=CASE WHEN COALESCE(reserved_quantity,0)>=? THEN reserved_quantity-? ELSE 0 END, updated_at=CURRENT_TIMESTAMP WHERE product_id=?`), activeReservations, activeReservations, productID); err != nil {
					return status, fmt.Errorf("reverse reserved inventory quantity: %w", err)
				}
			}
			status = "AVAILABLE"
		}
		if hasStatus {
			var convertedStock int
			movementExists, err := s.tableExists(ctx, tx, "inventory_movements")
			if err != nil {
				return status, err
			}
			if movementExists {
				required := []string{"item_id", "movement_type", "reference_type", "reference_id", "quantity"}
				complete := true
				for _, column := range required {
					hasColumn, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", column)
					if err != nil {
						return status, err
					}
					complete = complete && hasColumn
				}
				if complete {
					query := `SELECT COALESCE(SUM(m.quantity),0) FROM inventory_movements m JOIN reservations r ON CAST(r.id AS TEXT)=CAST(m.reference_id AS TEXT) WHERE r.item_id=? AND LOWER(TRIM(COALESCE(r.status,'')))='converted' AND UPPER(COALESCE(m.movement_type,''))='SALE' AND LOWER(COALESCE(m.reference_type,''))='reservation'`
					if err := tx.GetContext(ctx, &convertedStock, tx.Rebind(query), itemID); err != nil {
						return status, fmt.Errorf("load converted reservation stock effect: %w", err)
					}
				}
			}
			if convertedStock > 0 {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=COALESCE(quantity,0)+?, updated_at=CURRENT_TIMESTAMP WHERE product_id=?`), convertedStock, productID); err != nil {
					return status, fmt.Errorf("reverse converted reservation stock: %w", err)
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM inventory_movements WHERE item_id=? AND UPPER(COALESCE(movement_type,''))='SALE' AND LOWER(COALESCE(reference_type,''))='reservation' AND reference_id IN (SELECT id FROM reservations WHERE item_id=? AND LOWER(TRIM(COALESCE(status,'')))='converted')`), itemID, itemID); err != nil {
					return status, fmt.Errorf("remove reversed converted-reservation movement: %w", err)
				}
				status = "AVAILABLE"
			}
		}
	}

	if movementsExist, err := s.tableExists(ctx, tx, "inventory_movements"); err != nil {
		return status, err
	} else if movementsExist {
		hasItemID, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "item_id")
		if err != nil {
			return status, err
		}
		hasMovementType, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "movement_type")
		if err != nil {
			return status, err
		}
		if hasItemID && hasMovementType {
			var adjustmentCount int
			if err := tx.GetContext(ctx, &adjustmentCount, tx.Rebind(`SELECT COUNT(*) FROM inventory_movements WHERE item_id=? AND UPPER(COALESCE(movement_type,''))='ADJUSTMENT'`), itemID); err != nil {
				return status, fmt.Errorf("count item adjustment effects: %w", err)
			}
			var adjustmentDelta int
			if adjustmentCount > 0 {
				hasQuantity, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "quantity")
				if err != nil {
					return status, err
				}
				if !hasQuantity {
					return status, fmt.Errorf("item adjustment is missing its quantity delta")
				}
				if err := tx.GetContext(ctx, &adjustmentDelta, tx.Rebind(`SELECT COALESCE(SUM(quantity),0) FROM inventory_movements WHERE item_id=? AND UPPER(COALESCE(movement_type,''))='ADJUSTMENT'`), itemID); err != nil {
					return status, fmt.Errorf("sum item adjustment effects: %w", err)
				}
			}
			var damageRepairDelta int
			var damageRepairCount int
			if err := tx.GetContext(ctx, &damageRepairCount, tx.Rebind(`SELECT COUNT(*) FROM inventory_movements WHERE item_id=? AND UPPER(COALESCE(movement_type,'')) IN ('DAMAGE','REPAIR')`), itemID); err != nil {
				return status, fmt.Errorf("count item damage/repair effects: %w", err)
			}
			if damageRepairCount > 0 {
				hasBefore, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "before_quantity")
				if err != nil {
					return status, err
				}
				hasAfter, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "after_quantity")
				if err != nil {
					return status, err
				}
				hasQuantity, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "quantity")
				if err != nil {
					return status, err
				}
				deltaExpr := `0`
				if hasBefore && hasAfter && hasQuantity {
					deltaExpr = `CASE WHEN before_quantity IS NOT NULL AND after_quantity IS NOT NULL THEN after_quantity-before_quantity ELSE COALESCE(quantity,0) END`
				} else if hasBefore && hasAfter {
					deltaExpr = `CASE WHEN before_quantity IS NOT NULL AND after_quantity IS NOT NULL THEN after_quantity-before_quantity ELSE 0 END`
				} else if hasQuantity {
					deltaExpr = `COALESCE(quantity,0)`
				}
				query := `SELECT COALESCE(SUM(` + deltaExpr + `),0) FROM inventory_movements WHERE item_id=? AND UPPER(COALESCE(movement_type,'')) IN ('DAMAGE','REPAIR')`
				if err := tx.GetContext(ctx, &damageRepairDelta, tx.Rebind(query), itemID); err != nil {
					return status, fmt.Errorf("sum damage/repair stock effects: %w", err)
				}
			}
			unknownMovementDelta, unknownMovementCount, err := s.unknownItemMovementDeltaTx(ctx, tx, itemID)
			if err != nil {
				return status, err
			}
			totalDelta := adjustmentDelta + damageRepairDelta + unknownMovementDelta
			if totalDelta != 0 || damageRepairCount > 0 || unknownMovementCount > 0 {
				var currentQuantity int
				if err := tx.GetContext(ctx, &currentQuantity, tx.Rebind(`SELECT COALESCE(quantity,0) FROM inventory WHERE product_id=?`), productID); err != nil && err != sql.ErrNoRows {
					return status, fmt.Errorf("read inventory before reversing item adjustments: %w", err)
				}
				reversedQuantity := currentQuantity - totalDelta
				if reversedQuantity < 0 {
					reversedQuantity = 0
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=?, updated_at=CURRENT_TIMESTAMP WHERE product_id=?`), reversedQuantity, productID); err != nil {
					return status, fmt.Errorf("reverse item adjustment quantity: %w", err)
				}
				status = "AVAILABLE"
			}
		}
	}
	if status == "RETURNED" {
		status = "AVAILABLE"
	}
	if status != "AVAILABLE" {
		// The parent inventory record is being physically removed. Any state
		// still non-sellable after reversing its movement history contributes no
		// available aggregate stock, so mark it reversed and delete its history.
		status = "REVERSED"
	}
	return status, nil
}

// unknownItemMovementDeltaTx reverses older movement types through their
// recorded quantity delta or before/after snapshots. This keeps hard deletion
// available when a later workflow added a movement type unknown to the current
// application, while still deriving the exact aggregate stock effect.
func (s *SmartDeleteService) unknownItemMovementDeltaTx(ctx context.Context, tx *sqlx.Tx, itemID string) (int, int, error) {
	hasQuantity, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "quantity")
	if err != nil {
		return 0, 0, err
	}
	hasBefore, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "before_quantity")
	if err != nil {
		return 0, 0, err
	}
	hasAfter, err := purchaseColumnExistsTx(ctx, tx, s.db, "inventory_movements", "after_quantity")
	if err != nil {
		return 0, 0, err
	}
	quantityExpr, beforeExpr, afterExpr := `0`, `NULL`, `NULL`
	if hasQuantity {
		quantityExpr = `COALESCE(quantity,0)`
	}
	if hasBefore {
		beforeExpr = `before_quantity`
	}
	if hasAfter {
		afterExpr = `after_quantity`
	}
	query := `SELECT ` + quantityExpr + ` AS quantity, ` + beforeExpr + ` AS before_quantity, ` + afterExpr + ` AS after_quantity
		FROM inventory_movements WHERE item_id=? AND UPPER(COALESCE(movement_type,'')) NOT IN
		('PURCHASE','PURCHASE_REVERSAL','REVERSE_PURCHASE','TRANSFER','ADJUSTMENT','RELEASE','DAMAGE','REPAIR','SALE','RETURN','SUPPLIER_RETURN')`
	var rows []struct {
		Quantity int           `db:"quantity"`
		Before   sql.NullInt64 `db:"before_quantity"`
		After    sql.NullInt64 `db:"after_quantity"`
	}
	if err := tx.SelectContext(ctx, &rows, tx.Rebind(query), itemID); err != nil {
		return 0, 0, fmt.Errorf("load unrecognized inventory movement effects: %w", err)
	}
	delta := 0
	for _, row := range rows {
		movementDelta := row.Quantity
		if row.Before.Valid && row.After.Valid {
			movementDelta = int(row.After.Int64 - row.Before.Int64)
		}
		delta += movementDelta
	}
	return delta, len(rows), nil
}

// hasRecordedPayments prevents deleting a purchase while any payment history
// or supplier payment ledger entry still refers to it. Purchase deletion does
// not reverse the corresponding cash/bank account effects, so these records
// must remain available for the normal payment reversal flow.
func (s *SmartDeleteService) hasRecordedPayments(ctx context.Context, tx *sqlx.Tx, purchaseID uuid.UUID, paidAmount float64) (bool, error) {
	if paidAmount > 0.005 {
		return true, nil
	}

	paymentsExist, err := s.tableExists(ctx, tx, "payments")
	if err != nil {
		return false, err
	}
	if paymentsExist {
		hasPurchaseID, err := purchaseColumnExistsTx(ctx, tx, s.db, "payments", "purchase_id")
		if err != nil {
			return false, err
		}
		if !hasPurchaseID {
			// The payment schema is unknown, so deletion cannot be proven safe.
			return true, nil
		}
		var paymentCount int
		if err := tx.GetContext(ctx, &paymentCount, tx.Rebind(`SELECT COUNT(*) FROM payments WHERE purchase_id=?`), purchaseID.String()); err != nil {
			return false, err
		}
		if paymentCount > 0 {
			return true, nil
		}
	}

	ledgerExists, err := s.tableExists(ctx, tx, "supplier_ledger")
	if err != nil || !ledgerExists {
		return false, err
	}
	hasReferenceID, err := purchaseColumnExistsTx(ctx, tx, s.db, "supplier_ledger", "reference_id")
	if err != nil || !hasReferenceID {
		return false, err
	}
	var paymentKinds []string
	hasTransactionType, err := purchaseColumnExistsTx(ctx, tx, s.db, "supplier_ledger", "transaction_type")
	if err != nil {
		return false, err
	}
	hasType, err := purchaseColumnExistsTx(ctx, tx, s.db, "supplier_ledger", "type")
	if err != nil {
		return false, err
	}
	hasReferenceType, err := purchaseColumnExistsTx(ctx, tx, s.db, "supplier_ledger", "reference_type")
	if err != nil {
		return false, err
	}
	if hasTransactionType {
		paymentKinds = append(paymentKinds, `UPPER(COALESCE(transaction_type,''))='PAYMENT'`)
	}
	if hasType {
		paymentKinds = append(paymentKinds, `LOWER(COALESCE(type,''))='credit'`)
	}
	if hasReferenceType {
		paymentKinds = append(paymentKinds, `LOWER(COALESCE(reference_type,''))='payment'`)
	}
	if len(paymentKinds) == 0 {
		return false, nil
	}
	var ledgerPaymentCount int
	query := `SELECT COUNT(*) FROM supplier_ledger WHERE reference_id=? AND (` + strings.Join(paymentKinds, " OR ") + `)`
	if err := tx.GetContext(ctx, &ledgerPaymentCount, tx.Rebind(query), purchaseID.String()); err != nil {
		return false, err
	}
	return ledgerPaymentCount > 0, nil
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
	result, err := s.deleteReceivedPurchaseTx(ctx, tx, purchaseID, userID)
	if err != nil || result.Action != "deleted" {
		return result, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit received purchase deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("purchase_deleted")
	return result, nil
}

func (s *SmartDeleteService) deleteReceivedPurchaseTx(ctx context.Context, tx *sqlx.Tx, purchaseID uuid.UUID, userID uuid.UUID) (*SmartDeleteResult, error) {
	guardResult, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE purchases SET status=status WHERE id=?`), purchaseID.String())
	if err != nil {
		return nil, fmt.Errorf("lock received purchase for deletion: %w", err)
	}
	if affected, err := guardResult.RowsAffected(); err != nil {
		return nil, fmt.Errorf("verify received purchase for deletion: %w", err)
	} else if affected != 1 {
		return &SmartDeleteResult{
			Action: "blocked", Message: "تغيرت حالة الشراء قبل الحذف؛ حدّث القائمة وحاول مجددًا", CanProceed: false,
			Details: &SmartDeleteDetails{Reason: "لم تعد حالة الشراء تسمح بعكسه", SuggestedAction: "حدّث البيانات قبل إعادة المحاولة"},
		}, nil
	}
	var purchase struct {
		SupplierID uuid.UUID `db:"supplier_id"`
		Invoice    string    `db:"invoice_number"`
		Total      float64   `db:"total_amount"`
		PaidAmount float64   `db:"paid_amount"`
	}
	if err := tx.GetContext(ctx, &purchase, tx.Rebind(`SELECT supplier_id, COALESCE(invoice_number,'') AS invoice_number, total_amount, COALESCE(paid_amount,0) AS paid_amount FROM purchases WHERE id=?`), purchaseID.String()); err != nil {
		return nil, fmt.Errorf("load received purchase: %w", err)
	}
	if err := s.reversePurchasePaymentsTx(ctx, tx, purchaseID); err != nil {
		return nil, err
	}
	linkedSales, err := purchaseLinkedSales(ctx, tx, s, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("load sales using purchase inventory: %w", err)
	}
	saleDeleteService := sales.NewSmartDeleteService(s.db)
	for _, saleID := range linkedSales {
		if err := saleDeleteService.DeleteInTransaction(ctx, tx, saleID, userID); err != nil {
			return nil, fmt.Errorf("reverse dependent sale %s: %w", saleID, err)
		}
	}
	if err := s.reversePurchaseReturnsTx(ctx, tx, purchaseID, userID); err != nil {
		return nil, fmt.Errorf("reverse returns linked to purchase: %w", err)
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
	for index := range items {
		item := &items[index]
		status, err := s.reversePurchaseItemStateTx(ctx, tx, item.ID, item.ProductID, item.Status)
		if err != nil {
			return nil, fmt.Errorf("reverse inventory history for purchased item %s: %w", item.ID, err)
		}
		item.Status = status
		// reversePurchaseItemStateTx normalizes the item to AVAILABLE or REVERSED.
		if item.Status == "AVAILABLE" {
			productCounts[item.ProductID]++
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
				// If legacy item rows exceed the aggregate balance, clamp the reversal at zero.
				if quantity < 0 {
					quantity = 0
				}
				productCounts[productID] = quantity
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
				column := map[string]string{"inventory_movements": "item_id", "item_history": "inventory_item_id", "barcodes": "inventory_item_id", "reservations": "item_id", "acquisition_items": "inventory_item_id", "inspection_items": "item_id", "item_repair_costs": "inventory_item_id"}[dependent.table]
				hasColumn, err := purchaseColumnExistsTx(ctx, tx, s.db, dependent.table, column)
				if err != nil {
					return nil, err
				}
				if !hasColumn {
					continue
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(dependent.query), item.ID); err != nil {
					return nil, fmt.Errorf("delete purchase item dependency %s: %w", dependent.table, err)
				}
			}
		}
		for _, dependent := range []struct{ table, column string }{
			{"inspections", "inventory_item_id"}, {"trade_ins", "inventory_item_id"},
			{"warranty_claims", "inventory_item_id"}, {"item_warranties", "inventory_item_id"},
		} {
			exists, err := s.tableExists(ctx, tx, dependent.table)
			if err != nil {
				return nil, err
			}
			if !exists {
				continue
			}
			hasColumn, err := purchaseColumnExistsTx(ctx, tx, s.db, dependent.table, dependent.column)
			if err != nil {
				return nil, err
			}
			if hasColumn {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM `+dependent.table+` WHERE `+dependent.column+`=?`), item.ID); err != nil {
					return nil, fmt.Errorf("delete purchase item dependency %s: %w", dependent.table, err)
				}
			}
		}
		if exists, err := s.tableExists(ctx, tx, "audit_logs"); err != nil {
			return nil, err
		} else if exists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM audit_logs WHERE CAST(entity_id AS TEXT)=? AND LOWER(COALESCE(entity_type,'')) IN ('inventory_item','inventory_items','item')`), item.ID); err != nil {
				return nil, fmt.Errorf("delete audit history for purchased item: %w", err)
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
		{"purchase_reversals", `DELETE FROM purchase_reversals WHERE purchase_id=?`},
		{"supplier_debts", `DELETE FROM supplier_debts WHERE reference_id=? AND LOWER(COALESCE(reference_type,'')) IN ('purchase','purchase_debt')`},
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
	return &SmartDeleteResult{Action: "deleted", Message: "تم حذف الشراء وعكس أثره على المخزون والرصيد", CanProceed: true}, nil
}

// deleteDraftPurchase handles deletion of draft/pending purchases
func (s *SmartDeleteService) deleteDraftPurchase(ctx context.Context, purchaseID uuid.UUID, userID uuid.UUID) (*SmartDeleteResult, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin purchase deletion: %w", err)
	}
	defer tx.Rollback()
	result, err := s.deleteDraftPurchaseTx(ctx, tx, purchaseID, userID)
	if err != nil || result.Action != "deleted" {
		return result, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit purchase deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("purchase_deleted")
	return result, nil
}

func (s *SmartDeleteService) deleteDraftPurchaseTx(ctx context.Context, tx *sqlx.Tx, purchaseID uuid.UUID, userID uuid.UUID) (*SmartDeleteResult, error) {
	guardResult, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE purchases SET status=status WHERE id=?`), purchaseID.String())
	if err != nil {
		return nil, fmt.Errorf("lock draft purchase for deletion: %w", err)
	}
	if affected, err := guardResult.RowsAffected(); err != nil {
		return nil, fmt.Errorf("verify draft purchase for deletion: %w", err)
	} else if affected != 1 {
		return &SmartDeleteResult{
			Action: "blocked", Message: "تغيرت حالة الشراء قبل الحذف؛ حدّث القائمة وحاول مجددًا", CanProceed: false,
			Details: &SmartDeleteDetails{Reason: "لم تعد حالة الشراء تسمح بالحذف المباشر", SuggestedAction: "حدّث البيانات قبل إعادة المحاولة"},
		}, nil
	}

	var purchase struct {
		SupplierID uuid.UUID `db:"supplier_id"`
		Invoice    string    `db:"invoice_number"`
		Total      float64   `db:"total_amount"`
		PaidAmount float64   `db:"paid_amount"`
	}
	if err = tx.GetContext(ctx, &purchase, tx.Rebind(`SELECT supplier_id, COALESCE(invoice_number,'') AS invoice_number, total_amount, COALESCE(paid_amount,0) AS paid_amount FROM purchases WHERE id=?`), purchaseID.String()); err != nil {
		return nil, fmt.Errorf("failed to find purchase supplier: %w", err)
	}
	if err := s.reversePurchasePaymentsTx(ctx, tx, purchaseID); err != nil {
		return nil, err
	}
	if err := s.reversePurchaseReturnsTx(ctx, tx, purchaseID, userID); err != nil {
		return nil, fmt.Errorf("reverse returns linked to draft purchase: %w", err)
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
