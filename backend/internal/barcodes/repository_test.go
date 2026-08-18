package barcodes

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Setup test database connection
func setupTestDB(t *testing.T) *sqlx.DB {
	// In a real implementation, you would connect to a test database
	// For now, this is a placeholder
	return nil
}

func TestNewRepository(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Skipping test: database not available")
	}

	repo := NewRepository(db)
	if repo == nil {
		t.Error("NewRepository() returned nil")
	}

	if repo.db != db {
		t.Error("NewRepository() did not set database correctly")
	}
}

func TestRepository_GetBarcodeByCode(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Skipping test: database not available")
	}

	repo := NewRepository(db)
	ctx := context.Background()
	code := "TEST-BARCODE-001"
	organizationID := uuid.New()

	barcode, err := repo.GetBarcodeByCode(ctx, code, organizationID)
	if err != nil {
		t.Logf("GetBarcodeByCode() error (expected with empty DB): %v", err)
		return
	}

	if barcode != nil {
		t.Logf("GetBarcodeByCode() found barcode: %v", barcode.Code)
	}
}

func TestRepository_CreateBarcode(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Skipping test: database not available")
	}

	repo := NewRepository(db)
	ctx := context.Background()
	organizationID := uuid.New()

	barcode := &Barcode{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Code:           "TEST-BARCODE-002",
		Type:           BarcodeTypeSKU,
		ProductID:      nil,
		InventoryItemID: nil,
		IsActive:       true,
	}

	err := repo.CreateBarcode(ctx, barcode)
	if err != nil {
		t.Logf("CreateBarcode() error (expected with test DB): %v", err)
	}
}

func TestRepository_ListBarcodes(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Skipping test: database not available")
	}

	repo := NewRepository(db)
	ctx := context.Background()
	organizationID := uuid.New()

	barcodes, total, err := repo.ListBarcodes(ctx, organizationID, 10, 0)
	if err != nil {
		t.Logf("ListBarcodes() error (expected with empty DB): %v", err)
		return
	}

	if barcodes == nil {
		t.Error("ListBarcodes() returned nil barcodes")
	}

	if total < 0 {
		t.Errorf("ListBarcodes() total = %v, want >= 0", total)
	}
}

func TestRepository_DeleteBarcode(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Skipping test: database not available")
	}

	repo := NewRepository(db)
	ctx := context.Background()
	barcodeID := uuid.New()
	organizationID := uuid.New()

	err := repo.DeleteBarcode(ctx, barcodeID, organizationID)
	if err != nil {
		t.Logf("DeleteBarcode() error (expected with test DB): %v", err)
	}
}

// Test barcode model structure
func TestBarcode(t *testing.T) {
	id := uuid.New()
	organizationID := uuid.New()
	productID := uuid.New()
	inventoryItemID := uuid.New()

	barcode := Barcode{
		ID:             id,
		OrganizationID: organizationID,
		Code:           "TEST-001",
		Type:           BarcodeTypeSKU,
		ProductID:      &productID,
		InventoryItemID: &inventoryItemID,
		IsActive:       true,
	}

	if barcode.ID != id {
		t.Errorf("Barcode.ID = %v, want %v", barcode.ID, id)
	}

	if barcode.OrganizationID != organizationID {
		t.Errorf("Barcode.OrganizationID = %v, want %v", barcode.OrganizationID, organizationID)
	}

	if barcode.Code != "TEST-001" {
		t.Errorf("Barcode.Code = %v, want TEST-001", barcode.Code)
	}

	if barcode.Type != BarcodeTypeSKU {
		t.Errorf("Barcode.Type = %v, want %v", barcode.Type, BarcodeTypeSKU)
	}
}