package reports

import (
	"fmt"
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
	repo    *Repository
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
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			req.StartDate = &t
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			req.EndDate = &t
		}
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

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format")
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Default to this month if no dates provided
	startDate := time.Now().AddDate(0, -1, 0).Truncate(time.Hour * 24)
	endDate := time.Now().Truncate(time.Hour * 24).Add(24 * time.Hour)

	if startDateStr != "" {
		if parsed, err := parseDate(startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := parseDate(endDateStr); err == nil {
			endDate = parsed.Add(24 * time.Hour)
		}
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Default to this month if no dates provided
	startDate := time.Now().AddDate(0, -1, 0).Truncate(time.Hour * 24)
	endDate := time.Now().Truncate(time.Hour * 24).Add(24 * time.Hour)

	if startDateStr != "" {
		if parsed, err := parseDate(startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := parseDate(endDateStr); err == nil {
			endDate = parsed.Add(24 * time.Hour)
		}
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Default to this month if no dates provided
	startDate := time.Now().AddDate(0, -1, 0).Truncate(time.Hour * 24)
	endDate := time.Now().Truncate(time.Hour * 24).Add(24 * time.Hour)

	if startDateStr != "" {
		if parsed, err := parseDate(startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := parseDate(endDateStr); err == nil {
			endDate = parsed.Add(24 * time.Hour)
		}
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Default to this month if no dates provided
	startDate := time.Now().AddDate(0, -1, 0).Truncate(time.Hour * 24)
	endDate := time.Now().Truncate(time.Hour * 24).Add(24 * time.Hour)

	if startDateStr != "" {
		if parsed, err := parseDate(startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := parseDate(endDateStr); err == nil {
			endDate = parsed.Add(24 * time.Hour)
		}
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Default to this month if no dates provided
	startDate := time.Now().AddDate(0, -1, 0).Truncate(time.Hour * 24)
	endDate := time.Now().Truncate(time.Hour * 24).Add(24 * time.Hour)

	if startDateStr != "" {
		if parsed, err := parseDate(startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := parseDate(endDateStr); err == nil {
			endDate = parsed.Add(24 * time.Hour)
		}
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Default to this month if no dates provided
	startDate := time.Now().AddDate(0, -1, 0).Truncate(time.Hour * 24)
	endDate := time.Now().Truncate(time.Hour * 24).Add(24 * time.Hour)

	if startDateStr != "" {
		if parsed, err := parseDate(startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := parseDate(endDateStr); err == nil {
			endDate = parsed.Add(24 * time.Hour)
		}
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
		"SELECT COUNT(*) FROM products WHERE is_active = true")
	if err != nil {
		// If products table doesn't exist, return empty report
		reportData["total_products"] = 0
		reportData["by_category"] = make(map[string]int)
		reportData["low_stock_count"] = 0
		c.JSON(http.StatusOK, gin.H{"data": reportData})
		return
	}
	reportData["total_products"] = totalProducts

	// Get products by category - simplified
	byCategory := make(map[string]int)
	rows, err := h.repo.db.QueryContext(c.Request.Context(),
		`SELECT COALESCE(c.name, 'غير مصنف'), COUNT(p.id) as count
		 FROM categories c 
		 RIGHT JOIN products p ON c.id = p.category_id AND p.is_active = true
		 WHERE p.is_active = true
		 GROUP BY COALESCE(c.name, 'غير مصنف')`)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var category string
			var count int
			if err := rows.Scan(&category, &count); err != nil {
				continue
			}
			byCategory[category] = count
		}
	}
	reportData["by_category"] = byCategory

	// Get low stock products - simplified
	var lowStockCount int
	err = h.repo.db.GetContext(c.Request.Context(), &lowStockCount,
		`SELECT COUNT(*) FROM (
			SELECT p.id
			FROM inventory_items ii
			JOIN products p ON ii.product_id = p.id
			WHERE ii.status = 'AVAILABLE' AND p.min_stock_level > 0
			GROUP BY p.id, p.min_stock_level
			HAVING COUNT(ii.id) < p.min_stock_level
		) low_stock`)
	if err != nil {
		lowStockCount = 0
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
		// If suppliers table doesn't exist, return empty report
		reportData["total_suppliers"] = 0
		reportData["inactive_suppliers"] = 0
		reportData["suppliers_with_balance"] = []map[string]interface{}{}
		c.JSON(http.StatusOK, gin.H{"data": reportData})
		return
	}
	reportData["total_suppliers"] = activeSuppliers
	if err := h.repo.db.GetContext(c.Request.Context(), &inactiveSuppliers,
		"SELECT COUNT(*) FROM suppliers WHERE is_active = false"); err == nil {
		reportData["inactive_suppliers"] = inactiveSuppliers
	} else {
		reportData["inactive_suppliers"] = 0
	}

	var purchaseTotals struct {
		Total float64 `db:"total"`
		Paid  float64 `db:"paid"`
		Open  float64 `db:"open"`
	}
	if err := h.repo.db.GetContext(c.Request.Context(), &purchaseTotals,
		`SELECT COALESCE(SUM(total_amount), 0) AS total,
		        COALESCE(SUM(paid_amount), 0) AS paid,
		        COALESCE(SUM(total_amount - paid_amount), 0) AS open
		 FROM purchases
		 WHERE LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`); err == nil {
		reportData["total_purchases"] = purchaseTotals.Total
		reportData["total_paid"] = purchaseTotals.Paid
		reportData["total_outstanding"] = purchaseTotals.Open
	} else {
		reportData["total_purchases"] = 0
		reportData["total_paid"] = 0
		reportData["total_outstanding"] = 0
	}

	bySupplier := []map[string]interface{}{}
	rows, err := h.repo.db.QueryContext(c.Request.Context(), `
			SELECT COALESCE(s.name, 'مورد غير معروف'),
				COALESCE(SUM(p.total_amount), 0),
				COALESCE(SUM(p.paid_amount), 0),
				COALESCE(SUM(p.total_amount - p.paid_amount), 0)
			FROM purchases p
			LEFT JOIN suppliers s ON s.id = p.supplier_id
			WHERE LOWER(COALESCE(p.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY s.name
			HAVING SUM(p.total_amount) > 0
			ORDER BY SUM(p.total_amount) DESC`)
	if err == nil {
		for rows.Next() {
			var name string
			var total, paid, outstanding float64
			if err := rows.Scan(&name, &total, &paid, &outstanding); err == nil {
				bySupplier = append(bySupplier, map[string]interface{}{
					"supplier_name": name, "total_purchases": total,
					"total_paid": paid, "outstanding": outstanding,
				})
			}
		}
		rows.Close()
	}
	reportData["by_supplier"] = bySupplier

	suppliersWithBalance := []map[string]interface{}{}
	rows, err = h.repo.db.QueryContext(c.Request.Context(),
		`SELECT s.id, s.name,
		        COALESCE(SUM(p.total_amount), 0) AS total_purchases,
		        COALESCE(SUM(p.paid_amount), 0) AS total_paid,
		        COALESCE(SUM(p.total_amount - p.paid_amount), 0) AS balance
				FROM suppliers s
				JOIN purchases p ON p.supplier_id = s.id
				WHERE LOWER(COALESCE(p.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY s.id, s.name
		 HAVING SUM(p.total_amount - p.paid_amount) > 0
		 ORDER BY balance DESC`)
	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var id uuid.UUID
			var name string
			var totalPurchases float64
			var totalPaid float64
			var balance float64
			if err := rows.Scan(&id, &name, &totalPurchases, &totalPaid, &balance); err != nil {
				continue
			}
			suppliersWithBalance = append(suppliersWithBalance, map[string]interface{}{
				"id": id, "name": name, "total_purchases": totalPurchases,
				"total_paid": totalPaid, "balance": balance,
			})
		}
	}
	reportData["suppliers_with_balance"] = suppliersWithBalance

	c.JSON(http.StatusOK, gin.H{"data": reportData})
}
