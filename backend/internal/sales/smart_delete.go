package sales

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

// SmartDeleteService performs a physical, transactional sale deletion after
// reversing the sale's stock and customer-balance effects.
type SmartDeleteService struct {
	db *sqlx.DB
}

func NewSmartDeleteService(db *sqlx.DB) *SmartDeleteService {
	return &SmartDeleteService{db: db}
}

type DeleteResult struct {
	Action     string         `json:"action"`
	Message    string         `json:"message"`
	CanProceed bool           `json:"can_proceed"`
	Details    *DeleteDetails `json:"details,omitempty"`
}

type DeleteDetails struct {
	Reason          string `json:"reason"`
	SuggestedAction string `json:"suggested_action"`
}

func (s *SmartDeleteService) tableExists(ctx context.Context, tx *sqlx.Tx, table string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1 UNION ALL SELECT 1 FROM information_schema.views WHERE table_schema = current_schema() AND table_name = $1)`
	if dbutil.IsSQLite(s.db) {
		query = `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type IN ('table','view') AND name=$1)`
	}
	if err := tx.GetContext(ctx, &exists, query, table); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *SmartDeleteService) columnExists(ctx context.Context, tx *sqlx.Tx, table, column string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2)`
	if dbutil.IsSQLite(s.db) {
		// Table names here are internal constants; SQLite does not accept a bind
		// parameter as the argument to pragma_table_info on all supported versions.
		query = `SELECT EXISTS (SELECT 1 FROM pragma_table_info('` + table + `') WHERE name = $1)`
		err := tx.GetContext(ctx, &exists, query, column)
		return exists, err
	}
	err := tx.GetContext(ctx, &exists, query, table, column)
	return exists, err
}

