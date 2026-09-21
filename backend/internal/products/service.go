package products

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/dashboard"
)

func generateSKUFromName(name string) string {
	base := strings.ToUpper(strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "_", "-").Replace(strings.TrimSpace(name)))
	if base == "" {
		base = "SKU"
	}
	if len(base) > 18 {
		base = base[:18]
	}
	return fmt.Sprintf("%s-%s", base, strings.ToUpper(uuid.NewString()[:6]))
}

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
func (s *Service) CreateCategory(ctx context.Context, req *CategoryRequest) (*Category, error) {
	category := &Category{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Icon:        req.Icon,
		Color:       req.Color,
		IsActive:    req.IsActive != nil && *req.IsActive,
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
func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repo.ListCategories(ctx)
}

// UpdateCategory updates a category
func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, req *CategoryRequest) (*Category, error) {
	category, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	category.Name = req.Name
	category.Description = req.Description
	category.ParentID = req.ParentID
	category.Icon = req.Icon
	category.Color = req.Color
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteCategory(ctx, id)
}

// Brand operations

// CreateBrand creates a new brand
func (s *Service) CreateBrand(ctx context.Context, req *BrandRequest) (*Brand, error) {
	brand := &Brand{
		Name:        req.Name,
		Description: req.Description,
		LogoURL:     req.LogoURL,
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
func (s *Service) ListBrands(ctx context.Context) ([]Brand, error) {
	return s.repo.ListBrands(ctx)
}

// UpdateBrand updates a brand
func (s *Service) UpdateBrand(ctx context.Context, id uuid.UUID, req *BrandRequest) (*Brand, error) {
	brand, err := s.repo.GetBrandByID(ctx, id)
	if err != nil {
		return nil, err
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
func (s *Service) DeleteBrand(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteBrand(ctx, id)
}

// Product operations

// CreateProduct creates a new product
func (s *Service) CreateProduct(ctx context.Context, req *ProductRequest) (*Product, error) {
	if req == nil {
		return nil, fmt.Errorf("product request is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if req.SKU == "" {
		req.SKU = generateSKUFromName(req.Name)
	}

	product := &Product{
		CategoryID:          req.CategoryID,
		BrandID:             req.BrandID,
		PreferredSupplierID: req.PreferredSupplierID,
		Name:                strings.TrimSpace(req.Name),
		Description:         req.Description,
		Model:               req.Model,
		SKU:                 strings.TrimSpace(req.SKU),
		Barcode:             req.Barcode,
		CostPrice:           req.CostPrice,
		SellingPrice:        req.SellingPrice,
		TrackSerial:         req.TrackSerial,
		TrackIndividual:     req.TrackIndividual,
		MinStockLevel:       req.MinStockLevel,
		WarrantyDays:        req.WarrantyDays,
	}

	if err := s.repo.CreateProduct(ctx, product); err != nil {
		return nil, err
	}

	// Invalidate dashboard cache since products data changed
	dashboard.InvalidateDashboardCacheWithReason("product_created")

	return product, nil
}

// CreateProductsBulk creates multiple products in one request while preserving category validation and duplicate checks.
func (s *Service) CreateProductsBulk(ctx context.Context, requests []ProductRequest) ([]*Product, []BulkProductFailure, error) {
	created := make([]*Product, 0, len(requests))
	failed := make([]BulkProductFailure, 0)
	seenSKUs := make(map[string]struct{}, len(requests))

	for index, request := range requests {
		request.Name = strings.TrimSpace(request.Name)
		request.SKU = strings.TrimSpace(request.SKU)
		request.Barcode = strings.TrimSpace(request.Barcode)

		if request.CategoryID != nil {
			if _, err := s.repo.GetCategoryByID(ctx, *request.CategoryID); err != nil {
				failed = append(failed, BulkProductFailure{Index: index, Name: request.Name, SKU: request.SKU, Error: fmt.Sprintf("invalid category: %v", err)})
				continue
			}
		}
		if request.Name == "" {
			failed = append(failed, BulkProductFailure{Index: index, Name: request.Name, SKU: request.SKU, Error: "product name is required"})
			continue
		}
		if request.SKU == "" {
			request.SKU = generateSKUFromName(request.Name)
		}
		if _, exists := seenSKUs[request.SKU]; exists {
			failed = append(failed, BulkProductFailure{Index: index, Name: request.Name, SKU: request.SKU, Error: "duplicate sku in this batch"})
			continue
		}
		product, err := s.CreateProduct(ctx, &request)
		if err != nil {
			msg := err.Error()
			if strings.Contains(strings.ToLower(msg), "unique") || strings.Contains(strings.ToLower(msg), "duplicate") || strings.Contains(strings.ToLower(msg), "constraint") {
				msg = "product with this sku already exists"
			}
			failed = append(failed, BulkProductFailure{Index: index, Name: request.Name, SKU: request.SKU, Error: msg})
			continue
		}
		created = append(created, product)
		seenSKUs[request.SKU] = struct{}{}
	}

	if len(created) > 0 {
		dashboard.InvalidateDashboardCacheWithReason("product_created")
	}
	return created, failed, nil
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
func (s *Service) GetProductByBarcode(ctx context.Context, barcode string) (*ProductResponse, error) {
	product, err := s.repo.GetProductByBarcode(ctx, barcode)
	if err != nil {
		return nil, err
	}

	return s.GetProduct(ctx, product.ID)
}

// ListProducts retrieves products with pagination and filters
func (s *Service) ListProducts(ctx context.Context, req *ProductListRequest) ([]Product, int, error) {
	// Set default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}

	products, total, err := s.repo.ListProducts(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	for index := range products {
		stockCount, stockErr := s.repo.GetProductStockCount(ctx, products[index].ID)
		if stockErr != nil {
			return nil, 0, stockErr
		}
		products[index].CurrentQuantity = stockCount
	}

	return products, total, nil
}

// UpdateProduct updates a product
func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, req *ProductRequest) (*Product, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	product.CategoryID = req.CategoryID
	product.BrandID = req.BrandID
	product.PreferredSupplierID = req.PreferredSupplierID
	product.Name = req.Name
	product.Description = req.Description
	product.Model = req.Model
	product.SKU = req.SKU
	product.Barcode = req.Barcode
	product.CostPrice = req.CostPrice
	product.SellingPrice = req.SellingPrice
	product.TrackSerial = req.TrackSerial
	product.TrackIndividual = req.TrackIndividual
	product.MinStockLevel = req.MinStockLevel
	product.WarrantyDays = req.WarrantyDays

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return nil, err
	}

	// Invalidate dashboard cache since products data changed
	dashboard.InvalidateDashboardCacheWithReason("product_updated")

	return product, nil
}

func (s *Service) UpdateProductName(ctx context.Context, id uuid.UUID, name string) (*Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if err := s.repo.UpdateProductName(ctx, id, name); err != nil {
		return nil, err
	}
	dashboard.InvalidateDashboardCacheWithReason("product_name_updated")
	return s.repo.GetProductByID(ctx, id)
}

func (s *Service) UpdateMinimumStock(ctx context.Context, id uuid.UUID, minStockLevel int) error {
	if err := s.repo.UpdateMinimumStock(ctx, id, minStockLevel); err != nil {
		return err
	}
	dashboard.InvalidateDashboardCacheWithReason("product_min_stock_updated")
	return nil
}

// DeleteProduct permanently deletes a product and cleans all transaction
// history linked to that product so dashboard and report totals stay aligned.
func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		return err
	}
	dashboard.InvalidateDashboardCacheWithReason("product_deleted")
	return nil
}

// RestoreProduct restores a soft-deleted product
func (s *Service) RestoreProduct(ctx context.Context, id uuid.UUID) error {
	// Invalidate dashboard cache since products data changed
	dashboard.InvalidateDashboardCacheWithReason("product_restored")
	return s.repo.RestoreProduct(ctx, id)
}

// ArchiveProduct archives a product (soft delete)
func (s *Service) ArchiveProduct(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.ArchiveProduct(ctx, id)
}

// GenerateBarcode generates a new barcode for a product
func (s *Service) GenerateBarcode(ctx context.Context, productID uuid.UUID) (string, error) {
	product, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return "", err
	}

	// Generate internal barcode if not exists
	if product.Barcode == "" {
		barcode := generateInternalBarcode(product.ID, product.SKU)
		product.Barcode = barcode

		if err := s.repo.UpdateProduct(ctx, product); err != nil {
			return "", err
		}

		return barcode, nil
	}

	return product.Barcode, nil
}

// GetProductStock retrieves detailed stock information for a product
func (s *Service) GetProductStock(ctx context.Context, productID uuid.UUID) (*ProductStockInfo, error) {
	trackIndividual, minStockLevel, err := s.repo.GetProductStockSettings(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	stockCount, err := s.repo.GetProductStockCount(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Get available items count
	availableCount, err := s.repo.GetAvailableItemCount(ctx, productID)
	if err != nil {
		availableCount = 0
	}

	// Get reserved items count
	reservedCount, err := s.repo.GetReservedItemCount(ctx, productID)
	if err != nil {
		reservedCount = 0
	}

	return &ProductStockInfo{
		ProductID:       productID,
		TotalStock:      stockCount,
		Available:       availableCount,
		Reserved:        reservedCount,
		TrackIndividual: trackIndividual,
		MinStockLevel:   minStockLevel,
		IsLowStock:      stockCount < minStockLevel,
	}, nil
}

// SearchProducts searches products by name, SKU, or barcode
func (s *Service) SearchProducts(ctx context.Context, query string, limit int) ([]Product, error) {
	return s.repo.SearchProducts(ctx, query, limit)
}

// Helper functions

func generateInternalBarcode(productID uuid.UUID, sku string) string {
	if sku != "" {
		return "PF-" + sku
	}
	return "PF-" + productID.String()[:8]
}
