package settings

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	stdsync "sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/internal/localdb"
	"github.com/partflow/smart-store/internal/sync"
)

type LocalDatabaseHandler struct{}

// SQLite allows one writer at a time. A browser can start the initial sync in
// more than one React effect (or open two windows), so serialize snapshot
// writes and metadata updates to avoid transient SQLITE_BUSY/constraint
// failures while keeping the cloud fetch itself concurrent-safe.
var cloudSnapshotMu stdsync.Mutex

var bidirectionalSyncTables = []struct {
	table  string
	entity string
}{
	{table: "categories", entity: "category"},
	{table: "brands", entity: "brand"},
	{table: "suppliers", entity: "supplier"},
	{table: "customers", entity: "customer"},
	{table: "products", entity: "product"},
	{table: "inventory", entity: "inventory"},
	{table: "inventory_items", entity: "inventory_item"},
	{table: "sales", entity: "sale"},
	{table: "sale_items", entity: "sale_item"},
	{table: "purchases", entity: "purchase"},
	{table: "purchase_items", entity: "purchase_item"},
	{table: "payments", entity: "payment"},
	{table: "debts", entity: "debt"},
	{table: "expense_categories", entity: "expense_category"},
	{table: "expenses", entity: "expense"},
	{table: "seller_payments", entity: "seller_payment"},
	{table: "supplier_returns", entity: "supplier_return"},
	{table: "supplier_return_items", entity: "supplier_return_item"},
	{table: "part_types", entity: "part_type"},
	{table: "part_specifications", entity: "part_specification"},
	{table: "type_specifications", entity: "type_specification"},
	{table: "acquisitions", entity: "acquisition"},
	{table: "acquisition_items", entity: "acquisition_item"},
	{table: "trade_ins", entity: "trade_in"},
	{table: "item_specification_values", entity: "item_specification_value"},
	{table: "returns", entity: "return"},
	{table: "return_items", entity: "return_item"},
}

type syncSnapshotResponse struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data"`
	Error   string         `json:"error"`
}

type bidirectionalSyncOperation struct {
	ID             string `json:"id"`
	EntityType     string `json:"entity_type"`
	EntityID       string `json:"entity_id"`
	Operation      string `json:"operation"`
	Payload        string `json:"payload"`
	IdempotencyKey string `json:"idempotency_key"`
}

func cloudSyncBaseURL() string {
	configured := strings.TrimRight(strings.TrimSpace(os.Getenv("PARTFLOW_CLOUD_API_URL")), "/")
	if configured == "" {
		return "https://partflow-api.onrender.com/api/v1"
	}
	return configured
}

type OfflineSessionRequest struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

func NewLocalDatabaseHandler() *LocalDatabaseHandler {
	return &LocalDatabaseHandler{}
}

