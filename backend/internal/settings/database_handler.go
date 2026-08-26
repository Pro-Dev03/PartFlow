package settings

import (
	"log"
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type DatabaseHandler struct {
	db *sqlx.DB
}

func NewDatabaseHandler(db *sqlx.DB) *DatabaseHandler {
	return &DatabaseHandler{db: db}
}

// DeleteAllData يحذف جميع البيانات من قاعدة البيانات
func (h *DatabaseHandler) DeleteAllData(c *gin.Context) {
	log.Println("Starting database cleanup...")

	// قائمة الجداول المطلوب حذفها - بالترتيب الصحيح لتجنب مشاكل foreign keys
	// الجداول التابعة أولاً، ثم الجداول الرئيسية
	tables := []string{
		// Child tables first (containing foreign keys)
		"sale_items",
		"purchase_items",
		"audit_logs",
		"payments",
		
		// Parent tables then (referenced by foreign keys)
		"returns",          // بعد حذف payment
		"sales",            // بعد حذف sale_items و payments
		"purchases",        // بعد حذف purchase_items و payments
		"debts",            // بعد حذف payments
		"expenses",         // بعد حذف payments
		"inspections",
		"notifications",
		"automations",
		"inventory",        // بعد حذف products
		"warehouses",
		"customers",        // بعد حذف sales و payments و debts
		"suppliers",        // بعد حذف purchases و payments
		"products",         // بعد حذف inventory و sale_items و purchase_items
		"categories",
		"users",
	}

	tx, err := h.db.Beginx()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "فشل بدء المعاملة",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Transaction started successfully")

	// حذف البيانات من كل جدول
	deletedCount := 0
	for i, table := range tables {
		log.Printf("Processing table %d/%d: %s", i+1, len(tables), table)
		
		query := fmt.Sprintf("DELETE FROM %s", table)
		log.Printf("Executing query: %s", query)
		
		result, err := tx.Exec(query)
		if err != nil {
			log.Printf("Error deleting from table %s: %v", table, err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("فشل حذف البيانات من جدول: %s", table),
				"details": err.Error(),
			})
			return
		}
		
		rowsAffected, _ := result.RowsAffected()
		deletedCount += int(rowsAffected)
		log.Printf("Deleted %d rows from %s", rowsAffected, table)
	}

	log.Printf("Successfully deleted %d total rows", deletedCount)

	// ملاحظة: لا نحتاج لإعادة تعيين sequences لأن المشروع يستخدم UUIDs

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "فشل تأكيد المعاملة",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Transaction committed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "تم حذف جميع البيانات بنجاح",
		"deleted_tables": len(tables),
		"deleted_rows": deletedCount,
	})
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

-- Add comments
COMMENT ON COLUMN products.cost_price IS 'سعر التكلفة';
COMMENT ON COLUMN products.selling_price IS 'سعر البيع';
`

	_, err := h.db.Exec(migration)
	if err != nil {
		log.Printf("Failed to apply migration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "فشل تطبيق الـ migration",
			"details": err.Error(),
		})
		return
	}

	log.Println("Migration applied successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "تم تطبيق الـ migration بنجاح",
		"migration": "add_preferred_supplier_to_products",
	})
}