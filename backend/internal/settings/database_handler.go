package settings

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/dashboard"
	"github.com/partflow/smart-store/internal/localdb"
)

type DatabaseHandler struct {
	db *sqlx.DB
}

const destructiveResetConfirmation = "DELETE ALL DATA"

func NewDatabaseHandler(db *sqlx.DB) *DatabaseHandler {
	return &DatabaseHandler{db: db}
}

func (h *DatabaseHandler) resetPostgreSQL(c *gin.Context) {
	if c.Query("confirmation_token") != destructiveResetConfirmation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تأكيد التصفير غير صحيح"})
		return
	}

	log.Println("Starting PostgreSQL cleanup...")

	tx, err := h.db.Beginx()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل بدء المعاملة", "details": err.Error()})
		return
	}

	var tables []string
	if err := tx.Select(&tables, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = current_schema()
		  AND table_type = 'BASE TABLE'
		  AND table_name <> 'schema_migrations'
		  AND table_name <> 'users'
		ORDER BY table_name
	`); err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل قراءة جداول قاعدة البيانات", "details": err.Error()})
		return
	}

	if len(tables) == 0 {
		_ = tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"message": "لا توجد بيانات لحذفها", "deleted_tables": 0})
		return
	}

	quotedTables := make([]string, len(tables))
	for i, table := range tables {
		quotedTables[i] = `"` + strings.ReplaceAll(table, `"`, `""`) + `"`
	}
	// Without CASCADE, an unexpected dependency from a protected table makes
	// this transaction fail safely instead of deleting that protected table.
	query := "TRUNCATE TABLE " + strings.Join(quotedTables, ", ") + " RESTART IDENTITY"
	if _, err := tx.Exec(query); err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تصفير بيانات قاعدة البيانات", "details": err.Error()})
		return
	}
	if err := ensureDefaultSettingsWith(tx); err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل إعادة الإعدادات الافتراضية", "details": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تأكيد المعاملة", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "تم حذف بيانات المتجر مع الحفاظ على حسابات المالك والمشتركين",
		"deleted_tables":   len(tables),
		"preserved_tables": []string{"users", "schema_migrations"},
		"target":           "online",
	})
}

func (h *DatabaseHandler) resetSQLite(c *gin.Context) {
	if c.Query("confirmation_token") != destructiveResetConfirmation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تأكيد التصفير غير صحيح"})
		return
	}

	log.Println("Starting SQLite cleanup...")

	db, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فتح قاعدة البيانات المحلية", "details": err.Error()})
		return
	}
	defer db.DB.Close()

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل بدء معاملة تنظيف قاعدة البيانات المحلية", "details": err.Error()})
		return
	}
	defer tx.Rollback()
	if _, err := tx.Exec("PRAGMA defer_foreign_keys = ON"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تهيئة قيود قاعدة البيانات المحلية", "details": err.Error()})
		return
	}

	rows, err := tx.Query(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		  AND name <> 'local_metadata'
		ORDER BY name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل قراءة جداول قاعدة البيانات المحلية", "details": err.Error()})
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل قراءة اسم جدول", "details": err.Error()})
			return
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل إنهاء قراءة جداول قاعدة البيانات المحلية", "details": err.Error()})
		return
	}
	if err := rows.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل إغلاق قراءة جداول قاعدة البيانات المحلية", "details": err.Error()})
		return
	}

	for _, table := range tables {
		// Keep table definitions intact; reset only business rows.
		if table == "users" || table == "local_metadata" || table == "schema_migrations" {
			continue
		}
		if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %q", table)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تصفير بيانات قاعدة البيانات المحلية", "details": err.Error()})
			return
		}
	}
	if err := ensureDefaultSettingsWith(tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore default SQLite settings", "details": err.Error()})
		return
	}
	var hasSequenceTable int
	if err := tx.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sqlite_sequence'").Scan(&hasSequenceTable); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to inspect SQLite sequences", "details": err.Error()})
		return
	}
	if hasSequenceTable > 0 {
		if _, err := tx.Exec("DELETE FROM sqlite_sequence"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset SQLite sequences", "details": err.Error()})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تأكيد تنظيف قاعدة البيانات المحلية", "details": err.Error()})
		return
	}
	if service := dashboard.GetGlobalCacheService(); service != nil {
		service.InvalidateCache()
	}

	if err := localdb.SetMetadata(db.DB, "operating_mode", "offline"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حفظ حالة التشغيل المحلية", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "تم تصفير بيانات التشغيل المحلي مع الاحتفاظ بوضع التشغيل الحالي",
		"deleted_tables": len(tables),
		"target":         "offline",
		"preserved":      []string{"users", "local_metadata", "schema_migrations", "operating_mode"},
	})
}

// DeleteAllData deletes all application data from the database.
func (h *DatabaseHandler) DeleteAllData(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database is unavailable"})
		return
	}
	target := strings.ToLower(strings.TrimSpace(c.Query("target")))
	driver := strings.ToLower(h.db.DriverName())
	if driver == "sqlite" || driver == "sqlite3" {
		if target != "offline" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "local database reset requires target=offline", "code": "LOCAL_DATABASE_TARGET_REQUIRED"})
			return
		}
		h.resetSQLite(c)
		return
	}
	if driver != "pgx" && driver != "postgres" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unsupported database driver"})
		return
	}
	if target == "offline" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cloud service cannot reset a device SQLite database", "code": "LOCAL_DATABASE_NOT_HERE"})
		return
	}
	h.resetPostgreSQL(c)
}

// ApplyMigration applies database migrations
func (h *DatabaseHandler) ApplyMigration(c *gin.Context) {
	log.Println("Applying migration: Add preferred supplier to products...")

	migration := `
-- Add preferred supplier to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS preferred_supplier_id UUID REFERENCES suppliers(id);

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_products_preferred_supplier ON products(preferred_supplier_id);

-- Add comment
COMMENT ON COLUMN products.preferred_supplier_id IS 'المورد المفضل للمنتج';

-- Add missing deleted_at column
ALTER TABLE products ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Add missing cost_price and selling_price columns
ALTER TABLE products ADD COLUMN IF NOT EXISTS cost_price DECIMAL(10,2) DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS selling_price DECIMAL(10,2) DEFAULT 0;

-- Expense reporting and management tables/columns
CREATE TABLE IF NOT EXISTS expense_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(20),
    icon VARCHAR(50),
    budget DECIMAL(10,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES expense_categories(id);
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS title VARCHAR(255);
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'ILS';
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS reference TEXT;
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN DEFAULT false;
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS recurring_period VARCHAR(20);
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id);
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS status VARCHAR(30) DEFAULT 'approved';

-- Add comments
COMMENT ON COLUMN products.cost_price IS 'سعر التكلفة';
COMMENT ON COLUMN products.selling_price IS 'سعر البيع';
`

	_, err := h.db.Exec(migration)
	if err != nil {
		log.Printf("Failed to apply migration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "فشل تطبيق ترحيل قاعدة البيانات",
			"details": err.Error(),
		})
		return
	}

	log.Println("Migration applied successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":   "تم تطبيق ترحيل قاعدة البيانات بنجاح",
		"migration": "add_preferred_supplier_to_products",
	})
}
