package payments

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles HTTP requests for payments
type Handler struct {
	service *Service
}

// NewHandler creates a new payment handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreatePayment handles payment creation
func (h *Handler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
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

	payment, err := h.service.CreatePayment(c.Request.Context(), organizationID.(uuid.UUID), userID.(uuid.UUID), &req)
	if err != nil {
		switch err {
		case ErrInvalidPaymentType, ErrInvalidAmount:
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Created(c, payment, "Payment created successfully")
}

// GetPayment handles payment retrieval
func (h *Handler) GetPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	payment, err := h.service.GetPayment(c.Request.Context(), id, organizationID.(uuid.UUID))
	if err != nil {
		if err == ErrPaymentNotFound {
			response.NotFound(c, "payment not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, payment, "Payment retrieved successfully")
}

// ListPayments handles payment listing
func (h *Handler) ListPayments(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	page := 1
	perPage := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	filters := make(map[string]interface{})
	if paymentType := c.Query("type"); paymentType != "" {
		filters["type"] = paymentType
	}
	if referenceID := c.Query("reference_id"); referenceID != "" {
		if id, err := uuid.Parse(referenceID); err == nil {
			filters["reference_id"] = id
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if method := c.Query("method"); method != "" {
		filters["method"] = method
	}

	payments, total, err := h.service.ListPayments(c.Request.Context(), organizationID.(uuid.UUID), page, perPage, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, map[string]interface{}{
		"payments": payments,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	}, "Payments retrieved successfully")
}

// UpdatePayment handles payment update
func (h *Handler) UpdatePayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	var req UpdatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	payment, err := h.service.UpdatePayment(c.Request.Context(), id, organizationID.(uuid.UUID), &req)
	if err != nil {
		switch err {
		case ErrPaymentNotFound:
			response.NotFound(c, "payment not found")
		case ErrPaymentAlreadyProcessed:
			response.BadRequest(c, "payment already processed")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, payment, "Payment updated successfully")
}

// DeletePayment handles payment deletion
func (h *Handler) DeletePayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	if err := h.service.DeletePayment(c.Request.Context(), id, organizationID.(uuid.UUID)); err != nil {
		switch err {
		case ErrPaymentNotFound:
			response.NotFound(c, "payment not found")
		case ErrPaymentCannotBeCancelled:
			response.BadRequest(c, "payment cannot be cancelled")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, map[string]interface{}{"message": "payment deleted successfully"}, "Payment deleted successfully")
}

// CompletePayment handles marking a payment as completed
func (h *Handler) CompletePayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	payment, err := h.service.CompletePayment(c.Request.Context(), id, organizationID.(uuid.UUID))
	if err != nil {
		switch err {
		case ErrPaymentNotFound:
			response.NotFound(c, "payment not found")
		case ErrPaymentAlreadyProcessed:
			response.BadRequest(c, "payment already processed")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, payment, "Payment completed successfully")
}

// CancelPayment handles cancelling a payment
func (h *Handler) CancelPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	payment, err := h.service.CancelPayment(c.Request.Context(), id, organizationID.(uuid.UUID))
	if err != nil {
		switch err {
		case ErrPaymentNotFound:
			response.NotFound(c, "payment not found")
		case ErrPaymentCannotBeCancelled, ErrPaymentAlreadyProcessed:
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, payment, "Payment cancelled successfully")
}

// GetPaymentSummary handles payment summary retrieval
func (h *Handler) GetPaymentSummary(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	summary, err := h.service.GetPaymentSummary(c.Request.Context(), organizationID.(uuid.UUID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, summary, "Payment summary retrieved successfully")
}