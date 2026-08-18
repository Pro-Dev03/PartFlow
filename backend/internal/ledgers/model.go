package ledgers

import (
	"time"

	"github.com/google/uuid"
)

// LedgerType represents the type of ledger
type LedgerType string

const (
	LedgerTypeCustomer LedgerType = "CUSTOMER"
	LedgerTypeSupplier LedgerType = "SUPPLIER"
)

// TransactionType represents the type of ledger transaction
type TransactionType string

const (
	TransactionSale         TransactionType = "SALE"
	TransactionPayment      TransactionType = "PAYMENT"
	TransactionReturn       TransactionType = "RETURN"
	TransactionRefund       TransactionType = "REFUND"
	TransactionPurchase     TransactionType = "PURCHASE"
	TransactionPurchasePayment TransactionType = "PURCHASE_PAYMENT"
	TransactionAdjustment   TransactionType = "ADJUSTMENT"
)

// LedgerEntry represents a ledger entry
type LedgerEntry struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	LedgerType     LedgerType      `json:"ledger_type" db:"ledger_type"`
	EntityID       uuid.UUID       `json:"entity_id" db:"entity_id"` // customer_id or supplier_id
	TransactionType TransactionType `json:"transaction_type" db:"transaction_type"`
	ReferenceID    *uuid.UUID      `json:"reference_id" db:"reference_id"` // sale_id, payment_id, etc.
	Amount         int64           `json:"amount" db:"amount"` // positive for debit, negative for credit
	Balance        int64           `json:"balance" db:"balance"` // running balance
	Description    string          `json:"description" db:"description"`
	CreatedBy      uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// CustomerLedger represents customer ledger summary
type CustomerLedger struct {
	CustomerID        uuid.UUID `json:"customer_id"`
	CustomerName      string    `json:"customer_name"`
	CurrentBalance    int64     `json:"current_balance"`
	CreditLimit       int64     `json:"credit_limit"`
	AvailableCredit   int64     `json:"available_credit"`
	TotalPurchases    int64     `json:"total_purchases"`
	TotalPayments     int64     `json:"total_payments"`
	LastTransactionAt time.Time `json:"last_transaction_at"`
	DaysOverdue       int       `json:"days_overdue"`
	Status            string    `json:"status"` // current, overdue, blocked
}

// SupplierLedger represents supplier ledger summary
type SupplierLedger struct {
	SupplierID        uuid.UUID `json:"supplier_id"`
	SupplierName      string    `json:"supplier_name"`
	CurrentBalance    int64     `json:"current_balance"`
	TotalPurchases    int64     `json:"total_purchases"`
	TotalPayments     int64     `json:"total_payments"`
	LastTransactionAt time.Time `json:"last_transaction_at"`
}

// LedgerEntryRequest represents ledger entry creation request
type LedgerEntryRequest struct {
	LedgerType      LedgerType      `json:"ledger_type" binding:"required"`
	EntityID        uuid.UUID       `json:"entity_id" binding:"required"`
	TransactionType TransactionType `json:"transaction_type" binding:"required"`
	ReferenceID     *uuid.UUID      `json:"reference_id"`
	Amount          int64           `json:"amount" binding:"required"`
	Description     string          `json:"description"`
}
