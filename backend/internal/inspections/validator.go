package inspections

import (
	"github.com/google/uuid"
)

// ValidateInspectionStatus validates inspection status
func ValidateInspectionStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":      true,
		"passed":       true,
		"failed":       true,
		"needs_repair": true,
	}
	
	if !validStatuses[status] {
		return ErrInvalidInspectionStatus
	}
	return nil
}

// ValidateCondition validates condition
func ValidateCondition(condition string) error {
	validConditions := map[string]bool{
		"excellent":  true,
		"very_good":  true,
		"good":       true,
		"fair":       true,
		"poor":       true,
	}
	
	if !validConditions[condition] {
		return ErrInvalidCondition
	}
	return nil
}

// ValidateGrade validates grade
func ValidateGrade(grade string) error {
	validGrades := map[string]bool{
		"A": true,
		"B": true,
		"C": true,
		"D": true,
		"F": true,
	}
	
	if !validGrades[grade] {
		return ErrInvalidGrade
	}
	return nil
}

// ValidateProductID validates product ID
func ValidateProductID(productID uuid.UUID) error {
	if productID == uuid.Nil {
		return ErrProductNotFound
	}
	return nil
}

// ValidateInspectionID validates inspection ID
func ValidateInspectionID(inspectionID uuid.UUID) error {
	if inspectionID == uuid.Nil {
		return ErrInspectionNotFound
	}
	return nil
}
