package sales

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type saleCancellationMovement struct {
	ItemID    sql.NullString `db:"item_id"`
	ProductID sql.NullString `db:"product_id"`
	Quantity  int            `db:"quantity"`
}

func (s *Service) cancelSaleTransaction(ctx context.Context, userID, saleID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sale cancellation: %w", err)
	}
	defer tx.Rollback()

	isSQLite := dbutil.IsSQLite(s.db)
	saleIDArg := interface{}(saleID)
	if isSQLite {
		saleIDArg = saleID.String()
		// SQLite starts deferred transactions; take its write lock before reading
		// the sale so payments and cancellation cannot pass each other.
		result, err := tx.ExecContext(ctx, `UPDATE sales SET updated_at = updated_at WHERE id = ?`, saleIDArg)
		if err != nil {
			return fmt.Errorf("lock sale before cancellation: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return ErrSaleNotFound
		}
	}

	loadSale := `SELECT invoice_number, customer_id, status, payment_status, total_amount, COALESCE(paid_amount,0) AS paid_amount FROM sales WHERE id = ?`
	if !isSQLite {
		loadSale += ` FOR UPDATE`
	}
	var sale struct {
		InvoiceNumber string         `db:"invoice_number"`
		CustomerID    sql.NullString `db:"customer_id"`
		Status        string         `db:"status"`
		PaymentStatus string         `db:"payment_status"`
		TotalAmount   float64        `db:"total_amount"`
		PaidAmount    float64        `db:"paid_amount"`
	}
	if err := tx.GetContext(ctx, &sale, tx.Rebind(loadSale), saleIDArg); err != nil {
		if err == sql.ErrNoRows {
			return ErrSaleNotFound
		}
		return fmt.Errorf("load sale before cancellation: %w", err)
	}
	paymentStatus := strings.ToLower(strings.TrimSpace(sale.PaymentStatus))
	if !strings.EqualFold(strings.TrimSpace(sale.Status), "completed") || sale.PaidAmount > 0 || (paymentStatus != "" && paymentStatus != "unpaid" && paymentStatus != "pending" && paymentStatus != "debt") {
		return ErrInvalidSaleStatus
	}

	tables := NewSmartDeleteService(s.db)
	for _, table := range []string{"returns", "accounting_returns", "payment_transactions", "payments", "sale_payment_allocations"} {
		exists, err := tables.tableExists(ctx, tx, table)
		if err != nil {
			return fmt.Errorf("inspect %s before cancellation: %w", table, err)
		}
		if !exists {
			continue
		}
		var count int
		if err := tx.GetContext(ctx, &count, tx.Rebind(`SELECT COUNT(*) FROM `+table+` WHERE sale_id = ?`), saleIDArg); err != nil {
			return fmt.Errorf("check %s before cancellation: %w", table, err)
		}
		if count > 0 {
			return ErrInvalidSaleStatus
		}
	}

	for _, table := range []string{"debts", "customer_debts"} {
		exists, err := tables.tableExists(ctx, tx, table)
		if err != nil {
			return fmt.Errorf("inspect %s before cancellation: %w", table, err)
		}
		if !exists {
			continue
		}
		query := `SELECT COUNT(*) FROM ` + table + ` WHERE sale_id = ? AND COALESCE(paid_amount,0) > 0`
		if table == "customer_debts" {
			paidDebtCondition := `(COALESCE(paid_amount,0) > 0 OR COALESCE(is_paid,0) <> 0)`
			if !isSQLite {
				paidDebtCondition = `(COALESCE(paid_amount,0) > 0 OR COALESCE(is_paid,FALSE) = TRUE)`
			}
			query = `SELECT COUNT(*) FROM customer_debts WHERE reference_id = ? AND LOWER(COALESCE(reference_type,'')) = 'sale' AND ` + paidDebtCondition
		}
		var count int
		if err := tx.GetContext(ctx, &count, tx.Rebind(query), saleIDArg); err != nil {
			return fmt.Errorf("check collected %s before cancellation: %w", table, err)
		}
		if count > 0 {
			return ErrInvalidSaleStatus
		}
	}

	var expectedUnits int
	if err := tx.GetContext(ctx, &expectedUnits, tx.Rebind(`SELECT COALESCE(SUM(quantity),0) FROM sale_items WHERE sale_id = ?`), saleIDArg); err != nil {
		return fmt.Errorf("load sale item quantities: %w", err)
	}
	var movements []saleCancellationMovement
	if err := tx.SelectContext(ctx, &movements, tx.Rebind(`SELECT im.item_id, COALESCE(im.product_id,ii.product_id) AS product_id, im.quantity FROM inventory_movements im LEFT JOIN inventory_items ii ON ii.id=im.item_id WHERE LOWER(COALESCE(im.reference_type,''))='sale' AND im.reference_id=? AND UPPER(im.movement_type)='SALE'`), saleIDArg); err != nil {
		return fmt.Errorf("load sale stock movements: %w", err)
	}
	if expectedUnits > 0 && len(movements) == 0 {
		return fmt.Errorf("%w: sale has no verifiable stock movements", ErrInvalidSaleStatus)
	}
	actualUnits := 0
	for _, movement := range movements {
		if movement.Quantity >= 0 || !movement.ProductID.Valid || strings.TrimSpace(movement.ProductID.String) == "" {
			return fmt.Errorf("%w: invalid stock movement in sale", ErrInvalidSaleStatus)
		}
		reverseQuantity := -movement.Quantity
		actualUnits += reverseQuantity
		beforeQuantity, afterQuantity := 0, 0
		var itemArg interface{}
		if movement.ItemID.Valid && movement.ItemID.String != "" {
			if reverseQuantity != 1 {
				return fmt.Errorf("%w: tracked item movement quantity is invalid", ErrInvalidSaleStatus)
			}
			itemArg = movement.ItemID.String
			result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory_items SET status='AVAILABLE',sold_at=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=? AND UPPER(TRIM(COALESCE(status,'')))='SOLD'`), itemArg)
			if err != nil {
				return fmt.Errorf("restore sold item: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected != 1 {
				return fmt.Errorf("%w: sold item is no longer available to restore", ErrInvalidSaleStatus)
			}
			if exists, err := tables.tableExists(ctx, tx, "acquisition_items"); err != nil {
				return fmt.Errorf("inspect acquired item state: %w", err)
			} else if exists {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE acquisition_items SET item_status='available',updated_at=CURRENT_TIMESTAMP WHERE inventory_item_id=? AND LOWER(item_status)='sold'`), itemArg); err != nil {
					return fmt.Errorf("restore acquired item state: %w", err)
				}
			}
			if err := tx.GetContext(ctx, &afterQuantity, tx.Rebind(`SELECT COUNT(*) FROM inventory_items WHERE product_id=? AND UPPER(TRIM(COALESCE(status,'')))='AVAILABLE'`), movement.ProductID.String); err != nil {
				return fmt.Errorf("read available stock after restoring item: %w", err)
			}
			beforeQuantity = afterQuantity - 1
		} else {
			var exists bool
			if err := tx.GetContext(ctx, &exists, tx.Rebind(`SELECT EXISTS(SELECT 1 FROM inventory WHERE product_id=?)`), movement.ProductID.String); err != nil {
				return fmt.Errorf("check aggregate stock row: %w", err)
			}
			if exists {
				if err := tx.GetContext(ctx, &beforeQuantity, tx.Rebind(`SELECT COALESCE(quantity,0) FROM inventory WHERE product_id=?`), movement.ProductID.String); err != nil {
					return fmt.Errorf("read aggregate stock before restoring sale: %w", err)
				}
			}
			afterQuantity = beforeQuantity + reverseQuantity
		}

		stockResult, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=COALESCE(quantity,0)+?,updated_at=CURRENT_TIMESTAMP WHERE product_id=?`), reverseQuantity, movement.ProductID.String)
		if err != nil {
			return fmt.Errorf("restore aggregate stock: %w", err)
		}
		if affected, _ := stockResult.RowsAffected(); affected == 0 {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES (?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`), uuid.New().String(), movement.ProductID.String, reverseQuantity); err != nil {
				return fmt.Errorf("recreate aggregate stock row: %w", err)
			}
		}
		createdAt := interface{}(time.Now().UTC())
		if isSQLite {
			createdAt = time.Now().UTC().Format(time.RFC3339Nano)
		}
		reason := "Cancelled sale: " + sale.InvoiceNumber
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,reason,created_by,created_at) VALUES (?,?,?,'SALE_CANCELLATION',?,?,?,'sale_cancellation',?,?,?,?)`), uuid.New().String(), itemArg, movement.ProductID.String, reverseQuantity, beforeQuantity, afterQuantity, saleID.String(), reason, userID.String(), createdAt); err != nil {
			return fmt.Errorf("record sale cancellation stock movement: %w", err)
		}
		if itemArg != nil {
			if historyExists, err := tables.tableExists(ctx, tx, "item_history"); err != nil {
				return fmt.Errorf("inspect item history: %w", err)
			} else if historyExists {
				metadata, _ := json.Marshal(map[string]string{"sale_id": saleID.String(), "invoice_number": sale.InvoiceNumber})
				historyQuery := `INSERT INTO item_history (id,inventory_item_id,event_type,event_date,reference_type,reference_id,description,metadata,created_by,created_at) VALUES (?,?,'sale_cancelled',?,'sale_cancellation',?,?,?, ?,?)`
				if !isSQLite {
					historyQuery = `INSERT INTO item_history (id,inventory_item_id,event_type,event_date,reference_type,reference_id,description,metadata,created_by,created_at) VALUES (?,?,'sale_cancelled',?,'sale_cancellation',?,?,?::jsonb, ?,?)`
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(historyQuery), uuid.New().String(), movement.ItemID.String, createdAt, saleID.String(), reason, string(metadata), userID.String(), createdAt); err != nil {
					return fmt.Errorf("record restored item history: %w", err)
				}
			}
		}
	}
	if actualUnits != expectedUnits {
		return fmt.Errorf("%w: stock movement quantity %d does not match sale quantity %d", ErrInvalidSaleStatus, actualUnits, expectedUnits)
	}

	if sale.CustomerID.Valid {
		ledgerExists, err := tables.tableExists(ctx, tx, "customer_ledger")
		if err != nil {
			return fmt.Errorf("inspect customer ledger: %w", err)
		}
		if !ledgerExists {
			return fmt.Errorf("%w: customer ledger is unavailable for safe sale cancellation", ErrInvalidSaleStatus)
		}
		var saleDebit float64
			ledgerTypeExists, err := tables.columnExists(ctx, tx, "customer_ledger", "type")
			if err != nil {
				return fmt.Errorf("inspect customer ledger type: %w", err)
			}
			ledgerTransactionTypeExists, err := tables.columnExists(ctx, tx, "customer_ledger", "transaction_type")
			if err != nil {
				return fmt.Errorf("inspect customer ledger transaction type: %w", err)
			}
			if ledgerTypeExists {
				if err := tx.GetContext(ctx, &saleDebit, tx.Rebind(`SELECT COALESCE(SUM(amount),0) FROM customer_ledger WHERE customer_id=? AND reference_id=? AND LOWER(COALESCE(type,''))='debit'`), sale.CustomerID.String, saleIDArg); err != nil {
					return fmt.Errorf("read sale customer debit: %w", err)
				}
			} else if ledgerTransactionTypeExists {
				if err := tx.GetContext(ctx, &saleDebit, tx.Rebind(`SELECT COALESCE(SUM(amount),0) FROM customer_ledger WHERE customer_id=? AND reference_id=? AND UPPER(COALESCE(transaction_type,''))='SALE'`), sale.CustomerID.String, saleIDArg); err != nil {
					return fmt.Errorf("read sale customer debit: %w", err)
				}
			} else {
				return fmt.Errorf("%w: customer ledger has no supported transaction type", ErrInvalidSaleStatus)
			}
			if math.Abs(saleDebit-sale.TotalAmount) > 0.01 {
				return fmt.Errorf("%w: customer debt ledger does not match the unpaid sale total", ErrInvalidSaleStatus)
			}
			if saleDebit > 0 {
				var currentBalance float64
				if err := tx.GetContext(ctx, &currentBalance, tx.Rebind(`SELECT COALESCE(current_balance,0) FROM customers WHERE id=?`), sale.CustomerID.String); err != nil {
					return fmt.Errorf("read customer balance before sale cancellation: %w", err)
				}
				newBalance := currentBalance - saleDebit
				createdAt := interface{}(time.Now().UTC())
				if isSQLite {
					createdAt = time.Now().UTC().Format(time.RFC3339Nano)
				}
				if ledgerTypeExists {
					if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_ledger (id,customer_id,type,amount,balance,reference_id,description,created_at) VALUES (?,?,?,?,?,?,?,?)`), uuid.New().String(), sale.CustomerID.String, "credit", saleDebit, newBalance, saleID.String(), "Cancellation of sale "+sale.InvoiceNumber, createdAt); err != nil {
						return fmt.Errorf("record sale debit reversal: %w", err)
					}
				} else {
					if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO customer_ledger (id,customer_id,transaction_type,reference_id,amount,balance,description,created_by,created_at) VALUES (?,?,?,?,?,?,?,?,?)`), uuid.New().String(), sale.CustomerID.String, "ADJUSTMENT", saleID.String(), -saleDebit, newBalance, "Cancellation of sale "+sale.InvoiceNumber, userID.String(), createdAt); err != nil {
						return fmt.Errorf("record sale debit reversal: %w", err)
					}
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE customers SET current_balance=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`), newBalance, sale.CustomerID.String); err != nil {
					return fmt.Errorf("update customer balance after sale cancellation: %w", err)
				}
		}
	}

	if exists, err := tables.tableExists(ctx, tx, "debts"); err != nil {
		return fmt.Errorf("inspect debts: %w", err)
	} else if exists {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE debts SET remaining_amount=0,status='cancelled',updated_at=CURRENT_TIMESTAMP WHERE sale_id=? AND COALESCE(paid_amount,0)=0`), saleIDArg); err != nil {
			return fmt.Errorf("cancel unpaid sale debt: %w", err)
		}
	}
	if exists, err := tables.tableExists(ctx, tx, "customer_debts"); err != nil {
		return fmt.Errorf("inspect customer debts: %w", err)
	} else if exists {
		unpaidCondition := `COALESCE(paid_amount,0)=0 AND COALESCE(is_paid,0)=0`
		if !isSQLite {
			unpaidCondition = `COALESCE(paid_amount,0)=0 AND COALESCE(is_paid,FALSE)=FALSE`
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM customer_debts WHERE reference_id=? AND LOWER(COALESCE(reference_type,''))='sale' AND `+unpaidCondition), saleIDArg); err != nil {
			return fmt.Errorf("remove unpaid customer debt: %w", err)
		}
	}

	updatedAt := interface{}(time.Now().UTC())
	if isSQLite {
		updatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE sales SET status='cancelled',payment_status='cancelled',updated_at=? WHERE id=? AND LOWER(COALESCE(status,''))='completed' AND COALESCE(paid_amount,0)=0`), updatedAt, saleIDArg)
	if err != nil {
		return fmt.Errorf("mark sale cancelled: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrInvalidSaleStatus
	}

	if shiftsExist, err := tables.tableExists(ctx, tx, "pos_shifts"); err != nil {
		return fmt.Errorf("inspect POS shifts: %w", err)
	} else if shiftsExist {
		shiftQuery := `UPDATE pos_shifts AS sh SET
			sales_total=(SELECT COALESCE(SUM(s.total_amount),0) FROM sales s WHERE s.user_id=sh.user_id AND LOWER(COALESCE(s.status,'completed'))='completed' AND s.created_at>=sh.opened_at AND (sh.closed_at IS NULL OR s.created_at<sh.closed_at)),
			sale_count=(SELECT COUNT(*) FROM sales s WHERE s.user_id=sh.user_id AND LOWER(COALESCE(s.status,'completed'))='completed' AND s.created_at>=sh.opened_at AND (sh.closed_at IS NULL OR s.created_at<sh.closed_at))
			WHERE EXISTS (SELECT 1 FROM sales cancelled WHERE cancelled.id=? AND cancelled.user_id=sh.user_id AND cancelled.created_at>=sh.opened_at AND (sh.closed_at IS NULL OR cancelled.created_at<sh.closed_at))`
		if _, err := tx.ExecContext(ctx, tx.Rebind(shiftQuery), saleIDArg); err != nil {
			return fmt.Errorf("refresh POS shift after sale cancellation: %w", err)
		}
	}

	if exists, err := tables.tableExists(ctx, tx, "audit_logs"); err != nil {
		return fmt.Errorf("inspect audit log: %w", err)
	} else if exists {
		oldValues, _ := json.Marshal(map[string]interface{}{"status": sale.Status, "paid_amount": sale.PaidAmount})
		newValues, _ := json.Marshal(map[string]interface{}{"status": "cancelled", "stock_reversed": expectedUnits, "invoice_number": sale.InvoiceNumber})
		userArg := interface{}(userID.String())
		if userID == uuid.Nil {
			userArg = nil
		}
		auditQuery := `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,old_values,new_values,created_at) VALUES (?,?,?,?,?,?,?,?)`
		if !isSQLite {
			auditQuery = `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,old_values,new_values,created_at) VALUES (?,?,?,?,?,?::jsonb,?::jsonb,?)`
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(auditQuery), uuid.New().String(), userArg, "CANCEL", "sale", saleID.String(), string(oldValues), string(newValues), updatedAt); err != nil {
			return fmt.Errorf("record sale cancellation audit: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit sale cancellation: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("sale_cancelled")
	return nil
}
