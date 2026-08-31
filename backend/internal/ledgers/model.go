package ledgers

import (
	"time"

	"github.com/google/uuid"
)

// LedgerType represents the type of ledger
type LedgerType string

const (
	LedgerTypeCustomer  LedgerType = "CUSTOMER"
	LedgerTypeSupplier  LedgerType = "SUPPLIER"
	LedgerTypeInventory LedgerType = "INVENTORY"
)

// TransactionType represents the type of ledger transaction
type TransactionType string

const (
	// Customer transactions
	TransactionSale       TransactionType = "SALE"
	TransactionPayment    TransactionType = "PAYMENT"
	TransactionReturn     TransactionType = "RETURN"
	TransactionRefund     TransactionType = "REFUND"
	TransactionAdjustment TransactionType = "ADJUSTMENT"

	// Supplier transactions
	TransactionPurchase        TransactionType = "PURCHASE"
	TransactionPurchasePayment TransactionType = "PURCHASE_PAYMENT"

	// Inventory transactions
	TransactionStockIn         TransactionType = "STOCK_IN"
	TransactionStockOut        TransactionType = "STOCK_OUT"
	TransactionStockAdjustment TransactionType = "STOCK_ADJUSTMENT"
	TransactionTransfer        TransactionType = "TRANSFER"
	TransactionDamaged         TransactionType = "DAMAGED"
	TransactionRepair          TransactionType = "REPAIR"
)

// LedgerEntry represents a ledger entry
type LedgerEntry struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	LedgerType      LedgerType             `json:"ledger_type" db:"ledger_type"`
	EntityID        uuid.UUID              `json:"entity_id" db:"entity_id"` // customer_id, supplier_id, or product_id
	TransactionType TransactionType        `json:"transaction_type" db:"transaction_type"`
	ReferenceID     *uuid.UUID             `json:"reference_id" db:"reference_id"`         // sale_id, payment_id, etc.
	ReferenceType   *string                `json:"reference_type" db:"reference_type"`     // 'sale', 'payment', 'purchase', etc.
	Amount          float64                `json:"amount" db:"amount"`                     // positive for debit, negative for credit
	Balance         float64                `json:"balance" db:"balance"`                   // running balance
	PreviousBalance float64                `json:"previous_balance" db:"previous_balance"` // balance before this transaction
	Description     string                 `json:"description" db:"description"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedBy       uuid.UUID              `json:"created_by" db:"created_by"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// CustomerLedger represents customer ledger summary
type CustomerLedger struct {
	CustomerID        uuid.UUID `json:"customer_id" db:"customer_id"`
	CustomerName      string    `json:"customer_name" db:"customer_name"`
	CurrentBalance    float64   `json:"current_balance" db:"current_balance"`
	CreditLimit       float64   `json:"credit_limit" db:"credit_limit"`
	AvailableCredit   float64   `json:"available_credit" db:"available_credit"`
	TotalPurchases    float64   `json:"total_purchases" db:"total_purchases"`
	TotalPayments     float64   `json:"total_payments" db:"total_payments"`
	LastTransactionAt time.Time `json:"last_transaction_at" db:"last_transaction_at"`
	DaysOverdue       int       `json:"days_overdue" db:"days_overdue"`
	Status            string    `json:"status" db:"status"` // current, overdue, blocked
}

// SupplierLedger represents supplier ledger summary
type SupplierLedger struct {
	SupplierID        uuid.UUID `json:"supplier_id" db:"supplier_id"`
	SupplierName      string    `json:"supplier_name" db:"supplier_name"`
	CurrentBalance    float64   `json:"current_balance" db:"current_balance"`
	TotalPurchases    float64   `json:"total_purchases" db:"total_purchases"`
	TotalPayments     float64   `json:"total_payments" db:"total_payments"`
	LastTransactionAt time.Time `json:"last_transaction_at" db:"last_transaction_at"`
}

// InventoryLedger represents inventory ledger summary
type InventoryLedger struct {
	ProductID       uuid.UUID `json:"product_id" db:"product_id"`
	ProductName     string    `json:"product_name" db:"product_name"`
	ProductSKU      string    `json:"product_sku" db:"product_sku"`
	ProductBarcode  string    `json:"product_barcode" db:"product_barcode"`
	CurrentQuantity float64   `json:"current_quantity" db:"current_quantity"`
	TotalIn         float64   `json:"total_in" db:"total_in"`
	TotalOut        float64   `json:"total_out" db:"total_out"`
}

// LedgerEntryRequest represents ledger entry creation request
type LedgerEntryRequest struct {
	LedgerType      LedgerType             `json:"ledger_type" binding:"required"`
	EntityID        uuid.UUID              `json:"entity_id" binding:"required"`
	TransactionType TransactionType        `json:"transaction_type" binding:"required"`
	ReferenceID     *uuid.UUID             `json:"reference_id"`
	ReferenceType   *string                `json:"reference_type"`
	Amount          float64                `json:"amount" binding:"required"`
	Description     string                 `json:"description"`
	Metadata        map[string]interface{} `json:"metadata"`
}