func (s *SmartDeleteService) SmartDelete(ctx context.Context, saleID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin sale deletion: %w", err)
	}
	defer tx.Rollback()

	var sale struct {
		InvoiceNumber string         `db:"invoice_number"`
		CustomerID    sql.NullString `db:"customer_id"`
		Status        string         `db:"status"`
		PaymentStatus string         `db:"payment_status"`
		Total         float64        `db:"total_amount"`
		PaidAmount    float64        `db:"paid_amount"`
	}
	lock := ""
	if !dbutil.IsSQLite(s.db) {
		lock = " FOR UPDATE"
	} else {
		result, err := tx.ExecContext(ctx, `UPDATE sales SET updated_at = updated_at WHERE id = ?`, saleID.String())
		if err != nil {
			return nil, fmt.Errorf("lock sale before deletion: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return &DeleteResult{Action: "not_found", Message: "عملية البيع غير موجودة", CanProceed: false}, nil
		}
	}
	paidColumnExists, err := s.columnExists(ctx, tx, "sales", "paid_amount")
	if err != nil {
		return nil, fmt.Errorf("inspect sale payment amount: %w", err)
	}
	paidAmountExpr := `0 AS paid_amount`
	if paidColumnExists {
		paidAmountExpr = `COALESCE(paid_amount,0) AS paid_amount`
	}
	paymentStatusExpr := `'' AS payment_status`
	if paymentStatusColumnExists, err := s.columnExists(ctx, tx, "sales", "payment_status"); err != nil {
		return nil, fmt.Errorf("inspect sale payment status: %w", err)
	} else if paymentStatusColumnExists {
		paymentStatusExpr = `COALESCE(payment_status,'') AS payment_status`
	}
	if err := tx.GetContext(ctx, &sale, tx.Rebind(`SELECT invoice_number, customer_id, status, `+paymentStatusExpr+`, total_amount, `+paidAmountExpr+` FROM sales WHERE id = ?`)+lock, saleID.String()); err != nil {
		if err == sql.ErrNoRows {
			return &DeleteResult{Action: "not_found", Message: "عملية البيع غير موجودة", CanProceed: false}, nil
		}
		return nil, fmt.Errorf("load sale for deletion: %w", err)
	}
	paymentStatus := strings.ToLower(strings.TrimSpace(sale.PaymentStatus))
	if sale.PaidAmount > 0 || paymentStatus == "paid" || paymentStatus == "partial" || paymentStatus == "refunded" {
		return &DeleteResult{Action: "blocked", Message: "لا يمكن حذف بيع سُجل عليه تحصيل", CanProceed: false,
			Details: &DeleteDetails{Reason: "يوجد مبلغ مدفوع مسجل على الفاتورة", SuggestedAction: "استخدم مسار عكس التحصيلات قبل حذف البيع"}}, nil
	}

	for _, table := range []string{"returns", "accounting_returns", "payment_transactions"} {
		exists, tableErr := s.tableExists(ctx, tx, table)
		if tableErr != nil {
			return nil, fmt.Errorf("inspect %s dependencies: %w", table, tableErr)
		}
		if !exists {
			continue
		}
		var dependencyCount int
		if err := tx.GetContext(ctx, &dependencyCount, tx.Rebind(`SELECT COUNT(*) FROM `+table+` WHERE sale_id = ?`), saleID.String()); err != nil {
			return nil, fmt.Errorf("check %s dependencies: %w", table, err)
		}
		if dependencyCount > 0 {
			return &DeleteResult{
				Action: "blocked", Message: "لا يمكن حذف البيع قبل معالجة المرتجعات أو رد المدفوعات المرتبطة به", CanProceed: false,
				Details: &DeleteDetails{Reason: "توجد عمليات لاحقة مرتبطة بالبيع", SuggestedAction: "احذف أو اعكس العمليات المرتبطة أولاً"},
			}, nil
		}
	}

	debtsExist, err := s.tableExists(ctx, tx, "debts")
	if err != nil {
		return nil, fmt.Errorf("inspect debt dependencies: %w", err)
	}
	if debtsExist {
		var collectedDebtPayments int
		if err := tx.GetContext(ctx, &collectedDebtPayments, tx.Rebind(`SELECT COUNT(*) FROM debts WHERE sale_id = ? AND COALESCE(paid_amount, 0) > 0`), saleID.String()); err != nil {
			return nil, fmt.Errorf("check collected sale debt: %w", err)
		}
		if collectedDebtPayments > 0 {
			return &DeleteResult{
				Action: "blocked", Message: "لا يمكن حذف بيع سُدد جزء من دينه", CanProceed: false,
				Details: &DeleteDetails{Reason: "تم تسجيل تحصيلات على دين البيع", SuggestedAction: "اعكس التحصيلات المرتبطة أولاً"},
			}, nil
		}
	}

	salePaymentsExist, err := s.tableExists(ctx, tx, "payments")
	if err != nil {
		return nil, fmt.Errorf("inspect sale payments: %w", err)
	}
	if salePaymentsExist {
		var paymentCount int
		if err := tx.GetContext(ctx, &paymentCount, tx.Rebind(`SELECT COUNT(*) FROM payments WHERE sale_id=?`), saleID.String()); err != nil {
			return nil, fmt.Errorf("check sale payments: %w", err)
		}
		if paymentCount > 0 {
			return &DeleteResult{Action: "blocked", Message: "لا يمكن حذف بيع له دفعات مسجلة", CanProceed: false,
				Details: &DeleteDetails{Reason: "توجد دفعات مرتبطة بالفاتورة", SuggestedAction: "عالج الدفعات أو اعكسها قبل حذف البيع"}}, nil
		}
	}

	itemsTable, err := s.tableExists(ctx, tx, "sale_items")
	if err != nil || !itemsTable {
		if err == nil {
			err = fmt.Errorf("sale_items table is missing")
		}
		return nil, fmt.Errorf("inspect sale items: %w", err)
	}
	var itemCount int
	if err := tx.GetContext(ctx, &itemCount, tx.Rebind(`SELECT COUNT(*) FROM sale_items WHERE sale_id = ?`), saleID.String()); err != nil {
		return nil, fmt.Errorf("count sale items: %w", err)
	}
	var expectedStockUnits int
	if err := tx.GetContext(ctx, &expectedStockUnits, tx.Rebind(`SELECT COALESCE(SUM(quantity),0) FROM sale_items WHERE sale_id = ?`), saleID.String()); err != nil {
		return nil, fmt.Errorf("sum sale item quantities: %w", err)
	}

	stockDeltas := make(map[string]int)
	movementExists, err := s.tableExists(ctx, tx, "inventory_movements")
	if err != nil {
		return nil, fmt.Errorf("inspect inventory movement history: %w", err)
	}
	if movementExists {
		var movements []struct {
			ItemID    sql.NullString `db:"item_id"`
			ProductID sql.NullString `db:"product_id"`
			Quantity  int            `db:"quantity"`
		}
		query := `SELECT im.item_id, COALESCE(im.product_id, ii.product_id) AS product_id, im.quantity
			FROM inventory_movements im LEFT JOIN inventory_items ii ON ii.id = im.item_id
			WHERE LOWER(COALESCE(im.reference_type,''))='sale' AND im.reference_id=$1 AND UPPER(im.movement_type)='SALE'`
		if err := tx.SelectContext(ctx, &movements, query, saleID.String()); err != nil {
			return nil, fmt.Errorf("load sale inventory movements: %w", err)
		}
		if itemCount > 0 && len(movements) == 0 {
			return &DeleteResult{
				Action: "blocked", Message: "تعذر عكس المخزون لأن سجل حركة البيع غير موجود", CanProceed: false,
				Details: &DeleteDetails{Reason: "لا توجد حركة مخزون موثوقة مرتبطة بالبيع", SuggestedAction: "راجع المخزون قبل حذف الفاتورة"},
			}, nil
		}
		actualStockUnits := 0
		for _, movement := range movements {
			if movement.Quantity >= 0 {
				return nil, fmt.Errorf("invalid sale movement quantity %d", movement.Quantity)
			}
			actualStockUnits += -movement.Quantity
			if movement.ProductID.Valid {
				stockDeltas[movement.ProductID.String] += -movement.Quantity
			}
			if movement.ItemID.Valid {
				result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory_items SET status='AVAILABLE', sold_at=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=? AND UPPER(TRIM(COALESCE(status,'')))='SOLD'`), movement.ItemID.String)
				if err != nil {
					return nil, fmt.Errorf("restore sold inventory item: %w", err)
				}
				if affected, _ := result.RowsAffected(); affected != 1 {
					return nil, fmt.Errorf("inventory item %s is no longer in SOLD state; sale deletion rolled back", movement.ItemID.String)
				}
				if exists, _ := s.tableExists(ctx, tx, "acquisition_items"); exists {
					if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE acquisition_items SET item_status='available', updated_at=CURRENT_TIMESTAMP WHERE inventory_item_id=? AND LOWER(item_status)='sold'`), movement.ItemID.String); err != nil {
						return nil, fmt.Errorf("restore acquired item state: %w", err)
					}
				}
			}
		}
		if actualStockUnits != expectedStockUnits {
			return &DeleteResult{
				Action: "blocked", Message: "لا يمكن حذف البيع لأن كميات حركات المخزون لا تطابق سطور الفاتورة", CanProceed: false,
				Details: &DeleteDetails{Reason: "يوجد نقص أو زيادة في حركات المخزون المرتبطة", SuggestedAction: "طابق حركات المخزون مع كميات الفاتورة قبل الحذف"},
			}, nil
		}
	} else if itemCount > 0 {
		return &DeleteResult{Action: "blocked", Message: "تعذر حذف البيع لعدم توفر سجل حركات المخزون", CanProceed: false}, nil
	}

	inventoryExists, err := s.tableExists(ctx, tx, "inventory")
	if err != nil {
		return nil, fmt.Errorf("inspect aggregate inventory: %w", err)
	}
	if inventoryExists {
		for productID, delta := range stockDeltas {
			result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=COALESCE(quantity,0)+?, updated_at=CURRENT_TIMESTAMP WHERE product_id=?`), delta, productID)
			if err != nil {
				return nil, fmt.Errorf("restore aggregate inventory: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected == 0 {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`), uuid.New().String(), productID, delta); err != nil {
					return nil, fmt.Errorf("recreate aggregate inventory row: %w", err)
				}
			}
		}
	}

	paymentsExist, err := s.tableExists(ctx, tx, "payments")
	if err != nil {
		return nil, fmt.Errorf("inspect sale payments: %w", err)
	}
	if paymentsExist {
		if auditExists, err := s.tableExists(ctx, tx, "audit_logs"); err != nil {
			return nil, err
		} else if auditExists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM audit_logs WHERE entity_id IN (SELECT id FROM payments WHERE sale_id=?)`), saleID.String()); err != nil {
				return nil, fmt.Errorf("remove sale payment audit rows: %w", err)
			}
		}
		if ledgerExists, err := s.tableExists(ctx, tx, "customer_ledger"); err != nil {
			return nil, fmt.Errorf("inspect customer ledger: %w", err)
		} else if ledgerExists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM customer_ledger WHERE reference_id=? OR reference_id IN (SELECT id FROM payments WHERE sale_id=?)`), saleID.String(), saleID.String()); err != nil {
				return nil, fmt.Errorf("remove sale customer ledger rows: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM payments WHERE sale_id=?`), saleID.String()); err != nil {
			return nil, fmt.Errorf("delete sale payments: %w", err)
		}
	} else if ledgerExists, err := s.tableExists(ctx, tx, "customer_ledger"); err != nil {
		return nil, err
	} else if ledgerExists {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM customer_ledger WHERE reference_id=?`), saleID.String()); err != nil {
			return nil, fmt.Errorf("remove sale customer ledger rows: %w", err)
		}
	}

	for _, dependency := range []struct{ table, query string }{
		{"debts", `DELETE FROM debts WHERE sale_id=?`},
		{"customer_debts", `DELETE FROM customer_debts WHERE reference_id=? AND LOWER(COALESCE(reference_type,''))='sale'`},
		{"sale_payment_allocations", `DELETE FROM sale_payment_allocations WHERE sale_id=?`},
		{"item_history", `DELETE FROM item_history WHERE reference_id=? AND LOWER(COALESCE(reference_type,''))='sale'`},
		{"inventory_movements", `DELETE FROM inventory_movements WHERE reference_id=? AND LOWER(COALESCE(reference_type,''))='sale'`},
		{"ledger_entries", `DELETE FROM ledger_entries WHERE reference_id=? AND LOWER(COALESCE(reference_type,''))='sale'`},
	} {
		exists, err := s.tableExists(ctx, tx, dependency.table)
		if err != nil {
			return nil, fmt.Errorf("inspect %s: %w", dependency.table, err)
		}
		if exists {
			if dependency.table == "inventory_movements" || dependency.table == "item_history" || dependency.table == "ledger_entries" {
				if auditExists, err := s.tableExists(ctx, tx, "audit_logs"); err != nil {
					return nil, err
				} else if auditExists {
					if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM audit_logs WHERE entity_id IN (SELECT id FROM `+dependency.table+` WHERE reference_id=? AND LOWER(COALESCE(reference_type,''))='sale')`), saleID.String()); err != nil {
						return nil, fmt.Errorf("remove sale %s audit rows: %w", dependency.table, err)
					}
				}
			}
			if _, err := tx.ExecContext(ctx, tx.Rebind(dependency.query), saleID.String()); err != nil {
				return nil, fmt.Errorf("delete sale dependent %s: %w", dependency.table, err)
			}
		}
	}

	if sale.CustomerID.Valid {
		if exists, err := s.tableExists(ctx, tx, "customers"); err != nil {
			return nil, err
		} else if exists {
			if ledgerExists, err := s.tableExists(ctx, tx, "customer_ledger"); err != nil {
				return nil, err
			} else if ledgerExists {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE customers SET current_balance=COALESCE((SELECT SUM(CASE WHEN LOWER(type)='debit' THEN amount ELSE -amount END) FROM customer_ledger WHERE customer_id=?),0), updated_at=CURRENT_TIMESTAMP WHERE id=?`), sale.CustomerID.String, sale.CustomerID.String); err != nil {
					return nil, fmt.Errorf("recalculate customer balance: %w", err)
				}
			}
		}
	}

	for _, target := range []struct{ table, query string }{
		{"audit_logs", `DELETE FROM audit_logs WHERE entity_id=? AND LOWER(entity_type) IN ('sale','sales')`},
		{"audit_logs", `DELETE FROM audit_logs WHERE entity_id IN (SELECT id FROM sale_items WHERE sale_id=?)`},
		{"sale_items", `DELETE FROM sale_items WHERE sale_id=?`},
		{"sales", `DELETE FROM sales WHERE id=?`},
	} {
		exists, err := s.tableExists(ctx, tx, target.table)
		if err != nil {
			return nil, err
		}
		if exists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(target.query), saleID.String()); err != nil {
				return nil, fmt.Errorf("delete sale %s: %w", target.table, err)
			}
		}
	}

	if exists, err := s.tableExists(ctx, tx, "audit_logs"); err != nil {
		return nil, err
	} else if exists {
		snapshot, _ := json.Marshal(map[string]interface{}{"invoice_number": sale.InvoiceNumber, "total_amount": sale.Total, "deleted_at": time.Now().UTC()})
		userArg := interface{}(userID.String())
		if userID == uuid.Nil {
			userArg = nil
		}
		query := `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES (?,?, 'DELETE','sale',?,?,CURRENT_TIMESTAMP)`
		if _, err := tx.ExecContext(ctx, tx.Rebind(query), uuid.New().String(), userArg, saleID.String(), string(snapshot)); err != nil {
			return nil, fmt.Errorf("write sale deletion audit: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit sale deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("sale_deleted")
	return &DeleteResult{Action: "deleted", Message: "تم حذف عملية البيع وعكس آثار المخزون والرصيد", CanProceed: true}, nil
}

func (s *Service) DeleteSale(ctx context.Context, saleID, userID uuid.UUID) (*DeleteResult, error) {
	return NewSmartDeleteService(s.db).SmartDelete(ctx, saleID, userID)
}
