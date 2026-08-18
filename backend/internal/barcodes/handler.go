package barcodes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers barcode routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	barcodes := router.Group("/barcodes")
	{
		barcodes.GET("/:code", middleware.RequirePermission("inventory", "read"), h.LookupBarcode)
		barcodes.POST("/generate", middleware.RequirePermission("inventory", "adjust"), h.GenerateBarcode)
		barcodes.POST("/labels", middleware.RequirePermission("inventory", "read"), h.GenerateLabels)
		barcodes.GET("", middleware.RequirePermission("inventory", "read"), h.ListBarcodes)
		barcodes.DELETE("/:id", middleware.RequirePermission("inventory", "adjust"), h.DeleteBarcode)
	}
}

// LookupBarcode looks up a barcode by code
func (h *Handler) LookupBarcode(c *gin.Context) {
	code := c.Param("code")
	organizationID := getOrganizationID(c)

	barcode, err := h.service.LookupBarcode(c.Request.Context(), code, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "barcode not found"})
		return
	}

	c.JSON(http.StatusOK, barcode)
}

// GenerateBarcode generates a new barcode
func (h *Handler) GenerateBarcode(c *gin.Context) {
	var req BarcodeGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := getOrganizationID(c)

	barcode, err := h.service.GenerateBarcode(c.Request.Context(), &req, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, barcode)
}

// ListBarcodes lists all barcodes
func (h *Handler) ListBarcodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	organizationID := getOrganizationID(c)

	barcodes, total, err := h.service.ListBarcodes(c.Request.Context(), organizationID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"barcodes":  barcodes,
		"total":     total,
		"page":      page,
		"per_page":  perPage,
	})
}

// DeleteBarcode deletes a barcode
func (h *Handler) DeleteBarcode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	organizationID := getOrganizationID(c)

	if err := h.service.DeleteBarcode(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "barcode deleted successfully"})
}

// GenerateLabels generates printable labels for barcodes
func (h *Handler) GenerateLabels(c *gin.Context) {
	var req LabelGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := getOrganizationID(c)

	labels, err := h.service.GenerateLabels(c.Request.Context(), &req, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"count":  len(labels),
	})
}

// Helper function to get organization ID from context
func getOrganizationID(c *gin.Context) uuid.UUID {
	return uuid.MustParse(c.GetHeader("X-Organization-ID"))
}