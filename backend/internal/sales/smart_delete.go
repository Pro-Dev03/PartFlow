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

// SmartDelete performs the explicit delete requested by an authorized user.
func (s *SmartDeleteService) SmartDelete(ctx context.Context, saleID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	var exists bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM sales WHERE id = $1)", saleID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to check sale: %w", err)
	}
	if !exists {
		return &DeleteResult{
			Action:     "not_found",
			Message:    "عملية البيع غير موجودة",
			CanProceed: false,
		}, nil
	}
	return s.deleteDraft(ctx, saleID, userID)
}

// deleteDraft handles simple deletion of draft sales
func (s *SmartDeleteService) deleteDraft(ctx context.Context, saleID uuid.UUID, userID uuid.UUID) (*DeleteResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, query := range []string{
		"DELETE FROM customer_ledger WHERE reference_type = 'sale' AND reference_id = $1",
		"DELETE FROM inventory_movements WHERE reference_type = 'sale' AND reference_id = $1",
		"DELETE FROM item_history WHERE reference_type = 'sale' AND reference_id = $1",
	} {
		if _, err = tx.ExecContext(ctx, query, saleID); err != nil {
			return nil, fmt.Errorf("failed to delete sale dependent history: %w", err)
		}
	}

	// Delete sale items first; sale-linked rows with CASCADE/SET NULL follow
	// the existing database schema.
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
		Message:    "تم حذف عملية البيع بنجاح",
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
