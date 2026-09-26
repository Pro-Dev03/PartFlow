package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/customers"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/debts"
	"github.com/partflow/smart-store/internal/expenses"
	"github.com/partflow/smart-store/internal/inspections"
	"github.com/partflow/smart-store/internal/inventory"
	"github.com/partflow/smart-store/internal/payments"
	"github.com/partflow/smart-store/internal/products"
	"github.com/partflow/smart-store/internal/purchases"
	"github.com/partflow/smart-store/internal/returns"
	"github.com/partflow/smart-store/internal/sales"
	"github.com/partflow/smart-store/internal/supplierreturns"
	"github.com/partflow/smart-store/internal/suppliers"
	"github.com/partflow/smart-store/pkg/response"
)

type HistoricalCleanupPreview struct {
	Type           string   `json:"type"`
	Label          string   `json:"label"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
	Timezone       string   `json:"timezone"`
	CandidateCount int64    `json:"candidate_count"`
	CandidateIDs   []string `json:"candidate_ids,omitempty"`
	MaxBatch       int      `json:"max_batch"`
	Note           string   `json:"note"`
}

type HistoricalCleanupResult struct {
	Type         string                   `json:"type"`
	Total        int                      `json:"total"`
	Deleted      int                      `json:"deleted"`
	Blocked      int                      `json:"blocked"`
	Failed       int                      `json:"failed"`
	StoppedEarly bool                     `json:"stopped_early"`
	Issues       []HistoricalCleanupIssue `json:"issues,omitempty"`
}

type HistoricalCleanupIssue struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type historicalCleanupSpec struct {
	key               string
	label             string
	table             string
	dateCols          []string
	pendingStatusCols []string
	ownerColumn       string
}

const historicalCleanupMaxBatch = 1000

var historicalCleanupSpecs = map[string]historicalCleanupSpec{
	"inventory_adjustments": {key: "inventory_adjustments", label: "تعديلات كمية المخزون", table: "inventory_movements", dateCols: []string{"created_at"}},
	"sales":                 {key: "sales", label: "فواتير المبيعات", table: "sales", dateCols: []string{"sale_date", "created_at"}},
	"purchases":             {key: "purchases", label: "فواتير المشتريات", table: "purchases", dateCols: []string{"purchase_date", "created_at"}},
	"expenses":              {key: "expenses", label: "المصروفات", table: "expenses", dateCols: []string{"expense_date", "created_at"}},
	"income":                {key: "income", label: "الإيرادات المستقلة", table: "financial_transactions", dateCols: []string{"created_at"}},
	"returns":               {key: "returns", label: "مرتجعات العملاء", table: "returns", dateCols: []string{"return_date", "created_at"}},
	"supplier_returns":      {key: "supplier_returns", label: "مرتجعات الموردين", table: "supplier_returns", dateCols: []string{"return_date", "created_at"}},
	"debts":                 {key: "debts", label: "الديون اليدوية", table: "debts", dateCols: []string{"created_at"}},
	"payments":              {key: "payments", label: "الدفعات القابلة للعكس والحذف", table: "payments", dateCols: []string{"payment_date", "created_at"}, pendingStatusCols: []string{"status", "payment_status"}},
	"products":              {key: "products", label: "المنتجات غير المستخدمة", table: "products", dateCols: []string{"created_at"}},
	"inventory_items":       {key: "inventory_items", label: "قطع المخزون غير المرتبطة بحركة تجارية", table: "inventory_items", dateCols: []string{"created_at"}},
	"held_sales":            {key: "held_sales", label: "المبيعات المعلّقة", table: "held_sales", dateCols: []string{"created_at"}, ownerColumn: "user_id"},
	"inspections":           {key: "inspections", label: "الفحوص المعلّقة المستقلة", table: "inspections", dateCols: []string{"inspection_date", "created_at"}, pendingStatusCols: []string{"result", "status"}},
	"customers":             {key: "customers", label: "العملاء غير المرتبطين بعمليات", table: "customers", dateCols: []string{"created_at"}},
	"suppliers":             {key: "suppliers", label: "الموردون غير المرتبطين بعمليات", table: "suppliers", dateCols: []string{"created_at"}},
}

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

func historicalCleanupRange(rawStart, rawEnd string) (string, string, error) {
	start, err := time.Parse("2006-01-02", strings.TrimSpace(rawStart))
	if err != nil || start.Format("2006-01-02") != strings.TrimSpace(rawStart) {
		return "", "", fmt.Errorf("start_date must use YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", strings.TrimSpace(rawEnd))
	if err != nil || end.Format("2006-01-02") != strings.TrimSpace(rawEnd) {
		return "", "", fmt.Errorf("end_date must use YYYY-MM-DD")
	}
	if end.Before(start) {
		return "", "", fmt.Errorf("end_date must be on or after start_date")
	}
	return start.Format("2006-01-02"), end.AddDate(0, 0, 1).Format("2006-01-02"), nil
}

func historicalCleanupDateColumn(ctx context.Context, db cleanupExecutor, sqlite bool, spec historicalCleanupSpec) (string, bool, error) {
	for _, column := range spec.dateCols {
		exists, err := cleanupColumnExists(ctx, db, sqlite, spec.table, column)
		if err != nil {
			return "", false, err
		}
		if !exists {
			continue
		}
		if sqlite {
			return "store_date(" + column + ")", true, nil
		}
		var dataType string
		if err := db.GetContext(ctx, &dataType, `SELECT data_type FROM information_schema.columns WHERE table_schema=current_schema() AND table_name=$1 AND column_name=$2`, spec.table, column); err != nil {
			return "", false, err
		}
		if dataType == "date" {
			return column, true, nil
		}
		return accounting.PostgresStoreDateExpression(column), true, nil
	}
	return "", false, nil
}

func historicalCleanupStatusFilter(ctx context.Context, db cleanupExecutor, sqlite bool, spec historicalCleanupSpec) (string, error) {
	if spec.key == "inventory_adjustments" {
		return `UPPER(COALESCE(movement_type,'')) = 'ADJUSTMENT' AND ((LOWER(COALESCE(reference_type,'')) = 'product_quantity_adjustment' AND item_id IS NULL) OR (LOWER(COALESCE(reference_type,'')) = 'adjustment' AND item_id IS NOT NULL))`, nil
	}
	// Transaction records in any state are eligible. The per-entity deletion
	// service performs the required reversal and refuses only when source data
	// is genuinely inconsistent or an external settlement cannot be reversed.
	return "", nil
}

func historicalCleanupAdditionalFilter(ctx context.Context, db cleanupExecutor, sqlite bool, spec historicalCleanupSpec) (string, error) {
	if spec.key != "income" {
		return "", nil
	}
	for _, column := range []string{"sale_id", "type"} {
		exists, err := cleanupColumnExists(ctx, db, sqlite, spec.table, column)
		if err != nil {
			return "", fmt.Errorf("inspect income source column %s: %w", column, err)
		}
		if !exists {
			return "", fmt.Errorf("income deletion requires financial_transactions.%s; migrate or repair the financial transaction schema first", column)
		}
	}
	return `sale_id IS NULL AND LOWER(COALESCE(type,'')) IN ('income','revenue')`, nil
}

func quoteCleanupIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

// findHistoricalCleanupReferences finds customer/supplier IDs still referenced
// by any table column with the matching foreign-key convention. This allows
// hard deletion only for truly unused directory records without guessing which
// financial history may be removed.
func findHistoricalCleanupReferences(ctx context.Context, db cleanupExecutor, sqlite bool, ownerTable, referenceColumn string, ids []string) (map[string]struct{}, error) {
	referenced := make(map[string]struct{})
	var tables []string
	if sqlite {
		if err := db.SelectContext(ctx, &tables, `SELECT name FROM sqlite_master WHERE type='table' AND name <> ? ORDER BY name`, ownerTable); err != nil {
			return nil, err
		}
	} else {
		if err := db.SelectContext(ctx, &tables, `SELECT DISTINCT table_name FROM information_schema.columns WHERE table_schema=current_schema() AND table_name <> $1 AND column_name=$2 ORDER BY table_name`, ownerTable, referenceColumn); err != nil {
			return nil, err
		}
	}
	for _, table := range tables {
		if sqlite {
			exists, err := cleanupColumnExists(ctx, db, true, table, referenceColumn)
			if err != nil {
				return nil, err
			}
			if !exists {
				continue
			}
		}
		for start := 0; start < len(ids); start += 400 {
			end := start + 400
			if end > len(ids) {
				end = len(ids)
			}
			query, args, err := sqlx.In(`SELECT DISTINCT CAST(`+quoteCleanupIdentifier(referenceColumn)+` AS TEXT) FROM `+quoteCleanupIdentifier(table)+` WHERE CAST(`+quoteCleanupIdentifier(referenceColumn)+` AS TEXT) IN (?)`, ids[start:end])
			if err != nil {
				return nil, err
			}
			var rows []string
			if err := db.SelectContext(ctx, &rows, db.Rebind(query), args...); err != nil {
				return nil, fmt.Errorf("inspect references in %s: %w", table, err)
			}
			for _, rawID := range rows {
				if parsed, err := uuid.Parse(rawID); err == nil {
					referenced[parsed.String()] = struct{}{}
				} else {
					referenced[strings.ToLower(strings.TrimSpace(rawID))] = struct{}{}
				}
			}
		}
	}
	return referenced, nil
}

func deleteUnreferencedHistoricalDirectories(ctx context.Context, db *sqlx.DB, sqlite bool, spec historicalCleanupSpec, ids []string, dateExpression, filterSQL, startDate, exclusiveEnd string) (int, int, error) {
	if len(ids) == 0 {
		return 0, 0, nil
	}
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("begin %s cleanup: %w", spec.table, err)
	}
	defer tx.Rollback()

	quotedTable := quoteCleanupIdentifier(spec.table)
	if sqlite {
		lockQuery, args, err := sqlx.In(`UPDATE `+quotedTable+` SET id=id WHERE id IN (?)`, ids)
		if err != nil {
			return 0, 0, err
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(lockQuery), args...); err != nil {
			return 0, 0, fmt.Errorf("lock %s cleanup candidates: %w", spec.table, err)
		}
	} else {
		lockQuery, args, err := sqlx.In(`SELECT id FROM `+quotedTable+` WHERE id IN (?) ORDER BY id FOR UPDATE`, ids)
		if err != nil {
			return 0, 0, err
		}
		var lockedIDs []string
		if err := tx.SelectContext(ctx, &lockedIDs, tx.Rebind(lockQuery), args...); err != nil {
			return 0, 0, fmt.Errorf("lock %s cleanup candidates: %w", spec.table, err)
		}
	}

	referenceColumn := "customer_id"
	if spec.key == "suppliers" {
		referenceColumn = "supplier_id"
	}
	referenced, err := findHistoricalCleanupReferences(ctx, tx, sqlite, spec.table, referenceColumn, ids)
	if err != nil {
		return 0, 0, fmt.Errorf("check %s links before cleanup: %w", spec.table, err)
	}
	deletedCount, blockedCount := 0, 0
	for _, id := range ids {
		var stillEligible bool
		checkQuery := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id=? AND %s>=? AND %s<?%s)`, quotedTable, dateExpression, dateExpression, filterSQL)
		if err := tx.GetContext(ctx, &stillEligible, tx.Rebind(checkQuery), id, startDate, exclusiveEnd); err != nil {
			return 0, 0, fmt.Errorf("recheck %s cleanup candidate: %w", spec.table, err)
		}
		if !stillEligible {
			blockedCount++
			continue
		}
		if _, linked := referenced[strings.ToLower(id)]; linked {
			blockedCount++
			continue
		}
		deleteQuery := fmt.Sprintf(`DELETE FROM %s WHERE id=? AND %s>=? AND %s<?%s`, quotedTable, dateExpression, dateExpression, filterSQL)
		result, err := tx.ExecContext(ctx, tx.Rebind(deleteQuery), id, startDate, exclusiveEnd)
		if err != nil {
			return 0, 0, fmt.Errorf("delete unreferenced %s: %w", spec.table, err)
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return 0, 0, fmt.Errorf("count deleted %s: %w", spec.table, err)
		}
		if rows == 1 {
			deletedCount++
		} else {
			blockedCount++
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit %s cleanup: %w", spec.table, err)
	}
	return deletedCount, blockedCount, nil
}

