package products

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles products data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new products repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Category CRUD operations

// CreateCategory creates a new category
func (r *Repository) CreateCategory(ctx context.Context, category *Category) error {
	query := `
		INSERT INTO categories (id, organization_id, name, description, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	category.ID = uuid.New()
	category.CreatedAt = now
	category.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		category.ID,
		category.OrganizationID,
		category.Name,
		category.Description,
		category.ParentID,
		category.CreatedAt,
		category.UpdatedAt,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	return err
}

// GetCategoryByID retrieves a category by ID
func (r *Repository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	query := `
		SELECT id, organization_id, name, description, parent_id, created_at, updated_at
		FROM categories
		WHERE id = $1
	`
	var category Category
	err := r.db.GetContext(ctx, &category, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrCategoryNotFound
	}
	return &category, err
}

// ListCategories retrieves all categories for an organization
func (r *Repository) ListCategories(ctx context.Context, organizationID uuid.UUID) ([]Category, error) {
	query := `
		SELECT id, organization_id, name, description, parent_id, created_at, updated_at
		FROM categories
		WHERE organization_id = $1
		ORDER BY name
	`
	var categories []Category
	err := r.db.SelectContext(ctx, &categories, query, organizationID)
	return categories, err
}

// UpdateCategory updates a category
func (r *Repository) UpdateCategory(ctx context.Context, category *Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, parent_id = $3, updated_at = $4
		WHERE id = $5 AND organization_id = $6
	`
	category.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		category.Name,
		category.Description,
		category.ParentID,
		category.UpdatedAt,
		category.ID,
		category.OrganizationID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// DeleteCategory deletes a category
func (r *Repository) DeleteCategory(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	// Check if category has products
	var count int
	checkQuery := `SELECT COUNT(*) FROM products WHERE category_id = $1`
	err := r.db.GetContext(ctx, &count, checkQuery, id)
	if err != nil {
		return err
	}

	if count > 0 {
		return ErrCategoryHasProducts
	}

	query := `DELETE FROM categories WHERE id = $1 AND organization_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// Brand CRUD operations

// CreateBrand creates a new brand
func (r *Repository) CreateBrand(ctx context.Context, brand *Brand) error {
	query := `
		INSERT INTO brands (id, organization_id, name, description, logo_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	brand.ID = uuid.New()
	brand.CreatedAt = now
	brand.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		brand.ID,
		brand.OrganizationID,
		brand.Name,
		brand.Description,
		brand.LogoURL,
		brand.CreatedAt,
		brand.UpdatedAt,
	).Scan(&brand.ID, &brand.CreatedAt, &brand.UpdatedAt)

	return err
}

// GetBrandByID retrieves a brand by ID
func (r *Repository) GetBrandByID(ctx context.Context, id uuid.UUID) (*Brand, error) {
	query := `
		SELECT id, organization_id, name, description, logo_url, created_at, updated_at
		FROM brands
		WHERE id = $1
	`
	var brand Brand
	err := r.db.GetContext(ctx, &brand, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrBrandNotFound
	}
	return &brand, err
}

// ListBrands retrieves all brands for an organization
func (r *Repository) ListBrands(ctx context.Context, organizationID uuid.UUID) ([]Brand, error) {
	query := `
		SELECT id, organization_id, name, description, logo_url, created_at, updated_at
		FROM brands
		WHERE organization_id = $1
		ORDER BY name
	`
	var brands []Brand
	err := r.db.SelectContext(ctx, &brands, query, organizationID)
	return brands, err
}

// UpdateBrand updates a brand
func (r *Repository) UpdateBrand(ctx context.Context, brand *Brand) error {
	query := `
		UPDATE brands
		SET name = $1, description = $2, logo_url = $3, updated_at = $4
		WHERE id = $5 AND organization_id = $6
	`
	brand.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		brand.Name,
		brand.Description,
		brand.LogoURL,
		brand.UpdatedAt,
		brand.ID,
		brand.OrganizationID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrBrandNotFound
	}

	return nil
}

