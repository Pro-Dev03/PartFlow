package settings

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/pkg/response"
)

// CleanupCategory is a user-facing group of expired, non-financial records.
// EstimatedBytes is intentionally approximate and is not a promise that the
// database file or hosted-plan quota will immediately shrink by this amount.
type CleanupCategory struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Count          int64  `json:"count"`
	EstimatedBytes int64  `json:"estimated_bytes"`
}

type CleanupStorage struct {
	DatabaseBytes int64  `json:"database_bytes"`
	ReusableBytes int64  `json:"reusable_bytes"`
	Note          string `json:"note"`
}

type CleanupPreview struct {
	Target         string            `json:"target"`
	Categories     []CleanupCategory `json:"categories"`
	TotalCount     int64             `json:"total_count"`
	EstimatedBytes int64             `json:"estimated_bytes"`
	Storage        CleanupStorage    `json:"storage"`
}

type CleanupResult struct {
	Target       string         `json:"target"`
	DeletedCount int64          `json:"deleted_count"`
	Remaining    int64          `json:"remaining_count"`
	Storage      CleanupStorage `json:"storage"`
}

type cleanupSpec struct {
	key      string
	label    string
	table    string
	required []string
	depends  []string
	wherePG  string
	whereSQL string
	sizeSQL  string
}

type cleanupExecutor interface {
	GetContext(context.Context, any, string, ...any) error
	SelectContext(context.Context, any, string, ...any) error
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Rebind(string) string
}

func cleanupSpecs(sqlite bool) []cleanupSpec {
	now := "NOW()"
	cutoff90 := "NOW() - INTERVAL '90 days'"
	if sqlite {
		now = "datetime('now')"
		cutoff90 = "datetime('now', '-90 days')"
	}
	return []cleanupSpec{
		{
			key: "expired_idempotency", label: "طلبات منتهية لمنع التكرار", table: "idempotency_keys",
			required: []string{"id", "expires_at", "idempotency_key", "request_hash", "response_body"},
			wherePG:  "expires_at <= " + now,
			whereSQL: "datetime(expires_at) <= " + now,
			sizeSQL:  "COALESCE(LENGTH(CAST(idempotency_key AS TEXT)),0)+COALESCE(LENGTH(CAST(request_hash AS TEXT)),0)+COALESCE(LENGTH(CAST(response_body AS TEXT)),0)",
		},
		{
			key: "expired_refresh_tokens", label: "جلسات دخول منتهية", table: "refresh_tokens",
			required: []string{"id", "expires_at", "token"},
			wherePG:  "expires_at <= " + now,
			whereSQL: "datetime(expires_at) <= " + now,
			sizeSQL:  "COALESCE(LENGTH(CAST(token AS TEXT)),0)",
		},
		{
			key: "expired_local_sessions", label: "جلسات الجهاز المنتهية", table: "local_sessions",
			required: []string{"id", "expires_at", "access_token", "refresh_token"},
			wherePG:  "expires_at <= " + now,
			whereSQL: "datetime(expires_at) <= " + now,
			sizeSQL:  "COALESCE(LENGTH(CAST(access_token AS TEXT)),0)+COALESCE(LENGTH(CAST(refresh_token AS TEXT)),0)",
		},
		{
			key: "synced_queue_history", label: "سجل مزامنة قديم اكتمل رفعه", table: "sync_queue",
			required: []string{"id", "payload", "synced_at"},
			wherePG:  "synced_at IS NOT NULL AND synced_at <= " + cutoff90,
			whereSQL: "synced_at IS NOT NULL AND datetime(synced_at) <= " + cutoff90,
			sizeSQL:  "COALESCE(LENGTH(CAST(payload AS TEXT)),0)",
		},
		{
			key: "expired_password_resets", label: "روابط استعادة كلمة مرور منتهية أو مستخدمة", table: "password_reset_tokens",
			required: []string{"id", "expires_at", "used", "token"},
			wherePG:  "used = TRUE OR expires_at <= " + now,
			whereSQL: "COALESCE(used,0) = 1 OR datetime(expires_at) <= " + now,
			sizeSQL:  "COALESCE(LENGTH(CAST(token AS TEXT)),0)",
		},
		{
			key: "old_notifications", label: "إشعارات منتهية أو مقروءة منذ أكثر من 90 يومًا", table: "notifications",
			required: []string{"id", "created_at", "title", "message", "data"},
			sizeSQL:  "COALESCE(LENGTH(CAST(title AS TEXT)),0)+COALESCE(LENGTH(CAST(message AS TEXT)),0)+COALESCE(LENGTH(CAST(data AS TEXT)),0)",
		},
		{
			key: "old_reservations", label: "حجوزات منتهية أو ملغاة منذ أكثر من 90 يومًا", table: "reservations",
			required: []string{"id", "status", "updated_at", "notes"},
			wherePG:  "LOWER(TRIM(COALESCE(status,''))) IN ('expired','cancelled','converted') AND updated_at <= " + cutoff90,
			whereSQL: "LOWER(TRIM(COALESCE(status,''))) IN ('expired','cancelled','converted') AND datetime(updated_at) <= " + cutoff90,
			sizeSQL:  "COALESCE(LENGTH(CAST(notes AS TEXT)),0)",
		},
		{
			key: "orphan_item_specifications", label: "تفاصيل منتجات لا ترتبط بمنتج موجود", table: "item_specification_values",
			required: []string{"id", "inventory_item_id", "value_text", "value_number", "value_boolean"},
			depends:  []string{"inventory_items"},
			wherePG:  "NOT EXISTS (SELECT 1 FROM inventory_items i WHERE i.id = inventory_item_id)",
			whereSQL: "NOT EXISTS (SELECT 1 FROM inventory_items i WHERE i.id = inventory_item_id)",
			sizeSQL:  "COALESCE(LENGTH(CAST(value_text AS TEXT)),0)+COALESCE(LENGTH(CAST(value_number AS TEXT)),0)+COALESCE(LENGTH(CAST(value_boolean AS TEXT)),0)",
		},
	}
}

