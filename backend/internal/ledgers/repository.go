package ledgers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
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
	metadata := "{}"
	if entry.Metadata != nil {
		encoded, marshalErr := json.Marshal(entry.Metadata)
		if marshalErr != nil {
			return fmt.Errorf("failed to encode ledger metadata: %w", marshalErr)
		}
		metadata = string(encoded)
	}
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.LedgerType, entry.EntityID, entry.TransactionType,
		entry.ReferenceID, entry.ReferenceType, entry.Amount, entry.Balance,
		entry.PreviousBalance, entry.Description, metadata, entry.CreatedBy, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}
	return nil
}

// localLedgerRow keeps SQLite's TEXT UUID/timestamp/JSON representation out
// of the public model. PostgreSQL scans these values directly into the model,
// while SQLite needs an explicit conversion step.
type localLedgerRow struct {
	ID              string  `db:"id"`
	LedgerType      string  `db:"ledger_type"`
	EntityID        string  `db:"entity_id"`
	TransactionType string  `db:"transaction_type"`
	ReferenceID     *string `db:"reference_id"`
	ReferenceType   *string `db:"reference_type"`
	Amount          float64 `db:"amount"`
	Balance         float64 `db:"balance"`
	PreviousBalance float64 `db:"previous_balance"`
	Description     string  `db:"description"`
	Metadata        string  `db:"metadata"`
	CreatedBy       string  `db:"created_by"`
	CreatedAt       string  `db:"created_at"`
}

func (r localLedgerRow) model() (*LedgerEntry, error) {
	id, err := uuid.Parse(strings.TrimSpace(r.ID))
	if err != nil {
		return nil, fmt.Errorf("parse ledger entry id: %w", err)
	}
	entityID, err := uuid.Parse(strings.TrimSpace(r.EntityID))
	if err != nil {
		return nil, fmt.Errorf("parse ledger entity id: %w", err)
	}
	createdBy := uuid.Nil
	if strings.TrimSpace(r.CreatedBy) != "" {
		createdBy, err = uuid.Parse(strings.TrimSpace(r.CreatedBy))
		if err != nil {
			return nil, fmt.Errorf("parse ledger creator id: %w", err)
		}
	}
	createdAt, err := dbutil.ParseTimestamp(r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse ledger created_at: %w", err)
	}
	entry := &LedgerEntry{
		ID:              id,
		LedgerType:      LedgerType(r.LedgerType),
		EntityID:        entityID,
		TransactionType: TransactionType(r.TransactionType),
		ReferenceType:   r.ReferenceType,
		Amount:          r.Amount,
		Balance:         r.Balance,
		PreviousBalance: r.PreviousBalance,
		Description:     r.Description,
		CreatedBy:       createdBy,
		CreatedAt:       createdAt,
		Metadata:        map[string]interface{}{},
	}
	if r.ReferenceID != nil && strings.TrimSpace(*r.ReferenceID) != "" {
		value, parseErr := uuid.Parse(strings.TrimSpace(*r.ReferenceID))
		if parseErr != nil {
			return nil, fmt.Errorf("parse ledger reference id: %w", parseErr)
		}
		entry.ReferenceID = &value
	}
	if strings.TrimSpace(r.Metadata) != "" {
		if err := json.Unmarshal([]byte(r.Metadata), &entry.Metadata); err != nil {
			return nil, fmt.Errorf("parse ledger metadata: %w", err)
		}
	}
	return entry, nil
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
	if dbutil.IsSQLite(r.db) {
		if page < 1 {
			page = 1
		}
		if perPage < 1 {
			perPage = 50
		}
		offset := (page - 1) * perPage
		var rows []localLedgerRow
		if err := r.db.SelectContext(ctx, &rows, `SELECT id, ledger_type, entity_id, transaction_type, reference_id, reference_type, amount, balance, previous_balance, COALESCE(description,'') AS description, COALESCE(metadata,'{}') AS metadata, COALESCE(created_by,'') AS created_by, created_at FROM ledger_entries WHERE ledger_type = ? AND entity_id = ? ORDER BY datetime(created_at) DESC LIMIT ? OFFSET ?`, string(ledgerType), entityID.String(), perPage, offset); err != nil {
			return nil, 0, fmt.Errorf("failed to get local ledger entries: %w", err)
		}
		entries := make([]*LedgerEntry, 0, len(rows))
		for _, row := range rows {
			entry, err := row.model()
			if err != nil {
				return nil, 0, err
			}
			entries = append(entries, entry)
		}
		var total int64
		if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM ledger_entries WHERE ledger_type = ? AND entity_id = ?`, string(ledgerType), entityID.String()); err != nil {
			return nil, 0, fmt.Errorf("failed to get local ledger count: %w", err)
		}
		return entries, total, nil
	}
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
		SELECT name AS customer_name, credit_limit, current_balance
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
	var lastTxValue any
	if dbutil.IsSQLite(r.db) {
		lastTxQuery = `SELECT created_at FROM ledger_entries WHERE ledger_type = ? AND entity_id = ? ORDER BY datetime(created_at) DESC LIMIT 1`
		_ = r.db.GetContext(ctx, &lastTxValue, lastTxQuery, string(LedgerTypeCustomer), customerID.String())
	} else {
		var lastTx sql.NullTime
		_ = r.db.GetContext(ctx, &lastTx, lastTxQuery, customerID)
		if lastTx.Valid {
			summary.LastTransactionAt = lastTx.Time
		}
	}
	if dbutil.IsSQLite(r.db) && lastTxValue != nil {
		if parsed, parseErr := dbutil.ParseTimestamp(lastTxValue); parseErr == nil {
			summary.LastTransactionAt = parsed
		}
	}

	// Determine status
	if summary.CurrentBalance > 0 {
		summary.Status = "overdue"
		// A ledger has no due date of its own. Use the age of the latest
		// transaction as a conservative display value; debt endpoints expose
		// the authoritative due-date based value.
		if summary.LastTransactionAt.IsZero() {
			summary.DaysOverdue = 0
		} else {
			daysOverdue := int(time.Since(summary.LastTransactionAt).Hours() / 24)
			if daysOverdue < 0 {
				daysOverdue = 0
			}
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
		SELECT name AS supplier_name, current_balance
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
			COALESCE(SUM(CASE WHEN transaction_type IN ('PURCHASE_PAYMENT', 'PAYMENT') THEN ABS(amount) ELSE 0 END), 0) as total_payments
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
	if dbutil.IsSQLite(r.db) {
		var raw string
		if err := r.db.GetContext(ctx, &raw, `SELECT created_at FROM ledger_entries WHERE ledger_type = ? AND entity_id = ? ORDER BY datetime(created_at) DESC LIMIT 1`, string(LedgerTypeSupplier), supplierID.String()); err == nil {
			if parsed, parseErr := dbutil.ParseTimestamp(raw); parseErr == nil {
				summary.LastTransactionAt = parsed
			}
		}
	} else {
		var lastTx sql.NullTime
		r.db.GetContext(ctx, &lastTx, lastTxQuery, supplierID)
		if lastTx.Valid {
			summary.LastTransactionAt = lastTx.Time
		}
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
		SELECT name AS product_name, sku AS product_sku, barcode AS product_barcode
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
