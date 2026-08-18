package warranty

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles HTTP requests for warranty
type Handler struct {
	service *Service
}

// NewHandler creates a new warranty handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateWarrantyClaim handles warranty claim creation
func (h *Handler) CreateWarrantyClaim(c *gin.Context) {
	var req WarrantyClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.BadRequest(c, "user_id required")
		return
	}

	claim, err := h.service.CreateWarrantyClaim(c.Request.Context(), organizationID.(uuid.UUID), userID.(uuid.UUID), &req)
	if err != nil {
		switch err {
		case ErrWarrantyExpired:
			response.BadRequest(c, "warranty has expired")
		case ErrInvalidClaimType:
			response.BadRequest(c, "invalid claim type")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Created(c, claim, "Warranty claim created successfully")
}

// GetWarrantyClaim handles warranty claim retrieval
func (h *Handler) GetWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid warranty claim id")
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	claim, err := h.service.GetWarrantyClaim(c.Request.Context(), id, organizationID.(uuid.UUID))
	if err != nil {
		if err == ErrWarrantyClaimNotFound {
			response.NotFound(c, "warranty claim not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, claim, "Warranty claim retrieved successfully")
}

// ListWarrantyClaims handles warranty claim listing
func (h *Handler) ListWarrantyClaims(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	req := WarrantyClaimListRequest{
		Page:    1,
		PerPage: 20,
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}

	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil && perPage > 0 && perPage <= 100 {
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

	req.Status = c.Query("status")
	req.Priority = c.Query("priority")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "claim_date")
	req.SortOrder = c.DefaultQuery("sort_order", "DESC")

	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			req.StartDate = &t
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			req.EndDate = &t
		}
	}

	claims, total, err := h.service.ListWarrantyClaims(c.Request.Context(), organizationID.(uuid.UUID), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, map[string]interface{}{
		"claims":     claims,
		"total":      total,
		"page":       req.Page,
		"per_page":   req.PerPage,
	}, "Warranty claims retrieved successfully")
}

// UpdateWarrantyClaim handles warranty claim update
func (h *Handler) UpdateWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid warranty claim id")
		return
	}

	var req WarrantyClaimUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	claim, err := h.service.UpdateWarrantyClaim(c.Request.Context(), id, organizationID.(uuid.UUID), &req)
	if err != nil {
		switch err {
		case ErrWarrantyClaimNotFound:
			response.NotFound(c, "warranty claim not found")
		case ErrClaimAlreadyCompleted:
			response.BadRequest(c, "claim already completed")
		case ErrInvalidClaimStatus:
			response.BadRequest(c, "invalid claim status transition")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, claim, "Warranty claim updated successfully")
}

// ApproveWarrantyClaim handles warranty claim approval
func (h *Handler) ApproveWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid warranty claim id")
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	claim, err := h.service.ApproveWarrantyClaim(c.Request.Context(), id, organizationID.(uuid.UUID))
	if err != nil {
		switch err {
		case ErrWarrantyClaimNotFound:
			response.NotFound(c, "warranty claim not found")
		case ErrClaimAlreadyApproved:
			response.BadRequest(c, "claim already approved")
		case ErrClaimAlreadyRejected:
			response.BadRequest(c, "claim already rejected")
		case ErrClaimAlreadyCompleted:
			response.BadRequest(c, "claim already completed")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, claim, "Warranty claim approved successfully")
}

// RejectWarrantyClaim handles warranty claim rejection
func (h *Handler) RejectWarrantyClaim(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid warranty claim id")
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	claim, err := h.service.RejectWarrantyClaim(c.Request.Context(), id, organizationID.(uuid.UUID), req.Reason)
	if err != nil {
		switch err {
		case ErrWarrantyClaimNotFound:
			response.NotFound(c, "warranty claim not found")
		case ErrClaimAlreadyApproved:
			response.BadRequest(c, "claim already approved")
		case ErrClaimAlreadyRejected:
			response.BadRequest(c, "claim already rejected")
		case ErrClaimAlreadyCompleted:
			response.BadRequest(c, "claim already completed")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, claim, "Warranty claim rejected successfully")
}

// CreateWarranty handles warranty creation
func (h *Handler) CreateWarranty(c *gin.Context) {
	var req WarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	warranty, err := h.service.CreateWarranty(c.Request.Context(), organizationID.(uuid.UUID), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, warranty, "Warranty created successfully")
}

// GetWarranty handles warranty retrieval
func (h *Handler) GetWarranty(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid warranty id")
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	warranty, err := h.service.GetWarranty(c.Request.Context(), id, organizationID.(uuid.UUID))
	if err != nil {
		if err == ErrWarrantyNotFound {
			response.NotFound(c, "warranty not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, warranty, "Warranty retrieved successfully")
}

// GetWarrantyClaimsSummary handles warranty claims summary retrieval
func (h *Handler) GetWarrantyClaimsSummary(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	summary, err := h.service.GetWarrantyClaimsSummary(c.Request.Context(), organizationID.(uuid.UUID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, summary, "Warranty claims summary retrieved successfully")
}