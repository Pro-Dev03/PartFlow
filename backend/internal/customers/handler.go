package customers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/errors"
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
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	customer, err := h.service.CreateCustomer(c.Request.Context(), organizationID, &req)
	if err != nil {
		if err == ErrCustomerCodeExists {
			errors.HandleError(c, errors.NewConflictError("Customer code already exists", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to create customer"))
		return
	}

	response.Success(c, http.StatusCreated, customer, "Customer created successfully")
}

// GetCustomer handles customer retrieval
func (h *Handler) GetCustomer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	customer, err := h.service.GetCustomer(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve customer"))
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
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	customer, err := h.service.UpdateCustomer(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		if err == ErrCustomerCodeExists {
			errors.HandleError(c, errors.NewConflictError("Customer code already exists", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to update customer"))
		return
	}

	response.Success(c, http.StatusOK, customer, "Customer updated successfully")
}

// DeleteCustomer handles customer deletion
func (h *Handler) DeleteCustomer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.DeleteCustomer(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to delete customer"))
		return
	}

	response.Success(c, http.StatusOK, nil, "Customer deleted successfully")
}

// GetCustomerLedger handles customer ledger retrieval
func (h *Handler) GetCustomerLedger(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	ledger, err := h.service.GetCustomerLedger(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve customer ledger"))
		return
	}

	response.Success(c, http.StatusOK, ledger, "Customer ledger retrieved successfully")
}

// AddPayment handles payment addition
func (h *Handler) AddPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	payment, err := h.service.AddPayment(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		if err == ErrPaymentAmountInvalid {
			errors.HandleError(c, errors.NewValidationError("Invalid payment amount", err))
			return
		}
		if err == ErrPaymentExceedsBalance {
			errors.HandleError(c, errors.NewBusinessError("Payment amount exceeds customer balance", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to add payment"))
		return
	}

	response.Success(c, http.StatusCreated, payment, "Payment added successfully")
}

// GetCustomerDebtSummary handles customer debt summary retrieval
func (h *Handler) GetCustomerDebtSummary(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	summary, err := h.service.GetCustomerDebtSummary(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve debt summary"))
		return
	}

	response.Success(c, http.StatusOK, summary, "Debt summary retrieved successfully")
}

// UpdateCreditLimit handles credit limit update
func (h *Handler) UpdateCreditLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	var req UpdateCreditLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.UpdateCreditLimit(c.Request.Context(), id, organizationID, req.NewLimit)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		if err == ErrCreditLimitBelowBalance {
			errors.HandleError(c, errors.NewBusinessError("Credit limit cannot be set below current balance", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to update credit limit"))
		return
	}

	response.Success(c, http.StatusOK, nil, "Credit limit updated successfully")
}

// GetOverdueCustomers handles overdue customers retrieval
func (h *Handler) GetOverdueCustomers(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)

	overdueCustomers, err := h.service.GetOverdueCustomers(c.Request.Context(), organizationID)
	if err != nil {
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve overdue customers"))
		return
	}

	response.Success(c, http.StatusOK, overdueCustomers, "Overdue customers retrieved successfully")
}

// CreateDebtEntry handles debt entry creation
func (h *Handler) CreateDebtEntry(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	var req CreateDebtEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.CreateDebtEntry(c.Request.Context(), id, organizationID, req.Amount, req.ReferenceID, req.ReferenceType, req.DueDate)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		if err == ErrCreditLimitExceeded {
			errors.HandleError(c, errors.NewBusinessError("Credit limit exceeded", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to create debt entry"))
		return
	}

	response.Success(c, http.StatusCreated, nil, "Debt entry created successfully")
}

// GetDebtEntries handles debt entries retrieval
func (h *Handler) GetDebtEntries(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	debts, err := h.service.GetDebtEntries(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve debt entries"))
		return
	}

	response.Success(c, http.StatusOK, debts, "Debt entries retrieved successfully")
}

// CreateDebtCollection handles debt collection creation
func (h *Handler) CreateDebtCollection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	var req CreateDebtCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.CreateDebtCollection(c.Request.Context(), id, organizationID, req.Type, req.ScheduledDate, req.Notes)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to create debt collection"))
		return
	}

	response.Success(c, http.StatusCreated, nil, "Debt collection created successfully")
}

// GetDebtCollections handles debt collections retrieval
func (h *Handler) GetDebtCollections(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	collections, err := h.service.GetDebtCollections(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve debt collections"))
		return
	}

	response.Success(c, http.StatusOK, collections, "Debt collections retrieved successfully")
}

// GetPendingDebtCollections handles pending debt collections retrieval
func (h *Handler) GetPendingDebtCollections(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)

	collections, err := h.service.GetPendingDebtCollections(c.Request.Context(), organizationID)
	if err != nil {
		errors.HandleError(c, errors.WrapError(err, "Failed to retrieve pending debt collections"))
		return
	}

	response.Success(c, http.StatusOK, collections, "Pending debt collections retrieved successfully")
}

// ProcessDebtPayment handles debt payment processing
func (h *Handler) ProcessDebtPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("Invalid customer ID", err))
		return
	}

	var req ProcessDebtPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.ProcessDebtPayment(c.Request.Context(), id, organizationID, req.Amount, req.Method)
	if err != nil {
		if err == ErrCustomerNotFound {
			errors.HandleError(c, errors.NewNotFoundError("Customer", err))
			return
		}
		errors.HandleError(c, errors.WrapError(err, "Failed to process debt payment"))
		return
	}

	response.Success(c, http.StatusOK, nil, "Debt payment processed successfully")
}
