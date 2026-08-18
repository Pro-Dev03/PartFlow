package suppliers

import (
	"time"

	"github.com/google/uuid"
)

// Supplier represents a supplier
type Supplier struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Code           string     `json:"code" db:"code"`
	Name           string     `json:"name" db:"name"`
	Email          *string    `json:"email,omitempty" db:"email"`
	Phone          *string    `json:"phone,omitempty" db:"phone"`
	Address        *string    `json:"address,omitempty" db:"address"`
	City           *string    `json:"city,omitempty" db:"city"`
	Country        *string    `json:"country,omitempty" db:"country"`
	TaxID          *string    `json:"tax_id,omitempty" db:"tax_id"`
	PaymentTerms   *string    `json:"payment_terms,omitempty" db:"payment_terms"`
	CreditLimit    float64    `json:"credit_limit" db:"credit_limit"`
	CurrentBalance float64    `json:"current_balance" db:"current_balance"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for the Supplier model
func (Supplier) TableName() string {
	return "suppliers"
}

// NewSupplier creates a new Supplier instance
func NewSupplier(organizationID uuid.UUID, code, name string) *Supplier {
	return &Supplier{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Code:           code,
		Name:           name,
		CreditLimit:    0,
		CurrentBalance: 0,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}


// DebtEntry represents a debt entry with detailed information
type DebtEntry struct {
	ID            uuid.UUID `json:"id" db:"id"`
	SupplierID    uuid.UUID `json:"supplier_id" db:"supplier_id"`
	Amount        float64   `json:"amount" db:"amount"`
	ReferenceID   uuid.UUID `json:"reference_id" db:"reference_id"`
	ReferenceType string    `json:"reference_type" db:"reference_type"` // "purchase", "invoice", etc.
	DueDate       time.Time `json:"due_date" db:"due_date"`
	IsPaid        bool      `json:"is_paid" db:"is_paid"`
	PaidAmount    float64   `json:"paid_amount" db:"paid_amount"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// DebtCollection represents a debt collection action
type DebtCollection struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	SupplierID     uuid.UUID  `json:"supplier_id" db:"supplier_id"`
	Type           string     `json:"type" db:"type"` // "reminder", "warning", "legal_action"
	Status         string     `json:"status" db:"status"` // "pending", "sent", "resolved"
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	ScheduledDate  time.Time  `json:"scheduled_date" db:"scheduled_date"`
	CompletedDate  *time.Time `json:"completed_date,omitempty" db:"completed_date"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}