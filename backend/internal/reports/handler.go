package reports

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/pkg/middleware"
)

// Handler handles HTTP requests for reports
type Handler struct {
	service *Service
	repo    *Repository
}

func maxFloat(value, minimum float64) float64 {
	if value < minimum {
		return minimum
	}
	return value
}

// NewHandler creates a new report handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
		repo:    service.repo,
	}
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

	userID := middleware.GetUserID(c)

	report, err := h.service.GenerateReport(c.Request.Context(), userID, &req)
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

	report, err := h.service.GetReport(c.Request.Context(), id)
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
		t, err := parseReportListBound(startDate, false)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date: use RFC3339 or YYYY-MM-DD", "code": "INVALID_DATE"})
			return
		}
		req.StartDate = &t
	}

	if endDate := c.Query("end_date"); endDate != "" {
		t, err := parseReportListBound(endDate, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date: use RFC3339 or YYYY-MM-DD", "code": "INVALID_DATE"})
			return
		}
		req.EndDate = &t
	}

	reports, total, err := h.service.ListReports(c.Request.Context(), req)
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

	if err := h.service.DeleteReport(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// parseDate parses date string with multiple format support
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	// Try multiple date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006/01/02",
		"01-02-2006",
		"01/02/2006",
	}

	location, err := accounting.StoreLocation()
	if err != nil {
		return time.Time{}, err
	}
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, location); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format")
}

func parseReportListBound(dateStr string, upper bool) (time.Time, error) {
	parsed, err := parseDate(dateStr)
	if err != nil {
		return time.Time{}, err
	}
	calendarDateFormats := []string{"2006-01-02", "2006/01/02", "01-02-2006", "01/02/2006"}
	for _, format := range calendarDateFormats {
		if _, err := time.ParseInLocation(format, dateStr, parsed.Location()); err == nil {
			dateKey, err := accounting.StoreDate(parsed)
			if err != nil {
				return time.Time{}, err
			}
			start, end, err := accounting.StoreDateBounds(dateKey)
			if err != nil {
				return time.Time{}, err
			}
			if upper {
				return end.In(parsed.Location()), nil
			}
			return start.In(parsed.Location()), nil
		}
	}
	if upper {
		return parsed.Add(time.Nanosecond), nil
	}
	return parsed, nil
}

// parseReportDateRange parses the optional date range used by report endpoints.
// Invalid dates are rejected instead of silently falling back to the default
// range, which could otherwise produce a report for the wrong period.
func parseReportDateRange(c *gin.Context) (time.Time, time.Time, error) {
	location, err := accounting.StoreLocation()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	todayKey, err := accounting.StoreDate(time.Now())
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	today, err := time.Parse("2006-01-02", todayKey)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse store date %q: %w", todayKey, err)
	}
	startKey := today.AddDate(0, -1, 0).Format("2006-01-02")
	endKey := today.AddDate(0, 0, 1).Format("2006-01-02")

	if raw := c.Query("start_date"); raw != "" {
		parsed, err := parseDate(raw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date: use RFC3339 or YYYY-MM-DD")
		}
		startKey, err = accounting.StoreDate(parsed)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if raw := c.Query("end_date"); raw != "" {
		parsed, err := parseDate(raw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date: use RFC3339 or YYYY-MM-DD")
		}
		localEndKey, err := accounting.StoreDate(parsed)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		localEndDate, err := time.Parse("2006-01-02", localEndKey)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parse store date %q: %w", localEndKey, err)
		}
		endKey = localEndDate.AddDate(0, 0, 1).Format("2006-01-02")
	}

	startDate, _, err := accounting.StoreDateBounds(startKey)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, _, err := accounting.StoreDateBounds(endKey)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !endDate.After(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must be after start_date")
	}

	// Keep the same instants while retaining the store location for report
	// descriptions and serialized period labels.
	return startDate.In(location), endDate.In(location), nil
}

