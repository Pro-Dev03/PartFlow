package acquisitions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/sales"
)

// PrepareOwnerAcquisitions preflights external effects for sales made from
// items acquired from a supplier/customer. It must run before the parent
// deletion transaction starts.
func (s *Service) PrepareOwnerAcquisitions(ctx context.Context, ownerType string, ownerID, userID uuid.UUID) error {
	ids, err := s.AcquisitionSaleIDsForOwner(ctx, ownerType, ownerID)
	if err != nil {
		return err
	}
	service := sales.NewSmartDeleteService(s.db)
	for _, id := range ids {
		if err := service.PrepareDelete(ctx, id, userID); err != nil {
			return fmt.Errorf("prepare acquisition-linked sale %s: %w", id, err)
		}
	}
	return nil
}

func (s *Service) PrepareProductAcquisitions(ctx context.Context, productID, userID uuid.UUID) error {
	ids, err := s.ProductAcquisitionSaleIDs(ctx, productID)
	if err != nil {
		return err
	}
	service := sales.NewSmartDeleteService(s.db)
	for _, id := range ids {
		if err := service.PrepareDelete(ctx, id, userID); err != nil {
			return fmt.Errorf("prepare product acquisition sale %s: %w", id, err)
		}
	}
	return nil
}

func (s *Service) ProductAcquisitionSaleIDs(ctx context.Context, productID uuid.UUID) ([]uuid.UUID, error) {
	available, err := acquisitionSaleLinkTablesExist(ctx, s.db, nil)
	if err != nil || !available {
		return nil, err
	}
	var raw []string
	query := `SELECT DISTINCT CAST(si.sale_id AS TEXT) FROM acquisition_items ai JOIN sale_items si ON si.inventory_item_id=ai.inventory_item_id WHERE ai.product_id=? AND si.sale_id IS NOT NULL ORDER BY CAST(si.sale_id AS TEXT)`
	if err := s.db.SelectContext(ctx, &raw, s.db.Rebind(query), productID.String()); err != nil {
		return nil, fmt.Errorf("load product acquisition sales: %w", err)
	}
	return parseAcquisitionSaleIDs(raw, "product acquisition")
}

func (s *Service) ProductAcquisitionSaleIDsTx(ctx context.Context, tx *sqlx.Tx, productID uuid.UUID) ([]uuid.UUID, error) {
	available, err := acquisitionSaleLinkTablesExist(ctx, s.db, tx)
	if err != nil || !available {
		return nil, err
	}
	var raw []string
	query := `SELECT DISTINCT CAST(si.sale_id AS TEXT) FROM acquisition_items ai JOIN sale_items si ON si.inventory_item_id=ai.inventory_item_id WHERE ai.product_id=? AND si.sale_id IS NOT NULL ORDER BY CAST(si.sale_id AS TEXT)`
	if err := tx.SelectContext(ctx, &raw, tx.Rebind(query), productID.String()); err != nil {
		return nil, fmt.Errorf("load product acquisition sales: %w", err)
	}
	return parseAcquisitionSaleIDs(raw, "product acquisition")
}

