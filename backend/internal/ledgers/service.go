package ledgers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// CreateLedgerEntry creates a new ledger entry with automatic balance calculation
func (s *Service) CreateLedgerEntry(ctx context.Context, req *LedgerEntryRequest, organizationID uuid.UUID, userID uuid.UUID) (*LedgerEntry, error) {
	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get current balance
	var currentBalance int64
	balanceQuery := `
		SELECT COALESCE(balance, 0) 
		FROM ledger_entries 
		WHERE ledger_type = $1 AND entity_id = $2 AND organization_id = $3
		ORDER BY created_at DESC 
		LIMIT 1
	`
	err = tx.GetContext(ctx, &currentBalance, balanceQuery, req.LedgerType, req.EntityID, organizationID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get current balance: %w", err)
	}

	// Calculate new balance
	newBalance := currentBalance + req.Amount

	// Create ledger entry
	entry := &LedgerEntry{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		LedgerType:     req.LedgerType,
		EntityID:       req.EntityID,
		TransactionType: req.TransactionType,
		ReferenceID:    req.ReferenceID,
		Amount:         req.Amount,
		Balance:        newBalance,
		Description:    req.Description,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
	}

	insertQuery := `
		INSERT INTO ledger_entries (id, organization_id, ledger_type, entity_id, 
			transaction_type, reference_id, amount, balance, description, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		entry.ID, entry.OrganizationID, entry.LedgerType, entry.EntityID,
		entry.TransactionType, entry.ReferenceID, entry.Amount, entry.Balance,
		entry.Description, entry.CreatedBy, entry.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	// Update entity balance in relevant table
	if req.LedgerType == LedgerTypeCustomer {
		updateQuery := `UPDATE customers SET current_balance = $1, updated_at = NOW() WHERE id = $2`
		_, err = tx.ExecContext(ctx, updateQuery, newBalance, req.EntityID)
		if err != nil {
			return nil, fmt.Errorf("failed to update customer balance: %w", err)
		}
	} else if req.LedgerType == LedgerTypeSupplier {
		updateQuery := `UPDATE suppliers SET current_balance = $1, updated_at = NOW() WHERE id = $2`
		_, err = tx.ExecContext(ctx, updateQuery, newBalance, req.EntityID)
		if err != nil {
			return nil, fmt.Errorf("failed to update supplier balance: %w", err)
		}
	}

	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, changes, result, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	changes := fmt.Sprintf("Ledger entry: %s, amount: %d, new balance: %d", req.TransactionType, req.Amount, newBalance)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "CREATE_LEDGER_ENTRY", "ledger_entry", entry.ID,
		changes, "success", time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return entry, nil
}

// GetCustomerLedgerSummary retrieves customer ledger summary
func (s *Service) GetCustomerLedgerSummary(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID) (*CustomerLedger, error) {
	summary := &CustomerLedger{
		CustomerID: customerID,
	}

	// Get customer details
	customerQuery := `
		SELECT name, credit_limit 
		FROM customers 
		WHERE id = $1 AND organization_id = $2
	`
	err := s.db.GetContext(ctx, summary, customerQuery, customerID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer details: %w", err)
	}

	// Get current balance
	balanceQuery := `
		SELECT COALESCE(balance, 0) 
		FROM ledger_entries 
		WHERE ledger_type = 'CUSTOMER' AND entity_id = $1 AND organization_id = $2
		ORDER BY created_at DESC 
		LIMIT 1
	`
	err = s.db.GetContext(ctx, &summary.CurrentBalance, balanceQuery, customerID, organizationID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get current balance: %w", err)
	}

	// Calculate totals
	summary.AvailableCredit = summary.CreditLimit - summary.CurrentBalance

	// Get total purchases and payments
	purchasesQuery := `
		SELECT COALESCE(SUM(CASE WHEN transaction_type = 'SALE' THEN amount ELSE 0 END), 0) as total_purchases,
		       COALESCE(SUM(CASE WHEN transaction_type = 'PAYMENT' THEN amount ELSE 0 END), 0) as total_payments
		FROM ledger_entries 
		WHERE ledger_type = 'CUSTOMER' AND entity_id = $1 AND organization_id = $2
	`
	err = s.db.GetContext(ctx, summary, purchasesQuery, customerID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	// Get last transaction date
	lastTxQuery := `
		SELECT created_at 
		FROM ledger_entries 
		WHERE ledger_type = 'CUSTOMER' AND entity_id = $1 AND organization_id = $2
		ORDER BY created_at DESC 
		LIMIT 1
	`
	var lastTx time.Time
	s.db.GetContext(ctx, &lastTx, lastTxQuery, customerID, organizationID)
	summary.LastTransactionAt = lastTx

	// Determine status
	if summary.CurrentBalance > 0 {
		summary.Status = "overdue"
		// Calculate days overdue
		daysOverdue := int(time.Since(lastTx).Hours() / 24)
		summary.DaysOverdue = daysOverdue
	} else {
		summary.Status = "current"
	}

	return summary, nil
}

// GetSupplierLedgerSummary retrieves supplier ledger summary
func (s *Service) GetSupplierLedgerSummary(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID) (*SupplierLedger, error) {
	summary := &SupplierLedger{
		SupplierID: supplierID,
	}

	// Get supplier details
	supplierQuery := `
		SELECT name 
		FROM suppliers 
		WHERE id = $1 AND organization_id = $2
	`
	err := s.db.GetContext(ctx, summary, supplierQuery, supplierID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier details: %w", err)
	}

	// Get current balance
	balanceQuery := `
		SELECT COALESCE(balance, 0) 
		FROM ledger_entries 
		WHERE ledger_type = 'SUPPLIER' AND entity_id = $1 AND organization_id = $2
		ORDER BY created_at DESC 
		LIMIT 1
	`
	err = s.db.GetContext(ctx, &summary.CurrentBalance, balanceQuery, supplierID, organizationID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get current balance: %w", err)
	}

	// Get totals
	totalsQuery := `
		SELECT COALESCE(SUM(CASE WHEN transaction_type = 'PURCHASE' THEN amount ELSE 0 END), 0) as total_purchases,
		       COALESCE(SUM(CASE WHEN transaction_type = 'PURCHASE_PAYMENT' THEN amount ELSE 0 END), 0) as total_payments
		FROM ledger_entries 
		WHERE ledger_type = 'SUPPLIER' AND entity_id = $1 AND organization_id = $2
	`
	err = s.db.GetContext(ctx, summary, totalsQuery, supplierID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	// Get last transaction date
	lastTxQuery := `
		SELECT created_at 
		FROM ledger_entries 
		WHERE ledger_type = 'SUPPLIER' AND entity_id = $1 AND organization_id = $2
		ORDER BY created_at DESC 
		LIMIT 1
	`
	var lastTx time.Time
	s.db.GetContext(ctx, &lastTx, lastTxQuery, supplierID, organizationID)
	summary.LastTransactionAt = lastTx

	return summary, nil
}

// GetLedgerEntries retrieves ledger entries for an entity
func (s *Service) GetLedgerEntries(ctx context.Context, ledgerType LedgerType, entityID uuid.UUID, organizationID uuid.UUID, page, perPage int) ([]*LedgerEntry, int64, error) {
	offset := (page - 1) * perPage

	query := `
		SELECT id, organization_id, ledger_type, entity_id, transaction_type, 
		       reference_id, amount, balance, description, created_by, created_at
		FROM ledger_entries 
		WHERE ledger_type = $1 AND entity_id = $2 AND organization_id = $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	var entries []*LedgerEntry
	err := s.db.SelectContext(ctx, &entries, query, ledgerType, entityID, organizationID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ledger entries: %w", err)
	}

	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*) 
		FROM ledger_entries 
		WHERE ledger_type = $1 AND entity_id = $2 AND organization_id = $3
	`
	err = s.db.GetContext(ctx, &total, countQuery, ledgerType, entityID, organizationID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ledger entries count: %w", err)
	}

	return entries, total, nil
}

// GetOverdueEntities retrieves entities with overdue balances
func (s *Service) GetOverdueEntities(ctx context.Context, ledgerType LedgerType, organizationID uuid.UUID) ([]uuid.UUID, error) {
	var entityIDs []uuid.UUID

	query := `
		SELECT DISTINCT entity_id
		FROM ledger_entries
		WHERE ledger_type = $1 AND organization_id = $2 AND balance > 0
		ORDER BY created_at DESC
	`

	err := s.db.SelectContext(ctx, &entityIDs, query, ledgerType, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue entities: %w", err)
	}

	return entityIDs, nil
}
