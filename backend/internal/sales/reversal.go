package sales

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSaleCannotReverse     = errors.New("sale cannot be reversed - wrong status")
	ErrInvalidReversalReason = errors.New("reversal reason is required")
)

// ReversalService handles sale reversals (ARCHITECTURE-PRINCIPLES.md)
type ReversalService struct {
	db *sql.DB
}

// NewReversalService creates a new reversal service
func NewReversalService(db *sql.DB) *ReversalService {
	return &ReversalService{db: db}
}

// CanReverse checks if a sale can be reversed
func (s *ReversalService) CanReverse(ctx context.Context, saleID uuid.UUID) (bool, error) {
	var status string
	err := s.db.QueryRowContext(ctx,
		"SELECT status FROM sales WHERE id = $1",
		saleID,
	).Scan(&status)

	if err == sql.ErrNoRows {
		return false, ErrSaleNotFound
	}
	if err != nil {
		return false, fmt.Errorf("failed to check sale status: %w", err)
	}

	// Only completed sales can be reversed
	// Draft sales should be deleted, not reversed
	canReverse := status == "completed"

	return canReverse, nil
}

// ReverseSale reverses a sale (ARCHITECTURE-PRINCIPLES.md)
// This creates a reversal record and updates the sale status
func (s *ReversalService) ReverseSale(ctx context.Context, saleID uuid.UUID, req SaleReversalRequest, userID uuid.UUID) (*SaleReversal, error) {
	// Validate request
	if req.Reason == "" {
		return nil, ErrInvalidReversalReason
	}

	// Check if sale can be reversed
	canReverse, err := s.CanReverse(ctx, saleID)
	if err != nil {
		return nil, err
	}
	if !canReverse {
		return nil, ErrSaleCannotReverse
	}

	// Get sale details
	var sale Sale
	err = s.db.QueryRowContext(ctx,
		`SELECT id, invoice_number, total_amount, status, customer_id
		 FROM sales WHERE id = $1`,
		saleID,
	).Scan(&sale.ID, &sale.InvoiceNumber, &sale.TotalAmount, &sale.Status, &sale.CustomerID)

	if err != nil {
		return nil, fmt.Errorf("failed to get sale: %w", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create reversal record
	reversalID := uuid.New()
	now := time.Now()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO sale_reversals (id, sale_id, reason, reversed_by, reversed_at, original_total, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		reversalID, saleID, req.Reason, userID, now, sale.TotalAmount, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reversal record: %w", err)
	}

	// Update sale status and add reversal tracking
	_, err = tx.ExecContext(ctx,
		`UPDATE sales
		 SET status = 'reversed', reversed_at = $1, reversed_by = $2, reversal_reason = $3, updated_at = $4
		 WHERE id = $5`,
		now, userID, req.Reason, now, saleID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update sale status: %w", err)
	}

	// Create reverse inventory movements for all items in this sale
	// This will trigger the automatic current state update via the trigger
	_, err = tx.ExecContext(ctx,
		`INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at, is_reversed, reversed_by, reversed_at, reversal_reason)
		 SELECT
				gen_random_uuid(),
				si.id as item_id,
				si.product_id,
				'REVERSE_SALE' as movement_type,
				si.quantity as quantity,
				ii.current_quantity - si.quantity as before_quantity,
				ii.current_quantity + si.quantity as after_quantity,
				'sale' as reference_type,
				$1 as reference_id,
				$2 as reason,
				$3 as created_by,
				$4 as created_at,
				TRUE as is_reversed,
				$3 as reversed_by,
				$4 as reversed_at,
				$2 as reversal_reason
		 FROM sale_items si
		 LEFT JOIN inventory_items ii ON si.product_id = ii.product_id
		 WHERE si.sale_id = $5`,
		saleID, req.Reason, userID, now, saleID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reverse inventory movements: %w", err)
	}

	// If the sale was on debt, we need to reverse the debt effect
	if sale.PaymentStatus == "debt" && sale.CustomerID != nil {
		// Create a debt adjustment to reverse the sale effect
		adjustmentID := uuid.New()

		_, err = tx.ExecContext(ctx,
			`INSERT INTO ledgers (id, customer_id, type, amount, reference_type, reference_id, notes, created_by, created_at)
			 VALUES ($1, $2, 'DEBT_ADJUSTMENT', $3, 'sale_reversal', $4, $5, $6, $7)`,
			adjustmentID, *sale.CustomerID, -sale.TotalAmount, reversalID, fmt.Sprintf("Sale reversal: %s", req.Reason), userID, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create debt adjustment: %w", err)
		}

		// Update the reversal record with the debt adjustment ID
		_, err = tx.ExecContext(ctx,
			`UPDATE sale_reversals SET inventory_adjustment_ids = array_append(inventory_adjustment_ids, $1) WHERE id = $2`,
			adjustmentID, reversalID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update reversal with debt adjustment: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the reversal record
	reversal := &SaleReversal{
		ID:            reversalID,
		SaleID:        saleID,
		Reason:        req.Reason,
		ReversedBy:    userID,
		ReversedAt:    now,
		OriginalTotal: sale.TotalAmount,
		CreatedAt:     now,
	}

	return reversal, nil
}

// GetReversalHistory returns the reversal history for a sale
func (s *ReversalService) GetReversalHistory(ctx context.Context, saleID uuid.UUID) ([]SaleReversal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, sale_id, reason, reversed_by, reversed_at, original_total, inventory_adjustment_ids, payment_reversal_ids, created_at
		 FROM sale_reversals
		 WHERE sale_id = $1
		 ORDER BY reversed_at DESC`,
		saleID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get reversal history: %w", err)
	}
	defer rows.Close()

	var reversals []SaleReversal
	for rows.Next() {
		var reversal SaleReversal
		err := rows.Scan(
			&reversal.ID,
			&reversal.SaleID,
			&reversal.Reason,
			&reversal.ReversedBy,
			&reversal.ReversedAt,
			&reversal.OriginalTotal,
			&reversal.InventoryAdjustmentIDs,
			&reversal.PaymentReversalIDs,
			&reversal.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reversal: %w", err)
		}
		reversals = append(reversals, reversal)
	}

	return reversals, nil
}
