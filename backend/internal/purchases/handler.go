package purchases

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
)

// Handler handles HTTP requests for purchases
type Handler struct {
	service *Service
}

// NewHandler creates a new purchase handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreatePurchase handles purchase creation
// @Summary Create a new purchase
// @Description Create a new purchase with items
// @Tags purchases
// @Accept json
// @Produce json
// @Param request body PurchaseRequest true "Purchase request"
// @Success 201 {object} PurchaseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases [post]
func (h *Handler) CreatePurchase(c *gin.Context) {
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.CreatePurchase(c.Request.Context(), organizationID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetPurchase handles getting a purchase by ID
// @Summary Get a purchase
// @Description Get a purchase by ID
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path string true "Purchase ID"
// @Success 200 {object} PurchaseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/{id} [get]
func (h *Handler) GetPurchase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.GetPurchase(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListPurchases handles listing purchases
// @Summary List purchases
// @Description List purchases with pagination and filters
// @Tags purchases
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param supplier_id query string false "Supplier ID filter"
// @Param status query string false "Status filter" Enums(pending, received, cancelled)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param search query string false "Search in invoice number and notes"
// @Param sort_by query string false "Sort by field" default(purchase_date)
// @Param sort_order query string false "Sort order" default(DESC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases [get]
func (h *Handler) ListPurchases(c *gin.Context) {
	var req PurchaseListRequest
	
	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}
	
	if supplierID := c.Query("supplier_id"); supplierID != "" {
		if id, err := uuid.Parse(supplierID); err == nil {
			req.SupplierID = &id
		}
	}
	
	req.Status = c.Query("status")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "purchase_date")
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

	purchases, total, err := h.service.ListPurchases(c.Request.Context(), organizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": purchases,
		"meta": gin.H{
			"page":      req.Page,
			"per_page":  req.PerPage,
			"total":     total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// UpdatePurchase handles updating a purchase
// @Summary Update a purchase
// @Description Update a purchase by ID
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path string true "Purchase ID"
// @Param request body PurchaseUpdateRequest true "Purchase update request"
// @Success 200 {object} PurchaseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/{id} [put]
func (h *Handler) UpdatePurchase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase ID"})
		return
	}

	var req PurchaseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.UpdatePurchase(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeletePurchase handles deleting a purchase
// @Summary Delete a purchase
// @Description Delete a purchase by ID
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path string true "Purchase ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/{id} [delete]
func (h *Handler) DeletePurchase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeletePurchase(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ReceivePurchase handles marking a purchase as received
// @Summary Receive a purchase
// @Description Mark a purchase as received
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path string true "Purchase ID"
// @Success 200 {object} PurchaseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/{id}/receive [post]
func (h *Handler) ReceivePurchase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.ReceivePurchase(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelPurchase handles cancelling a purchase
// @Summary Cancel a purchase
// @Description Cancel a purchase
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path string true "Purchase ID"
// @Success 200 {object} PurchaseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/{id}/cancel [post]
func (h *Handler) CancelPurchase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.CancelPurchase(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AddPayment handles adding a payment to a purchase
// @Summary Add payment to purchase
// @Description Add a payment to a purchase
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path string true "Purchase ID"
// @Param amount body number true "Payment amount"
// @Success 200 {object} PurchaseResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/{id}/payment [post]
func (h *Handler) AddPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase ID"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,min=0.01"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.AddPayment(c.Request.Context(), id, organizationID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AddPurchaseItem handles adding an item to a purchase
// @Summary Add item to purchase
// @Description Add an item to a purchase
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path string true "Purchase ID"
// @Param request body PurchaseItemRequest true "Purchase item request"
// @Success 201 {object} PurchaseItem
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/{id}/items [post]
func (h *Handler) AddPurchaseItem(c *gin.Context) {
	purchaseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase ID"})
		return
	}

	var req PurchaseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	item, err := h.service.AddPurchaseItem(c.Request.Context(), purchaseID, organizationID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdatePurchaseItem handles updating a purchase item
// @Summary Update purchase item
// @Description Update a purchase item
// @Tags purchases
// @Accept json
// @Produce json
// @Param item_id path string true "Purchase Item ID"
// @Param request body PurchaseItemRequest true "Purchase item request"
// @Success 200 {object} PurchaseItem
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/items/{item_id} [put]
func (h *Handler) UpdatePurchaseItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	var req PurchaseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	item, err := h.service.UpdatePurchaseItem(c.Request.Context(), itemID, organizationID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeletePurchaseItem handles deleting a purchase item
// @Summary Delete purchase item
// @Description Delete a purchase item
// @Tags purchases
// @Accept json
// @Produce json
// @Param item_id path string true "Purchase Item ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/purchases/items/{item_id} [delete]
func (h *Handler) DeletePurchaseItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	if err := h.service.DeletePurchaseItem(c.Request.Context(), itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
