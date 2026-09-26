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
	"github.com/partflow/smart-store/internal/payments"
	"github.com/partflow/smart-store/internal/paymenttransactions"
	returnrepo "github.com/partflow/smart-store/internal/returns"
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

type SalesCleanupSummary struct {
	Total        int  `json:"total"`
	Deleted      int  `json:"deleted"`
	Blocked      int  `json:"blocked"`
	Failed       int  `json:"failed"`
	StoppedEarly bool `json:"stopped_early"`
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
	deletionOrder, err := s.PrepareCascade(ctx, saleID, userID)
	if err != nil {
		return nil, fmt.Errorf("prepare sale cascade: %w", err)
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin sale deletion: %w", err)
	}
	defer tx.Rollback()
	for _, dependentSaleID := range deletionOrder[:len(deletionOrder)-1] {
		if err := s.deleteCascadeEntryTx(ctx, tx, dependentSaleID, userID); err != nil {
			return nil, fmt.Errorf("delete dependent sale %s before %s: %w", dependentSaleID, saleID, err)
		}
	}
	result, err := s.smartDeleteTx(ctx, tx, saleID, userID)
	if err != nil {
		return nil, err
	}
	if result.Action != "deleted" {
		_ = tx.Rollback()
		return result, nil
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit sale deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("sale_deleted")
	return result, nil
}

// SmartDeleteTx performs a sale deletion in a caller-owned transaction after
// its caller has reconciled any external payment provider effects.
func (s *SmartDeleteService) SmartDeleteTx(ctx context.Context, tx *sqlx.Tx, saleID, userID uuid.UUID) (*DeleteResult, error) {
	return s.smartDeleteTx(ctx, tx, saleID, userID)
}

// DeleteInTransaction adapts transactional sale deletion for higher-level
// cascades such as removing a return whose restored item was sold later.
func (s *SmartDeleteService) DeleteInTransaction(ctx context.Context, tx *sqlx.Tx, saleID, userID uuid.UUID) error {
	return s.DeleteCascadeInTransaction(ctx, tx, saleID, userID)
}

func (s *SmartDeleteService) PrepareDelete(ctx context.Context, saleID, userID uuid.UUID) error {
	_, err := s.PrepareCascade(ctx, saleID, userID)
	return err
}

// PrepareCascade resolves downstream sales restored by returns and prepares
// every sale's external payment effects before a caller starts its transaction.
// The returned order is deepest dependent first and always ends with saleID.
func (s *SmartDeleteService) PrepareCascade(ctx context.Context, saleID, userID uuid.UUID) ([]uuid.UUID, error) {
	order, err := s.planDependentSalesForSale(ctx, saleID)
	if err != nil {
		return nil, err
	}
	order = append(order, saleID)
	for _, id := range order {
		if err := s.prepareLinkedPayments(ctx, id, userID); err != nil {
			return nil, fmt.Errorf("prepare payment reversal for sale %s: %w", id, err)
		}
	}
	return order, nil
}

// DeleteCascadeInTransaction reverses all downstream sales and the root sale
// atomically. Call PrepareCascade before opening the transaction.
func (s *SmartDeleteService) DeleteCascadeInTransaction(ctx context.Context, tx *sqlx.Tx, saleID, userID uuid.UUID) error {
	order, err := s.planDependentSalesForSaleTx(ctx, tx, saleID)
	if err != nil {
		return err
	}
	for _, dependentID := range order {
		if err := s.deleteCascadeEntryTx(ctx, tx, dependentID, userID); err != nil {
			return fmt.Errorf("delete dependent sale %s: %w", dependentID, err)
		}
	}
	return s.deleteCascadeEntryTx(ctx, tx, saleID, userID)
}

func (s *SmartDeleteService) planDependentSalesForSaleTx(ctx context.Context, tx *sqlx.Tx, saleID uuid.UUID) ([]uuid.UUID, error) {
	rootReturns, err := s.saleReturnIDsTx(ctx, tx, saleID)
	if err != nil {
		return nil, err
	}
	visited := make(map[uuid.UUID]bool)
	visiting := make(map[uuid.UUID]bool)
	ordered := make([]uuid.UUID, 0)
	var visitSale func(uuid.UUID) error
	visitSale = func(currentSaleID uuid.UUID) error {
		if visited[currentSaleID] {
			return nil
		}
		if visiting[currentSaleID] {
			return fmt.Errorf("cyclic sale/return inventory dependency at sale %s", currentSaleID)
		}
		visiting[currentSaleID] = true
		returnIDs, err := s.saleReturnIDsTx(ctx, tx, currentSaleID)
		if err != nil {
			return err
		}
		for _, returnID := range returnIDs {
			children, err := s.dependentSalesForReturnTx(ctx, tx, returnID)
			if err != nil {
				return err
			}
			for _, childID := range children {
				if err := visitSale(childID); err != nil {
					return err
				}
			}
		}
		delete(visiting, currentSaleID)
		visited[currentSaleID] = true
		if currentSaleID != saleID {
			ordered = append(ordered, currentSaleID)
		}
		return nil
	}
	for _, returnID := range rootReturns {
		children, err := s.dependentSalesForReturnTx(ctx, tx, returnID)
		if err != nil {
			return nil, err
		}
		for _, childID := range children {
			if err := visitSale(childID); err != nil {
				return nil, err
			}
		}
	}
	return ordered, nil
}

func (s *SmartDeleteService) saleReturnIDsTx(ctx context.Context, tx *sqlx.Tx, saleID uuid.UUID) ([]uuid.UUID, error) {
	exists, err := s.tableExists(ctx, tx, "returns")
	if err != nil || !exists {
		return nil, err
	}
	var rawIDs []string
	if err := tx.SelectContext(ctx, &rawIDs, tx.Rebind(`SELECT CAST(id AS TEXT) FROM returns WHERE sale_id=? ORDER BY id`), saleID.String()); err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("invalid linked return id %q", rawID)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *SmartDeleteService) dependentSalesForReturnTx(ctx context.Context, tx *sqlx.Tx, returnID uuid.UUID) ([]uuid.UUID, error) {
	for _, table := range []string{"return_items", "inventory_movements"} {
		exists, err := s.tableExists(ctx, tx, table)
		if err != nil || !exists {
			return nil, err
		}
	}
	hasItemLink, err := s.columnExists(ctx, tx, "return_items", "inventory_item_id")
	if err != nil || !hasItemLink {
		return nil, err
	}
	query := `SELECT DISTINCT CAST(m.reference_id AS TEXT) FROM returns r JOIN return_items ri ON ri.return_id=r.id JOIN inventory_movements m ON m.item_id=ri.inventory_item_id WHERE r.id=? AND UPPER(COALESCE(m.movement_type,''))='SALE' AND LOWER(COALESCE(m.reference_type,''))='sale' AND CAST(m.reference_id AS TEXT)<>CAST(r.sale_id AS TEXT) ORDER BY CAST(m.reference_id AS TEXT)`
	var rawIDs []string
	if err := tx.SelectContext(ctx, &rawIDs, tx.Rebind(query), returnID.String()); err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("invalid dependent sale id %q", rawID)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *SmartDeleteService) deleteCascadeEntryTx(ctx context.Context, tx *sqlx.Tx, saleID, userID uuid.UUID) error {
	var exists bool
	if err := tx.GetContext(ctx, &exists, tx.Rebind(`SELECT EXISTS(SELECT 1 FROM sales WHERE id=?)`), saleID.String()); err != nil {
		return fmt.Errorf("check sale %s before cascade deletion: %w", saleID, err)
	}
	if !exists {
		// A prior root in the same parent cascade may already have removed this
		// shared descendant. Hard-delete cascades are idempotent for that case.
		return nil
	}
	result, err := s.smartDeleteTx(ctx, tx, saleID, userID)
	if err != nil {
		return err
	}
	if result == nil || result.Action != "deleted" || !result.CanProceed {
		if result == nil {
			return fmt.Errorf("sale %s did not return a deletion result", saleID)
		}
		return fmt.Errorf("sale %s could not be deleted: %s", saleID, result.Message)
	}
	return nil
}

func (s *SmartDeleteService) planDependentSalesForSale(ctx context.Context, saleID uuid.UUID) ([]uuid.UUID, error) {
	rootReturns, err := s.saleReturnIDs(ctx, saleID)
	if err != nil {
		return nil, err
	}
	visited := make(map[uuid.UUID]bool)
	visiting := make(map[uuid.UUID]bool)
	ordered := make([]uuid.UUID, 0)
	var visitSale func(uuid.UUID) error
	visitSale = func(currentSaleID uuid.UUID) error {
		if visited[currentSaleID] {
			return nil
		}
		if visiting[currentSaleID] {
			return fmt.Errorf("cyclic sale/return inventory dependency at sale %s", currentSaleID)
		}
		visiting[currentSaleID] = true
		returnIDs, err := s.saleReturnIDs(ctx, currentSaleID)
		if err != nil {
			return err
		}
		for _, returnID := range returnIDs {
			children, err := s.dependentSalesForReturn(ctx, returnID)
			if err != nil {
				return err
			}
			for _, childID := range children {
				if err := visitSale(childID); err != nil {
					return err
				}
			}
		}
		delete(visiting, currentSaleID)
		visited[currentSaleID] = true
		if currentSaleID != saleID {
			ordered = append(ordered, currentSaleID)
		}
		return nil
	}
	for _, returnID := range rootReturns {
		children, err := s.dependentSalesForReturn(ctx, returnID)
		if err != nil {
			return nil, err
		}
		for _, childID := range children {
			if err := visitSale(childID); err != nil {
				return nil, err
			}
		}
	}
	return ordered, nil
}

func (s *SmartDeleteService) saleReturnIDs(ctx context.Context, saleID uuid.UUID) ([]uuid.UUID, error) {
	exists, err := s.dbTableExists(ctx, "returns")
	if err != nil || !exists {
		return nil, err
	}
	var rawIDs []string
	if err := s.db.SelectContext(ctx, &rawIDs, s.db.Rebind(`SELECT CAST(id AS TEXT) FROM returns WHERE sale_id=? ORDER BY id`), saleID.String()); err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("invalid linked return id %q", rawID)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *SmartDeleteService) dependentSalesForReturn(ctx context.Context, returnID uuid.UUID) ([]uuid.UUID, error) {
	for _, table := range []string{"return_items", "inventory_movements"} {
		exists, err := s.dbTableExists(ctx, table)
		if err != nil || !exists {
			return nil, err
		}
	}
	hasItemLink, err := s.dbColumnExists(ctx, "return_items", "inventory_item_id")
	if err != nil || !hasItemLink {
		return nil, err
	}
	query := `SELECT DISTINCT CAST(m.reference_id AS TEXT) FROM returns r JOIN return_items ri ON ri.return_id=r.id JOIN inventory_movements m ON m.item_id=ri.inventory_item_id WHERE r.id=? AND UPPER(COALESCE(m.movement_type,''))='SALE' AND LOWER(COALESCE(m.reference_type,''))='sale' AND CAST(m.reference_id AS TEXT)<>CAST(r.sale_id AS TEXT) ORDER BY CAST(m.reference_id AS TEXT)`
	var rawIDs []string
	if err := s.db.SelectContext(ctx, &rawIDs, s.db.Rebind(query), returnID.String()); err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("invalid dependent sale id %q", rawID)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *SmartDeleteService) dbTableExists(ctx context.Context, table string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1)`
	if dbutil.IsSQLite(s.db) {
		query = `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type IN ('table','view') AND name=$1)`
	}
	if err := s.db.GetContext(ctx, &exists, query, table); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *SmartDeleteService) dbColumnExists(ctx context.Context, table, column string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name=$1 AND column_name=$2)`
	args := []any{table, column}
	if dbutil.IsSQLite(s.db) {
		query = `SELECT EXISTS(SELECT 1 FROM pragma_table_info('` + table + `') WHERE name=$1)`
		args = []any{column}
	}
	if err := s.db.GetContext(ctx, &exists, query, args...); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *SmartDeleteService) smartDeleteTx(ctx context.Context, tx *sqlx.Tx, saleID, userID uuid.UUID) (*DeleteResult, error) {
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
			return &DeleteResult{Action: "not_found", Message: "Ø¹Ù…Ù„ÙŠØ© Ø§Ù„Ø¨ÙŠØ¹ ØºÙŠØ± Ù…ÙˆØ¬ÙˆØ¯Ø©", CanProceed: false}, nil
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
	if err := tx.GetContext(ctx, &sale, tx.Rebind(`SELECT COALESCE(invoice_number,'') AS invoice_number, customer_id, status, `+paymentStatusExpr+`, total_amount, `+paidAmountExpr+` FROM sales WHERE id = ?`)+lock, saleID.String()); err != nil {
		if err == sql.ErrNoRows {
			return &DeleteResult{Action: "not_found", Message: "Ø¹Ù…Ù„ÙŠØ© Ø§Ù„Ø¨ÙŠØ¹ ØºÙŠØ± Ù…ÙˆØ¬ÙˆØ¯Ø©", CanProceed: false}, nil
		}
		return nil, fmt.Errorf("load sale for deletion: %w", err)
	}
	// Reverse and remove returns before the parent sale. The return repository
	// owns the stock/debt/ledger logic; running it on this transaction keeps the
	// whole sale cascade atomic. accounting_returns is a reporting view in
	// current schemas and is rebuilt from these source rows.
	returnsExist, err := s.tableExists(ctx, tx, "returns")
	if err != nil {
		return nil, fmt.Errorf("inspect sale returns: %w", err)
	}
	if returnsExist {
		var linkedReturns []string
		if err := tx.SelectContext(ctx, &linkedReturns, tx.Rebind(`SELECT id FROM returns WHERE sale_id=? ORDER BY id`), saleID.String()); err != nil {
			return nil, fmt.Errorf("load sale returns for reversal: %w", err)
		}
		returnRepository := returnrepo.NewRepository(s.db)
		for _, rawID := range linkedReturns {
			returnID, err := uuid.Parse(rawID)
			if err != nil {
				return nil, fmt.Errorf("parse linked return id %q: %w", rawID, err)
			}
			if err := returnRepository.DeleteReturnTx(ctx, tx, returnID); err != nil {
				return nil, fmt.Errorf("reverse linked return %s before deleting sale: %w", returnID, err)
			}
		}
	}

	if debtsExist, err := s.tableExists(ctx, tx, "debts"); err != nil {
		return nil, err
	} else if debtsExist {
		if hasCustomerID, err := s.columnExists(ctx, tx, "debts", "customer_id"); err != nil {
			return nil, err
		} else if hasCustomerID {
			var customerIDs []string
			if err := tx.SelectContext(ctx, &customerIDs, tx.Rebind(`SELECT DISTINCT CAST(customer_id AS TEXT) FROM debts WHERE sale_id=? AND customer_id IS NOT NULL AND COALESCE(paid_amount,0)>0.000001`), saleID.String()); err != nil {
				return nil, fmt.Errorf("load customers with payments on this sale: %w", err)
			}
			paymentRepo := payments.NewRepository(s.db)
			for _, rawCustomerID := range customerIDs {
				customerID, err := uuid.Parse(rawCustomerID)
				if err != nil {
					return nil, ErrSaleDebtAllocationHistoryInconsistent
				}
				if err := paymentRepo.ReconstructLegacyCustomerAllocationsTx(ctx, tx, customerID); err != nil {
					return nil, fmt.Errorf("reconstruct debt payment allocations for customer %s: %w", customerID, err)
				}
			}
		}
	}
	paymentIDs, err := linkedSalePayments(ctx, tx, s, saleID)
	if err != nil {
		return nil, fmt.Errorf("load sale payment reversals: %w", err)
	}
	paymentRepo := payments.NewRepository(s.db)
	for _, paymentID := range paymentIDs {
		if err := paymentRepo.DeleteInTx(ctx, tx, paymentID); err != nil {
			return nil, fmt.Errorf("reverse payment %s before deleting sale: %w", paymentID, err)
		}
		if auditExists, err := s.tableExists(ctx, tx, "audit_logs"); err != nil {
			return nil, err
		} else if auditExists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM audit_logs WHERE entity_id=?`), paymentID.String()); err != nil {
				return nil, fmt.Errorf("remove deleted payment audit rows: %w", err)
			}
		}
	}
	var remainingPaid float64
	if paidColumnExists {
		if err := tx.GetContext(ctx, &remainingPaid, tx.Rebind(`SELECT COALESCE(paid_amount,0) FROM sales WHERE id=?`), saleID.String()); err != nil {
			return nil, fmt.Errorf("verify sale payments were reversed: %w", err)
		}
	}
	if remainingPaid > 0.000001 && len(paymentIDs) == 0 {
		// Legacy invoices can have a paid_amount summary without a corresponding
		// payment row. There is no remaining payment transaction to reverse; the
		// sale itself is the sole source of that reported amount and will be hard
		// deleted below. Clear the denormalized summary inside this transaction
		// so it cannot be mistaken for an unreversed payment effect.
		updates := []string{"paid_amount=0"}
		if exists, err := s.columnExists(ctx, tx, "sales", "payment_status"); err != nil {
			return nil, fmt.Errorf("inspect legacy sale payment status: %w", err)
		} else if exists {
			updates = append(updates, "payment_status='unpaid'")
		}
		if exists, err := s.columnExists(ctx, tx, "sales", "remaining_amount"); err != nil {
			return nil, fmt.Errorf("inspect legacy sale remaining amount: %w", err)
		} else if exists {
			updates = append(updates, "remaining_amount=0")
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE sales SET `+strings.Join(updates, ",")+` WHERE id=?`), saleID.String()); err != nil {
			return nil, fmt.Errorf("clear legacy sale payment summary before deletion: %w", err)
		}
		remainingPaid = 0
	}
	if remainingPaid > 0.000001 {
		return nil, ErrSalePaymentHistoryInconsistent
	}
	if debtsExist, err := s.tableExists(ctx, tx, "debts"); err != nil {
		return nil, err
	} else if debtsExist {
		if hasPaidAmount, err := s.columnExists(ctx, tx, "debts", "paid_amount"); err != nil {
			return nil, err
		} else if hasPaidAmount {
			var unresolvedAllocations int
			if err := tx.GetContext(ctx, &unresolvedAllocations, tx.Rebind(`SELECT COUNT(*) FROM debts WHERE sale_id=? AND COALESCE(paid_amount,0)>0.000001`), saleID.String()); err != nil {
				return nil, fmt.Errorf("verify sale debt payment reversals: %w", err)
			}
			if unresolvedAllocations > 0 {
				return nil, ErrSaleDebtAllocationHistoryInconsistent
			}
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
		// No movement means no stock delta was committed for an unfinished sale.
		if itemCount > 0 && len(movements) == 0 && strings.EqualFold(strings.TrimSpace(sale.Status), "completed") {
			legacyDeltas, err := restoreLegacySaleInventoryTx(ctx, tx, s, saleID)
			if err != nil {
				return nil, fmt.Errorf("reconstruct legacy sale inventory reversal: %w", err)
			}
			for productID, delta := range legacyDeltas {
				stockDeltas[productID] += delta
			}
		}
		restoredItems := make(map[string]bool)
		for _, movement := range movements {
			if movement.Quantity >= 0 {
				return nil, fmt.Errorf("invalid sale movement quantity %d", movement.Quantity)
			}
			if !movement.ProductID.Valid || strings.TrimSpace(movement.ProductID.String) == "" {
				return nil, fmt.Errorf("sale movement has no product link; cannot reverse its stock delta")
			}
			stockDeltas[movement.ProductID.String] += -movement.Quantity
			if movement.ItemID.Valid {
				if restoredItems[movement.ItemID.String] {
					continue
				}
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
				restoredItems[movement.ItemID.String] = true
			}
		}
		// Inventory movements record the committed stock effect, including legacy invoice mismatches.
		// Reverse those deltas atomically instead of making the transaction undeletable.
	}
	if !movementExists && itemCount > 0 && strings.EqualFold(strings.TrimSpace(sale.Status), "completed") {
		legacyDeltas, err := restoreLegacySaleInventoryTx(ctx, tx, s, saleID)
		if err != nil {
			return nil, fmt.Errorf("reconstruct legacy sale inventory reversal: %w", err)
		}
		for productID, delta := range legacyDeltas {
			stockDeltas[productID] += delta
		}
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
	paymentsHaveSaleID := false
	if paymentsExist {
		paymentsHaveSaleID, err = s.columnExists(ctx, tx, "payments", "sale_id")
		if err != nil {
			return nil, fmt.Errorf("inspect sale payment links: %w", err)
		}
	}
	if paymentsExist && paymentsHaveSaleID {
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
	if exists, err := s.tableExists(ctx, tx, "payment_transactions"); err != nil {
		return nil, err
	} else if exists {
		if err := paymenttransactions.NewService(paymenttransactions.NewRepository(s.db)).DeleteSaleHistoryTx(ctx, tx, saleID); err != nil {
			return nil, fmt.Errorf("delete sale provider transaction history: %w", err)
		}
	}

	for _, dependency := range []struct{ table, query string }{
		{"financial_transactions", `DELETE FROM financial_transactions WHERE sale_id=?`},
		{"profit_entries", `DELETE FROM profit_entries WHERE sale_id=?`},
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

	shiftsExist, err := s.tableExists(ctx, tx, "pos_shifts")
	if err != nil {
		return nil, fmt.Errorf("inspect POS shift summaries: %w", err)
	}
	if shiftsExist {
		for _, column := range []struct{ table, name string }{
			{"sales", "user_id"},
			{"sales", "created_at"},
			{"pos_shifts", "user_id"},
			{"pos_shifts", "opened_at"},
			{"pos_shifts", "closed_at"},
			{"pos_shifts", "sales_total"},
			{"pos_shifts", "sale_count"},
		} {
			exists, err := s.columnExists(ctx, tx, column.table, column.name)
			if err != nil {
				return nil, fmt.Errorf("inspect POS shift column %s.%s: %w", column.table, column.name, err)
			}
			if !exists {
				return nil, fmt.Errorf("cannot safely reconcile POS shifts: missing column %s.%s", column.table, column.name)
			}
		}

		refreshShiftQuery := `UPDATE pos_shifts AS sh SET
			sales_total=(SELECT COALESCE(SUM(s.total_amount),0) FROM sales s WHERE CAST(s.user_id AS TEXT)=CAST(sh.user_id AS TEXT) AND LOWER(COALESCE(s.status,'completed'))='completed' AND s.created_at>=sh.opened_at AND (sh.closed_at IS NULL OR s.created_at<sh.closed_at) AND s.id<>?),
			sale_count=(SELECT COUNT(*) FROM sales s WHERE CAST(s.user_id AS TEXT)=CAST(sh.user_id AS TEXT) AND LOWER(COALESCE(s.status,'completed'))='completed' AND s.created_at>=sh.opened_at AND (sh.closed_at IS NULL OR s.created_at<sh.closed_at) AND s.id<>?)
			WHERE EXISTS (SELECT 1 FROM sales affected WHERE affected.id=? AND CAST(affected.user_id AS TEXT)=CAST(sh.user_id AS TEXT) AND affected.created_at>=sh.opened_at AND (sh.closed_at IS NULL OR affected.created_at<sh.closed_at))`
		if _, err := tx.ExecContext(ctx, tx.Rebind(refreshShiftQuery), saleID.String(), saleID.String(), saleID.String()); err != nil {
			return nil, fmt.Errorf("refresh POS shift after sale deletion: %w", err)
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
	return &DeleteResult{Action: "deleted", Message: "Sale deleted and effects reversed", CanProceed: true}, nil
}

func (s *SmartDeleteService) prepareLinkedPayments(ctx context.Context, saleID, userID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sale payment preflight: %w", err)
	}
	defer tx.Rollback()
	ids, err := linkedSalePayments(ctx, tx, s, saleID)
	if err != nil {
		return err
	}
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("release sale payment preflight: %w", err)
	}
	var providerTransactionsExist bool
	providerQuery := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name='payment_transactions')`
	if dbutil.IsSQLite(s.db) {
		providerQuery = `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='payment_transactions')`
	}
	if err := s.db.GetContext(ctx, &providerTransactionsExist, providerQuery); err != nil {
		return fmt.Errorf("inspect sale provider payments: %w", err)
	}
	if providerTransactionsExist {
		if err := paymenttransactions.NewConfiguredService(s.db).PrepareSaleDeletion(ctx, saleID, userID); err != nil {
			return fmt.Errorf("reconcile sale provider payments: %w", err)
		}
	}
	service := payments.NewService(payments.NewRepository(s.db))
	for _, id := range ids {
		if err := service.PrepareDelete(ctx, id, userID); err != nil {
			return fmt.Errorf("reconcile external payment %s: %w", id, err)
		}
	}
	return nil
}

func restoreLegacySaleInventoryTx(ctx context.Context, tx *sqlx.Tx, s *SmartDeleteService, saleID uuid.UUID) (map[string]int, error) {
	var lines []struct {
		ProductID       sql.NullString `db:"product_id"`
		InventoryItemID sql.NullString `db:"inventory_item_id"`
		Quantity        int            `db:"quantity"`
	}
	query := `SELECT CAST(product_id AS TEXT) AS product_id, CAST(inventory_item_id AS TEXT) AS inventory_item_id, quantity FROM sale_items WHERE sale_id=? ORDER BY id`
	if err := tx.SelectContext(ctx, &lines, tx.Rebind(query), saleID.String()); err != nil {
		return nil, fmt.Errorf("load legacy sale lines: %w", err)
	}
	itemTable, err := s.tableExists(ctx, tx, "inventory_items")
	if err != nil {
		return nil, err
	}
	deltas := make(map[string]int)
	for _, line := range lines {
		if line.Quantity <= 0 {
			return nil, fmt.Errorf("legacy sale contains nonpositive line quantity %d", line.Quantity)
		}
		productID := ""
		if line.ProductID.Valid {
			productID = strings.TrimSpace(line.ProductID.String)
		}
		if line.InventoryItemID.Valid && strings.TrimSpace(line.InventoryItemID.String) != "" {
			itemID := strings.TrimSpace(line.InventoryItemID.String)
			if !itemTable {
				return nil, fmt.Errorf("legacy sale line links inventory item %s but inventory_items is missing", itemID)
			}
			var itemProductID string
			if err := tx.GetContext(ctx, &itemProductID, tx.Rebind(`SELECT CAST(product_id AS TEXT) FROM inventory_items WHERE id=?`), itemID); err != nil {
				return nil, fmt.Errorf("load product for legacy inventory item %s: %w", itemID, err)
			}
			if productID == "" {
				productID = itemProductID
			} else if productID != itemProductID {
				return nil, fmt.Errorf("legacy sale item %s product does not match its sale line", itemID)
			}
			result, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory_items SET status='AVAILABLE', sold_at=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=? AND UPPER(TRIM(COALESCE(status,'')))='SOLD'`), itemID)
			if err != nil {
				return nil, fmt.Errorf("restore legacy sold inventory item %s: %w", itemID, err)
			}
			if affected, _ := result.RowsAffected(); affected != 1 {
				return nil, fmt.Errorf("legacy inventory item %s is not SOLD; refusing an ambiguous stock reversal", itemID)
			}
			acquisitionTable, err := s.tableExists(ctx, tx, "acquisition_items")
			if err != nil {
				return nil, err
			}
			if acquisitionTable {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE acquisition_items SET item_status='available', updated_at=CURRENT_TIMESTAMP WHERE inventory_item_id=? AND LOWER(item_status)='sold'`), itemID); err != nil {
					return nil, fmt.Errorf("restore legacy acquired item state: %w", err)
				}
			}
		}
		if productID == "" {
			return nil, fmt.Errorf("legacy sale line has no product or inventory-item link")
		}
		deltas[productID] += line.Quantity
	}
	return deltas, nil
}

func linkedSalePayments(ctx context.Context, tx *sqlx.Tx, s *SmartDeleteService, saleID uuid.UUID) ([]uuid.UUID, error) {
	exists, err := s.tableExists(ctx, tx, "payments")
	if err != nil || !exists {
		return nil, err
	}
	hasSaleID, err := s.columnExists(ctx, tx, "payments", "sale_id")
	if err != nil || !hasSaleID {
		return nil, err
	}
	hasID, err := s.columnExists(ctx, tx, "payments", "id")
	if err != nil || !hasID {
		return nil, ErrSalePaymentHistoryInconsistent
	}
	query := `SELECT CAST(id AS TEXT) FROM payments WHERE sale_id=?`
	allocationTable, err := s.tableExists(ctx, tx, "payment_debt_allocations")
	if err != nil {
		return nil, err
	}
	debtTable, err := s.tableExists(ctx, tx, "debts")
	if err != nil {
		return nil, err
	}
	if allocationTable && debtTable {
		query = `SELECT CAST(payment.id AS TEXT) FROM payments payment WHERE payment.sale_id=?
			UNION SELECT CAST(payment.id AS TEXT) FROM payments payment
			JOIN payment_debt_allocations allocation ON allocation.payment_id=payment.id
			JOIN debts debt ON debt.id=allocation.debt_id
			WHERE debt.sale_id=?`
		var rawIDs []string
		if err := tx.SelectContext(ctx, &rawIDs, tx.Rebind(query), saleID.String(), saleID.String()); err != nil {
			return nil, fmt.Errorf("list sale and debt payments: %w", err)
		}
		return parseLinkedSalePaymentIDs(rawIDs)
	}
	var rawIDs []string
	if err := tx.SelectContext(ctx, &rawIDs, tx.Rebind(query), saleID.String()); err != nil {
		return nil, fmt.Errorf("list sale payments: %w", err)
	}
	return parseLinkedSalePaymentIDs(rawIDs)
}

func parseLinkedSalePaymentIDs(rawIDs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("parse linked sale payment id %q: %w", raw, ErrSalePaymentHistoryInconsistent)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *Service) DeleteSale(ctx context.Context, saleID, userID uuid.UUID) (*DeleteResult, error) {
	return NewSmartDeleteService(s.db).SmartDelete(ctx, saleID, userID)
}

// CleanSalesHistory attempts to remove each sale through the same transactional
// reversal path used by individual sale deletion. Sales with payments, returns,
// or inconsistent stock history remain untouched and are reported as blocked.
func (s *Service) CleanSalesHistory(ctx context.Context, userID uuid.UUID) (*SalesCleanupSummary, error) {
	const pageSize = 100
	var saleIDs []uuid.UUID
	for page := 1; ; page++ {
		sales, total, err := s.repo.ListSales(ctx, page, pageSize, nil)
		if err != nil {
			return nil, fmt.Errorf("load sales for cleanup: %w", err)
		}
		for _, sale := range sales {
			saleIDs = append(saleIDs, sale.ID)
		}
		if len(sales) == 0 || page*pageSize >= total {
			break
		}
	}

	summary := &SalesCleanupSummary{Total: len(saleIDs)}
	for index, saleID := range saleIDs {
		if err := ctx.Err(); err != nil {
			summary.Failed += len(saleIDs) - index
			summary.StoppedEarly = true
			break
		}

		result, err := s.DeleteSale(ctx, saleID, userID)
		if err != nil {
			summary.Failed++
			continue
		}
		switch result.Action {
		case "deleted":
			summary.Deleted++
		case "blocked", "not_found":
			summary.Blocked++
		default:
			summary.Failed++
		}
	}
	return summary, nil
}
