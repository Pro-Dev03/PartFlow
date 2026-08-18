package inspections

import (
	"time"

	"github.com/google/uuid"
)

// ToInspectionResponse converts Inspection to InspectionResponse
func (i *Inspection) ToInspectionResponse(product *ProductInfo, inspector *UserInfo) *InspectionResponse {
	return &InspectionResponse{
		Inspection: *i,
		Product:   product,
		Inspector: inspector,
	}
}

// ToInspectionListItem converts Inspection to list item format
func (i *Inspection) ToInspectionListItem(productName string, inspectorName string) map[string]interface{} {
	return map[string]interface{}{
		"id":              i.ID,
		"product_name":    productName,
		"serial_number":   i.SerialNumber,
		"inspection_date": i.InspectionDate,
		"status":          i.Status,
		"condition":       i.Condition,
		"grade":           i.Grade,
		"inspector_name":  inspectorName,
		"created_at":      i.CreatedAt,
	}
}

// CreateInspection creates an Inspection from request
func CreateInspection(organizationID uuid.UUID, userID uuid.UUID, req *InspectionRequest) *Inspection {
	return &Inspection{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ProductID:      req.ProductID,
		SerialNumber:   req.SerialNumber,
		InspectionDate: req.InspectionDate,
		InspectedBy:    userID,
		Status:         "pending",
		Condition:      req.Condition,
		Grade:          req.Grade,
		Notes:          req.Notes,
		Photos:         req.Photos,
		TestResults:    req.TestResults,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// ValidateInspectionRequest validates inspection request
func ValidateInspectionRequest(req *InspectionRequest) error {
	if req.ProductID == uuid.Nil {
		return ErrProductNotFound
	}
	if req.Condition != "excellent" && req.Condition != "very_good" && 
		req.Condition != "good" && req.Condition != "fair" && req.Condition != "poor" {
		return ErrInvalidCondition
	}
	if req.Grade != "A" && req.Grade != "B" && req.Grade != "C" && req.Grade != "D" && req.Grade != "F" {
		return ErrInvalidGrade
	}
	return nil
}

// IsPassed checks if inspection is passed
func (i *Inspection) IsPassed() bool {
	return i.Status == "passed"
}

// IsFailed checks if inspection is failed
func (i *Inspection) IsFailed() bool {
	return i.Status == "failed"
}

// IsCompleted checks if inspection is completed
func (i *Inspection) IsCompleted() bool {
	return i.Status == "passed" || i.Status == "failed"
}

// CalculateOverallStatus calculates overall status based on test results
func (tr *TestResults) CalculateOverallStatus() string {
	tests := []bool{
		tr.PowerTest,
		tr.TemperatureTest,
		tr.PerformanceTest,
		tr.PortsTest,
		tr.StorageTest,
		tr.VisualTest,
		tr.SerialTest,
	}
	
	allPassed := true
	for _, test := range tests {
		if !test {
			allPassed = false
			break
		}
	}
	
	if allPassed {
		return "passed"
	}
	return "failed"
}
