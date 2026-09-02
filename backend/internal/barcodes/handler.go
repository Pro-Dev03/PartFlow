package barcodes

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

// RegisterRoutes registers barcode routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	barcodes := router.Group("/barcodes")
	{
		barcodes.GET("/:code", h.LookupBarcode)
		barcodes.GET("/product/:code", h.LookupProductByBarcode)
		barcodes.GET("/sku/:sku", h.LookupProductBySKU)
		barcodes.POST("/generate", h.GenerateBarcode)
		barcodes.POST("/labels", h.GenerateLabels)
		barcodes.GET("", h.ListBarcodes)
		barcodes.DELETE("/:id", h.DeleteBarcode)
	}
}

// LookupBarcode looks up a barcode by code
func (h *Handler) LookupBarcode(c *gin.Context) {
	code := c.Param("code")

	barcode, err := h.service.LookupBarcode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "barcode not found"})
		return
	}

	c.JSON(http.StatusOK, barcode)
}

// LookupProductByBarcode looks up product information by barcode code
func (h *Handler) LookupProductByBarcode(c *gin.Context) {
	code := c.Param("code")

	product, err := h.service.LookupProductByBarcode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// LookupProductBySKU looks up product information by SKU
func (h *Handler) LookupProductBySKU(c *gin.Context) {
	sku := c.Param("sku")

	product, err := h.service.LookupProductBySKU(c.Request.Context(), sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GenerateBarcode generates a new barcode
func (h *Handler) GenerateBarcode(c *gin.Context) {
	var req BarcodeGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	barcode, err := h.service.GenerateBarcode(c.Request.Context(), &req)
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

	barcodes, total, err := h.service.ListBarcodes(c.Request.Context(), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"barcodes": barcodes,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// DeleteBarcode deletes a barcode
func (h *Handler) DeleteBarcode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteBarcode(c.Request.Context(), id); err != nil {
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

	labels, err := h.service.GenerateLabels(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"count":  len(labels),
	})
}