// SyncCloudData downloads the authenticated user's snapshot from the cloud
// API and merges it into the local SQLite database. The desktop process must
// never use its local SQLite connection as if it were the cloud database.
func (h *LocalDatabaseHandler) SyncCloudData(c *gin.Context) {
	if strings.TrimSpace(c.GetHeader("Authorization")) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "جلسة التحقق السحابي غير موجودة"})
		return
	}
	cloudToken := strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token"))
	if cloudToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "رمز السحابة غير موجود"})
		return
	}

	cloudBaseURL := cloudSyncBaseURL()
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, cloudBaseURL+"/sync/initial-data", nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "تعذر تجهيز طلب المزامنة السحابية", "details": err.Error()})
		return
	}
	// The cloud API validates its own token, not the local session JWT.
	request.Header.Set("Authorization", "Bearer "+cloudToken)
	request.Header.Set("X-PartFlow-Cloud-Token", cloudToken)
	request.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "تعذر الاتصال بالخادم السحابي للمزامنة", "details": err.Error()})
		return
	}
	defer response.Body.Close()

	var payload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
		Error   string         `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "استجابة المزامنة السحابية غير صالحة", "details": err.Error()})
		return
	}
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		message := payload.Error
		if message == "" {
			message = "رفض الخادم السحابي جلسة المستخدم"
		}
		c.JSON(response.StatusCode, gin.H{"error": message})
		return
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !payload.Success {
		message := payload.Error
		if message == "" {
			message = "فشل جلب البيانات من الخادم السحابي"
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": message})
		return
	}

	sqliteDB, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة البيانات المحلية", "details": err.Error()})
		return
	}
	defer sqliteDB.DB.Close()
	cloudSnapshotMu.Lock()
	defer cloudSnapshotMu.Unlock()
	if err := localdb.SeedLocalSnapshot(sqliteDB.DB, payload.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حفظ بيانات السحابة محلياً", "details": err.Error()})
		return
	}
	syncCompletedAt := time.Now().UTC().Format(time.RFC3339)
	if err := localdb.SetMetadata(sqliteDB.DB, "last_cloud_sync_at", syncCompletedAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تسجيل اكتمال المزامنة المحلية", "details": err.Error()})
		return
	}

	rowCounts := make(map[string]int, len(payload.Data))
	for table, rawRows := range payload.Data {
		if rows, ok := rawRows.([]any); ok {
			rowCounts[table] = len(rows)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"source":       "cloud",
			"destination":  "local_sqlite",
			"completed_at": syncCompletedAt,
			"tables":       rowCounts,
		},
	})
}

// SyncLocalDataToCloud reconciles pending operations and existing local data
// after the caller has passed the cloud-backed authentication middleware.
func (h *LocalDatabaseHandler) SyncLocalDataToCloud(c *gin.Context) {
	sqliteDB, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة البيانات المحلية", "details": err.Error()})
		return
	}
	defer sqliteDB.DB.Close()

	entries, err := localdb.ListPendingSyncOperations(sqliteDB.DB, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل قراءة طابور المزامنة المحلي", "details": err.Error()})
		return
	}
	if len(entries) == 0 {
		cloudToken := strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token"))
		result, reconcileErr := reconcileLocalAndCloud(c, sqliteDB.DB, cloudToken)
		if reconcileErr != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": reconcileErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
		return
	}

	// Push only pending queue entries through the cloud API; the local service
	// must never use its PostgreSQL connection as an implicit sync channel.
	operations := make([]sync.PushOperation, 0, len(entries))
	for _, entry := range entries {
		payload, normalizeErr := sync.NormalizeCloudPayloadJSON(entry.EntityType, entry.Payload)
		if normalizeErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تجهيز عملية المزامنة المحلية", "details": normalizeErr.Error()})
			return
		}
		operations = append(operations, sync.PushOperation{
			ID: entry.ID, EntityType: entry.EntityType, EntityID: entry.EntityID,
			Operation: entry.Operation, Payload: string(payload), IdempotencyKey: entry.Idempotency,
		})
	}
	payload, err := json.Marshal(sync.PushRequest{Operations: operations})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تجهيز طابور المزامنة", "details": err.Error()})
		return
	}
	cloudBaseURL := cloudSyncBaseURL()
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, cloudBaseURL+"/sync/push", strings.NewReader(string(payload)))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "تعذر تجهيز طلب رفع المزامنة", "details": err.Error()})
		return
	}
	request.Header.Set("Content-Type", "application/json")
	cloudToken := strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token"))
	// The cloud API validates its own token, not the local session JWT.
	request.Header.Set("Authorization", "Bearer "+cloudToken)
	request.Header.Set("X-PartFlow-Cloud-Token", cloudToken)
	response, err := (&http.Client{Timeout: 20 * time.Second}).Do(request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "تعذر الاتصال بالخادم السحابي للمزامنة", "details": err.Error()})
		return
	}
	defer response.Body.Close()
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			AcceptedIDs []string `json:"accepted_ids"`
			Rejected    []struct {
				ID       string `json:"id"`
				Error    string `json:"error"`
				Conflict bool   `json:"conflict"`
			} `json:"rejected"`
			Processed int `json:"processed"`
			Failed    int `json:"failed"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil || response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !result.Success {
		c.JSON(http.StatusBadGateway, gin.H{"error": "فشل رفع التغييرات المحلية إلى السحابة"})
		return
	}

	// Mark only cloud-accepted entries as synced so rejected or conflicted data
	// remains locally retryable and visible to the owner.
	accepted := make(map[string]struct{}, len(result.Data.AcceptedIDs))
	for _, id := range result.Data.AcceptedIDs {
		accepted[id] = struct{}{}
	}
	for _, entry := range entries {
		if _, ok := accepted[entry.ID]; ok {
			_ = localdb.MarkSyncOperationSucceeded(sqliteDB.DB, entry.ID)
		}
	}
	for _, rejected := range result.Data.Rejected {
		for _, entry := range entries {
			if entry.ID != rejected.ID {
				continue
			}
			if rejected.Conflict {
				_ = localdb.RecordSyncConflict(sqliteDB.DB, entry.EntityType, entry.EntityID, entry.EntityType, entry.Operation, "", "", rejected.Error, entry.Payload)
			}
			_ = localdb.MarkSyncOperationFailed(sqliteDB.DB, entry.ID, rejected.Error)
		}
	}

	acceptedIDs := make(map[string]struct{}, len(result.Data.AcceptedIDs))
	for _, id := range result.Data.AcceptedIDs {
		acceptedIDs[id] = struct{}{}
	}
	operationDetails := make([]gin.H, 0, len(entries))
	for _, entry := range entries {
		detail := gin.H{
			"id": entry.ID, "entity_type": entry.EntityType, "entity_id": entry.EntityID,
			"operation": entry.Operation, "status": "failed",
		}
		if _, ok := acceptedIDs[entry.ID]; ok {
			detail["status"] = "processed"
		} else {
			for _, rejected := range result.Data.Rejected {
				if rejected.ID == entry.ID {
					detail["error"] = rejected.Error
					if rejected.Conflict {
						detail["status"] = "conflict"
					}
					break
				}
			}
		}
		operationDetails = append(operationDetails, detail)
	}

	reconcileResult, reconcileErr := reconcileLocalAndCloud(c, sqliteDB.DB, strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token")))
	if reconcileErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": reconcileErr.Error()})
		return
	}
	reconcileProcessed, _ := reconcileResult["processed"].(int)
	reconcileFailed, _ := reconcileResult["failed"].(int)
	reconcileOperations, _ := reconcileResult["operations"].([]gin.H)
	operationDetails = append(operationDetails, reconcileOperations...)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"processed":  result.Data.Processed + reconcileProcessed,
			"failed":     result.Data.Failed + reconcileFailed,
			"direction":  "bidirectional",
			"operations": operationDetails,
		},
	})
}

