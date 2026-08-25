package ledgers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetCustomerLedger retrieves customer ledger entries
func (h *Handler) GetCustomerLedger(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	entries, total, err := h.service.GetLedgerEntries(c.Request.Context(), LedgerTypeCustomer, customerID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   entries,
		"total":  total,
		"page":   page,
		"per_page": perPage,
	})
}

// GetCustomerLedgerSummary retrieves customer ledger summary
func (h *Handler) GetCustomerLedgerSummary(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	summary, err := h.service.GetCustomerLedgerSummary(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetSupplierLedger retrieves supplier ledger entries
func (h *Handler) GetSupplierLedger(c *gin.Context) {
	supplierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	entries, total, err := h.service.GetLedgerEntries(c.Request.Context(), LedgerTypeSupplier, supplierID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   entries,
		"total":  total,
		"page":   page,
		"per_page": perPage,
	})
}

// GetSupplierLedgerSummary retrieves supplier ledger summary
func (h *Handler) GetSupplierLedgerSummary(c *gin.Context) {
	supplierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier ID"})
		return
	}

	summary, err := h.service.GetSupplierLedgerSummary(c.Request.Context(), supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetInventoryLedger retrieves inventory ledger entries
func (h *Handler) GetInventoryLedger(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	entries, total, err := h.service.GetLedgerEntries(c.Request.Context(), LedgerTypeInventory, productID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   entries,
		"total":  total,
		"page":   page,
		"per_page": perPage,
	})
}

// GetInventoryLedgerSummary retrieves inventory ledger summary
func (h *Handler) GetInventoryLedgerSummary(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	summary, err := h.service.GetInventoryLedgerSummary(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// CreateLedgerEntry creates a new ledger entry (manual entry)
func (h *Handler) CreateLedgerEntry(c *gin.Context) {
	var req LedgerEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (assuming auth middleware sets this)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	entry, err := h.service.CreateLedgerEntry(c.Request.Context(), &req, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}
