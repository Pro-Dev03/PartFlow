package reports

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
)

// Handler handles HTTP requests for reports
type Handler struct {
	service *Service
}

// NewHandler creates a new report handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GenerateReport handles report generation
// @Summary Generate a new report
// @Description Generate a new report with specified type and parameters
// @Tags reports
// @Accept json
// @Produce json
// @Param request body ReportRequest true "Report request"
// @Success 201 {object} Report
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports [post]
func (h *Handler) GenerateReport(c *gin.Context) {
	var req ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateReport(c.Request.Context(), organizationID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// GetReport handles getting a report by ID
// @Summary Get a report
// @Description Get a report by ID
// @Tags reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} Report
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/{id} [get]
func (h *Handler) GetReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	report, err := h.service.GetReport(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// ListReports handles listing reports
// @Summary List reports
// @Description List reports with pagination and filters
// @Tags reports
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param type query string false "Report type filter" Enums(sales, inventory, expenses, profits, debts, purchases, returns, warranties)
// @Param status query string false "Status filter" Enums(pending, completed, failed)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param generated_by query string false "Generator ID filter"
// @Param search query string false "Search in title and description"
// @Param sort_by query string false "Sort by field" default(generated_at)
// @Param sort_order query string false "Sort order" default(DESC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports [get]
func (h *Handler) ListReports(c *gin.Context) {
	var req ReportListRequest
	
	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}
	
	if generatedBy := c.Query("generated_by"); generatedBy != "" {
		if id, err := uuid.Parse(generatedBy); err == nil {
			req.GeneratedBy = &id
		}
	}
	
	req.Type = c.Query("type")
	req.Status = c.Query("status")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "generated_at")
	req.SortOrder = c.DefaultQuery("sort_order", "DESC")
	
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			req.StartDate = &t
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			req.EndDate = &t
		}
	}

	organizationID := middleware.GetOrganizationID(c)

	reports, total, err := h.service.ListReports(c.Request.Context(), organizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reports,
		"meta": gin.H{
			"page":        req.Page,
			"per_page":    req.PerPage,
			"total":       total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// DeleteReport handles deleting a report
// @Summary Delete a report
// @Description Delete a report by ID
// @Tags reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/{id} [delete]
func (h *Handler) DeleteReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeleteReport(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GenerateSalesReport handles generating a sales report
// @Summary Generate sales report
// @Description Generate a sales report for specified date range
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string true "Start date"
// @Param end_date query string true "End date"
// @Success 200 {object} SalesReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/sales [get]
func (h *Handler) GenerateSalesReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateSalesReport(c.Request.Context(), organizationID, userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateInventoryReport handles generating an inventory report
// @Summary Generate inventory report
// @Description Generate an inventory report
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} InventoryReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/inventory [get]
func (h *Handler) GenerateInventoryReport(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateInventoryReport(c.Request.Context(), organizationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateExpensesReport handles generating an expenses report
// @Summary Generate expenses report
// @Description Generate an expenses report for specified date range
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string true "Start date"
// @Param end_date query string true "End date"
// @Success 200 {object} ExpensesReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/expenses [get]
func (h *Handler) GenerateExpensesReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateExpensesReport(c.Request.Context(), organizationID, userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateProfitsReport handles generating a profits report
// @Summary Generate profits report
// @Description Generate a profits report for specified date range
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string true "Start date"
// @Param end_date query string true "End date"
// @Success 200 {object} ProfitsReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/profits [get]
func (h *Handler) GenerateProfitsReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateProfitsReport(c.Request.Context(), organizationID, userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateDebtsReport handles generating a debts report
// @Summary Generate debts report
// @Description Generate a debts report
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} DebtsReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/debts [get]
func (h *Handler) GenerateDebtsReport(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateDebtsReport(c.Request.Context(), organizationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GeneratePurchasesReport handles generating a purchases report
// @Summary Generate purchases report
// @Description Generate a purchases report for specified date range
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string true "Start date"
// @Param end_date query string true "End date"
// @Success 200 {object} PurchasesReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/purchases [get]
func (h *Handler) GeneratePurchasesReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GeneratePurchasesReport(c.Request.Context(), organizationID, userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateReturnsReport handles generating a returns report
// @Summary Generate returns report
// @Description Generate a returns report for specified date range
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string true "Start date"
// @Param end_date query string true "End date"
// @Success 200 {object} ReturnsReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/returns [get]
func (h *Handler) GenerateReturnsReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateReturnsReport(c.Request.Context(), organizationID, userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateWarrantyReport handles generating a warranty report
// @Summary Generate warranty report
// @Description Generate a warranty report
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} WarrantyReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/warranty [get]
func (h *Handler) GenerateWarrantyReport(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateWarrantyReport(c.Request.Context(), organizationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}