func reconcileLocalAndCloud(c *gin.Context, sqliteDB *sql.DB, cloudToken string) (gin.H, error) {
	if cloudToken == "" {
		return nil, fmt.Errorf("رمز السحابة غير موجود")
	}
	cloudBaseURL := cloudSyncBaseURL()
	cloudSnapshot, err := fetchCloudSnapshot(c, cloudBaseURL, cloudToken)
	if err != nil {
		return nil, err
	}
	localSnapshot, err := localdb.ExportLocalSnapshot(sqliteDB)
	if err != nil {
		return nil, fmt.Errorf("فشل قراءة البيانات المحلية: %w", err)
	}

	operations := make([]bidirectionalSyncOperation, 0)
	for _, table := range bidirectionalSyncTables {
		localRows := snapshotRows(localSnapshot[table.table])
		cloudRows := snapshotRows(cloudSnapshot[table.table])
		cloudByID := make(map[string]map[string]any, len(cloudRows))
		for _, row := range cloudRows {
			if id := snapshotRowID(row); id != "" {
				cloudByID[id] = row
			}
		}
		for _, row := range localRows {
			id := snapshotRowID(row)
			if id == "" {
				continue
			}
			remote, exists := cloudByID[id]
			if exists && (snapshotRowsEquivalent(table.table, row, remote) || !localSnapshotIsNewer(row, remote)) {
				continue
			}
			normalizedRow := normalizeOutboundSnapshotRow(table.table, row)
			if len(normalizedRow) == 0 {
				continue
			}
			payload, marshalErr := json.Marshal(normalizedRow)
			if marshalErr != nil {
				return nil, fmt.Errorf("فشل تجهيز %s/%s: %w", table.table, id, marshalErr)
			}
			operations = append(operations, bidirectionalSyncOperation{
				ID:             fmt.Sprintf("snapshot:%s:%s", table.table, id),
				EntityType:     table.entity,
				EntityID:       id,
				Operation:      "upsert",
				Payload:        string(payload),
				IdempotencyKey: fmt.Sprintf("snapshot:%s:%s", table.table, id),
			})
		}
	}

	processed, failed, details, err := pushSnapshotOperations(c, cloudBaseURL, cloudToken, operations)
	if err != nil {
		return nil, err
	}
	cloudSnapshot, err = fetchCloudSnapshot(c, cloudBaseURL, cloudToken)
	if err != nil {
		return nil, err
	}
	cloudSnapshotMu.Lock()
	seedErr := localdb.SeedLocalSnapshot(sqliteDB, cloudSnapshot)
	cloudSnapshotMu.Unlock()
	if seedErr != nil {
		return nil, fmt.Errorf("فشل دمج البيانات السحابية محليًا: %w", seedErr)
	}

	return gin.H{
		"processed":  processed,
		"failed":     failed,
		"direction":  "bidirectional",
		"operations": details,
	}, nil
}

