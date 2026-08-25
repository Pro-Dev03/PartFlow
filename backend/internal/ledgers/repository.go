package ledgers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateLedgerEntry creates a new ledger entry
func (r *Repository) CreateLedgerEntry(ctx context.Context, entry *LedgerEntry) error {
	query := `
		INSERT INTO ledger_entries 
		(id, ledger_type, entity_id, transaction_type, reference_id, reference_type, 
		 amount, balance, previous_balance, description, metadata, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.LedgerType, entry.EntityID, entry.TransactionType,
		entry.ReferenceID, entry.ReferenceType, entry.Amount, entry.Balance,
		entry.PreviousBalance, entry.Description, entry.Metadata, entry.CreatedBy, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}
	return nil
}

// GetCurrentBalance gets the current balance for an entity
func (r *Repository) GetCurrentBalance(ctx context.Context, ledgerType LedgerType, entityID uuid.UUID) (float64, error) {
	var balance float64
	query := `
		SELECT COALESCE(balance, 0) 
		FROM ledger_entries 
		WHERE ledger_type = $1 AND entity_id = $2
		ORDER BY created_at DESC 
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &balance, query, ledgerType, entityID)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get current balance: %w", err)
	}
	return balance, nil
}

// GetLedgerEntries retrieves ledger entries for an entity with pagination
func (r *Repository) GetLedgerEntries(ctx context.Context, ledgerType LedgerType, entityID uuid.UUID, page, perPage int) ([]*LedgerEntry, int64, error) {
	offset := (page - 1) * perPage

	query := `
		SELECT id, ledger_type, entity_id, transaction_type, reference_id, reference_type,
		       amount, balance, previous_balance, description, metadata, created_by, created_at
		FROM ledger_entries 
		WHERE ledger_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	var entries []*LedgerEntry
	err := r.db.SelectContext(ctx, &entries, query, ledgerType, entityID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ledger entries: %w", err)
	}

	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*) 
		FROM ledger_entries 
		WHERE ledger_type = $1 AND entity_id = $2
	`
	err = r.db.GetContext(ctx, &total, countQuery, ledgerType, entityID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ledger entries count: %w", err)
	}

	return entries, total, nil
}

// GetCustomerLedgerSummary retrieves customer ledger summary
func (r *Repository) GetCustomerLedgerSummary(ctx context.Context, customerID uuid.UUID) (*CustomerLedger, error) {
	summary := &CustomerLedger{
		CustomerID: customerID,
	}

	// Get customer details
	customerQuery := `
		SELECT name, credit_limit, current_balance 
		FROM customers 
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, summary, customerQuery, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer details: %w", err)
	}

	// Calculate available credit
	summary.AvailableCredit = summary.CreditLimit - summary.CurrentBalance

	// Get total purchases and payments from ledger
	totalsQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN transaction_type = 'SALE' THEN ABS(amount) ELSE 0 END), 0) as total_purchases,
			COALESCE(SUM(CASE WHEN transaction_type = 'PAYMENT' THEN ABS(amount) ELSE 0 END), 0) as total_payments
		FROM ledger_entries 
		WHERE ledger_type = 'CUSTOMER' AND entity_id = $1
	`
	err = r.db.GetContext(ctx, summary, totalsQuery, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	// Get last transaction date
	lastTxQuery := `
		SELECT created_at 
		FROM ledger_entries 
		WHERE ledger_type = 'CUSTOMER' AND entity_id = $1
		ORDER BY created_at DESC 
		LIMIT 1
	`
	var lastTx sql.NullTime
	r.db.GetContext(ctx, &lastTx, lastTxQuery, customerID)
	if lastTx.Valid {
		summary.LastTransactionAt = lastTx.Time
	}

	// Determine status
	if summary.CurrentBalance > 0 {
		summary.Status = "overdue"
		// Calculate days overdue (simplified - in production, use due_date from debts table)
		if !lastTx.Valid {
			summary.DaysOverdue = 0
		} else {
			daysOverdue := int(lastTx.Time.Sub(summary.LastTransactionAt).Hours() / 24)
			summary.DaysOverdue = daysOverdue
		}
	} else {
		summary.Status = "current"
	}

	return summary, nil
}

// GetSupplierLedgerSummary retrieves supplier ledger summary
func (r *Repository) GetSupplierLedgerSummary(ctx context.Context, supplierID uuid.UUID) (*SupplierLedger, error) {
	summary := &SupplierLedger{
		SupplierID: supplierID,
	}

	// Get supplier details
	supplierQuery := `
		SELECT name, current_balance 
		FROM suppliers 
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, summary, supplierQuery, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier details: %w", err)
	}

	// Get totals from ledger
	totalsQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN transaction_type = 'PURCHASE' THEN ABS(amount) ELSE 0 END), 0) as total_purchases,
			COALESCE(SUM(CASE WHEN transaction_type = 'PURCHASE_PAYMENT' THEN ABS(amount) ELSE 0 END), 0) as total_payments
		FROM ledger_entries 
		WHERE ledger_type = 'SUPPLIER' AND entity_id = $1
	`
	err = r.db.GetContext(ctx, summary, totalsQuery, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	// Get last transaction date
	lastTxQuery := `
		SELECT created_at 
		FROM ledger_entries 
		WHERE ledger_type = 'SUPPLIER' AND entity_id = $1
		ORDER BY created_at DESC 
		LIMIT 1
	`
	var lastTx sql.NullTime
	r.db.GetContext(ctx, &lastTx, lastTxQuery, supplierID)
	if lastTx.Valid {
		summary.LastTransactionAt = lastTx.Time
	}

	return summary, nil
}

// GetInventoryLedgerSummary retrieves inventory ledger summary
func (r *Repository) GetInventoryLedgerSummary(ctx context.Context, productID uuid.UUID) (*InventoryLedger, error) {
	summary := &InventoryLedger{
		ProductID: productID,
	}

	// Get product details
	productQuery := `
		SELECT name, sku, barcode 
		FROM products 
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, summary, productQuery, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product details: %w", err)
	}

	// Get current quantity from ledger
	quantityQuery := `
		SELECT COALESCE(balance, 0) as current_quantity
		FROM ledger_entries 
		WHERE ledger_type = 'INVENTORY' AND entity_id = $1
		ORDER BY created_at DESC 
		LIMIT 1
	`
	err = r.db.GetContext(ctx, &summary.CurrentQuantity, quantityQuery, productID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get current quantity: %w", err)
	}

	// Get totals from ledger
	totalsQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN transaction_type = 'STOCK_IN' THEN ABS(amount) ELSE 0 END), 0) as total_in,
			COALESCE(SUM(CASE WHEN transaction_type = 'STOCK_OUT' THEN ABS(amount) ELSE 0 END), 0) as total_out
		FROM ledger_entries 
		WHERE ledger_type = 'INVENTORY' AND entity_id = $1
	`
	err = r.db.GetContext(ctx, summary, totalsQuery, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	return summary, nil
}

// GetOverdueEntities retrieves entities with overdue balances
func (r *Repository) GetOverdueEntities(ctx context.Context, ledgerType LedgerType) ([]uuid.UUID, error) {
	var entityIDs []uuid.UUID

	query := `
		SELECT DISTINCT entity_id
		FROM ledger_entries
		WHERE ledger_type = $1 AND balance > 0
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &entityIDs, query, ledgerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue entities: %w", err)
	}

	return entityIDs, nil
}
