package supplierreturns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
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
	result, err := s.db.ExecContext(ctx, fmt.Sprintf(`UPDATE supplier_returns
		SET status = 'REJECTED', updated_at = %s
		WHERE id = $1 AND status IN ('SHIPPED', 'RECEIVED')`, dbutil.NowSQL(s.db)), id)
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
	var rows []struct {
		ID           string  `db:"id"`
		PurchaseID   string  `db:"purchase_id"`
		SupplierID   string  `db:"supplier_id"`
		ReturnNumber string  `db:"return_number"`
		Status       string  `db:"status"`
		Reason       string  `db:"reason"`
		RefundAmount float64 `db:"refund_amount"`
		Notes        string  `db:"notes"`
		CreatedAt    string  `db:"created_at"`
	}
	query := `SELECT id, purchase_id, supplier_id, return_number, status, reason,
		refund_amount, COALESCE(notes, '') AS notes, created_at FROM supplier_returns`
	args := []interface{}{}
	if status != "" {
		if strings.EqualFold(s.db.DriverName(), "sqlite") {
			query += " WHERE status = ?"
		} else {
			query += " WHERE status = $1"
		}
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT 100"
	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("list supplier returns: %w", err)
	}
	out := make([]SupplierReturn, 0, len(rows))
	for _, row := range rows {
		createdAt, err := parseSQLiteTime(row.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return created_at: %w", err)
		}
		purchaseID, err := uuid.Parse(row.PurchaseID)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return purchase_id: %w", err)
		}
		supplierID, err := uuid.Parse(row.SupplierID)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return supplier_id: %w", err)
		}
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return id: %w", err)
		}
		out = append(out, SupplierReturn{
			ID:           id,
			PurchaseID:   purchaseID,
			SupplierID:   supplierID,
			ReturnNumber: row.ReturnNumber,
			Status:       row.Status,
			Reason:       row.Reason,
			RefundAmount: row.RefundAmount,
			Notes:        row.Notes,
			CreatedAt:    createdAt,
		})
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateRequest) (*SupplierReturn, error) {
	var r SupplierReturn
	if strings.EqualFold(s.db.DriverName(), "sqlite") {
		id := uuid.New()
		returnNumber := "SRET-" + strings.ToUpper(strings.ReplaceAll(id.String()[:10], "-", ""))
		query := `INSERT INTO supplier_returns
			(id, purchase_id, supplier_id, return_number, status, reason, notes, created_by, created_at, updated_at)
			SELECT ?, ?, supplier_id, ?, 'PENDING', ?, ?, ?, datetime('now'), datetime('now')
			FROM purchases WHERE id = ? AND status IN ('received', 'partially_received', 'completed')`
		result, err := s.db.ExecContext(ctx, query, id.String(), req.PurchaseID.String(), returnNumber, req.Reason, req.Notes, userID.String(), req.PurchaseID.String())
		if err != nil {
			return nil, fmt.Errorf("create supplier return: %w", err)
		}
		rows, err := result.RowsAffected()
		if err != nil || rows == 0 {
			return nil, fmt.Errorf("create supplier return: purchase is not eligible for supplier return")
		}
		purchaseRow := struct {
			ID           string  `db:"id"`
			PurchaseID   string  `db:"purchase_id"`
			SupplierID   string  `db:"supplier_id"`
			ReturnNumber string  `db:"return_number"`
			Status       string  `db:"status"`
			Reason       string  `db:"reason"`
			RefundAmount float64 `db:"refund_amount"`
			Notes        string  `db:"notes"`
			CreatedAt    string  `db:"created_at"`
		}{
			ID:           id.String(),
			PurchaseID:   req.PurchaseID.String(),
			SupplierID:   "",
			ReturnNumber: returnNumber,
			Status:       "PENDING",
			Reason:       req.Reason,
			RefundAmount: 0,
			Notes:        req.Notes,
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		}
		var supplierID string
		if err := s.db.GetContext(ctx, &supplierID, `SELECT supplier_id FROM purchases WHERE id = ?`, req.PurchaseID); err != nil {
			return nil, fmt.Errorf("fetch supplier id for supplier return: %w", err)
		}
		purchaseRow.SupplierID = supplierID
		createdAt, err := parseSQLiteTime(purchaseRow.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return created_at: %w", err)
		}
		pid, err := uuid.Parse(purchaseRow.PurchaseID)
		if err != nil {
			return nil, fmt.Errorf("parse purchase_id: %w", err)
		}
		sid, err := uuid.Parse(purchaseRow.SupplierID)
		if err != nil {
			return nil, fmt.Errorf("parse supplier_id: %w", err)
		}
		rid, err := uuid.Parse(purchaseRow.ID)
		if err != nil {
			return nil, fmt.Errorf("parse id: %w", err)
		}
		return &SupplierReturn{ID: rid, PurchaseID: pid, SupplierID: sid, ReturnNumber: purchaseRow.ReturnNumber, Status: purchaseRow.Status, Reason: purchaseRow.Reason, RefundAmount: purchaseRow.RefundAmount, Notes: purchaseRow.Notes, CreatedAt: createdAt}, nil
	}

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

func parseSQLiteTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported sqlite timestamp %q", value)
}