func fetchCloudSnapshot(c *gin.Context, cloudBaseURL, cloudToken string) (map[string]any, error) {
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, cloudBaseURL+"/sync/initial-data", nil)
	if err != nil {
		return nil, fmt.Errorf("تعذر تجهيز تنزيل المزامنة: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+cloudToken)
	request.Header.Set("X-PartFlow-Cloud-Token", cloudToken)
	response, err := (&http.Client{Timeout: 20 * time.Second}).Do(request)
	if err != nil {
		return nil, fmt.Errorf("تعذر الاتصال بالخادم السحابي للمزامنة: %w", err)
	}
	defer response.Body.Close()
	var payload syncSnapshotResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("استجابة المزامنة السحابية غير صالحة: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !payload.Success {
		if payload.Error == "" {
			payload.Error = "فشل جلب بيانات السحابة"
		}
		return nil, fmt.Errorf("%s", payload.Error)
	}
	return payload.Data, nil
}

func pushSnapshotOperations(c *gin.Context, cloudBaseURL, cloudToken string, operations []bidirectionalSyncOperation) (int, int, []gin.H, error) {
	processed, failed := 0, 0
	details := make([]gin.H, 0, len(operations))
	for start := 0; start < len(operations); start += 50 {
		end := start + 50
		if end > len(operations) {
			end = len(operations)
		}
		body, err := json.Marshal(sync.PushRequest{Operations: toPushOperations(operations[start:end])})
		if err != nil {
			return processed, failed, details, fmt.Errorf("فشل تجهيز دفعة المزامنة: %w", err)
		}
		request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, cloudBaseURL+"/sync/push", strings.NewReader(string(body)))
		if err != nil {
			return processed, failed, details, fmt.Errorf("تعذر تجهيز رفع المزامنة: %w", err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+cloudToken)
		request.Header.Set("X-PartFlow-Cloud-Token", cloudToken)
		response, err := (&http.Client{Timeout: 20 * time.Second}).Do(request)
		if err != nil {
			return processed, failed, details, fmt.Errorf("تعذر رفع بيانات المزامنة: %w", err)
		}
		var result struct {
			Success bool `json:"success"`
			Data    struct {
				AcceptedIDs []string `json:"accepted_ids"`
				Rejected    []struct {
					ID       string `json:"id"`
					Error    string `json:"error"`
					Conflict bool   `json:"conflict"`
				} `json:"rejected"`
				Processed int `json:"processed"`
				Failed    int `json:"failed"`
			} `json:"data"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&result)
		response.Body.Close()
		if decodeErr != nil || response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !result.Success {
			return processed, failed, details, fmt.Errorf("فشل رفع دفعة المزامنة")
		}
		processed += result.Data.Processed
		failed += result.Data.Failed
		accepted := make(map[string]struct{}, len(result.Data.AcceptedIDs))
		for _, id := range result.Data.AcceptedIDs {
			accepted[id] = struct{}{}
		}
		for _, operation := range operations[start:end] {
			status := "failed"
			var operationError string
			if _, ok := accepted[operation.ID]; ok {
				status = "processed"
			} else {
				for _, rejected := range result.Data.Rejected {
					if rejected.ID == operation.ID {
						operationError = rejected.Error
						break
					}
				}
			}
			detail := gin.H{"id": operation.ID, "entity_type": operation.EntityType, "entity_id": operation.EntityID, "operation": operation.Operation, "status": status}
			if operationError != "" {
				detail["error"] = operationError
			}
			details = append(details, detail)
		}
	}
	return processed, failed, details, nil
}

func toPushOperations(operations []bidirectionalSyncOperation) []sync.PushOperation {
	result := make([]sync.PushOperation, len(operations))
	for index, operation := range operations {
		result[index] = sync.PushOperation{ID: operation.ID, EntityType: operation.EntityType, EntityID: operation.EntityID, Operation: operation.Operation, Payload: operation.Payload, IdempotencyKey: operation.IdempotencyKey}
	}
	return result
}

func snapshotRows(raw any) []map[string]any {
	switch rows := raw.(type) {
	case []map[string]any:
		return rows
	case []any:
		result := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			if mapped, ok := row.(map[string]any); ok {
				result = append(result, mapped)
			}
		}
		return result
	default:
		return nil
	}
}

func snapshotRowID(row map[string]any) string {
	if value, ok := row["id"]; ok && value != nil {
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return ""
}

func localSnapshotIsNewer(localRow, cloudRow map[string]any) bool {
	localTime, localOK := snapshotTimestamp(localRow)
	cloudTime, cloudOK := snapshotTimestamp(cloudRow)
	return localOK && (!cloudOK || localTime.After(cloudTime))
}

func snapshotTimestamp(row map[string]any) (time.Time, bool) {
	for _, key := range []string{"updated_at", "created_at"} {
		value, ok := row[key]
		if !ok || value == nil || strings.TrimSpace(fmt.Sprint(value)) == "" {
			continue
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
			if parsed, err := time.Parse(layout, fmt.Sprint(value)); err == nil {
				return parsed, true
			}
		}
	}
	return time.Time{}, false
}

func normalizeOutboundSnapshotRow(table string, row map[string]any) map[string]any {
	result := make(map[string]any, len(row))
	for key, value := range row {
		result[key] = normalizeOutboundSnapshotValue(value)
	}
	if table == "sales" && (result["invoice_number"] == nil || strings.TrimSpace(fmt.Sprint(result["invoice_number"])) == "") {
		result["invoice_number"] = result["sale_number"]
	}
	if table == "purchases" && (result["invoice_number"] == nil || strings.TrimSpace(fmt.Sprint(result["invoice_number"])) == "") {
		result["invoice_number"] = result["purchase_number"]
	}
	if table == "expenses" && (result["reference_number"] == nil || strings.TrimSpace(fmt.Sprint(result["reference_number"])) == "") {
		result["reference_number"] = result["id"]
	}
	if table == "expenses" && (result["category"] == nil || strings.TrimSpace(fmt.Sprint(result["category"])) == "") {
		result["category"] = "عام"
	}
	if table == "return_items" {
		inspectionResult := strings.ToLower(strings.TrimSpace(fmt.Sprint(result["inspection_result"])))
		switch inspectionResult {
		case "passed", "failed", "needs_repair", "condemned":
			result["inspection_result"] = inspectionResult
		default:
			delete(result, "inspection_result")
		}
	}
	delete(result, "sale_number")
	delete(result, "purchase_number")
	return sync.NormalizeCloudPayload(table, sync.FilterPushPayload(table, result))
}

func normalizeOutboundSnapshotValue(value any) any {
	text, ok := value.(string)
	if !ok {
		return value
	}
	if index := strings.Index(text, " m="); index > 0 {
		text = strings.TrimSpace(text[:index])
	}
	for _, layout := range []string{
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
	} {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed.UTC().Format(time.RFC3339Nano)
		}
	}
	return text
}

func snapshotRowsEquivalent(table string, localRow, cloudRow map[string]any) bool {
	localPayload, localErr := json.Marshal(normalizeOutboundSnapshotRow(table, localRow))
	cloudPayload, cloudErr := json.Marshal(normalizeOutboundSnapshotRow(table, cloudRow))
	return localErr == nil && cloudErr == nil && string(localPayload) == string(cloudPayload)
}

// SyncOfflineQueue is kept as a compatibility alias for older clients. The
// current policy is cloud-to-local synchronization, so it delegates to the
// same safe implementation.
func (h *LocalDatabaseHandler) SyncOfflineQueue(c *gin.Context) {
	h.SyncCloudData(c)
}

func (h *LocalDatabaseHandler) GetSyncConflicts(c *gin.Context) {
	db, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة بيانات التشغيل المحلية", "details": err.Error()})
		return
	}
	defer db.DB.Close()

	conflicts, err := localdb.ListSyncConflicts(db.DB, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل قراءة تعارضات المزامنة", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"conflicts": conflicts,
			"count":     len(conflicts),
		},
	})
}

func (h *LocalDatabaseHandler) ClearSyncConflicts(c *gin.Context) {
	db, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة بيانات التشغيل المحلية", "details": err.Error()})
		return
	}
	defer db.DB.Close()

	if err := localdb.ClearSyncConflicts(db.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف تعارضات المزامنة", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"cleared": true}})
}

func (h *LocalDatabaseHandler) ResolveSyncConflict(c *gin.Context) {
	conflictID := c.Param("id")
	if strings.TrimSpace(conflictID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف التعارض غير موجود"})
		return
	}

	var request struct {
		Policy     string         `json:"policy"`
		MergedData map[string]any `json:"merged_data"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تعذر قراءة سياسة حل التعارض", "details": err.Error()})
		return
	}

	policy := strings.ToLower(strings.TrimSpace(request.Policy))
	if policy == "" {
		policy = "manual_merge"
	}
	if policy != "server_wins" && policy != "client_wins" && policy != "manual_merge" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "سياسة حل التعارض غير صالحة", "details": "policy must be one of server_wins, client_wins, manual_merge"})
		return
	}

	db, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة بيانات التشغيل المحلية", "details": err.Error()})
		return
	}
	defer db.DB.Close()
	conflict, err := localdb.GetSyncConflict(db.DB, conflictID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "تعذر قراءة التعارض", "details": err.Error()})
		return
	}
	if conflict == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "التعارض غير موجود"})
		return
	}

	// Apply the chosen policy before removing the conflict record. This makes
	// resolution durable and retryable if the cloud is unavailable.
	switch policy {
	case "server_wins":
		if err := h.applyCloudSnapshot(c, db.DB); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "فشل تطبيق بيانات الخادم السحابي", "details": err.Error()})
			return
		}
	case "client_wins":
		var payload map[string]any
		if strings.TrimSpace(conflict["payload"]) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "بيان العميل غير موجود للتعارض"})
			return
		}
		if err := json.Unmarshal([]byte(conflict["payload"]), &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "بيان التعارض غير صالح", "details": err.Error()})
			return
		}
		cloudSnapshotMu.Lock()
		err = localdb.SeedLocalSnapshot(db.DB, map[string]any{conflict["entity_table"]: []any{payload}})
		cloudSnapshotMu.Unlock()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تطبيق تغيير العميل محليًا", "details": err.Error()})
			return
		}
	case "manual_merge":
		if len(request.MergedData) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "manual_merge يتطلب merged_data"})
			return
		}
		cloudSnapshotMu.Lock()
		err = localdb.SeedLocalSnapshot(db.DB, map[string]any{conflict["entity_table"]: []any{request.MergedData}})
		cloudSnapshotMu.Unlock()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تطبيق الدمج اليدوي", "details": err.Error()})
			return
		}
	}

	if err := localdb.DeleteSyncConflict(db.DB, conflictID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حل التعارض", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"resolved": true,
			"policy":   policy,
			"id":       conflictID,
		},
	})
}

