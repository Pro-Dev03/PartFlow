package payments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
)

const paymentDeleteTolerance = 0.000001

type paymentDeleteRow struct {
	CustomerID sql.NullString `db:"customer_id"`
	SupplierID sql.NullString `db:"supplier_id"`
	SaleID     sql.NullString `db:"sale_id"`
	PurchaseID sql.NullString `db:"purchase_id"`
	Amount     float64        `db:"amount"`
	Status     string         `db:"status"`
}

type paymentAllocation struct {
	DebtID string  `db:"debt_id"`
	Amount float64 `db:"amount"`
}

// smartDelete reverses all recorded application-side effects for a completed
// account payment and physically deletes it in the same transaction. For
// legacy payments without allocation rows, it reconstructs deterministic FIFO
// allocations from the debt balances that are currently recorded.
func (r *Repository) smartDelete(ctx context.Context, paymentID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin payment deletion: %w", err)
	}
	defer tx.Rollback()
	if err := r.deleteInTx(ctx, tx, paymentID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit payment deletion: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("payment_deleted")
	return nil
}

// DeleteInTx reverses a payment using an existing transaction. Callers that
// delete a parent sale/purchase can include the payment reversal and parent
// deletion in the same local database commit. The caller must reconcile any
// external payment-provider settlement before opening its transaction.
func (r *Repository) DeleteInTx(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID) error {
	return r.deleteInTx(ctx, tx, paymentID)
}

// DeleteCustomerCascadeTx removes a customer's payment source row after the
// customer deletion workflow has refunded external settlements and is deleting
// every sale, debt, and ledger row owned by that customer in the same
// transaction. There is no surviving account balance to allocate a payment
// reversal against, so replaying legacy debt allocations here would be both
// unnecessary and unsafe.
func (r *Repository) DeleteCustomerCascadeTx(ctx context.Context, tx *sqlx.Tx, paymentID, customerID uuid.UUID) error {
	var rawCustomerID sql.NullString
	if err := tx.GetContext(ctx, &rawCustomerID, tx.Rebind(`SELECT CAST(customer_id AS TEXT) FROM payments WHERE id=?`), paymentID.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPaymentNotFound
		}
		return fmt.Errorf("verify customer payment ownership: %w", err)
	}
	if !rawCustomerID.Valid || rawCustomerID.String != customerID.String() {
		return ErrPaymentHistoryInconsistent
	}
	if err := ensureProviderPaymentCanBeDeleted(ctx, tx, r.db, paymentID); err != nil {
		return err
	}
	if err := hardDeleteProviderPaymentHistory(ctx, tx, r.db, paymentID); err != nil {
		return err
	}
	return deletePaymentRow(ctx, tx, r.db, paymentID)
}

// DeleteSupplierCascadeTx is the supplier counterpart used only when the
// caller reverses/deletes all purchases, supplier returns, debts, and ledger
// rows in the same transaction.
func (r *Repository) DeleteSupplierCascadeTx(ctx context.Context, tx *sqlx.Tx, paymentID, supplierID uuid.UUID) error {
	var rawSupplierID sql.NullString
	if err := tx.GetContext(ctx, &rawSupplierID, tx.Rebind(`SELECT CAST(supplier_id AS TEXT) FROM payments WHERE id=?`), paymentID.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPaymentNotFound
		}
		return fmt.Errorf("verify supplier payment ownership: %w", err)
	}
	if !rawSupplierID.Valid || rawSupplierID.String != supplierID.String() {
		return ErrPaymentHistoryInconsistent
	}
	if err := ensureProviderPaymentCanBeDeleted(ctx, tx, r.db, paymentID); err != nil {
		return err
	}
	if err := hardDeleteProviderPaymentHistory(ctx, tx, r.db, paymentID); err != nil {
		return err
	}
	return deletePaymentRow(ctx, tx, r.db, paymentID)
}

// ReconstructLegacyCustomerAllocationsTx records the FIFO debt mapping used
// by current payment writes. Legacy overpayments that are not represented in
// the debt paid_amount remain unallocated account credits; unexplained legacy
// debt balances are preserved rather than assigned to a payment that may not
// have caused them.
func (r *Repository) ReconstructLegacyCustomerAllocationsTx(ctx context.Context, tx *sqlx.Tx, customerID uuid.UUID) error {
	if err := ensurePaymentAllocationSchema(ctx, tx, r.db); err != nil {
		return err
	}
	return reconstructLegacyCustomerAllocations(ctx, tx, r.db, customerID.String())
}

