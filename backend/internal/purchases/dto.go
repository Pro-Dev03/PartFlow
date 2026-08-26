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

// ToAPIMap converts Purchase to a map with created_by for API compatibility
func (p *Purchase) ToAPIMap() map[string]interface{} {
	result := map[string]interface{}{
		"id":                    p.ID,
		"supplier_id":           p.SupplierID,
		"invoice_number":        p.InvoiceNumber,
		"purchase_date":         p.PurchaseDate,
		"expected_delivery_date": p.ExpectedDeliveryDate,
		"total_amount":          p.TotalAmount,
		"paid_amount":           p.PaidAmount,
		"status":                p.Status,
		"notes":                 p.Notes,
		"created_at":            p.CreatedAt,
		"updated_at":            p.UpdatedAt,
	}

	// Only include created_by if UserID is not nil
	if p.UserID != nil {
		result["created_by"] = *p.UserID
	}

	return result
}

// PurchaseListItem represents a purchase in list view
type PurchaseListItem struct {
	ID             string     `json:"id"`
	InvoiceNumber  string     `json:"invoice_number"`
	PurchaseDate   time.Time  `json:"purchase_date"`
	TotalAmount    float64    `json:"total_amount"`
	PaidAmount     float64    `json:"paid_amount"`
	Remaining      float64    `json:"remaining"`
	Status         string     `json:"status"`
	SupplierName   string     `json:"supplier_name"`
	TotalItems     int        `json:"total_items"`
	CreatedAt      time.Time  `json:"created_at"`
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
