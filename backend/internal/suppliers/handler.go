package suppliers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles HTTP requests for suppliers
type Handler struct {
	service *Service
}

// NewHandler creates a new supplier handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateSupplier handles supplier creation
func (h *Handler) CreateSupplier(c *gin.Context) {
	var req SupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	supplier, err := h.service.CreateSupplier(c.Request.Context(), organizationID, &req)
	if err != nil {
		if err == ErrSupplierCodeExists {
			response.Error(c, http.StatusConflict, http.StatusConflict, "Supplier code already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create supplier", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, supplier, "Supplier created successfully")
}

// GetSupplier handles supplier retrieval
func (h *Handler) GetSupplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	supplier, err := h.service.GetSupplier(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve supplier", err.Error())
		return
	}

	response.Success(c, http.StatusOK, supplier, "Supplier retrieved successfully")
}

// ListSuppliers handles supplier listing
func (h *Handler) ListSuppliers(c *gin.Context) {
	var req SupplierListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	suppliers, total, err := h.service.ListSuppliers(c.Request.Context(), organizationID, req.Page, req.PerPage, req.Search, req.IsActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve suppliers", err.Error())
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, suppliers, total, req.Page, req.PerPage, "Suppliers retrieved successfully")
}

// UpdateSupplier handles supplier update
func (h *Handler) UpdateSupplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	var req SupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	supplier, err := h.service.UpdateSupplier(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		if err == ErrSupplierCodeExists {
			response.Error(c, http.StatusConflict, http.StatusConflict, "Supplier code already exists", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to update supplier", err.Error())
		return
	}

	response.Success(c, http.StatusOK, supplier, "Supplier updated successfully")
}

// DeleteSupplier handles supplier deletion
func (h *Handler) DeleteSupplier(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.DeleteSupplier(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to delete supplier", err.Error())
		return
	}

	response.Success(c, http.StatusOK, nil, "Supplier deleted successfully")
}

// GetSupplierLedger handles supplier ledger retrieval
func (h *Handler) GetSupplierLedger(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	ledger, err := h.service.GetSupplierLedger(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve supplier ledger", err.Error())
		return
	}

	response.Success(c, http.StatusOK, ledger, "Supplier ledger retrieved successfully")
}

// AddPayment handles payment addition
func (h *Handler) AddPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
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
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		if err == ErrPaymentAmountInvalid {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid payment amount", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to add payment", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, payment, "Payment added successfully")
}

// GetSupplierDebtSummary handles supplier debt summary retrieval
func (h *Handler) GetSupplierDebtSummary(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	summary, err := h.service.GetSupplierDebtSummary(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve debt summary", err.Error())
		return
	}

	response.Success(c, http.StatusOK, summary, "Debt summary retrieved successfully")
}

// UpdateCreditLimit handles credit limit update
func (h *Handler) UpdateCreditLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	var req UpdateCreditLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.UpdateCreditLimit(c.Request.Context(), id, organizationID, req.NewLimit)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		if err == ErrCreditLimitBelowBalance {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Credit limit cannot be set below current balance", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to update credit limit", err.Error())
		return
	}

	response.Success(c, http.StatusOK, nil, "Credit limit updated successfully")
}

// GetOverdueSuppliers handles overdue suppliers retrieval
func (h *Handler) GetOverdueSuppliers(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)

	overdueSuppliers, err := h.service.GetOverdueSuppliers(c.Request.Context(), organizationID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve overdue suppliers", err.Error())
		return
	}

	response.Success(c, http.StatusOK, overdueSuppliers, "Overdue suppliers retrieved successfully")
}

// CreateDebtEntry handles debt entry creation
func (h *Handler) CreateDebtEntry(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	var req CreateDebtEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.CreateDebtEntry(c.Request.Context(), id, organizationID, req.Amount, req.ReferenceID, req.ReferenceType, req.DueDate)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		if err == ErrCreditLimitExceeded {
			response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Credit limit exceeded", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create debt entry", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, nil, "Debt entry created successfully")
}

// GetDebtEntries handles debt entries retrieval
func (h *Handler) GetDebtEntries(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	debts, err := h.service.GetDebtEntries(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve debt entries", err.Error())
		return
	}

	response.Success(c, http.StatusOK, debts, "Debt entries retrieved successfully")
}

// CreateDebtCollection handles debt collection creation
func (h *Handler) CreateDebtCollection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	var req CreateDebtCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.CreateDebtCollection(c.Request.Context(), id, organizationID, req.Type, req.ScheduledDate, req.Notes)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to create debt collection", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, nil, "Debt collection created successfully")
}

// GetDebtCollections handles debt collections retrieval
func (h *Handler) GetDebtCollections(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	collections, err := h.service.GetDebtCollections(c.Request.Context(), id, organizationID)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve debt collections", err.Error())
		return
	}

	response.Success(c, http.StatusOK, collections, "Debt collections retrieved successfully")
}

// GetPendingDebtCollections handles pending debt collections retrieval
func (h *Handler) GetPendingDebtCollections(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)

	collections, err := h.service.GetPendingDebtCollections(c.Request.Context(), organizationID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to retrieve pending debt collections", err.Error())
		return
	}

	response.Success(c, http.StatusOK, collections, "Pending debt collections retrieved successfully")
}

// ProcessDebtPayment handles debt payment processing
func (h *Handler) ProcessDebtPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid supplier ID", err.Error())
		return
	}

	var req ProcessDebtPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	err = h.service.ProcessDebtPayment(c.Request.Context(), id, organizationID, req.Amount, req.Method)
	if err != nil {
		if err == ErrSupplierNotFound {
			response.Error(c, http.StatusNotFound, http.StatusNotFound, "Supplier not found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Failed to process debt payment", err.Error())
		return
	}

	response.Success(c, http.StatusOK, nil, "Debt payment processed successfully")
}