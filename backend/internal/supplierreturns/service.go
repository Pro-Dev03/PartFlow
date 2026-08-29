package supplierreturns

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct{ db *sqlx.DB }
type Handler struct{ service *Service }

type SupplierReturn struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PurchaseID   uuid.UUID `json:"purchase_id" db:"purchase_id"`
	SupplierID   uuid.UUID `json:"supplier_id" db:"supplier_id"`
	ReturnNumber string    `json:"return_number" db:"return_number"`
	Status       string    `json:"status" db:"status"`
	Reason       string    `json:"reason" db:"reason"`
	RefundAmount float64   `json:"refund_amount" db:"refund_amount"`
	Notes        string    `json:"notes" db:"notes"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type CreateRequest struct {
	PurchaseID uuid.UUID `json:"purchase_id" binding:"required"`
	Reason     string    `json:"reason" binding:"required"`
	Notes      string    `json:"notes"`
}
type AddItemRequest struct {
	PurchaseItemID uuid.UUID `json:"purchase_item_id" binding:"required"`
	Quantity       int       `json:"quantity" binding:"required,min=1"`
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM supplier_returns
		WHERE id = $1 AND status IN ('DRAFT', 'PENDING')
		AND NOT EXISTS (SELECT 1 FROM supplier_return_items WHERE supplier_return_id = $1)`, id)
	if err != nil {
		return fmt.Errorf("delete supplier return: %w", err)
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return fmt.Errorf("only an empty, unprocessed supplier return can be deleted")
	}
	return nil
}

func (s *Service) Reject(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `UPDATE supplier_returns
		SET status = 'REJECTED', updated_at = NOW()
		WHERE id = $1 AND status IN ('SHIPPED', 'RECEIVED')`, id)
	if err != nil {
		return fmt.Errorf("reject supplier return: %w", err)
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return fmt.Errorf("only a supplier return in processing can be rejected")
	}
	return nil
}

func NewService(db *sqlx.DB) *Service      { return &Service{db: db} }
func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (s *Service) List(ctx context.Context, status string) ([]SupplierReturn, error) {
	var rows []SupplierReturn
	query := `SELECT id, purchase_id, supplier_id, return_number, status, reason,
		refund_amount, COALESCE(notes, '') AS notes, created_at FROM supplier_returns`
	args := []interface{}{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT 100"
	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("list supplier returns: %w", err)
	}
	return rows, nil
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateRequest) (*SupplierReturn, error) {
	var r SupplierReturn
	err := s.db.GetContext(ctx, &r, `INSERT INTO supplier_returns
		(purchase_id, supplier_id, return_number, status, reason, notes, created_by)
		SELECT $1, supplier_id, 'SRET-' || upper(substr(replace(uuid_generate_v4()::text, '-', ''), 1, 10)),
		'PENDING', $2, $3, $4 FROM purchases WHERE id = $1
		AND status IN ('received', 'partially_received', 'completed')
		RETURNING id, purchase_id, supplier_id, return_number, status, reason, refund_amount, COALESCE(notes, '') AS notes, created_at`,
		req.PurchaseID, req.Reason, req.Notes, userID)
	if err != nil {
		return nil, fmt.Errorf("create supplier return: %w", err)
	}

	return &r, nil
}

func (s *Service) AddItem(ctx context.Context, id uuid.UUID, req AddItemRequest) error {
	result, err := s.db.ExecContext(ctx, `INSERT INTO supplier_return_items
			(supplier_return_id, purchase_item_id, product_id, quantity, unit_cost)
			SELECT $1, pi.id, pi.product_id, $3::int, pi.unit_price
			FROM purchase_items pi
			JOIN supplier_returns sr ON sr.purchase_id = pi.purchase_id
			JOIN purchases p ON p.id = pi.purchase_id
			WHERE sr.id = $1 AND pi.id = $2
			AND p.status IN ('received', 'partially_received', 'completed')
			AND $3::int <= (
				SELECT COUNT(*) FROM inventory_items
				WHERE product_id = pi.product_id AND status = 'AVAILABLE'
			)`,
		id, req.PurchaseItemID, req.Quantity)
	if err != nil {
		return fmt.Errorf("add supplier return item: %w", err)
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return fmt.Errorf("purchase item is not available for this supplier return")
	}
	return nil
}

