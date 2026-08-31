package sales

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	dbutil "github.com/partflow/smart-store/internal/database"
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
	if dbutil.IsSQLite(s.db) {
		held.ID = uuid.New()
		held.Items = payload
		held.CreatedAt = time.Now().UTC()
		_, err = s.db.ExecContext(ctx, `INSERT INTO held_sales (id,user_id,items,created_at) VALUES (?,?,?,?)`, held.ID.String(), userID.String(), payload, held.CreatedAt.Format(time.RFC3339Nano))
		if err != nil {
			return nil, fmt.Errorf("failed to hold sale: %w", err)
		}
		return held, nil
	}
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
	if dbutil.IsSQLite(s.db) {
		var rows []struct {
			ID        string `db:"id"`
			Items     []byte `db:"items"`
			CreatedAt string `db:"created_at"`
		}
		if err := s.db.SelectContext(ctx, &rows, `SELECT id,items,created_at FROM held_sales WHERE user_id = ? ORDER BY created_at DESC`, userID.String()); err != nil {
			return nil, fmt.Errorf("failed to list held sales: %w", err)
		}
		held := make([]HeldSale, 0, len(rows))
		for _, row := range rows {
			id, _ := uuid.Parse(row.ID)
			t, _ := dbutil.ParseTimestamp(row.CreatedAt)
			held = append(held, HeldSale{ID: id, Items: json.RawMessage(row.Items), CreatedAt: t})
		}
		return held, nil
	}
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
	query := `DELETE FROM held_sales WHERE id = $1 AND user_id = $2`
	args := []interface{}{id, userID}
	if dbutil.IsSQLite(s.db) {
		query = `DELETE FROM held_sales WHERE id = ? AND user_id = ?`
		args = []interface{}{id.String(), userID.String()}
	}
	result, err := s.db.ExecContext(ctx, query, args...)
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