// GenerateSalesReport handles generating a sales report
// @Summary Generate sales report
// @Description Generate a sales report for specified date range
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} SalesReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/sales [get]
func (h *Handler) GenerateSalesReport(c *gin.Context) {
	startDate, endDate, err := parseReportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_DATE_RANGE"})
		return
	}

	userID := middleware.GetUserID(c)
	// If userID is empty (not authenticated), use a default UUID
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	report, err := h.service.GenerateSalesReport(c.Request.Context(), userID, startDate, endDate)
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
	userID := middleware.GetUserID(c)
	// If userID is empty (not authenticated), use a default UUID
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	report, err := h.service.GenerateInventoryReport(c.Request.Context(), userID)
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
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} ExpensesReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/expenses [get]
func (h *Handler) GenerateExpensesReport(c *gin.Context) {
	startDate, endDate, err := parseReportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_DATE_RANGE"})
		return
	}

	userID := middleware.GetUserID(c)
	// If userID is empty (not authenticated), use a default UUID
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	report, err := h.service.GenerateExpensesReport(c.Request.Context(), userID, startDate, endDate)
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
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} ProfitsReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/profits [get]
func (h *Handler) GenerateProfitsReport(c *gin.Context) {
	startDate, endDate, err := parseReportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_DATE_RANGE"})
		return
	}

	userID := middleware.GetUserID(c)
	// If userID is empty (not authenticated), use a default UUID
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	report, err := h.service.GenerateProfitsReport(c.Request.Context(), userID, startDate, endDate)
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
	userID := middleware.GetUserID(c)
	// If userID is empty (not authenticated), use a default UUID
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	report, err := h.service.GenerateDebtsReport(c.Request.Context(), userID)
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
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} PurchasesReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/purchases [get]
func (h *Handler) GeneratePurchasesReport(c *gin.Context) {
	startDate, endDate, err := parseReportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_DATE_RANGE"})
		return
	}

	userID := middleware.GetUserID(c)
	// If userID is empty (not authenticated), use a default UUID
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	report, err := h.service.GeneratePurchasesReport(c.Request.Context(), userID, startDate, endDate)
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
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} ReturnsReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/returns [get]
func (h *Handler) GenerateReturnsReport(c *gin.Context) {
	startDate, endDate, err := parseReportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_DATE_RANGE"})
		return
	}

	userID := middleware.GetUserID(c)
	// If userID is empty (not authenticated), use a default UUID
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	report, err := h.service.GenerateReturnsReport(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateNetSalesReport handles generating a net sales report
// @Summary Generate net sales report
// @Description Generate a net sales report (gross sales minus returns) for specified date range
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} NetSalesReport
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/net-sales [get]
func (h *Handler) GenerateNetSalesReport(c *gin.Context) {
	startDate, endDate, err := parseReportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_DATE_RANGE"})
		return
	}

	userID := middleware.GetUserID(c)
	// If userID is empty (not authenticated), use a default UUID
	if userID == uuid.Nil {
		userID = uuid.New()
	}

	report, err := h.service.GenerateNetSalesReport(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *Handler) GenerateTaxReport(c *gin.Context) {
	startDate, endDate, err := parseReportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_DATE_RANGE"})
		return
	}
	report, err := h.service.GenerateTaxReport(c.Request.Context(), middleware.GetUserID(c), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

// GenerateProductsReport handles generating a products report
// @Summary Generate products report
// @Description Generate a products report
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/products [get]
func (h *Handler) GenerateProductsReport(c *gin.Context) {
	// Get products data directly from database
	reportData := make(map[string]interface{})

	// Get total products - simplified
	var totalProducts int
	err := h.repo.db.GetContext(c.Request.Context(), &totalProducts,
		"SELECT COUNT(*) FROM products WHERE is_active = true AND deleted_at IS NULL")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product count"})
		return
	}
	reportData["total_products"] = totalProducts

	// Get products by category - simplified
	byCategory := make(map[string]int)
	rows, err := h.repo.db.QueryContext(c.Request.Context(),
		`SELECT COALESCE(c.name, 'غير مصنف'), COUNT(p.id) as count
		 FROM categories c 
		 RIGHT JOIN products p ON c.id = p.category_id AND p.is_active = true
		 WHERE p.is_active = true AND p.deleted_at IS NULL
		 GROUP BY COALESCE(c.name, 'غير مصنف')`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products by category"})
		return
	}
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			rows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read products by category"})
			return
		}
		byCategory[category] = count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read products by category"})
		return
	}
	rows.Close()
	reportData["by_category"] = byCategory

	// Get low stock products - simplified
	var lowStockCount int
	err = h.repo.db.GetContext(c.Request.Context(), &lowStockCount,
		`WITH serialized_stock AS (
			SELECT product_id, COUNT(*) AS quantity
			FROM inventory_items
			WHERE status = 'AVAILABLE' AND UPPER(COALESCE(condition, '')) <> 'USED'
			GROUP BY product_id
		), product_stock AS (
			SELECT p.id, p.min_stock_level, COALESCE(inv.quantity, serialized_stock.quantity, 0) AS quantity
			FROM products p
			LEFT JOIN (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id
			LEFT JOIN serialized_stock ON serialized_stock.product_id = p.id
			WHERE p.is_active = TRUE AND p.deleted_at IS NULL AND p.min_stock_level > 0
		)
		SELECT COUNT(*) FROM product_stock WHERE quantity <= min_stock_level`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve low stock count"})
		return
	}
	reportData["low_stock_count"] = lowStockCount

	c.JSON(http.StatusOK, gin.H{"data": reportData})
}

