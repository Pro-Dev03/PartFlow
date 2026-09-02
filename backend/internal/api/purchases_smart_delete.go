package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/purchases"
)

// PurchaseSmartDeleteHandler handles smart delete operations for purchases (PRODUCT-PHILOSOPHY.md)
type PurchaseSmartDeleteHandler struct {
	smartDeleteService *purchases.SmartDeleteService
}

// NewPurchaseSmartDeleteHandler creates a new smart delete handler
func NewPurchaseSmartDeleteHandler(db *sqlx.DB) *PurchaseSmartDeleteHandler {
	return &PurchaseSmartDeleteHandler{
		smartDeleteService: purchases.NewSmartDeleteService(db),
	}
}

// SmartDelete handles the smart delete endpoint
// DELETE /api/purchases/:id (now returns SmartDeleteResult instead of simple delete)
func (h *PurchaseSmartDeleteHandler) SmartDelete(c *gin.Context) {
	purchaseIDStr := c.Param("id")
	purchaseID, err := uuid.Parse(purchaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	// Get user ID from context (assuming auth middleware sets this)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Execute smart delete
	result, err := h.smartDeleteService.SmartDelete(c.Request.Context(), purchaseID, userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUsedItemsInfo returns information about used items for a purchase
// GET /api/purchases/:id/used-items
func (h *PurchaseSmartDeleteHandler) GetUsedItemsInfo(c *gin.Context) {
	purchaseIDStr := c.Param("id")
	purchaseID, err := uuid.Parse(purchaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	usedItems, err := h.smartDeleteService.GetUsedItemsInfo(c.Request.Context(), purchaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": usedItems})
}
