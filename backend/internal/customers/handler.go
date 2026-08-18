package customers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles HTTP requests for customers
type Handler struct {
	service *Service
}

// NewHandler creates a new customer handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateCustomer handles customer creation
func (h *Handler) CreateCustomer(c *gin.Context) {
	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	customer, err := h.service.CreateCustomer(c.Request.Context(), organizationID, &req)
	if err != nil {
		if err == ErrCustomerCodeExists {
			response.Error(c, http.StatusConflict, http.StatusConflict, "Customer code already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create customer", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, customer, "Customer created successfully")
}

// GetCustomer handles customer retrieval
func (h *Handler) GetCustomer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid customer ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	customer, err := h.service.GetCustomer(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Customer not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve customer", err.Error())
		return
	}

	response.Success(c, http.StatusOK, customer, "Customer retrieved successfully")
}

// ListCustomers handles customer listing
func (h *Handler) ListCustomers(c *gin.Context) {
	var req CustomerListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	customers, total, err := h.service.ListCustomers(c.Request.Context(), organizationID, req.Page, req.PerPage, req.Search, req.IsActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve customers", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, customers, total, req.Page, req.PerPage, "Customers retrieved successfully")
}

// UpdateCustomer handles customer update
func (h *Handler) UpdateCustomer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid customer ID", err.Error())
		return
	}

	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	customer, err := h.service.UpdateCustomer(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		if err == ErrCustomerNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Customer not found", err.Error())
			return
		}
		if err == ErrCustomerCodeExists {
			response.Error(c, http.StatusConflict, http.StatusConflict, "Customer code already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to update customer", err.Error())
		return
	}

	response.Success(c, http.StatusOK, customer, "Customer updated successfully")
}

// DeleteCustomer handles customer deletion
func (h *Handler) DeleteCustomer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid customer ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.DeleteCustomer(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Customer not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to delete customer", err.Error())
		return
	}

	response.Success(c, http.StatusOK, nil, "Customer deleted successfully")
}

// GetCustomerLedger handles customer ledger retrieval
func (h *Handler) GetCustomerLedger(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid customer ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	ledger, err := h.service.GetCustomerLedger(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Customer not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve customer ledger", err.Error())
		return
	}

	response.Success(c, http.StatusOK, ledger, "Customer ledger retrieved successfully")
}

// AddPayment handles payment addition
func (h *Handler) AddPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid customer ID", err.Error())
		return
	}

	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	payment, err := h.service.AddPayment(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		if err == ErrCustomerNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Customer not found", err.Error())
			return
		}
		if err == ErrPaymentAmountInvalid {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid payment amount", err.Error())
			return
		}
		if err == ErrPaymentExceedsBalance {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Payment amount exceeds customer balance", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to add payment", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, payment, "Payment added successfully")
}