func cleanupTableExists(ctx context.Context, db cleanupExecutor, sqlite bool, table string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1)`
	if sqlite {
		query = `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=$1)`
	}
	if err := db.GetContext(ctx, &exists, query, table); err != nil {
		return false, err
	}
	return exists, nil
}

func cleanupColumnExists(ctx context.Context, db cleanupExecutor, sqlite bool, table, column string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name=$1 AND column_name=$2)`
	if sqlite {
		query = `SELECT EXISTS (SELECT 1 FROM pragma_table_info('` + table + `') WHERE name=$1)`
		if err := db.GetContext(ctx, &exists, query, column); err != nil {
			return false, err
		}
		return exists, nil
	}
	if err := db.GetContext(ctx, &exists, query, table, column); err != nil {
		return false, err
	}
	return exists, nil
}

func cleanupAvailableSpecs(ctx context.Context, db cleanupExecutor, sqlite bool) ([]cleanupSpec, error) {
	available := make([]cleanupSpec, 0)
	for _, spec := range cleanupSpecs(sqlite) {
		exists, err := cleanupTableExists(ctx, db, sqlite, spec.table)
		if err != nil {
			return nil, fmt.Errorf("check cleanup table %s: %w", spec.table, err)
		}
		if !exists {
			continue
		}
		dependenciesExist := true
		for _, dependency := range spec.depends {
			exists, err := cleanupTableExists(ctx, db, sqlite, dependency)
			if err != nil {
				return nil, fmt.Errorf("check cleanup dependency %s: %w", dependency, err)
			}
			if !exists {
				dependenciesExist = false
				break
			}
		}
		if !dependenciesExist {
			continue
		}
		complete := true
		for _, column := range spec.required {
			exists, err := cleanupColumnExists(ctx, db, sqlite, spec.table, column)
			if err != nil {
				return nil, fmt.Errorf("check cleanup column %s.%s: %w", spec.table, column, err)
			}
			if !exists {
				complete = false
				break
			}
		}
		if complete {
			if spec.key == "old_reservations" {
				hasExpiry, err := cleanupColumnExists(ctx, db, sqlite, spec.table, "expires_at")
				if err != nil {
					return nil, fmt.Errorf("check cleanup column reservations.expires_at: %w", err)
				}
				if hasExpiry {
					spec.wherePG = "(" + spec.wherePG + ") OR expires_at <= NOW() - INTERVAL '90 days'"
					spec.whereSQL = "(" + spec.whereSQL + ") OR datetime(expires_at) <= datetime('now', '-90 days')"
				}
			}
			if spec.key == "old_notifications" {
				var hasExpiry, hasStatus, hasReadFlag bool
				if hasExpiry, err = cleanupColumnExists(ctx, db, sqlite, spec.table, "expires_at"); err != nil {
					return nil, fmt.Errorf("check cleanup column notifications.expires_at: %w", err)
				}
				if hasStatus, err = cleanupColumnExists(ctx, db, sqlite, spec.table, "status"); err != nil {
					return nil, fmt.Errorf("check cleanup column notifications.status: %w", err)
				}
				if hasReadFlag, err = cleanupColumnExists(ctx, db, sqlite, spec.table, "is_read"); err != nil {
					return nil, fmt.Errorf("check cleanup column notifications.is_read: %w", err)
				}
				conditionsPG := make([]string, 0, 2)
				conditionsSQL := make([]string, 0, 2)
				if hasExpiry {
					conditionsPG = append(conditionsPG, "(expires_at IS NOT NULL AND expires_at <= NOW())")
					conditionsSQL = append(conditionsSQL, "(expires_at IS NOT NULL AND datetime(expires_at) <= datetime('now'))")
				}
				if hasStatus {
					conditionsPG = append(conditionsPG, "(LOWER(COALESCE(status,'')) = 'read' AND created_at <= NOW() - INTERVAL '90 days')")
					conditionsSQL = append(conditionsSQL, "(LOWER(COALESCE(status,'')) = 'read' AND datetime(created_at) <= datetime('now', '-90 days'))")
				} else if hasReadFlag {
					conditionsPG = append(conditionsPG, "(is_read IS TRUE AND created_at <= NOW() - INTERVAL '90 days')")
					conditionsSQL = append(conditionsSQL, "(COALESCE(is_read,0) = 1 AND datetime(created_at) <= datetime('now', '-90 days'))")
				}
				if len(conditionsPG) == 0 {
					continue
				}
				spec.wherePG = strings.Join(conditionsPG, " OR ")
				spec.whereSQL = strings.Join(conditionsSQL, " OR ")
			}
			available = append(available, spec)
		}
	}
	return available, nil
}