func (s *Service) AddItem(ctx context.Context, id uuid.UUID, req AddItemRequest) error {
	result, err := s.db.ExecContext(ctx, `INSERT INTO supplier_return_items
			(supplier_return_id, purchase_item_id, product_id, quantity, unit_cost)
			SELECT $1, pi.id, pi.product_id, $3, pi.unit_price
			FROM purchase_items pi
			JOIN supplier_returns sr ON sr.purchase_id = pi.purchase_id
			JOIN purchases p ON p.id = pi.purchase_id
			WHERE sr.id = $1 AND pi.id = $2
			AND p.status IN ('received', 'partially_received', 'completed')
			AND $3 <= (
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
	if dbutil.IsSQLite(s.db) {
		return s.completeSQLite(ctx, id, userID)
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin supplier return completion: %w", err)
	}
	defer tx.Rollback()
	var status string
	statusQuery := `SELECT status FROM supplier_returns WHERE id = $1`
	if !dbutil.IsSQLite(s.db) {
		statusQuery += " FOR UPDATE"
	}
	if err = tx.Get(&status, statusQuery, id); err != nil {
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
		availableQuery := `SELECT id FROM inventory_items
			WHERE product_id = $1 AND status = 'AVAILABLE'
			ORDER BY created_at`
		if !dbutil.IsSQLite(s.db) {
			availableQuery += " FOR UPDATE"
		}
		if err = tx.Select(&availableIDs, availableQuery, item.ProductID); err != nil {
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
			if _, err = tx.Exec(fmt.Sprintf(`UPDATE inventory_items SET status = 'RETURNED', updated_at = %s WHERE id = $1`, dbutil.NowSQL(s.db)), availableIDs[index]); err != nil {
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
	if _, err = tx.Exec(fmt.Sprintf(`UPDATE supplier_returns SET status = 'COMPLETED', refund_amount = $1, updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db)), total, id); err != nil {
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

// completeSQLite mirrors the PostgreSQL completion workflow without FOR
// UPDATE/casts. SQLite is the authoritative store for desktop operations, so
// it must record the same inventory, movement, and supplier-ledger effects.
func (s *Service) completeSQLite(ctx context.Context, id, userID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin supplier return completion: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	var status string
	if err = tx.GetContext(ctx, &status, `SELECT status FROM supplier_returns WHERE id = $1`, id); err != nil {
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
	if err = tx.SelectContext(ctx, &items, `SELECT product_id, quantity, unit_cost FROM supplier_return_items WHERE supplier_return_id = $1`, id); err != nil {
		return fmt.Errorf("get supplier return items: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("supplier return must contain at least one item")
	}
	var total float64
	for _, item := range items {
		var available []uuid.UUID
		if err = tx.SelectContext(ctx, &available, `SELECT id FROM inventory_items WHERE product_id = $1 AND status = 'AVAILABLE' ORDER BY created_at`, item.ProductID); err != nil {
			return fmt.Errorf("get available inventory items: %w", err)
		}
		if len(available) < item.Quantity {
			return fmt.Errorf("insufficient inventory for supplier return")
		}
		for _, itemID := range available[:item.Quantity] {
			var before int
			if err = tx.GetContext(ctx, &before, `SELECT COUNT(*) FROM inventory_items WHERE product_id = $1 AND status = 'AVAILABLE'`, item.ProductID); err != nil {
				return fmt.Errorf("count available inventory items: %w", err)
			}
			if _, err = tx.ExecContext(ctx, `UPDATE inventory_items SET status = 'RETURNED', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, itemID); err != nil {
				return fmt.Errorf("update inventory item: %w", err)
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at) VALUES ($1, $2, $3, 'SUPPLIER_RETURN', -1, $4, $5, 'supplier_return', $6, 'Supplier return', $7, CURRENT_TIMESTAMP)`, uuid.New(), itemID, item.ProductID, before, before-1, id, userID); err != nil {
				return fmt.Errorf("record inventory movement: %w", err)
			}
		}
		_, _ = tx.ExecContext(ctx, `UPDATE inventory SET quantity = CASE WHEN quantity > $1 THEN quantity - $1 ELSE 0 END, updated_at = CURRENT_TIMESTAMP WHERE product_id = $2`, item.Quantity, item.ProductID)
		total += float64(item.Quantity) * item.UnitCost
	}
	if _, err = tx.ExecContext(ctx, `UPDATE supplier_returns SET status = 'COMPLETED', refund_amount = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, total, id); err != nil {
		return fmt.Errorf("complete supplier return: %w", err)
	}
	var supplierID uuid.UUID
	if err = tx.GetContext(ctx, &supplierID, `SELECT supplier_id FROM supplier_returns WHERE id = $1`, id); err != nil {
		return fmt.Errorf("get supplier: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT $1, $2, 'credit', 'SUPPLIER_RETURN', $3, COALESCE((SELECT balance FROM supplier_ledger WHERE supplier_id = $2 ORDER BY created_at DESC LIMIT 1), 0) - $3, 'Supplier return ' || $4, $5, CURRENT_TIMESTAMP`, uuid.New(), supplierID, total, id.String(), id); err != nil {
		return fmt.Errorf("record supplier ledger entry: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit supplier return: %w", err)
	}
	committed = true
	return nil
}