// GenerateSuppliersReport handles generating a suppliers report
// @Summary Generate suppliers report
// @Description Generate a suppliers report
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/reports/suppliers [get]
func (h *Handler) GenerateSuppliersReport(c *gin.Context) {
	// Get suppliers data directly from database
	reportData := make(map[string]interface{})

	var activeSuppliers, inactiveSuppliers int
	err := h.repo.db.GetContext(c.Request.Context(), &activeSuppliers,
		"SELECT COUNT(*) FROM suppliers WHERE is_active = true")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve supplier count"})
		return
	}
	reportData["total_suppliers"] = activeSuppliers
	if err := h.repo.db.GetContext(c.Request.Context(), &inactiveSuppliers,
		"SELECT COUNT(*) FROM suppliers WHERE is_active = false"); err == nil {
		reportData["inactive_suppliers"] = inactiveSuppliers
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve inactive supplier count"})
		return
	}

	var purchaseTotals struct {
		Total    float64 `db:"total"`
		Paid     float64 `db:"paid"`
		Returns  float64 `db:"returns"`
		Payments float64 `db:"payments"`
		Open     float64 `db:"open"`
	}
	if err := h.repo.db.GetContext(c.Request.Context(), &purchaseTotals,
		`SELECT COALESCE(SUM(total_amount), 0) AS total,
		        COALESCE(SUM(paid_amount), 0) AS paid,
		        COALESCE((SELECT SUM(refund_amount) FROM supplier_returns WHERE UPPER(COALESCE(status, '')) = 'COMPLETED'), 0) AS returns,
				COALESCE(SUM(paid_amount), 0) + COALESCE((SELECT SUM(amount) FROM supplier_ledger sl WHERE sl.type = 'credit' AND sl.transaction_type = 'PAYMENT' AND NOT EXISTS (SELECT 1 FROM purchases p2 WHERE p2.id = sl.reference_id)), 0) AS payments,
		        0 AS open
		 FROM purchases
		 WHERE LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve supplier purchase totals"})
		return
	}
	reportData["total_purchases"] = purchaseTotals.Total
	reportData["supplier_return_credits"] = purchaseTotals.Returns
	reportData["supplier_payments"] = purchaseTotals.Payments
	reportData["total_paid"] = purchaseTotals.Payments
	var supplierBalances struct {
		Outstanding float64 `db:"outstanding"`
		Credit      float64 `db:"credit"`
	}
	if err := h.repo.db.GetContext(c.Request.Context(), &supplierBalances, `
		SELECT COALESCE(SUM(CASE WHEN current_balance > 0 THEN current_balance ELSE 0 END), 0) AS outstanding,
		       COALESCE(SUM(CASE WHEN current_balance < 0 THEN -current_balance ELSE 0 END), 0) AS credit
		FROM suppliers`); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve supplier balances"})
		return
	}
	reportData["total_outstanding"] = supplierBalances.Outstanding
	reportData["supplier_credit_balance"] = supplierBalances.Credit

	bySupplier := []map[string]interface{}{}
	rows, err := h.repo.db.QueryContext(c.Request.Context(), `
			SELECT COALESCE(s.name, 'مورد غير معروف'),
				COALESCE(SUM(p.total_amount), 0),
				COALESCE(SUM(p.paid_amount), 0) + COALESCE((SELECT SUM(amount) FROM supplier_ledger sl WHERE sl.supplier_id = s.id AND sl.type = 'credit' AND sl.transaction_type = 'PAYMENT' AND NOT EXISTS (SELECT 1 FROM purchases p2 WHERE p2.id = sl.reference_id)), 0),
				COALESCE(s.current_balance, 0)
			FROM suppliers s
			LEFT JOIN purchases p ON p.supplier_id = s.id AND LOWER(COALESCE(p.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY s.id, s.name, s.current_balance
			HAVING COALESCE(SUM(p.total_amount), 0) > 0 OR COALESCE(s.current_balance, 0) <> 0
			ORDER BY SUM(p.total_amount) DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve purchases by supplier"})
		return
	}
	for rows.Next() {
		var name string
		var total, paid, outstanding float64
		if err := rows.Scan(&name, &total, &paid, &outstanding); err != nil {
			rows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read purchases by supplier"})
			return
		}
		rawOutstanding := outstanding
		bySupplier = append(bySupplier, map[string]interface{}{
			"supplier_name": name, "total_purchases": total,
			"total_paid": paid, "outstanding": maxFloat(outstanding, 0),
			"credit_balance": maxFloat(-rawOutstanding, 0),
		})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read purchases by supplier"})
		return
	}
	rows.Close()
	reportData["by_supplier"] = bySupplier

	suppliersWithBalance := []map[string]interface{}{}
	rows, err = h.repo.db.QueryContext(c.Request.Context(),
		`SELECT s.id, s.name,
		        COALESCE(SUM(p.total_amount), 0) AS total_purchases,
		        COALESCE(SUM(p.paid_amount), 0) + COALESCE((SELECT SUM(amount) FROM supplier_ledger sl WHERE sl.supplier_id = s.id AND sl.type = 'credit' AND sl.transaction_type = 'PAYMENT' AND NOT EXISTS (SELECT 1 FROM purchases p2 WHERE p2.id = sl.reference_id)), 0) AS total_paid,
		        COALESCE(s.current_balance, 0) AS balance
				FROM suppliers s
				LEFT JOIN purchases p ON p.supplier_id = s.id AND LOWER(COALESCE(p.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
				WHERE COALESCE(s.current_balance, 0) > 0
				GROUP BY s.id, s.name, s.current_balance
				ORDER BY balance DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve suppliers with balance"})
		return
	}
	for rows.Next() {
		var id uuid.UUID
		var name string
		var totalPurchases float64
		var totalPaid float64
		var balance float64
		if err := rows.Scan(&id, &name, &totalPurchases, &totalPaid, &balance); err != nil {
			rows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read suppliers with balance"})
			return
		}
		suppliersWithBalance = append(suppliersWithBalance, map[string]interface{}{
			"id": id, "name": name, "total_purchases": totalPurchases,
			"total_paid": totalPaid, "balance": balance,
			"credit_balance": float64(0),
		})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read suppliers with balance"})
		return
	}
	rows.Close()
	reportData["suppliers_with_balance"] = suppliersWithBalance

	c.JSON(http.StatusOK, gin.H{"data": reportData})
}
