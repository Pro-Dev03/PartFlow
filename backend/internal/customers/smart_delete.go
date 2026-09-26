package customers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/acquisitions"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/payments"
	returnrepo "github.com/partflow/smart-store/internal/returns"
	"github.com/partflow/smart-store/internal/sales"
)

type customerTradeInDeleteRow struct {
	ID              string `db:"id"`
	InventoryItemID string `db:"inventory_item_id"`
	ProductID       string `db:"product_id"`
	Status          string `db:"status"`
}

// DeleteCustomer reverses linked sales, payments, returns and trade-ins in
// one transaction, then physically removes the customer and its remaining
// customer-owned records.
func (s *Service) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	preflight, err := s.repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin customer deletion preflight: %w", err)
	}
	saleIDs, paymentIDs, tradeIns, err := loadCustomerDeleteDependencies(ctx, preflight, s.repo.db, id)
	_ = preflight.Rollback()
	if err != nil {
		return err
	}
	saleService := sales.NewSmartDeleteService(s.repo.db)
	for _, saleID := range saleIDs {
		if err := saleService.PrepareDelete(ctx, saleID, uuid.Nil); err != nil {
			return fmt.Errorf("prepare sale %s before deleting customer: %w", saleID, err)
		}
	}
	paymentService := payments.NewService(payments.NewRepository(s.repo.db))
	for _, paymentID := range paymentIDs {
		if err := paymentService.PrepareDelete(ctx, paymentID, uuid.Nil); err != nil {
			return fmt.Errorf("prepare payment %s before deleting customer: %w", paymentID, err)
		}
	}
	acquisitionService := acquisitions.NewService(s.repo.db)
	if err := acquisitionService.PrepareOwnerAcquisitions(ctx, acquisitions.TypeCustomer, id, uuid.Nil); err != nil {
		return fmt.Errorf("prepare customer acquisitions before deletion: %w", err)
	}

	tx, err := s.repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin customer hard delete: %w", err)
	}
	defer tx.Rollback()
	if err := lockCustomerForDelete(ctx, tx, s.repo.db, id); err != nil {
		return err
	}

	paymentRepo := payments.NewRepository(s.repo.db)
	for _, paymentID := range paymentIDs {
		var exists bool
		if err := tx.GetContext(ctx, &exists, tx.Rebind(`SELECT EXISTS(SELECT 1 FROM payments WHERE id=?)`), paymentID.String()); err != nil {
			return fmt.Errorf("check customer payment %s: %w", paymentID, err)
		}
		if exists {
			if err := paymentRepo.DeleteCustomerCascadeTx(ctx, tx, paymentID, id); err != nil {
				return fmt.Errorf("remove customer payment %s after provider reversal: %w", paymentID, err)
			}
		}
	}

	returnRepository := returnrepo.NewRepository(s.repo.db)
	// The complete customer account is being removed, so remove its debt and
	// ledger sources after returns have been reversed and before deleting sales.
	// This lets sale deletion handle inventory/accounting without attempting to
	// invent allocations for legacy payments that are being erased with the
	// customer.
	for _, table := range []string{"customer_payments", "customer_debts", "debts", "debt_collections", "customer_ledger"} {
		if err := deleteCustomerRowsByReferenceTx(ctx, tx, s.repo.db, table, "customer_id", id.String()); err != nil {
			return err
		}
	}
	if err := clearCustomerSalesPaymentStateTx(ctx, tx, s.repo.db, id); err != nil {
		return err
	}
	saleIDs, _, tradeIns, err = loadCustomerDeleteDependencies(ctx, tx, s.repo.db, id)
	if err != nil {
		return err
	}
	for _, saleID := range saleIDs {
		if err := saleService.DeleteInTransaction(ctx, tx, saleID, uuid.Nil); err != nil {
			return fmt.Errorf("reverse customer sale %s: %w", saleID, err)
		}
	}

	remainingReturns, err := customerReturnIDs(ctx, tx, s.repo.db, id, false)
	if err != nil {
		return err
	}
	for _, returnID := range remainingReturns {
		if err := returnRepository.DeleteReturnTx(ctx, tx, returnID); err != nil {
			return fmt.Errorf("reverse remaining customer return %s: %w", returnID, err)
		}
	}
	if err := acquisitionService.DeleteAcquisitionsForOwnerTx(ctx, tx, acquisitions.TypeCustomer, id, uuid.Nil); err != nil {
		return fmt.Errorf("reverse customer acquisitions: %w", err)
	}
	if err := reverseCustomerTradeInsTx(ctx, tx, s.repo.db, tradeIns); err != nil {
		return err
	}
	if err := deleteCustomerOwnedRowsTx(ctx, tx, s.repo.db, id); err != nil {
		return err
	}
	if err := writeCustomerDeleteAuditTx(ctx, tx, s.repo.db, id); err != nil {
		return err
	}
	repository := Repository{db: s.repo.db}
	if err := repository.DeleteTx(ctx, tx, id); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit customer deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("customer_deleted")
	return nil
}

func clearCustomerSalesPaymentStateTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, customerID uuid.UUID) error {
	isSQLite := dbutil.IsSQLite(db)
	for column, expression := range map[string]string{
		"paid_amount":      "0",
		"remaining_amount": "COALESCE(total_amount,0)",
		"payment_status":   "'unpaid'",
	} {
		hasColumn, err := customerColumnExistsTx(ctx, tx, isSQLite, "sales", column)
		if err != nil {
			return err
		}
		if !hasColumn {
			continue
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE sales SET `+column+`=`+expression+` WHERE customer_id=?`), customerID.String()); err != nil {
			return fmt.Errorf("clear customer sale payment state before cascade: %w", err)
		}
	}
	return nil
}

func loadCustomerDeleteDependencies(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, customerID uuid.UUID) ([]uuid.UUID, []uuid.UUID, []customerTradeInDeleteRow, error) {
	var salesIDs, paymentIDs []uuid.UUID
	var tradeIns []customerTradeInDeleteRow
	isSQLite := dbutil.IsSQLite(db)
	if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "sales"); err != nil {
		return nil, nil, nil, err
	} else if exists {
		if hasCustomer, err := customerColumnExistsTx(ctx, tx, isSQLite, "sales", "customer_id"); err != nil {
			return nil, nil, nil, err
		} else if hasCustomer {
			var ids []string
			query := `SELECT CAST(id AS TEXT) FROM sales WHERE customer_id=? ORDER BY created_at DESC, id DESC`
			if !customerColumnExists(ctx, tx, isSQLite, "sales", "created_at") {
				query = `SELECT CAST(id AS TEXT) FROM sales WHERE customer_id=? ORDER BY id DESC`
			}
			if err := tx.SelectContext(ctx, &ids, tx.Rebind(query), customerID.String()); err != nil {
				return nil, nil, nil, fmt.Errorf("load customer sales: %w", err)
			}
			for _, raw := range ids {
				id, err := uuid.Parse(raw)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("parse customer sale %q: %w", raw, err)
				}
				salesIDs = appendUniqueCustomerID(salesIDs, id)
			}
		}
	}
	if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "payments"); err != nil {
		return nil, nil, nil, err
	} else if exists {
		if hasCustomer, err := customerColumnExistsTx(ctx, tx, isSQLite, "payments", "customer_id"); err != nil {
			return nil, nil, nil, err
		} else if hasCustomer {
			var ids []string
			if err := tx.SelectContext(ctx, &ids, tx.Rebind(`SELECT CAST(id AS TEXT) FROM payments WHERE customer_id=? ORDER BY id`), customerID.String()); err != nil {
				return nil, nil, nil, fmt.Errorf("load customer payments: %w", err)
			}
			for _, raw := range ids {
				id, err := uuid.Parse(raw)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("parse customer payment %q: %w", raw, err)
				}
				paymentIDs = appendUniqueCustomerID(paymentIDs, id)
			}
		}
	}
	if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "trade_ins"); err != nil {
		return nil, nil, nil, err
	} else if exists {
		if hasCustomer, err := customerColumnExistsTx(ctx, tx, isSQLite, "trade_ins", "customer_id"); err != nil {
			return nil, nil, nil, err
		} else if hasCustomer {
			tradeInQuery := `SELECT CAST(t.id AS TEXT) AS id, COALESCE(CAST(t.inventory_item_id AS TEXT),'') AS inventory_item_id, COALESCE(CAST(i.product_id AS TEXT),'') AS product_id, COALESCE(UPPER(i.status),'') AS status FROM trade_ins t LEFT JOIN inventory_items i ON i.id=t.inventory_item_id WHERE t.customer_id=? ORDER BY t.created_at DESC, t.id DESC`
			if !customerColumnExists(ctx, tx, isSQLite, "inventory_items", "product_id") {
				tradeInQuery = `SELECT CAST(id AS TEXT) AS id, COALESCE(CAST(inventory_item_id AS TEXT),'') AS inventory_item_id, '' AS product_id, '' AS status FROM trade_ins WHERE customer_id=? ORDER BY id DESC`
			}
			if err := tx.SelectContext(ctx, &tradeIns, tx.Rebind(tradeInQuery), customerID.String()); err != nil {
				return nil, nil, nil, fmt.Errorf("load customer trade-ins: %w", err)
			}
			for _, item := range tradeIns {
				if item.InventoryItemID == "" {
					continue
				}
				if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "sale_items"); err != nil {
					return nil, nil, nil, err
				} else if exists {
					if hasItemID, err := customerColumnExistsTx(ctx, tx, isSQLite, "sale_items", "inventory_item_id"); err != nil {
						return nil, nil, nil, err
					} else if hasItemID {
						var ids []string
						query := `SELECT DISTINCT CAST(s.id AS TEXT) FROM sales s JOIN sale_items si ON si.sale_id=s.id WHERE si.inventory_item_id=? ORDER BY s.created_at DESC, s.id DESC`
						if !customerColumnExists(ctx, tx, isSQLite, "sales", "created_at") {
							query = `SELECT DISTINCT CAST(s.id AS TEXT) FROM sales s JOIN sale_items si ON si.sale_id=s.id WHERE si.inventory_item_id=? ORDER BY s.id DESC`
						}
						if err := tx.SelectContext(ctx, &ids, tx.Rebind(query), item.InventoryItemID); err != nil {
							return nil, nil, nil, fmt.Errorf("load trade-in sales: %w", err)
						}
						for _, raw := range ids {
							id, err := uuid.Parse(raw)
							if err != nil {
								return nil, nil, nil, fmt.Errorf("parse trade-in sale %q: %w", raw, err)
							}
							salesIDs = appendUniqueCustomerID(salesIDs, id)
						}
					}
				}
			}
		}
	}
	return salesIDs, paymentIDs, tradeIns, nil
}

func customerReturnIDs(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, customerID uuid.UUID, onlyUnlinked bool) ([]uuid.UUID, error) {
	isSQLite := dbutil.IsSQLite(db)
	if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "returns"); err != nil || !exists {
		return nil, err
	}
	hasCustomer, err := customerColumnExistsTx(ctx, tx, isSQLite, "returns", "customer_id")
	if err != nil || !hasCustomer {
		return nil, err
	}
	query := `SELECT CAST(id AS TEXT) FROM returns WHERE customer_id=?`
	if onlyUnlinked && customerColumnExists(ctx, tx, isSQLite, "returns", "sale_id") {
		query += ` AND sale_id IS NULL`
	}
	query += ` ORDER BY id`
	var rawIDs []string
	if err := tx.SelectContext(ctx, &rawIDs, tx.Rebind(query), customerID.String()); err != nil {
		return nil, fmt.Errorf("load customer returns: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("parse customer return %q: %w", raw, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func reverseCustomerTradeInsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, tradeIns []customerTradeInDeleteRow) error {
	isSQLite := dbutil.IsSQLite(db)
	for _, tradeIn := range tradeIns {
		if tradeIn.InventoryItemID == "" {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM trade_ins WHERE id=?`), tradeIn.ID); err != nil {
				return fmt.Errorf("delete customer trade-in record: %w", err)
			}
			continue
		}
		var status, productID string
		if err := tx.GetContext(ctx, &status, tx.Rebind(`SELECT UPPER(TRIM(COALESCE(status,''))) FROM inventory_items WHERE id=?`), tradeIn.InventoryItemID); err != nil {
			return fmt.Errorf("load trade-in inventory item %s: %w", tradeIn.InventoryItemID, err)
		}
		if err := tx.GetContext(ctx, &productID, tx.Rebind(`SELECT CAST(product_id AS TEXT) FROM inventory_items WHERE id=?`), tradeIn.InventoryItemID); err != nil {
			return fmt.Errorf("load trade-in product: %w", err)
		}
		if status != "AVAILABLE" {
			return fmt.Errorf("trade-in item %s still has %s status after reversing its sales; reverse its later stock operation first", tradeIn.InventoryItemID, status)
		}
		var movementCount int
		if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "inventory_movements"); err != nil {
			return err
		} else if exists {
			if err := tx.GetContext(ctx, &movementCount, tx.Rebind(`SELECT COUNT(*) FROM inventory_movements WHERE item_id=?`), tradeIn.InventoryItemID); err != nil {
				return err
			}
			if movementCount > 0 {
				return fmt.Errorf("trade-in item %s has inventory movements outside its deleted sales; reverse those movements before deleting its owner", tradeIn.InventoryItemID)
			}
		}
		if tradeIn.Status == "SOLD" {
			if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "inventory"); err != nil {
				return err
			} else if exists {
				result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=quantity-1, updated_at=CURRENT_TIMESTAMP WHERE product_id=? AND COALESCE(quantity,0)>0`), productID)
				if err != nil {
					return fmt.Errorf("remove stock restored from deleted trade-in sales: %w", err)
				}
				if affected, _ := result.RowsAffected(); affected == 0 {
					return fmt.Errorf("cannot reconcile aggregate stock for sold trade-in item %s", tradeIn.InventoryItemID)
				}
			}
		}
		for _, related := range []struct{ table, column string }{
			{"trade_ins", "inventory_item_id"},
			{"item_specification_values", "inventory_item_id"},
			{"barcodes", "inventory_item_id"},
			{"warranty_claims", "inventory_item_id"},
			{"item_repair_costs", "inventory_item_id"},
			{"item_history", "inventory_item_id"},
			{"inventory_movements", "item_id"},
			{"reservations", "item_id"},
		} {
			if err := deleteCustomerRowsByReferenceTx(ctx, tx, db, related.table, related.column, tradeIn.InventoryItemID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM inventory_items WHERE id=?`), tradeIn.InventoryItemID); err != nil {
			return fmt.Errorf("hard delete customer trade-in stock item: %w", err)
		}
		if hasSKU, err := customerColumnExistsTx(ctx, tx, isSQLite, "products", "sku"); err != nil {
			return err
		} else if hasSKU {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM products WHERE id=? AND sku LIKE 'TRD-%' AND NOT EXISTS(SELECT 1 FROM inventory_items WHERE product_id=products.id) AND NOT EXISTS(SELECT 1 FROM sale_items WHERE product_id=products.id) AND NOT EXISTS(SELECT 1 FROM purchase_items WHERE product_id=products.id)`), productID); err != nil {
				return fmt.Errorf("delete unused product created for trade-in: %w", err)
			}
		}
	}
	return nil
}

func deleteCustomerOwnedRowsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, customerID uuid.UUID) error {
	isSQLite := dbutil.IsSQLite(db)
	for _, table := range []string{"customer_payments", "customer_debts", "debts", "debt_collections", "customer_ledger", "warranty_claims"} {
		if err := deleteCustomerRowsByReferenceTx(ctx, tx, db, table, "customer_id", customerID.String()); err != nil {
			return err
		}
	}
	if hasCustomerID, err := customerColumnExistsTx(ctx, tx, isSQLite, "inventory_items", "customer_id"); err != nil {
		return err
	} else if hasCustomerID {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory_items SET customer_id=NULL WHERE customer_id=?`), customerID.String()); err != nil {
			return fmt.Errorf("detach retained inventory items from customer: %w", err)
		}
	}
	if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "reservations"); err != nil {
		return err
	} else if exists {
		if hasCustomerID, err := customerColumnExistsTx(ctx, tx, isSQLite, "reservations", "customer_id"); err != nil {
			return err
		} else if hasCustomerID {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory_items SET status='AVAILABLE', updated_at=CURRENT_TIMESTAMP WHERE status='RESERVED' AND id IN (SELECT item_id FROM reservations WHERE customer_id=?)`), customerID.String()); err != nil {
				return fmt.Errorf("release customer item reservations: %w", err)
			}
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM reservations WHERE customer_id=?`), customerID.String()); err != nil {
				return fmt.Errorf("delete customer reservations: %w", err)
			}
		}
	}
	for _, table := range []string{"audit_logs"} {
		if exists, err := customerTableExistsTx(ctx, tx, isSQLite, table); err != nil {
			return err
		} else if exists {
			if hasEntityID, err := customerColumnExistsTx(ctx, tx, isSQLite, table, "entity_id"); err != nil {
				return err
			} else if hasEntityID {
				query := `DELETE FROM audit_logs WHERE entity_id=?`
				if customerColumnExists(ctx, tx, isSQLite, table, "entity_type") {
					query += ` AND LOWER(COALESCE(entity_type,'')) IN ('customer','customers')`
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(query), customerID.String()); err != nil {
					return fmt.Errorf("remove old customer audit rows: %w", err)
				}
			}
		}
	}
	return nil
}

func deleteCustomerRowsByReferenceTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table, column, value string) error {
	isSQLite := dbutil.IsSQLite(db)
	exists, err := customerTableExistsTx(ctx, tx, isSQLite, table)
	if err != nil || !exists {
		return err
	}
	hasColumn, err := customerColumnExistsTx(ctx, tx, isSQLite, table, column)
	if err != nil || !hasColumn {
		return err
	}
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s=?`, table, column)
	if _, err := tx.ExecContext(ctx, tx.Rebind(query), value); err != nil {
		return fmt.Errorf("delete customer dependency %s.%s: %w", table, column, err)
	}
	return nil
}

func lockCustomerForDelete(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, id uuid.UUID) error {
	isSQLite := dbutil.IsSQLite(db)
	if isSQLite {
		result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE customers SET is_active=is_active WHERE id=?`), id.String())
		if err != nil {
			return fmt.Errorf("lock customer before deletion: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return ErrCustomerNotFound
		}
		return nil
	}
	var rawID string
	if err := tx.GetContext(ctx, &rawID, `SELECT CAST(id AS TEXT) FROM customers WHERE id=$1 FOR UPDATE`, id); err != nil {
		if err == sql.ErrNoRows {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("lock customer before deletion: %w", err)
	}
	return nil
}

func customerTableExistsTx(ctx context.Context, tx *sqlx.Tx, isSQLite bool, table string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1)`
	if isSQLite {
		query = `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name=$1)`
	}
	if err := tx.GetContext(ctx, &exists, query, table); err != nil {
		return false, err
	}
	return exists, nil
}

func customerColumnExistsTx(ctx context.Context, tx *sqlx.Tx, isSQLite bool, table, column string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name=$1 AND column_name=$2)`
	if isSQLite {
		query = `SELECT EXISTS(SELECT 1 FROM pragma_table_info('` + strings.ReplaceAll(table, "'", "''") + `') WHERE name=$1)`
	}
	if isSQLite {
		if err := tx.GetContext(ctx, &exists, query, column); err != nil {
			return false, err
		}
		return exists, nil
	}
	if err := tx.GetContext(ctx, &exists, query, table, column); err != nil {
		if isSQLite {
			return false, err
		}
		return false, err
	}
	return exists, nil
}

func customerColumnExists(ctx context.Context, tx *sqlx.Tx, isSQLite bool, table, column string) bool {
	exists, err := customerColumnExistsTx(ctx, tx, isSQLite, table, column)
	return err == nil && exists
}

func appendUniqueCustomerID(ids []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}

func writeCustomerDeleteAuditTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, customerID uuid.UUID) error {
	isSQLite := dbutil.IsSQLite(db)
	if exists, err := customerTableExistsTx(ctx, tx, isSQLite, "audit_logs"); err != nil || !exists {
		return err
	}
	if !customerColumnExists(ctx, tx, isSQLite, "audit_logs", "entity_type") || !customerColumnExists(ctx, tx, isSQLite, "audit_logs", "entity_id") {
		return nil
	}
	payload, _ := json.Marshal(map[string]any{"hard_deleted": true})
	query := `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES (?,NULL,'DELETE','customer',?,?,CURRENT_TIMESTAMP)`
	if !isSQLite {
		query = `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES ($1,NULL,'DELETE','customer',$2,$3::jsonb,NOW())`
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(query), uuid.New().String(), customerID.String(), string(payload)); err != nil {
		return fmt.Errorf("write customer deletion audit: %w", err)
	}
	return nil
}
