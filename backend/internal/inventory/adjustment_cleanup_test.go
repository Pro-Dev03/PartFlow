package inventory

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestDeleteProductQuantityAdjustmentReversesItsDeltaWithLaterMovementsSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "inventory-adjustment-cleanup.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatalf("open local database: %v", err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	ctx := context.Background()
	service := NewService(NewRepository(db), db)
	now := time.Now().UTC()

	safeProduct := uuid.New()
	blockedProduct := uuid.New()
	negativeProduct := uuid.New()
	for i, productID := range []uuid.UUID{safeProduct, blockedProduct, negativeProduct} {
		if _, err := db.Exec(`INSERT INTO products (id,sku,name,cost_price,selling_price,is_active,created_at,updated_at) VALUES (?,?,?,0,10,1,?,?)`, productID, uuid.NewString(), "Adjustment cleanup test", now, now); err != nil {
			t.Fatalf("insert product %d: %v", i, err)
		}
	}
	for _, row := range []struct {
		productID uuid.UUID
		quantity  int
	}{{safeProduct, 5}, {blockedProduct, 5}, {negativeProduct, 1}} {
		if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES (?,?,?,?,?)`, uuid.New(), row.productID, row.quantity, now, now); err != nil {
			t.Fatalf("insert inventory: %v", err)
		}
	}

	insertMovement := func(id, productID uuid.UUID, quantity, before, after int, referenceType string, itemID *uuid.UUID, createdAt time.Time) {
		t.Helper()
		if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,reason,created_by,created_at) VALUES (?,?,?,'ADJUSTMENT',?,?,?,?,?,?,?,?)`, id, itemID, productID, quantity, before, after, referenceType, productID, "test adjustment", uuid.New(), createdAt.Format(time.RFC3339Nano)); err != nil {
			t.Fatalf("insert movement: %v", err)
		}
	}

	safeID := uuid.New()
	insertMovement(safeID, safeProduct, 3, 2, 5, "product_quantity_adjustment", nil, now.Add(-time.Hour))
	if err := service.DeleteProductQuantityAdjustment(ctx, safeID); err != nil {
		t.Fatalf("delete safe adjustment: %v", err)
	}
	var safeQuantity, safeMovementCount int
	if err := db.Get(&safeQuantity, `SELECT quantity FROM inventory WHERE product_id=?`, safeProduct); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&safeMovementCount, `SELECT COUNT(*) FROM inventory_movements WHERE id=?`, safeID); err != nil {
		t.Fatal(err)
	}
	if safeQuantity != 2 || safeMovementCount != 0 {
		t.Fatalf("safe adjustment deletion left quantity=%d movement_count=%d; want 2 and 0", safeQuantity, safeMovementCount)
	}

	blockedID := uuid.New()
	insertMovement(blockedID, blockedProduct, 3, 2, 5, "product_quantity_adjustment", nil, now.Add(-time.Hour))
	insertMovement(uuid.New(), blockedProduct, 2, 5, 7, "sale", nil, now)
	if _, err := db.Exec(`UPDATE inventory SET quantity=7 WHERE product_id=?`, blockedProduct); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteProductQuantityAdjustment(ctx, blockedID); err != nil {
		t.Fatalf("reverse earlier adjustment while retaining later movement: %v", err)
	}
	var blockedQuantity, blockedMovementCount int
	if err := db.Get(&blockedQuantity, `SELECT quantity FROM inventory WHERE product_id=?`, blockedProduct); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&blockedMovementCount, `SELECT COUNT(*) FROM inventory_movements WHERE product_id=?`, blockedProduct); err != nil {
		t.Fatal(err)
	}
	if blockedQuantity != 4 || blockedMovementCount != 1 {
		t.Fatalf("adjustment reversal left quantity=%d movements=%d; want 4 and 1", blockedQuantity, blockedMovementCount)
	}

	// Removing a positive adjustment that later sales consumed must still be
	// possible. The negative balance exposes the resulting historical stock
	// deficit while preserving the sale and rebasing its movement snapshots.
	negativeAdjustmentID := uuid.New()
	negativeSaleID := uuid.New()
	negativeSaleReferenceID := uuid.New()
	insertMovement(negativeAdjustmentID, negativeProduct, 3, 2, 5, "product_quantity_adjustment", nil, now.Add(-2*time.Hour))
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,reason,created_by,created_at) VALUES (?,NULL,?,'SALE',-4,5,1,'sale',?,'historical sale',?,?)`, negativeSaleID, negativeProduct, negativeSaleReferenceID, uuid.New(), now.Add(-time.Hour).Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("insert later sale movement: %v", err)
	}
	if err := service.DeleteProductQuantityAdjustment(ctx, negativeAdjustmentID); err != nil {
		t.Fatalf("delete adjustment consumed by a later sale: %v", err)
	}
	var negativeQuantity, remainingMovements, rebasedBefore, rebasedAfter int
	if err := db.Get(&negativeQuantity, `SELECT quantity FROM inventory WHERE product_id=?`, negativeProduct); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&remainingMovements, `SELECT COUNT(*) FROM inventory_movements WHERE product_id=?`, negativeProduct); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT before_quantity,after_quantity FROM inventory_movements WHERE reference_type='sale' AND reference_id=?`, negativeSaleReferenceID.String()).Scan(&rebasedBefore, &rebasedAfter); err != nil {
		t.Fatal(err)
	}
	if negativeQuantity != -2 || remainingMovements != 1 || rebasedBefore != 2 || rebasedAfter != -2 {
		t.Fatalf("consumed adjustment reversal left quantity=%d movements=%d sale=%d->%d; want -2/1/2->-2", negativeQuantity, remainingMovements, rebasedBefore, rebasedAfter)
	}
	negativeSnapshotAdjustmentID := uuid.New()
	insertMovement(negativeSnapshotAdjustmentID, negativeProduct, 1, -2, -1, "product_quantity_adjustment", nil, now)
	if _, err := db.Exec(`UPDATE inventory SET quantity=-1 WHERE product_id=?`, negativeProduct); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteProductQuantityAdjustment(ctx, negativeSnapshotAdjustmentID); err != nil {
		t.Fatalf("delete adjustment with negative historical snapshots: %v", err)
	}
	if err := db.Get(&negativeQuantity, `SELECT quantity FROM inventory WHERE product_id=?`, negativeProduct); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&remainingMovements, `SELECT COUNT(*) FROM inventory_movements WHERE product_id=?`, negativeProduct); err != nil {
		t.Fatal(err)
	}
	if negativeQuantity != -2 || remainingMovements != 1 {
		t.Fatalf("negative-snapshot adjustment deletion left quantity=%d movements=%d; want -2/1", negativeQuantity, remainingMovements)
	}
}

