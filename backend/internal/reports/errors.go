package reports

import "errors"

var (
	// ErrReportNotFound is returned when report is not found
	ErrReportNotFound = errors.New("report not found")

	// ErrInvalidReportType is returned when report type is invalid
	ErrInvalidReportType = errors.New("invalid report type")

	// ErrInvalidReportStatus is returned when report status is invalid
	ErrInvalidReportStatus = errors.New("invalid report status")

	// ErrInvalidDateRange is returned when date range is invalid
	ErrInvalidDateRange = errors.New("invalid date range")

	// ErrReportGenerationFailed is returned when report generation fails
	ErrReportGenerationFailed = errors.New("report generation failed")

	// ErrInsufficientData is returned when there's insufficient data for report
	ErrInsufficientData = errors.New("insufficient data for report")

	// ErrInvalidParameters is returned when report parameters are invalid
	ErrInvalidParameters = errors.New("invalid report parameters")
)
