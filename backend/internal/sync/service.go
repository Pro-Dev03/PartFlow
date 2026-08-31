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
		key   string
		query string
	}{
		{key: "categories", query: `SELECT * FROM categories ORDER BY created_at ASC`},
		{key: "products", query: `SELECT * FROM products ORDER BY created_at ASC`},
		{key: "customers", query: `SELECT * FROM customers ORDER BY created_at ASC`},
		{key: "suppliers", query: `SELECT * FROM suppliers ORDER BY created_at ASC`},
		{key: "sales", query: `SELECT * FROM sales ORDER BY created_at ASC`},
		{key: "sale_items", query: `SELECT * FROM sale_items ORDER BY created_at ASC`},
		{key: "purchases", query: `SELECT * FROM purchases ORDER BY created_at ASC`},
		{key: "purchase_items", query: `SELECT * FROM purchase_items ORDER BY created_at ASC`},
		{key: "payments", query: `SELECT * FROM payments ORDER BY created_at ASC`},
		{key: "debts", query: `SELECT * FROM debts ORDER BY created_at ASC`},
		{key: "expenses", query: `SELECT * FROM expenses ORDER BY created_at ASC`},
		{key: "inspections", query: `SELECT * FROM inspections ORDER BY created_at ASC`},
		{key: "part_types", query: `SELECT * FROM part_types ORDER BY created_at ASC`},
		{key: "returns", query: `SELECT * FROM returns ORDER BY created_at ASC`},
		{key: "return_items", query: `SELECT * FROM return_items ORDER BY created_at ASC`},
		{key: "inventory_items", query: `SELECT * FROM inventory_items ORDER BY created_at ASC`},
	}

	snapshot := make(map[string]any, len(queries))
	for _, item := range queries {
		rows, err := fetchRowsAsMaps(postgresDB, item.query)
		if err != nil {
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
	index := 1
	for key, value := range payload {
		switch value.(type) {
		case []any, map[string]any:
			continue
		}
		column := normalizeFieldName(key)
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
	default:
		return "", fmt.Errorf("unsupported entity type %q for sync", entityType)
	}
}
