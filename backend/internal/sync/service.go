package sync

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

type SyncResult struct {
	Processed int      `json:"processed"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

func SeedLocalDatabaseFromOnline(postgresDB *sqlx.DB, sqliteDB *sql.DB) error {
	if postgresDB == nil {
		return fmt.Errorf("online database is nil")
	}
	if sqliteDB == nil {
		return fmt.Errorf("local database is nil")
	}

	queries := []struct {
		key      string
		query    string
		required bool
	}{
		{key: "categories", query: `SELECT * FROM categories ORDER BY created_at ASC`, required: true},
		{key: "brands", query: `SELECT * FROM brands ORDER BY created_at ASC`},
		{key: "products", query: `SELECT * FROM products ORDER BY created_at ASC`, required: true},
		{key: "customers", query: `SELECT * FROM customers ORDER BY created_at ASC`, required: true},
		{key: "suppliers", query: `SELECT * FROM suppliers ORDER BY created_at ASC`, required: true},
		{key: "customer_ledger", query: `SELECT * FROM customer_ledger ORDER BY created_at ASC`},
		{key: "supplier_ledger", query: `SELECT * FROM supplier_ledger ORDER BY created_at ASC`},
		{key: "ledger_entries", query: `SELECT * FROM ledger_entries ORDER BY created_at ASC`},
		{key: "inventory", query: `SELECT * FROM inventory ORDER BY created_at ASC`},
		{key: "locations", query: `SELECT * FROM locations ORDER BY created_at ASC`},
		{key: "inventory_items", query: `SELECT * FROM inventory_items ORDER BY created_at ASC`, required: true},
		{key: "inventory_movements", query: `SELECT * FROM inventory_movements ORDER BY created_at ASC`},
		{key: "reservations", query: `SELECT * FROM reservations ORDER BY created_at ASC`},
		{key: "barcodes", query: `SELECT * FROM barcodes ORDER BY created_at ASC`},
		{key: "sales", query: `SELECT * FROM sales ORDER BY created_at ASC`, required: true},
		{key: "sale_items", query: `SELECT * FROM sale_items ORDER BY created_at ASC`, required: true},
		{key: "purchases", query: `SELECT * FROM purchases ORDER BY created_at ASC`, required: true},
		{key: "purchase_items", query: `SELECT * FROM purchase_items ORDER BY created_at ASC`, required: true},
		{key: "payments", query: `SELECT * FROM payments ORDER BY created_at ASC`, required: true},
		{key: "debts", query: `SELECT * FROM debts ORDER BY created_at ASC`, required: true},
		{key: "expenses", query: `SELECT * FROM expenses ORDER BY created_at ASC`, required: true},
		{key: "expense_categories", query: `SELECT * FROM expense_categories ORDER BY created_at ASC`},
		{key: "inspections", query: `SELECT * FROM inspections ORDER BY created_at ASC`},
		{key: "inspection_items", query: `SELECT * FROM inspection_items ORDER BY created_at ASC`},
		{key: "part_types", query: `SELECT * FROM part_types ORDER BY created_at ASC`},
		{key: "part_specifications", query: `SELECT * FROM part_specifications ORDER BY created_at ASC`},
		{key: "type_specifications", query: `SELECT * FROM type_specifications ORDER BY created_at ASC`},
		{key: "acquisitions", query: `SELECT * FROM acquisitions ORDER BY created_at ASC`},
		{key: "acquisition_items", query: `SELECT * FROM acquisition_items ORDER BY created_at ASC`},
		{key: "trade_ins", query: `SELECT * FROM trade_ins ORDER BY created_at ASC`},
		{key: "item_specification_values", query: `SELECT * FROM item_specification_values ORDER BY created_at ASC`},
		{key: "returns", query: `SELECT * FROM returns ORDER BY created_at ASC`},
		{key: "return_items", query: `SELECT * FROM return_items ORDER BY created_at ASC`},
		{key: "supplier_returns", query: `SELECT * FROM supplier_returns ORDER BY created_at ASC`},
		{key: "supplier_return_items", query: `SELECT * FROM supplier_return_items ORDER BY created_at ASC`},
		{key: "notifications", query: `SELECT * FROM notifications ORDER BY created_at ASC`},
		{key: "notification_preferences", query: `SELECT * FROM notification_preferences ORDER BY created_at ASC`},
		{key: "reports", query: `SELECT * FROM reports ORDER BY created_at ASC`},
		{key: "settings", query: `SELECT * FROM settings ORDER BY created_at ASC`},
		{key: "held_sales", query: `SELECT * FROM held_sales ORDER BY created_at ASC`},
	}

	snapshot := make(map[string]any, len(queries))
	for _, item := range queries {
		rows, err := fetchRowsAsMaps(postgresDB, item.query)
		if err != nil {
			if !item.required && isMissingTableError(err) {
				continue
			}
			return fmt.Errorf("fetch %s from online db: %w", item.key, err)
		}
		snapshot[item.key] = rows
	}

	return localdb.SeedLocalSnapshot(sqliteDB, snapshot)
}

func fetchRowsAsMaps(postgresDB *sqlx.DB, query string) ([]map[string]any, error) {
	rows, err := postgresDB.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		record := make(map[string]any)
		if err := rows.MapScan(record); err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func ProcessPendingOfflineQueue(postgresDB *sqlx.DB, sqliteDB *sql.DB, limit int) (*SyncResult, error) {
	entries, err := localdb.ListPendingSyncOperations(sqliteDB, limit)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{}
	for _, entry := range entries {
		if err := syncOneItem(postgresDB, sqliteDB, entry); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s/%s: %v", entry.EntityType, entry.EntityID, err))
			if markErr := localdb.MarkSyncOperationFailed(sqliteDB, entry.ID, err.Error()); markErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("%s/%s: unable to mark failed: %v", entry.EntityType, entry.EntityID, markErr))
			}
			continue
		}
		result.Processed++
		if err := localdb.MarkSyncOperationSucceeded(sqliteDB, entry.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s/%s: unable to mark synced: %v", entry.EntityType, entry.EntityID, err))
		}
	}
	return result, nil
}

func syncOneItem(postgresDB *sqlx.DB, sqliteDB *sql.DB, entry localdb.SyncQueueEntry) error {
	if entry.EntityType == "" {
		return fmt.Errorf("missing entity type")
	}

	payloadMap, err := parsePayload(entry.Payload)
	if err != nil {
		return err
	}
	if len(payloadMap) == 0 {
		return fmt.Errorf("empty payload")
	}
	if entry.EntityID != "" {
		payloadMap["id"] = entry.EntityID
	}
	if _, ok := payloadMap["id"]; !ok {
		if entry.EntityID == "" {
			return fmt.Errorf("missing id in payload")
		}
		payloadMap["id"] = entry.EntityID
	}

	tableName, err := tableNameForEntity(entry.EntityType)
	if err != nil {
		return err
	}
	if err := validateSyncPayload(tableName, payloadMap); err != nil {
		return err
	}
	if err := detectSyncConflict(postgresDB, sqliteDB, tableName, entry, payloadMap); err != nil {
		return err
	}

	switch strings.ToLower(entry.Operation) {
	case "delete":
		_, err = postgresDB.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, quoteIdentifier(tableName)), payloadMap["id"])
		return err
	case "create", "update", "upsert":
		return upsertEntity(postgresDB, tableName, payloadMap)
	default:
		return fmt.Errorf("unsupported operation %q", entry.Operation)
	}
}

func detectSyncConflict(postgresDB *sqlx.DB, sqliteDB *sql.DB, tableName string, entry localdb.SyncQueueEntry, payload map[string]any) error {
	localUpdatedAt := extractUpdatedAt(payload)
	if localUpdatedAt == "" || entry.EntityID == "" {
		return nil
	}

	remoteUpdatedAt, err := fetchEntityUpdatedAt(postgresDB, tableName, entry.EntityID)
	if err != nil {
		return err
	}
	if remoteUpdatedAt == "" {
		return nil
	}

	localTime, err := parseTimestamp(localUpdatedAt)
	if err != nil {
		return nil
	}
	remoteTime, err := parseTimestamp(remoteUpdatedAt)
	if err != nil {
		return nil
	}
	if !remoteTime.After(localTime) {
		return nil
	}

	reason := fmt.Sprintf("remote data is newer than local change (%s > %s)", remoteUpdatedAt, localUpdatedAt)
	if recordErr := localdb.RecordSyncConflict(sqliteDB, entry.EntityType, entry.EntityID, tableName, entry.Operation, localUpdatedAt, remoteUpdatedAt, reason, entry.Payload); recordErr != nil {
		return fmt.Errorf("detect sync conflict: %w", recordErr)
	}
	return fmt.Errorf("sync conflict: %s", reason)
}

func extractUpdatedAt(payload map[string]any) string {
	for _, key := range []string{"updated_at", "updatedAt", "last_updated", "lastUpdated"} {
		if value, ok := payload[key]; ok {
			switch v := value.(type) {
			case string:
				if strings.TrimSpace(v) != "" {
					return v
				}
			case time.Time:
				return v.Format(time.RFC3339)
			}
		}
	}
	return ""
}

func fetchEntityUpdatedAt(postgresDB *sqlx.DB, tableName, entityID string) (string, error) {
	if tableName == "" || entityID == "" {
		return "", nil
	}
	var updatedAt sql.NullString
	query := fmt.Sprintf(`SELECT updated_at FROM %s WHERE id = $1 LIMIT 1`, quoteIdentifier(tableName))
	if err := postgresDB.QueryRow(query, entityID).Scan(&updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("fetch %s updated_at: %w", tableName, err)
	}
	if !updatedAt.Valid {
		return "", nil
	}
	return updatedAt.String, nil
}

func parseTimestamp(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		return ts, nil
	}
	if ts, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return ts, nil
	}
	if ts, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
		return ts, nil
	}
	if ts, err := time.Parse("2006-01-02", value); err == nil {
		return ts, nil
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp format: %s", value)
}

func parsePayload(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("empty payload")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("parse payload: %w", err)
	}
	return normalizeMap(payload), nil
}

func normalizeMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		cleanKey := normalizeFieldName(key)
		if value == nil {
			continue
		}
		if mapValue, ok := value.(map[string]any); ok {
			for nestedKey, nestedValue := range mapValue {
				out[normalizeFieldName(nestedKey)] = nestedValue
			}
			continue
		}
		if _, ok := value.([]any); ok {
			continue
		}
		out[cleanKey] = value
	}
	return out
}

func normalizeFieldName(field string) string {
	if field == "" {
		return field
	}
	var out strings.Builder
	for i, r := range field {
		if unicode.IsUpper(r) {
			if i > 0 {
				out.WriteRune('_')
			}
			out.WriteRune(unicode.ToLower(r))
			continue
		}
		out.WriteRune(r)
	}
	return strings.Trim(strings.ReplaceAll(out.String(), "-", "_"), "_")
}

func upsertEntity(postgresDB *sqlx.DB, tableName string, payload map[string]any) error {
	if len(payload) == 0 {
		return fmt.Errorf("no values to sync")
	}

	columns := make([]string, 0, len(payload))
	values := make([]any, 0, len(payload))
	placeholders := make([]string, 0, len(payload))
	updates := make([]string, 0, len(payload))
	allowed := syncColumns(tableName)
	index := 1
	for key, value := range payload {
		switch value.(type) {
		case []any, map[string]any:
			continue
		}
		column := normalizeFieldName(key)
		if _, ok := allowed[column]; !ok {
			continue
		}
		columns = append(columns, column)
		values = append(values, value)
		placeholders = append(placeholders, fmt.Sprintf("$%d", index))
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", quoteIdentifier(column), quoteIdentifier(column)))
		index++
	}

	if len(columns) == 0 {
		return fmt.Errorf("no scalar fields found to sync")
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) DO UPDATE SET %s`,
		quoteIdentifier(tableName),
		joinQuoted(columns),
		strings.Join(placeholders, ", "),
		strings.Join(updates, ", "),
	)

	if _, err := postgresDB.Exec(query, values...); err != nil {
		return fmt.Errorf("upsert into %s: %w", tableName, err)
	}
	return nil
}

