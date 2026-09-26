package inspections

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func inspectionWorkflowDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	tables := []string{
		`CREATE TABLE inspections (id TEXT PRIMARY KEY, product_id TEXT, inventory_item_id TEXT, inspector_id TEXT, inspection_date TEXT, result TEXT, condition TEXT, grade TEXT, notes TEXT, images TEXT, test_results TEXT, acquisition_item_id TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE inspection_items (id TEXT PRIMARY KEY, inspection_id TEXT NOT NULL, item_id TEXT, checkpoint_name TEXT, status TEXT, notes TEXT, images TEXT, created_at TEXT)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, status TEXT, updated_at TEXT)`,
		`CREATE TABLE acquisitions (id TEXT PRIMARY KEY, status TEXT, updated_at TEXT)`,
		`CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, acquisition_id TEXT, inventory_item_id TEXT, inspection_id TEXT, inspection_status TEXT, item_status TEXT, updated_at TEXT)`,
		`CREATE TABLE inspection_workflow_snapshots (inspection_id TEXT PRIMARY KEY, inventory_item_id TEXT, inventory_status_before TEXT, acquisition_item_id TEXT, acquisition_inventory_item_id_before TEXT, acquisition_inspection_id_before TEXT, acquisition_inspection_status_before TEXT, acquisition_item_status_before TEXT, acquisition_id TEXT, acquisition_status_before TEXT, captured_at TEXT DEFAULT CURRENT_TIMESTAMP)`,
	}
	for _, statement := range tables {
		if _, err := db.Exec(statement); err != nil {
			db.Close()
			t.Fatalf("create inspection test schema: %v", err)
		}
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestDeleteCompletedInspectionRestoresWorkflowAtomicallySQLite(t *testing.T) {
	ctx := context.Background()
	db := inspectionWorkflowDB(t)
	now := time.Now().UTC()
	acquisitionID, acquisitionItemID, inventoryItemID := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO acquisitions (id,status) VALUES (?, 'draft')`, acquisitionID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO acquisition_items (id,acquisition_id,inventory_item_id,item_status) VALUES (?,?,?,'available')`, acquisitionItemID, acquisitionID, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id,status) VALUES (?, 'AVAILABLE')`, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	inspection := &Inspection{
		ID: uuid.New(), ProductID: uuidPtr(uuid.New()), InventoryItemID: &inventoryItemID,
		AcquisitionItemID: &acquisitionItemID, InspectedBy: uuid.New(), InspectionDate: now,
		Status: "pending", Photos: []string{}, CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.CreateInspectionWithWorkflow(ctx, inspection); err != nil {
		t.Fatalf("create linked inspection: %v", err)
	}
	if err := db.Get(new(string), `SELECT status FROM inventory_items WHERE id=? AND status='INSPECTION'`, inventoryItemID); err != nil {
		t.Fatalf("inspection did not move inventory into inspection: %v", err)
	}
	inspection.Status = "passed"
	if err := repo.UpdateInspectionWithWorkflow(ctx, inspection); err != nil {
		t.Fatalf("complete inspection: %v", err)
	}
	if err := repo.DeleteInspection(ctx, inspection.ID); err != nil {
		t.Fatalf("delete completed inspection: %v", err)
	}
	var inventoryStatus, itemStatus, acquisitionStatus string
	var inspectionLink, inspectionStatus sql.NullString
	if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, inventoryItemID).Scan(&inventoryStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT item_status, inspection_id, inspection_status FROM acquisition_items WHERE id=?`, acquisitionItemID).Scan(&itemStatus, &inspectionLink, &inspectionStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT status FROM acquisitions WHERE id=?`, acquisitionID).Scan(&acquisitionStatus); err != nil {
		t.Fatal(err)
	}
	if inventoryStatus != "AVAILABLE" || itemStatus != "available" || inspectionLink.Valid || inspectionStatus.Valid || acquisitionStatus != "draft" {
		t.Fatalf("workflow was not restored: inventory=%s item=%s inspection_id=%v inspection_status=%v acquisition=%s", inventoryStatus, itemStatus, inspectionLink, inspectionStatus, acquisitionStatus)
	}
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM inspections WHERE id=?`, inspection.ID); err != nil || count != 0 {
		t.Fatalf("inspection remains after delete: count=%d err=%v", count, err)
	}
}

func TestDeleteInspectionPreservesLaterInventoryStatusSQLite(t *testing.T) {
	ctx := context.Background()
	db := inspectionWorkflowDB(t)
	now := time.Now().UTC()
	acquisitionID, acquisitionItemID, inventoryItemID := uuid.New(), uuid.New(), uuid.New()
	_, _ = db.Exec(`INSERT INTO acquisitions (id,status) VALUES (?, 'draft')`, acquisitionID)
	_, _ = db.Exec(`INSERT INTO acquisition_items (id,acquisition_id,inventory_item_id,item_status) VALUES (?,?,?,'available')`, acquisitionItemID, acquisitionID, inventoryItemID)
	_, _ = db.Exec(`INSERT INTO inventory_items (id,status) VALUES (?, 'AVAILABLE')`, inventoryItemID)
	repo := NewRepository(db)
	inspection := &Inspection{ID: uuid.New(), ProductID: uuidPtr(uuid.New()), InventoryItemID: &inventoryItemID, AcquisitionItemID: &acquisitionItemID, InspectedBy: uuid.New(), InspectionDate: now, Status: "pending", Photos: []string{}, CreatedAt: now, UpdatedAt: now}
	if err := repo.CreateInspectionWithWorkflow(ctx, inspection); err != nil {
		t.Fatal(err)
	}
	inspection.Status = "passed"
	if err := repo.UpdateInspectionWithWorkflow(ctx, inspection); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE inventory_items SET status='SOLD' WHERE id=?`, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteInspection(ctx, inspection.ID); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := db.Get(&status, `SELECT status FROM inventory_items WHERE id=?`, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if status != "SOLD" {
		t.Fatalf("delete overwrote later inventory state: got %s", status)
	}
}

func uuidPtr(value uuid.UUID) *uuid.UUID { return &value }
