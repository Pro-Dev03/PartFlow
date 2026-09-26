package parttypes

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestDeletePartTypeDetachesInventoryAndRemovesSpecificationLinksSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "part-type-delete.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	partTypeID, specificationID, productID, itemID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, row := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO part_types (id,name_ar,name_en,created_at,updated_at) VALUES (?,?,?, ?,?)`, []any{partTypeID, "نوع اختبار", "Test type", now, now}},
		{`INSERT INTO part_specifications (id,name_ar,name_en,data_type,options,is_required,created_at) VALUES (?,?,?,'text','[]',0,?)`, []any{specificationID, "تفصيل", "Detail", now}},
		{`INSERT INTO type_specifications (id,part_type_id,specification_id,sort_order,created_at) VALUES (?,?,?,0,?)`, []any{uuid.New(), partTypeID, specificationID, now}},
		{`INSERT INTO products (id,sku,name,cost_price,selling_price,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`, []any{productID, "PART-TYPE-PRODUCT", "Existing product", 1, 2, now, now}},
		{`INSERT INTO inventory_items (id,product_id,part_type_id,item_code,status,created_at,updated_at) VALUES (?,?,?,?,'AVAILABLE',?,?)`, []any{itemID, productID, partTypeID, "PART-TYPE-ITEM", now, now}},
	} {
		if _, err := db.ExecContext(context.Background(), row.query, row.args...); err != nil {
			t.Fatal(err)
		}
	}
	if err := NewRepository(db).DeletePartType(context.Background(), partTypeID); err != nil {
		t.Fatalf("delete part type: %v", err)
	}
	var typeCount, linkCount int
	var itemPartType sql.NullString
	if err := db.Get(&typeCount, `SELECT COUNT(*) FROM part_types WHERE id=?`, partTypeID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&linkCount, `SELECT COUNT(*) FROM type_specifications WHERE part_type_id=?`, partTypeID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&itemPartType, `SELECT part_type_id FROM inventory_items WHERE id=?`, itemID); err != nil {
		t.Fatal(err)
	}
	if typeCount != 0 || linkCount != 0 || itemPartType.Valid {
		t.Fatalf("delete left type/links/inventory reference=%d/%d/%v; want 0/0/NULL", typeCount, linkCount, itemPartType)
	}
}
