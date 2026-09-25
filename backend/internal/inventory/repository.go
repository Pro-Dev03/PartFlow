package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func inventoryQuantityExpressions(ctx context.Context, db *sqlx.DB, productRef string) (string, string) {
	var tableExists int
	if err := db.GetContext(ctx, &tableExists, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'inventory'`); err == nil && tableExists > 0 {
		return fmt.Sprintf(`CASE WHEN EXISTS (SELECT 1 FROM inventory_items ii0 WHERE ii0.product_id = %s) THEN (SELECT COUNT(*) FROM inventory_items ii2 WHERE ii2.product_id = %s) ELSE COALESCE((SELECT quantity FROM inventory WHERE product_id = %s), 0) END`, productRef, productRef, productRef), fmt.Sprintf(`CASE WHEN EXISTS (SELECT 1 FROM inventory_items ii0 WHERE ii0.product_id = %s) THEN (SELECT COUNT(*) FROM inventory_items ii3 WHERE ii3.product_id = %s AND UPPER(TRIM(COALESCE(ii3.status, ''))) = 'AVAILABLE') ELSE MAX(0, COALESCE((SELECT quantity - COALESCE(reserved_quantity, 0) FROM inventory WHERE product_id = %s), 0)) END`, productRef, productRef, productRef)
	}
	return fmt.Sprintf(`(SELECT COUNT(*) FROM inventory_items ii2 WHERE ii2.product_id = %s)`, productRef), fmt.Sprintf(`(SELECT COUNT(*) FROM inventory_items ii3 WHERE ii3.product_id = %s AND UPPER(TRIM(COALESCE(ii3.status, ''))) = 'AVAILABLE')`, productRef)
}

func sqliteHasColumn(db *sqlx.DB, tableName string, columnName string) bool {
	if !dbutil.IsSQLite(db) {
		return true
	}
	var count int
	err := db.Get(&count, `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, tableName, columnName)
	return err == nil && count > 0
}

func inventoryItemFromMap(record map[string]any) (*InventoryItem, error) {
	item := &InventoryItem{}
	if raw, ok := record["product_name"]; ok && raw != nil && raw != "" {
		value := fmt.Sprint(raw)
		item.ProductName = &value
	}
	if raw, ok := record["id"]; ok && raw != nil {
		parsed, err := parseUUIDValue(raw)
		if err != nil {
			return nil, fmt.Errorf("parse item id: %w", err)
		}
		item.ID = parsed
	}

	if raw, ok := record["product_id"]; ok && raw != nil {
		parsed, err := parseNullableUUIDValue(raw)
		if err != nil {
			return nil, fmt.Errorf("parse product_id: %w", err)
		}
		item.ProductID = parsed
	}

	if raw, ok := record["category_id"]; ok && raw != nil {
		parsed, err := parseNullableUUIDValue(raw)
		if err != nil {
			return nil, fmt.Errorf("parse category_id: %w", err)
		}
		item.CategoryID = parsed
	}

	if raw, ok := record["part_type_id"]; ok && raw != nil {
		parsed, err := parseNullableUUIDValue(raw)
		if err != nil {
			return nil, fmt.Errorf("parse part_type_id: %w", err)
		}
		item.PartTypeID = parsed
	}

	if raw, ok := record["item_code"]; ok && raw != nil {
		if value, ok := stringValue(raw); ok && strings.TrimSpace(value) != "" {
			item.ItemCode = &value
		}
	}
	if raw, ok := record["barcode"]; ok && raw != nil {
		if value, ok := stringValue(raw); ok && strings.TrimSpace(value) != "" {
			item.Barcode = &value
		}
	}
	if raw, ok := record["serial_number"]; ok && raw != nil {
		if value, ok := stringValue(raw); ok && strings.TrimSpace(value) != "" {
			item.SerialNumber = &value
		}
	}
	if raw, ok := record["condition"]; ok && raw != nil {
		item.Condition = fmt.Sprint(raw)
	}
	if raw, ok := record["grade"]; ok && raw != nil {
		if value, ok := stringValue(raw); ok && strings.TrimSpace(value) != "" {
			item.Grade = &value
		}
	}
	if raw, ok := record["purchase_cost"]; ok && raw != nil {
		item.PurchaseCost = toFloat64(raw)
	}
	if raw, ok := record["selling_price"]; ok && raw != nil {
		item.SellingPrice = toFloat64(raw)
	}
	if raw, ok := record["status"]; ok && raw != nil {
		item.Status = fmt.Sprint(raw)
	}
	if raw, ok := record["location_id"]; ok && raw != nil {
		parsed, err := parseNullableUUIDValue(raw)
		if err != nil {
			return nil, fmt.Errorf("parse location_id: %w", err)
		}
		item.LocationID = parsed
	}
	if raw, ok := record["supplier_id"]; ok && raw != nil {
		parsed, err := parseNullableUUIDValue(raw)
		if err != nil {
			return nil, fmt.Errorf("parse supplier_id: %w", err)
		}
		item.SupplierID = parsed
	}
	if raw, ok := record["purchase_date"]; ok && raw != nil {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return nil, fmt.Errorf("parse purchase_date: %w", err)
		}
		if !parsed.IsZero() {
			item.PurchaseDate = &parsed
		}
	}
	if raw, ok := record["sold_at"]; ok && raw != nil {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return nil, fmt.Errorf("parse sold_at: %w", err)
		}
		if !parsed.IsZero() {
			item.SoldAt = &parsed
		}
	}
	if raw, ok := record["notes"]; ok && raw != nil {
		if value, ok := stringValue(raw); ok && strings.TrimSpace(value) != "" {
			item.Notes = &value
		}
	}
	if raw, ok := record["created_at"]; ok && raw != nil {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		item.CreatedAt = parsed
	}
	if raw, ok := record["updated_at"]; ok && raw != nil {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at: %w", err)
		}
		item.UpdatedAt = parsed
	}
	if raw, ok := record["current_quantity"]; ok && raw != nil {
		item.CurrentQuantity = int(toFloat64(raw))
	}
	if raw, ok := record["available_quantity"]; ok && raw != nil {
		item.AvailableQuantity = int(toFloat64(raw))
	}
	return item, nil
}

