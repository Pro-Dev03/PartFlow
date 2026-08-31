package settings

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/internal/localdb"
	"github.com/partflow/smart-store/internal/sync"
	"github.com/partflow/smart-store/pkg/database"
)

type LocalDatabaseHandler struct{}

type OperatingModeRequest struct {
	Mode string `json:"mode"`
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

func (h *LocalDatabaseHandler) GetOperatingMode(c *gin.Context) {
	database, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "فشل فتح قاعدة البيانات المحلية",
			"details": err.Error(),
		})
		return
	}
	defer database.DB.Close()

	mode, err := localdb.GetMetadata(database.DB, "operating_mode")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "فشل قراءة وضع التشغيل المحلي",
			"details": err.Error(),
		})
		return
	}
	if mode == "" {
		mode = "offline"
	}

	pendingSyncCount, err := localdb.GetPendingSyncCount(database.DB)
	if err != nil {
		pendingSyncCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"mode":                 mode,
			"local_database":       database.Path,
			"database_initialized": true,
			"pending_sync_count":   pendingSyncCount,
		},
	})
}

func (h *LocalDatabaseHandler) SetOperatingMode(c *gin.Context) {
	var request OperatingModeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تعذر قراءة وضع التشغيل", "details": err.Error()})
		return
	}
	if request.Mode != "offline" && request.Mode != "online" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "وضع التشغيل غير صالح", "details": "mode must be offline or online"})
		return
	}

	localDB, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "فشل تهيئة قاعدة البيانات المحلية",
			"details": err.Error(),
		})
		return
	}
	defer localDB.DB.Close()

	if err := localdb.SetMetadata(localDB.DB, "operating_mode", request.Mode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "فشل حفظ وضع التشغيل المحلي",
			"details": err.Error(),
		})
		return
	}

	if request.Mode == "offline" {
		postgresDB := database.GetDB()
		if postgresDB == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "قاعدة البيانات الرئيسية غير متاحة"})
			return
		}
		if err := sync.SeedLocalDatabaseFromOnline(postgresDB, localDB.DB); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "فشل تجهيز البيانات للـ offline",
				"details": err.Error(),
			})
			return
		}
	}

	pendingSyncCount, err := localdb.GetPendingSyncCount(localDB.DB)
	if err != nil {
		pendingSyncCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"mode":                 request.Mode,
			"local_database":       localDB.Path,
			"database_initialized": true,
			"pending_sync_count":   pendingSyncCount,
		},
	})
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
		Policy string `json:"policy"`
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