func (r *Repository) deleteInTx(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID) error {

	statusExpr := `COALESCE(status, payment_status, 'completed')`
	if dbutil.IsSQLite(r.db) {
		statusExpr = `COALESCE(payment_status, 'completed')`
	}
	lock := ""
	if !dbutil.IsSQLite(r.db) {
		lock = " FOR UPDATE"
	}
	purchaseIDExpr := `NULL AS purchase_id`
	if hasPurchaseID, err := paymentColumnExistsTx(ctx, tx, r.db, "payments", "purchase_id"); err != nil {
		return fmt.Errorf("inspect purchase payment link: %w", err)
	} else if hasPurchaseID {
		purchaseIDExpr = `purchase_id`
	}
	query := fmt.Sprintf(`SELECT customer_id, supplier_id, sale_id, %s, amount, LOWER(%s) AS status FROM payments WHERE id = ?%s`, purchaseIDExpr, statusExpr, lock)
	var payment paymentDeleteRow
	if err := tx.GetContext(ctx, &payment, tx.Rebind(query), paymentID.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPaymentNotFound
		}
		return fmt.Errorf("load payment for deletion: %w", err)
	}

	ownerType, ownerID, err := paymentOwner(payment)
	if err != nil {
		return err
	}
	if payment.Status != "pending" && payment.Status != "completed" && payment.Status != "cancelled" && payment.Status != "failed" {
		return ErrPaymentCannotBeCancelled
	}
	if err := ensureProviderPaymentCanBeDeleted(ctx, tx, r.db, paymentID); err != nil {
		return err
	}
	if payment.Status != "completed" {
		// A non-completed payment has no posted financial effect. If a ledger
		// row nevertheless exists, the record is inconsistent and must not be
		// erased without reversing that effect.
		if ownerType != "" {
			var hasAllocationBatch bool
			var batchTableExists bool
			if dbutil.IsSQLite(r.db) {
				if err := tx.GetContext(ctx, &batchTableExists, `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='payment_allocation_batches')`); err != nil {
					return fmt.Errorf("inspect payment allocation batches: %w", err)
				}
			} else if err := tx.GetContext(ctx, &batchTableExists, `SELECT to_regclass('payment_allocation_batches') IS NOT NULL`); err != nil {
				return fmt.Errorf("inspect payment allocation batches: %w", err)
			}
			if batchTableExists {
				if err := tx.GetContext(ctx, &hasAllocationBatch, tx.Rebind(`SELECT EXISTS (SELECT 1 FROM payment_allocation_batches WHERE payment_id = ?)`), paymentID.String()); err != nil {
					return fmt.Errorf("check non-completed payment allocation effects: %w", err)
				}
			}
			if hasAllocationBatch {
				return ErrPaymentHistoryInconsistent
			}
			ledgerTable, ownerColumn := paymentLedger(ownerType)
			var ledgerExists bool
			check := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE reference_id = ? AND %s = ?)`, ledgerTable, ownerColumn)
			if err := tx.GetContext(ctx, &ledgerExists, tx.Rebind(check), paymentID.String(), ownerID); err != nil {
				return fmt.Errorf("check pending payment ledger effects: %w", err)
			}
			if ledgerExists {
				return ErrPaymentHistoryInconsistent
			}
		}
		if err := hardDeleteProviderPaymentHistory(ctx, tx, r.db, paymentID); err != nil {
			return err
		}
		if err := deletePaymentRow(ctx, tx, r.db, paymentID); err != nil {
			return err
		}
		return nil
	}
	if ownerType == "" {
		// Generic completed payment records do not post to customer/supplier
		// ledgers. They are safe to delete only when nothing references them.
		if err := ensureNoPostedPaymentEffects(ctx, tx, paymentID); err != nil {
			return err
		}
		if payment.SaleID.Valid && strings.TrimSpace(payment.SaleID.String) != "" {
			if err := reverseSalePayment(ctx, tx, r.db, "", payment.SaleID.String, payment.Amount); err != nil {
				return err
			}
		}
		if err := hardDeleteProviderPaymentHistory(ctx, tx, r.db, paymentID); err != nil {
			return err
		}
		if err := deletePaymentRow(ctx, tx, r.db, paymentID); err != nil {
			return err
		}
		return nil
	}
	if payment.Amount <= 0 {
		return ErrPaymentHistoryInconsistent
	}
	if err := ensurePaymentAllocationSchema(ctx, tx, r.db); err != nil {
		return err
	}

	allocations, err := loadPaymentAllocations(ctx, tx, paymentID, ownerType, ownerID)
	if errors.Is(err, ErrPaymentAllocationHistoryMissing) {
		if ownerType == "customer" {
			err = reconstructLegacyCustomerAllocations(ctx, tx, r.db, ownerID)
		} else {
			err = reconstructLegacySupplierMarkers(ctx, tx, r.db, ownerID)
		}
		if err == nil {
			allocations, err = loadPaymentAllocations(ctx, tx, paymentID, ownerType, ownerID)
		}
	}
	if err != nil {
		return err
	}
	allocated := 0.0
	for _, allocation := range allocations {
		allocated += allocation.Amount
	}
	if allocated > payment.Amount+paymentDeleteTolerance {
		return ErrPaymentHistoryInconsistent
	}

	if err := reversePaymentAllocations(ctx, tx, r.db, ownerType, ownerID, allocations); err != nil {
		return err
	}
	if ownerType == "customer" {
		if payment.SaleID.Valid && strings.TrimSpace(payment.SaleID.String) != "" {
			if err := reverseSalePayment(ctx, tx, r.db, ownerID, payment.SaleID.String, payment.Amount); err != nil {
				return err
			}
		}
		if err := adjustAccountBalance(ctx, tx, r.db, "customers", ownerID, payment.Amount); err != nil {
			return err
		}
	} else {
		if payment.PurchaseID.Valid && strings.TrimSpace(payment.PurchaseID.String) != "" {
			if err := reversePurchasePayment(ctx, tx, r.db, payment.PurchaseID.String, payment.Amount); err != nil {
				return err
			}
		}
		if err := adjustAccountBalance(ctx, tx, r.db, "suppliers", ownerID, payment.Amount); err != nil {
			return err
		}
	}

	ledgerTable, ownerColumn := paymentLedger(ownerType)
	deleteLedger := fmt.Sprintf(`DELETE FROM %s WHERE %s = ? AND reference_id = ? AND (LOWER(COALESCE(type, '')) = 'credit' OR UPPER(COALESCE(transaction_type, '')) = 'PAYMENT')`, ledgerTable, ownerColumn)
	if _, err := tx.ExecContext(ctx, tx.Rebind(deleteLedger), ownerID, paymentID.String()); err != nil {
		return fmt.Errorf("remove payment ledger entry: %w", err)
	}
	if err := rebuildAccountLedgerBalances(ctx, tx, ledgerTable, ownerColumn, ownerID); err != nil {
		return err
	}
	if err := hardDeleteProviderPaymentHistory(ctx, tx, r.db, paymentID); err != nil {
		return err
	}
	if err := deletePaymentRow(ctx, tx, r.db, paymentID); err != nil {
		return err
	}
	if err := rebuildDebtSummaryTables(ctx, tx, r.db); err != nil {
		return err
	}
	return nil
}

func paymentOwner(payment paymentDeleteRow) (string, string, error) {
	if payment.CustomerID.Valid && strings.TrimSpace(payment.CustomerID.String) != "" {
		if payment.SupplierID.Valid && strings.TrimSpace(payment.SupplierID.String) != "" {
			return "", "", ErrPaymentHistoryInconsistent
		}
		return "customer", payment.CustomerID.String, nil
	}
	if payment.SupplierID.Valid && strings.TrimSpace(payment.SupplierID.String) != "" {
		return "supplier", payment.SupplierID.String, nil
	}
	return "", "", nil
}

func paymentLedger(ownerType string) (string, string) {
	if ownerType == "customer" {
		return "customer_ledger", "customer_id"
	}
	return "supplier_ledger", "supplier_id"
}

func loadPaymentAllocations(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID, ownerType, ownerID string) ([]paymentAllocation, error) {
	var batch struct {
		OwnerType string `db:"owner_type"`
		OwnerID   string `db:"owner_id"`
	}
	if err := tx.GetContext(ctx, &batch, tx.Rebind(`SELECT owner_type, CAST(owner_id AS TEXT) AS owner_id FROM payment_allocation_batches WHERE payment_id = ?`), paymentID.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentAllocationHistoryMissing
		}
		return nil, fmt.Errorf("load payment allocation marker: %w", err)
	}
	if batch.OwnerType != ownerType || batch.OwnerID != ownerID {
		return nil, ErrPaymentHistoryInconsistent
	}
	var allocations []paymentAllocation
	if err := tx.SelectContext(ctx, &allocations, tx.Rebind(`SELECT CAST(debt_id AS TEXT) AS debt_id, amount FROM payment_debt_allocations WHERE payment_id = ? ORDER BY debt_id`), paymentID.String()); err != nil {
		return nil, fmt.Errorf("load payment debt allocations: %w", err)
	}
	return allocations, nil
}

func ensurePaymentAllocationSchema(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB) error {
	var exists bool
	if dbutil.IsSQLite(db) {
		if err := tx.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='payment_allocation_batches')`); err != nil {
			return fmt.Errorf("check payment allocation schema: %w", err)
		}
	} else if err := tx.GetContext(ctx, &exists, `SELECT to_regclass('payment_allocation_batches') IS NOT NULL`); err != nil {
		return fmt.Errorf("check payment allocation schema: %w", err)
	}
	if !exists {
		return ErrPaymentAllocationTrackingUnavailable
	}
	if dbutil.IsSQLite(db) {
		if err := tx.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='payment_debt_allocations')`); err != nil {
			return fmt.Errorf("check payment debt allocation schema: %w", err)
		}
	} else if err := tx.GetContext(ctx, &exists, `SELECT to_regclass('payment_debt_allocations') IS NOT NULL`); err != nil {
		return fmt.Errorf("check payment debt allocation schema: %w", err)
	}
	if !exists {
		return ErrPaymentAllocationTrackingUnavailable
	}
	return nil
}

type legacyPaymentRow struct {
	ID        string         `db:"id"`
	Amount    float64        `db:"amount"`
	SaleID    sql.NullString `db:"sale_id"`
	CreatedAt string         `db:"created_at"`
}

type legacyDebtRow struct {
	ID        string         `db:"id"`
	Amount    float64        `db:"amount"`
	Paid      float64        `db:"paid_amount"`
	SaleID    sql.NullString `db:"sale_id"`
	CreatedAt string         `db:"created_at"`
}

// reconstructLegacyCustomerAllocations replays legacy customer payments FIFO
// against debts, capped by each debt's currently recorded paid amount. This
// preserves unexplained legacy debt balances and leaves payment overages as
// account credits instead of assigning them to unrelated debts.
func reconstructLegacyCustomerAllocations(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, customerID string) error {
	statusExpr := `COALESCE(status, payment_status, 'completed')`
	if dbutil.IsSQLite(db) {
		statusExpr = `COALESCE(payment_status, 'completed')`
	}
	var debts []legacyDebtRow
	debtQuery := `SELECT id, amount, COALESCE(paid_amount, 0) AS paid_amount, sale_id, created_at FROM debts WHERE customer_id = ? ORDER BY due_date, created_at, id`
	if err := tx.SelectContext(ctx, &debts, tx.Rebind(debtQuery), customerID); err != nil {
		return fmt.Errorf("load legacy customer debts for allocation reconstruction: %w", err)
	}
	var payments []legacyPaymentRow
	paymentQuery := fmt.Sprintf(`SELECT id, amount, sale_id, created_at FROM payments WHERE customer_id = ? AND LOWER(%s) = 'completed' ORDER BY created_at, id`, statusExpr)
	if err := tx.SelectContext(ctx, &payments, tx.Rebind(paymentQuery), customerID); err != nil {
		return fmt.Errorf("load legacy customer payments for allocation reconstruction: %w", err)
	}
	type mappedPayment struct {
		ID        string
		Amount    float64
		SaleID    sql.NullString
		CreatedAt string
		Tracked   bool
		Items     []paymentAllocation
	}
	tracked := map[string]*mappedPayment{}
	var existing []struct {
		PaymentID string   `db:"payment_id"`
		OwnerID   string   `db:"owner_id"`
		DebtID    *string  `db:"debt_id"`
		Amount    *float64 `db:"amount"`
	}
	if err := tx.SelectContext(ctx, &existing, tx.Rebind(`SELECT b.payment_id, CAST(b.owner_id AS TEXT) AS owner_id, CAST(a.debt_id AS TEXT) AS debt_id, a.amount FROM payment_allocation_batches b LEFT JOIN payment_debt_allocations a ON a.payment_id = b.payment_id WHERE b.owner_type = 'customer' AND b.owner_id = ?`), customerID); err != nil {
		return fmt.Errorf("load recorded customer payment allocations: %w", err)
	}
	for _, row := range existing {
		item := tracked[row.PaymentID]
		if item == nil {
			item = &mappedPayment{ID: row.PaymentID, Tracked: true}
			tracked[row.PaymentID] = item
		}
		if row.DebtID != nil && row.Amount != nil {
			item.Items = append(item.Items, paymentAllocation{DebtID: *row.DebtID, Amount: *row.Amount})
		}
	}
	actual := make(map[string]float64, len(debts))
	for _, debt := range debts {
		actual[debt.ID] = debt.Paid
	}
	expected := make(map[string]float64, len(debts))
	newMaps := make([]mappedPayment, 0)
	for _, payment := range payments {
		id := strings.ToLower(payment.ID)
		if item := tracked[id]; item != nil {
			for _, allocation := range item.Items {
				if _, ok := actual[allocation.DebtID]; !ok {
					return ErrPaymentAllocationHistoryMissing
				}
				expected[allocation.DebtID] += allocation.Amount
			}
			continue
		}
		paymentTime, parseErr := dbutil.ParseTimestamp(payment.CreatedAt)
		if parseErr != nil {
			return fmt.Errorf("parse legacy customer payment time: %w", parseErr)
		}
		remaining := payment.Amount
		item := mappedPayment{ID: payment.ID, Amount: payment.Amount, SaleID: payment.SaleID, CreatedAt: payment.CreatedAt}
		for _, debt := range debts {
			if remaining <= paymentDeleteTolerance {
				break
			}
			debtTime, parseErr := dbutil.ParseTimestamp(debt.CreatedAt)
			if parseErr != nil {
				return fmt.Errorf("parse legacy customer debt time: %w", parseErr)
			}
			if debtTime.After(paymentTime) || (payment.SaleID.Valid && (!debt.SaleID.Valid || debt.SaleID.String != payment.SaleID.String)) {
				continue
			}
			open := math.Min(debt.Amount, debt.Paid) - expected[debt.ID]
			if open <= paymentDeleteTolerance {
				continue
			}
			applied := math.Min(open, remaining)
			item.Items = append(item.Items, paymentAllocation{DebtID: debt.ID, Amount: applied})
			expected[debt.ID] += applied
			remaining -= applied
		}
		newMaps = append(newMaps, item)
	}
	for id, applied := range expected {
		if applied > actual[id]+paymentDeleteTolerance {
			return ErrPaymentHistoryInconsistent
		}
	}
	for _, item := range newMaps {
		var saleID any
		if item.SaleID.Valid {
			saleID = item.SaleID.String
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO payment_allocation_batches (payment_id, owner_type, owner_id, sale_id, tracked_at) VALUES (?, 'customer', ?, ?, ?)`), item.ID, customerID, saleID, item.CreatedAt); err != nil {
			return fmt.Errorf("persist reconstructed customer allocation marker: %w", err)
		}
		for _, allocation := range item.Items {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO payment_debt_allocations (id, payment_id, debt_id, amount, created_at) VALUES (?, ?, ?, ?, ?)`), uuid.New().String(), item.ID, allocation.DebtID, allocation.Amount, item.CreatedAt); err != nil {
				return fmt.Errorf("persist reconstructed customer debt allocation: %w", err)
			}
		}
	}
	return nil
}

// reconstructLegacySupplierMarkers replays legacy supplier payments FIFO
// against supplier debts using only each debt's currently recorded paid
// amount. Any payment amount left after those allocations remains an account
// credit; unmatched legacy debt balances remain untouched.
func reconstructLegacySupplierMarkers(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, supplierID string) error {
	statusExpr := `COALESCE(status, payment_status, 'completed')`
	if dbutil.IsSQLite(db) {
		statusExpr = `COALESCE(payment_status, 'completed')`
	}
	var payments []legacyPaymentRow
	query := fmt.Sprintf(`SELECT id, amount, NULL AS sale_id, created_at FROM payments WHERE supplier_id = ? AND LOWER(%s) = 'completed' ORDER BY created_at, id`, statusExpr)
	if err := tx.SelectContext(ctx, &payments, tx.Rebind(query), supplierID); err != nil {
		return fmt.Errorf("load legacy supplier payments: %w", err)
	}
	var debts []legacyDebtRow
	if err := tx.SelectContext(ctx, &debts, tx.Rebind(`SELECT id, amount, COALESCE(paid_amount,0) AS paid_amount, NULL AS sale_id, created_at FROM supplier_debts WHERE supplier_id=? ORDER BY due_date, created_at, id`), supplierID); err != nil {
		return fmt.Errorf("load legacy supplier debts for allocation reconstruction: %w", err)
	}
	var existing []struct {
		PaymentID string   `db:"payment_id"`
		DebtID    *string  `db:"debt_id"`
		Amount    *float64 `db:"amount"`
	}
	if err := tx.SelectContext(ctx, &existing, tx.Rebind(`SELECT b.payment_id, CAST(a.debt_id AS TEXT) AS debt_id, a.amount FROM payment_allocation_batches b LEFT JOIN payment_debt_allocations a ON a.payment_id=b.payment_id WHERE b.owner_type='supplier' AND b.owner_id=?`), supplierID); err != nil {
		return fmt.Errorf("load recorded supplier payment allocations: %w", err)
	}
	type supplierMapping struct {
		Payment legacyPaymentRow
		Items   []paymentAllocation
	}
	tracked := make(map[string]bool)
	expected := make(map[string]float64)
	debtByID := make(map[string]legacyDebtRow, len(debts))
	for _, debt := range debts {
		debtByID[debt.ID] = debt
	}
	for _, row := range existing {
		tracked[strings.ToLower(row.PaymentID)] = true
		if row.DebtID == nil || row.Amount == nil {
			continue
		}
		debt, ok := debtByID[*row.DebtID]
		if !ok {
			return ErrPaymentHistoryInconsistent
		}
		expected[debt.ID] += *row.Amount
		if expected[debt.ID] > math.Min(debt.Amount, debt.Paid)+paymentDeleteTolerance {
			return ErrPaymentHistoryInconsistent
		}
	}
	newMappings := make([]supplierMapping, 0)
	for _, payment := range payments {
		if tracked[strings.ToLower(payment.ID)] {
			continue
		}
		paymentTime, err := dbutil.ParseTimestamp(payment.CreatedAt)
		if err != nil {
			return fmt.Errorf("parse legacy supplier payment time: %w", err)
		}
		remaining := payment.Amount
		mapping := supplierMapping{Payment: payment}
		for _, debt := range debts {
			if remaining <= paymentDeleteTolerance {
				break
			}
			debtTime, err := dbutil.ParseTimestamp(debt.CreatedAt)
			if err != nil {
				return fmt.Errorf("parse legacy supplier debt time: %w", err)
			}
			if debtTime.After(paymentTime) {
				continue
			}
			open := math.Min(debt.Amount, debt.Paid) - expected[debt.ID]
			if open <= paymentDeleteTolerance {
				continue
			}
			applied := math.Min(open, remaining)
			mapping.Items = append(mapping.Items, paymentAllocation{DebtID: debt.ID, Amount: applied})
			expected[debt.ID] += applied
			remaining -= applied
		}
		newMappings = append(newMappings, mapping)
	}
	for _, mapping := range newMappings {
		payment := mapping.Payment
		if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO payment_allocation_batches (payment_id, owner_type, owner_id, sale_id, tracked_at) VALUES (?, 'supplier', ?, NULL, ?)`), payment.ID, supplierID, payment.CreatedAt); err != nil {
			return fmt.Errorf("persist reconstructed supplier payment marker: %w", err)
		}
		for _, allocation := range mapping.Items {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`INSERT INTO payment_debt_allocations (id, payment_id, debt_id, amount, created_at) VALUES (?, ?, ?, ?, ?)`), uuid.New().String(), payment.ID, allocation.DebtID, allocation.Amount, payment.CreatedAt); err != nil {
				return fmt.Errorf("persist reconstructed supplier debt allocation: %w", err)
			}
		}
	}
	return nil
}