func loadHistoricalCleanupPreview(ctx context.Context, db *sqlx.DB, sqlite bool, cleanupType, rawStart, rawEnd string) (*HistoricalCleanupPreview, error) {
	return loadHistoricalCleanupPreviewForUser(ctx, db, sqlite, cleanupType, rawStart, rawEnd, uuid.Nil)
}

func loadHistoricalCleanupPreviewForUser(ctx context.Context, db *sqlx.DB, sqlite bool, cleanupType, rawStart, rawEnd string, userID uuid.UUID) (*HistoricalCleanupPreview, error) {
	spec, ok := historicalCleanupSpecs[strings.ToLower(strings.TrimSpace(cleanupType))]
	if !ok {
		return nil, fmt.Errorf("unsupported historical cleanup type")
	}
	startDate, endDate, err := historicalCleanupRange(rawStart, rawEnd)
	if err != nil {
		return nil, err
	}
	exists, err := cleanupTableExists(ctx, db, sqlite, spec.table)
	if err != nil {
		return nil, fmt.Errorf("inspect %s table: %w", spec.table, err)
	}
	preview := &HistoricalCleanupPreview{Type: spec.key, Label: spec.label, StartDate: startDate, EndDate: strings.TrimSpace(rawEnd), Timezone: accounting.CurrentStoreTimezone(), MaxBatch: historicalCleanupMaxBatch,
		Note: "يمر كل سجل محدد عبر مسار عكس آثاره المالية والمخزنية ثم حذفه نهائيًا. إذا تعذر العكس بسبب نقص أو تعارض في بيانات قديمة، لن يُحذف السجل وستظهر هويته وسبب التعذر في النتيجة."}
	if spec.key == "expenses" {
		preview.Note = "\u0633\u064a\u062d\u0630\u0641 \u0627\u0644\u0645\u0635\u0631\u0648\u0641 \u0627\u0644\u0645\u062d\u062f\u062f \u0646\u0647\u0627\u0626\u064a\u064b\u0627\u060c \u0648\u0633\u064a\u0632\u064a\u0644 \u0645\u0628\u0644\u063a\u0647 \u0645\u0646 \u062a\u0642\u0627\u0631\u064a\u0631 \u0627\u0644\u0645\u0635\u0631\u0648\u0641\u0627\u062a \u0648\u0635\u0627\u0641\u064a \u0627\u0644\u0631\u0628\u062d."
	}
	if spec.key == "inventory_adjustments" {
		preview.Note += " تُعكس التعديلات على المخزون المجمع وحالة القطعة باستخدام قيم ما قبل التعديل وبعده. إذا كانت حركة لاحقة تعتمد على كمية لا يمكن استعادتها، يظهر سبب التعذر ولا يُحذف التعديل."
	}
	if !exists {
		return preview, nil
	}
	if spec.ownerColumn != "" && userID == uuid.Nil {
		return nil, fmt.Errorf("authenticated user is required to preview %s", spec.table)
	}
	dateExpression, found, err := historicalCleanupDateColumn(ctx, db, sqlite, spec)
	if err != nil {
		return nil, fmt.Errorf("inspect %s date column: %w", spec.table, err)
	}
	if !found {
		return nil, fmt.Errorf("%s has no supported business date column", spec.table)
	}
	statusFilter, err := historicalCleanupStatusFilter(ctx, db, sqlite, spec)
	if err != nil {
		return nil, fmt.Errorf("inspect %s cleanup status: %w", spec.table, err)
	}
	filterSQL := ""
	if statusFilter != "" {
		filterSQL = " AND " + statusFilter
	}
	additionalFilter, err := historicalCleanupAdditionalFilter(ctx, db, sqlite, spec)
	if err != nil {
		return nil, err
	}
	if additionalFilter != "" {
		filterSQL += " AND " + additionalFilter
	}
	ownerFilter := ""
	args := []any{startDate, endDate}
	if spec.ownerColumn != "" {
		ownerFilter = " AND " + quoteCleanupIdentifier(spec.ownerColumn) + " = ?"
		if sqlite {
			args = append(args, userID.String())
		} else {
			args = append(args, userID)
		}
	}
	idsQuery := fmt.Sprintf(`SELECT CAST(id AS TEXT) AS id, COUNT(*) OVER() AS total_count FROM %s WHERE %s >= ? AND %s < ?%s%s ORDER BY %s, id LIMIT ?`, quoteCleanupIdentifier(spec.table), dateExpression, dateExpression, filterSQL, ownerFilter, dateExpression)
	args = append(args, historicalCleanupMaxBatch+1)
	var rows []struct {
		ID    string `db:"id"`
		Count int64  `db:"total_count"`
	}
	if err := db.SelectContext(ctx, &rows, db.Rebind(idsQuery), args...); err != nil {
		return nil, fmt.Errorf("load %s cleanup preview: %w", spec.table, err)
	}
	if len(rows) > 0 {
		preview.CandidateCount = rows[0].Count
		if preview.CandidateCount <= historicalCleanupMaxBatch {
			preview.CandidateIDs = make([]string, 0, len(rows))
			for _, row := range rows {
				preview.CandidateIDs = append(preview.CandidateIDs, row.ID)
			}
		}
	}
	return preview, nil
}

