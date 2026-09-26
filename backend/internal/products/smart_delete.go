package products

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/acquisitions"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/purchases"
	"github.com/partflow/smart-store/internal/returns"
	"github.com/partflow/smart-store/internal/sales"
	"github.com/partflow/smart-store/internal/supplierreturns"
)

type productDeleteDependencies struct {
	Sales            []uuid.UUID
	Purchases        []uuid.UUID
	Returns          []uuid.UUID
	SupplierReturns  []uuid.UUID
	AcquisitionSales []uuid.UUID
}

func (s *Service) deleteProductCascade(ctx context.Context, productID uuid.UUID) error {
	preflight, err := s.repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin product deletion preflight: %w", err)
	}
	dependencies, err := loadProductDeleteDependencies(ctx, preflight, s.repo.db, productID)
	_ = preflight.Rollback()
	if err != nil {
		return err
	}
	saleService := sales.NewSmartDeleteService(s.repo.db)
	for _, id := range dependencies.Sales {
		if err := saleService.PrepareDelete(ctx, id, uuid.Nil); err != nil {
			return fmt.Errorf("prepare product sale %s: %w", id, err)
		}
	}
	purchaseService := purchases.NewSmartDeleteService(s.repo.db)
	for _, id := range dependencies.Purchases {
		if err := purchaseService.PrepareDelete(ctx, id, uuid.Nil); err != nil {
			return fmt.Errorf("prepare product purchase %s: %w", id, err)
		}
	}
	acquisitionService := acquisitions.NewService(s.repo.db)
	if err := acquisitionService.PrepareProductAcquisitions(ctx, productID, uuid.Nil); err != nil {
		return err
	}

	tx, err := s.repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin product hard delete: %w", err)
	}
	defer tx.Rollback()
	if err := lockProductForCascadeDelete(ctx, tx, s.repo.db, productID); err != nil {
		return err
	}
	current, err := loadProductDeleteDependencies(ctx, tx, s.repo.db, productID)
	if err != nil {
		return err
	}
	current.AcquisitionSales, err = acquisitionService.ProductAcquisitionSaleIDsTx(ctx, tx, productID)
	if err != nil {
		return err
	}
	if !sameProductUUIDs(dependencies.Sales, current.Sales) ||
		!sameProductUUIDs(dependencies.Purchases, current.Purchases) ||
		!sameProductUUIDs(dependencies.Returns, current.Returns) ||
		!sameProductUUIDs(dependencies.SupplierReturns, current.SupplierReturns) ||
		!sameProductUUIDs(dependencies.AcquisitionSales, current.AcquisitionSales) {
		return fmt.Errorf("product transactions changed during deletion; retry so all linked effects can be reconciled first")
	}

	returnService := returns.NewRepository(s.repo.db)
	supplierReturnService := supplierreturns.NewService(s.repo.db)
	for _, id := range current.Sales {
		if err := saleService.DeleteInTransaction(ctx, tx, id, uuid.Nil); err != nil {
			return fmt.Errorf("reverse product sale %s: %w", id, err)
		}
	}
	for _, id := range current.Returns {
		var stillExists bool
		if err := tx.GetContext(ctx, &stillExists, tx.Rebind(`SELECT EXISTS(SELECT 1 FROM returns WHERE id=?)`), id.String()); err != nil {
			return fmt.Errorf("recheck product return %s: %w", id, err)
		}
		if stillExists {
			if err := returnService.DeleteReturnTx(ctx, tx, id); err != nil {
				return fmt.Errorf("reverse product return %s: %w", id, err)
			}
		}
	}
	for _, id := range current.SupplierReturns {
		// A linked supplier return may have been removed while reversing its
		// customer return above; skip the already-reversed child.
		var stillExists bool
		if err := tx.GetContext(ctx, &stillExists, tx.Rebind(`SELECT EXISTS(SELECT 1 FROM supplier_returns WHERE id=?)`), id.String()); err != nil {
			return fmt.Errorf("recheck product supplier return %s: %w", id, err)
		}
		if stillExists {
			if err := supplierReturnService.DeleteTx(ctx, tx, id, true, uuid.Nil); err != nil {
				return fmt.Errorf("reverse product supplier return %s: %w", id, err)
			}
		}
	}
	if err := acquisitionService.DeleteAcquisitionsForProductTx(ctx, tx, productID, uuid.Nil); err != nil {
		return fmt.Errorf("reverse product acquisitions: %w", err)
	}
	for _, id := range current.Purchases {
		result, err := purchaseService.SmartDeleteTx(ctx, tx, id, uuid.Nil)
		if err != nil {
			return fmt.Errorf("reverse product purchase %s: %w", id, err)
		}
		if result.Action != "deleted" {
			return fmt.Errorf("product purchase %s could not be safely deleted: %s", id, result.Message)
		}
	}
	if err := s.repo.DeleteProductCascadeTx(ctx, tx, productID); err != nil {
		return fmt.Errorf("hard delete product after reversing dependencies: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit product hard delete: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("product_deleted")
	return nil
}

func loadProductDeleteDependencies(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, productID uuid.UUID) (productDeleteDependencies, error) {
	var result productDeleteDependencies
	queries := []struct {
		table string
		query string
		into  *[]uuid.UUID
	}{
		{"sale_items", `SELECT DISTINCT CAST(sale_id AS TEXT) FROM sale_items WHERE product_id=? AND sale_id IS NOT NULL ORDER BY CAST(sale_id AS TEXT)`, &result.Sales},
		{"purchase_items", `SELECT DISTINCT CAST(purchase_id AS TEXT) FROM purchase_items WHERE product_id=? AND purchase_id IS NOT NULL ORDER BY CAST(purchase_id AS TEXT)`, &result.Purchases},
		{"return_items", `SELECT DISTINCT CAST(return_id AS TEXT) FROM return_items WHERE product_id=? AND return_id IS NOT NULL ORDER BY CAST(return_id AS TEXT)`, &result.Returns},
		{"supplier_return_items", `SELECT DISTINCT CAST(supplier_return_id AS TEXT) FROM supplier_return_items WHERE product_id=? AND supplier_return_id IS NOT NULL ORDER BY CAST(supplier_return_id AS TEXT)`, &result.SupplierReturns},
	}
	for _, item := range queries {
		if exists, err := productTableExistsTx(ctx, tx, db, item.table); err != nil || !exists {
			if err != nil {
				return result, err
			}
			continue
		}
		var raw []string
		if err := tx.SelectContext(ctx, &raw, tx.Rebind(item.query), productID.String()); err != nil {
			return result, fmt.Errorf("load product %s: %w", item.table, err)
		}
		for _, value := range raw {
			id, err := uuid.Parse(value)
			if err != nil {
				return result, fmt.Errorf("parse product %s id %q: %w", item.table, value, err)
			}
			*item.into = append(*item.into, id)
		}
	}
	acquisitionSales, err := acquisitions.NewService(db).ProductAcquisitionSaleIDsTx(ctx, tx, productID)
	if err != nil {
		return result, err
	}
	result.AcquisitionSales = acquisitionSales
	return result, nil
}

func productTableExistsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table string) (bool, error) {
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

func lockProductForCascadeDelete(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, productID uuid.UUID) error {
	if dbutil.IsSQLite(db) {
		result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE products SET id=id WHERE id=?`), productID.String())
		if err != nil {
			return fmt.Errorf("lock product before deletion: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return ErrProductNotFound
		}
		return nil
	}
	var locked string
	if err := tx.GetContext(ctx, &locked, `SELECT CAST(id AS TEXT) FROM products WHERE id=$1 FOR UPDATE`, productID); err != nil {
		if err == sql.ErrNoRows {
			return ErrProductNotFound
		}
		return fmt.Errorf("lock product before deletion: %w", err)
	}
	return nil
}

func sameProductUUIDs(left, right []uuid.UUID) bool {
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
