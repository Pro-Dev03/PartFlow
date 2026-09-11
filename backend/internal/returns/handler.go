package returns

import (
	stderrors "errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/partflow/smart-store/pkg/errors"
	"github.com/partflow/smart-store/pkg/middleware"
)

// Handler handles HTTP requests for returns
type Handler struct {
	service *Service
}

// NewHandler creates a new return handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateReturn handles return creation
// @Summary Create a new return
// @Description Create a new return with items
// @Tags returns
// @Accept json
// @Produce json
// @Param request body ReturnRequest true "Return request"
// @Success 201 {object} ReturnResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns [post]
func (h *Handler) CreateReturn(c *gin.Context) {
	var req ReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	userID := middleware.GetUserID(c)

	response, err := h.service.CreateReturn(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, ErrInsufficientStock) {
			apperrors.HandleError(c, apperrors.NewConflictError("return quantity exceeds the quantity sold", err))
			return
		}
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to create return"))
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetReturn handles getting a return by ID
// @Summary Get a return
// @Description Get a return by ID
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Success 200 {object} ReturnResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id} [get]
func (h *Handler) GetReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	response, err := h.service.GetReturn(c.Request.Context(), id)
	if err != nil {
		apperrors.HandleError(c, apperrors.NewNotFoundError("Return", err))
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListReturns handles listing returns
// @Summary List returns
// @Description List returns with pagination and filters
// @Tags returns
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param customer_id query string false "Customer ID filter"
// @Param sale_id query string false "Sale ID filter"
// @Param status query string false "Status filter" Enums(PENDING, APPROVED, PROCESSING, COMPLETED, REJECTED, CANCELLED)
// @Param return_type query string false "Return type filter" Enums(FULL, PARTIAL, QUANTITY_PARTIAL)
// @Param refund_method query string false "Refund method filter" Enums(CASH, DEBT_ADJUSTMENT)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param search query string false "Search in return number, reference number, reason, notes"
// @Param sort_by query string false "Sort by field" default(return_date)
// @Param sort_order query string false "Sort order" default(DESC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns [get]
func (h *Handler) ListReturns(c *gin.Context) {
	var req ReturnListRequest

	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}

	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := uuid.Parse(customerID); err == nil {
			req.CustomerID = &id
		}
	}

	if saleID := c.Query("sale_id"); saleID != "" {
		if id, err := uuid.Parse(saleID); err == nil {
			req.SaleID = &id
		}
	}

	req.Status = c.Query("status")
	req.ReturnType = c.Query("return_type")
	req.RefundMethod = c.Query("refund_method")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "return_date")
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

	returns, total, err := h.service.ListReturns(c.Request.Context(), req)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to list returns"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": returns,
		"meta": gin.H{
			"page":        req.Page,
			"per_page":    req.PerPage,
			"total":       total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// UpdateReturn handles updating a return
// @Summary Update a return
// @Description Update a return by ID
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Param request body ReturnUpdateRequest true "Return update request"
// @Success 200 {object} ReturnResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id} [put]
func (h *Handler) UpdateReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	var req ReturnUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	response, err := h.service.UpdateReturn(c.Request.Context(), id, &req)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to update return"))
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteReturn handles deleting a return
// @Summary Delete a return
// @Description Delete a return by ID
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id} [delete]
func (h *Handler) DeleteReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	if err := h.service.DeleteReturn(c.Request.Context(), id); err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	c.Status(http.StatusNoContent)
}

// ApproveReturn handles approving a return
// @Summary Approve a return
// @Description Approve a return
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Success 200 {object} ReturnResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id}/approve [post]
func (h *Handler) ApproveReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	response, err := h.service.ApproveReturn(c.Request.Context(), id)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to approve return"))
		return
	}

	c.JSON(http.StatusOK, response)
}

// RejectReturn handles rejecting a return
// @Summary Reject a return
// @Description Reject a return
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Success 200 {object} ReturnResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id}/reject [post]
func (h *Handler) RejectReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	response, err := h.service.RejectReturn(c.Request.Context(), id)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to process return"))
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessRefund handles processing refund for a return
// @Summary Process refund
// @Description Process refund for a return
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Success 200 {object} ReturnResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id}/refund [post]
func (h *Handler) ProcessRefund(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	response, err := h.service.ProcessRefund(c.Request.Context(), id)
	if err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	c.JSON(http.StatusOK, response)
}

// AddReturnItem handles adding an item to a return
// @Summary Add item to return
// @Description Add an item to a return
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Param request body ReturnItemRequest true "Return item request"
// @Success 201 {object} ReturnItem
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id}/items [post]
func (h *Handler) AddReturnItem(c *gin.Context) {
	returnID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	var req ReturnItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to create return item"))
		return
	}

	item, err := h.service.AddReturnItem(c.Request.Context(), returnID, req)
	if err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateReturnItem handles updating a return item
// @Summary Update return item
// @Description Update a return item
// @Tags returns
// @Accept json
// @Produce json
// @Param item_id path string true "Return Item ID"
// @Param request body ReturnItemRequest true "Return item request"
// @Success 200 {object} ReturnItem
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/items/{item_id} [put]
func (h *Handler) UpdateReturnItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid item ID", err))
		return
	}

	var req ReturnItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	item, err := h.service.UpdateReturnItem(c.Request.Context(), itemID, req)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to update return item"))
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteReturnItem handles deleting a return item
// @Summary Delete return item
// @Description Delete a return item
// @Tags returns
// @Accept json
// @Produce json
// @Param item_id path string true "Return Item ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/items/{item_id} [delete]
func (h *Handler) DeleteReturnItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid item ID", err))
		return
	}

	if err := h.service.DeleteReturnItem(c.Request.Context(), itemID); err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to delete return item"))
		return
	}

	c.Status(http.StatusNoContent)
}

