package reports

import (
	"time"
)

// ValidateReportType validates report type
func ValidateReportType(reportType string) error {
	validTypes := map[string]bool{
		"sales":     true,
		"inventory": true,
		"expenses":  true,
		"profits":   true,
		"debts":     true,
		"purchases": true,
		"returns":   true,
		"warranties": true,
	}
	
	if !validTypes[reportType] {
		return ErrInvalidReportType
	}
	return nil
}

// ValidateDateRange validates date range
func ValidateDateRange(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return ErrInvalidDateRange
	}
	return nil
}

// ValidateReportParameters validates report parameters
func ValidateReportParameters(parameters map[string]interface{}) error {
	// Validate date range if provided
	if startDate, ok := parameters["start_date"].(string); ok {
		if endDate, ok := parameters["end_date"].(string); ok {
			start, err := time.Parse(time.RFC3339, startDate)
			if err != nil {
				return ErrInvalidParameters
			}
			end, err := time.Parse(time.RFC3339, endDate)
			if err != nil {
				return ErrInvalidParameters
			}
			return ValidateDateRange(start, end)
		}
	}
	return nil
}