func cleanupStats(ctx context.Context, db cleanupExecutor, sqlite bool, spec cleanupSpec) (CleanupCategory, error) {
	where := spec.wherePG
	bytesExpr := "COALESCE(SUM(pg_column_size(t)),0)"
	if sqlite {
		where = spec.whereSQL
		bytesExpr = "COALESCE(SUM(" + spec.sizeSQL + "),0)"
	}
	var category CleanupCategory
	category.Key = spec.key
	category.Label = spec.label
	var stats struct {
		Count          int64 `db:"count"`
		EstimatedBytes int64 `db:"estimated_bytes"`
	}
	query := fmt.Sprintf(`SELECT COUNT(*) AS count, %s AS estimated_bytes FROM %s t WHERE %s`, bytesExpr, spec.table, where)
	if err := db.GetContext(ctx, &stats, query); err != nil {
		return category, fmt.Errorf("preview cleanup table %s: %w", spec.table, err)
	}
	category.Count = stats.Count
	category.EstimatedBytes = stats.EstimatedBytes
	return category, nil
}

func storageStats(ctx context.Context, db cleanupExecutor, sqlite bool) (CleanupStorage, error) {
	var storage CleanupStorage
	if sqlite {
		var pageCount, pageSize, freePages int64
		if err := db.GetContext(ctx, &pageCount, `PRAGMA page_count`); err != nil {
			return storage, err
		}
		if err := db.GetContext(ctx, &pageSize, `PRAGMA page_size`); err != nil {
			return storage, err
		}
		if err := db.GetContext(ctx, &freePages, `PRAGMA freelist_count`); err != nil {
			return storage, err
		}
		storage.DatabaseBytes = pageCount * pageSize
		storage.ReusableBytes = freePages * pageSize
		storage.Note = "المساحة القابلة لإعادة الاستخدام من SQLite؛ تقليص حجم الملف يتطلب صيانة منفصلة."
		return storage, nil
	}
	if err := db.GetContext(ctx, &storage.DatabaseBytes, `SELECT pg_database_size(current_database())`); err != nil {
		return storage, err
	}
	storage.Note = "حذف PostgreSQL يجعل الصفحات قابلة لإعادة الاستخدام بعد VACUUM/autovacuum؛ لا يضمن خفض حجم قاعدة البيانات أو حصة Supabase مباشرة."
	return storage, nil
}

