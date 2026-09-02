package payments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPaymentCannotReverse  = errors.New("payment cannot be reversed - wrong status")
	ErrInvalidReversalReason = errors.New("reversal reason is required")
)

// ReversalService handles payment reversals (ARCHITECTURE-PRINCIPLES.md)
type ReversalService struct {
	db *sql.DB
}

// NewReversalService creates a new reversal service
func NewReversalService(db *sql.DB) *ReversalService {
	return &ReversalService{db: db}
}

// CanReverse checks if a payment can be reversed
func (s *ReversalService) CanReverse(ctx context.Context, paymentID uuid.UUID) (bool, error) {
	var status string
	var isReversed bool

	err := s.db.QueryRowContext(ctx,
		"SELECT status, is_reversed FROM payments WHERE id = $1",
		paymentID,
	).Scan(&status, &isReversed)

	if err == sql.ErrNoRows {
		return false, ErrPaymentNotFound
	}
	if err != nil {
		return false, fmt.Errorf("failed to check payment status: %w", err)
	}

	// Only completed payments that are not already reversed can be reversed
	canReverse := status == "completed" && !isReversed

	return canReverse, nil
}

// ReversePayment reverses a payment (ARCHITECTURE-PRINCIPLES.md)
// This creates a reversal record and marks the payment as reversed
func (s *ReversalService) ReversePayment(ctx context.Context, paymentID uuid.UUID, req PaymentReversalRequest, userID uuid.UUID) (*PaymentReversal, error) {
	// Validate request
	if req.Reason == "" {
		return nil, ErrInvalidReversalReason
	}

	// Check if payment can be reversed
	canReverse, err := s.CanReverse(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if !canReverse {
		return nil, ErrPaymentCannotReverse
	}

	// Get payment details
	var payment Payment
	err = s.db.QueryRowContext(ctx,
		`SELECT id, type, reference_id, amount, payment_date, method, status
		 FROM payments WHERE id = $1`,
		paymentID,
	).Scan(&payment.ID, &payment.Type, &payment.ReferenceID, &payment.Amount, &payment.PaymentDate, &payment.Method, &payment.Status)

	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// Create reversal record
	reversalID := uuid.New()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO payment_reversals (id, payment_id, reason, reversed_by, reversed_at, original_amount, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		reversalID, paymentID, req.Reason, userID, now, payment.Amount, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reversal record: %w", err)
	}

	// Mark the original payment as reversed
	_, err = tx.ExecContext(ctx,
		`UPDATE payments
		 SET is_reversed = TRUE, reversed_at = $1, reversed_by = $2, reversal_reason = $3, reversal_payment_id = $4, updated_at = $5
		 WHERE id = $6`,
		now, userID, req.Reason, reversalID, now, paymentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to mark payment as reversed: %w", err)
	}

	// If this is a customer payment, we need to adjust the debt
	if payment.Type == "customer" {
		// Create a debt adjustment to reverse the payment effect
		adjustmentID := uuid.New()

		_, err = tx.ExecContext(ctx,
			`INSERT INTO ledgers (id, customer_id, type, amount, reference_type, reference_id, notes, created_by, created_at)
			 VALUES ($1, $2, 'DEBT_ADJUSTMENT', $3, 'payment_reversal', $4, $5, $6, $7)`,
			adjustmentID, payment.ReferenceID, payment.Amount, reversalID, fmt.Sprintf("Payment reversal: %s", req.Reason), userID, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create debt adjustment: %w", err)
		}

		// Update the reversal record with the debt adjustment ID
		_, err = tx.ExecContext(ctx,
			`UPDATE payment_reversals SET debt_adjustment_id = $1 WHERE id = $2`,
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
	reversal := &PaymentReversal{
		ID:             reversalID,
		PaymentID:      paymentID,
		Reason:         req.Reason,
		ReversedBy:     userID,
		ReversedAt:     now,
		OriginalAmount: payment.Amount,
		CreatedAt:      now,
	}

	return reversal, nil
}

// GetReversalHistory returns the reversal history for a payment
func (s *ReversalService) GetReversalHistory(ctx context.Context, paymentID uuid.UUID) ([]PaymentReversal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, payment_id, reason, reversed_by, reversed_at, original_amount, debt_adjustment_id, created_at
		 FROM payment_reversals
		 WHERE payment_id = $1
		 ORDER BY reversed_at DESC`,
		paymentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get reversal history: %w", err)
	}
	defer rows.Close()

	var reversals []PaymentReversal
	for rows.Next() {
		var reversal PaymentReversal
		var debtAdjustmentID *uuid.UUID
		err := rows.Scan(
			&reversal.ID,
			&reversal.PaymentID,
			&reversal.Reason,
			&reversal.ReversedBy,
			&reversal.ReversedAt,
			&reversal.OriginalAmount,
			&debtAdjustmentID,
			&reversal.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reversal: %w", err)
		}
		reversal.DebtAdjustmentID = debtAdjustmentID
		reversals = append(reversals, reversal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate payment reversals: %w", err)
	}

	return reversals, nil
}