func reversePaymentAllocations(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, ownerType, ownerID string, allocations []paymentAllocation) error {
	for _, allocation := range allocations {
		if allocation.Amount <= 0 {
			return ErrPaymentHistoryInconsistent
		}
		if ownerType == "customer" {
			var amount, paid float64
			var dueDate sql.NullString
			var status string
			query := `SELECT amount, COALESCE(paid_amount, 0), due_date, LOWER(COALESCE(status, 'pending')) FROM debts WHERE id = ? AND customer_id = ?`
			if err := tx.QueryRowxContext(ctx, tx.Rebind(query), allocation.DebtID, ownerID).Scan(&amount, &paid, &dueDate, &status); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return ErrPaymentHistoryInconsistent
				}
				return fmt.Errorf("load customer debt for payment reversal: %w", err)
			}
			if paid+paymentDeleteTolerance < allocation.Amount || amount+paymentDeleteTolerance < paid {
				return ErrPaymentHistoryInconsistent
			}
			newPaid := paid - allocation.Amount
			if newPaid < paymentDeleteTolerance {
				newPaid = 0
			}
			remaining := math.Max(0, amount-newPaid)
			newStatus := reopenedCustomerDebtStatus(status, newPaid, remaining)
			if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE debts SET paid_amount = ?, remaining_amount = ?, status = ?, updated_at = `+dbutil.NowSQL(db)+` WHERE id = ? AND customer_id = ?`), newPaid, remaining, newStatus, allocation.DebtID, ownerID); err != nil {
				return fmt.Errorf("reverse customer debt allocation: %w", err)
			}
			if exists, err := paymentTableExistsTx(ctx, tx, db, "customer_debts"); err != nil {
				return err
			} else if exists {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE customer_debts SET paid_amount = ?, is_paid = ? WHERE id = ? AND customer_id = ?`), newPaid, remaining <= paymentDeleteTolerance, allocation.DebtID, ownerID); err != nil {
					return fmt.Errorf("reverse customer debt projection: %w", err)
				}
			}
			_ = dueDate
		} else {
			var amount, paid float64
			var isPaid bool
			if err := tx.QueryRowxContext(ctx, tx.Rebind(`SELECT amount, COALESCE(paid_amount, 0), COALESCE(is_paid, FALSE) FROM supplier_debts WHERE id = ? AND supplier_id = ?`), allocation.DebtID, ownerID).Scan(&amount, &paid, &isPaid); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return ErrPaymentHistoryInconsistent
				}
				return fmt.Errorf("load supplier debt for payment reversal: %w", err)
			}
			if paid+paymentDeleteTolerance < allocation.Amount || amount+paymentDeleteTolerance < paid {
				return ErrPaymentHistoryInconsistent
			}
			newPaid := math.Max(0, paid-allocation.Amount)
			newPaid = math.Min(amount, newPaid)
			if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE supplier_debts SET paid_amount = ?, is_paid = ? WHERE id = ? AND supplier_id = ?`), newPaid, newPaid+paymentDeleteTolerance >= amount, allocation.DebtID, ownerID); err != nil {
				return fmt.Errorf("reverse supplier debt allocation: %w", err)
			}
			_ = isPaid
		}
	}
	return nil
}

func reopenedCustomerDebtStatus(previous string, paid, remaining float64) string {
	if remaining <= paymentDeleteTolerance {
		return "paid"
	}
	if paid > paymentDeleteTolerance {
		if previous == "overdue" {
			return "overdue"
		}
		return "partial"
	}
	if previous == "overdue" {
		return "overdue"
	}
	return "pending"
}

func reverseSalePayment(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, customerID, saleID string, amount float64) error {
	var paid float64
	loadQuery := `SELECT COALESCE(paid_amount, 0) FROM sales WHERE id = ?`
	loadArgs := []any{saleID}
	whereCustomer := ""
	if strings.TrimSpace(customerID) != "" {
		loadQuery += ` AND customer_id = ?`
		loadArgs = append(loadArgs, customerID)
		whereCustomer = ` AND customer_id = ?`
	}
	if err := tx.GetContext(ctx, &paid, tx.Rebind(loadQuery), loadArgs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPaymentHistoryInconsistent
		}
		return fmt.Errorf("load sale for payment reversal: %w", err)
	}
	if paid+paymentDeleteTolerance < amount {
		return ErrPaymentHistoryInconsistent
	}
	paidExpression := `GREATEST(0, COALESCE(paid_amount, 0) - ?)`
	remainingExpression := `LEAST(total_amount, COALESCE(remaining_amount, GREATEST(total_amount - COALESCE(paid_amount, 0), 0)) + ?)`
	if dbutil.IsSQLite(db) {
		paidExpression = `MAX(0, COALESCE(paid_amount, 0) - ?)`
		remainingExpression = `MIN(total_amount, COALESCE(remaining_amount, MAX(total_amount - COALESCE(paid_amount, 0), 0)) + ?)`
	}
	query := fmt.Sprintf(`UPDATE sales SET paid_amount = %[1]s, remaining_amount = %[2]s, payment_status = CASE WHEN %[1]s <= ? THEN 'unpaid' WHEN %[2]s <= ? THEN 'paid' ELSE 'partial' END, updated_at = %[3]s WHERE id = ?%[4]s`, paidExpression, remainingExpression, dbutil.NowSQL(db), whereCustomer)
	args := []any{amount, amount, amount, paymentDeleteTolerance, amount, paymentDeleteTolerance, saleID}
	if strings.TrimSpace(customerID) != "" {
		args = append(args, customerID)
	}
	result, err := tx.ExecContext(ctx, tx.Rebind(query), args...)
	if err != nil {
		return fmt.Errorf("reverse sale payment totals: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrPaymentHistoryInconsistent
	}
	return nil
}

func reversePurchasePayment(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, purchaseID string, amount float64) error {
	var row struct {
		Total float64 `db:"total_amount"`
		Paid  float64 `db:"paid_amount"`
	}
	if err := tx.GetContext(ctx, &row, tx.Rebind(`SELECT COALESCE(total_amount,0) AS total_amount, COALESCE(paid_amount,0) AS paid_amount FROM purchases WHERE id=?`), purchaseID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPaymentHistoryInconsistent
		}
		return fmt.Errorf("reload purchase for payment reversal: %w", err)
	}
	if amount <= 0 || row.Paid+paymentDeleteTolerance < amount || row.Total+paymentDeleteTolerance < row.Paid {
		return ErrPaymentHistoryInconsistent
	}
	newPaid := math.Max(0, row.Paid-amount)
	remaining := math.Max(0, row.Total-newPaid)
	hasRemaining, err := paymentColumnExistsTx(ctx, tx, db, "purchases", "remaining_amount")
	if err != nil {
		return err
	}
	hasUpdatedAt, err := paymentColumnExistsTx(ctx, tx, db, "purchases", "updated_at")
	if err != nil {
		return err
	}
	query := `UPDATE purchases SET paid_amount=? WHERE id=?`
	args := []any{newPaid, purchaseID}
	if hasRemaining {
		query = `UPDATE purchases SET paid_amount=?,remaining_amount=? WHERE id=?`
		args = []any{newPaid, remaining, purchaseID}
	}
	if hasUpdatedAt {
		query = strings.TrimSuffix(query, ` WHERE id=?`) + `,updated_at=` + dbutil.NowSQL(db) + ` WHERE id=?`
	}
	result, err := tx.ExecContext(ctx, tx.Rebind(query), args...)
	if err != nil {
		return fmt.Errorf("reverse purchase paid balance: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrPaymentHistoryInconsistent
	}
	return nil
}

func adjustAccountBalance(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table, ownerID string, amount float64) error {
	query := fmt.Sprintf(`UPDATE %s SET current_balance = COALESCE(current_balance, 0) + ?, updated_at = %s WHERE id = ?`, table, dbutil.NowSQL(db))
	result, err := tx.ExecContext(ctx, tx.Rebind(query), amount, ownerID)
	if err != nil {
		return fmt.Errorf("restore %s balance after payment deletion: %w", strings.TrimSuffix(table, "s"), err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrPaymentHistoryInconsistent
	}
	return nil
}

func rebuildAccountLedgerBalances(ctx context.Context, tx *sqlx.Tx, table, ownerColumn, ownerID string) error {
	query := fmt.Sprintf(`SELECT id, LOWER(COALESCE(type, '')) AS entry_type, amount FROM %s WHERE %s = ? ORDER BY created_at, id`, table, ownerColumn)
	type ledgerRow struct {
		ID    string  `db:"id"`
		Type  string  `db:"entry_type"`
		Value float64 `db:"amount"`
	}
	var rows []ledgerRow
	if err := tx.SelectContext(ctx, &rows, tx.Rebind(query), ownerID); err != nil {
		return fmt.Errorf("read %s for balance rebuild: %w", table, err)
	}
	balance := 0.0
	for _, row := range rows {
		switch row.Type {
		case "debit":
			balance += row.Value
		case "credit":
			balance -= row.Value
		default:
			return ErrPaymentHistoryInconsistent
		}
		update := fmt.Sprintf(`UPDATE %s SET balance = ? WHERE id = ? AND %s = ?`, table, ownerColumn)
		if _, err := tx.ExecContext(ctx, tx.Rebind(update), balance, row.ID, ownerID); err != nil {
			return fmt.Errorf("rebuild running balance in %s: %w", table, err)
		}
	}
	return nil
}

// rebuildDebtSummaryTables recalculates existing daily/monthly debt projections
// from their source tables after the payment and debt effects have been
// reversed. The Postgres expressions use the configured store timezone.
func rebuildDebtSummaryTables(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB) error {
	now := dbutil.NowSQL(db)
	dailyTotal := `COALESCE((SELECT SUM(debt.remaining_amount) FROM debts debt WHERE store_date(debt.created_at) <= summary.date AND debt.remaining_amount > 0), 0)`
	dailyNew := `COALESCE((SELECT SUM(debt.amount) FROM debts debt WHERE store_date(debt.created_at) = summary.date), 0)`
	dailyPayments := `COALESCE((SELECT SUM(payment.amount) FROM payments payment WHERE payment.customer_id IS NOT NULL AND store_date(payment.created_at) = summary.date), 0)`
	dailyOverdue := `COALESCE((SELECT SUM(debt.remaining_amount) FROM debts debt WHERE debt.due_date < summary.date AND debt.remaining_amount > 0), 0)`
	dailyOverdueCount := `COALESCE((SELECT COUNT(*) FROM debts debt WHERE debt.due_date < summary.date AND debt.remaining_amount > 0), 0)`
	dailyPaid := `COALESCE((SELECT SUM(debt.amount) FROM debts debt WHERE debt.remaining_amount <= 0 AND store_date(debt.updated_at) = summary.date), 0)`
	monthlyStart := `printf('%04d-%02d-01', summary.year, summary.month)`
	monthlyEnd := `date(` + monthlyStart + `, '+1 month')`
	monthlyTotal := `COALESCE((SELECT SUM(debt.remaining_amount) FROM debts debt WHERE store_date(debt.created_at) < ` + monthlyEnd + ` AND debt.remaining_amount > 0), 0)`
	monthlyNew := `COALESCE((SELECT SUM(debt.amount) FROM debts debt WHERE store_date(debt.created_at) >= ` + monthlyStart + ` AND store_date(debt.created_at) < ` + monthlyEnd + `), 0)`
	monthlyPayments := `COALESCE((SELECT SUM(payment.amount) FROM payments payment WHERE payment.customer_id IS NOT NULL AND store_date(payment.created_at) >= ` + monthlyStart + ` AND store_date(payment.created_at) < ` + monthlyEnd + `), 0)`
	monthlyOverdue := `COALESCE((SELECT SUM(debt.remaining_amount) FROM debts debt WHERE debt.due_date < ` + monthlyEnd + ` AND debt.remaining_amount > 0), 0)`
	monthlyOverdueCount := `COALESCE((SELECT COUNT(*) FROM debts debt WHERE debt.due_date < ` + monthlyEnd + ` AND debt.remaining_amount > 0), 0)`
	monthlyPaid := `COALESCE((SELECT SUM(debt.amount) FROM debts debt WHERE debt.remaining_amount <= 0 AND store_date(debt.updated_at) >= ` + monthlyStart + ` AND store_date(debt.updated_at) < ` + monthlyEnd + `), 0)`
	if !dbutil.IsSQLite(db) {
		dailyTotal = `COALESCE((SELECT SUM(debt.remaining_amount) FROM debts debt WHERE ` + accounting.PostgresStoreDateExpression("debt.created_at") + ` <= summary.date AND debt.remaining_amount > 0), 0)`
		dailyNew = `COALESCE((SELECT SUM(debt.amount) FROM debts debt WHERE ` + accounting.PostgresStoreDateExpression("debt.created_at") + ` = summary.date), 0)`
		dailyPayments = `COALESCE((SELECT SUM(payment.amount) FROM payments payment WHERE payment.customer_id IS NOT NULL AND ` + accounting.PostgresStoreDateExpression("payment.payment_date") + ` = summary.date), 0)`
		dailyOverdue = `COALESCE((SELECT SUM(debt.remaining_amount) FROM debts debt WHERE debt.due_date < summary.date AND debt.remaining_amount > 0), 0)`
		dailyOverdueCount = `COALESCE((SELECT COUNT(*) FROM debts debt WHERE debt.due_date < summary.date AND debt.remaining_amount > 0), 0)`
		dailyPaid = `COALESCE((SELECT SUM(debt.amount) FROM debts debt WHERE debt.remaining_amount <= 0 AND ` + accounting.PostgresStoreDateExpression("debt.updated_at") + ` = summary.date), 0)`
		monthlyStart = `make_date(summary.year, summary.month, 1)`
		monthlyEnd = `make_date(summary.year, summary.month, 1) + INTERVAL '1 month'`
		monthlyTotal = `COALESCE((SELECT SUM(debt.remaining_amount) FROM debts debt WHERE ` + accounting.PostgresStoreDateExpression("debt.created_at") + ` < ` + monthlyEnd + ` AND debt.remaining_amount > 0), 0)`
		monthlyNew = `COALESCE((SELECT SUM(debt.amount) FROM debts debt WHERE ` + accounting.PostgresStoreDateExpression("debt.created_at") + ` >= ` + monthlyStart + ` AND ` + accounting.PostgresStoreDateExpression("debt.created_at") + ` < ` + monthlyEnd + `), 0)`
		monthlyPayments = `COALESCE((SELECT SUM(payment.amount) FROM payments payment WHERE payment.customer_id IS NOT NULL AND ` + accounting.PostgresStoreDateExpression("payment.payment_date") + ` >= ` + monthlyStart + ` AND ` + accounting.PostgresStoreDateExpression("payment.payment_date") + ` < ` + monthlyEnd + `), 0)`
		monthlyOverdue = `COALESCE((SELECT SUM(debt.remaining_amount) FROM debts debt WHERE debt.due_date < ` + monthlyEnd + ` AND debt.remaining_amount > 0), 0)`
		monthlyOverdueCount = `COALESCE((SELECT COUNT(*) FROM debts debt WHERE debt.due_date < ` + monthlyEnd + ` AND debt.remaining_amount > 0), 0)`
		monthlyPaid = `COALESCE((SELECT SUM(debt.amount) FROM debts debt WHERE debt.remaining_amount <= 0 AND ` + accounting.PostgresStoreDateExpression("debt.updated_at") + ` >= ` + monthlyStart + ` AND ` + accounting.PostgresStoreDateExpression("debt.updated_at") + ` < ` + monthlyEnd + `), 0)`
	}
	dailyQuery := fmt.Sprintf(`UPDATE daily_debt_summary AS summary SET total_debt=%s,new_debt=%s,payments_received=%s,overdue_debt=%s,overdue_count=%s,paid_debt=%s,updated_at=%s`, dailyTotal, dailyNew, dailyPayments, dailyOverdue, dailyOverdueCount, dailyPaid, now)
	if exists, err := paymentTableExistsTx(ctx, tx, db, "daily_debt_summary"); err != nil {
		return err
	} else if exists {
		if _, err := tx.ExecContext(ctx, dailyQuery); err != nil {
			return fmt.Errorf("rebuild daily debt summaries after payment deletion: %w", err)
		}
	}
	monthlyQuery := fmt.Sprintf(`UPDATE monthly_debt_summary AS summary SET total_debt=%s,new_debt=%s,payments_received=%s,overdue_debt=%s,overdue_count=%s,paid_debt=%s,updated_at=%s`, monthlyTotal, monthlyNew, monthlyPayments, monthlyOverdue, monthlyOverdueCount, monthlyPaid, now)
	if exists, err := paymentTableExistsTx(ctx, tx, db, "monthly_debt_summary"); err != nil {
		return err
	} else if exists {
		if _, err := tx.ExecContext(ctx, monthlyQuery); err != nil {
			return fmt.Errorf("rebuild monthly debt summaries after payment deletion: %w", err)
		}
	}
	return nil
}

func paymentTableExistsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table string) (bool, error) {
	var exists bool
	query := `SELECT to_regclass(?) IS NOT NULL`
	if dbutil.IsSQLite(db) {
		query = `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)`
	}
	if err := tx.GetContext(ctx, &exists, tx.Rebind(query), table); err != nil {
		return false, fmt.Errorf("inspect %s summary table: %w", table, err)
	}
	return exists, nil
}

func deletePaymentRow(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, paymentID uuid.UUID) error {
	for _, table := range []struct {
		name  string
		query string
	}{
		{"payment_debt_allocations", `DELETE FROM payment_debt_allocations WHERE payment_id=?`},
		{"payment_allocation_batches", `DELETE FROM payment_allocation_batches WHERE payment_id=?`},
		{"audit_logs", `DELETE FROM audit_logs WHERE entity_id=?`},
	} {
		exists, err := paymentTableExistsTx(ctx, tx, db, table.name)
		if err != nil {
			return err
		}
		if exists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(table.query), paymentID.String()); err != nil {
				return fmt.Errorf("remove %s records for deleted payment: %w", table.name, err)
			}
		}
	}
	result, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM payments WHERE id = ?`), paymentID.String())
	if err != nil {
		return fmt.Errorf("hard delete payment: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrPaymentNotFound
	}
	return nil
}

