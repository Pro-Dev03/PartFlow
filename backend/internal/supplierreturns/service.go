package supplierreturns

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type Service struct{ db *sqlx.DB }
type Handler struct{ service *Service }

type SupplierReturn struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	CustomerReturnID      uuid.UUID `json:"customer_return_id" db:"customer_return_id"`
	SaleID                uuid.UUID `json:"sale_id" db:"sale_id"`
	PurchaseID            uuid.UUID `json:"purchase_id" db:"purchase_id"`
	SupplierID            uuid.UUID `json:"supplier_id" db:"supplier_id"`
	SupplierName          string    `json:"supplier_name" db:"supplier_name"`
	ProductName           string    `json:"product_name" db:"product_name"`
	InventoryItemID       uuid.UUID `json:"inventory_item_id" db:"inventory_item_id"`
	Barcode               string    `json:"barcode" db:"barcode"`
	SerialNumber          string    `json:"serial_number" db:"serial_number"`
	Quantity              int       `json:"quantity" db:"quantity"`
	PurchaseCost          float64   `json:"purchase_cost" db:"purchase_cost"`
	ReturnReason          string    `json:"return_reason" db:"return_reason"`
	ReturnDate            time.Time `json:"return_date" db:"return_date"`
	NeedsSourceResolution bool      `json:"needs_source_resolution" db:"needs_source_resolution"`
	Source                string    `json:"source" db:"source"`
	ReturnNumber          string    `json:"return_number" db:"return_number"`
	Status                string    `json:"status" db:"status"`
	SourceStatus          string    `json:"source_status" db:"source_status"`
	Reason                string    `json:"reason" db:"reason"`
	RefundAmount          float64   `json:"refund_amount" db:"refund_amount"`
	Notes                 string    `json:"notes" db:"notes"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
}

type CreateRequest struct {
	PurchaseID     uuid.UUID `json:"purchase_id" binding:"required"`
	Reason         string    `json:"reason" binding:"required"`
	Notes          string    `json:"notes"`
	PurchaseItemID uuid.UUID `json:"purchase_item_id"`
	Quantity       int       `json:"quantity" binding:"omitempty,min=1"`
}
type AddItemRequest struct {
	PurchaseItemID uuid.UUID `json:"purchase_item_id" binding:"required"`
	Quantity       int       `json:"quantity" binding:"required,min=1"`
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, force bool, actorIDs ...uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete supplier return: %w", err)
	}
	defer tx.Rollback()

	lockQuery := `SELECT id::text FROM supplier_returns WHERE id = $1 FOR UPDATE`
	if dbutil.IsSQLite(s.db) {
		lockQuery = `UPDATE supplier_returns SET updated_at = updated_at WHERE id = ?`
	}
	if dbutil.IsSQLite(s.db) {
		result, lockErr := tx.ExecContext(ctx, lockQuery, id)
		if lockErr != nil {
			return fmt.Errorf("lock supplier return for deletion: %w", lockErr)
		}
		if count, rowsErr := result.RowsAffected(); rowsErr != nil || count != 1 {
			return fmt.Errorf("get supplier return source: supplier return not found")
		}
	} else {
		var lockedID string
		if err = tx.GetContext(ctx, &lockedID, lockQuery, id); err != nil {
			return fmt.Errorf("lock supplier return for deletion: %w", err)
		}
	}

	var status string
	if err = tx.GetContext(ctx, &status, tx.Rebind(`SELECT status FROM supplier_returns WHERE id = ?`), id); err != nil {
		return fmt.Errorf("get supplier return source: %w", err)
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	var actor interface{}
	if len(actorIDs) > 0 && actorIDs[0] != uuid.Nil {
		actor = actorIDs[0]
	}
	var snapshotJSON string
	switch status {
	case "DRAFT", "PENDING", "REJECTED":
		// These states have not posted financial or stock effects. A linked
		// customer return does not prevent cleanup; its own record remains.
	case "COMPLETED":
		snapshotJSON, err = s.snapshotCompletedSupplierReturn(ctx, tx, id)
		if err != nil {
			return err
		}
		if dbutil.IsSQLite(s.db) {
			_, err = tx.ExecContext(ctx, `INSERT INTO deleted_operation_snapshots (entity_type, operation_id, business_number, status, snapshot, deleted_by) VALUES ('supplier_return', ?, (SELECT return_number FROM supplier_returns WHERE id = ?), ?, ?, ?)`, id.String(), id.String(), status, snapshotJSON, actor)
		} else {
			_, err = tx.ExecContext(ctx, `INSERT INTO deleted_operation_snapshots (entity_type, operation_id, business_number, status, snapshot, deleted_by) SELECT 'supplier_return', $1, return_number, $2, $3::jsonb, $4 FROM supplier_returns WHERE id = $1`, id.String(), status, snapshotJSON, actor)
		}
		if err != nil {
			return fmt.Errorf("preserve completed supplier return history: %w", err)
		}
		// The posted ledger and stock rows remain intact. Their reference now
		// resolves to the durable cleanup snapshot instead of the removed row.
		for _, table := range []string{"inventory_movements", "supplier_ledger"} {
			if !dbutil.IsSQLite(s.db) && table == "supplier_ledger" {
				_, err = tx.ExecContext(ctx, `UPDATE supplier_ledger SET reference_type = 'supplier_return_effect' WHERE reference_id = $1 AND UPPER(COALESCE(transaction_type,'')) = 'SUPPLIER_RETURN'`, id)
			} else if !dbutil.IsSQLite(s.db) {
				_, err = tx.ExecContext(ctx, `UPDATE inventory_movements SET reference_type = 'supplier_return_effect' WHERE reference_id = $1 AND UPPER(COALESCE(movement_type,'')) = 'SUPPLIER_RETURN'`, id)
			} else if table == "supplier_ledger" {
				_, err = tx.ExecContext(ctx, `UPDATE supplier_ledger SET reference_type = 'supplier_return_effect' WHERE reference_id = ? AND UPPER(COALESCE(transaction_type,'')) = 'SUPPLIER_RETURN'`, id.String())
			} else {
				_, err = tx.ExecContext(ctx, `UPDATE inventory_movements SET reference_type = 'supplier_return_effect' WHERE reference_id = ? AND UPPER(COALESCE(movement_type,'')) = 'SUPPLIER_RETURN'`, id.String())
			}
			if err != nil {
				return fmt.Errorf("detach supplier return %s history: %w", table, err)
			}
		}
	default:
		return fmt.Errorf("supplier return is still in %s processing; complete or reject it before cleanup", status)
	}

	// force remains accepted for API compatibility; linked historical records
	// are now handled by explicit state-aware cleanup instead of a bypass flag.
	_ = force
	if status == "COMPLETED" {
		auditJSON, _ := json.Marshal(map[string]any{"financial_effect_preserved": true, "snapshot_entity": "supplier_return", "status": status})
		if dbutil.IsSQLite(s.db) {
			_, err = tx.ExecContext(ctx, `INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at) VALUES (?, ?, 'DELETE', 'supplier_return', ?, ?, CURRENT_TIMESTAMP)`, uuid.NewString(), actor, id.String(), string(auditJSON))
		} else {
			_, err = tx.ExecContext(ctx, `INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at) VALUES ($1, $2, 'DELETE', 'supplier_return', $3, $4::jsonb, NOW())`, uuid.New(), actor, id, string(auditJSON))
		}
		if err != nil {
			return fmt.Errorf("write supplier return deletion audit: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, tx.Rebind(`DELETE FROM supplier_return_items WHERE supplier_return_id = ?`), id); err != nil {
		return fmt.Errorf("delete supplier return items: %w", err)
	}
	deleteQuery := `DELETE FROM supplier_returns WHERE id = ? AND UPPER(status) IN ('DRAFT', 'PENDING', 'REJECTED', 'COMPLETED')`
	result, err := tx.ExecContext(ctx, tx.Rebind(deleteQuery), id)
	if err != nil {
		return fmt.Errorf("delete supplier return: %w", err)
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		return fmt.Errorf("supplier return could not be cleaned up in its current state")
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit delete supplier return: %w", err)
	}
	return nil
}

func (s *Service) snapshotCompletedSupplierReturn(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (string, error) {
	var quantity, movementCount, movementQuantity, ledgerCount int
	var ledgerAmount, refundAmount float64
	if err := tx.GetContext(ctx, &refundAmount, tx.Rebind(`SELECT refund_amount FROM supplier_returns WHERE id = ?`), id); err != nil {
		return "", fmt.Errorf("load completed supplier return amount: %w", err)
	}
	if err := tx.GetContext(ctx, &quantity, tx.Rebind(`SELECT COALESCE(SUM(quantity), 0) FROM supplier_return_items WHERE supplier_return_id = ?`), id); err != nil {
		return "", fmt.Errorf("validate supplier return items: %w", err)
	}
	if quantity <= 0 {
		return "", fmt.Errorf("completed supplier return has no item history; cleanup was stopped to protect stock integrity")
	}
	if err := tx.GetContext(ctx, &movementCount, tx.Rebind(`SELECT COUNT(*) FROM inventory_movements WHERE reference_id = ? AND UPPER(COALESCE(movement_type,'')) = 'SUPPLIER_RETURN'`), id); err != nil {
		return "", fmt.Errorf("validate supplier return stock history: %w", err)
	}
	if err := tx.GetContext(ctx, &movementQuantity, tx.Rebind(`SELECT COALESCE(SUM(quantity),0) FROM inventory_movements WHERE reference_id = ? AND UPPER(COALESCE(movement_type,'')) = 'SUPPLIER_RETURN'`), id); err != nil {
		return "", fmt.Errorf("validate supplier return stock quantity: %w", err)
	}
	if movementCount != quantity || movementQuantity != -quantity {
		return "", fmt.Errorf("completed supplier return stock history is incomplete (expected %d units, found %d); reconcile its history before cleanup", quantity, movementCount)
	}
	if err := tx.GetContext(ctx, &ledgerCount, tx.Rebind(`SELECT COUNT(*) FROM supplier_ledger WHERE reference_id = ? AND UPPER(COALESCE(transaction_type,'')) = 'SUPPLIER_RETURN'`), id); err != nil {
		return "", fmt.Errorf("validate supplier return financial history: %w", err)
	}
	if err := tx.GetContext(ctx, &ledgerAmount, tx.Rebind(`SELECT COALESCE(SUM(amount),0) FROM supplier_ledger WHERE reference_id = ? AND UPPER(COALESCE(transaction_type,'')) = 'SUPPLIER_RETURN'`), id); err != nil {
		return "", fmt.Errorf("validate supplier return ledger amount: %w", err)
	}
	if ledgerCount != 1 || math.Abs(ledgerAmount-refundAmount) > 0.01 {
		return "", fmt.Errorf("completed supplier return supplier-ledger history does not match its refund amount; reconcile its history before cleanup")
	}

	type header struct {
		ID               string  `db:"id"`
		CustomerReturnID string  `db:"customer_return_id"`
		SaleID           string  `db:"sale_id"`
		PurchaseID       string  `db:"purchase_id"`
		SupplierID       string  `db:"supplier_id"`
		ReturnNumber     string  `db:"return_number"`
		Status           string  `db:"status"`
		SourceStatus     string  `db:"source_status"`
		Reason           string  `db:"reason"`
		Notes            string  `db:"notes"`
		CreatedBy        string  `db:"created_by"`
		ReturnReason     string  `db:"return_reason"`
		ReturnDate       string  `db:"return_date"`
		CreatedAt        string  `db:"created_at"`
		UpdatedAt        string  `db:"updated_at"`
		RefundAmount     float64 `db:"refund_amount"`
	}
	var h header
	query := `SELECT CAST(id AS TEXT) AS id, COALESCE(CAST(customer_return_id AS TEXT), '') AS customer_return_id, COALESCE(CAST(sale_id AS TEXT), '') AS sale_id, CAST(purchase_id AS TEXT) AS purchase_id, CAST(supplier_id AS TEXT) AS supplier_id, return_number, status, COALESCE(source_status,'') AS source_status, reason, COALESCE(notes,'') AS notes, COALESCE(CAST(created_by AS TEXT),'') AS created_by, COALESCE(return_reason,'') AS return_reason, COALESCE(CAST(return_date AS TEXT),'') AS return_date, CAST(created_at AS TEXT) AS created_at, CAST(updated_at AS TEXT) AS updated_at, refund_amount FROM supplier_returns WHERE id = ?`
	if err := tx.GetContext(ctx, &h, tx.Rebind(query), id); err != nil {
		return "", fmt.Errorf("snapshot supplier return header: %w", err)
	}
	type item struct {
		ID               string  `db:"id"`
		CustomerReturnID string  `db:"customer_return_id"`
		SaleID           string  `db:"sale_id"`
		SaleItemID       string  `db:"sale_item_id"`
		InventoryItemID  string  `db:"inventory_item_id"`
		PurchaseItemID   string  `db:"purchase_item_id"`
		ProductID        string  `db:"product_id"`
		Barcode          string  `db:"barcode"`
		SerialNumber     string  `db:"serial_number"`
		ReturnReason     string  `db:"return_reason"`
		ReturnDate       string  `db:"return_date"`
		CreatedAt        string  `db:"created_at"`
		Quantity         int     `db:"quantity"`
		UnitCost         float64 `db:"unit_cost"`
		PurchaseCost     float64 `db:"purchase_cost"`
	}
	items := make([]item, 0)
	itemQuery := `SELECT COALESCE(CAST(id AS TEXT),'') AS id, COALESCE(CAST(customer_return_id AS TEXT),'') AS customer_return_id, COALESCE(CAST(sale_id AS TEXT),'') AS sale_id, COALESCE(CAST(sale_item_id AS TEXT),'') AS sale_item_id, COALESCE(CAST(inventory_item_id AS TEXT),'') AS inventory_item_id, CAST(purchase_item_id AS TEXT) AS purchase_item_id, CAST(product_id AS TEXT) AS product_id, COALESCE(barcode,'') AS barcode, COALESCE(serial_number,'') AS serial_number, COALESCE(return_reason,'') AS return_reason, COALESCE(CAST(return_date AS TEXT),'') AS return_date, CAST(created_at AS TEXT) AS created_at, quantity, unit_cost, COALESCE(purchase_cost,0) AS purchase_cost FROM supplier_return_items WHERE supplier_return_id = ? ORDER BY created_at, id`
	if err := tx.SelectContext(ctx, &items, tx.Rebind(itemQuery), id); err != nil {
		return "", fmt.Errorf("snapshot supplier return items: %w", err)
	}
	encoded, err := json.Marshal(map[string]any{"supplier_return": h, "items": items, "posted_effects": map[string]any{"stock_movement_count": movementCount, "stock_quantity": movementQuantity, "supplier_ledger_count": ledgerCount, "supplier_ledger_amount": ledgerAmount}})
	if err != nil {
		return "", fmt.Errorf("encode supplier return cleanup snapshot: %w", err)
	}
	return string(encoded), nil
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
		ID               string  `db:"id"`
		CustomerReturnID string  `db:"customer_return_id"`
		SaleID           string  `db:"sale_id"`
		PurchaseID       string  `db:"purchase_id"`
		SupplierID       string  `db:"supplier_id"`
		SupplierName     string  `db:"supplier_name"`
		ProductName      string  `db:"product_name"`
		InventoryItemID  string  `db:"inventory_item_id"`
		Barcode          string  `db:"barcode"`
		SerialNumber     string  `db:"serial_number"`
		ReturnNumber     string  `db:"return_number"`
		Status           string  `db:"status"`
		Reason           string  `db:"reason"`
		RefundAmount     float64 `db:"refund_amount"`
		Notes            string  `db:"notes"`
		CreatedAt        string  `db:"created_at"`
		SourceStatus     string  `db:"source_status"`
		ReturnDate       string  `db:"return_date"`
		ReturnReason     string  `db:"return_reason"`
		Quantity         int     `db:"quantity"`
		PurchaseCost     float64 `db:"purchase_cost"`
	}
	query := `SELECT sr.id, COALESCE(sr.customer_return_id, '') AS customer_return_id,
		COALESCE(sr.sale_id, '') AS sale_id, COALESCE(sr.purchase_id, '') AS purchase_id, COALESCE(sr.supplier_id, '') AS supplier_id,
		COALESCE(s.name, '') AS supplier_name,
		COALESCE((SELECT p0.name FROM supplier_return_items sri0 JOIN products p0 ON p0.id = sri0.product_id WHERE sri0.supplier_return_id = sr.id ORDER BY sri0.created_at LIMIT 1),
			(SELECT p1.name FROM supplier_return_items sri1 JOIN purchase_items pi1 ON pi1.id = sri1.purchase_item_id JOIN products p1 ON p1.id = pi1.product_id WHERE sri1.supplier_return_id = sr.id ORDER BY sri1.created_at LIMIT 1),
			(SELECT p2.name FROM return_items ri2 JOIN products p2 ON p2.id = ri2.product_id WHERE ri2.return_id = sr.customer_return_id ORDER BY ri2.created_at LIMIT 1),
			(SELECT p3.name FROM return_items ri3 JOIN inventory_items ii3 ON ii3.id = ri3.inventory_item_id JOIN products p3 ON p3.id = ii3.product_id WHERE ri3.return_id = sr.customer_return_id ORDER BY ri3.created_at LIMIT 1),
			(SELECT p4.name FROM return_items ri4 JOIN inventory_items ii4 ON ii4.barcode = ri4.barcode JOIN products p4 ON p4.id = ii4.product_id WHERE ri4.return_id = sr.customer_return_id ORDER BY ri4.created_at LIMIT 1),
			(SELECT p5.name FROM return_items ri5 JOIN products p5 ON p5.barcode = ri5.barcode WHERE ri5.return_id = sr.customer_return_id ORDER BY ri5.created_at LIMIT 1), '') AS product_name,
		COALESCE((SELECT sri.inventory_item_id FROM supplier_return_items sri WHERE sri.supplier_return_id = sr.id ORDER BY sri.created_at LIMIT 1), (SELECT ri.inventory_item_id FROM return_items ri WHERE ri.return_id = sr.customer_return_id ORDER BY ri.created_at LIMIT 1), '') AS inventory_item_id,
		COALESCE((SELECT sri.barcode FROM supplier_return_items sri WHERE sri.supplier_return_id = sr.id ORDER BY sri.created_at LIMIT 1), (SELECT ri.barcode FROM return_items ri WHERE ri.return_id = sr.customer_return_id ORDER BY ri.created_at LIMIT 1), '') AS barcode,
		COALESCE((SELECT sri.serial_number FROM supplier_return_items sri WHERE sri.supplier_return_id = sr.id ORDER BY sri.created_at LIMIT 1), (SELECT ri.serial_number FROM return_items ri WHERE ri.return_id = sr.customer_return_id ORDER BY ri.created_at LIMIT 1), '') AS serial_number,
		sr.return_number, sr.status, COALESCE(sr.source_status, 'RESOLVED') AS source_status, sr.reason, sr.refund_amount, COALESCE(sr.notes, '') AS notes,
		COALESCE(sr.return_reason, sr.reason, '') AS return_reason, COALESCE(sr.return_date, sr.created_at) AS return_date,
		COALESCE((SELECT sri.quantity FROM supplier_return_items sri WHERE sri.supplier_return_id = sr.id ORDER BY sri.created_at LIMIT 1), (SELECT ri.quantity_returned FROM return_items ri WHERE ri.return_id = sr.customer_return_id ORDER BY ri.created_at LIMIT 1), 0) AS quantity,
		COALESCE((SELECT sri.purchase_cost FROM supplier_return_items sri WHERE sri.supplier_return_id = sr.id ORDER BY sri.created_at LIMIT 1), (SELECT ri.original_cost FROM return_items ri WHERE ri.return_id = sr.customer_return_id ORDER BY ri.created_at LIMIT 1), 0) AS purchase_cost,
		sr.created_at FROM supplier_returns sr LEFT JOIN suppliers s ON s.id = sr.supplier_id`
	if strings.EqualFold(s.db.DriverName(), "sqlite") && (!sqliteTableExists(ctx, s.db, "suppliers") || !sqliteTableExists(ctx, s.db, "returns") || !sqliteColumnExists(ctx, s.db, "supplier_returns", "source_status")) {
		query = `SELECT sr.id, COALESCE(sr.customer_return_id, '') AS customer_return_id,
			COALESCE(sr.sale_id, '') AS sale_id, COALESCE(sr.purchase_id, '') AS purchase_id, COALESCE(sr.supplier_id, '') AS supplier_id,
			'' AS supplier_name, '' AS product_name, COALESCE((SELECT sri.inventory_item_id FROM supplier_return_items sri WHERE sri.supplier_return_id = sr.id LIMIT 1), '') AS inventory_item_id,
			COALESCE((SELECT sri.barcode FROM supplier_return_items sri WHERE sri.supplier_return_id = sr.id LIMIT 1), '') AS barcode,
			COALESCE((SELECT sri.serial_number FROM supplier_return_items sri WHERE sri.supplier_return_id = sr.id LIMIT 1), '') AS serial_number,
			sr.return_number, sr.status, '' AS source_status, sr.reason, sr.refund_amount, COALESCE(sr.notes, '') AS notes,
			'' AS return_reason, sr.created_at AS return_date, 0 AS quantity, 0 AS purchase_cost, sr.created_at
			FROM supplier_returns sr`
	}
	if !strings.EqualFold(s.db.DriverName(), "sqlite") {
		query = strings.ReplaceAll(query, "COALESCE(sr.customer_return_id, '')", "COALESCE(sr.customer_return_id::text, '')")
		query = strings.ReplaceAll(query, "COALESCE(sr.sale_id, '')", "COALESCE(sr.sale_id::text, '')")
		query = strings.ReplaceAll(query, "COALESCE(sr.purchase_id, '')", "COALESCE(sr.purchase_id::text, '')")
		query = strings.ReplaceAll(query, "COALESCE(sr.supplier_id, '')", "COALESCE(sr.supplier_id::text, '')")
		query = strings.ReplaceAll(query, "COALESCE((SELECT sri.inventory_item_id", "COALESCE((SELECT sri.inventory_item_id::text")
	}
	args := []interface{}{}
	if status != "" {
		if strings.EqualFold(s.db.DriverName(), "sqlite") {
			query += " WHERE status = ?"
		} else {
			query += " WHERE status = $1"
		}
		args = append(args, status)
	}
	query += " ORDER BY sr.created_at DESC LIMIT 100"
	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("list supplier returns: %w", err)
	}
	out := make([]SupplierReturn, 0, len(rows))
	for _, row := range rows {
		createdAt, err := parseSQLiteTime(row.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return created_at: %w", err)
		}
		purchaseID, err := parseOptionalUUID(row.PurchaseID)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return purchase_id: %w", err)
		}
		supplierID, err := parseOptionalUUID(row.SupplierID)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return supplier_id: %w", err)
		}
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return id: %w", err)
		}
		customerReturnID, err := parseOptionalUUID(row.CustomerReturnID)
		if err != nil {
			return nil, fmt.Errorf("parse customer_return_id: %w", err)
		}
		saleID, err := parseOptionalUUID(row.SaleID)
		if err != nil {
			return nil, fmt.Errorf("parse sale_id: %w", err)
		}
		inventoryItemID, err := parseOptionalUUID(row.InventoryItemID)
		if err != nil {
			return nil, fmt.Errorf("parse inventory_item_id: %w", err)
		}
		returnDate, err := parseSQLiteTime(row.ReturnDate)
		if err != nil {
			return nil, fmt.Errorf("parse supplier return date: %w", err)
		}
		out = append(out, SupplierReturn{
			ID: id, CustomerReturnID: customerReturnID, SaleID: saleID, PurchaseID: purchaseID,
			SupplierID: supplierID, SupplierName: row.SupplierName, ProductName: row.ProductName, InventoryItemID: inventoryItemID, Barcode: row.Barcode,
			SerialNumber: row.SerialNumber, Quantity: row.Quantity, PurchaseCost: row.PurchaseCost, ReturnReason: row.ReturnReason, ReturnDate: returnDate,
			ReturnNumber: row.ReturnNumber, Status: row.Status, SourceStatus: row.SourceStatus, NeedsSourceResolution: row.SourceStatus == "NEEDS_SOURCE_DATA" || row.Status == "NEEDS_SOURCE_DATA",
			Source: func() string {
				if customerReturnID != uuid.Nil {
					return "Customer Return"
				}
				return "Supplier Return"
			}(), Reason: row.Reason, RefundAmount: row.RefundAmount, Notes: row.Notes, CreatedAt: createdAt,
		})
	}
	return out, nil
}

func sqliteTableExists(ctx context.Context, db *sqlx.DB, table string) bool {
	var count int
	return db.GetContext(ctx, &count, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table) == nil && count > 0
}

func sqliteColumnExists(ctx context.Context, db *sqlx.DB, table, column string) bool {
	rows, err := db.QueryxContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, pk int
		var name, dataType string
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err == nil && strings.EqualFold(name, column) {
			return true
		}
	}
	return false
}

func parseOptionalUUID(value string) (uuid.UUID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "00000000-0000-0000-0000-000000000000" {
		return uuid.Nil, nil
	}
	return uuid.Parse(trimmed)
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateRequest) (*SupplierReturn, error) {
	hasPurchaseItem := req.PurchaseItemID != uuid.Nil
	if hasPurchaseItem != (req.Quantity > 0) {
		return nil, fmt.Errorf("purchase item and a positive quantity must be provided together")
	}
	if req.Quantity < 0 {
		return nil, fmt.Errorf("supplier return quantity must be positive")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin supplier return creation: %w", err)
	}
	defer tx.Rollback()

	var r SupplierReturn
	if strings.EqualFold(s.db.DriverName(), "sqlite") {
		id := uuid.New()
		returnNumber := "SRET-" + strings.ToUpper(strings.ReplaceAll(id.String()[:10], "-", ""))
		query := `INSERT INTO supplier_returns
			(id, purchase_id, supplier_id, return_number, status, reason, notes, created_by, created_at, updated_at)
			SELECT ?, ?, supplier_id, ?, 'PENDING', ?, ?, ?, datetime('now'), datetime('now')
			FROM purchases WHERE id = ? AND status IN ('received', 'partially_received', 'completed')`
		result, err := tx.ExecContext(ctx, query, id.String(), req.PurchaseID.String(), returnNumber, req.Reason, req.Notes, userID.String(), req.PurchaseID.String())
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
		if err := tx.GetContext(ctx, &supplierID, `SELECT supplier_id FROM purchases WHERE id = ?`, req.PurchaseID); err != nil {
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
		r = SupplierReturn{ID: rid, PurchaseID: pid, SupplierID: sid, ReturnNumber: purchaseRow.ReturnNumber, Status: purchaseRow.Status, Reason: purchaseRow.Reason, RefundAmount: purchaseRow.RefundAmount, Notes: purchaseRow.Notes, CreatedAt: createdAt}
	} else {
		err := tx.GetContext(ctx, &r, `INSERT INTO supplier_returns
			(purchase_id, supplier_id, return_number, status, reason, notes, created_by)
			SELECT $1, supplier_id, 'SRET-' || upper(substr(replace(uuid_generate_v4()::text, '-', ''), 1, 10)),
			'PENDING', $2, $3, $4 FROM purchases WHERE id = $1
			AND status IN ('received', 'partially_received', 'completed')
			RETURNING id, purchase_id, supplier_id, return_number, status, reason, refund_amount, COALESCE(notes, '') AS notes, created_at`,
			req.PurchaseID, req.Reason, req.Notes, userID)
		if err != nil {
			return nil, fmt.Errorf("create supplier return: %w", err)
		}
	}

	if hasPurchaseItem {
		if err := addSupplierReturnItem(ctx, tx, r.ID, AddItemRequest{PurchaseItemID: req.PurchaseItemID, Quantity: req.Quantity}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit supplier return creation: %w", err)
	}
	return &r, nil
}

func parseSQLiteTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05 -0700 MST", "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported sqlite timestamp %q", value)
}

func (s *Service) AddItem(ctx context.Context, id uuid.UUID, req AddItemRequest) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin supplier return item creation: %w", err)
	}
	defer tx.Rollback()
	if err := addSupplierReturnItem(ctx, tx, id, req); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit supplier return item: %w", err)
	}
	return nil
}

func addSupplierReturnItem(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, req AddItemRequest) error {
	if req.PurchaseItemID == uuid.Nil || req.Quantity < 1 {
		return fmt.Errorf("supplier return item and a positive quantity are required")
	}
	if !strings.EqualFold(tx.DriverName(), "sqlite") {
		var lockedPurchaseID string
		if err := tx.GetContext(ctx, &lockedPurchaseID, `
			SELECT p.id::text
			FROM purchases p
			JOIN supplier_returns sr ON sr.purchase_id = p.id
			JOIN purchase_items pi ON pi.purchase_id = p.id
			WHERE sr.id = $1 AND pi.id = $2 AND sr.status IN ('DRAFT', 'PENDING')
			FOR UPDATE OF p`, id, req.PurchaseItemID); err != nil {
			return fmt.Errorf("supplier return is not open or purchase item is unavailable: %w", err)
		}
	}

	query := `INSERT INTO supplier_return_items
			(id, supplier_return_id, purchase_item_id, product_id, quantity, unit_cost)
			SELECT $4, $1, pi.id, pi.product_id, $3, pi.unit_price
			FROM purchase_items pi
			JOIN supplier_returns sr ON sr.purchase_id = pi.purchase_id
			JOIN purchases p ON p.id = pi.purchase_id
			WHERE sr.id = $1 AND pi.id = $2
			AND sr.status IN ('DRAFT', 'PENDING')
			AND p.status IN ('received', 'partially_received', 'completed')
			AND $3 <= (
				SELECT COUNT(*) FROM inventory_items ii
				WHERE ii.product_id = pi.product_id
				AND ii.item_code LIKE 'ITM-' || substr(replace(CAST(pi.purchase_id AS TEXT), '-', ''), 1, 8) || '-%'
				AND ii.status = 'AVAILABLE'
			)
			- COALESCE((SELECT SUM(sri.quantity) FROM supplier_return_items sri
				JOIN supplier_returns existing_sr ON existing_sr.id = sri.supplier_return_id
				JOIN purchase_items existing_pi ON existing_pi.id = sri.purchase_item_id
				WHERE existing_pi.purchase_id = pi.purchase_id
				AND existing_pi.product_id = pi.product_id
				AND existing_sr.status IN ('PENDING', 'SHIPPED', 'RECEIVED')), 0)`
	result, err := tx.ExecContext(ctx, tx.Rebind(query),
		id, req.PurchaseItemID, req.Quantity, uuid.New())
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
		ProductID       uuid.UUID `db:"product_id"`
		InventoryItemID string    `db:"inventory_item_id"`
		Quantity        int       `db:"quantity"`
		UnitCost        float64   `db:"unit_cost"`
	}
	if err = tx.Select(&items, `SELECT product_id, COALESCE(inventory_item_id::text, '') AS inventory_item_id, quantity, unit_cost FROM supplier_return_items WHERE supplier_return_id = $1`, id); err != nil {
		return fmt.Errorf("get supplier return items: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("supplier return must contain at least one item")
	}
	var total float64
	for _, item := range items {
		var availableIDs []uuid.UUID
		if item.InventoryItemID != "" {
			if err = tx.Select(&availableIDs, `SELECT id FROM inventory_items WHERE id = $1 AND product_id = $2 AND status IN ('AVAILABLE', 'RETURNED') FOR UPDATE`, item.InventoryItemID, item.ProductID); err != nil {
				return fmt.Errorf("get linked inventory item: %w", err)
			}
		} else if err = tx.Select(&availableIDs, `SELECT id FROM inventory_items WHERE product_id = $1 AND status = 'AVAILABLE' AND item_code LIKE 'ITM-' || substr(replace((SELECT sr.purchase_id FROM supplier_returns sr WHERE sr.id = $2)::text, '-', ''), 1, 8) || '-%' ORDER BY created_at FOR UPDATE`, item.ProductID, id); err != nil {
			return fmt.Errorf("get available inventory items: %w", err)
		}
		if len(availableIDs) < item.Quantity {
			return fmt.Errorf("insufficient inventory for supplier return")
		}
		for index := 0; index < item.Quantity; index++ {
			var before int
			if err = tx.Get(&before, `SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id = $1 AND ii.status IN ('AVAILABLE', 'RETURNED')`, item.ProductID); err != nil {
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

		var aggregateTotal int
		if err = tx.Get(&aggregateTotal, `SELECT COALESCE(SUM(quantity), 0) FROM inventory WHERE product_id = $1`, item.ProductID); err != nil {
			return fmt.Errorf("get aggregate product inventory: %w", err)
		}
		var inventoryRows []struct {
			ID       uuid.UUID `db:"id"`
			Quantity int       `db:"quantity"`
		}
		if err = tx.Select(&inventoryRows, `SELECT id, quantity FROM inventory WHERE product_id = $1 ORDER BY quantity DESC, created_at ASC FOR UPDATE`, item.ProductID); err != nil {
			return fmt.Errorf("get inventory rows for product: %w", err)
		}
		remaining := item.Quantity
		for _, row := range inventoryRows {
			if remaining <= 0 {
				break
			}
			if row.Quantity <= 0 {
				continue
			}
			deducted := row.Quantity
			if deducted > remaining {
				deducted = remaining
			}
			if _, err = tx.Exec(fmt.Sprintf(`UPDATE inventory SET quantity = quantity - $1, updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db)), deducted, row.ID); err != nil {
				return fmt.Errorf("update aggregate inventory: %w", err)
			}
			remaining -= deducted
		}
		if remaining != 0 && aggregateTotal < item.Quantity && len(availableIDs) >= item.Quantity {
			remaining = 0
		}
		if remaining != 0 {
			return fmt.Errorf("insufficient aggregate inventory for supplier return")
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
			COALESCE((SELECT SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END) FROM supplier_ledger WHERE supplier_id = $1), 0) - $2,
			'مرتجع مورد ' || $3, $4`, supplierID, total, id.String(), id); err != nil {
		return fmt.Errorf("record supplier ledger entry: %w", err)
	}
	if _, err = tx.Exec(`UPDATE suppliers SET current_balance = COALESCE(current_balance, 0) - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, total, supplierID); err != nil {
		return fmt.Errorf("update supplier balance after supplier return: %w", err)
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
		ProductID       uuid.UUID `db:"product_id"`
		InventoryItemID string    `db:"inventory_item_id"`
		Quantity        int       `db:"quantity"`
		UnitCost        float64   `db:"unit_cost"`
	}
	if err = tx.SelectContext(ctx, &items, `SELECT product_id, COALESCE(inventory_item_id, '') AS inventory_item_id, quantity, unit_cost FROM supplier_return_items WHERE supplier_return_id = $1`, id); err != nil {
		return fmt.Errorf("get supplier return items: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("supplier return must contain at least one item")
	}
	var total float64
	for _, item := range items {
		var available []uuid.UUID
		if item.InventoryItemID != "" {
			if err = tx.SelectContext(ctx, &available, `SELECT id FROM inventory_items WHERE id = $1 AND product_id = $2 AND status IN ('AVAILABLE', 'RETURNED')`, item.InventoryItemID, item.ProductID); err != nil {
				return fmt.Errorf("get linked inventory item: %w", err)
			}
		} else if err = tx.SelectContext(ctx, &available, `SELECT id FROM inventory_items WHERE product_id = $1 AND status = 'AVAILABLE' AND item_code LIKE 'ITM-' || substr(replace((SELECT p.id FROM purchases p JOIN supplier_returns sr ON sr.purchase_id = p.id WHERE sr.id = $2), '-', ''), 1, 8) || '-%' ORDER BY created_at`, item.ProductID, id); err != nil {
			return fmt.Errorf("get available inventory items: %w", err)
		}
		if len(available) < item.Quantity {
			return fmt.Errorf("insufficient inventory for supplier return")
		}
		for _, itemID := range available[:item.Quantity] {
			var before int
			if err = tx.GetContext(ctx, &before, `SELECT COUNT(*) FROM inventory_items WHERE product_id = $1 AND status IN ('AVAILABLE', 'RETURNED')`, item.ProductID); err != nil {
				return fmt.Errorf("count available inventory items: %w", err)
			}
			if _, err = tx.ExecContext(ctx, `UPDATE inventory_items SET status = 'RETURNED', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, itemID); err != nil {
				return fmt.Errorf("update inventory item: %w", err)
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at) VALUES ($1, $2, $3, 'SUPPLIER_RETURN', -1, $4, $5, 'supplier_return', $6, 'Supplier return', $7, CURRENT_TIMESTAMP)`, uuid.New(), itemID, item.ProductID, before, before-1, id, userID); err != nil {
				return fmt.Errorf("record inventory movement: %w", err)
			}
		}
		var aggregateTotal int
		if err = tx.GetContext(ctx, &aggregateTotal, `SELECT COALESCE(SUM(quantity), 0) FROM inventory WHERE product_id = $1`, item.ProductID); err != nil {
			return fmt.Errorf("get aggregate product inventory: %w", err)
		}
		var inventoryRows []struct {
			ID       uuid.UUID `db:"id"`
			Quantity int       `db:"quantity"`
		}
		if err = tx.SelectContext(ctx, &inventoryRows, `SELECT id, quantity FROM inventory WHERE product_id = $1 ORDER BY quantity DESC, created_at ASC`, item.ProductID); err != nil {
			return fmt.Errorf("get inventory rows for product: %w", err)
		}
		remaining := item.Quantity
		for _, row := range inventoryRows {
			if remaining <= 0 {
				break
			}
			if row.Quantity <= 0 {
				continue
			}
			deducted := row.Quantity
			if deducted > remaining {
				deducted = remaining
			}
			if _, err = tx.ExecContext(ctx, `UPDATE inventory SET quantity = quantity - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, deducted, row.ID); err != nil {
				return fmt.Errorf("update aggregate inventory: %w", err)
			}
			remaining -= deducted
		}
		if remaining != 0 && aggregateTotal < item.Quantity && len(available) >= item.Quantity {
			remaining = 0
		}
		if remaining != 0 {
			return fmt.Errorf("insufficient aggregate inventory for supplier return")
		}
		total += float64(item.Quantity) * item.UnitCost
	}
	if _, err = tx.ExecContext(ctx, `UPDATE supplier_returns SET status = 'COMPLETED', refund_amount = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, total, id); err != nil {
		return fmt.Errorf("complete supplier return: %w", err)
	}
	var supplierID uuid.UUID
	if err = tx.GetContext(ctx, &supplierID, `SELECT supplier_id FROM supplier_returns WHERE id = $1`, id); err != nil {
		return fmt.Errorf("get supplier: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT $1, $2, 'credit', 'SUPPLIER_RETURN', $3, COALESCE((SELECT SUM(CASE WHEN type = 'debit' OR transaction_type = 'PURCHASE' THEN amount ELSE -amount END) FROM supplier_ledger WHERE supplier_id = $2), 0) - $3, 'Supplier return ' || $4, $5, CURRENT_TIMESTAMP`, uuid.New(), supplierID, total, id.String(), id); err != nil {
		return fmt.Errorf("record supplier ledger entry: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE suppliers SET current_balance = COALESCE(current_balance, 0) - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, total, supplierID); err != nil {
		return fmt.Errorf("update supplier balance after supplier return: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit supplier return: %w", err)
	}
	committed = true
	return nil
}
