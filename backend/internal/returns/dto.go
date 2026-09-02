package returns

import (
	"time"

	"github.com/google/uuid"
)

// ToReturnResponse converts Return to ReturnResponse
func (r *Return) ToReturnResponse(items []ReturnItem, customer *CustomerInfo, sale *SaleInfo) *ReturnResponse {
	return &ReturnResponse{
		Return:     *r,
		Items:      items,
		Customer:   customer,
		Sale:       sale,
		TotalItems: len(items),
	}
}

// ToReturnListItem converts Return to list item format
func (r *Return) ToReturnListItem(itemCount int, customerName string, saleInvoiceNumber string) map[string]interface{} {
	return map[string]interface{}{
		"id":                  r.ID,
		"return_number":       r.ReturnNumber,
		"reference_number":    r.ReferenceNumber,
		"return_date":         r.ReturnDate,
		"customer_name":       customerName,
		"sale_invoice":        saleInvoiceNumber,
		"total_refund_amount": r.TotalRefundAmount,
		"refund_method":       r.RefundMethod,
		"status":              r.Status,
		"return_type":         r.ReturnType,
		"reason":              r.Reason,
		"total_items":         itemCount,
		"created_at":          r.CreatedAt,
	}
}

// CreateReturn creates a Return from request
func CreateReturn(userID uuid.UUID, req *ReturnRequest) *Return {
	returnRecord := &Return{
		ID:                       uuid.New(),
		SaleID:                   uuid.Nil, // Will be set from request
		PurchaseID:               uuid.Nil, // Will be set from request
		CustomerID:               uuid.Nil, // Will be set from sale or request
		ReturnNumber:             generateReturnNumber(),
		ReferenceNumber:          generateReferenceNumber(),
		ReturnDate:               req.ReturnDate,
		ReturnType:               req.ReturnType,
		Status:                   "PENDING",
		TotalRefundAmount:        0,
		RefundMethod:             req.RefundMethod,
		DebtID:                   req.DebtID,
		DebtAdjustment:           0,
		CustomerCredit:           0,
		Reason:                   req.Reason,
		ReasonDetail:             req.ReasonDetail,
		ItemConditionAfterReturn: req.ItemConditionAfterReturn,
		Notes:                    req.Notes,
		InternalNotes:            req.InternalNotes,
		IsWarrantyClaim:          req.IsWarrantyClaim,
		WarrantyID:               req.WarrantyID,
		CreatedBy:                &userID,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	// Set optional fields
	if req.SaleID != nil {
		returnRecord.SaleID = *req.SaleID
	}
	if req.PurchaseID != nil {
		returnRecord.PurchaseID = *req.PurchaseID
	}
	if req.CustomerID != nil {
		returnRecord.CustomerID = *req.CustomerID
	}

	return returnRecord
}

// CreateReturnItem creates a ReturnItem from request
func CreateReturnItem(returnID uuid.UUID, req ReturnItemRequest, unitPrice float64) *ReturnItem {
	totalRefundAmount := req.TotalRefundAmount
	if totalRefundAmount == 0 {
		totalRefundAmount = float64(req.QuantityReturned) * unitPrice
	}

	return &ReturnItem{
		ID:                 uuid.New(),
		ReturnID:           returnID,
		SaleItemID:         req.SaleItemID,
		ProductID:          req.ProductID,
		InventoryItemID:    req.InventoryItemID,
		SerialNumber:       req.SerialNumber,
		Barcode:            req.Barcode,
		QuantityReturned:   req.QuantityReturned,
		OriginalQuantity:   nil, // Will be set from sale item
		UnitPrice:          unitPrice,
		TotalRefundAmount:  totalRefundAmount,
		OriginalCondition:  req.OriginalCondition,
		ReturnedCondition:  req.ReturnedCondition,
		ConditionNotes:     req.ConditionNotes,
		Resolution:         req.Resolution,
		InventoryStatus:    "RETURNED",
		InspectionRequired: req.InspectionRequired,
		OriginalCost:       req.OriginalCost,
		RepairCost:         req.RepairCost,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// ValidateReturnRequest validates return request
func ValidateReturnRequest(req *ReturnRequest) error {
	if (req.SaleID == nil || *req.SaleID == uuid.Nil) && (req.PurchaseID == nil || *req.PurchaseID == uuid.Nil) {
		return ErrSaleNotFound
	}
	if req.Reason == "" {
		return ErrReturnNotFound
	}
	if req.ReturnType != "FULL" && req.ReturnType != "PARTIAL" && req.ReturnType != "QUANTITY_PARTIAL" {
		return ErrInvalidCondition
	}
	if len(req.Items) == 0 {
		return ErrNoItems
	}
	if req.RefundMethod != "CASH" && req.RefundMethod != "CREDIT" &&
		req.RefundMethod != "DEBT_ADJUSTMENT" && req.RefundMethod != "EXCHANGE" &&
		req.RefundMethod != "BANK_TRANSFER" && req.RefundMethod != "STORE_CREDIT" {
		return ErrInvalidRefundMethod
	}
	for _, item := range req.Items {
		if item.SaleItemID == nil && (item.ProductID == nil || *item.ProductID == uuid.Nil) {
			return ErrSaleItemNotFound
		}
		if item.QuantityReturned <= 0 {
			return ErrInvalidQuantity
		}
		if item.ReturnedCondition != "NEW" && item.ReturnedCondition != "USED" &&
			item.ReturnedCondition != "DAMAGED" && item.ReturnedCondition != "DEFECTIVE" &&
			item.ReturnedCondition != "OPEN_BOX" && item.ReturnedCondition != "REFURBISHED" {
			return ErrInvalidCondition
		}
	}
	return nil
}

// generateReturnNumber generates a unique return number
func generateReturnNumber() string {
	return "RET-" + uuid.New().String()[:8]
}

// generateReferenceNumber generates a unique reference number
func generateReferenceNumber() string {
	return "REF-" + uuid.New().String()[:8]
}
