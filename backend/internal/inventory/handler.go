package inventory

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/pkg/logger"
)

type Handler struct {
	service *Service
	db      *sqlx.DB
	cache   *inventoryCache
}

type inventoryCache struct {
	data       interface{}
	expiration time.Time
	mu         sync.RWMutex
}

func newInventoryCache() *inventoryCache {
	return &inventoryCache{}
}

func (c *inventoryCache) get() (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if time.Now().Before(c.expiration) {
		return c.data, true
	}
	return nil, false
}

func (c *inventoryCache) set(data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	c.expiration = time.Now().Add(ttl)
}

func NewHandler(service *Service, db *sqlx.DB) *Handler {
	return &Handler{service, db, newInventoryCache()}
}

// RegisterRoutes registers inventory routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	inventory := router.Group("/inventory")
	{
		inventory.POST("/items", h.CreateInventoryItem)
		inventory.GET("/items-with-supplier", h.ListInventoryItemsWithSupplierInfo)
		inventory.GET("/items", h.ListInventoryItems)
		inventory.GET("/items/:id", h.GetInventoryItem)
		inventory.DELETE("/items/:id", h.DeleteInventoryItem)
		inventory.PATCH("/items/:id/status", h.UpdateItemStatus)
		inventory.POST("/items/:id/receive", h.ReceiveItem)
		inventory.POST("/items/:id/adjust", h.AdjustInventory)
		inventory.POST("/items/:id/transfer", h.TransferItem)
		inventory.GET("/items/:id/history", h.GetItemHistory)
		inventory.GET("/barcode/:code", h.LookupBarcode)
	}

	locations := router.Group("/locations")
	{
		locations.POST("", h.CreateLocation)
		locations.GET("/:id", h.GetLocation)
		locations.GET("", h.ListLocations)
	}

	reservations := router.Group("/reservations")
	{
		reservations.POST("", h.CreateReservation)
		reservations.POST("/:id/release", h.ReleaseReservation)
	}

	tradeIns := inventory.Group("/trade-ins")
	{
		tradeIns.POST("", h.CreateTradeIn)
	}
}

// CreateInventoryItem creates a new inventory item
func (h *Handler) CreateInventoryItem(c *gin.Context) {
	var req InventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)

	item, err := h.service.CreateInventoryItem(c.Request.Context(), &req, userID)
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

	item, err := h.service.GetInventoryItem(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteInventoryItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteInventoryItem(c.Request.Context(), id, getUserID(c)); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "inventory item deleted successfully"})
}

