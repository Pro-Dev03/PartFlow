package inventory

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type MainHandler struct {
	db *sqlx.DB
}

func NewMainHandler(db *sqlx.DB) *MainHandler {
	return &MainHandler{db: db}
}

// RegisterMainRoutes registers main inventory routes
func (h *MainHandler) RegisterMainRoutes(router *gin.RouterGroup) {
	inventory := router.Group("/inventory")
	{
		inventory.GET("", h.ListInventory)
		inventory.GET("/:id", h.GetInventory)
		inventory.GET("/summary", h.GetInventorySummary)
		inventory.GET("/low-stock", h.GetLowStockItems)
		inventory.GET("/out-of-stock", h.GetOutOfStockItems)
	}
}

// ListInventory lists all inventory records with pagination
func (h *MainHandler) ListInventory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	offset := (page - 1) * perPage

	var inventory []struct {
		ID             uuid.UUID  `json:"id"`
		ProductID      uuid.UUID  `json:"product_id"`
		Quantity       int        `json:"quantity"`
		ReservedQuantity int      `json:"reserved_quantity"`
		Location       string     `json:"location"`
		WarehouseID    *uuid.UUID `json:"warehouse_id"`
		LastRestockedAt *string   `json:"last_restocked_at"`
		CreatedAt      string     `json:"created_at"`
		UpdatedAt      string     `json:"updated_at"`
	}

	query := `
		SELECT i.id, i.product_id, i.quantity, i.reserved_quantity, 
		       i.location, i.warehouse_id, i.last_restocked_at, i.created_at, i.updated_at
		FROM inventory i
		ORDER BY i.updated_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := h.db.Query(query, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item struct {
			ID             uuid.UUID  `json:"id"`
			ProductID      uuid.UUID  `json:"product_id"`
			Quantity       int        `json:"quantity"`
			ReservedQuantity int      `json:"reserved_quantity"`
			Location       string     `json:"location"`
			WarehouseID    *uuid.UUID `json:"warehouse_id"`
			LastRestockedAt *string   `json:"last_restocked_at"`
			CreatedAt      string     `json:"created_at"`
			UpdatedAt      string     `json:"updated_at"`
		}
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.ReservedQuantity, &item.Location, &item.WarehouseID, &item.LastRestockedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		inventory = append(inventory, item)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var total int
	h.db.Get(&total, "SELECT COUNT(*) FROM inventory")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inventory,
		"meta": gin.H{
			"page":      page,
			"per_page":  perPage,
			"total":     total,
		},
	})
}

// GetInventory retrieves an inventory record by ID
func (h *MainHandler) GetInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory ID"})
		return
	}

	var inventory struct {
		ID             uuid.UUID `json:"id"`
		ProductID      uuid.UUID `json:"product_id"`
		ProductName    string    `json:"product_name"`
		Quantity       int       `json:"quantity"`
		ReservedQuantity int     `json:"reserved_quantity"`
		Location       string    `json:"location"`
		WarehouseID    *uuid.UUID `json:"warehouse_id"`
		LastRestockedAt *string  `json:"last_restocked_at"`
		CreatedAt      string    `json:"created_at"`
		UpdatedAt      string    `json:"updated_at"`
	}

	query := `
		SELECT i.id, i.product_id, p.name as product_name, i.quantity, i.reserved_quantity, 
		       i.location, i.warehouse_id, i.last_restocked_at, i.created_at, i.updated_at
		FROM inventory i
		JOIN products p ON i.product_id = p.id
		WHERE i.id = $1
	`

	err = h.db.Get(&inventory, query, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "inventory not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inventory,
	})
}

// GetInventorySummary retrieves inventory summary statistics
func (h *MainHandler) GetInventorySummary(c *gin.Context) {
	var summary struct {
		TotalProducts    int     `json:"total_products"`
		TotalQuantity    int     `json:"total_quantity"`
		TotalValue       float64 `json:"total_value"`
		LowStockItems    int     `json:"low_stock_items"`
		OutOfStockItems  int     `json:"out_of_stock_items"`
		ReservedQuantity int     `json:"reserved_quantity"`
		AvailableQuantity int    `json:"available_quantity"`
	}

	query := `
		SELECT 
			(SELECT COUNT(DISTINCT product_id) FROM inventory) as total_products,
			(SELECT COALESCE(SUM(quantity), 0) FROM inventory) as total_quantity,
			(SELECT COALESCE(SUM(purchase_cost), 0) FROM inventory_items WHERE status = 'AVAILABLE') as total_value,
			(SELECT COUNT(*) FROM inventory WHERE quantity < 5 AND quantity > 0) as low_stock_items,
			(SELECT COUNT(*) FROM inventory WHERE quantity = 0) as out_of_stock_items,
			(SELECT COALESCE(SUM(reserved_quantity), 0) FROM inventory) as reserved_quantity,
			(SELECT COALESCE(SUM(quantity - reserved_quantity), 0) FROM inventory) as available_quantity
	`

	row := h.db.QueryRow(query)
	err := row.Scan(&summary.TotalProducts, &summary.TotalQuantity, &summary.TotalValue, &summary.LowStockItems, &summary.OutOfStockItems, &summary.ReservedQuantity, &summary.AvailableQuantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// GetLowStockItems retrieves items with low stock
func (h *MainHandler) GetLowStockItems(c *gin.Context) {
	var items []struct {
		ID          uuid.UUID `json:"id"`
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Quantity    int       `json:"quantity"`
		MinStock    int       `json:"min_stock_level"`
		Location    string    `json:"location"`
	}

	query := `
		SELECT i.id, i.product_id, p.name as product_name, i.quantity, p.min_stock_level, i.location
		FROM inventory i
		JOIN products p ON i.product_id = p.id
		WHERE i.quantity < p.min_stock_level AND i.quantity > 0
		ORDER BY i.quantity ASC
	`

	err := h.db.Select(&items, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}

// GetOutOfStockItems retrieves items that are out of stock
func (h *MainHandler) GetOutOfStockItems(c *gin.Context) {
	var items []struct {
		ID          uuid.UUID `json:"id"`
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Quantity    int       `json:"quantity"`
		Location    string    `json:"location"`
	}

	query := `
		SELECT i.id, i.product_id, p.name as product_name, i.quantity, i.location
		FROM inventory i
		JOIN products p ON i.product_id = p.id
		WHERE i.quantity = 0
		ORDER BY p.name
	`

	err := h.db.Select(&items, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}