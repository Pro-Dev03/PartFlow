package sync

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

type Handler struct {
	db *sqlx.DB
}

// The cloud schema is single-tenant: most operational tables do not have a
// user_id column. The protected route already validates the subscriber, so
// filtering those tables by user_id would silently return an empty snapshot.
var snapshotTables = []struct {
	key      string
	table    string
	required bool
}{
	{key: "categories", table: "categories", required: true},
	{key: "brands", table: "brands"},
	{key: "suppliers", table: "suppliers", required: true},
	{key: "customers", table: "customers", required: true},
	{key: "customer_ledger", table: "customer_ledger"},
	{key: "supplier_ledger", table: "supplier_ledger"},
	{key: "ledger_entries", table: "ledger_entries"},
	{key: "products", table: "products", required: true},
	{key: "inventory", table: "inventory"},
	{key: "locations", table: "locations"},
	{key: "inventory_items", table: "inventory_items", required: true},
	{key: "inventory_movements", table: "inventory_movements"},
	{key: "reservations", table: "reservations"},
	{key: "barcodes", table: "barcodes"},
	{key: "sales", table: "sales", required: true},
	{key: "sale_items", table: "sale_items", required: true},
	{key: "purchases", table: "purchases", required: true},
	{key: "purchase_items", table: "purchase_items", required: true},
	{key: "payments", table: "payments", required: true},
	{key: "debts", table: "debts", required: true},
	{key: "expenses", table: "expenses", required: true},
	{key: "expense_categories", table: "expense_categories"},
	{key: "seller_payments", table: "seller_payments"},
	{key: "supplier_returns", table: "supplier_returns"},
	{key: "supplier_return_items", table: "supplier_return_items"},
	{key: "part_types", table: "part_types"},
	{key: "part_specifications", table: "part_specifications"},
	{key: "type_specifications", table: "type_specifications"},
	{key: "acquisitions", table: "acquisitions"},
	{key: "acquisition_items", table: "acquisition_items"},
	{key: "trade_ins", table: "trade_ins"},
	{key: "item_specification_values", table: "item_specification_values"},
	{key: "returns", table: "returns"},
	{key: "return_items", table: "return_items"},
	{key: "notifications", table: "notifications"},
	{key: "notification_preferences", table: "notification_preferences"},
	{key: "reports", table: "reports"},
	{key: "settings", table: "settings"},
	{key: "held_sales", table: "held_sales"},
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}

type PushOperation struct {
	ID             string `json:"id"`
	EntityType     string `json:"entity_type"`
	EntityID       string `json:"entity_id"`
	Operation      string `json:"operation"`
	Payload        string `json:"payload"`
	IdempotencyKey string `json:"idempotency_key"`
}

type PushRequest struct {
	Operations []PushOperation `json:"operations"`
}

type PushRejectedOperation struct {
	ID       string `json:"id"`
	Error    string `json:"error"`
	Conflict bool   `json:"conflict"`
}

// PushData accepts queue entries only, keeping the cloud contract separate
// from full SQLite snapshots and preserving manual owner-controlled sync.
func (h *Handler) PushData(c *gin.Context) {
	var request PushRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid sync push payload"})
		return
	}
	if len(request.Operations) == 0 || len(request.Operations) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "sync push must contain between 1 and 50 operations"})
		return
	}

	accepted := make([]string, 0, len(request.Operations))
	rejected := make([]PushRejectedOperation, 0)
	for _, operation := range request.Operations {
		entry := localdb.SyncQueueEntry{
			ID: operation.ID, EntityType: operation.EntityType, EntityID: operation.EntityID,
			Operation: operation.Operation, Payload: operation.Payload, Idempotency: operation.IdempotencyKey,
		}
		if err := ApplyCloudOperation(h.db, entry); err != nil {
			rejected = append(rejected, PushRejectedOperation{
				ID: operation.ID, Error: err.Error(),
				Conflict: strings.Contains(strings.ToLower(err.Error()), "sync conflict"),
			})
			continue
		}
		accepted = append(accepted, operation.ID)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"accepted_ids": accepted, "rejected": rejected,
		"processed": len(accepted), "failed": len(rejected),
	}})
}

// GetInitialData returns all user data for initial sync
func (h *Handler) GetInitialData(c *gin.Context) {
	// The authentication middleware stores the canonical string form under
	// user_id_string, while older tests/integrations may still provide a plain
	// user_id string. Accept both keys so an authenticated local request is not
	// rejected with a misleading 401.
	userID := c.GetString("user_id_string")
	if userID == "" {
		userID = c.GetString("user_id")
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found in context",
		})
		return
	}

	data := make(gin.H, len(snapshotTables))
	for _, item := range snapshotTables {
		var rows []map[string]interface{}
		var err error
		if item.table == "notifications" || item.table == "notification_preferences" {
			rows, err = h.getUserSnapshotTable(item.table, userID)
		} else {
			rows, err = h.getSnapshotTable(item.table)
		}
		if err != nil {
			// Optional tables may not exist on older cloud deployments. They do
			// not prevent the core customer/sales snapshot from being imported.
			if !item.required && isMissingTableError(err) {
				continue
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("failed to fetch cloud table %s", item.table),
			})
			return
		}
		data[item.key] = rows
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (h *Handler) getUserSnapshotTable(table, userID string) ([]map[string]interface{}, error) {
	if table != "notifications" && table != "notification_preferences" {
		return nil, fmt.Errorf("unsupported user snapshot table %q", table)
	}
	query := fmt.Sprintf(`SELECT * FROM "%s" WHERE user_id = $1 ORDER BY 1 ASC`, strings.ReplaceAll(table, `"`, `""`))
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			return nil, err
		}
		for key, value := range record {
			if bytes, ok := value.([]byte); ok {
				record[key] = string(bytes)
			}
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

// getSnapshotTable returns complete rows. SeedLocalSnapshot filters each row
// against the local SQLite schema, so cloud fields from later migrations do
// not cause the row to be dropped.
func (h *Handler) getSnapshotTable(table string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`SELECT * FROM "%s" ORDER BY 1 ASC`, strings.ReplaceAll(table, `"`, `""`))
	rows, err := h.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			return nil, err
		}
		// PostgreSQL returns JSON/JSONB values as []byte. Sending them as UTF-8
		// strings avoids encoding them as base64 in the HTTP response.
		for key, value := range record {
			if bytes, ok := value.([]byte); ok {
				record[key] = string(bytes)
			}
		}
		result = append(result, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func isMissingTableError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "does not exist") || strings.Contains(message, "no such table")
}