func (s *Service) Complete(ctx context.Context, id, userID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin supplier return completion: %w", err)
	}
	defer tx.Rollback()
	var status string
	if err = tx.Get(&status, `SELECT status FROM supplier_returns WHERE id = $1 FOR UPDATE`, id); err != nil {
		return fmt.Errorf("get supplier return: %w", err)
	}
	if status != "PENDING" && status != "SHIPPED" && status != "RECEIVED" {
		return fmt.Errorf("supplier return is not ready to complete")
	}
	var items []struct {
		ProductID uuid.UUID `db:"product_id"`
		Quantity  int       `db:"quantity"`
		UnitCost  float64   `db:"unit_cost"`
	}
	if err = tx.Select(&items, `SELECT product_id, quantity, unit_cost FROM supplier_return_items WHERE supplier_return_id = $1`, id); err != nil {
		return fmt.Errorf("get supplier return items: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("supplier return must contain at least one item")
	}
	var total float64
	for _, item := range items {
		var availableIDs []uuid.UUID
		if err = tx.Select(&availableIDs, `SELECT id FROM inventory_items
			WHERE product_id = $1 AND status = 'AVAILABLE'
			ORDER BY created_at FOR UPDATE`, item.ProductID); err != nil {
			return fmt.Errorf("get available inventory items: %w", err)
		}
		if len(availableIDs) < item.Quantity {
			return fmt.Errorf("insufficient inventory for supplier return")
		}
		for index := 0; index < item.Quantity; index++ {
			var before int
			if err = tx.Get(&before, `SELECT COUNT(*) FROM inventory_items
				WHERE product_id = $1 AND status = 'AVAILABLE'`, item.ProductID); err != nil {
				return fmt.Errorf("count available inventory items: %w", err)
			}
			if _, err = tx.Exec(`UPDATE inventory_items SET status = 'RETURNED', updated_at = NOW() WHERE id = $1`, availableIDs[index]); err != nil {
				return fmt.Errorf("update inventory item: %w", err)
			}
			if _, err = tx.Exec(`INSERT INTO inventory_movements
				(item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by)
				VALUES ($1, $2, 'SUPPLIER_RETURN', -1, $3, $4, 'supplier_return', $5, 'إرجاع بضاعة إلى المورد', NULLIF($6::text, '00000000-0000-0000-0000-000000000000')::uuid)`,
				availableIDs[index], item.ProductID, before, before-1, id, userID); err != nil {
				return fmt.Errorf("record inventory movement: %w", err)
			}
		}
		total += float64(item.Quantity) * item.UnitCost
	}
	if _, err = tx.Exec(`UPDATE supplier_returns SET status = 'COMPLETED', refund_amount = $1, updated_at = NOW() WHERE id = $2`, total, id); err != nil {
		return fmt.Errorf("complete supplier return: %w", err)
	}
	var supplierID uuid.UUID
	if err = tx.Get(&supplierID, `SELECT supplier_id FROM supplier_returns WHERE id = $1`, id); err != nil {
		return fmt.Errorf("get supplier: %w", err)
	}
	if _, err = tx.Exec(`INSERT INTO supplier_ledger (supplier_id, type, amount, balance, description, reference_id)
		SELECT $1, 'credit', $2,
			COALESCE((SELECT balance FROM supplier_ledger WHERE supplier_id = $1 ORDER BY created_at DESC LIMIT 1), 0) - $2,
			'مرتجع مورد ' || $3, $4`, supplierID, total, id.String(), id); err != nil {
		return fmt.Errorf("record supplier ledger entry: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit supplier return: %w", err)
	}
	return nil
}
