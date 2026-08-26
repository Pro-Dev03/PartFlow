package sales

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// SmartDeleteService handles intelligent deletion logic for sales (ARCHITECTURE-PRINCIPLES.md)
// From user perspective: it's just "delete"
// Internally: decides between DELETE, REVERSE, or BLOCK based on state
type SmartDeleteService struct {
	db             *sql.DB
	reversalService *ReversalService
}

// NewSmartDeleteService creates a new smart delete service
func NewSmartDeleteService(db *sql.DB) *SmartDeleteService {
	return &SmartDeleteService{
		db:             db,
		reversalService: NewReversalService(db),
	}
}

// DeleteResult represents the result of a smart delete operation
type DeleteResult struct {
	Action      string `json:"action"`      // "deleted", "reversed", "blocked"
	Message     string `json:"message"`     // User-friendly message
	CanProceed  bool   `json:"can_proceed"`  // Whether the operation was successful
	Details     *DeleteDetails `json:"details,omitempty"` // Additional details if blocked
}

// DeleteDetails contains information when deletion is blocked
type DeleteDetails struct {
	Reason          string `json:"reason"`
	SuggestedAction string `json:"suggested_action"`
}

// SmartDelete intelligently handles deletion based on sale state
// Frontend just calls "delete", backend decides what to do
func (s *SmartDeleteService) SmartDelete(ctx context.Context, saleID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	// Get sale status
	var status string
	err := s.db.QueryRowContext(ctx,
		"SELECT status FROM sales WHERE id = $1",
		saleID,
	).Scan(&status)

	if err == sql.ErrNoRows {
		return &DeleteResult{
			Action:     "blocked",
			Message:    "عملية البيع غير موجودة",
			CanProceed: false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sale status: %w", err)
	}

	// Case 1: Draft sales → Simple DELETE
	if status == "draft" {
		return s.deleteDraft(ctx, saleID, userID)
	}

	// Case 2: Check if can be reversed
	canReverse, err := s.reversalService.CanReverse(ctx, saleID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if sale can be reversed: %w", err)
	}

	// Case 3: Completed sales → Delete with inventory adjustment (internally REVERSE)
	if canReverse {
		return s.deleteWithInventoryAdjustment(ctx, saleID, userID)
	}

	// Case 4: Other statuses → Block
	return &DeleteResult{
		Action:     "blocked",
		Message:    "لا يمكن حذف عملية البيع في حالتها الحالية",
		CanProceed: false,
		Details: &DeleteDetails{
			Reason:          fmt.Sprintf("الحالة الحالية: %s", status),
			SuggestedAction: "يمكنك إلغاء البيع إذا كان في حالة مسودة فقط",
		},
	}, nil
}

// deleteDraft handles simple deletion of draft sales
func (s *SmartDeleteService) deleteDraft(ctx context.Context, saleID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete sale items first
	_, err = tx.ExecContext(ctx,
		"DELETE FROM sale_items WHERE sale_id = $1",
		saleID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to delete sale items: %w", err)
	}

	// Delete sale
	_, err = tx.ExecContext(ctx,
		"DELETE FROM sales WHERE id = $1",
		saleID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to delete sale: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &DeleteResult{
		Action:     "deleted",
		Message:    "تم حذف عملية البيع المسودة بنجاح",
		CanProceed: true,
	}, nil
}

// deleteWithInventoryAdjustment handles deletion of completed sales
// Internally uses REVERSE but presents as simple delete to user
func (s *SmartDeleteService) deleteWithInventoryAdjustment(ctx context.Context, saleID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	// Use reversal service internally
	req := SaleReversalRequest{
		Reason: "User requested deletion",
	}

	_, err := s.reversalService.ReverseSale(ctx, saleID, req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to reverse sale: %w", err)
	}

	return &DeleteResult{
		Action:     "reversed",
		Message:    "تم حذف عملية البيع وإرجاع القطع إلى المخزون",
		CanProceed: true,
	}, nil
}