func stringValue(raw any) (string, bool) {
	switch v := raw.(type) {
	case nil:
		return "", false
	case string:
		return v, true
	case []byte:
		return string(v), true
	case fmt.Stringer:
		return v.String(), true
	default:
		return strings.TrimSpace(fmt.Sprint(v)), true
	}
}

func parseUUIDValue(raw any) (uuid.UUID, error) {
	if raw == nil {
		return uuid.Nil, nil
	}
	switch v := raw.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		if strings.TrimSpace(v) == "" {
			return uuid.Nil, nil
		}
		return uuid.Parse(v)
	case []byte:
		if len(v) == 0 {
			return uuid.Nil, nil
		}
		return uuid.Parse(string(v))
	default:
		text := strings.TrimSpace(fmt.Sprint(v))
		if text == "" || text == "<nil>" {
			return uuid.Nil, nil
		}
		return uuid.Parse(text)
	}
}

func parseNullableUUIDValue(raw any) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, nil
		}
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	case []byte:
		if len(v) == 0 {
			return nil, nil
		}
		parsed, err := uuid.Parse(string(v))
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	case uuid.UUID:
		return &v, nil
	default:
		text := strings.TrimSpace(fmt.Sprint(v))
		if text == "" || text == "<nil>" {
			return nil, nil
		}
		parsed, err := uuid.Parse(text)
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	}
}

func toFloat64(raw any) float64 {
	switch v := raw.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if strings.TrimSpace(v) == "" {
			return 0
		}
		var out float64
		_, _ = fmt.Sscan(v, &out)
		return out
	default:
		return 0
	}
}

// CreateInventoryItem creates a new inventory item
func (r *Repository) CreateInventoryItem(ctx context.Context, item *InventoryItem) error {
	query := `
		INSERT INTO inventory_items (id, product_id, part_type_id, item_code, barcode, serial_number,
			condition, grade, purchase_cost, selling_price, status, location_id,
			supplier_id, purchase_date, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProductID, item.PartTypeID, item.ItemCode, item.Barcode, item.SerialNumber,
		item.Condition, item.Grade, item.PurchaseCost, item.SellingPrice,
		item.Status, item.LocationID, item.SupplierID, item.PurchaseDate, item.Notes,
		item.CreatedAt, item.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create inventory item: %w", err)
	}

	return nil
}

func (r *Repository) HasProtectedHistory(ctx context.Context, itemID uuid.UUID) (bool, error) {
	var count int
	if err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM inventory_movements WHERE item_id = $1`, itemID); err != nil {
		return false, fmt.Errorf("failed to check inventory item history: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) DeleteInventoryItem(ctx context.Context, itemID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM inventory_items WHERE id = $1`, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *Repository) deleteInventoryItemRelatedRows(ctx context.Context, tx *sqlx.Tx, tableName, columnName string, itemID uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = ?`, tableName, columnName)
	if _, err := tx.ExecContext(ctx, query, itemID); err != nil {
		errText := strings.ToLower(err.Error())
		if dbutil.IsSQLite(r.db) && (strings.Contains(errText, "no such table:") || strings.Contains(errText, "no such column:") || strings.Contains(errText, "duplicate column") || strings.Contains(errText, "undefined column")) {
			return nil
		}
		return fmt.Errorf("failed to delete related rows from %s: %w", tableName, err)
	}
	return nil
}

func (r *Repository) DeleteUsedInventoryItem(ctx context.Context, itemID, userID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start permanent inventory deletion: %w", err)
	}
	defer tx.Rollback()

	var row struct {
		ProductID string `db:"product_id"`
		Status    string `db:"status"`
	}
	if err := tx.GetContext(ctx, &row, `SELECT product_id, status FROM inventory_items WHERE id = $1`, itemID); err != nil {
		if err == sql.ErrNoRows {
			return ErrItemNotFound
		}
		return fmt.Errorf("failed to load item before permanent delete: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM inventory_movements WHERE item_id = $1`, itemID); err != nil {
		return fmt.Errorf("failed to remove item history: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM reservations WHERE item_id = $1`, itemID); err != nil {
		return fmt.Errorf("failed to remove item reservations: %w", err)
	}

	for _, table := range []struct {
		tableName string
		column    string
	}{
		{tableName: "item_history", column: "inventory_item_id"},
		{tableName: "item_repair_costs", column: "inventory_item_id"},
		{tableName: "acquisition_items", column: "inventory_item_id"},
		{tableName: "trade_ins", column: "inventory_item_id"},
		{tableName: "barcodes", column: "inventory_item_id"},
		{tableName: "inspections", column: "inventory_item_id"},
		{tableName: "sale_items", column: "inventory_item_id"},
		{tableName: "return_items", column: "inventory_item_id"},
		{tableName: "supplier_return_items", column: "inventory_item_id"},
		{tableName: "item_specification_values", column: "inventory_item_id"},
	} {
		if err := r.deleteInventoryItemRelatedRows(ctx, tx, table.tableName, table.column, itemID); err != nil {
			return err
		}
	}

	if strings.EqualFold(row.Status, "AVAILABLE") || strings.EqualFold(row.Status, "RETURNED") {
		if _, err := tx.ExecContext(ctx, `UPDATE inventory SET quantity = CASE WHEN COALESCE(quantity, 0) > 0 THEN quantity - 1 ELSE 0 END, updated_at = CURRENT_TIMESTAMP WHERE product_id = $1`, row.ProductID); err != nil {
			return fmt.Errorf("failed to update inventory quantity after delete: %w", err)
		}
	}
	if strings.EqualFold(row.Status, "RESERVED") {
		if _, err := tx.ExecContext(ctx, `UPDATE inventory SET reserved_quantity = CASE WHEN COALESCE(reserved_quantity, 0) > 0 THEN reserved_quantity - 1 ELSE 0 END, updated_at = CURRENT_TIMESTAMP WHERE product_id = $1`, row.ProductID); err != nil {
			return fmt.Errorf("failed to update reserved quantity after delete: %w", err)
		}
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM inventory_items WHERE id = $1`, itemID)
	if err != nil {
		return fmt.Errorf("failed to permanently delete inventory item: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return ErrItemNotFound
	}
	var actor interface{}
	if userID != uuid.Nil {
		actor = userID
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES ($1,$2,'DELETE','inventory_item',$3,$4,CURRENT_TIMESTAMP)`, uuid.New(), actor, itemID, fmt.Sprintf(`{"product_id":%q,"status":%q}`, row.ProductID, row.Status)); err != nil {
		return fmt.Errorf("write inventory deletion audit: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit permanent inventory deletion: %w", err)
	}
	return nil
}

// GetInventoryItemByID retrieves an inventory item by ID
func (r *Repository) GetInventoryItemByID(ctx context.Context, id uuid.UUID) (*InventoryItem, error) {
	query := `
		SELECT id, product_id, category_id, part_type_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii2
			   WHERE ii2.product_id = inventory_items.product_id
			   ), 0) AS current_quantity,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii3
			   WHERE ii3.product_id = inventory_items.product_id
			   AND ii3.status = 'AVAILABLE'
			   ), 0) AS available_quantity
		FROM inventory_items
		WHERE id = $1
	`

	row := r.db.QueryRowxContext(ctx, query, id)
	if row == nil {
		return nil, ErrItemNotFound
	}

	record := map[string]any{}
	if err := row.MapScan(record); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}
	item, err := inventoryItemFromMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inventory item: %w", err)
	}
	return item, nil
}