func syncColumns(tableName string) map[string]struct{} {
	columns := map[string]string{
		"products":                  "id sku barcode name description category_id brand_id preferred_supplier_id cost_price selling_price min_stock_level max_stock_level is_active created_at updated_at",
		"customers":                 "id code name email phone address city country tax_id credit_limit current_balance notes is_active created_at updated_at",
		"suppliers":                 "id code name email phone address city country tax_id payment_terms credit_limit current_balance notes is_active created_at updated_at",
		"inventory_items":           "id product_id part_type_id item_code barcode serial_number condition grade purchase_cost selling_price status location_id supplier_id purchase_date sold_at notes created_at updated_at",
		"sales":                     "id invoice_number customer_id user_id sale_date subtotal tax_amount discount_amount total_amount paid_amount remaining_amount payment_method payment_status status notes created_at updated_at",
		"purchases":                 "id invoice_number supplier_id user_id purchase_date subtotal tax_amount discount_amount total_amount paid_amount remaining_amount payment_method payment_status status notes created_at updated_at",
		"payments":                  "id reference_number sale_id purchase_id customer_id supplier_id amount payment_method payment_date notes created_at updated_at",
		"categories":                "id name description parent_id icon color is_active created_at updated_at",
		"brands":                    "id name description logo_url created_at updated_at",
		"sale_items":                "id sale_id inventory_item_id product_id quantity unit_price item_total total_amount unit_cost created_at",
		"purchase_items":            "id purchase_id product_id quantity unit_price item_total total_amount unit_cost created_at",
		"expenses":                  "id title category_id reference_number category amount description expense_date payment_method receipt_url created_by currency reference notes status is_recurring recurring_period approved_by created_at updated_at",
		"expense_categories":        "id name description color icon budget is_active created_at updated_at",
		"inventory":                 "id product_id quantity reserved_quantity location warehouse_id current_quantity available_quantity current_cost current_value last_movement_id last_restocked_at created_at updated_at",
		"inventory_movements":       "id item_id product_id movement_type quantity before_quantity after_quantity reference_type reference_id reason created_by created_at is_reversed reversed_by reversed_at reversal_reason",
		"locations":                 "id name type parent_id warehouse_id description is_active created_at updated_at",
		"reservations":              "id item_id customer_id user_id reserved_at expires_at status notes created_at updated_at",
		"barcodes":                  "id code product_id inventory_item_id type is_active generated_at created_at updated_at",
		"returns":                   "id return_number reference_number sale_id purchase_id customer_id return_date return_type status total_refund_amount refund_method refund_date refund_reference debt_id debt_adjustment customer_credit reason reason_detail item_condition_after_return is_warranty_claim warranty_id warranty_valid_until created_by processed_by approved_by approved_at notes internal_notes refund_status created_at updated_at",
		"return_items":              "id return_id sale_item_id product_id inventory_item_id quantity quantity_returned original_quantity serial_number barcode unit_price total_refund_amount reason original_condition returned_condition condition_notes resolution inventory_status inspection_required inspection_date inspection_result inspection_notes original_cost repair_cost created_at updated_at",
		"acquisitions":              "id type acquisition_date supplier_id customer_id total_cost paid_amount payment_status status notes user_id created_at updated_at reversed_at reversed_by reversal_reason",
		"acquisition_items":         "id acquisition_id product_id inventory_item_id inspection_id item_code serial_number condition grade unit_cost total_cost inspection_status item_status notes created_at updated_at",
		"trade_ins":                 "id customer_id inventory_item_id purchase_price purchase_date notes created_at updated_at",
		"inspections":               "id product_id inventory_item_id inspector_id inspection_date result condition grade notes images created_at updated_at",
		"inspection_items":          "id inspection_id item_id checkpoint_name status notes images created_at",
		"supplier_returns":          "id purchase_id supplier_id return_number status reason refund_amount notes created_by created_at updated_at",
		"supplier_return_items":     "id supplier_return_id purchase_item_id product_id quantity unit_cost created_at",
		"seller_payments":           "id acquisition_id customer_id amount payment_method payment_date notes user_id created_at",
		"part_types":                "id name_ar name_en icon color is_active sort_order created_at updated_at",
		"part_specifications":       "id name_ar name_en data_type options is_required created_at",
		"type_specifications":       "id part_type_id specification_id sort_order created_at",
		"item_specification_values": "id inventory_item_id specification_id value_text value_number value_boolean created_at updated_at",
	}
	set := make(map[string]struct{})
	for _, column := range strings.Fields(columns[tableName]) {
		set[column] = struct{}{}
	}
	return set
}

