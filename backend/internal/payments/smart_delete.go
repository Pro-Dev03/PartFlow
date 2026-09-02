package payments

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// SmartDeleteService handles intelligent deletion logic for payments (ARCHITECTURE-PRINCIPLES.md)
// From user perspective: it's just "delete"
// Internally: decides between DELETE, REVERSE, or BLOCK based on state
type SmartDeleteService struct {
	db              *sql.DB
	reversalService *ReversalService
}

// NewSmartDeleteService creates a new smart delete service
func NewSmartDeleteService(db *sql.DB) *SmartDeleteService {
	return &SmartDeleteService{
		db:              db,
		reversalService: NewReversalService(db),
	}
}

// DeleteResult represents the result of a smart delete operation
type DeleteResult struct {
	Action     string         `json:"action"`            // "deleted", "reversed", "blocked"
	Message    string         `json:"message"`           // User-friendly message
	CanProceed bool           `json:"can_proceed"`       // Whether the operation was successful
	Details    *DeleteDetails `json:"details,omitempty"` // Additional details if blocked
}

// DeleteDetails contains information when deletion is blocked
type DeleteDetails struct {
	Reason          string `json:"reason"`
	SuggestedAction string `json:"suggested_action"`
}

// SmartDelete intelligently handles deletion based on payment state
// Frontend just calls "delete", backend decides what to do
func (s *SmartDeleteService) SmartDelete(ctx context.Context, paymentID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	// Get payment status
	var status string
	var isReversed bool
	err := s.db.QueryRowContext(ctx,
		"SELECT status, is_reversed FROM payments WHERE id = $1",
		paymentID,
	).Scan(&status, &isReversed)

	if err == sql.ErrNoRows {
		return &DeleteResult{
			Action:     "blocked",
			Message:    "الدفعة غير موجودة",
			CanProceed: false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment status: %w", err)
	}

	// Case 1: Already reversed → Block
	if isReversed {
		return &DeleteResult{
			Action:     "blocked",
			Message:    "الدفعة ملغاة بالفعل",
			CanProceed: false,
			Details: &DeleteDetails{
				Reason:          "هذه الدفعة تم عكسها مسبقاً",
				SuggestedAction: "يمكنك عرض سجل العكس في التاريخ",
			},
		}, nil
	}

	// Case 2: Pending payments → Simple DELETE
	if status == "pending" {
		return s.deletePending(ctx, paymentID, userID)
	}

	// Case 3: Completed payments → Delete with debt adjustment (internally REVERSE)
	if status == "completed" {
		return s.deleteWithDebtAdjustment(ctx, paymentID, userID)
	}

	// Case 4: Other statuses → Block
	return &DeleteResult{
		Action:     "blocked",
		Message:    "لا يمكن حذف الدفعة في حالتها الحالية",
		CanProceed: false,
		Details: &DeleteDetails{
			Reason:          fmt.Sprintf("الحالة الحالية: %s", status),
			SuggestedAction: "يمكنك حذف الدفعات المعلقة فقط",
		},
	}, nil
}

// deletePending handles simple deletion of pending payments
func (s *SmartDeleteService) deletePending(ctx context.Context, paymentID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM payments WHERE id = $1",
		paymentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to delete payment: %w", err)
	}

	return &DeleteResult{
		Action:     "deleted",
		Message:    "تم حذف الدفعة المعلقة بنجاح",
		CanProceed: true,
	}, nil
}

// deleteWithDebtAdjustment handles deletion of completed payments
// Internally uses REVERSE but presents as simple delete to user
func (s *SmartDeleteService) deleteWithDebtAdjustment(ctx context.Context, paymentID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	// Use reversal service internally
	req := PaymentReversalRequest{
		Reason: "User requested deletion",
	}

	_, err := s.reversalService.ReversePayment(ctx, paymentID, req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to reverse payment: %w", err)
	}

	return &DeleteResult{
		Action:     "reversed",
		Message:    "تم حذف الدفعة وتم تعديل رصيد الدين بشكل تلقائي",
		CanProceed: true,
	}, nil
}