func buildCleanupPreview(ctx context.Context, db cleanupExecutor, sqlite bool) (*CleanupPreview, error) {
	specs, err := cleanupAvailableSpecs(ctx, db, sqlite)
	if err != nil {
		return nil, err
	}
	preview := &CleanupPreview{Categories: make([]CleanupCategory, 0, len(specs))}
	for _, spec := range specs {
		category, err := cleanupStats(ctx, db, sqlite, spec)
		if err != nil {
			return nil, err
		}
		preview.Categories = append(preview.Categories, category)
		preview.TotalCount += category.Count
		preview.EstimatedBytes += category.EstimatedBytes
	}
	preview.Storage, err = storageStats(ctx, db, sqlite)
	if err != nil {
		return nil, fmt.Errorf("read database storage status: %w", err)
	}
	return preview, nil
}

func (h *DatabaseHandler) cleanupDatabase(c *gin.Context, localOnly bool) (*sqlx.DB, bool, string, bool) {
	if h.db == nil {
		response.InternalError(c, "database is unavailable")
		return nil, false, "", false
	}
	sqlite := dbutil.IsSQLite(h.db)
	if localOnly && !sqlite {
		c.JSON(http.StatusConflict, gin.H{"error": "local SQLite cleanup is available only through the local database service"})
		return nil, false, "", false
	}
	target := "cloud"
	if sqlite {
		target = "local"
	}
	return h.db, sqlite, target, true
}

func (h *DatabaseHandler) PreviewCleanup(c *gin.Context) {
	h.previewCleanup(c, false)
}

func (h *DatabaseHandler) PreviewLocalCleanup(c *gin.Context) {
	h.previewCleanup(c, true)
}

func (h *DatabaseHandler) previewCleanup(c *gin.Context, localOnly bool) {
	db, sqlite, target, ok := h.cleanupDatabase(c, localOnly)
	if !ok {
		return
	}
	preview, err := buildCleanupPreview(c.Request.Context(), db, sqlite)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	preview.Target = target
	response.OK(c, preview, "Cleanup preview ready")
}

func (h *DatabaseHandler) RunCleanup(c *gin.Context) {
	h.runCleanup(c, false)
}

func (h *DatabaseHandler) RunLocalCleanup(c *gin.Context) {
	h.runCleanup(c, true)
}

func (h *DatabaseHandler) runCleanup(c *gin.Context, localOnly bool) {
	db, sqlite, target, ok := h.cleanupDatabase(c, localOnly)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	specs, err := cleanupAvailableSpecs(ctx, db, sqlite)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("begin cleanup: %v", err))
		return
	}
	defer tx.Rollback()

	const batchSize = 250
	const maxRowsPerCategory = 2000
	var deleted int64
	for _, spec := range specs {
		where := spec.wherePG
		if sqlite {
			where = spec.whereSQL
		}
		for count := 0; count < maxRowsPerCategory; count += batchSize {
			if err := ctx.Err(); err != nil {
				response.InternalError(c, "cleanup canceled before completion")
				return
			}
			var ids []string
			if err := tx.SelectContext(ctx, &ids, tx.Rebind(fmt.Sprintf(`SELECT id FROM %s WHERE %s ORDER BY id LIMIT ?`, spec.table, where)), batchSize); err != nil {
				response.InternalError(c, fmt.Sprintf("load cleanup candidates from %s: %v", spec.table, err))
				return
			}
			if len(ids) == 0 {
				break
			}
			query, args, err := sqlx.In(`DELETE FROM `+spec.table+` WHERE id IN (?)`, ids)
			if err != nil {
				response.InternalError(c, fmt.Sprintf("prepare cleanup for %s: %v", spec.table, err))
				return
			}
			result, err := tx.ExecContext(ctx, tx.Rebind(query), args...)
			if err != nil {
				response.InternalError(c, fmt.Sprintf("clean %s: %v", spec.table, err))
				return
			}
			rows, err := result.RowsAffected()
			if err != nil {
				response.InternalError(c, fmt.Sprintf("count cleaned rows for %s: %v", spec.table, err))
				return
			}
			deleted += rows
			if len(ids) < batchSize || rows == 0 {
				break
			}
		}
	}
	if err := tx.Commit(); err != nil {
		response.InternalError(c, fmt.Sprintf("commit cleanup: %v", err))
		return
	}
	preview, err := buildCleanupPreview(ctx, db, sqlite)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	result := CleanupResult{Target: target, DeletedCount: deleted, Remaining: preview.TotalCount, Storage: preview.Storage}
	message := "تم تنظيف البيانات المنتهية بأمان"
	if result.Remaining > 0 {
		message = "تم تنظيف دفعة من البيانات المنتهية؛ يمكن متابعة التنظيف لإزالة الباقي"
	}
	response.OK(c, result, message)
}
