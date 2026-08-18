package warranties

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
)

// Handler handles HTTP requests for warranties
type Handler struct {
	service *Service
}

// NewHandler creates a new warranty handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateWarranty handles warranty creation
// @Summary Create a new warranty
// @Description Create a new warranty
// @Tags warranties
// @Accept json
// @Produce json
// @Param request body WarrantyRequest true "Warranty request"
// @Success 201 {object} WarrantyResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranties [post]
func (h *Handler) CreateWarranty(c *gin.Context) {
	var req WarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.CreateWarranty(c.Request.Context(), organizationID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetWarranty handles getting a warranty by ID
// @Summary Get a warranty
// @Description Get a warranty by ID
// @Tags warranties
// @Accept json
// @Produce json
// @Param id path string true "Warranty ID"
// @Success 200 {object} WarrantyResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranties/{id} [get]
func (h *Handler) GetWarranty(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.GetWarranty(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListWarranties handles listing warranties
// @Summary List warranties
// @Description List warranties with pagination and filters
// @Tags warranties
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param product_id query string false "Product ID filter"
// @Param customer_id query string false "Customer ID filter"
// @Param sale_id query string false "Sale ID filter"
// @Param status query string false "Status filter" Enums(active, expired, claimed, voided)
// @Param warranty_type query string false "Warranty type filter" Enums(manufacturer, seller, extended)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param search query string false "Search in warranty number, serial number, notes"
// @Param sort_by query string false "Sort by field" default(end_date)
// @Param sort_order query string false "Sort order" default(ASC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranties [get]
func (h *Handler) ListWarranties(c *gin.Context) {
	var req WarrantyListRequest
	
	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}
	
	if productID := c.Query("product_id"); productID != "" {
		if id, err := uuid.Parse(productID); err == nil {
			req.ProductID = &id
		}
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
	req.WarrantyType = c.Query("warranty_type")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "end_date")
	req.SortOrder = c.DefaultQuery("sort_order", "ASC")
	
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

	warranties, total, err := h.service.ListWarranties(c.Request.Context(), organizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": warranties,
		"meta": gin.H{
			"page":        req.Page,
			"per_page":    req.PerPage,
			"total":       total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// UpdateWarranty handles updating a warranty
// @Summary Update a warranty
// @Description Update a warranty by ID
// @Tags warranties
// @Accept json
// @Produce json
// @Param id path string true "Warranty ID"
// @Param request body WarrantyUpdateRequest true "Warranty update request"
// @Success 200 {object} WarrantyResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranties/{id} [put]
func (h *Handler) UpdateWarranty(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty ID"})
		return
	}

	var req WarrantyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.UpdateWarranty(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteWarranty handles deleting a warranty
// @Summary Delete a warranty
// @Description Delete a warranty by ID
// @Tags warranties
// @Accept json
// @Produce json
// @Param id path string true "Warranty ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranties/{id} [delete]
func (h *Handler) DeleteWarranty(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeleteWarranty(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateWarrantyClaim handles warranty claim creation
// @Summary Create a new warranty claim
// @Description Create a new warranty claim
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Param request body WarrantyClaimRequest true "Warranty claim request"
// @Success 201 {object} WarrantyClaimResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranty-claims [post]
func (h *Handler) CreateWarrantyClaim(c *gin.Context) {
	var req WarrantyClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.CreateWarrantyClaim(c.Request.Context(), organizationID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetWarrantyClaim handles getting a warranty claim by ID
// @Summary Get a warranty claim
// @Description Get a warranty claim by ID
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Param id path string true "Warranty Claim ID"
// @Success 200 {object} WarrantyClaimResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranty-claims/{id} [get]
func (h *Handler) GetWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty claim ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.GetWarrantyClaim(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListWarrantyClaims handles listing warranty claims
// @Summary List warranty claims
// @Description List warranty claims with pagination and filters
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param warranty_id query string false "Warranty ID filter"
// @Param customer_id query string false "Customer ID filter"
// @Param status query string false "Status filter" Enums(pending, approved, rejected, in_progress, completed)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param search query string false "Search in claim number, issue description, notes"
// @Param sort_by query string false "Sort by field" default(claim_date)
// @Param sort_order query string false "Sort order" default(DESC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranty-claims [get]
func (h *Handler) ListWarrantyClaims(c *gin.Context) {
	var req WarrantyClaimListRequest
	
	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}
	
	if warrantyID := c.Query("warranty_id"); warrantyID != "" {
		if id, err := uuid.Parse(warrantyID); err == nil {
			req.WarrantyID = &id
		}
	}
	
	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := uuid.Parse(customerID); err == nil {
			req.CustomerID = &id
		}
	}
	
	req.Status = c.Query("status")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "claim_date")
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

	claims, total, err := h.service.ListWarrantyClaims(c.Request.Context(), organizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": claims,
		"meta": gin.H{
			"page":        req.Page,
			"per_page":    req.PerPage,
			"total":       total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// UpdateWarrantyClaim handles updating a warranty claim
// @Summary Update a warranty claim
// @Description Update a warranty claim by ID
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Param id path string true "Warranty Claim ID"
// @Param request body WarrantyClaimUpdateRequest true "Warranty claim update request"
// @Success 200 {object} WarrantyClaimResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranty-claims/{id} [put]
func (h *Handler) UpdateWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty claim ID"})
		return
	}

	var req WarrantyClaimUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.UpdateWarrantyClaim(c.Request.Context(), id, organizationID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteWarrantyClaim handles deleting a warranty claim
// @Summary Delete a warranty claim
// @Description Delete a warranty claim by ID
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Param id path string true "Warranty Claim ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranty-claims/{id} [delete]
func (h *Handler) DeleteWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty claim ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeleteWarrantyClaim(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetWarrantiesExpiringSoon handles getting warranties expiring soon
// @Summary Get warranties expiring soon
// @Description Get warranties that will expire within specified days
// @Tags warranties
// @Accept json
// @Produce json
// @Param days query int false "Days until expiration" default(30)
// @Success 200 {array} WarrantyExpiringSoon
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranties/expiring-soon [get]
func (h *Handler) GetWarrantiesExpiringSoon(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
	if err != nil || days <= 0 {
		days = 30
	}

	organizationID := middleware.GetOrganizationID(c)

	warranties, err := h.service.GetWarrantiesExpiringSoon(c.Request.Context(), organizationID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, warranties)
}

// ApproveWarrantyClaim handles approving a warranty claim
// @Summary Approve a warranty claim
// @Description Approve a warranty claim
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Param id path string true "Warranty Claim ID"
// @Success 200 {object} WarrantyClaimResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranty-claims/{id}/approve [post]
func (h *Handler) ApproveWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty claim ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.ApproveWarrantyClaim(c.Request.Context(), id, organizationID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RejectWarrantyClaim handles rejecting a warranty claim
// @Summary Reject a warranty claim
// @Description Reject a warranty claim
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Param id path string true "Warranty Claim ID"
// @Success 200 {object} WarrantyClaimResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranty-claims/{id}/reject [post]
func (h *Handler) RejectWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty claim ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.RejectWarrantyClaim(c.Request.Context(), id, organizationID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CompleteWarrantyClaim handles completing a warranty claim
// @Summary Complete a warranty claim
// @Description Complete a warranty claim with resolution
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Param id path string true "Warranty Claim ID"
// @Param resolution body string true "Resolution description"
// @Success 200 {object} WarrantyClaimResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/warranty-claims/{id}/complete [post]
func (h *Handler) CompleteWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty claim ID"})
		return
	}

	var req struct {
		Resolution string `json:"resolution" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.CompleteWarrantyClaim(c.Request.Context(), id, organizationID, userID, req.Resolution)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