// ListInventoryItems lists inventory items with pagination
func (h *Handler) ListInventoryItems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	status := c.Query("status")
	condition := c.Query("condition")
	locationID := c.Query("location_id")
	productID := c.Query("product_id")
	partTypeID := c.Query("part_type_id")

	// Try cache first for default first page request without filters
	if page == 1 && perPage == 10 && status == "" && condition == "" && locationID == "" && productID == "" && partTypeID == "" {
		if cached, found := h.cache.get(); found {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	filters := make(map[string]interface{})
	if status != "" {
		filters["status"] = status
	}
	if condition != "" {
		filters["condition"] = condition
	}
	if locationID != "" {
		if id, err := uuid.Parse(locationID); err == nil {
			filters["location_id"] = id
		}
	}
	if productID != "" {
		if id, err := uuid.Parse(productID); err == nil {
			filters["product_id"] = id
		}
	}
	if partTypeID != "" {
		if id, err := uuid.Parse(partTypeID); err == nil {
			filters["part_type_id"] = id
		}
	}

	items, total, err := h.service.ListInventoryItems(c.Request.Context(), page, perPage, filters)
	if err != nil {
		handleError(c, err)
		return
	}

	// Ensure items is never null
	if items == nil {
		items = []*InventoryItem{}
	}

	responseData := gin.H{
		"success": true,
		"data": gin.H{
			"items":    items,
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	}

	// Cache the response for default first page without filters
	if page == 1 && perPage == 10 && status == "" && condition == "" && locationID == "" && productID == "" && partTypeID == "" {
		h.cache.set(responseData, 2*time.Minute)
	}

	c.JSON(http.StatusOK, responseData)
}

// ListInventoryItemsWithSupplierInfo lists inventory items with supplier information
func (h *Handler) ListInventoryItemsWithSupplierInfo(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if condition := c.Query("condition"); condition != "" {
		filters["condition"] = condition
	}
	if excludeCondition := c.Query("exclude_condition"); excludeCondition != "" {
		filters["exclude_condition"] = excludeCondition
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
	if partTypeID := c.Query("part_type_id"); partTypeID != "" {
		if id, err := uuid.Parse(partTypeID); err == nil {
			filters["part_type_id"] = id
		}
	}
	// New filters for supplier and purchase
	if supplierID := c.Query("supplier_id"); supplierID != "" {
		if id, err := uuid.Parse(supplierID); err == nil {
			filters["supplier_id"] = id
		}
	}
	if purchaseDateFrom := c.Query("purchase_date_from"); purchaseDateFrom != "" {
		filters["purchase_date_from"] = purchaseDateFrom
	}
	if purchaseDateTo := c.Query("purchase_date_to"); purchaseDateTo != "" {
		filters["purchase_date_to"] = purchaseDateTo
	}
	if minPurchaseCost := c.Query("min_purchase_cost"); minPurchaseCost != "" {
		if cost, err := strconv.ParseFloat(minPurchaseCost, 64); err == nil {
			filters["min_purchase_cost"] = cost
		}
	}
	if maxPurchaseCost := c.Query("max_purchase_cost"); maxPurchaseCost != "" {
		if cost, err := strconv.ParseFloat(maxPurchaseCost, 64); err == nil {
			filters["max_purchase_cost"] = cost
		}
	}

	items, total, err := h.service.ListInventoryItemsWithSupplierInfo(c.Request.Context(), page, perPage, filters)
	if err != nil {
		handleError(c, err)
		return
	}

	// Ensure items is never null
	if items == nil {
		items = []*InventoryItemWithSupplier{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":    items,
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
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
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateItemStatus(c.Request.Context(), id, req.Status); err != nil {
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

	userID := getUserID(c)

	if err := h.service.ReceiveItem(c.Request.Context(), id, req.LocationID, userID); err != nil {
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
	userID := getUserID(c)

	if err := h.service.AdjustInventory(c.Request.Context(), &req, userID); err != nil {
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
	userID := getUserID(c)

	if err := h.service.TransferItem(c.Request.Context(), &req, userID); err != nil {
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

	movements, total, err := h.service.GetItemHistory(c.Request.Context(), id, page, perPage)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"movements": movements,
			"total":     total,
			"page":      page,
			"per_page":  perPage,
		},
	})
}

// LookupBarcode looks up a product or item by barcode
func (h *Handler) LookupBarcode(c *gin.Context) {
	barcode := c.Param("code")

	item, err := h.service.LookupBarcode(c.Request.Context(), barcode)
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

	location, err := h.service.CreateLocation(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, location)
}

// GetLocation retrieves a location by ID
func (h *Handler) GetLocation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	location, err := h.service.GetLocation(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, location)
}

// ListLocations lists all locations
func (h *Handler) ListLocations(c *gin.Context) {
	locations, err := h.service.ListLocations(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": locations})
}

// CreateReservation creates a new reservation
func (h *Handler) CreateReservation(c *gin.Context) {
	var req ReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)

	reservation, err := h.service.ReserveItem(c.Request.Context(), &req, userID)
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

	userID := getUserID(c)

	if err := h.service.ReleaseReservation(c.Request.Context(), id, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reservation released successfully"})
}

// CreateTradeIn creates a new inventory item from a customer trade-in (buying used items)
func (h *Handler) CreateTradeIn(c *gin.Context) {
	var req struct {
		CustomerID     *uuid.UUID `json:"customer_id"`
		CustomerName   string     `json:"customer_name"`
		ProductID      *uuid.UUID `json:"product_id"`
		ProductName    string     `json:"product_name"`
		PartTypeID     *uuid.UUID `json:"part_type_id"`
		PurchaseCost   float64    `json:"purchase_cost" binding:"required"`
		SellingPrice   float64    `json:"selling_price"`
		Notes          *string    `json:"notes"`
		Specifications []struct {
			SpecificationID uuid.UUID `json:"specification_id"`
			ValueText       *string   `json:"value_text"`
			ValueNumber     *float64  `json:"value_number"`
			ValueBoolean    *bool     `json:"value_boolean"`
		} `json:"specifications"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that customer is provided
	if req.CustomerID == nil && req.CustomerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either customer_id or customer_name must be provided"})
		return
	}

	// Validate that either product or part type is provided
	if req.ProductID == nil && req.ProductName == "" && req.PartTypeID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either product_id/product_name or part_type_id must be provided"})
		return
	}

	userID := getUserID(c)

	// If manual names are provided, we need to create or find the customer/product
	// For now, we'll use a simple approach: if manual names, store them directly in the item notes
	// In a production system, you'd want to create the customer/product records first

	// For manual entries, we'll use a placeholder ID and store the name in notes
	var customerID uuid.UUID
	var productID uuid.UUID

	if req.CustomerID != nil {
		customerID = *req.CustomerID
	} else {
		// For manual customer, use a nil UUID and store name in notes
		customerID = uuid.New()
		name := strings.TrimSpace(req.CustomerName)
		code := "TRD-" + strings.ToUpper(strings.ReplaceAll(customerID.String()[:8], "-", ""))
		if _, createErr := h.db.ExecContext(c.Request.Context(), fmt.Sprintf(`INSERT INTO customers (id, code, name, is_active, created_at, updated_at) VALUES ($1, $2, $3, 1, %s, %s)`, dbutil.NowSQL(h.db), dbutil.NowSQL(h.db)), customerID, code, name); createErr != nil {
			handleError(c, fmt.Errorf("failed to create trade-in customer: %w", createErr))
			return
		}
		existingNotes := ""
		if req.Notes != nil {
			existingNotes = *req.Notes
		}
		customerNote := fmt.Sprintf("زبون: %s - %s", req.CustomerName, existingNotes)
		req.Notes = &customerNote
	}

	// Handle product/part type logic
	if req.ProductID != nil {
		productID = *req.ProductID
	} else if req.ProductName != "" {
		// For manual product, use a nil UUID and store name in notes
		productID = uuid.New()
		name := strings.TrimSpace(req.ProductName)
		sku := "TRD-" + strings.ToUpper(strings.ReplaceAll(productID.String()[:8], "-", ""))
		if _, createErr := h.db.ExecContext(c.Request.Context(), fmt.Sprintf(`INSERT INTO products (id, sku, name, cost_price, selling_price, min_stock_level, max_stock_level, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, 0, 0, 1, %s, %s)`, dbutil.NowSQL(h.db), dbutil.NowSQL(h.db)), productID, sku, name, req.PurchaseCost, req.SellingPrice); createErr != nil {
			handleError(c, fmt.Errorf("failed to create trade-in product: %w", createErr))
			return
		}
		existingNotes := ""
		if req.Notes != nil {
			existingNotes = *req.Notes
		}
		productNote := fmt.Sprintf("منتج: %s - %s", req.ProductName, existingNotes)
		req.Notes = &productNote
	} else {
		// No product, using part type only - add note
		productID = uuid.New()
		name := "Trade-in item"
		sku := "TRD-" + strings.ToUpper(strings.ReplaceAll(productID.String()[:8], "-", ""))
		if req.PartTypeID != nil {
			_ = h.db.GetContext(c.Request.Context(), &name, "SELECT COALESCE(name_en, name_ar) FROM part_types WHERE id = $1", *req.PartTypeID)
		}
		if _, createErr := h.db.ExecContext(c.Request.Context(), fmt.Sprintf(`INSERT INTO products (id, sku, name, cost_price, selling_price, min_stock_level, max_stock_level, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, 0, 0, 1, %s, %s)`, dbutil.NowSQL(h.db), dbutil.NowSQL(h.db)), productID, sku, name, req.PurchaseCost, req.SellingPrice); createErr != nil {
			handleError(c, fmt.Errorf("failed to create trade-in product: %w", createErr))
			return
		}
		if req.PartTypeID != nil {
			existingNotes := ""
			if req.Notes != nil {
				existingNotes = *req.Notes
			}
			partNote := fmt.Sprintf("قطعة بدون منتج - %s", existingNotes)
			req.Notes = &partNote
		}
	}

	// Default selling price if not provided (50% markup)
	if req.SellingPrice == 0 {
		req.SellingPrice = req.PurchaseCost * 1.5
	}

	// Create inventory item request
	// For used parts, we might not have a product_id, so handle nil case
	productIDPtr := &productID

	grade := GradeGood
	inventoryReq := &InventoryItemRequest{
		ProductID:    productIDPtr,
		PartTypeID:   req.PartTypeID,
		Condition:    ConditionUsed,
		Grade:        &grade,
		PurchaseCost: req.PurchaseCost,
		SellingPrice: req.SellingPrice,
		Status:       StatusAvailable,
		Notes:        req.Notes,
	}

	item, err := h.service.CreateInventoryItem(c.Request.Context(), inventoryReq, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	// Save specifications if provided
	if len(req.Specifications) > 0 {
		for _, spec := range req.Specifications {
			specQuery := fmt.Sprintf(`
				INSERT INTO item_specification_values (id, inventory_item_id, specification_id, value_text, value_number, value_boolean, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, %s, %s)
			`, dbutil.NowSQL(h.db), dbutil.NowSQL(h.db))
			_, err = h.db.ExecContext(c.Request.Context(), specQuery,
				uuid.New(), item.ID, spec.SpecificationID, spec.ValueText, spec.ValueNumber, spec.ValueBoolean)
			if err != nil {
				handleError(c, fmt.Errorf("failed to save specification value: %w", err))
				return
			}
		}
	}

	// Also save to trade_ins table for tracking
	// Only save if we have a valid customer ID (not manual entry)
	tradeInQuery := fmt.Sprintf(`
		INSERT INTO trade_ins (id, customer_id, inventory_item_id, purchase_price, purchase_date, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, %s, $5, %s, %s)
	`, dbutil.NowSQL(h.db), dbutil.NowSQL(h.db), dbutil.NowSQL(h.db))
	if _, err = h.db.ExecContext(c.Request.Context(), tradeInQuery,
		uuid.New(), customerID, item.ID, req.PurchaseCost, req.Notes); err != nil {
		handleError(c, fmt.Errorf("failed to save trade-in record: %w", err))
		return
	}

	c.JSON(http.StatusCreated, item)
}

// Helper functions

func getUserID(c *gin.Context) uuid.UUID {
	// Extract user ID from context (set by Auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if uuid, ok := userID.(uuid.UUID); ok {
			return uuid
		}
	}
	// Fallback to header for backward compatibility
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
	case ErrCannotDeleteSoldItem:
		status = http.StatusConflict
		message = err.Error()
	}

	logger.Error("Inventory handler error", err)
	c.JSON(status, gin.H{"error": message})
}