func validateSyncPayload(tableName string, payload map[string]any) error {
	allowed := syncColumns(tableName)
	if len(allowed) == 0 {
		return fmt.Errorf("no sync schema registered for %s", tableName)
	}
	if id, ok := payload["id"]; !ok || strings.TrimSpace(fmt.Sprint(id)) == "" {
		return fmt.Errorf("sync payload for %s is missing id", tableName)
	}
	for key, value := range payload {
		if _, ok := allowed[normalizeFieldName(key)]; !ok {
			return fmt.Errorf("field %q is not allowed for %s", key, tableName)
		}
		if _, nested := value.(map[string]any); nested {
			return fmt.Errorf("nested field %q is not allowed", key)
		}
		if _, nested := value.([]any); nested {
			return fmt.Errorf("array field %q is not allowed", key)
		}
	}
	return nil
}

func joinQuoted(cols []string) string {
	quoted := make([]string, 0, len(cols))
	for _, col := range cols {
		quoted = append(quoted, quoteIdentifier(col))
	}
	return strings.Join(quoted, ", ")
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func tableNameForEntity(entityType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "product", "products":
		return "products", nil
	case "customer", "customers":
		return "customers", nil
	case "supplier", "suppliers":
		return "suppliers", nil
	case "inventory_item", "inventory_items", "inventoryitem":
		return "inventory_items", nil
	case "sale", "sales":
		return "sales", nil
	case "purchase", "purchases":
		return "purchases", nil
	case "payment", "payments":
		return "payments", nil
	case "category", "categories":
		return "categories", nil
	case "brand", "brands":
		return "brands", nil
	case "sale_item", "sale_items":
		return "sale_items", nil
	case "purchase_item", "purchase_items":
		return "purchase_items", nil
	case "expense", "expenses":
		return "expenses", nil
	case "expense_category", "expense_categories":
		return "expense_categories", nil
	case "inventory", "inventories":
		return "inventory", nil
	case "inventory_movement", "inventory_movements":
		return "inventory_movements", nil
	case "location", "locations":
		return "locations", nil
	case "reservation", "reservations":
		return "reservations", nil
	case "barcode", "barcodes":
		return "barcodes", nil
	case "return", "returns":
		return "returns", nil
	case "return_item", "return_items":
		return "return_items", nil
	case "acquisition", "acquisitions":
		return "acquisitions", nil
	case "acquisition_item", "acquisition_items":
		return "acquisition_items", nil
	case "trade_in", "trade_ins":
		return "trade_ins", nil
	case "inspection", "inspections":
		return "inspections", nil
	case "inspection_item", "inspection_items":
		return "inspection_items", nil
	case "supplier_return", "supplier_returns":
		return "supplier_returns", nil
	case "supplier_return_item", "supplier_return_items":
		return "supplier_return_items", nil
	case "seller_payment", "seller_payments":
		return "seller_payments", nil
	case "part_type", "part_types":
		return "part_types", nil
	case "part_specification", "part_specifications":
		return "part_specifications", nil
	case "type_specification", "type_specifications":
		return "type_specifications", nil
	case "item_specification_value", "item_specification_values":
		return "item_specification_values", nil
	default:
		return "", fmt.Errorf("unsupported entity type %q for sync", entityType)
	}
}
