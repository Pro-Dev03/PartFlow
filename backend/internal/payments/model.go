package payments

import (
	"time"

	"github.com/google/uuid"
)

// Payment represents a payment transaction
type Payment struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"`                 // customer, supplier, expense
	ReferenceID uuid.UUID `json:"reference_id" db:"reference_id"` // customer_id, supplier_id, or expense_id
	Amount      float64   `json:"amount" db:"amount"`
	PaymentDate time.Time `json:"payment_date" db:"payment_date"`
	Method      string    `json:"method" db:"method"` // cash, card, bank_transfer, check, etc.
	Reference   *string   `json:"reference,omitempty" db:"reference"`
	Notes       *string   `json:"notes,omitempty" db:"notes"`
	Status      string    `json:"status" db:"status"` // pending, completed, cancelled, failed
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Reversal tracking (ARCHITECTURE-PRINCIPLES.md - Reverse instead of Delete)
	IsReversed        bool       `json:"is_reversed" db:"is_reversed"`                 // هل تم عكس الدفعة؟
	ReversedAt        *time.Time `json:"reversed_at" db:"reversed_at"`                 // متى تم العكس
	ReversedBy        *uuid.UUID `json:"reversed_by" db:"reversed_by"`                 // من قام بالعكس
	ReversalReason    *string    `json:"reversal_reason" db:"reversal_reason"`         // سبب العكس
	ReversalPaymentID *uuid.UUID `json:"reversal_payment_id" db:"reversal_payment_id"` // ID الدفعة العكسية
}

// TableName returns the table name for the Payment model
func (Payment) TableName() string {
	return "payments"
}

// NewPayment creates a new Payment instance
func NewPayment(paymentType string, referenceID uuid.UUID, amount float64, method string, userID uuid.UUID) *Payment {
	return &Payment{
		ID:          uuid.New(),
		Type:        paymentType,
		ReferenceID: referenceID,
		Amount:      amount,
		PaymentDate: time.Now(),
		Method:      method,
		Status:      "pending",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// PaymentReversal represents a reversal of a payment (ARCHITECTURE-PRINCIPLES.md)
type PaymentReversal struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	PaymentID        uuid.UUID  `json:"payment_id" db:"payment_id"`
	Reason           string     `json:"reason" db:"reason"`                         // سبب العكس
	ReversedBy       uuid.UUID  `json:"reversed_by" db:"reversed_by"`               // من قام بالعكس
	ReversedAt       time.Time  `json:"reversed_at" db:"reversed_at"`               // متى تم العكس
	OriginalAmount   float64    `json:"original_amount" db:"original_amount"`       // المبلغ الأصلي
	DebtAdjustmentID *uuid.UUID `json:"debt_adjustment_id" db:"debt_adjustment_id"` // تعديل الدين المرتبط
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// PaymentReversalRequest represents a request to reverse a payment
type PaymentReversalRequest struct {
	Reason string `json:"reason" binding:"required"` // سبب العكس (مثلاً: Wrong payment, Duplicate payment)
}