// CompleteReturn handles completing a return
// @Summary Complete a return
// @Description Complete a return and process financial effects
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Success 200 {object} ReturnResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id}/complete [post]
func (h *Handler) CompleteReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	userID := middleware.GetUserID(c)

	response, err := h.service.CompleteReturn(c.Request.Context(), id, userID)
	if err != nil {
		if stderrors.Is(err, ErrReturnAlreadyCompleted) {
			apperrors.HandleError(c, apperrors.NewConflictError("return already completed", err))
			return
		}
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to complete return"))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetReturnsBySale handles getting returns for a specific sale
// @Summary Get returns by sale
// @Description Get all returns for a specific sale
// @Tags returns
// @Accept json
// @Produce json
// @Param sale_id path string true "Sale ID"
// @Success 200 {array} Return
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/sale/{sale_id} [get]
func (h *Handler) GetReturnsBySale(c *gin.Context) {
	saleID, err := uuid.Parse(c.Param("sale_id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid sale ID", err))
		return
	}

	returns, err := h.service.GetReturnBySale(c.Request.Context(), saleID)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to retrieve sale returns"))
		return
	}

	c.JSON(http.StatusOK, returns)
}

// GetMonthlyReturnsAnalysis handles getting monthly returns analysis
// @Summary Get monthly returns analysis
// @Description Get monthly returns analysis data
// @Tags returns
// @Accept json
// @Produce json
// @Success 200 {array} MonthlyReturnsAnalysis
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/analysis/monthly [get]
func (h *Handler) GetMonthlyReturnsAnalysis(c *gin.Context) {
	analysis, err := h.service.GetMonthlyReturnsAnalysis(c.Request.Context())
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to retrieve return analysis"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": analysis})
}

// GetSalesReturnsAnalysis handles getting sales vs returns analysis
// @Summary Get sales vs returns analysis
// @Description Get sales vs returns analysis data
// @Tags returns
// @Accept json
// @Produce json
// @Success 200 {array} SalesReturnsAnalysis
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/analysis/sales-returns [get]
func (h *Handler) GetSalesReturnsAnalysis(c *gin.Context) {
	analysis, err := h.service.GetSalesReturnsAnalysis(c.Request.Context())
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to retrieve return trends"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": analysis})
}

// ValidateReturnQuantity handles validating return quantity against original sale
// @Summary Validate return quantity
// @Description Validate return quantity against original sale
// @Tags returns
// @Accept json
// @Produce json
// @Param sale_item_id path string true "Sale Item ID"
// @Param quantity query int true "Quantity to return"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/validate/{sale_item_id} [get]
func (h *Handler) ValidateReturnQuantity(c *gin.Context) {
	saleItemID, err := uuid.Parse(c.Param("sale_item_id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid sale item ID", err))
		return
	}

	quantity, err := strconv.Atoi(c.Query("quantity"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid quantity", nil))
		return
	}

	err = h.service.ValidateReturnQuantity(c.Request.Context(), saleItemID, quantity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// GetReturnSummary handles getting return summary
// @Summary Get return summary
// @Description Get summary of returns with related data
// @Tags returns
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/summary [get]
func (h *Handler) GetReturnSummary(c *gin.Context) {
	summary, err := h.service.GetReturnSummary(c.Request.Context())
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to retrieve return summary"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// ReverseReturn handles reversing a return (instead of deleting)
// @Summary Reverse a return
// @Description Reverse a return instead of deleting it
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Success 200 {object} ReturnResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id}/reverse [post]
func (h *Handler) ReverseReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	userID := middleware.GetUserID(c)

	response, err := h.service.ReverseReturn(c.Request.Context(), id, userID)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to validate return"))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetReturnsByCustomer handles getting returns for a specific customer
// @Summary Get returns by customer
// @Description Get all returns for a specific customer
// @Tags returns
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Success 200 {array} Return
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/customer/{customer_id} [get]
func (h *Handler) GetReturnsByCustomer(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid customer ID", err))
		return
	}

	returns, err := h.service.GetReturnsByCustomer(c.Request.Context(), customerID)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to retrieve customer returns"))
		return
	}

	c.JSON(http.StatusOK, returns)
}

// GetPendingReturns handles getting returns that are pending approval
// @Summary Get pending returns
// @Description Get all returns that are pending approval
// @Tags returns
// @Accept json
// @Produce json
// @Success 200 {array} Return
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/pending [get]
func (h *Handler) GetPendingReturns(c *gin.Context) {
	returns, err := h.service.GetPendingReturns(c.Request.Context())
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to retrieve customer returns"))
		return
	}

	c.JSON(http.StatusOK, returns)
}

// GetReturnStatistics handles getting return statistics
// @Summary Get return statistics
// @Description Get statistics about returns for the last 30 days
// @Tags returns
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/statistics [get]
func (h *Handler) GetReturnStatistics(c *gin.Context) {
	stats, err := h.service.GetReturnStatistics(c.Request.Context())
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to retrieve return statistics"))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetReturnWithItems handles getting a return with all its items
// @Summary Get return with items
// @Description Get a return with all its items
// @Tags returns
// @Accept json
// @Produce json
// @Param id path string true "Return ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/returns/{id}/with-items [get]
func (h *Handler) GetReturnWithItems(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("invalid return ID", err))
		return
	}

	returnRecord, items, err := h.service.GetReturnWithItems(c.Request.Context(), id)
	if err != nil {
		apperrors.HandleError(c, apperrors.NewNotFoundError("Return", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"return": returnRecord,
		"items":  items,
	})
}
