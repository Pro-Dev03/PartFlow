package customers

import (
	"time"

	"github.com/google/uuid"
)

// CustomerRequest represents customer creation/update request
type CustomerRequest struct {
	Code        string   `json:"code" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Email       *string  `json:"email,omitempty"`
	Phone       *string  `json:"phone,omitempty"`
	Address     *string  `json:"address,omitempty"`
	City        *string  `json:"city,omitempty"`
	Country     *string  `json:"country,omitempty"`
	TaxID       *string  `json:"tax_id,omitempty"`
	CreditLimit float64  `json:"credit_limit"`
	Notes       *string  `json:"notes,omitempty"`
	IsActive    bool     `json:"is_active"`
}

// CustomerResponse represents customer response
type CustomerResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	Address        *string    `json:"address,omitempty"`
	City           *string    `json:"city,omitempty"`
	Country        *string    `json:"country,omitempty"`
	TaxID          *string    `json:"tax_id,omitempty"`
	CreditLimit    float64    `json:"credit_limit"`
	CurrentBalance float64    `json:"current_balance"`
	Notes          *string    `json:"notes,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CustomerListRequest represents customer list query parameters
type CustomerListRequest struct {
	Page        int    `form:"page" binding:"min=1"`
	PerPage     int    `form:"per_page" binding:"min=1,max=100"`
	Search      string `form:"search"`
	IsActive    *bool  `form:"is_active"`
	SortBy      string `form:"sort_by"`
	SortOrder   string `form:"sort_order"`
}

// PaymentRequest represents payment request
type PaymentRequest struct {
	Amount      float64  `json:"amount" binding:"required,gt=0"`
	PaymentDate *time.Time `json:"payment_date,omitempty"`
	Method      string   `json:"method" binding:"required"`
	Reference   *string  `json:"reference,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
}

// PaymentResponse represents payment response
type PaymentResponse struct {
	ID            uuid.UUID `json:"id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	Amount        float64   `json:"amount"`
	PaymentDate   time.Time `json:"payment_date"`
	Method        string    `json:"method"`
	Reference     *string   `json:"reference,omitempty"`
	Notes         *string   `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// LedgerEntry represents a ledger entry
type LedgerEntry struct {
	ID          uuid.UUID `json:"id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	Type        string    `json:"type"` // debit, credit
	Amount      float64   `json:"amount"`
	Balance     float64   `json:"balance"`
	Description string    `json:"description"`
	ReferenceID *uuid.UUID `json:"reference_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CustomerLedgerResponse represents customer ledger response
type CustomerLedgerResponse struct {
	CustomerID    uuid.UUID      `json:"customer_id"`
	CustomerName  string         `json:"customer_name"`
	TotalPurchases float64      `json:"total_purchases"`
	TotalPayments  float64      `json:"total_payments"`
	CurrentBalance float64      `json:"current_balance"`
	Entries       []LedgerEntry `json:"entries"`
}
