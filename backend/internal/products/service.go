package products

import (
	"context"

	"github.com/google/uuid"
)

// Service handles products business logic
type Service struct {
	repo *Repository
}

// NewService creates a new products service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Category operations

// CreateCategory creates a new category
func (s *Service) CreateCategory(ctx context.Context, organizationID uuid.UUID, req *CategoryRequest) (*Category, error) {
	category := &Category{
		OrganizationID: organizationID,
		Name:           req.Name,
		Description:    req.Description,
		ParentID:       req.ParentID,
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *Service) GetCategory(ctx context.Context, id uuid.UUID) (*Category, error) {
	return s.repo.GetCategoryByID(ctx, id)
}

// ListCategories retrieves all categories
func (s *Service) ListCategories(ctx context.Context, organizationID uuid.UUID) ([]Category, error) {
	return s.repo.ListCategories(ctx, organizationID)
}

// UpdateCategory updates a category
func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *CategoryRequest) (*Category, error) {
	category, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if category.OrganizationID != organizationID {
		return nil, ErrCategoryNotFound
	}

	category.Name = req.Name
	category.Description = req.Description
	category.ParentID = req.ParentID

	if err := s.repo.UpdateCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.DeleteCategory(ctx, id, organizationID)
}

// Brand operations

// CreateBrand creates a new brand
func (s *Service) CreateBrand(ctx context.Context, organizationID uuid.UUID, req *BrandRequest) (*Brand, error) {
	brand := &Brand{
		OrganizationID: organizationID,
		Name:           req.Name,
		Description:    req.Description,
		LogoURL:        req.LogoURL,
	}

	if err := s.repo.CreateBrand(ctx, brand); err != nil {
		return nil, err
	}

	return brand, nil
}

// GetBrand retrieves a brand by ID
func (s *Service) GetBrand(ctx context.Context, id uuid.UUID) (*Brand, error) {
	return s.repo.GetBrandByID(ctx, id)
}

// ListBrands retrieves all brands
func (s *Service) ListBrands(ctx context.Context, organizationID uuid.UUID) ([]Brand, error) {
	return s.repo.ListBrands(ctx, organizationID)
}

// UpdateBrand updates a brand
func (s *Service) UpdateBrand(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *BrandRequest) (*Brand, error) {
	brand, err := s.repo.GetBrandByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if brand.OrganizationID != organizationID {
		return nil, ErrBrandNotFound
	}

	brand.Name = req.Name
	brand.Description = req.Description
	brand.LogoURL = req.LogoURL

	if err := s.repo.UpdateBrand(ctx, brand); err != nil {
		return nil, err
	}

	return brand, nil
}

// DeleteBrand deletes a brand
func (s *Service) DeleteBrand(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.DeleteBrand(ctx, id, organizationID)
}

// Product operations

// CreateProduct creates a new product
func (s *Service) CreateProduct(ctx context.Context, organizationID uuid.UUID, req *ProductRequest) (*Product, error) {
	product := &Product{
		OrganizationID:  organizationID,
		CategoryID:      req.CategoryID,
		BrandID:         req.BrandID,
		Name:            req.Name,
		Description:     req.Description,
		Model:           req.Model,
		SKU:             req.SKU,
		Barcode:         req.Barcode,
		TrackSerial:     req.TrackSerial,
		TrackIndividual: req.TrackIndividual,
		MinStockLevel:   req.MinStockLevel,
		WarrantyDays:    req.WarrantyDays,
	}

	if err := s.repo.CreateProduct(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// GetProduct retrieves a product by ID
func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (*ProductResponse, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := &ProductResponse{
		Product: *product,
	}

	// Load category if exists
	if product.CategoryID != nil {
		category, err := s.repo.GetCategoryByID(ctx, *product.CategoryID)
		if err == nil {
			response.Category = category
		}
	}

	// Load brand if exists
	if product.BrandID != nil {
		brand, err := s.repo.GetBrandByID(ctx, *product.BrandID)
		if err == nil {
			response.Brand = brand
		}
	}

	// Get stock count
	stockCount, err := s.repo.GetProductStockCount(ctx, product.ID)
	if err == nil {
		response.StockCount = stockCount
	}

	return response, nil
}

// GetProductByBarcode retrieves a product by barcode
func (s *Service) GetProductByBarcode(ctx context.Context, barcode string, organizationID uuid.UUID) (*ProductResponse, error) {
	product, err := s.repo.GetProductByBarcode(ctx, barcode, organizationID)
	if err != nil {
		return nil, err
	}

	return s.GetProduct(ctx, product.ID)
}

// ListProducts retrieves products with pagination and filters
func (s *Service) ListProducts(ctx context.Context, organizationID uuid.UUID, req *ProductListRequest) ([]Product, int, error) {
	// Set default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}

	return s.repo.ListProducts(ctx, organizationID, req)
}

// UpdateProduct updates a product
func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *ProductRequest) (*Product, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if product.OrganizationID != organizationID {
		return nil, ErrProductNotFound
	}

	product.CategoryID = req.CategoryID
	product.BrandID = req.BrandID
	product.Name = req.Name
	product.Description = req.Description
	product.Model = req.Model
	product.SKU = req.SKU
	product.Barcode = req.Barcode
	product.TrackSerial = req.TrackSerial
	product.TrackIndividual = req.TrackIndividual
	product.MinStockLevel = req.MinStockLevel
	product.WarrantyDays = req.WarrantyDays

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// DeleteProduct deletes a product
func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.DeleteProduct(ctx, id, organizationID)
}
