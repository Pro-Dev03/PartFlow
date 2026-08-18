package payments

import (
	"time"

	"github.com/google/uuid"
)

// CreatePaymentRequest represents payment creation request
type CreatePaymentRequest struct {
	Type        string    `json:"type" binding:"required,oneof=customer supplier expense"`
	ReferenceID uuid.UUID `json:"reference_id" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Method      string    `json:"method" binding:"required"`
	PaymentDate *time.Time `json:"payment_date,omitempty"`
	Reference   *string   `json:"reference,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
}

// UpdatePaymentRequest represents payment update request
type UpdatePaymentRequest struct {
	Status     string    `json:"status" binding:"omitempty,oneof=pending completed cancelled failed"`
	Method     string    `json:"method" binding:"omitempty"`
	Reference  *string   `json:"reference,omitempty"`
	Notes      *string   `json:"notes,omitempty"`
	PaymentDate *time.Time `json:"payment_date,omitempty"`
}

// PaymentListRequest represents payment list query parameters
type PaymentListRequest struct {
	Page        int        `form:"page" binding:"min=1"`
	PerPage     int        `form:"per_page" binding:"min=1,max=100"`
	Type        string     `form:"type" binding:"omitempty,oneof=customer supplier expense"`
	ReferenceID *uuid.UUID `form:"reference_id"`
	Status      string     `form:"status" binding:"omitempty,oneof=pending completed cancelled failed"`
	Method      string     `form:"method"`
	StartDate   *time.Time `form:"start_date"`
	EndDate     *time.Time `form:"end_date"`
	SortBy      string     `form:"sort_by"`
	SortOrder   string     `form:"sort_order"`
}

// PaymentResponse represents payment response
type PaymentResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Type           string     `json:"type"`
	ReferenceID    uuid.UUID `json:"reference_id"`
	ReferenceName  *string    `json:"reference_name,omitempty"` // customer/supplier name
	Amount         float64    `json:"amount"`
	PaymentDate    time.Time  `json:"payment_date"`
	Method         string     `json:"method"`
	Reference      *string    `json:"reference,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	Status         string     `json:"status"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PaymentSummary represents payment summary statistics
type PaymentSummary struct {
	TotalPayments      float64 `json:"total_payments"`
	CompletedPayments  float64 `json:"completed_payments"`
	PendingPayments    float64 `json:"pending_payments"`
	CancelledPayments  float64 `json:"cancelled_payments"`
	FailedPayments     float64 `json:"failed_payments"`
	TotalCount         int     `json:"total_count"`
}