package inventory

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

// RegisterRoutes registers inventory routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	inventory := router.Group("/inventory")
	{
		inventory.POST("/items", middleware.RequirePermission("inventory", "adjust"), h.CreateInventoryItem)
		inventory.GET("/items/:id", middleware.RequirePermission("inventory", "read"), h.GetInventoryItem)
		inventory.GET("/items", middleware.RequirePermission("inventory", "read"), h.ListInventoryItems)
		inventory.PATCH("/items/:id/status", middleware.RequirePermission("inventory", "adjust"), h.UpdateItemStatus)
		inventory.POST("/items/:id/receive", middleware.RequirePermission("inventory", "adjust"), h.ReceiveItem)
		inventory.POST("/items/:id/adjust", middleware.RequirePermission("inventory", "adjust"), h.AdjustInventory)
		inventory.POST("/items/:id/transfer", middleware.RequirePermission("inventory", "transfer"), h.TransferItem)
		inventory.GET("/items/:id/history", middleware.RequirePermission("inventory", "read"), h.GetItemHistory)
		inventory.GET("/barcode/:code", middleware.RequirePermission("inventory", "read"), h.LookupBarcode)
	}

	locations := router.Group("/locations")
	{
		locations.POST("", middleware.RequirePermission("inventory", "adjust"), h.CreateLocation)
		locations.GET("/:id", middleware.RequirePermission("inventory", "read"), h.GetLocation)
		locations.GET("", middleware.RequirePermission("inventory", "read"), h.ListLocations)
	}

	reservations := router.Group("/reservations")
	{
		reservations.POST("", middleware.RequirePermission("inventory", "adjust"), h.CreateReservation)
		reservations.POST("/:id/release", middleware.RequirePermission("inventory", "adjust"), h.ReleaseReservation)
	}
}

// CreateInventoryItem creates a new inventory item
func (h *Handler) CreateInventoryItem(c *gin.Context) {
	var req InventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := getOrganizationID(c)
	userID := getUserID(c)

	item, err := h.service.CreateInventoryItem(c.Request.Context(), &req, organizationID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

// GetInventoryItem retrieves an inventory item by ID
func (h *Handler) GetInventoryItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	organizationID := getOrganizationID(c)

	item, err := h.service.GetInventoryItem(c.Request.Context(), id, organizationID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// ListInventoryItems lists inventory items with pagination
func (h *Handler) ListInventoryItems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	organizationID := getOrganizationID(c)

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if condition := c.Query("condition"); condition != "" {
		filters["condition"] = condition
	}
	if locationID := c.Query("location_id"); locationID != "" {
		if id, err := uuid.Parse(locationID); err == nil {
			filters["location_id"] = id
		}
	}
	if productID := c.Query("product_id"); productID != "" {
		if id, err := uuid.Parse(productID); err == nil {
			filters["product_id"] = id
		}
	}

	items, total, err := h.service.ListInventoryItems(c.Request.Context(), organizationID, page, perPage, filters)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"per_page": perPage,
	})
}

// UpdateItemStatus updates the status of an inventory item
func (h *Handler) UpdateItemStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Status Status `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := getOrganizationID(c)

	if err := h.service.UpdateItemStatus(c.Request.Context(), id, organizationID, req.Status); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

// ReceiveItem marks an item as received
func (h *Handler) ReceiveItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		LocationID *uuid.UUID `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := getOrganizationID(c)
	userID := getUserID(c)

	if err := h.service.ReceiveItem(c.Request.Context(), id, organizationID, req.LocationID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item received successfully"})
}

// AdjustInventory adjusts inventory quantity
func (h *Handler) AdjustInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req AdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ItemID = id
	organizationID := getOrganizationID(c)
	userID := getUserID(c)

	if err := h.service.AdjustInventory(c.Request.Context(), &req, organizationID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "inventory adjusted successfully"})
}

// TransferItem transfers an item between locations
func (h *Handler) TransferItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ItemID = id
	organizationID := getOrganizationID(c)
	userID := getUserID(c)

	if err := h.service.TransferItem(c.Request.Context(), &req, organizationID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item transferred successfully"})
}

// GetItemHistory retrieves movement history for an item
func (h *Handler) GetItemHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	organizationID := getOrganizationID(c)

	movements, total, err := h.service.GetItemHistory(c.Request.Context(), id, organizationID, page, perPage)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"movements": movements,
		"total":     total,
		"page":      page,
		"per_page":  perPage,
	})
}

// LookupBarcode looks up a product or item by barcode
func (h *Handler) LookupBarcode(c *gin.Context) {
	barcode := c.Param("code")
	organizationID := getOrganizationID(c)

	item, err := h.service.LookupBarcode(c.Request.Context(), barcode, organizationID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// CreateLocation creates a new location
func (h *Handler) CreateLocation(c *gin.Context) {
	var req LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := getOrganizationID(c)

	location := &Location{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           req.Name,
		Type:           req.Type,
		ParentID:       req.ParentID,
		WarehouseID:    req.WarehouseID,
		Description:    req.Description,
		IsActive:       true,
	}

	// This would need repository and service implementation
	c.JSON(http.StatusCreated, location)
}

// GetLocation retrieves a location by ID
func (h *Handler) GetLocation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	organizationID := getOrganizationID(c)

	// This would need service implementation
	c.JSON(http.StatusOK, gin.H{"id": id, "organization_id": organizationID})
}

// ListLocations lists all locations
func (h *Handler) ListLocations(c *gin.Context) {
	organizationID := getOrganizationID(c)

	// This would need service implementation
	c.JSON(http.StatusOK, gin.H{"organization_id": organizationID})
}

// CreateReservation creates a new reservation
func (h *Handler) CreateReservation(c *gin.Context) {
	var req ReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := getOrganizationID(c)
	userID := getUserID(c)

	reservation, err := h.service.ReserveItem(c.Request.Context(), &req, organizationID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

// ReleaseReservation releases a reservation
func (h *Handler) ReleaseReservation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	organizationID := getOrganizationID(c)

	if err := h.service.ReleaseReservation(c.Request.Context(), id, organizationID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reservation released successfully"})
}

// Helper functions

func getOrganizationID(c *gin.Context) uuid.UUID {
	// This would extract organization ID from JWT token or context
	return uuid.MustParse(c.GetHeader("X-Organization-ID"))
}

func getUserID(c *gin.Context) uuid.UUID {
	// This would extract user ID from JWT token or context
	return uuid.MustParse(c.GetHeader("X-User-ID"))
}

func handleError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"

	switch err {
	case ErrItemNotFound, ErrLocationNotFound:
		status = http.StatusNotFound
		message = err.Error()
	case ErrInvalidStatus, ErrInvalidCondition, ErrInvalidGrade:
		status = http.StatusBadRequest
		message = err.Error()
	case ErrInsufficientStock, ErrItemAlreadyReserved, ErrDuplicateBarcode, ErrDuplicateSerialNumber:
		status = http.StatusConflict
		message = err.Error()
	}

	c.JSON(status, gin.H{"error": message})
}
