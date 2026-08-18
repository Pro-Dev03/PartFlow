package sales

import (
	"time"

	"github.com/google/uuid"
)

// Sale represents a sales transaction
type Sale struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	InvoiceNumber  string     `json:"invoice_number" db:"invoice_number"`
	CustomerID     *uuid.UUID `json:"customer_id,omitempty" db:"customer_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	SaleDate       time.Time  `json:"sale_date" db:"sale_date"`
	Subtotal       float64    `json:"subtotal" db:"subtotal"`
	TaxAmount      float64    `json:"tax_amount" db:"tax_amount"`
	DiscountAmount float64    `json:"discount_amount" db:"discount_amount"`
	TotalAmount    float64    `json:"total_amount" db:"total_amount"`
	CostAmount     float64    `json:"cost_amount" db:"cost_amount"`
	GrossProfit    float64    `json:"gross_profit" db:"gross_profit"`
	NetProfit      float64    `json:"net_profit" db:"net_profit"`
	PaidAmount     float64    `json:"paid_amount" db:"paid_amount"`
	PaymentMethod  *string    `json:"payment_method,omitempty" db:"payment_method"`
	PaymentStatus  string     `json:"payment_status" db:"payment_status"`
	Status         string     `json:"status" db:"status"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// SaleItem represents an item in a sale
type SaleItem struct {
	ID            uuid.UUID `json:"id" db:"id"`
	SaleID        uuid.UUID `json:"sale_id" db:"sale_id"`
	ProductID     uuid.UUID `json:"product_id" db:"product_id"`
	Quantity      int       `json:"quantity" db:"quantity"`
	UnitPrice     float64   `json:"unit_price" db:"unit_price"`
	UnitCost      float64   `json:"unit_cost" db:"unit_cost"`
	DiscountAmount float64  `json:"discount_amount" db:"discount_amount"`
	TaxAmount     float64   `json:"tax_amount" db:"tax_amount"`
	TotalAmount   float64   `json:"total_amount" db:"total_amount"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// TableName returns the table name for the Sale model
func (Sale) TableName() string {
	return "sales"
}

// TableName returns the table name for the SaleItem model
func (SaleItem) TableName() string {
	return "sale_items"
}

// NewSale creates a new Sale instance
func NewSale(organizationID uuid.UUID, invoiceNumber string, userID uuid.UUID) *Sale {
	return &Sale{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		InvoiceNumber:  invoiceNumber,
		UserID:         userID,
		SaleDate:       time.Now(),
		Subtotal:       0,
		TaxAmount:      0,
		DiscountAmount: 0,
		TotalAmount:    0,
		PaidAmount:     0,
		PaymentStatus:  "pending",
		Status:         "completed",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Transaction represents a financial transaction
type Transaction struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	SaleID         *uuid.UUID `json:"sale_id,omitempty" db:"sale_id"`
	Type           string     `json:"type" db:"type"` // "sale", "refund", "payment", "expense"
	Amount         float64    `json:"amount" db:"amount"`
	Currency       string     `json:"currency" db:"currency"`
	Reference      string     `json:"reference" db:"reference"`
	Description    *string    `json:"description,omitempty" db:"description"`
	DebitAccount   string     `json:"debit_account" db:"debit_account"`
	CreditAccount  string     `json:"credit_account" db:"credit_account"`
	Status         string     `json:"status" db:"status"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ProfitEntry represents a profit/loss entry
type ProfitEntry struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	SaleID         *uuid.UUID `json:"sale_id,omitempty" db:"sale_id"`
	Period         string     `json:"period" db:"period"` // "daily", "weekly", "monthly"
	StartDate      time.Time  `json:"start_date" db:"start_date"`
	EndDate        time.Time  `json:"end_date" db:"end_date"`
	Revenue        float64    `json:"revenue" db:"revenue"`
	Cost           float64    `json:"cost" db:"cost"`
	GrossProfit    float64    `json:"gross_profit" db:"gross_profit"`
	NetProfit      float64    `json:"net_profit" db:"net_profit"`
	Margin         float64    `json:"margin" db:"margin"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// TableName returns the table name for the Transaction model
func (Transaction) TableName() string {
	return "financial_transactions"
}

// TableName returns the table name for the ProfitEntry model
func (ProfitEntry) TableName() string {
	return "profit_entries"
}