func TestDeleteItemAdjustmentReversesStoredStatusSnapshotSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "inventory-item-adjustment.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	ctx := context.Background()
	service := NewService(NewRepository(db), db)
	productID, itemID := uuid.New(), uuid.New()
	now := time.Now().UTC()
	if _, err := db.Exec(`INSERT INTO products (id,sku,name,cost_price,selling_price,is_active,created_at,updated_at) VALUES (?,?,?,0,10,1,?,?)`, productID, uuid.NewString(), "Item status reversal", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES (?,?,?,?,?)`, uuid.New(), productID, 5, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id,item_code,status,created_at,updated_at) VALUES (?,?,?,'AVAILABLE',?,?)`, itemID, productID, uuid.NewString(), now, now); err != nil {
		t.Fatal(err)
	}
	damaged := "DAMAGED"
	if err := service.AdjustInventory(ctx, &AdjustmentRequest{ItemID: itemID, NewQuantity: 6, NewStatus: &damaged}, uuid.Nil); err != nil {
		t.Fatalf("create item adjustment: %v", err)
	}
	var movementID string
	if err := db.Get(&movementID, `SELECT id FROM inventory_movements WHERE item_id=? AND reference_type='adjustment' ORDER BY created_at DESC LIMIT 1`, itemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE inventory SET quantity=8 WHERE product_id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE inventory_items SET status='AVAILABLE' WHERE id=?`, itemID); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteProductQuantityAdjustment(ctx, uuid.MustParse(movementID)); err != nil {
		t.Fatalf("delete item adjustment after later quantity/status changes: %v", err)
	}
	var quantity int
	var status string
	if err := db.Get(&quantity, `SELECT quantity FROM inventory WHERE product_id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&status, `SELECT status FROM inventory_items WHERE id=?`, itemID); err != nil {
		t.Fatal(err)
	}
	if quantity != 7 || status != "AVAILABLE" {
		t.Fatalf("reversed adjustment left quantity=%d status=%s; want 7/AVAILABLE", quantity, status)
	}
}