// GetInventoryItemByBarcode retrieves an inventory item by barcode
func (r *Repository) GetInventoryItemByBarcode(ctx context.Context, barcode string) (*InventoryItem, error) {
	query := `
		SELECT id, product_id, category_id, part_type_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii2
			   WHERE ii2.product_id = inventory_items.product_id
			   ), 0) AS current_quantity,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii3
			   WHERE ii3.product_id = inventory_items.product_id
			   AND ii3.status = 'AVAILABLE'
			   ), 0) AS available_quantity
		FROM inventory_items
		WHERE barcode = $1
	`

	row := r.db.QueryRowxContext(ctx, query, barcode)
	if row == nil {
		return nil, ErrItemNotFound
	}
	record := map[string]any{}
	if err := row.MapScan(record); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get inventory item by barcode: %w", err)
	}
	item, err := inventoryItemFromMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inventory item: %w", err)
	}
	return item, nil
}

// GetInventoryItemBySerialNumber retrieves an inventory item by serial number
func (r *Repository) GetInventoryItemBySerialNumber(ctx context.Context, serialNumber string) (*InventoryItem, error) {
	query := `
		SELECT id, product_id, category_id, part_type_id, item_code, barcode, serial_number,
			   condition, grade, purchase_cost, selling_price, status, location_id,
			   supplier_id, purchase_date, sold_at, notes, created_at, updated_at,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii2
			   WHERE ii2.product_id = inventory_items.product_id
			   ), 0) AS current_quantity,
			   COALESCE((
			   SELECT COUNT(*)
			   FROM inventory_items ii3
			   WHERE ii3.product_id = inventory_items.product_id
			   AND ii3.status = 'AVAILABLE'
			   ), 0) AS available_quantity
		FROM inventory_items
		WHERE serial_number = $1
	`

	row := r.db.QueryRowxContext(ctx, query, serialNumber)
	if row == nil {
		return nil, ErrItemNotFound
	}
	record := map[string]any{}
	if err := row.MapScan(record); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get inventory item by serial number: %w", err)
	}
	item, err := inventoryItemFromMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inventory item: %w", err)
	}
	return item, nil
}

