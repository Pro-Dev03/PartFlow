package settings

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

type DatabaseHandler struct {
	db *sqlx.DB
}

func NewDatabaseHandler(db *sqlx.DB) *DatabaseHandler {
	return &DatabaseHandler{db: db}
}

func resolveResetTarget(target, currentMode string) string {
	target = strings.TrimSpace(strings.ToLower(target))
	if target == "online" || target == "offline" {
		return target
	}
	if currentMode == "online" || currentMode == "offline" {
		return currentMode
	}
	return "online"
}

func getOperatingModeFromLocalDB() (string, error) {
	db, err := localdb.Open()
	if err != nil {
		return "online", err
	}
	defer db.DB.Close()

	var mode string
	err = db.DB.QueryRow("SELECT value FROM local_metadata WHERE key = 'operating_mode' LIMIT 1").Scan(&mode)
	if err == sql.ErrNoRows {
		return "online", nil
	}
	if err != nil {
		return "online", err
	}
	return strings.TrimSpace(strings.ToLower(mode)), nil
}

func (h *DatabaseHandler) resetPostgreSQL(c *gin.Context) {
	if c.Query("confirmation") != "احذف جميع البيانات" {
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
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		  AND table_name <> 'schema_migrations'
		  AND table_name <> 'users'
		  AND table_name <> 'settings'
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
	query := "TRUNCATE TABLE " + strings.Join(quotedTables, ", ") + " CASCADE"
	if _, err := tx.Exec(query); err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تصفير بيانات قاعدة البيانات", "details": err.Error()})
		return
	}

	var deletedUsers int64
	result, err := tx.Exec(`DELETE FROM users WHERE email IS DISTINCT FROM 'owner@partflow.com'`)
	if err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تنظيف حسابات المستخدمين", "details": err.Error()})
		return
	}
	deletedUsers, _ = result.RowsAffected()

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تأكيد المعاملة", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "تم تصفير بيانات النظام مع الحفاظ على حساب المالك والإعدادات",
		"deleted_tables":   len(tables),
		"deleted_users":    deletedUsers,
		"preserved_tables": []string{"owner user", "settings", "schema_migrations"},
		"target":           "online",
	})
}

func (h *DatabaseHandler) resetSQLite(c *gin.Context) {
	if c.Query("confirmation") != "احذف جميع البيانات" {
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

	rows, err := db.DB.Query(`
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

	for _, table := range tables {
		// Keep table definitions intact; reset only business rows.
		if table == "users" || table == "local_sessions" || table == "refresh_tokens" || table == "settings" || table == "schema_migrations" {
			continue
		}
		if _, err := db.DB.Exec(fmt.Sprintf("DELETE FROM %q", table)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تصفير بيانات قاعدة البيانات المحلية", "details": err.Error()})
			return
		}
	}

	if err := localdb.SetMetadata(db.DB, "operating_mode", "offline"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حفظ حالة التشغيل المحلية", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "تم تصفير بيانات التشغيل المحلي مع الاحتفاظ بوضع التشغيل الحالي",
		"deleted_tables": len(tables),
		"target":         "offline",
		"preserved":      []string{"local_metadata", "operating_mode"},
	})
}

// DeleteAllData يحذف جميع البيانات من قاعدة البيانات
func (h *DatabaseHandler) DeleteAllData(c *gin.Context) {
	// The embedded desktop backend owns SQLite. A client must never be able to
	// select an "online" target and accidentally clear cloud data (or use the
	// old operating-mode switch as a destructive escape hatch).
	if h.db != nil && (strings.EqualFold(h.db.DriverName(), "sqlite") || strings.EqualFold(h.db.DriverName(), "sqlite3")) {
		h.resetSQLite(c)
		return
	}

	currentMode, err := getOperatingModeFromLocalDB()
	if err != nil {
		log.Printf("Unable to read current operating mode: %v", err)
	}

	target := resolveResetTarget(c.Query("target"), c.Query("mode"))
	if target == "" {
		target = resolveResetTarget("", currentMode)
	}
	if target == "offline" {
		h.resetSQLite(c)
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
			"error":   "فشل تطبيق الـ migration",
			"details": err.Error(),
		})
		return
	}

	log.Println("Migration applied successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":   "تم تطبيق الـ migration بنجاح",
		"migration": "add_preferred_supplier_to_products",
	})
}