func ensureNoPostedPaymentEffects(ctx context.Context, tx *sqlx.Tx, paymentID uuid.UUID) error {
	for _, item := range []struct{ table, column string }{{"customer_ledger", "reference_id"}, {"supplier_ledger", "reference_id"}} {
		var exists bool
		query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE %s = ?)`, item.table, item.column)
		if err := tx.GetContext(ctx, &exists, tx.Rebind(query), paymentID.String()); err != nil {
			return fmt.Errorf("check linked %s effect: %w", item.table, err)
		}
		if exists {
			return ErrPaymentHistoryInconsistent
		}
	}
	return nil
}

func ensureProviderPaymentCanBeDeleted(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, paymentID uuid.UUID) error {
	var exists bool
	if dbutil.IsSQLite(db) {
		if err := tx.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='payment_transactions')`); err != nil {
			return fmt.Errorf("check payment transaction schema: %w", err)
		}
	} else if err := tx.GetContext(ctx, &exists, `SELECT to_regclass('payment_transactions') IS NOT NULL`); err != nil {
		return fmt.Errorf("check payment transaction schema: %w", err)
	}
	if !exists {
		return nil
	}
	if hasPaymentID, err := paymentColumnExistsTx(ctx, tx, db, "payment_transactions", "payment_id"); err != nil {
		return fmt.Errorf("inspect provider payment link: %w", err)
	} else if !hasPaymentID {
		return nil
	}
	var status string
	query := `SELECT LOWER(status) FROM payment_transactions WHERE payment_id = ? AND LOWER(status) IN ('paid','partially_refunded','processing','pending','succeeded','authorized') LIMIT 1`
	if err := tx.GetContext(ctx, &status, tx.Rebind(query), paymentID.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check external payment settlement: %w", err)
	}
	return ErrPaymentRequiresProviderRefund
}

func paymentColumnExistsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table, column string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name=? AND column_name=?)`
	if dbutil.IsSQLite(db) {
		query = `SELECT EXISTS (SELECT 1 FROM pragma_table_info('` + table + `') WHERE name=?)`
		return exists, tx.GetContext(ctx, &exists, tx.Rebind(query), column)
	}
	err := tx.GetContext(ctx, &exists, tx.Rebind(query), table, column)
	return exists, err
}

func hardDeleteProviderPaymentHistory(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, paymentID uuid.UUID) error {
	var exists bool
	if dbutil.IsSQLite(db) {
		if err := tx.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='payment_transactions')`); err != nil {
			return fmt.Errorf("inspect provider transaction history: %w", err)
		}
	} else if err := tx.GetContext(ctx, &exists, `SELECT to_regclass('payment_transactions') IS NOT NULL`); err != nil {
		return fmt.Errorf("inspect provider transaction history: %w", err)
	}
	if !exists {
		return nil
	}
	if dbutil.IsSQLite(db) {
		err := tx.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='payment_webhook_events')`)
		if err != nil {
			return fmt.Errorf("inspect provider webhook history: %w", err)
		}
	} else if err := tx.GetContext(ctx, &exists, `SELECT to_regclass('payment_webhook_events') IS NOT NULL`); err != nil {
		return fmt.Errorf("inspect provider webhook history: %w", err)
	}
	if exists {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM payment_webhook_events WHERE payment_transaction_id IN (SELECT id FROM payment_transactions WHERE payment_id = ?)`), paymentID.String()); err != nil {
			return fmt.Errorf("delete provider webhook history: %w", err)
		}
	}
	if dbutil.IsSQLite(db) {
		err := tx.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='payment_refunds')`)
		if err != nil {
			return fmt.Errorf("inspect provider refund history: %w", err)
		}
	} else if err := tx.GetContext(ctx, &exists, `SELECT to_regclass('payment_refunds') IS NOT NULL`); err != nil {
		return fmt.Errorf("inspect provider refund history: %w", err)
	}
	if exists {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM payment_refunds WHERE payment_transaction_id IN (SELECT id FROM payment_transactions WHERE payment_id = ?)`), paymentID.String()); err != nil {
			return fmt.Errorf("delete provider refund history: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM payment_transactions WHERE payment_id = ?`), paymentID.String()); err != nil {
		return fmt.Errorf("delete provider transaction history: %w", err)
	}
	return nil
}
