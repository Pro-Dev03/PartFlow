package products

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/errors"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles products HTTP requests
type Handler struct {
	service       *Service
	cache         *productsCache
	categoryCache *productsCache
}

type productsCache struct {
	data       interface{}
	expiration time.Time
	mu         sync.RWMutex
}

func newProductsCache() *productsCache {
	return &productsCache{}
}

func (c *productsCache) get() (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if time.Now().Before(c.expiration) {
		return c.data, true
	}
	return nil, false
}

func (c *productsCache) set(data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	c.expiration = time.Now().Add(ttl)
}

func (c *productsCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = nil
	c.expiration = time.Time{}
}

// NewHandler creates a new products handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:       service,
		cache:         newProductsCache(),
		categoryCache: newProductsCache(),
	}
}

// Category handlers

// CreateCategory creates a new category
// @Summary Create Category
// @Description Create a new product category
// @Tags categories
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CategoryRequest true "Category data"
// @Success 201 {object} response.Response{data=Category}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/categories [post]
func (h *Handler) CreateCategory(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	category, err := h.service.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		errors.HandleError(c, errors.WrapError(err, "Failed to create category"))
		return
	}

	h.categoryCache.clear()
	response.Created(c, category, "Category created successfully")
}

// GetCategory retrieves a category by ID
// @Summary Get Category
// @Description Get a category by ID
// @Tags categories
// @Produce json
// @Security Bearer
// @Param id path string true "Category ID"
// @Success 200 {object} response.Response{data=Category}
// @Failure 404 {object} response.Response
// @Router /api/v1/categories/{id} [get]
func (h *Handler) GetCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.HandleError(c, errors.NewValidationError("invalid category id", err))
		return
	}

	category, err := h.service.GetCategory(c.Request.Context(), id)
	if err != nil {
		errors.HandleError(c, errors.WrapErrorWithType(err, errors.ErrorTypeNotFound, "Category not found", http.StatusNotFound))
		return
	}

	response.OK(c, category, "Category retrieved successfully")
}

// ListCategories retrieves all categories
// @Summary List Categories
// @Description Get all categories
// @Tags categories
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]Category}
// @Router /api/v1/categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	// Try cache first
	if cached, found := h.categoryCache.get(); found {
		c.JSON(http.StatusOK, cached)
		return
	}

	categories, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	responseData := gin.H{
		"success": true,
		"data":    categories,
		"message": "Categories retrieved successfully",
	}

	// Cache the response
	h.categoryCache.set(responseData, 5*time.Minute)

	c.JSON(http.StatusOK, responseData)
}

// UpdateCategory updates a category
// @Summary Update Category
// @Description Update a category
// @Tags categories
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Category ID"
// @Param request body CategoryRequest true "Category data"
// @Success 200 {object} response.Response{data=Category}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/categories/{id} [put]
func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid category id")
		return
	}

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	category, err := h.service.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	h.categoryCache.clear()
	response.OK(c, category, "Operation successful")
}

// DeleteCategory deletes a category
// @Summary Delete Category
// @Description Delete a category
// @Tags categories
// @Security Bearer
// @Param id path string true "Category ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/categories/{id} [delete]
func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid category id")
		return
	}

	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	h.categoryCache.clear()
	response.OK(c, gin.H{"message": "category deleted successfully"}, "Operation successful")
}

// Brand handlers

// CreateBrand creates a new brand
// @Summary Create Brand
// @Description Create a new product brand
// @Tags brands
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body BrandRequest true "Brand data"
// @Success 201 {object} response.Response{data=Brand}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/brands [post]
func (h *Handler) CreateBrand(c *gin.Context) {
	var req BrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	brand, err := h.service.CreateBrand(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, brand, "Operation successful")
}

// GetBrand retrieves a brand by ID
// @Summary Get Brand
// @Description Get a brand by ID
// @Tags brands
// @Produce json
// @Security Bearer
// @Param id path string true "Brand ID"
// @Success 200 {object} response.Response{data=Brand}
// @Failure 404 {object} response.Response
// @Router /api/v1/brands/{id} [get]
func (h *Handler) GetBrand(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid brand id")
		return
	}

	brand, err := h.service.GetBrand(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, brand, "Operation successful")
}