func (h *DatabaseHandler) PreviewHistoricalCleanup(c *gin.Context) {
	if h.db == nil {
		response.InternalError(c, "database is unavailable")
		return
	}
	spec, ok := historicalCleanupSpecs[strings.ToLower(strings.TrimSpace(c.Query("type")))]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported historical cleanup type"})
		return
	}
	var userID uuid.UUID
	if value, exists := c.Get("user_id"); exists {
		userID, _ = value.(uuid.UUID)
	}
	if spec.ownerColumn != "" && userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user is required for this cleanup type"})
		return
	}
	preview, err := loadHistoricalCleanupPreviewForUser(c.Request.Context(), h.db, dbutil.IsSQLite(h.db), spec.key, c.Query("start_date"), c.Query("end_date"), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response.OK(c, preview, "Historical cleanup preview ready")
}

func (h *DatabaseHandler) RunHistoricalCleanup(c *gin.Context) {
	if h.db == nil {
		response.InternalError(c, "database is unavailable")
		return
	}
	var request struct {
		Type         string   `json:"type" binding:"required"`
		StartDate    string   `json:"start_date" binding:"required"`
		EndDate      string   `json:"end_date" binding:"required"`
		CandidateIDs []string `json:"candidate_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type, date range, and preview candidate IDs are required"})
		return
	}
	spec, ok := historicalCleanupSpecs[strings.ToLower(strings.TrimSpace(request.Type))]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported cleanup type"})
		return
	}
	startDate, exclusiveEnd, err := historicalCleanupRange(request.StartDate, request.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(request.CandidateIDs) == 0 || len(request.CandidateIDs) > historicalCleanupMaxBatch {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("provide between 1 and %d previewed records", historicalCleanupMaxBatch)})
		return
	}
	dateExpression, found, err := historicalCleanupDateColumn(c.Request.Context(), h.db, dbutil.IsSQLite(h.db), spec)
	if err != nil || !found {
		if err == nil {
			err = fmt.Errorf("%s has no supported business date column", spec.table)
		}
		response.InternalError(c, err.Error())
		return
	}
	statusFilter, err := historicalCleanupStatusFilter(c.Request.Context(), h.db, dbutil.IsSQLite(h.db), spec)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filterSQL := ""
	if statusFilter != "" {
		filterSQL = " AND " + statusFilter
	}
	additionalFilter, err := historicalCleanupAdditionalFilter(c.Request.Context(), h.db, dbutil.IsSQLite(h.db), spec)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if additionalFilter != "" {
		filterSQL += " AND " + additionalFilter
	}
	result := HistoricalCleanupResult{Type: spec.key, Total: len(request.CandidateIDs)}
	var userID uuid.UUID
	if value, exists := c.Get("user_id"); exists {
		userID, _ = value.(uuid.UUID)
	}
	if spec.ownerColumn != "" && userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user is required for this cleanup type"})
		return
	}
	candidateIDs := make([]uuid.UUID, 0, len(request.CandidateIDs))
	canonicalIDs := make([]string, 0, len(request.CandidateIDs))
	seenIDs := make(map[string]struct{}, len(request.CandidateIDs))
	for _, rawID := range request.CandidateIDs {
		id, err := uuid.Parse(rawID)
		if err != nil || id == uuid.Nil {
			result.Failed++
			result.Issues = append(result.Issues, HistoricalCleanupIssue{ID: rawID, Status: "failed", Reason: "معرّف السجل غير صالح"})
			continue
		}
		if _, exists := seenIDs[id.String()]; exists {
			result.Failed++
			result.Issues = append(result.Issues, HistoricalCleanupIssue{ID: id.String(), Status: "failed", Reason: "تكرر معرّف السجل في الطلب"})
			continue
		}
		seenIDs[id.String()] = struct{}{}
		candidateIDs = append(candidateIDs, id)
		canonicalIDs = append(canonicalIDs, id.String())
	}
	for index, id := range candidateIDs {
		if err := c.Request.Context().Err(); err != nil {
			for _, remainingID := range candidateIDs[index:] {
				recordHistoricalCleanupIssue(&result, remainingID.String(), "failed", err)
			}
			result.StoppedEarly = true
			break
		}
		canonicalID := id.String()
		var stillInRange bool
		ownerFilter := ""
		checkArgs := []any{canonicalID, startDate, exclusiveEnd}
		if spec.ownerColumn != "" {
			ownerFilter = " AND " + quoteCleanupIdentifier(spec.ownerColumn) + " = ?"
			if dbutil.IsSQLite(h.db) {
				checkArgs = append(checkArgs, userID.String())
			} else {
				checkArgs = append(checkArgs, userID)
			}
		}
		checkQuery := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id = ? AND %s >= ? AND %s < ?%s%s)`, quoteCleanupIdentifier(spec.table), dateExpression, dateExpression, filterSQL, ownerFilter)
		if err := h.db.GetContext(c.Request.Context(), &stillInRange, h.db.Rebind(checkQuery), checkArgs...); err != nil {
			recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
			continue
		}
		if !stillInRange {
			recordHistoricalCleanupIssue(&result, canonicalID, "blocked", fmt.Errorf("تغير السجل أو لم يعد ضمن شروط المعاينة"))
			continue
		}
		switch spec.key {
		case "customers":
			if err := customers.NewService(customers.NewRepository(h.db)).DeleteCustomer(c.Request.Context(), id); err != nil {
				recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
			} else {
				result.Deleted++
			}
		case "suppliers":
			if err := suppliers.NewService(suppliers.NewRepository(h.db), h.db).DeleteSupplier(c.Request.Context(), id); err != nil {
				recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
			} else {
				result.Deleted++
			}
		case "sales":
			deleted, deleteErr := sales.NewSmartDeleteService(h.db).SmartDelete(c.Request.Context(), id, userID)
			if deleteErr != nil {
				recordHistoricalCleanupIssue(&result, canonicalID, "failed", deleteErr)
			} else if deleted.Action == "deleted" {
				result.Deleted++
			} else {
				recordHistoricalCleanupIssue(&result, canonicalID, "blocked", fmt.Errorf("%s", deleted.Message))
			}
		case "purchases":
			deleted, deleteErr := purchases.NewSmartDeleteService(h.db).SmartDelete(c.Request.Context(), id, userID)
			if deleteErr != nil {
				recordHistoricalCleanupIssue(&result, canonicalID, "failed", deleteErr)
			} else if deleted.Action == "deleted" || deleted.Action == "reversed" {
				result.Deleted++
			} else {
				recordHistoricalCleanupIssue(&result, canonicalID, "blocked", fmt.Errorf("%s", deleted.Message))
			}
		case "expenses":
			service := expenses.NewService(expenses.NewRepository(h.db))
			if err := service.PermanentlyDeleteExpense(c.Request.Context(), id); err != nil {
				if errors.Is(err, expenses.ErrInvalidExpenseStatus) || errors.Is(err, expenses.ErrExpenseAlreadyArchived) {
					recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
				} else {
					recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
				}
			} else {
				result.Deleted++
			}
		case "returns":
			service := returns.NewService(returns.NewRepository(h.db))
			if err := service.DeleteReturn(c.Request.Context(), id); err != nil {
				if errors.Is(err, returns.ErrInvalidReturnStatus) || errors.Is(err, returns.ErrReturnNotFound) {
					recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
				} else {
					recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
				}
			} else {
				result.Deleted++
			}
		case "supplier_returns":
			service := supplierreturns.NewService(h.db)
			if err := service.Delete(c.Request.Context(), id, false, userID); err != nil {
				recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
			} else {
				result.Deleted++
			}
		case "debts":
			err := debts.NewHandler(h.db).DeleteDebtByID(c.Request.Context(), id)
			if err == nil {
				result.Deleted++
			} else if errors.Is(err, debts.ErrDebtDeletionBlocked) || errors.Is(err, debts.ErrDebtNotFound) {
				recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
			} else {
				recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
			}
		case "payments":
			service := payments.NewService(payments.NewRepository(h.db))
			if err := service.DeletePayment(c.Request.Context(), id, userID); err != nil {
				if errors.Is(err, payments.ErrPaymentCannotBeCancelled) || errors.Is(err, payments.ErrPaymentNotFound) || errors.Is(err, payments.ErrPaymentAllocationHistoryMissing) || errors.Is(err, payments.ErrPaymentAllocationTrackingUnavailable) || errors.Is(err, payments.ErrPaymentHistoryInconsistent) || errors.Is(err, payments.ErrPaymentRequiresProviderRefund) {
					recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
				} else {
					recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
				}
			} else {
				result.Deleted++
			}
		case "products":
			service := products.NewService(products.NewRepository(h.db))
			if err := service.DeleteProduct(c.Request.Context(), id); err != nil {
				if errors.Is(err, products.ErrProductHasHistory) || errors.Is(err, products.ErrProductNotFound) {
					recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
				} else {
					recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
				}
			} else {
				result.Deleted++
			}
		case "inventory_items":
			service := inventory.NewService(inventory.NewRepository(h.db), h.db)
			if err := service.DeleteInventoryItem(c.Request.Context(), id, userID); err != nil {
				if errors.Is(err, inventory.ErrCannotDeleteItemWithHistory) || errors.Is(err, inventory.ErrItemNotFound) {
					recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
				} else {
					recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
				}
			} else {
				result.Deleted++
			}
		case "held_sales":
			if err := sales.NewService(sales.NewRepository(h.db), h.db).DeleteHeldSale(c.Request.Context(), userID, id); err != nil {
				if errors.Is(err, sales.ErrHeldSaleNotFound) {
					recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
				} else {
					recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
				}
			} else {
				result.Deleted++
			}
		case "inspections":
			service := inspections.NewService(inspections.NewRepository(h.db))
			if err := service.DeleteInspection(c.Request.Context(), id); err != nil {
				if errors.Is(err, inspections.ErrCannotDeleteCompletedInspection) || errors.Is(err, inspections.ErrCannotDeleteLinkedInspection) || errors.Is(err, inspections.ErrInspectionNotFound) {
					recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
				} else {
					recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
				}
			} else {
				result.Deleted++
			}
		case "inventory_adjustments":
			service := inventory.NewService(inventory.NewRepository(h.db), h.db)
			if err := service.DeleteProductQuantityAdjustment(c.Request.Context(), id); err != nil {
				if errors.Is(err, inventory.ErrCannotDeleteInventoryAdjustment) || errors.Is(err, inventory.ErrInventoryAdjustmentNotFound) {
					recordHistoricalCleanupIssue(&result, canonicalID, "blocked", err)
				} else {
					recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
				}
			} else {
				result.Deleted++
			}
		case "income":
			service := sales.NewService(sales.NewRepository(h.db), h.db)
			if err := service.DeleteTransaction(c.Request.Context(), id, userID); err != nil {
				recordHistoricalCleanupIssue(&result, canonicalID, "failed", err)
			} else {
				result.Deleted++
			}
		}
	}
	if result.Deleted > 0 {
		dashboard.InvalidateDashboardCacheWithReason("historical_cleanup")
	}
	response.OK(c, result, "Historical cleanup completed")
}

func recordHistoricalCleanupIssue(result *HistoricalCleanupResult, id, status string, err error) {
	reason := "تعذر إكمال الحذف بسبب عدم اتساق البيانات أو تعارضها. راجع سجل الخادم لهذا المعرّف."
	if err != nil && strings.TrimSpace(err.Error()) != "" {
		reason = err.Error()
	}
	if status == "blocked" {
		result.Blocked++
	} else {
		result.Failed++
	}
	result.Issues = append(result.Issues, HistoricalCleanupIssue{ID: id, Status: status, Reason: reason})
}
