package returns

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.CreateReturn(c.Request.Context(), organizationID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.GetReturn(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
// @Param status query string false "Status filter" Enums(pending, approved, rejected, completed)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param search query string false "Search in return number, reason, notes"
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

	organizationID := middleware.GetOrganizationID(c)

	returns, total, err := h.service.ListReturns(c.Request.Context(), organizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return ID"})
		return
	}

	var req ReturnUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.UpdateReturn(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeleteReturn(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.ApproveReturn(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.RejectReturn(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.ProcessRefund(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return ID"})
		return
	}

	var req ReturnItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	item, err := h.service.AddReturnItem(c.Request.Context(), returnID, organizationID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	var req ReturnItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	item, err := h.service.UpdateReturnItem(c.Request.Context(), itemID, organizationID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	if err := h.service.DeleteReturnItem(c.Request.Context(), itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
