package acquisitions

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/partflow/smart-store/pkg/errors"
)

// Handler handles HTTP requests for acquisitions
type Handler struct {
	service *Service
}

// NewHandler creates a new acquisitions handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateAcquisition handles POST /api/v1/acquisitions
func (h *Handler) CreateAcquisition(c *gin.Context) {
	var req AcquisitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	// Get user ID from context (assuming middleware sets it)
	userID, exists := c.Get("user_id")
	if !exists {
		apperrors.HandleError(c, apperrors.NewUnauthorizedError("User not authenticated", nil))
		return
	}

	acquisition, err := h.service.CreateAcquisition(c.Request.Context(), &req, userID.(uuid.UUID))
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to create acquisition"))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": acquisition})
}

// GetAcquisition handles GET /api/v1/acquisitions/:id
func (h *Handler) GetAcquisition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("Invalid acquisition ID", err))
		return
	}

	response, err := h.service.GetAcquisitionWithItems(c.Request.Context(), id)
	if err != nil {
		apperrors.HandleError(c, apperrors.NewNotFoundError("Acquisition", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// ListAcquisitions handles GET /api/v1/acquisitions
func (h *Handler) ListAcquisitions(c *gin.Context) {
	req := &AcquisitionListRequest{
		Page:    1,
		PerPage: 20,
	}

	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}

	req.Type = c.Query("type")
	req.Status = c.Query("status")
	req.PaymentStatus = c.Query("payment_status")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "acquisition_date")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	if supplierID := c.Query("supplier_id"); supplierID != "" {
		if id, err := uuid.Parse(supplierID); err == nil {
			req.SupplierID = &id
		}
	}

	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := uuid.Parse(customerID); err == nil {
			req.CustomerID = &id
		}
	}

	acquisitions, total, err := h.service.ListAcquisitions(c.Request.Context(), req)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to list acquisitions"))
		return
	}
	if acquisitions == nil {
		acquisitions = []Acquisition{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": acquisitions,
		"meta": gin.H{
			"page":     req.Page,
			"per_page": req.PerPage,
			"total":    total,
		},
	})
}

// UpdateAcquisitionStatus handles PUT /api/v1/acquisitions/:id/status
func (h *Handler) UpdateAcquisitionStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("Invalid acquisition ID", err))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	err = h.service.UpdateAcquisitionStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to update acquisition status"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// CreateSellerPayment handles POST /api/v1/acquisitions/:id/payments
func (h *Handler) CreateSellerPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("Invalid acquisition ID", err))
		return
	}

	var req SellerPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}

	// Set acquisition ID from URL
	req.AcquisitionID = id

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		apperrors.HandleError(c, apperrors.NewUnauthorizedError("User not authenticated", nil))
		return
	}

	payment, err := h.service.CreateSellerPayment(c.Request.Context(), &req, userID.(uuid.UUID))
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to create seller payment"))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": payment})
}

// GetSellerBalances handles GET /api/v1/acquisitions/seller-balances
func (h *Handler) GetSellerBalances(c *gin.Context) {
	balances, err := h.service.GetSellerBalances(c.Request.Context())
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to retrieve seller balances"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": balances})
}

// CreateSellerBalancePayment applies a payment to the seller's oldest unpaid acquisition.
func (h *Handler) CreateSellerBalancePayment(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewValidationError("Invalid customer ID", err))
		return
	}
	var req struct {
		Amount float64 `json:"amount" binding:"required,min=0.01"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.ValidateRequest(err))
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		apperrors.HandleError(c, apperrors.NewUnauthorizedError("User not authenticated", nil))
		return
	}

	payment, err := h.service.CreateSellerBalancePayment(c.Request.Context(), customerID, req.Amount, userID.(uuid.UUID))
	if err != nil {
		apperrors.HandleError(c, apperrors.WrapError(err, "Failed to create seller balance payment"))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": payment})
}