func parseAcquisitionSaleIDs(raw []string, source string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(raw))
	for _, value := range raw {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("parse %s sale %q: %w", source, value, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *Service) InventoryItemAcquisitionIDs(ctx context.Context, inventoryItemID uuid.UUID) ([]uuid.UUID, error) {
	exists, err := acquisitionTableExists(ctx, s.db, nil, "acquisition_items")
	if err != nil || !exists {
		return nil, err
	}
	var raw []string
	if err := s.db.SelectContext(ctx, &raw, s.db.Rebind(`SELECT DISTINCT CAST(acquisition_id AS TEXT) FROM acquisition_items WHERE inventory_item_id=? ORDER BY CAST(acquisition_id AS TEXT)`), inventoryItemID.String()); err != nil {
		return nil, fmt.Errorf("load inventory item acquisitions: %w", err)
	}
	return parseInventoryAcquisitionIDs(raw)
}

func (s *Service) InventoryItemAcquisitionIDsTx(ctx context.Context, tx *sqlx.Tx, inventoryItemID uuid.UUID) ([]uuid.UUID, error) {
	exists, err := acquisitionTableExists(ctx, s.db, tx, "acquisition_items")
	if err != nil || !exists {
		return nil, err
	}
	var raw []string
	if err := tx.SelectContext(ctx, &raw, tx.Rebind(`SELECT DISTINCT CAST(acquisition_id AS TEXT) FROM acquisition_items WHERE inventory_item_id=? ORDER BY CAST(acquisition_id AS TEXT)`), inventoryItemID.String()); err != nil {
		return nil, fmt.Errorf("load inventory item acquisitions: %w", err)
	}
	return parseInventoryAcquisitionIDs(raw)
}

func parseInventoryAcquisitionIDs(raw []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(raw))
	for _, value := range raw {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("parse inventory item acquisition %q: %w", value, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func acquisitionSaleLinkTablesExist(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx) (bool, error) {
	for _, table := range []string{"acquisition_items", "sale_items"} {
		exists, err := acquisitionTableExists(ctx, db, tx, table)
		if err != nil || !exists {
			return false, err
		}
	}
	return true, nil
}

func acquisitionTableExists(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx, table string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1)`
	if dbutil.IsSQLite(db) {
		query = `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type IN ('table','view') AND name=$1)`
	}
	var exists bool
	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &exists, query, table)
	} else {
		err = db.GetContext(ctx, &exists, query, table)
	}
	return exists, err
}

func (s *Service) PrepareInventoryItemAcquisitions(ctx context.Context, inventoryItemID, userID uuid.UUID) error {
	exists, err := acquisitionSaleLinkTablesExist(ctx, s.db, nil)
	if err != nil {
		return fmt.Errorf("check inventory item acquisition links: %w", err)
	}
	if !exists {
		return nil
	}
	var raw []string
	query := `SELECT DISTINCT CAST(si.sale_id AS TEXT) FROM acquisition_items ai JOIN sale_items si ON si.inventory_item_id=ai.inventory_item_id WHERE ai.inventory_item_id=? AND si.sale_id IS NOT NULL ORDER BY CAST(si.sale_id AS TEXT)`
	if err := s.db.SelectContext(ctx, &raw, s.db.Rebind(query), inventoryItemID.String()); err != nil {
		return fmt.Errorf("load inventory item acquisition sales: %w", err)
	}
	service := sales.NewSmartDeleteService(s.db)
	for _, value := range raw {
		id, err := uuid.Parse(value)
		if err != nil {
			return fmt.Errorf("parse inventory item acquisition sale %q: %w", value, err)
		}
		if err := service.PrepareDelete(ctx, id, userID); err != nil {
			return fmt.Errorf("prepare inventory item acquisition sale %s: %w", id, err)
		}
	}
	return nil
}

func (s *Service) DeleteAcquisitionsForProductTx(ctx context.Context, tx *sqlx.Tx, productID, userID uuid.UUID) error {
	exists, err := acquisitionTableExists(ctx, s.db, tx, "acquisition_items")
	if err != nil || !exists {
		return err
	}
	var items []struct {
		ID            string `db:"id"`
		AcquisitionID string `db:"acquisition_id"`
		InventoryID   string `db:"inventory_item_id"`
	}
	if err := tx.SelectContext(ctx, &items, tx.Rebind(`SELECT CAST(id AS TEXT) AS id, CAST(acquisition_id AS TEXT) AS acquisition_id, COALESCE(CAST(inventory_item_id AS TEXT),'') AS inventory_item_id FROM acquisition_items WHERE product_id=? ORDER BY created_at DESC,id DESC`), productID.String()); err != nil {
		return fmt.Errorf("load product acquisition items: %w", err)
	}
	for _, item := range items {
		if item.InventoryID != "" {
			var salesIDs []string
			if err := tx.SelectContext(ctx, &salesIDs, tx.Rebind(`SELECT DISTINCT CAST(sale_id AS TEXT) FROM sale_items WHERE inventory_item_id=? AND sale_id IS NOT NULL ORDER BY CAST(sale_id AS TEXT)`), item.InventoryID); err != nil {
				return fmt.Errorf("load acquisition item sales: %w", err)
			}
			for _, rawSaleID := range salesIDs {
				saleID, err := uuid.Parse(rawSaleID)
				if err != nil {
					return fmt.Errorf("parse acquisition sale id: %w", err)
				}
				result, err := sales.NewSmartDeleteService(s.db).SmartDeleteTx(ctx, tx, saleID, userID)
				if err != nil || result.Action != "deleted" {
					if err != nil {
						return fmt.Errorf("reverse acquisition item sale %s: %w", saleID, err)
					}
					return fmt.Errorf("acquisition item sale %s cannot be reversed: %s", saleID, result.Message)
				}
			}
			var status string
			if err := tx.GetContext(ctx, &status, tx.Rebind(`SELECT UPPER(TRIM(COALESCE(status,''))) FROM inventory_items WHERE id=?`), item.InventoryID); err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("load acquired inventory item: %w", err)
			} else if err == nil {
				if status == "SOLD" {
					return fmt.Errorf("acquisition inventory item %s remains sold after reversing linked sales", item.InventoryID)
				}
				if status == "AVAILABLE" || status == "RESERVED" {
					result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=COALESCE(quantity,0)-1, updated_at=CURRENT_TIMESTAMP WHERE product_id=? AND COALESCE(quantity,0)>0`), productID.String())
					if err != nil {
						return fmt.Errorf("reverse acquired product stock: %w", err)
					}
					if affected, _ := result.RowsAffected(); affected != 1 {
						return fmt.Errorf("cannot reconcile stock for acquired item %s", item.InventoryID)
					}
				}
				for _, dependency := range []struct{ table, column string }{
					{"inventory_movements", "item_id"}, {"item_history", "inventory_item_id"},
					{"barcodes", "inventory_item_id"}, {"reservations", "item_id"}, {"trade_ins", "inventory_item_id"},
					{"warranty_claims", "inventory_item_id"}, {"item_warranties", "inventory_item_id"},
					{"item_repair_costs", "inventory_item_id"}, {"inspections", "inventory_item_id"},
					{"item_specification_values", "inventory_item_id"},
				} {
					if err := deleteAcquisitionItemDependencyTx(ctx, tx, s.db, dependency.table, dependency.column, item.InventoryID); err != nil {
						return err
					}
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM inventory_items WHERE id=?`), item.InventoryID); err != nil {
					return fmt.Errorf("delete product acquisition inventory item: %w", err)
				}
			}
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM acquisition_items WHERE id=?`), item.ID); err != nil {
			return fmt.Errorf("delete product acquisition item: %w", err)
		}
		var remaining int
		if err := tx.GetContext(ctx, &remaining, tx.Rebind(`SELECT COUNT(*) FROM acquisition_items WHERE acquisition_id=?`), item.AcquisitionID); err != nil {
			return err
		}
		if remaining == 0 {
			for _, table := range []string{"seller_payments", "acquisition_reversals"} {
				if err := deleteAcquisitionItemDependencyTx(ctx, tx, s.db, table, "acquisition_id", item.AcquisitionID); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM acquisitions WHERE id=?`), item.AcquisitionID); err != nil {
				return fmt.Errorf("delete empty product acquisition: %w", err)
			}
		} else if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE acquisitions SET total_cost=COALESCE((SELECT SUM(total_cost) FROM acquisition_items WHERE acquisition_id=?),0), updated_at=CURRENT_TIMESTAMP WHERE id=?`), item.AcquisitionID, item.AcquisitionID); err != nil {
			return fmt.Errorf("recalculate acquisition total after product deletion: %w", err)
		}
	}
	return nil
}