// UpdateInventoryItem updates an inventory item
func (r *Repository) UpdateInventoryItem(ctx context.Context, item *InventoryItem) error {
	var categoryID interface{}
	if item.CategoryID != nil {
		categoryID = item.CategoryID.String()
	}
	query := `
		UPDATE inventory_items
		SET product_id = $2, category_id = $3, part_type_id = $4, item_code = $5, barcode = $6, serial_number = $7,
		    condition = $8, grade = $9, purchase_cost = $10, selling_price = $11,
		    status = $12, location_id = $13, supplier_id = $14, purchase_date = $15,
		    sold_at = $16, notes = $17, updated_at = $18
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProductID, categoryID, item.PartTypeID, item.ItemCode, item.Barcode, item.SerialNumber,
		item.Condition, item.Grade, item.PurchaseCost, item.SellingPrice,
		item.Status, item.LocationID, item.SupplierID, item.PurchaseDate, item.SoldAt,
		item.Notes, item.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update inventory item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrItemNotFound
	}

	return nil
}

// UpdateItemStatus updates the status of an inventory item
func (r *Repository) UpdateItemStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := fmt.Sprintf(`
		UPDATE inventory_items
		SET status = $2, updated_at = %s
		WHERE id = $1
	`, dbutil.NowSQL(r.db))

	result, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrItemNotFound
	}

	return nil
}

// ListInventoryItems retrieves a list of inventory items with pagination
func (r *Repository) ListInventoryItems(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*InventoryItem, int64, error) {
	currentQuantityExpr, availableQuantityExpr := inventoryQuantityExpressions(ctx, r.db, "inventory_items.product_id")
	// Build base query
	baseQuery := fmt.Sprintf(`
		SELECT inventory_items.id, inventory_items.product_id, p.name AS product_name,
		       inventory_items.part_type_id, inventory_items.item_code, inventory_items.barcode, inventory_items.serial_number,
		       inventory_items.condition, inventory_items.grade, inventory_items.purchase_cost, inventory_items.selling_price,
		       inventory_items.status, inventory_items.location_id, inventory_items.supplier_id, inventory_items.purchase_date,
		       inventory_items.sold_at, inventory_items.notes, inventory_items.created_at, inventory_items.updated_at,
			   %s AS current_quantity,
			   %s AS available_quantity
		FROM inventory_items
		LEFT JOIN products p ON inventory_items.product_id = p.id
		WHERE UPPER(COALESCE(inventory_items.status, '')) <> 'ARCHIVED'
		AND p.deleted_at IS NULL
	`, currentQuantityExpr, availableQuantityExpr)
	countQuery := `SELECT COUNT(*) FROM inventory_items WHERE UPPER(COALESCE(status, '')) <> 'ARCHIVED' AND EXISTS (SELECT 1 FROM products p WHERE p.id = inventory_items.product_id AND p.deleted_at IS NULL)`

	args := []interface{}{}
	argCount := 0

	// Add filters if provided
	if condition, ok := filters["condition"].(string); ok && condition != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND inventory_items.condition = ` + param
		countQuery += ` AND inventory_items.condition = ` + param
		args = append(args, condition)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND inventory_items.status = ` + param
		countQuery += ` AND inventory_items.status = ` + param
		args = append(args, status)
	}

	if productID, ok := filters["product_id"].(uuid.UUID); ok && productID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND inventory_items.product_id = ` + param
		countQuery += ` AND inventory_items.product_id = ` + param
		args = append(args, productID)
	}

	if partTypeID, ok := filters["part_type_id"].(uuid.UUID); ok && partTypeID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND inventory_items.part_type_id = ` + param
		countQuery += ` AND inventory_items.part_type_id = ` + param
		args = append(args, partTypeID)
	}

	if locationID, ok := filters["location_id"].(uuid.UUID); ok && locationID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND inventory_items.location_id = ` + param
		countQuery += ` AND inventory_items.location_id = ` + param
		args = append(args, locationID)
	}

	// Get total count
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory items: %w", err)
	}

	// Add pagination
	argCount++
	param := fmt.Sprintf("$%d", argCount)
	baseQuery += ` ORDER BY inventory_items.created_at DESC LIMIT ` + param
	args = append(args, limit)

	argCount++
	param = fmt.Sprintf("$%d", argCount)
	baseQuery += ` OFFSET ` + param
	args = append(args, offset)

	// Execute query
	rows, err := r.db.QueryxContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory items: %w", err)
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, 0, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		item, err := inventoryItemFromMap(record)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse inventory item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate inventory items: %w", err)
	}

	return items, total, nil
}

