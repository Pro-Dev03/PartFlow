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
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if authorization == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "جلسة التحقق السحابي غير موجودة"})
		return
	}

	cloudBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PARTFLOW_CLOUD_API_URL")), "/")
	if cloudBaseURL == "" {
		cloudBaseURL = "https://partflow-api.onrender.com/api/v1"
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, cloudBaseURL+"/sync/initial-data", nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "تعذر تجهيز طلب المزامنة السحابية", "details": err.Error()})
		return
	}
	request.Header.Set("Authorization", authorization)
	// Forward the cloud token because the cloud API does not trust the local JWT.
	request.Header.Set("X-PartFlow-Cloud-Token", strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token")))
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

// SyncLocalDataToCloud pushes only pending local operations after the caller
// has passed the cloud-backed admin middleware. Full SQLite snapshots are
// never uploaded by this endpoint.
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
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"processed": 0, "failed": 0, "direction": "local_to_cloud"}})
		return
	}

	// Push only pending queue entries through the cloud API; the local service
	// must never use its PostgreSQL connection as an implicit sync channel.
	operations := make([]sync.PushOperation, 0, len(entries))
	for _, entry := range entries {
		operations = append(operations, sync.PushOperation{
			ID: entry.ID, EntityType: entry.EntityType, EntityID: entry.EntityID,
			Operation: entry.Operation, Payload: entry.Payload, IdempotencyKey: entry.Idempotency,
		})
	}
	payload, err := json.Marshal(sync.PushRequest{Operations: operations})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تجهيز طابور المزامنة", "details": err.Error()})
		return
	}
	cloudBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PARTFLOW_CLOUD_API_URL")), "/")
	if cloudBaseURL == "" {
		cloudBaseURL = "https://partflow-api.onrender.com/api/v1"
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, cloudBaseURL+"/sync/push", strings.NewReader(string(payload)))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "تعذر تجهيز طلب رفع المزامنة", "details": err.Error()})
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", strings.TrimSpace(c.GetHeader("Authorization")))
	// The cloud guard validates this token before accepting any queue entry.
	request.Header.Set("X-PartFlow-Cloud-Token", strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token")))
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"processed": result.Data.Processed,
			"failed":    result.Data.Failed,
			"direction": "local_to_cloud",
		},
	})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ØªØ¹Ø°Ø± Ù‚Ø±Ø§Ø¡Ø© Ø§Ù„ØªØ¹Ø§Ø±Ø¶", "details": err.Error()})
		return
	}
	if conflict == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ø§Ù„ØªØ¹Ø§Ø±Ø¶ ØºÙŠØ± Ù…ÙˆØ¬ÙˆØ¯"})
		return
	}

	// Apply the chosen policy before removing the conflict record. This makes
	// resolution durable and retryable if the cloud is unavailable.
	switch policy {
	case "server_wins":
		if err := h.applyCloudSnapshot(c, db.DB); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "ÙØ´Ù„ ØªØ·Ø¨ÙŠÙ‚ Ø¨ÙŠØ§Ù†Ø§Øª Ø§Ù„Ø®Ø§Ø¯Ù… Ø§Ù„Ø³Ø­Ø§Ø¨ÙŠ", "details": err.Error()})
			return
		}
	case "client_wins":
		var payload map[string]any
		if strings.TrimSpace(conflict["payload"]) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ø¨ÙŠØ§Ù† Ø§Ù„Ø¹Ù…ÙŠÙ„ ØºÙŠØ± Ù…ÙˆØ¬ÙˆØ¯ Ù„Ù„ØªØ¹Ø§Ø±Ø¶"})
			return
		}
		if err := json.Unmarshal([]byte(conflict["payload"]), &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ø¨ÙŠØ§Ù† Ø§Ù„ØªØ¹Ø§Ø±Ø¶ ØºÙŠØ± ØµØ§Ù„Ø­", "details": err.Error()})
			return
		}
		cloudSnapshotMu.Lock()
		err = localdb.SeedLocalSnapshot(db.DB, map[string]any{conflict["entity_table"]: []any{payload}})
		cloudSnapshotMu.Unlock()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ÙØ´Ù„ ØªØ·Ø¨ÙŠÙ‚ ØªØºÙŠÙŠØ± Ø§Ù„Ø¹Ù…ÙŠÙ„ Ù…Ø­Ù„ÙŠØ§Ù‹", "details": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ÙØ´Ù„ ØªØ·Ø¨ÙŠÙ‚ Ø§Ù„Ø¯Ù…Ø¬ Ø§Ù„ÙŠØ¯ÙˆÙŠ", "details": err.Error()})
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
	cloudBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PARTFLOW_CLOUD_API_URL")), "/")
	if cloudBaseURL == "" {
		cloudBaseURL = "https://partflow-api.onrender.com/api/v1"
	}
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