// AcquisitionSaleIDsForOwner returns sales that consumed units recorded by
// acquisitions owned by the given supplier or customer.
func (s *Service) AcquisitionSaleIDsForOwner(ctx context.Context, ownerType string, ownerID uuid.UUID) ([]uuid.UUID, error) {
	ownerColumn := "supplier_id"
	if ownerType == TypeCustomer {
		ownerColumn = "customer_id"
	} else if ownerType != TypeSupplier {
		return nil, fmt.Errorf("invalid acquisition owner type %q", ownerType)
	}
	available, err := acquisitionSaleLinkTablesExist(ctx, s.db, nil)
	if err != nil || !available {
		return nil, err
	}
	query := fmt.Sprintf(`SELECT DISTINCT CAST(si.sale_id AS TEXT) FROM acquisitions a JOIN acquisition_items ai ON ai.acquisition_id=a.id JOIN sale_items si ON si.inventory_item_id=ai.inventory_item_id WHERE a.type=? AND a.%s=? AND si.sale_id IS NOT NULL ORDER BY CAST(si.sale_id AS TEXT)`, ownerColumn)
	var raw []string
	if err := s.db.SelectContext(ctx, &raw, query, ownerType, ownerID.String()); err != nil {
		return nil, fmt.Errorf("load acquisition-linked sales: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(raw))
	for _, value := range raw {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("parse acquisition-linked sale %q: %w", value, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// DeleteAcquisitionsForOwnerTx reverses all linked sale and stock effects,
// then hard-deletes acquisition rows inside the parent owner's transaction.
func (s *Service) DeleteAcquisitionsForOwnerTx(ctx context.Context, tx *sqlx.Tx, ownerType string, ownerID, userID uuid.UUID) error {
	ownerColumn := "supplier_id"
	if ownerType == TypeCustomer {
		ownerColumn = "customer_id"
	} else if ownerType != TypeSupplier {
		return fmt.Errorf("invalid acquisition owner type %q", ownerType)
	}
	exists, err := acquisitionTableExists(ctx, s.db, tx, "acquisitions")
	if err != nil || !exists {
		return err
	}
	var acquisitionIDs []string
	query := fmt.Sprintf(`SELECT CAST(id AS TEXT) FROM acquisitions WHERE type=? AND %s=? ORDER BY created_at DESC, id DESC`, ownerColumn)
	if err := tx.SelectContext(ctx, &acquisitionIDs, tx.Rebind(query), ownerType, ownerID.String()); err != nil {
		return fmt.Errorf("load owned acquisitions: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(acquisitionIDs))
	for _, rawID := range acquisitionIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return fmt.Errorf("parse acquisition id %q: %w", rawID, err)
		}
		ids = append(ids, id)
	}
	return s.DeleteAcquisitionsByIDsTx(ctx, tx, ids, userID)
}

// DeleteAcquisitionsByIDsTx reverses complete acquisition documents selected
// by their IDs. A caller deleting one acquired inventory item must remove its
// whole source document because paid totals and seller settlements belong to
// that document, not to a single child row.
func (s *Service) DeleteAcquisitionsByIDsTx(ctx context.Context, tx *sqlx.Tx, ids []uuid.UUID, userID uuid.UUID) error {
	for _, acquisitionID := range ids {
		rawID := acquisitionID.String()
		var saleIDs []string
		if err := tx.SelectContext(ctx, &saleIDs, tx.Rebind(`SELECT DISTINCT CAST(si.sale_id AS TEXT) FROM acquisition_items ai JOIN sale_items si ON si.inventory_item_id=ai.inventory_item_id WHERE ai.acquisition_id=? AND si.sale_id IS NOT NULL ORDER BY CAST(si.sale_id AS TEXT)`), rawID); err != nil {
			return fmt.Errorf("load sales linked to acquisition %s: %w", acquisitionID, err)
		}
		saleService := sales.NewSmartDeleteService(s.db)
		for _, rawSaleID := range saleIDs {
			saleID, err := uuid.Parse(rawSaleID)
			if err != nil {
				return fmt.Errorf("parse acquisition sale id %q: %w", rawSaleID, err)
			}
			result, err := saleService.SmartDeleteTx(ctx, tx, saleID, userID)
			if err != nil {
				return fmt.Errorf("reverse acquisition-linked sale %s: %w", saleID, err)
			}
			if result.Action != "deleted" {
				return fmt.Errorf("acquisition-linked sale %s could not be reversed: %s", saleID, result.Message)
			}
		}

		var items []struct {
			ID          string `db:"id"`
			ProductID   string `db:"product_id"`
			InventoryID string `db:"inventory_item_id"`
		}
		if err := tx.SelectContext(ctx, &items, tx.Rebind(`SELECT CAST(ai.id AS TEXT) AS id, CAST(ai.product_id AS TEXT) AS product_id, COALESCE(CAST(ai.inventory_item_id AS TEXT),'') AS inventory_item_id FROM acquisition_items ai WHERE ai.acquisition_id=? ORDER BY ai.id`), rawID); err != nil {
			return fmt.Errorf("load acquisition items: %w", err)
		}
		for _, item := range items {
			if item.InventoryID != "" {
				var status string
				if err := tx.GetContext(ctx, &status, tx.Rebind(`SELECT UPPER(TRIM(COALESCE(status,''))) FROM inventory_items WHERE id=?`), item.InventoryID); err != nil {
					if err != sql.ErrNoRows {
						return fmt.Errorf("load acquired item status: %w", err)
					}
				} else if status == "AVAILABLE" {
					result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=COALESCE(quantity,0)-1, updated_at=CURRENT_TIMESTAMP WHERE product_id=? AND COALESCE(quantity,0)>0`), item.ProductID)
					if err != nil {
						return fmt.Errorf("reverse acquisition aggregate stock: %w", err)
					}
					if affected, _ := result.RowsAffected(); affected != 1 {
						return fmt.Errorf("cannot reconcile stock for acquired inventory item %s", item.InventoryID)
					}
				} else if status != "REVERSED" {
					return fmt.Errorf("acquired inventory item %s remains %s after reversing its sales", item.InventoryID, status)
				}
				for _, dependency := range []struct{ table, column string }{
					{"inventory_movements", "item_id"}, {"item_history", "inventory_item_id"},
					{"barcodes", "inventory_item_id"}, {"reservations", "item_id"}, {"trade_ins", "inventory_item_id"},
					{"warranty_claims", "inventory_item_id"}, {"item_warranties", "inventory_item_id"},
					{"item_repair_costs", "inventory_item_id"}, {"inspections", "inventory_item_id"},
					{"item_specification_values", "inventory_item_id"},
				} {
					if err := deleteAcquisitionItemDependencyTx(ctx, tx, s.db, dependency.table, dependency.column, item.InventoryID); err != nil {
						return err
					}
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM inventory_items WHERE id=?`), item.InventoryID); err != nil {
					return fmt.Errorf("delete acquired inventory item: %w", err)
				}
			}
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM acquisition_items WHERE id=?`), item.ID); err != nil {
				return fmt.Errorf("delete acquisition item: %w", err)
			}
		}
		for _, table := range []string{"seller_payments", "acquisition_reversals"} {
			if err := deleteAcquisitionItemDependencyTx(ctx, tx, s.db, table, "acquisition_id", rawID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM acquisitions WHERE id=?`), rawID); err != nil {
			return fmt.Errorf("hard delete acquisition: %w", err)
		}
	}
	return nil
}

func deleteAcquisitionItemDependencyTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table, column, value string) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1)`
	if dbutil.IsSQLite(db) {
		query = `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name=$1)`
	}
	if err := tx.GetContext(ctx, &exists, query, table); err != nil || !exists {
		return err
	}
	columnQuery := `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name=$1 AND column_name=$2)`
	if dbutil.IsSQLite(db) {
		columnQuery = `SELECT EXISTS(SELECT 1 FROM pragma_table_info('` + table + `') WHERE name=$1)`
		if err := tx.GetContext(ctx, &exists, columnQuery, column); err != nil || !exists {
			return err
		}
	} else if err := tx.GetContext(ctx, &exists, columnQuery, table, column); err != nil || !exists {
		return err
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(fmt.Sprintf(`DELETE FROM %s WHERE %s=?`, table, column)), value); err != nil {
		return fmt.Errorf("delete acquisition dependency %s.%s: %w", table, column, err)
	}
	return nil
}