// ListInventoryItemsWithSupplierInfo retrieves inventory items with supplier information
func (r *Repository) ListInventoryItemsWithSupplierInfo(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*InventoryItemWithSupplier, int64, error) {
	currentQuantityExpr, availableQuantityExpr := inventoryQuantityExpressions(ctx, r.db, "ii.product_id")
	categoryExpr := "p.category_id"
	if !sqliteHasColumn(r.db, "products", "category_id") {
		categoryExpr = "ii.category_id"
	}
	categoryValueExpr := "p.category_id"
	if dbutil.IsSQLite(r.db) {
		if sqliteHasColumn(r.db, "products", "category_id") && sqliteHasColumn(r.db, "inventory_items", "category_id") {
			categoryValueExpr = "COALESCE(ii.category_id, p.category_id)"
		} else if sqliteHasColumn(r.db, "inventory_items", "category_id") {
			categoryValueExpr = "ii.category_id"
		}
	}
	// Backfill legacy inventory rows whenever the list is read. This also covers
	// products imported after SQLite startup, when the startup migration ran too early.
	if categoryExpr == "p.category_id" {
		_, _ = r.db.ExecContext(ctx, `
		UPDATE inventory_items
		SET category_id = (SELECT p.category_id FROM products p WHERE p.id = inventory_items.product_id)
		WHERE category_id IS NULL
		  AND product_id IS NOT NULL
		  AND EXISTS (SELECT 1 FROM products p WHERE p.id = inventory_items.product_id AND p.category_id IS NOT NULL)
		`)
	}

	manualOnly, _ := filters["manual_only"].(bool)
	supplierOnly, _ := filters["supplier_only"].(bool)
	filterPrefix := "ii"
	baseQuery := fmt.Sprintf(`
		SELECT
			CAST(ii.id AS TEXT) AS id, ii.product_id, ii.part_type_id, ii.item_code, ii.barcode, ii.serial_number,
			ii.condition, ii.grade, ii.purchase_cost, ii.selling_price, ii.status, ii.location_id,
			ii.supplier_id, ii.purchase_date, ii.sold_at, ii.notes, ii.created_at, ii.updated_at,
			%s AS current_quantity,
			%s AS available_quantity,
			p.name as product_name,
			p.selling_price as product_selling_price,
			CAST(%s AS TEXT) as category_id,
			(SELECT c2.name FROM categories c2 WHERE CAST(c2.id AS TEXT) = CAST(%s AS TEXT) LIMIT 1) as category_name,
			s.name as supplier_name,
			s.phone as supplier_phone
		FROM inventory_items ii
		LEFT JOIN products p ON ii.product_id = p.id
			LEFT JOIN categories c ON c.id = %s
		LEFT JOIN suppliers s ON ii.supplier_id = s.id
	`, currentQuantityExpr, availableQuantityExpr, categoryValueExpr, categoryValueExpr, categoryValueExpr)
	countQuery := `
		SELECT COUNT(*) FROM inventory_items ii
	`
	if includeArchived, ok := filters["include_archived"].(bool); !ok || !includeArchived {
		baseQuery += ` WHERE UPPER(COALESCE(ii.status, '')) <> 'ARCHIVED' AND p.deleted_at IS NULL`
		countQuery += ` WHERE UPPER(COALESCE(ii.status, '')) <> 'ARCHIVED' AND EXISTS (SELECT 1 FROM products p WHERE p.id = ii.product_id AND p.deleted_at IS NULL)`
	} else {
		baseQuery += ` WHERE 1=1 AND p.deleted_at IS NULL`
		countQuery += ` WHERE 1=1 AND EXISTS (SELECT 1 FROM products p WHERE p.id = ii.product_id AND p.deleted_at IS NULL)`
	}

	if manualOnly {
		filterPrefix = "manual_inventory"
		unionQuery := fmt.Sprintf(`
			SELECT
				CAST(ii.id AS TEXT) AS id, ii.product_id, ii.part_type_id, ii.item_code, ii.barcode, ii.serial_number,
				ii.condition, ii.grade, ii.purchase_cost, ii.selling_price, ii.status, ii.location_id,
				ii.supplier_id, ii.purchase_date, ii.sold_at, ii.notes, ii.created_at, ii.updated_at,
				%s AS current_quantity,
				%s AS available_quantity,
				p.name as product_name,
				p.selling_price as product_selling_price,
				CAST(%s AS TEXT) as category_id,
				(SELECT c2.name FROM categories c2 WHERE CAST(c2.id AS TEXT) = CAST(%s AS TEXT) LIMIT 1) as category_name,
				s.name as supplier_name,
				s.phone as supplier_phone
			FROM inventory_items ii
			LEFT JOIN products p ON ii.product_id = p.id
			LEFT JOIN categories c ON c.id = %s
			LEFT JOIN suppliers s ON ii.supplier_id = s.id
			WHERE UPPER(COALESCE(ii.status, '')) <> 'ARCHIVED'
			AND p.deleted_at IS NULL
			AND ii.supplier_id IS NULL
			UNION ALL
			SELECT
				CAST(inv.product_id AS TEXT) AS id,
				inv.product_id,
				NULL AS part_type_id,
				NULL AS item_code,
				NULL AS barcode,
				NULL AS serial_number,
				'NEW' AS condition,
				NULL AS grade,
				COALESCE(p.cost_price, 0) AS purchase_cost,
				p.selling_price AS selling_price,
				'AVAILABLE' AS status,
				NULL AS location_id,
				NULL AS supplier_id,
				(
					SELECT MAX(im.business_date)
					FROM inventory_movements im
					WHERE im.product_id = inv.product_id
					  AND im.item_id IS NULL
					  AND UPPER(COALESCE(im.source_type, '')) = 'OPENING_STOCK'
				) AS purchase_date,
				NULL AS sold_at,
				NULL AS notes,
				inv.updated_at AS created_at,
				inv.updated_at AS updated_at,
				inv.quantity AS current_quantity,
				inv.quantity AS available_quantity,
				p.name AS product_name,
				p.selling_price AS product_selling_price,
				CAST(p.category_id AS TEXT) AS category_id,
				(SELECT c2.name FROM categories c2 WHERE CAST(c2.id AS TEXT) = CAST(p.category_id AS TEXT) LIMIT 1) AS category_name,
				NULL AS supplier_name,
				NULL AS supplier_phone
			FROM inventory inv
			LEFT JOIN products p ON p.id = inv.product_id
			WHERE inv.quantity > 0
			AND p.deleted_at IS NULL
			AND EXISTS (
				SELECT 1 FROM inventory_movements im
				WHERE im.product_id = inv.product_id
				AND im.item_id IS NULL
				AND UPPER(COALESCE(im.source_type, '')) = 'OPENING_STOCK'
			)
		`, currentQuantityExpr, availableQuantityExpr, categoryValueExpr, categoryValueExpr, categoryValueExpr)
		baseQuery = `SELECT * FROM (` + unionQuery + `) AS manual_inventory WHERE 1=1`
		countQuery = `SELECT COUNT(*) FROM (` + unionQuery + `) AS manual_inventory WHERE 1=1`
	} else if !supplierOnly {
		filterPrefix = "combined_inventory"
		unionQuery := fmt.Sprintf(`
			SELECT
				CAST(ii.id AS TEXT) AS id, ii.product_id, ii.part_type_id, ii.item_code, ii.barcode, ii.serial_number,
				ii.condition, ii.grade, ii.purchase_cost, ii.selling_price, ii.status, ii.location_id,
				ii.supplier_id, ii.purchase_date, ii.sold_at, ii.notes, ii.created_at, ii.updated_at,
				%s AS current_quantity,
				%s AS available_quantity,
				p.name as product_name,
				p.selling_price as product_selling_price,
				CAST(%s AS TEXT) as category_id,
				(SELECT c2.name FROM categories c2 WHERE CAST(c2.id AS TEXT) = CAST(%s AS TEXT) LIMIT 1) as category_name,
				s.name as supplier_name,
				s.phone as supplier_phone
			FROM inventory_items ii
			LEFT JOIN products p ON ii.product_id = p.id
			LEFT JOIN categories c ON c.id = %s
			LEFT JOIN suppliers s ON ii.supplier_id = s.id
			WHERE UPPER(COALESCE(ii.status, '')) <> 'ARCHIVED'
			AND p.deleted_at IS NULL
			UNION ALL
			SELECT
				CAST(inv.product_id AS TEXT) AS id,
				inv.product_id,
				NULL AS part_type_id,
				NULL AS item_code,
				NULL AS barcode,
				NULL AS serial_number,
				'NEW' AS condition,
				NULL AS grade,
				COALESCE(p.cost_price, 0) AS purchase_cost,
				p.selling_price AS selling_price,
				'AVAILABLE' AS status,
				NULL AS location_id,
				NULL AS supplier_id,
				(
					SELECT MAX(im.business_date)
					FROM inventory_movements im
					WHERE im.product_id = inv.product_id
					  AND im.item_id IS NULL
					  AND UPPER(COALESCE(im.source_type, '')) = 'OPENING_STOCK'
				) AS purchase_date,
				NULL AS sold_at,
				NULL AS notes,
				inv.updated_at AS created_at,
				inv.updated_at AS updated_at,
				inv.quantity AS current_quantity,
				inv.quantity AS available_quantity,
				p.name AS product_name,
				p.selling_price AS product_selling_price,
				CAST(p.category_id AS TEXT) AS category_id,
				(SELECT c2.name FROM categories c2 WHERE CAST(c2.id AS TEXT) = CAST(p.category_id AS TEXT) LIMIT 1) AS category_name,
				NULL AS supplier_name,
				NULL AS supplier_phone
			FROM inventory inv
			LEFT JOIN products p ON p.id = inv.product_id
			WHERE inv.quantity > 0
			AND p.deleted_at IS NULL
			AND EXISTS (
				SELECT 1 FROM inventory_movements im
				WHERE im.product_id = inv.product_id
				AND im.item_id IS NULL
				AND UPPER(COALESCE(im.source_type, '')) = 'OPENING_STOCK'
			)
		`, currentQuantityExpr, availableQuantityExpr, categoryValueExpr, categoryValueExpr, categoryValueExpr)
		baseQuery = `SELECT * FROM (` + unionQuery + `) AS combined_inventory WHERE 1=1`
		countQuery = `SELECT COUNT(*) FROM (` + unionQuery + `) AS combined_inventory WHERE 1=1`
	}

	if supplierOnly {
		filterPrefix = "ii"
	}

	args := []interface{}{}
	argCount := 0

	// Add filters if provided
	if condition, ok := filters["condition"].(string); ok && condition != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ` + filterPrefix + `.condition = ` + param
		countQuery += ` AND ` + filterPrefix + `.condition = ` + param
		args = append(args, condition)
	}
	if excludeCondition, ok := filters["exclude_condition"].(string); ok && excludeCondition != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ` + filterPrefix + `.condition <> ` + param
		countQuery += ` AND ` + filterPrefix + `.condition <> ` + param
		args = append(args, excludeCondition)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ` + filterPrefix + `.status = ` + param
		countQuery += ` AND ` + filterPrefix + `.status = ` + param
		args = append(args, status)
	}

	if productID, ok := filters["product_id"].(uuid.UUID); ok && productID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ` + filterPrefix + `.product_id = ` + param
		countQuery += ` AND ` + filterPrefix + `.product_id = ` + param
		args = append(args, productID)
	}
	if categoryID, ok := filters["category_id"].(uuid.UUID); ok && categoryID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		targetColumn := filterPrefix + ".category_id"
		if filterPrefix == "ii" && !manualOnly && !supplierOnly {
			targetColumn = categoryExpr
		}
		baseQuery += ` AND ` + targetColumn + ` = ` + param
		countQuery += ` AND ` + targetColumn + ` = ` + param
		args = append(args, categoryID)
	}

	if partTypeID, ok := filters["part_type_id"].(uuid.UUID); ok && partTypeID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ` + filterPrefix + `.part_type_id = ` + param
		countQuery += ` AND ` + filterPrefix + `.part_type_id = ` + param
		args = append(args, partTypeID)
	}

	if locationID, ok := filters["location_id"].(uuid.UUID); ok && locationID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ` + filterPrefix + `.location_id = ` + param
		countQuery += ` AND ` + filterPrefix + `.location_id = ` + param
		args = append(args, locationID)
	}

	// New filters for supplier and purchase
	if supplierID, ok := filters["supplier_id"].(uuid.UUID); ok && supplierID != uuid.Nil {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		baseQuery += ` AND ` + filterPrefix + `.supplier_id = ` + param
		countQuery += ` AND ` + filterPrefix + `.supplier_id = ` + param
		args = append(args, supplierID)
	}
	if supplierOnly, ok := filters["supplier_only"].(bool); ok && supplierOnly {
		baseQuery += ` AND ` + filterPrefix + `.supplier_id IS NOT NULL`
		countQuery += ` AND ` + filterPrefix + `.supplier_id IS NOT NULL`
	}
	if manualOnly, ok := filters["manual_only"].(bool); ok && manualOnly {
		baseQuery += ` AND ` + filterPrefix + `.supplier_id IS NULL`
		countQuery += ` AND ` + filterPrefix + `.supplier_id IS NULL`
	}

	if purchaseDateFrom, ok := filters["purchase_date_from"].(string); ok && purchaseDateFrom != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		purchaseColumn := filterPrefix + ".purchase_date"
		if filterPrefix == "ii" {
			purchaseColumn = "ii.purchase_date"
		}
		baseQuery += ` AND ` + purchaseColumn + ` >= ` + param
		countQuery += ` AND ` + purchaseColumn + ` >= ` + param
		args = append(args, purchaseDateFrom)
	}

	if purchaseDateTo, ok := filters["purchase_date_to"].(string); ok && purchaseDateTo != "" {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		purchaseColumn := filterPrefix + ".purchase_date"
		if filterPrefix == "ii" {
			purchaseColumn = "ii.purchase_date"
		}
		baseQuery += ` AND ` + purchaseColumn + ` <= ` + param
		countQuery += ` AND ` + purchaseColumn + ` <= ` + param
		args = append(args, purchaseDateTo)
	}

	if minPurchaseCost, ok := filters["min_purchase_cost"].(float64); ok && minPurchaseCost > 0 {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		purchaseColumn := filterPrefix + ".purchase_cost"
		if filterPrefix == "ii" {
			purchaseColumn = "ii.purchase_cost"
		}
		baseQuery += ` AND ` + purchaseColumn + ` >= ` + param
		countQuery += ` AND ` + purchaseColumn + ` >= ` + param
		args = append(args, minPurchaseCost)
	}

	if maxPurchaseCost, ok := filters["max_purchase_cost"].(float64); ok && maxPurchaseCost > 0 {
		argCount++
		param := fmt.Sprintf("$%d", argCount)
		purchaseColumn := filterPrefix + ".purchase_cost"
		if filterPrefix == "ii" {
			purchaseColumn = "ii.purchase_cost"
		}
		baseQuery += ` AND ` + purchaseColumn + ` <= ` + param
		countQuery += ` AND ` + purchaseColumn + ` <= ` + param
		args = append(args, maxPurchaseCost)
	}

	// Get total count
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory items: %w", err)
	}

	// Add pagination
	argCount++
	param := fmt.Sprintf("$%d", argCount)
	orderColumn := filterPrefix + ".created_at"
	if filterPrefix == "ii" {
		orderColumn = "ii.created_at"
	}
	baseQuery += ` ORDER BY ` + orderColumn + ` DESC LIMIT ` + param
	args = append(args, limit)

	argCount++
	param = fmt.Sprintf("$%d", argCount)
	baseQuery += ` OFFSET ` + param
	args = append(args, offset)

	// Execute query
	rows, err := r.db.QueryxContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory items with supplier info: %w", err)
	}
	defer rows.Close()

	var items []*InventoryItemWithSupplier
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, 0, fmt.Errorf("failed to scan inventory item with supplier info: %w", err)
		}
		item, err := inventoryItemWithSupplierFromMap(record)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse inventory item with supplier info: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate inventory items with supplier info: %w", err)
	}

	return items, total, nil
}