// ListBrands retrieves all brands
// @Summary List Brands
// @Description Get all brands
// @Tags brands
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]Brand}
// @Router /api/v1/brands [get]
func (h *Handler) ListBrands(c *gin.Context) {

	brands, err := h.service.ListBrands(c.Request.Context())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, brands, "Operation successful")
}

// UpdateBrand updates a brand
// @Summary Update Brand
// @Description Update a brand
// @Tags brands
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Brand ID"
// @Param request body BrandRequest true "Brand data"
// @Success 200 {object} response.Response{data=Brand}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/brands/{id} [put]
func (h *Handler) UpdateBrand(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid brand id")
		return
	}

	var req BrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	brand, err := h.service.UpdateBrand(c.Request.Context(), id, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, brand, "Operation successful")
}

// DeleteBrand deletes a brand
// @Summary Delete Brand
// @Description Delete a brand
// @Tags brands
// @Security Bearer
// @Param id path string true "Brand ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/brands/{id} [delete]
func (h *Handler) DeleteBrand(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid brand id")
		return
	}

	if err := h.service.DeleteBrand(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "brand deleted successfully"}, "Operation successful")
}

// Product handlers

// CreateProduct creates a new product
// @Summary Create Product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ProductRequest true "Product data"
// @Success 201 {object} response.Response{data=Product}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.ValidateRequest(err))
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		errors.HandleError(c, errors.WrapError(err, "Failed to create product"))
		return
	}

	response.OK(c, product, "Operation successful")
}

// GetProduct retrieves a product by ID
// @Summary Get Product
// @Description Get a product by ID
// @Tags products
// @Produce json
// @Security Bearer
// @Param id path string true "Product ID"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 404 {object} response.Response
// @Router /api/v1/products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, product, "Operation successful")
}

// GetProductByBarcode retrieves a product by barcode
// @Summary Get Product by Barcode
// @Description Get a product by barcode
// @Tags products
// @Produce json
// @Security Bearer
// @Param barcode path string true "Barcode"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 404 {object} response.Response
// @Router /api/v1/products/barcode/{barcode} [get]
func (h *Handler) GetProductByBarcode(c *gin.Context) {
	barcode := c.Param("barcode")

	product, err := h.service.GetProductByBarcode(c.Request.Context(), barcode)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, product, "Operation successful")
}

