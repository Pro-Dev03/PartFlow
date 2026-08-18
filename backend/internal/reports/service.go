package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles report business logic
type Service struct {
	repo *Repository
}

// NewService creates a new report service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GenerateReport generates a new report
func (s *Service) GenerateReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *ReportRequest) (*Report, error) {
	// Validate request
	if err := ValidateReportRequest(req); err != nil {
		return nil, err
	}

	// Create report
	report := CreateReport(organizationID, userID, req)

	if err := s.repo.CreateReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Generate report data based on type
	var data interface{}
	var err error

	params, _ := report.ParseParameters()
	startDate, _ := params["start_date"].(string)
	endDate, _ := params["end_date"].(string)

	var start, end time.Time
	if startDate != "" {
		start, _ = time.Parse(time.RFC3339, startDate)
	}
	if endDate != "" {
		end, _ = time.Parse(time.RFC3339, endDate)
	}

	switch req.Type {
	case "sales":
		data, err = s.repo.GetSalesData(ctx, organizationID, start, end)
	case "inventory":
		data, err = s.repo.GetInventoryData(ctx, organizationID)
	case "expenses":
		data, err = s.repo.GetExpensesData(ctx, organizationID, start, end)
	case "profits":
		data, err = s.repo.GetProfitsData(ctx, organizationID, start, end)
	case "debts":
		data, err = s.repo.GetDebtsData(ctx, organizationID)
	case "purchases":
		data, err = s.repo.GetPurchasesData(ctx, organizationID, start, end)
	case "returns":
		data, err = s.repo.GetReturnsData(ctx, organizationID, start, end)
	case "warranties":
		data, err = s.repo.GetWarrantyData(ctx, organizationID)
	default:
		err = ErrInvalidReportType
	}

	if err != nil {
		report.Status = "failed"
		s.repo.UpdateReport(ctx, report)
		return nil, fmt.Errorf("failed to generate report data: %w", err)
	}

	// Set report data
	if err := report.SetData(data); err != nil {
		report.Status = "failed"
		s.repo.UpdateReport(ctx, report)
		return nil, fmt.Errorf("failed to set report data: %w", err)
	}

	report.Status = "completed"
	report.UpdatedAt = time.Now()

	if err := s.repo.UpdateReport(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

// GetReport retrieves a report by ID
func (s *Service) GetReport(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Report, error) {
	return s.repo.GetReportByID(ctx, id, organizationID)
}

// ListReports retrieves reports with pagination and filters
func (s *Service) ListReports(ctx context.Context, organizationID uuid.UUID, req ReportListRequest) ([]map[string]interface{}, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	reports, total, err := s.repo.ListReports(ctx, organizationID, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items
	var result []map[string]interface{}
	for _, report := range reports {
		generatorName, err := s.repo.GetUserName(ctx, report.GeneratedBy)
		if err != nil {
			generatorName = "Unknown"
		}
		result = append(result, report.ToReportListItem(generatorName))
	}

	return result, total, nil
}

// DeleteReport deletes a report
func (s *Service) DeleteReport(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.DeleteReport(ctx, id, organizationID)
}

// GenerateSalesReport generates a sales report
func (s *Service) GenerateSalesReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, startDate, endDate time.Time) (*SalesReport, error) {
	req := &ReportRequest{
		Type:        "sales",
		Title:       "Sales Report",
		Description: fmt.Sprintf("Sales report from %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Parameters: map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		},
	}

	report, err := s.GenerateReport(ctx, organizationID, userID, req)
	if err != nil {
		return nil, err
	}

	data, err := report.ParseData()
	if err != nil {
		return nil, err
	}

	salesReport, ok := data.(*SalesReport)
	if !ok {
		return nil, ErrReportGenerationFailed
	}

	return salesReport, nil
}

// GenerateInventoryReport generates an inventory report
func (s *Service) GenerateInventoryReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID) (*InventoryReport, error) {
	req := &ReportRequest{
		Type:        "inventory",
		Title:       "Inventory Report",
		Description: "Current inventory status",
		Parameters:  map[string]interface{}{},
	}

	report, err := s.GenerateReport(ctx, organizationID, userID, req)
	if err != nil {
		return nil, err
	}

	data, err := report.ParseData()
	if err != nil {
		return nil, err
	}

	inventoryReport, ok := data.(*InventoryReport)
	if !ok {
		return nil, ErrReportGenerationFailed
	}

	return inventoryReport, nil
}

// GenerateExpensesReport generates an expenses report
func (s *Service) GenerateExpensesReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, startDate, endDate time.Time) (*ExpensesReport, error) {
	req := &ReportRequest{
		Type:        "expenses",
		Title:       "Expenses Report",
		Description: fmt.Sprintf("Expenses report from %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Parameters: map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		},
	}

	report, err := s.GenerateReport(ctx, organizationID, userID, req)
	if err != nil {
		return nil, err
	}

	data, err := report.ParseData()
	if err != nil {
		return nil, err
	}

	expensesReport, ok := data.(*ExpensesReport)
	if !ok {
		return nil, ErrReportGenerationFailed
	}

	return expensesReport, nil
}

// GenerateProfitsReport generates a profits report
func (s *Service) GenerateProfitsReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, startDate, endDate time.Time) (*ProfitsReport, error) {
	req := &ReportRequest{
		Type:        "profits",
		Title:       "Profits Report",
		Description: fmt.Sprintf("Profits report from %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Parameters: map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		},
	}

	report, err := s.GenerateReport(ctx, organizationID, userID, req)
	if err != nil {
		return nil, err
	}

	data, err := report.ParseData()
	if err != nil {
		return nil, err
	}

	profitsReport, ok := data.(*ProfitsReport)
	if !ok {
		return nil, ErrReportGenerationFailed
	}

	return profitsReport, nil
}

// GenerateDebtsReport generates a debts report
func (s *Service) GenerateDebtsReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID) (*DebtsReport, error) {
	req := &ReportRequest{
		Type:        "debts",
		Title:       "Debts Report",
		Description: "Current debts status",
		Parameters:  map[string]interface{}{},
	}

	report, err := s.GenerateReport(ctx, organizationID, userID, req)
	if err != nil {
		return nil, err
	}

	data, err := report.ParseData()
	if err != nil {
		return nil, err
	}

	debtsReport, ok := data.(*DebtsReport)
	if !ok {
		return nil, ErrReportGenerationFailed
	}

	return debtsReport, nil
}

// GeneratePurchasesReport generates a purchases report
func (s *Service) GeneratePurchasesReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, startDate, endDate time.Time) (*PurchasesReport, error) {
	req := &ReportRequest{
		Type:        "purchases",
		Title:       "Purchases Report",
		Description: fmt.Sprintf("Purchases report from %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Parameters: map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		},
	}

	report, err := s.GenerateReport(ctx, organizationID, userID, req)
	if err != nil {
		return nil, err
	}

	data, err := report.ParseData()
	if err != nil {
		return nil, err
	}

	purchasesReport, ok := data.(*PurchasesReport)
	if !ok {
		return nil, ErrReportGenerationFailed
	}

	return purchasesReport, nil
}

// GenerateReturnsReport generates a returns report
func (s *Service) GenerateReturnsReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, startDate, endDate time.Time) (*ReturnsReport, error) {
	req := &ReportRequest{
		Type:        "returns",
		Title:       "Returns Report",
		Description: fmt.Sprintf("Returns report from %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Parameters: map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		},
	}

	report, err := s.GenerateReport(ctx, organizationID, userID, req)
	if err != nil {
		return nil, err
	}

	data, err := report.ParseData()
	if err != nil {
		return nil, err
	}

	returnsReport, ok := data.(*ReturnsReport)
	if !ok {
		return nil, ErrReportGenerationFailed
	}

	return returnsReport, nil
}

// GenerateWarrantyReport generates a warranty report
func (s *Service) GenerateWarrantyReport(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID) (*WarrantyReport, error) {
	// Generate report directly without using the general report generation
	// since warranty reports don't require date ranges
	report, err := s.repo.GetWarrantyData(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	return report, nil
}
