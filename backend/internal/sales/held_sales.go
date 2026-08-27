package sales

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type HeldSale struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	Items     json.RawMessage `json:"items" db:"items"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type HoldSaleRequest struct {
	Items []json.RawMessage `json:"items" binding:"required,min=1"`
}

func (s *Service) HoldSale(ctx context.Context, userID uuid.UUID, items []json.RawMessage) (*HeldSale, error) {
	payload, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("failed to encode held sale: %w", err)
	}
	held := &HeldSale{}
	err = s.db.QueryRowxContext(ctx, `
		INSERT INTO held_sales (user_id, items) VALUES ($1, $2)
		RETURNING id, items, created_at
	`, userID, payload).StructScan(held)
	if err != nil {
		return nil, fmt.Errorf("failed to hold sale: %w", err)
	}
	return held, nil
}

func (s *Service) ListHeldSales(ctx context.Context, userID uuid.UUID) ([]HeldSale, error) {
	var held []HeldSale
	if err := s.db.SelectContext(ctx, &held, `
		SELECT id, items, created_at FROM held_sales
		WHERE user_id = $1 ORDER BY created_at DESC
	`, userID); err != nil {
		return nil, fmt.Errorf("failed to list held sales: %w", err)
	}
	return held, nil
}

func (s *Service) DeleteHeldSale(ctx context.Context, userID, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM held_sales WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete held sale: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify held sale deletion: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("held sale not found")
	}
	return nil
}