// ListProducts retrieves products with pagination and filters
// @Summary List Products
// @Description Get products with pagination and filters
// @Tags products
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param category_id query string false "Filter by category ID"
// @Param brand_id query string false "Filter by brand ID"
// @Param search query string false "Search by name, model, or SKU"
// @Param track_serial query bool false "Filter by track_serial"
// @Param track_individual query bool false "Filter by track_individual"
// @Param sort_by query string false "Sort field" default(name)
// @Param sort_order query string false "Sort order" default(ASC)
// @Success 200 {object} response.Response{data=[]Product}
// @Router /api/v1/products [get]
func (h *Handler) ListProducts(c *gin.Context) {
	// Try cache first for default first page request
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	categoryID := c.Query("category_id")
	brandID := c.Query("brand_id")

	// Only cache default first page without filters
	if page == 1 && perPage == 20 && search == "" && categoryID == "" && brandID == "" {
		if cached, found := h.cache.get(); found {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	req := &ProductListRequest{
		Page:    page,
		PerPage: perPage,
	}

	if categoryID != "" {
		if id, err := uuid.Parse(categoryID); err == nil {
			req.CategoryID = &id
		}
	}

	if brandID != "" {
		if id, err := uuid.Parse(brandID); err == nil {
			req.BrandID = &id
		}
	}

	req.Search = search
	req.SortBy = c.DefaultQuery("sort_by", "name")
	req.SortOrder = c.DefaultQuery("sort_order", "ASC")

	if trackSerial := c.Query("track_serial"); trackSerial != "" {
		if val, err := strconv.ParseBool(trackSerial); err == nil {
			req.TrackSerial = &val
		}
	}

	if trackIndividual := c.Query("track_individual"); trackIndividual != "" {
		if val, err := strconv.ParseBool(trackIndividual); err == nil {
			req.TrackIndividual = &val
		}
	}

	products, total, err := h.service.ListProducts(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Ensure products is never null
	if products == nil {
		products = []Product{}
	}

	responseData := gin.H{
		"success": true,
		"data": gin.H{
			"products": products,
			"total":    total,
			"page":     req.Page,
			"per_page": req.PerPage,
		},
		"message": "Products retrieved successfully",
	}

	// Cache the response for default first page without filters
	if page == 1 && perPage == 20 && search == "" && categoryID == "" && brandID == "" {
		h.cache.set(responseData, 3*time.Minute)
	}

	c.JSON(http.StatusOK, responseData)
}

// UpdateProduct updates a product
// @Summary Update Product
// @Description Update a product
// @Tags products
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Product ID"
// @Param request body ProductRequest true "Product data"
// @Success 200 {object} response.Response{data=Product}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, product, "Operation successful")
}

func (h *Handler) UpdateMinimumStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}
	var req struct {
		MinStockLevel int `json:"min_stock_level" binding:"min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.UpdateMinimumStock(c.Request.Context(), id, req.MinStockLevel); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"min_stock_level": req.MinStockLevel}, "Minimum stock updated successfully")
}

// DeleteProduct deletes a product (soft delete)
// @Summary Delete Product
// @Description Delete a product (soft delete - marks as deleted but keeps record)
// @Tags products
// @Security Bearer
// @Param id path string true "Product ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "product deleted successfully (soft delete)"}, "Operation successful")
}

// RestoreProduct restores a soft-deleted product
// @Summary Restore Product
// @Description Restore a soft-deleted product
// @Tags products
// @Security Bearer
// @Param id path string true "Product ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/products/{id}/restore [post]
func (h *Handler) RestoreProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	if err := h.service.RestoreProduct(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "product restored successfully"}, "Operation successful")
}

// ArchiveProduct archives a product (soft delete)
// @Summary Archive Product
// @Description Archive a product (soft delete)
// @Tags products
// @Security Bearer
// @Param id path string true "Product ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/products/{id}/archive [post]
func (h *Handler) ArchiveProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	if err := h.service.ArchiveProduct(c.Request.Context(), id); err != nil {
		if err == ErrProductNotFound {
			response.NotFound(c, "product not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "product archived successfully"}, "Operation successful")
}

// GenerateBarcode generates a new barcode for a product
// @Summary Generate Barcode
// @Description Generate a new barcode for a product
// @Tags products
// @Produce json
// @Security Bearer
// @Param id path string true "Product ID"
// @Success 200 {object} response.Response{data=gin.H}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/products/{id}/barcode [post]
func (h *Handler) GenerateBarcode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	barcode, err := h.service.GenerateBarcode(c.Request.Context(), id)
	if err != nil {
		if err == ErrProductNotFound {
			response.NotFound(c, "product not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"barcode": barcode}, "Barcode generated successfully")
}

// GetProductStock retrieves detailed stock information for a product
// @Summary Get Product Stock
// @Description Get detailed stock information for a product
// @Tags products
// @Produce json
// @Security Bearer
// @Param id path string true "Product ID"
// @Success 200 {object} response.Response{data=ProductStockInfo}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/products/{id}/stock [get]
func (h *Handler) GetProductStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	stockInfo, err := h.service.GetProductStock(c.Request.Context(), id)
	if err != nil {
		if err == ErrProductNotFound {
			response.NotFound(c, "product not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, stockInfo, "Stock information retrieved successfully")
}

// SearchProducts searches products by name, SKU, or barcode
// @Summary Search Products
// @Description Search products by name, SKU, or barcode
// @Tags products
// @Produce json
// @Security Bearer
// @Param q query string true "Search query"
// @Param limit query int false "Result limit" default(20)
// @Success 200 {object} response.Response{data=gin.H}
// @Failure 400 {object} response.Response
// @Router /api/v1/products/search [get]
func (h *Handler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.BadRequest(c, "query parameter 'q' is required")
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	products, err := h.service.SearchProducts(c.Request.Context(), query, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"products": products,
		"count":    len(products),
	}, "Search completed successfully")
}
