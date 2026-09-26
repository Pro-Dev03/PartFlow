package suppliers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/acquisitions"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/payments"
	"github.com/partflow/smart-store/internal/purchases"
	"github.com/partflow/smart-store/internal/supplierreturns"
)

func (s *Service) deleteSupplierCascade(ctx context.Context, supplierID uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, supplierID); err != nil {
		return err
	}
	preflight, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin supplier deletion preflight: %w", err)
	}
	purchaseIDs, paymentIDs, returnIDs, err := supplierDeleteDependencies(ctx, preflight, s.db, supplierID)
	_ = preflight.Rollback()
	if err != nil {
		return err
	}
	purchaseService := purchases.NewSmartDeleteService(s.db)
	for _, id := range purchaseIDs {
		if err := purchaseService.PrepareDelete(ctx, id, uuid.Nil); err != nil {
			return fmt.Errorf("prepare purchase %s before deleting supplier: %w", id, err)
		}
	}
	paymentService := payments.NewService(payments.NewRepository(s.db))
	for _, id := range paymentIDs {
		if err := paymentService.PrepareDelete(ctx, id, uuid.Nil); err != nil {
			return fmt.Errorf("prepare supplier payment %s before deletion: %w", id, err)
		}
	}
	acquisitionService := acquisitions.NewService(s.db)
	if err := acquisitionService.PrepareOwnerAcquisitions(ctx, acquisitions.TypeSupplier, supplierID, uuid.Nil); err != nil {
		return err
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin supplier hard delete: %w", err)
	}
	defer tx.Rollback()
	if err := lockSupplierForDelete(ctx, tx, s.db, supplierID); err != nil {
		return err
	}
	currentPurchases, currentPayments, currentReturns, err := supplierDeleteDependencies(ctx, tx, s.db, supplierID)
	if err != nil {
		return err
	}
	if !sameSupplierIDs(purchaseIDs, currentPurchases) || !sameSupplierIDs(paymentIDs, currentPayments) || !sameSupplierIDs(returnIDs, currentReturns) {
		return fmt.Errorf("supplier transactions changed during deletion; retry so external payments can be reconciled first")
	}

	returnService := supplierreturns.NewService(s.db)
	for _, id := range currentReturns {
		if err := returnService.DeleteTx(ctx, tx, id, true, uuid.Nil); err != nil {
			return fmt.Errorf("reverse supplier return %s: %w", id, err)
		}
	}
	paymentRepository := payments.NewRepository(s.db)
	for _, id := range currentPayments {
		if err := paymentRepository.DeleteSupplierCascadeTx(ctx, tx, id, supplierID); err != nil {
			return fmt.Errorf("remove supplier payment %s after provider reversal: %w", id, err)
		}
	}
	if err := acquisitionService.DeleteAcquisitionsForOwnerTx(ctx, tx, acquisitions.TypeSupplier, supplierID, uuid.Nil); err != nil {
		return fmt.Errorf("reverse supplier acquisitions: %w", err)
	}
	for _, table := range []string{"supplier_debt_collections", "supplier_debts", "supplier_payments", "supplier_ledger"} {
		if err := deleteSupplierRowsTx(ctx, tx, s.db, table, "supplier_id", supplierID.String()); err != nil {
			return err
		}
	}
	if hasPaidAmount, err := supplierColumnExistsTx(ctx, tx, s.db, "purchases", "paid_amount"); err != nil {
		return err
	} else if hasPaidAmount {
		query := `UPDATE purchases SET paid_amount=0`
		if hasRemaining, err := supplierColumnExistsTx(ctx, tx, s.db, "purchases", "remaining_amount"); err != nil {
			return err
		} else if hasRemaining {
			query += `, remaining_amount=COALESCE(total_amount,0)`
		}
		query += ` WHERE supplier_id=?`
		if _, err := tx.ExecContext(ctx, tx.Rebind(query), supplierID.String()); err != nil {
			return fmt.Errorf("clear supplier purchase payment state before cascade: %w", err)
		}
	}
	for _, id := range currentPurchases {
		result, err := purchaseService.SmartDeleteTx(ctx, tx, id, uuid.Nil)
		if err != nil {
			return fmt.Errorf("reverse supplier purchase %s: %w", id, err)
		}
		if result.Action != "deleted" {
			return fmt.Errorf("supplier purchase %s could not be safely deleted: %s", id, result.Message)
		}
	}
	for _, row := range []struct{ table, column string }{
		{"products", "preferred_supplier_id"}, {"inventory_items", "supplier_id"},
	} {
		if exists, err := supplierTableExistsTx(ctx, tx, s.db, row.table); err != nil {
			return err
		} else if exists {
			if hasColumn, err := supplierColumnExistsTx(ctx, tx, s.db, row.table, row.column); err != nil {
				return err
			} else if hasColumn {
				if _, err := tx.ExecContext(ctx, tx.Rebind(fmt.Sprintf(`UPDATE %s SET %s=NULL WHERE %s=?`, row.table, row.column, row.column)), supplierID.String()); err != nil {
					return fmt.Errorf("detach supplier reference from %s: %w", row.table, err)
				}
			}
		}
	}
	for _, table := range []string{"audit_logs"} {
		if exists, err := supplierTableExistsTx(ctx, tx, s.db, table); err != nil {
			return err
		} else if exists {
			if hasEntity, err := supplierColumnExistsTx(ctx, tx, s.db, table, "entity_id"); err != nil {
				return err
			} else if hasEntity {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM audit_logs WHERE entity_id=? AND LOWER(COALESCE(entity_type,'')) IN ('supplier','suppliers')`), supplierID.String()); err != nil {
					return fmt.Errorf("remove old supplier audit rows: %w", err)
				}
			}
		}
	}
	if exists, err := supplierTableExistsTx(ctx, tx, s.db, "audit_logs"); err != nil {
		return err
	} else if exists {
		if hasEntityType, err := supplierColumnExistsTx(ctx, tx, s.db, "audit_logs", "entity_type"); err != nil {
			return err
		} else if hasEntityType {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES (?,NULL,'DELETE','supplier',?,'{"hard_deleted":true}',CURRENT_TIMESTAMP)`), uuid.NewString(), supplierID.String()); err != nil {
				return fmt.Errorf("write supplier deletion audit: %w", err)
			}
		}
	}
	result, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM suppliers WHERE id=?`), supplierID.String())
	if err != nil {
		return fmt.Errorf("hard delete supplier: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrSupplierNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit supplier hard delete: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("supplier_deleted")
	return nil
}

func supplierDeleteDependencies(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, supplierID uuid.UUID) ([]uuid.UUID, []uuid.UUID, []uuid.UUID, error) {
	purchaseIDs, err := supplierIDsByOwnerTx(ctx, tx, db, "purchases", "supplier_id", supplierID.String())
	if err != nil {
		return nil, nil, nil, err
	}
	paymentIDs, err := supplierIDsByOwnerTx(ctx, tx, db, "payments", "supplier_id", supplierID.String())
	if err != nil {
		return nil, nil, nil, err
	}
	returnIDs, err := supplierIDsByOwnerTx(ctx, tx, db, "supplier_returns", "supplier_id", supplierID.String())
	if err != nil {
		return nil, nil, nil, err
	}
	return purchaseIDs, paymentIDs, returnIDs, nil
}

func supplierIDsByOwnerTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table, column, ownerID string) ([]uuid.UUID, error) {
	if exists, err := supplierTableExistsTx(ctx, tx, db, table); err != nil || !exists {
		return nil, err
	}
	if hasColumn, err := supplierColumnExistsTx(ctx, tx, db, table, column); err != nil || !hasColumn {
		return nil, err
	}
	var rawIDs []string
	query := fmt.Sprintf(`SELECT CAST(id AS TEXT) FROM %s WHERE %s=? ORDER BY id`, table, column)
	if err := tx.SelectContext(ctx, &rawIDs, tx.Rebind(query), ownerID); err != nil {
		return nil, fmt.Errorf("load supplier %s dependencies: %w", table, err)
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("parse supplier %s id %q: %w", table, raw, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func lockSupplierForDelete(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, supplierID uuid.UUID) error {
	if dbutil.IsSQLite(db) {
		result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE suppliers SET is_active=is_active WHERE id=?`), supplierID.String())
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return ErrSupplierNotFound
		}
		return nil
	}
	var locked string
	if err := tx.GetContext(ctx, &locked, `SELECT CAST(id AS TEXT) FROM suppliers WHERE id=$1 FOR UPDATE`, supplierID); err != nil {
		if err == sql.ErrNoRows {
			return ErrSupplierNotFound
		}
		return fmt.Errorf("lock supplier for deletion: %w", err)
	}
	return nil
}

func deleteSupplierRowsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table, column, value string) error {
	if exists, err := supplierTableExistsTx(ctx, tx, db, table); err != nil || !exists {
		return err
	}
	if hasColumn, err := supplierColumnExistsTx(ctx, tx, db, table, column); err != nil || !hasColumn {
		return err
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(fmt.Sprintf(`DELETE FROM %s WHERE %s=?`, table, column)), value); err != nil {
		return fmt.Errorf("delete supplier dependency %s.%s: %w", table, column, err)
	}
	return nil
}

func supplierTableExistsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1)`
	if dbutil.IsSQLite(db) {
		query = `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name=$1)`
	}
	var exists bool
	if err := tx.GetContext(ctx, &exists, query, table); err != nil {
		return false, err
	}
	return exists, nil
}

func supplierColumnExistsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table, column string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name=$1 AND column_name=$2)`
	if dbutil.IsSQLite(db) {
		query = `SELECT EXISTS(SELECT 1 FROM pragma_table_info('` + table + `') WHERE name=$1)`
		var exists bool
		if err := tx.GetContext(ctx, &exists, query, column); err != nil {
			return false, err
		}
		return exists, nil
	}
	var exists bool
	if err := tx.GetContext(ctx, &exists, query, table, column); err != nil {
		return false, err
	}
	return exists, nil
}

func sameSupplierIDs(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	set := make(map[uuid.UUID]int, len(left))
	for _, id := range left {
		set[id]++
	}
	for _, id := range right {
		set[id]--
		if set[id] < 0 {
			return false
		}
	}
	return true
}