// CreateLocation creates a new location
func inventoryItemWithSupplierFromMap(record map[string]any) (*InventoryItemWithSupplier, error) {
	item, err := inventoryItemFromMap(record)
	if err != nil {
		return nil, err
	}
	out := &InventoryItemWithSupplier{
		ID:                item.ID,
		ProductID:         item.ProductID,
		CategoryID:        item.CategoryID,
		PartTypeID:        item.PartTypeID,
		ItemCode:          item.ItemCode,
		Barcode:           item.Barcode,
		SerialNumber:      item.SerialNumber,
		Condition:         item.Condition,
		Grade:             item.Grade,
		PurchaseCost:      item.PurchaseCost,
		SellingPrice:      item.SellingPrice,
		Status:            item.Status,
		LocationID:        item.LocationID,
		SupplierID:        item.SupplierID,
		PurchaseDate:      item.PurchaseDate,
		SoldAt:            item.SoldAt,
		Notes:             item.Notes,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		CurrentQuantity:   item.CurrentQuantity,
		AvailableQuantity: item.AvailableQuantity,
	}
	if raw, ok := record["product_name"]; ok && raw != nil && raw != "" {
		value := fmt.Sprint(raw)
		out.ProductName = &value
	}
	if raw, ok := record["product_selling_price"]; ok && raw != nil {
		value, err := strconv.ParseFloat(fmt.Sprint(raw), 64)
		if err == nil {
			out.ProductSellingPrice = &value
		}
	}
	if raw, ok := record["category_name"]; ok && raw != nil && raw != "" {
		value := fmt.Sprint(raw)
		out.CategoryName = &value
	}
	if raw, ok := record["supplier_name"]; ok && raw != nil && raw != "" {
		value := fmt.Sprint(raw)
		out.SupplierName = &value
	}
	if raw, ok := record["supplier_phone"]; ok && raw != nil && raw != "" {
		value := fmt.Sprint(raw)
		out.SupplierPhone = &value
	}
	return out, nil
}

