package warranties

import (
	"time"

	"github.com/google/uuid"
)

// ToWarrantyResponse converts Warranty to WarrantyResponse
func (w *Warranty) ToWarrantyResponse(product *ProductInfo, customer *CustomerInfo, claims []WarrantyClaim) *WarrantyResponse {
	return &WarrantyResponse{
		Warranty: *w,
		Product:  product,
		Customer: customer,
		Claims:   claims,
	}
}

// ToWarrantyClaimResponse converts WarrantyClaim to WarrantyClaimResponse
func (wc *WarrantyClaim) ToWarrantyClaimResponse(warranty *Warranty, customer *CustomerInfo) *WarrantyClaimResponse {
	return &WarrantyClaimResponse{
		Claim:    *wc,
		Warranty: warranty,
		Customer: customer,
	}
}

// ToWarrantyListItem converts Warranty to list item format
func (w *Warranty) ToWarrantyListItem(productName string, customerName string) map[string]interface{} {
	return map[string]interface{}{
		"id":              w.ID,
		"warranty_number": w.WarrantyNumber,
		"product_name":    productName,
		"serial_number":   w.SerialNumber,
		"warranty_type":   w.WarrantyType,
		"start_date":      w.StartDate,
		"end_date":        w.EndDate,
		"status":          w.Status,
		"customer_name":   customerName,
		"created_at":      w.CreatedAt,
	}
}

// ToWarrantyClaimListItem converts WarrantyClaim to list item format
func (wc *WarrantyClaim) ToWarrantyClaimListItem(warrantyNumber string, customerName string, productName string) map[string]interface{} {
	return map[string]interface{}{
		"id":                  wc.ID,
		"claim_number":        wc.ClaimNumber,
		"warranty_number":     warrantyNumber,
		"product_name":        productName,
		"customer_name":       customerName,
		"claim_date":          wc.ClaimDate,
		"issue_description":   wc.IssueDescription,
		"status":              wc.Status,
		"resolution":          wc.Resolution,
		"resolution_date":     wc.ResolutionDate,
		"created_at":          wc.CreatedAt,
	}
}

// CreateWarranty creates a Warranty from request
func CreateWarranty(organizationID uuid.UUID, userID uuid.UUID, req *WarrantyRequest) *Warranty {
	endDate := req.StartDate.AddDate(0, 0, req.WarrantyPeriod)
	
	return &Warranty{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ProductID:      req.ProductID,
		SerialNumber:   req.SerialNumber,
		WarrantyNumber: generateWarrantyNumber(),
		WarrantyType:   req.WarrantyType,
		WarrantyPeriod: req.WarrantyPeriod,
		StartDate:      req.StartDate,
		EndDate:        endDate,
		Status:         "active",
		CustomerID:     req.CustomerID,
		SaleID:         req.SaleID,
		Terms:          req.Terms,
		Notes:          req.Notes,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// CreateWarrantyClaim creates a WarrantyClaim from request
func CreateWarrantyClaim(organizationID uuid.UUID, userID uuid.UUID, req *WarrantyClaimRequest) *WarrantyClaim {
	return &WarrantyClaim{
		ID:               uuid.New(),
		OrganizationID:   organizationID,
		WarrantyID:       req.WarrantyID,
		ClaimNumber:      generateClaimNumber(),
		ClaimDate:        time.Now(),
		CustomerID:       req.CustomerID,
		IssueDescription: req.IssueDescription,
		Status:           "pending",
		Notes:            req.Notes,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// ValidateWarrantyRequest validates warranty request
func ValidateWarrantyRequest(req *WarrantyRequest) error {
	if req.ProductID == uuid.Nil {
		return ErrProductNotFound
	}
	if req.WarrantyType != "manufacturer" && req.WarrantyType != "seller" && req.WarrantyType != "extended" {
		return ErrInvalidWarrantyType
	}
	if req.WarrantyPeriod <= 0 {
		return ErrInvalidWarrantyPeriod
	}
	return nil
}

// ValidateWarrantyClaimRequest validates warranty claim request
func ValidateWarrantyClaimRequest(req *WarrantyClaimRequest) error {
	if req.WarrantyID == uuid.Nil {
		return ErrWarrantyNotFound
	}
	if req.CustomerID == uuid.Nil {
		return ErrCustomerNotFound
	}
	if req.IssueDescription == "" {
		return ErrWarrantyClaimNotFound
	}
	return nil
}

// generateWarrantyNumber generates a unique warranty number
func generateWarrantyNumber() string {
	return "WRT-" + uuid.New().String()[:8]
}

// generateClaimNumber generates a unique claim number
func generateClaimNumber() string {
	return "CLM-" + uuid.New().String()[:8]
}

// IsExpired checks if warranty is expired
func (w *Warranty) IsExpired() bool {
	return time.Now().After(w.EndDate)
}

// DaysRemaining returns the number of days remaining in warranty
func (w *Warranty) DaysRemaining() int {
	if w.IsExpired() {
		return 0
	}
	duration := time.Until(w.EndDate)
	return int(duration.Hours() / 24)
}
