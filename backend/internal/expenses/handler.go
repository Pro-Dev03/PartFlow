package expenses

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
)

// Handler handles HTTP requests for expenses
type Handler struct {
	service *Service
}

// NewHandler creates a new expense handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateExpense handles expense creation
// @Summary Create a new expense
// @Description Create a new expense
// @Tags expenses
// @Accept json
// @Produce json
// @Param request body ExpenseRequest true "Expense request"
// @Success 201 {object} ExpenseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expenses [post]
func (h *Handler) CreateExpense(c *gin.Context) {
	var req ExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.CreateExpense(c.Request.Context(), organizationID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetExpense handles getting an expense by ID
// @Summary Get an expense
// @Description Get an expense by ID
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expenses/{id} [get]
func (h *Handler) GetExpense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.GetExpense(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListExpenses handles listing expenses
// @Summary List expenses
// @Description List expenses with pagination and filters
// @Tags expenses
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param category_id query string false "Category ID filter"
// @Param status query string false "Status filter" Enums(pending, approved, rejected)
// @Param payment_method query string false "Payment method filter" Enums(cash, card, bank_transfer, check)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param is_recurring query bool false "Recurring filter"
// @Param search query string false "Search in title, description, reference"
// @Param sort_by query string false "Sort by field" default(expense_date)
// @Param sort_order query string false "Sort order" default(DESC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expenses [get]
func (h *Handler) ListExpenses(c *gin.Context) {
	var req ExpenseListRequest
	
	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}
	
	if categoryID := c.Query("category_id"); categoryID != "" {
		if id, err := uuid.Parse(categoryID); err == nil {
			req.CategoryID = &id
		}
	}
	
	req.Status = c.Query("status")
	req.PaymentMethod = c.Query("payment_method")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "expense_date")
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
	
	if isRecurring := c.Query("is_recurring"); isRecurring != "" {
		if val, err := strconv.ParseBool(isRecurring); err == nil {
			req.IsRecurring = &val
		}
	}

	organizationID := middleware.GetOrganizationID(c)

	expenses, total, err := h.service.ListExpenses(c.Request.Context(), organizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": expenses,
		"meta": gin.H{
			"page":        req.Page,
			"per_page":    req.PerPage,
			"total":       total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// UpdateExpense handles updating an expense
// @Summary Update an expense
// @Description Update an expense by ID
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Param request body ExpenseUpdateRequest true "Expense update request"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expenses/{id} [put]
func (h *Handler) UpdateExpense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	var req ExpenseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.UpdateExpense(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteExpense handles deleting an expense
// @Summary Delete an expense
// @Description Delete an expense by ID
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expenses/{id} [delete]
func (h *Handler) DeleteExpense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeleteExpense(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ApproveExpense handles approving an expense
// @Summary Approve an expense
// @Description Approve an expense
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expenses/{id}/approve [post]
func (h *Handler) ApproveExpense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.ApproveExpense(c.Request.Context(), id, organizationID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RejectExpense handles rejecting an expense
// @Summary Reject an expense
// @Description Reject an expense
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} ExpenseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expenses/{id}/reject [post]
func (h *Handler) RejectExpense(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.RejectExpense(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateExpenseCategory handles expense category creation
// @Summary Create a new expense category
// @Description Create a new expense category
// @Tags expense-categories
// @Accept json
// @Produce json
// @Param request body ExpenseCategoryRequest true "Expense category request"
// @Success 201 {object} ExpenseCategory
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expense-categories [post]
func (h *Handler) CreateExpenseCategory(c *gin.Context) {
	var req ExpenseCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	category, err := h.service.CreateExpenseCategory(c.Request.Context(), organizationID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GetExpenseCategory handles getting an expense category by ID
// @Summary Get an expense category
// @Description Get an expense category by ID
// @Tags expense-categories
// @Accept json
// @Produce json
// @Param id path string true "Expense Category ID"
// @Success 200 {object} ExpenseCategory
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expense-categories/{id} [get]
func (h *Handler) GetExpenseCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense category ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	category, err := h.service.GetExpenseCategory(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// ListExpenseCategories handles listing expense categories
// @Summary List expense categories
// @Description List expense categories with pagination and filters
// @Tags expense-categories
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param is_active query bool false "Active filter"
// @Param search query string false "Search in name and description"
// @Param sort_by query string false "Sort by field" default(name)
// @Param sort_order query string false "Sort order" default(ASC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expense-categories [get]
func (h *Handler) ListExpenseCategories(c *gin.Context) {
	var req ExpenseCategoryListRequest
	
	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}
	
	if isActive := c.Query("is_active"); isActive != "" {
		if val, err := strconv.ParseBool(isActive); err == nil {
			req.IsActive = &val
		}
	}
	
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "name")
	req.SortOrder = c.DefaultQuery("sort_order", "ASC")

	organizationID := middleware.GetOrganizationID(c)

	categories, total, err := h.service.ListExpenseCategories(c.Request.Context(), organizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
		"meta": gin.H{
			"page":        req.Page,
			"per_page":    req.PerPage,
			"total":       total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// UpdateExpenseCategory handles updating an expense category
// @Summary Update an expense category
// @Description Update an expense category by ID
// @Tags expense-categories
// @Accept json
// @Produce json
// @Param id path string true "Expense Category ID"
// @Param request body ExpenseCategoryUpdateRequest true "Expense category update request"
// @Success 200 {object} ExpenseCategory
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expense-categories/{id} [put]
func (h *Handler) UpdateExpenseCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense category ID"})
		return
	}

	var req ExpenseCategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	category, err := h.service.UpdateExpenseCategory(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteExpenseCategory handles deleting an expense category
// @Summary Delete an expense category
// @Description Delete an expense category by ID
// @Tags expense-categories
// @Accept json
// @Produce json
// @Param id path string true "Expense Category ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expense-categories/{id} [delete]
func (h *Handler) DeleteExpenseCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense category ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeleteExpenseCategory(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetExpenseSummary handles getting expense summary
// @Summary Get expense summary
// @Description Get expense summary statistics
// @Tags expenses
// @Accept json
// @Produce json
// @Success 200 {object} ExpenseSummary
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/expenses/summary [get]
func (h *Handler) GetExpenseSummary(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)

	summary, err := h.service.GetExpenseSummary(c.Request.Context(), organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