func (r *Repository) CreateLocation(ctx context.Context, location *Location) error {
	query := `
		INSERT INTO locations (id, name, type, parent_id, warehouse_id, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		location.ID, location.Name, location.Type, location.ParentID, location.WarehouseID, location.Description,
		location.IsActive, location.CreatedAt, location.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create location: %w", err)
	}

	return nil
}

// GetLocationByID retrieves a location by ID
func (r *Repository) GetLocationByID(ctx context.Context, id uuid.UUID) (*Location, error) {
	query := `
		SELECT id, name, type, parent_id, warehouse_id, description, is_active, created_at, updated_at
		FROM locations
		WHERE id = $1
	`

	record := map[string]any{}
	err := r.db.QueryRowxContext(ctx, query, id).MapScan(record)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLocationNotFound
		}
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	location, err := locationFromMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse location: %w", err)
	}
	return location, nil
}

// ListLocations retrieves all locations
func (r *Repository) ListLocations(ctx context.Context) ([]*Location, error) {
	query := `
		SELECT id, name, type, parent_id, warehouse_id, description, is_active, created_at, updated_at
		FROM locations
		ORDER BY name ASC
	`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}
	defer rows.Close()
	locations := make([]*Location, 0)
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, fmt.Errorf("failed to scan location: %w", err)
		}
		location, err := locationFromMap(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse location: %w", err)
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}
	return locations, nil
}

func locationFromMap(record map[string]any) (*Location, error) {
	location := &Location{ID: parseInventoryUUID(record["id"]), Name: strings.TrimSpace(fmt.Sprint(record["name"])), Type: strings.TrimSpace(fmt.Sprint(record["type"])), Description: parseInventoryNullableString(record["description"])}
	location.ParentID = parseInventoryNullableUUID(record["parent_id"])
	location.WarehouseID = parseInventoryNullableUUID(record["warehouse_id"])
	if raw, ok := record["is_active"]; ok {
		location.IsActive = fmt.Sprint(raw) == "1" || strings.EqualFold(fmt.Sprint(raw), "true")
	}
	if value, err := dbutil.ParseTimestamp(record["created_at"]); err == nil {
		location.CreatedAt = value
	}
	if value, err := dbutil.ParseTimestamp(record["updated_at"]); err == nil {
		location.UpdatedAt = value
	}
	return location, nil
}

// CreateMovement creates a new inventory movement
func (r *Repository) CreateMovement(ctx context.Context, movement *InventoryMovement) error {
	query := `
		INSERT INTO inventory_movements (id, item_id, movement_type, quantity,
		 before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		movement.ID, movement.ItemID, movement.MovementType, movement.Quantity, movement.BeforeQuantity,
		movement.AfterQuantity, movement.ReferenceType, movement.ReferenceID,
		movement.Reason, movement.CreatedBy, movement.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	return nil
}

// GetMovementsByItem retrieves movements for a specific item
func (r *Repository) GetMovementsByItem(ctx context.Context, itemID uuid.UUID, limit, offset int) ([]*InventoryMovement, int64, error) {
	query := `
		SELECT im.id, im.item_id, im.movement_type, im.quantity,
		       im.before_quantity, im.after_quantity, im.reference_type, im.reference_id,
		       c.name AS customer_name, s.invoice_number,
		       im.reason, im.created_by, im.created_at
		FROM inventory_movements im
		LEFT JOIN sales s ON s.id = im.reference_id
		LEFT JOIN customers c ON c.id = s.customer_id
		WHERE im.item_id = $1
		ORDER BY im.created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*)
		FROM inventory_movements
		WHERE item_id = $1
	`

	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, itemID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count movements: %w", err)
	}

	rows, err := r.db.QueryxContext(ctx, query, itemID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get movements: %w", err)
	}
	defer rows.Close()
	movements := make([]*InventoryMovement, 0)
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, 0, fmt.Errorf("failed to scan movement: %w", err)
		}
		movement, err := movementFromMap(record)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse movement: %w", err)
		}
		movements = append(movements, movement)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to get movements: %w", err)
	}
	return movements, total, nil
}

