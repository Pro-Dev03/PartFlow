package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/acquisitions"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/returns"
	"github.com/partflow/smart-store/internal/sales"
	"github.com/partflow/smart-store/internal/supplierreturns"
)

type inventoryItemDeleteDependencies struct {
	Sales           []uuid.UUID
	Returns         []uuid.UUID
	SupplierReturns []uuid.UUID
	Acquisitions    []uuid.UUID
}

func (s *Service) deleteInventoryItemCascade(ctx context.Context, itemID, userID uuid.UUID) error {
	preflight, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin inventory item delete preflight: %w", err)
	}
	dependencies, err := loadInventoryItemDeleteDependencies(ctx, preflight, s.db, itemID)
	_ = preflight.Rollback()
	if err != nil {
		return err
	}
	saleService := sales.NewSmartDeleteService(s.db)
	for _, id := range dependencies.Sales {
		if err := saleService.PrepareDelete(ctx, id, userID); err != nil {
			return fmt.Errorf("prepare inventory item sale %s: %w", id, err)
		}
	}
	acquisitionService := acquisitions.NewService(s.db)
	if err := acquisitionService.PrepareInventoryItemAcquisitions(ctx, itemID, userID); err != nil {
		return err
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin inventory item cascade delete: %w", err)
	}
	defer tx.Rollback()
	lock := `UPDATE inventory_items SET id=id WHERE id=?`
	if _, err := tx.ExecContext(ctx, tx.Rebind(lock), itemID.String()); err != nil {
		return fmt.Errorf("lock inventory item before deletion: %w", err)
	}
	var itemSnapshot struct {
		ProductID string `db:"product_id"`
		Status    string `db:"status"`
	}
	if err := tx.GetContext(ctx, &itemSnapshot, tx.Rebind(`SELECT CAST(product_id AS TEXT) AS product_id,status FROM inventory_items WHERE id=?`), itemID.String()); err != nil {
		if err == sql.ErrNoRows {
			return ErrItemNotFound
		}
		return fmt.Errorf("load inventory item snapshot: %w", err)
	}
	current, err := loadInventoryItemDeleteDependencies(ctx, tx, s.db, itemID)
	if err != nil {
		return err
	}
	if !sameInventoryUUIDs(dependencies.Sales, current.Sales) ||
		!sameInventoryUUIDs(dependencies.Returns, current.Returns) ||
		!sameInventoryUUIDs(dependencies.SupplierReturns, current.SupplierReturns) ||
		!sameInventoryUUIDs(dependencies.Acquisitions, current.Acquisitions) {
		return fmt.Errorf("inventory item transactions changed during deletion; retry so linked effects can be reconciled")
	}
	for _, id := range current.Returns {
		if err := returns.NewRepository(s.db).DeleteReturnTx(ctx, tx, id); err != nil {
			return fmt.Errorf("reverse inventory item return %s: %w", id, err)
		}
	}
	returnService := supplierreturns.NewService(s.db)
	for _, id := range current.SupplierReturns {
		var exists bool
		if err := tx.GetContext(ctx, &exists, tx.Rebind(`SELECT EXISTS(SELECT 1 FROM supplier_returns WHERE id=?)`), id.String()); err != nil {
			return fmt.Errorf("recheck inventory item supplier return %s: %w", id, err)
		}
		if exists {
			if err := returnService.DeleteTx(ctx, tx, id, true, userID); err != nil {
				return fmt.Errorf("reverse inventory item supplier return %s: %w", id, err)
			}
		}
	}
	for _, id := range current.Sales {
		result, err := saleService.SmartDeleteTx(ctx, tx, id, userID)
		if err != nil {
			return fmt.Errorf("reverse inventory item sale %s: %w", id, err)
		}
		if result == nil || result.Action != "deleted" {
			if result != nil && result.Message != "" {
				return fmt.Errorf("inventory item sale %s cannot be reversed: %s", id, result.Message)
			}
			return fmt.Errorf("inventory item sale %s cannot be reversed", id)
		}
	}
	if err := acquisitionService.DeleteAcquisitionsByIDsTx(ctx, tx, current.Acquisitions, userID); err != nil {
		return fmt.Errorf("reverse inventory item acquisition documents: %w", err)
	}
	var stillExists bool
	if err := tx.GetContext(ctx, &stillExists, tx.Rebind(`SELECT EXISTS(SELECT 1 FROM inventory_items WHERE id=?)`), itemID.String()); err != nil {
		return fmt.Errorf("check inventory item after dependency reversal: %w", err)
	}
	if stillExists {
		for _, dependency := range []struct{ table, column string }{
			{"sale_items", "inventory_item_id"}, {"return_items", "inventory_item_id"},
			{"supplier_return_items", "inventory_item_id"}, {"acquisition_items", "inventory_item_id"},
		} {
			if !inventoryHasColumn(s.db, dependency.table, dependency.column) {
				continue
			}
			var linked bool
			query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s=?)`, dependency.table, dependency.column)
			if err := tx.GetContext(ctx, &linked, tx.Rebind(query), itemID.String()); err != nil {
				return fmt.Errorf("verify item dependency %s: %w", dependency.table, err)
			}
			if linked {
				return fmt.Errorf("inventory item still belongs to a %s transaction after reversal", dependency.table)
			}
		}
		if err := s.repo.DeleteInventoryItemAfterReversalTx(ctx, tx, itemID, userID); err != nil {
			return err
		}
	} else {
		if exists, err := inventoryTableExistsTx(ctx, tx, s.db, "audit_logs"); err != nil {
			return err
		} else if exists {
			var actor any
			if userID != uuid.Nil {
				actor = userID.String()
			}
			snapshot := fmt.Sprintf(`{"product_id":%q,"status":%q,"effects_reversed":true}`, itemSnapshot.ProductID, itemSnapshot.Status)
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES (?,?,'DELETE','inventory_item',?,?,CURRENT_TIMESTAMP)`), uuid.New().String(), actor, itemID.String(), snapshot); err != nil {
				return fmt.Errorf("write inventory item cascade deletion audit: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inventory item cascade delete: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("inventory_item_deleted")
	return nil
}

func loadInventoryItemDeleteDependencies(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, itemID uuid.UUID) (inventoryItemDeleteDependencies, error) {
	var result inventoryItemDeleteDependencies
	queries := []struct {
		table        string
		parentColumn string
		query        string
		into         *[]uuid.UUID
	}{
		{"sale_items", "sale_id", `SELECT DISTINCT CAST(sale_id AS TEXT) FROM sale_items WHERE inventory_item_id=? AND sale_id IS NOT NULL ORDER BY CAST(sale_id AS TEXT)`, &result.Sales},
		{"return_items", "return_id", `SELECT DISTINCT CAST(return_id AS TEXT) FROM return_items WHERE inventory_item_id=? AND return_id IS NOT NULL ORDER BY CAST(return_id AS TEXT)`, &result.Returns},
		{"supplier_return_items", "supplier_return_id", `SELECT DISTINCT CAST(supplier_return_id AS TEXT) FROM supplier_return_items WHERE inventory_item_id=? AND supplier_return_id IS NOT NULL ORDER BY CAST(supplier_return_id AS TEXT)`, &result.SupplierReturns},
	}
	for _, item := range queries {
		if !inventoryHasColumn(db, item.table, "inventory_item_id") || !inventoryHasColumn(db, item.table, item.parentColumn) {
			continue
		}
		var raw []string
		if err := tx.SelectContext(ctx, &raw, tx.Rebind(item.query), itemID.String()); err != nil {
			return result, fmt.Errorf("load inventory item %s: %w", item.table, err)
		}
		for _, value := range raw {
			id, err := uuid.Parse(value)
			if err != nil {
				return result, fmt.Errorf("parse inventory item %s id %q: %w", item.table, value, err)
			}
			*item.into = append(*item.into, id)
		}
	}
	var err error
	if inventoryHasColumn(db, "acquisition_items", "inventory_item_id") && inventoryHasColumn(db, "acquisition_items", "acquisition_id") {
		result.Acquisitions, err = acquisitions.NewService(db).InventoryItemAcquisitionIDsTx(ctx, tx, itemID)
		if err != nil {
			if dbutil.IsSQLite(db) && strings.Contains(strings.ToLower(err.Error()), "no such table") {
				return result, nil
			}
			return result, err
		}
	}
	return result, nil
}

func sameInventoryUUIDs(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[uuid.UUID]int, len(left))
	for _, id := range left {
		counts[id]++
	}
	for _, id := range right {
		counts[id]--
		if counts[id] < 0 {
			return false
		}
	}
	return true
}
