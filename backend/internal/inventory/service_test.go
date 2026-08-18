package inventory

import (
	"testing"

	"github.com/google/uuid"
)

func TestIsValidCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		want      bool
	}{
		{
			name:      "valid new condition",
			condition: ConditionNew,
			want:      true,
		},
		{
			name:      "valid used condition",
			condition: ConditionUsed,
			want:      true,
		},
		{
			name:      "valid refurbished condition",
			condition: ConditionRefurbished,
			want:      true,
		},
		{
			name:      "valid damaged condition",
			condition: ConditionDamaged,
			want:      true,
		},
		{
			name:      "valid for parts condition",
			condition: ConditionForParts,
			want:      true,
		},
		{
			name:      "invalid condition",
			condition: Condition("invalid"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidCondition(tt.condition); got != tt.want {
				t.Errorf("isValidCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidGrade(t *testing.T) {
	tests := []struct {
		name  string
		grade Grade
		want  bool
	}{
		{
			name:  "valid excellent grade",
			grade: GradeExcellent,
			want:  true,
		},
		{
			name:  "valid very good grade",
			grade: GradeVeryGood,
			want:  true,
		},
		{
			name:  "valid good grade",
			grade: GradeGood,
			want:  true,
		},
		{
			name:  "valid fair grade",
			grade: GradeFair,
			want:  true,
		},
		{
			name:  "valid poor grade",
			grade: GradePoor,
			want:  true,
		},
		{
			name:  "invalid grade",
			grade: Grade("invalid"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidGrade(tt.grade); got != tt.want {
				t.Errorf("isValidGrade() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidStatusTransition(t *testing.T) {
	tests := []struct {
		name         string
		currentStatus Status
		newStatus     Status
		want         bool
	}{
		{
			name:         "valid transition: purchased to received",
			currentStatus: StatusPurchased,
			newStatus:     StatusReceived,
			want:         true,
		},
		{
			name:         "valid transition: available to reserved",
			currentStatus: StatusAvailable,
			newStatus:     StatusReserved,
			want:         true,
		},
		{
			name:         "valid transition: reserved to sold",
			currentStatus: StatusReserved,
			newStatus:     StatusSold,
			want:         true,
		},
		{
			name:         "invalid transition: sold to purchased",
			currentStatus: StatusSold,
			newStatus:     StatusPurchased,
			want:         false,
		},
		{
			name:         "invalid transition: available to purchased",
			currentStatus: StatusAvailable,
			newStatus:     StatusPurchased,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidStatusTransition(tt.currentStatus, tt.newStatus); got != tt.want {
				t.Errorf("isValidStatusTransition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateItemCode(t *testing.T) {
	code := generateItemCode()
	if code == "" {
		t.Error("generateItemCode() returned empty string")
	}

	if len(code) < 5 {
		t.Errorf("generateItemCode() = %v, too short", code)
	}
}

func TestGenerateBarcode(t *testing.T) {
	productID := uuid.New()
	barcode := generateBarcode(productID)

	if barcode == "" {
		t.Error("generateBarcode() returned empty string")
	}

	if len(barcode) < 4 {
		t.Errorf("generateBarcode() = %v, too short", barcode)
	}
}

func TestInventoryItemRequest(t *testing.T) {
	productID := uuid.New()
	locationID := uuid.New()
	supplierID := uuid.New()

	req := InventoryItemRequest{
		ProductID:     productID,
		ItemCode:      "ITEM-001",
		Barcode:       "BAR-001",
		SerialNumber:  "SN-001",
		Condition:     ConditionNew,
		Grade:         GradeExcellent,
		PurchaseCost:  100.0,
		SellingPrice:  150.0,
		LocationID:    &locationID,
		SupplierID:    &supplierID,
		Notes:         "Test item",
	}

	if req.ProductID != productID {
		t.Errorf("InventoryItemRequest.ProductID = %v, want %v", req.ProductID, productID)
	}

	if req.Condition != ConditionNew {
		t.Errorf("InventoryItemRequest.Condition = %v, want %v", req.Condition, ConditionNew)
	}

	if req.Grade != GradeExcellent {
		t.Errorf("InventoryItemRequest.Grade = %v, want %v", req.Grade, GradeExcellent)
	}
}

func TestReservationRequest(t *testing.T) {
	itemID := uuid.New()
	customerID := uuid.New()

	req := ReservationRequest{
		ItemID:      itemID,
		CustomerID:  customerID,
		ExpiresIn:   60,
		Notes:       "Test reservation",
	}

	if req.ItemID != itemID {
		t.Errorf("ReservationRequest.ItemID = %v, want %v", req.ItemID, itemID)
	}

	if req.CustomerID != customerID {
		t.Errorf("ReservationRequest.CustomerID = %v, want %v", req.CustomerID, customerID)
	}

	if req.ExpiresIn != 60 {
		t.Errorf("ReservationRequest.ExpiresIn = %v, want 60", req.ExpiresIn)
	}
}

func TestAdjustmentRequest(t *testing.T) {
	itemID := uuid.New()

	req := AdjustmentRequest{
		ItemID:     itemID,
		NewQuantity: 5,
		NewStatus:  "AVAILABLE",
		Reason:     "Stock adjustment",
	}

	if req.ItemID != itemID {
		t.Errorf("AdjustmentRequest.ItemID = %v, want %v", req.ItemID, itemID)
	}

	if req.NewQuantity != 5 {
		t.Errorf("AdjustmentRequest.NewQuantity = %v, want 5", req.NewQuantity)
	}

	if req.Reason != "Stock adjustment" {
		t.Errorf("AdjustmentRequest.Reason = %v, want 'Stock adjustment'", req.Reason)
	}
}

func TestTransferRequest(t *testing.T) {
	itemID := uuid.New()
	fromLocationID := uuid.New()
	toLocationID := uuid.New()

	req := TransferRequest{
		ItemID:        itemID,
		FromLocationID: fromLocationID,
		ToLocationID:  toLocationID,
		Reason:        "Stock transfer",
	}

	if req.ItemID != itemID {
		t.Errorf("TransferRequest.ItemID = %v, want %v", req.ItemID, itemID)
	}

	if req.FromLocationID != fromLocationID {
		t.Errorf("TransferRequest.FromLocationID = %v, want %v", req.FromLocationID, fromLocationID)
	}

	if req.ToLocationID != toLocationID {
		t.Errorf("TransferRequest.ToLocationID = %v, want %v", req.ToLocationID, toLocationID)
	}
}