func movementFromMap(record map[string]any) (*InventoryMovement, error) {
	movement := &InventoryMovement{
		ID:             parseInventoryUUID(record["id"]),
		ItemID:         parseInventoryNullableUUID(record["item_id"]),
		ProductID:      parseInventoryNullableUUID(record["product_id"]),
		MovementType:   MovementType(strings.TrimSpace(fmt.Sprint(record["movement_type"]))),
		Quantity:       int(toFloat64(record["quantity"])),
		BeforeQuantity: int(toFloat64(record["before_quantity"])),
		AfterQuantity:  int(toFloat64(record["after_quantity"])),
		ReferenceType:  strings.TrimSpace(fmt.Sprint(record["reference_type"])),
		ReferenceID:    parseInventoryNullableUUID(record["reference_id"]),
		CustomerName:   parseInventoryNullableString(record["customer_name"]),
		InvoiceNumber:  parseInventoryNullableString(record["invoice_number"]),
		Reason:         parseInventoryNullableString(record["reason"]),
		CreatedBy:      parseInventoryUUID(record["created_by"]),
	}
	if value, err := dbutil.ParseTimestamp(record["created_at"]); err == nil {
		movement.CreatedAt = value
	}
	return movement, nil
}

// CreateReservation creates a new reservation
func (r *Repository) CreateReservation(ctx context.Context, reservation *Reservation) error {
	query := `
		INSERT INTO reservations (id, item_id, customer_id, user_id, reserved_at, expires_at, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		reservation.ID, reservation.ItemID, reservation.CustomerID, reservation.UserID, reservation.ReservedAt,
		reservation.ExpiresAt, reservation.Status, reservation.Notes,
		reservation.CreatedAt, reservation.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	return nil
}

// GetActiveReservationByItem retrieves active reservation for an item
func (r *Repository) GetActiveReservationByItem(ctx context.Context, itemID uuid.UUID) (*Reservation, error) {
	query := `
		SELECT id, item_id, customer_id, user_id, reserved_at,
		       expires_at, status, notes, created_at, updated_at
		FROM reservations
		WHERE item_id = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`

	record := map[string]any{}
	err := r.db.QueryRowxContext(ctx, query, itemID).MapScan(record)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active reservation
		}
		return nil, fmt.Errorf("failed to get active reservation: %w", err)
	}

	reservation, err := reservationFromMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse active reservation: %w", err)
	}
	return reservation, nil
}

func reservationFromMap(record map[string]any) (*Reservation, error) {
	reservation := &Reservation{ID: parseInventoryUUID(record["id"]), ItemID: parseInventoryUUID(record["item_id"]), CustomerID: parseInventoryNullableUUID(record["customer_id"]), UserID: parseInventoryUUID(record["user_id"]), Status: strings.TrimSpace(fmt.Sprint(record["status"])), Notes: parseInventoryNullableString(record["notes"])}
	for field, destination := range map[string]*time.Time{"reserved_at": &reservation.ReservedAt, "expires_at": &reservation.ExpiresAt, "created_at": &reservation.CreatedAt, "updated_at": &reservation.UpdatedAt} {
		if value, err := dbutil.ParseTimestamp(record[field]); err == nil {
			*destination = value
		}
	}
	return reservation, nil
}

func parseInventoryUUID(value any) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	if parsed, ok := value.(uuid.UUID); ok {
		return parsed
	}
	if raw, ok := value.([]byte); ok {
		if len(raw) == 16 {
			var parsed uuid.UUID
			copy(parsed[:], raw)
			return parsed
		}
		value = string(raw)
	}
	parsed, _ := uuid.Parse(strings.TrimSpace(fmt.Sprint(value)))
	return parsed
}

func parseInventoryNullableUUID(value any) *uuid.UUID {
	parsed := parseInventoryUUID(value)
	if parsed == uuid.Nil {
		return nil
	}
	return &parsed
}

func parseInventoryNullableString(value any) *string {
	if value == nil {
		return nil
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" || text == "<nil>" {
		return nil
	}
	return &text
}

// UpdateReservationStatus updates reservation status
func (r *Repository) UpdateReservationStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := fmt.Sprintf(`
		UPDATE reservations
		SET status = $2, updated_at = %s
		WHERE id = $1
	`, dbutil.NowSQL(r.db))

	result, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrItemNotFound
	}

	return nil
}