// DeleteBrand deletes a brand
func (r *Repository) DeleteBrand(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	// Check if brand has products
	var count int
	checkQuery := `SELECT COUNT(*) FROM products WHERE brand_id = $1`
	err := r.db.GetContext(ctx, &count, checkQuery, id)
	if err != nil {
		return err
	}

	if count > 0 {
		return ErrBrandHasProducts
	}

	query := `DELETE FROM brands WHERE id = $1 AND organization_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrBrandNotFound
	}

	return nil
}

// Product CRUD operations

// CreateProduct creates a new product
func (r *Repository) CreateProduct(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (id, organization_id, category_id, brand_id, name, description, model, sku, barcode, track_serial, track_individual, min_stock_level, warranty_days, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	product.ID = uuid.New()
	product.CreatedAt = now
	product.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		product.ID,
		product.OrganizationID,
		product.CategoryID,
		product.BrandID,
		product.Name,
		product.Description,
		product.Model,
		product.SKU,
		product.Barcode,
		product.TrackSerial,
		product.TrackIndividual,
		product.MinStockLevel,
		product.WarrantyDays,
		product.CreatedAt,
		product.UpdatedAt,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	return err
}

// GetProductByID retrieves a product by ID
func (r *Repository) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	query := `
		SELECT id, organization_id, category_id, brand_id, name, description, model, sku, barcode, track_serial, track_individual, min_stock_level, warranty_days, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	var product Product
	err := r.db.GetContext(ctx, &product, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	return &product, err
}

// GetProductByBarcode retrieves a product by barcode
func (r *Repository) GetProductByBarcode(ctx context.Context, barcode string, organizationID uuid.UUID) (*Product, error) {
	query := `
		SELECT id, organization_id, category_id, brand_id, name, description, model, sku, barcode, track_serial, track_individual, min_stock_level, warranty_days, created_at, updated_at
		FROM products
		WHERE barcode = $1 AND organization_id = $2
	`
	var product Product
	err := r.db.GetContext(ctx, &product, query, barcode, organizationID)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	return &product, err
}

// ListProducts retrieves products with pagination and filters
func (r *Repository) ListProducts(ctx context.Context, organizationID uuid.UUID, req *ProductListRequest) ([]Product, int, error) {
	// Build base query
	baseQuery := `
		FROM products
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) ` + baseQuery

	args := []interface{}{organizationID}
	argCount := 1

	// Add filters
	if req.CategoryID != nil {
		argCount++
		baseQuery += ` AND category_id = $` + string(rune(argCount+'0'))
		countQuery += ` AND category_id = $` + string(rune(argCount+'0'))
		args = append(args, req.CategoryID)
	}

	if req.BrandID != nil {
		argCount++
		baseQuery += ` AND brand_id = $` + string(rune(argCount+'0'))
		countQuery += ` AND brand_id = $` + string(rune(argCount+'0'))
		args = append(args, req.BrandID)
	}

	if req.Search != "" {
		argCount++
		baseQuery += ` AND (name ILIKE $` + string(rune(argCount+'0')) + ` OR model ILIKE $` + string(rune(argCount+'0')) + ` OR sku ILIKE $` + string(rune(argCount+'0')) + `)`
		countQuery += ` AND (name ILIKE $` + string(rune(argCount+'0')) + ` OR model ILIKE $` + string(rune(argCount+'0')) + ` OR sku ILIKE $` + string(rune(argCount+'0')) + `)`
		args = append(args, "%"+req.Search+"%")
	}

	if req.TrackSerial != nil {
		argCount++
		baseQuery += ` AND track_serial = $` + string(rune(argCount+'0'))
		countQuery += ` AND track_serial = $` + string(rune(argCount+'0'))
		args = append(args, req.TrackSerial)
	}

	if req.TrackIndividual != nil {
		argCount++
		baseQuery += ` AND track_individual = $` + string(rune(argCount+'0'))
		countQuery += ` AND track_individual = $` + string(rune(argCount+'0'))
		args = append(args, req.TrackIndividual)
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Add sorting
	sortBy := "name"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}

	sortOrder := "ASC"
	if req.SortOrder == "DESC" {
		sortOrder = "DESC"
	}

	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	baseQuery += ` ORDER BY ` + sortBy + ` ` + sortOrder + ` LIMIT $` + string(rune(argCount+1+'0')) + ` OFFSET $` + string(rune(argCount+2+'0'))
	args = append(args, req.PerPage, offset)

	// Execute query
	query := `SELECT id, organization_id, category_id, brand_id, name, description, model, sku, barcode, track_serial, track_individual, min_stock_level, warranty_days, created_at, updated_at ` + baseQuery
	var products []Product
	err = r.db.SelectContext(ctx, &products, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// UpdateProduct updates a product
func (r *Repository) UpdateProduct(ctx context.Context, product *Product) error {
	query := `
		UPDATE products
		SET category_id = $1, brand_id = $2, name = $3, description = $4, model = $5, sku = $6, barcode = $7, track_serial = $8, track_individual = $9, min_stock_level = $10, warranty_days = $11, updated_at = $12
		WHERE id = $13 AND organization_id = $14
	`
	product.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		product.CategoryID,
		product.BrandID,
		product.Name,
		product.Description,
		product.Model,
		product.SKU,
		product.Barcode,
		product.TrackSerial,
		product.TrackIndividual,
		product.MinStockLevel,
		product.WarrantyDays,
		product.UpdatedAt,
		product.ID,
		product.OrganizationID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrProductNotFound
	}

	return nil
}

// DeleteProduct deletes a product
func (r *Repository) DeleteProduct(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1 AND organization_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrProductNotFound
	}

	return nil
}

// GetProductStockCount returns the stock count for a product
func (r *Repository) GetProductStockCount(ctx context.Context, productID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM inventory_items
		WHERE product_id = $1 AND status = 'AVAILABLE'
	`
	var count int
	err := r.db.GetContext(ctx, &count, query, productID)
	return count, err
}
