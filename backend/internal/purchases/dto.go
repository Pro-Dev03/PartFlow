package purchases

import (
	"time"

	"github.com/google/uuid"
)

// ToPurchaseResponse converts Purchase to PurchaseResponse
func (p *Purchase) ToPurchaseResponse(items []PurchaseItem, supplier *SupplierInfo) *PurchaseResponse {
	return &PurchaseResponse{
		Purchase:   *p,
		Items:      items,
		Supplier:   supplier,
		TotalItems: len(items),
		Remaining:  p.TotalAmount - p.PaidAmount,
	}
}

// ToPurchaseListItem converts Purchase to list item format
func (p *Purchase) ToPurchaseListItem(itemCount int, supplierName string) map[string]interface{} {
	return map[string]interface{}{
		"id":             p.ID,
		"invoice_number": p.InvoiceNumber,
		"purchase_date":  p.PurchaseDate,
		"total_amount":   p.TotalAmount,
		"paid_amount":    p.PaidAmount,
		"remaining":      p.TotalAmount - p.PaidAmount,
		"status":         p.Status,
		"supplier_name":  supplierName,
		"total_items":    itemCount,
		"created_at":     p.CreatedAt,
	}
}

// CreatePurchaseItem creates a PurchaseItem from request
func CreatePurchaseItem(purchaseID uuid.UUID, req PurchaseItemRequest) *PurchaseItem {
	totalCost := float64(req.Quantity) * req.UnitCost
	return &PurchaseItem{
		ID:          uuid.New(),
		PurchaseID:  purchaseID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		UnitCost:    req.UnitCost,
		TotalCost:   totalCost,
		SerialNumber: req.SerialNumber,
		Condition:   req.Condition,
		LocationID:  req.LocationID,
		Notes:       req.Notes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ValidatePurchaseRequest validates purchase request
func ValidatePurchaseRequest(req *PurchaseRequest) error {
	if req.SupplierID == uuid.Nil {
		return ErrSupplierNotFound
	}
	if req.InvoiceNumber == "" {
		return ErrPurchaseNotFound
	}
	if len(req.Items) == 0 {
		return ErrNoItems
	}
	for _, item := range req.Items {
		if item.ProductID == uuid.Nil {
			return ErrProductNotFound
		}
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.UnitCost < 0 {
			return ErrInvalidCost
		}
		if item.Condition != "new" && item.Condition != "used" && item.Condition != "refurbished" {
			return ErrInvalidCondition
		}
	}
	return nil
}