func (h *LocalDatabaseHandler) applyCloudSnapshot(c *gin.Context, sqliteDB *sql.DB) error {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if authorization == "" {
		return fmt.Errorf("missing authorization")
	}
	cloudBaseURL := cloudSyncBaseURL()
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, cloudBaseURL+"/sync/initial-data", nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", authorization)
	response, err := (&http.Client{Timeout: 20 * time.Second}).Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var payload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
		Error   string         `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !payload.Success {
		return fmt.Errorf("cloud snapshot rejected: %s", payload.Error)
	}
	cloudSnapshotMu.Lock()
	defer cloudSnapshotMu.Unlock()
	if err := localdb.SeedLocalSnapshot(sqliteDB, payload.Data); err != nil {
		return err
	}
	return localdb.SetMetadata(sqliteDB, "last_cloud_sync_at", time.Now().UTC().Format(time.RFC3339))
}

func (h *LocalDatabaseHandler) GetOfflineSession(c *gin.Context) {
	db, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة بيانات التشغيل المحلية", "details": err.Error()})
		return
	}
	defer db.DB.Close()

	session, err := localdb.GetLocalSession(db.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل قراءة جلسة أوفلاين", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"session":      session,
			"is_available": session != nil,
		},
	})
}

func (h *LocalDatabaseHandler) SaveOfflineSession(c *gin.Context) {
	var request OfflineSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تعذر قراءة بيانات الجلسة المحلية", "details": err.Error()})
		return
	}
	if strings.TrimSpace(request.Email) == "" || strings.TrimSpace(request.AccessToken) == "" || strings.TrimSpace(request.RefreshToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الجلسة غير مكتملة"})
		return
	}

	db, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة بيانات التشغيل المحلية", "details": err.Error()})
		return
	}
	defer db.DB.Close()

	if err := localdb.SaveLocalSession(db.DB, request.UserID, request.Email, request.DisplayName, request.AccessToken, request.RefreshToken, request.ExpiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حفظ الجلسة المحلية", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"saved": true}})
}

func (h *LocalDatabaseHandler) ClearOfflineSession(c *gin.Context) {
	db, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة بيانات التشغيل المحلية", "details": err.Error()})
		return
	}
	defer db.DB.Close()

	if err := localdb.ClearLocalSession(db.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف الجلسة المحلية", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"cleared": true}})
}
