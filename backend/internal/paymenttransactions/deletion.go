package paymenttransactions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/paymentproviders"
)

// PrepareSaleDeletion settles gateway effects for every provider transaction
// attached to a sale before the local sale deletion transaction begins.
func (s *Service) PrepareSaleDeletion(ctx context.Context, saleID, userID uuid.UUID) error {
	var ids []uuid.UUID
	if err := s.repo.db.SelectContext(ctx, &ids, `SELECT id FROM payment_transactions WHERE sale_id=$1 ORDER BY created_at,id`, saleID); err != nil {
		return fmt.Errorf("list provider transactions for sale: %w", err)
	}
	var createdBy *uuid.UUID
	if userID != uuid.Nil {
		createdBy = &userID
	}
	for _, id := range ids {
		transaction, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("load provider transaction %s: %w", id, err)
		}
		switch strings.ToLower(strings.TrimSpace(transaction.Status)) {
		case paymentproviders.StatusPending, paymentproviders.StatusProcessing:
			if _, err := s.CancelPayment(ctx, id); err != nil {
				return fmt.Errorf("cancel provider transaction %s: %w", id, err)
			}
		case paymentproviders.StatusPaid, paymentproviders.StatusPartiallyRefunded:
			var refunded int64
			if err := s.repo.db.GetContext(ctx, &refunded, `SELECT COALESCE(SUM(amount_minor),0) FROM payment_refunds WHERE payment_transaction_id=$1 AND status IN ('refunded','partially_refunded')`, id); err != nil {
				return fmt.Errorf("read transaction %s refunds: %w", id, err)
			}
			remaining := transaction.AmountMinor - refunded
			if remaining <= 0 {
				continue
			}
			request := paymentproviders.RefundRequest{
				Amount: remaining, Currency: transaction.Currency,
				IdempotencyKey: "sale-delete:" + saleID.String() + ":" + id.String(),
				Reason:         "Sale deletion requested",
			}
			if _, _, err := s.RefundPayment(ctx, id, request, createdBy); err != nil {
				return fmt.Errorf("refund provider transaction %s: %w", id, err)
			}
		case paymentproviders.StatusRefunded, paymentproviders.StatusFailed, paymentproviders.StatusCancelled:
			// No captured amount remains to reverse.
		default:
			return fmt.Errorf("provider transaction %s has unsupported status %q", id, transaction.Status)
		}
	}
	return nil
}

// DeleteSaleHistoryTx removes local provider event and refund rows only after
// PrepareSaleDeletion has reversed all external funds. It is part of the same
// transaction as deleting the source sale.
func (s *Service) DeleteSaleHistoryTx(ctx context.Context, tx *sqlx.Tx, saleID uuid.UUID) error {
	for _, table := range []string{"payment_webhook_events", "payment_refunds"} {
		exists, err := paymentTransactionTableExistsTx(ctx, tx, s.repo.db, table)
		if err != nil {
			return fmt.Errorf("inspect sale provider %s: %w", table, err)
		}
		if !exists {
			continue
		}
		query := fmt.Sprintf(`DELETE FROM %s WHERE payment_transaction_id IN (SELECT id FROM payment_transactions WHERE sale_id=?)`, table)
		if _, err := tx.ExecContext(ctx, tx.Rebind(query), saleID.String()); err != nil {
			return fmt.Errorf("delete sale provider %s: %w", table, err)
		}
	}
	if exists, err := paymentTransactionTableExistsTx(ctx, tx, s.repo.db, "return_payment_refunds"); err != nil {
		return fmt.Errorf("inspect sale return refund links: %w", err)
	} else if exists {
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM return_payment_refunds WHERE payment_transaction_id IN (SELECT id FROM payment_transactions WHERE sale_id=?)`), saleID.String()); err != nil {
			return fmt.Errorf("delete sale return refund links: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM payment_transactions WHERE sale_id=?`), saleID.String()); err != nil {
		return fmt.Errorf("delete sale provider transactions: %w", err)
	}
	return nil
}

func paymentTransactionTableExistsTx(ctx context.Context, tx *sqlx.Tx, db *sqlx.DB, table string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1)`
	if strings.EqualFold(db.DriverName(), "sqlite") {
		query = `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type IN ('table','view') AND name=$1)`
	}
	var exists bool
	err := tx.GetContext(ctx, &exists, query, table)
	return exists, err
}
