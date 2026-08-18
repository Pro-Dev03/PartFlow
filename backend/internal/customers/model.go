package customers

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a customer
type Customer struct {
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
	CreditLimit    float64    `json:"credit_limit" db:"credit_limit"`
	CurrentBalance float64    `json:"current_balance" db:"current_balance"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for the Customer model
func (Customer) TableName() string {
	return "customers"
}

// NewCustomer creates a new Customer instance
func NewCustomer(organizationID uuid.UUID, code, name string) *Customer {
	return &Customer{
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
