package payments

import (
	"time"

	"github.com/google/uuid"
)

// Payment represents a payment transaction
type Payment struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Type           string     `json:"type" db:"type"` // customer, supplier, expense
	ReferenceID    uuid.UUID  `json:"reference_id" db:"reference_id"` // customer_id, supplier_id, or expense_id
	Amount         float64    `json:"amount" db:"amount"`
	PaymentDate    time.Time  `json:"payment_date" db:"payment_date"`
	Method         string     `json:"method" db:"method"` // cash, card, bank_transfer, check, etc.
	Reference      *string    `json:"reference,omitempty" db:"reference"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	Status         string     `json:"status" db:"status"` // pending, completed, cancelled, failed
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for the Payment model
func (Payment) TableName() string {
	return "payments"
}

// NewPayment creates a new Payment instance
func NewPayment(organizationID uuid.UUID, paymentType string, referenceID uuid.UUID, amount float64, method string, userID uuid.UUID) *Payment {
	return &Payment{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Type:           paymentType,
		ReferenceID:    referenceID,
		Amount:         amount,
		PaymentDate:    time.Now(),
		Method:         method,
		Status:         "pending",
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}