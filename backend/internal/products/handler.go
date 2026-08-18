package products

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles products HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new products handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	category, err := h.service.CreateCategory(c.Request.Context(), organizationID.(uuid.UUID), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

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
		response.BadRequest(c, "invalid category id")
		return
	}

	category, err := h.service.GetCategory(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, category, "Category retrieved successfully")
}

// ListCategories retrieves all categories
// @Summary List Categories
// @Description Get all categories for the organization
// @Tags categories
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]Category}
// @Router /api/v1/categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	categories, err := h.service.ListCategories(c.Request.Context(), organizationID.(uuid.UUID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, categories, "Categories retrieved successfully")
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

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	category, err := h.service.UpdateCategory(c.Request.Context(), id, organizationID.(uuid.UUID), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

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

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	if err := h.service.DeleteCategory(c.Request.Context(), id, organizationID.(uuid.UUID)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

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
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	var req BrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	brand, err := h.service.CreateBrand(c.Request.Context(), organizationID.(uuid.UUID), &req)
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
// @Description Get all brands for the organization
// @Tags brands
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]Brand}
// @Router /api/v1/brands [get]
func (h *Handler) ListBrands(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	brands, err := h.service.ListBrands(c.Request.Context(), organizationID.(uuid.UUID))
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

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	var req BrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	brand, err := h.service.UpdateBrand(c.Request.Context(), id, organizationID.(uuid.UUID), &req)
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

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	if err := h.service.DeleteBrand(c.Request.Context(), id, organizationID.(uuid.UUID)); err != nil {
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
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), organizationID.(uuid.UUID), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
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
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	product, err := h.service.GetProductByBarcode(c.Request.Context(), barcode, organizationID.(uuid.UUID))
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
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	req := &ProductListRequest{
		Page:    1,
		PerPage: 20,
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}

	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}

	if categoryID := c.Query("category_id"); categoryID != "" {
		if id, err := uuid.Parse(categoryID); err == nil {
			req.CategoryID = &id
		}
	}

	if brandID := c.Query("brand_id"); brandID != "" {
		if id, err := uuid.Parse(brandID); err == nil {
			req.BrandID = &id
		}
	}

	req.Search = c.Query("search")
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

	products, total, err := h.service.ListProducts(c.Request.Context(), organizationID.(uuid.UUID), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"products": products,
		"total":    total,
		"page":     req.Page,
		"per_page": req.PerPage,
	}, "Products retrieved successfully")
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

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, organizationID.(uuid.UUID), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, product, "Operation successful")
}

// DeleteProduct deletes a product
// @Summary Delete Product
// @Description Delete a product
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

	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id, organizationID.(uuid.UUID)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "product deleted successfully"}, "Operation successful")
}
