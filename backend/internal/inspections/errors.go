package inspections

import "errors"

var (
	// ErrInspectionNotFound is returned when inspection is not found
	ErrInspectionNotFound = errors.New("inspection not found")

	// ErrProductNotFound is returned when product is not found
	ErrProductNotFound = errors.New("product not found")

	// ErrInspectionExists is returned when inspection already exists
	ErrInspectionExists = errors.New("inspection already exists")

	// ErrInvalidInspectionStatus is returned when inspection status is invalid
	ErrInvalidInspectionStatus = errors.New("invalid inspection status")

	// ErrInvalidCondition is returned when condition is invalid
	ErrInvalidCondition = errors.New("invalid condition")

	// ErrInvalidGrade is returned when grade is invalid
	ErrInvalidGrade = errors.New("invalid grade")

	// ErrInspectionAlreadyPassed is returned when inspection is already passed
	ErrInspectionAlreadyPassed = errors.New("inspection already passed")

	// ErrInspectionAlreadyFailed is returned when inspection is already failed
	ErrInspectionAlreadyFailed = errors.New("inspection already failed")

	// ErrCannotUpdateCompletedInspection is returned when trying to update completed inspection
	ErrCannotUpdateCompletedInspection = errors.New("cannot update completed inspection")
)
