package reports

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ToReportListItem converts Report to list item format
func (r *Report) ToReportListItem(generatorName string) map[string]interface{} {
	return map[string]interface{}{
		"id":           r.ID,
		"type":         r.Type,
		"title":        r.Title,
		"description":  r.Description,
		"status":       r.Status,
		"generated_by": generatorName,
		"generated_at": r.GeneratedAt,
		"created_at":   r.CreatedAt,
	}
}

// CreateReport creates a Report from request
func CreateReport(organizationID uuid.UUID, userID uuid.UUID, req *ReportRequest) *Report {
	parametersJSON, _ := json.Marshal(req.Parameters)
	
	return &Report{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Type:           req.Type,
		Title:          req.Title,
		Description:    req.Description,
		Parameters:     string(parametersJSON),
		Status:         "pending",
		GeneratedBy:    userID,
		GeneratedAt:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// ValidateReportRequest validates report request
func ValidateReportRequest(req *ReportRequest) error {
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
	
	if !validTypes[req.Type] {
		return ErrInvalidReportType
	}
	
	if req.Title == "" {
		return ErrReportNotFound
	}
	
	// Validate date range if provided
	if startDate, ok := req.Parameters["start_date"]; ok {
		if endDate, ok := req.Parameters["end_date"]; ok {
			if start, err := time.Parse(time.RFC3339, startDate.(string)); err == nil {
				if end, err := time.Parse(time.RFC3339, endDate.(string)); err == nil {
					if end.Before(start) {
						return ErrInvalidDateRange
					}
				}
			}
		}
	}
	
	return nil
}

// ValidateReportStatus validates report status
func ValidateReportStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":  true,
		"completed": true,
		"failed":   true,
	}
	
	if !validStatuses[status] {
		return ErrInvalidReportStatus
	}
	return nil
}

// ParseParameters parses parameters JSON string
func (r *Report) ParseParameters() (map[string]interface{}, error) {
	var params map[string]interface{}
	if r.Parameters != "" {
		err := json.Unmarshal([]byte(r.Parameters), &params)
		if err != nil {
			return nil, err
		}
	}
	return params, nil
}

// ParseData parses data JSON string
func (r *Report) ParseData() (interface{}, error) {
	var data interface{}
	if r.Data != "" {
		err := json.Unmarshal([]byte(r.Data), &data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// SetData sets data as JSON string
func (r *Report) SetData(data interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r.Data = string(dataJSON)
	return nil
}

// CalculateProfitMargin calculates profit margin
func (pr *ProfitsReport) CalculateProfitMargin() float64 {
	if pr.TotalRevenue == 0 {
		return 0
	}
	return (pr.NetProfit / pr.TotalRevenue) * 100
}

// CalculateGrossProfitMargin calculates gross profit margin
func (sr *SalesReport) CalculateProfitMargin() float64 {
	if sr.TotalRevenue == 0 {
		return 0
	}
	return (sr.GrossProfit / sr.TotalRevenue) * 100
}
