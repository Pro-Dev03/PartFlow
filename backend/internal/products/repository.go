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
		INSERT INTO categories (id, name, description, parent_id, icon, color, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	category.ID = uuid.New()
	category.CreatedAt = now
	category.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		category.ID,
		category.Name,
		category.Description,
		category.ParentID,
		category.Icon,
		category.Color,
		category.IsActive,
		category.CreatedAt,
		category.UpdatedAt,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	return err
}

// GetCategoryByID retrieves a category by ID
func (r *Repository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	query := `
		SELECT id, name, description, parent_id, icon, color, is_active, created_at, updated_at
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

// ListCategories retrieves all categories
func (r *Repository) ListCategories(ctx context.Context, ) ([]Category, error) {
	query := `
		SELECT id, name, description, parent_id, icon, color, is_active, created_at, updated_at
		FROM categories
		ORDER BY name
	`
	var categories []Category
	err := r.db.SelectContext(ctx, &categories, query)
	return categories, err
}

// UpdateCategory updates a category
func (r *Repository) UpdateCategory(ctx context.Context, category *Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, parent_id = $3, icon = $4, color = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`
	category.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		category.Name,
		category.Description,
		category.ParentID,
		category.Icon,
		category.Color,
		category.IsActive,
		category.UpdatedAt,
		category.ID,
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
func (r *Repository) DeleteCategory(ctx context.Context, id uuid.UUID, ) error {
	// First, set category_id to NULL for all products in this category
	updateQuery := `UPDATE products SET category_id = NULL WHERE category_id = $1`
	_, err := r.db.ExecContext(ctx, updateQuery, id)
	if err != nil {
		return err
	}

	// Now delete the category
	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
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
		INSERT INTO brands (id, name, description, logo_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	brand.ID = uuid.New()
	brand.CreatedAt = now
	brand.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		brand.ID,
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
		SELECT id, name, description, logo_url, created_at, updated_at
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

// ListBrands retrieves all brands
func (r *Repository) ListBrands(ctx context.Context, ) ([]Brand, error) {
	query := `
		SELECT id, name, description, logo_url, created_at, updated_at
		FROM brands
		ORDER BY name
	`
	var brands []Brand
	err := r.db.SelectContext(ctx, &brands, query)
	return brands, err
}

// UpdateBrand updates a brand
func (r *Repository) UpdateBrand(ctx context.Context, brand *Brand) error {
	query := `
		UPDATE brands
		SET name = $1, description = $2, logo_url = $3, updated_at = $4
	`
	brand.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		brand.Name,
		brand.Description,
		brand.LogoURL,
		brand.UpdatedAt,
		brand.ID,
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
func (r *Repository) DeleteBrand(ctx context.Context, id uuid.UUID, ) error {
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

	query := `DELETE FROM brands WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
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
		INSERT INTO products (id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	product.ID = uuid.New()
	product.CreatedAt = now
	product.UpdatedAt = now
	product.IsActive = true

	err := r.db.QueryRowContext(ctx, query,
		product.ID,
		product.CategoryID,
		product.BrandID,
		product.PreferredSupplierID,
		product.Name,
		product.Description,
		product.Model,
		product.SKU,
		product.Barcode,
		product.CostPrice,
		product.SellingPrice,
		product.TrackSerial,
		product.TrackIndividual,
		product.MinStockLevel,
		product.WarrantyDays,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	return err
}

// GetProductByID retrieves a product by ID
func (r *Repository) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	query := `
		SELECT id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`
	var product Product
	err := r.db.GetContext(ctx, &product, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	return &product, err
}

// GetProductByBarcode retrieves a product by barcode
func (r *Repository) GetProductByBarcode(ctx context.Context, barcode string, ) (*Product, error) {
	query := `
		SELECT id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		FROM products
		WHERE barcode = $1 AND deleted_at IS NULL
	`
	var product Product
	err := r.db.GetContext(ctx, &product, query, barcode)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	return &product, err
}

// ListProducts retrieves products with pagination and filters
func (r *Repository) ListProducts(ctx context.Context, req *ProductListRequest) ([]Product, int, error) {
	// Build base query
	baseQuery := `
		SELECT id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		FROM products
		WHERE deleted_at IS NULL
	`
	countQuery := `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`

	args := []interface{}{}
	argCount := 0

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
	var products []Product
	err = r.db.SelectContext(ctx, &products, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// UpdateProduct updates a product
func (r *Repository) UpdateProduct(ctx context.Context, product *Product) error {
	query := `
		UPDATE products
		SET category_id = $1, brand_id = $2, preferred_supplier_id = $3, name = $4, description = $5, model = $6, sku = $7, barcode = $8, cost_price = $9, selling_price = $10, track_serial = $11, track_individual = $12, min_stock_level = $13, warranty_days = $14, updated_at = $15
		WHERE id = $16
	`
	product.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		product.CategoryID,
		product.BrandID,
		product.PreferredSupplierID,
		product.Name,
		product.Description,
		product.Model,
		product.SKU,
		product.Barcode,
		product.CostPrice,
		product.SellingPrice,
		product.TrackSerial,
		product.TrackIndividual,
		product.MinStockLevel,
		product.WarrantyDays,
		product.UpdatedAt,
		product.ID,
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

// DeleteProduct deletes a product (soft delete)
func (r *Repository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
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

// RestoreProduct restores a soft-deleted product
func (r *Repository) RestoreProduct(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
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

// ArchiveProduct archives a product (sets is_active to false)
func (r *Repository) ArchiveProduct(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET is_active = false, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
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

// GetAvailableItemCount returns the count of available items for a product
func (r *Repository) GetAvailableItemCount(ctx context.Context, productID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM inventory_items
		WHERE product_id = $1 AND status = 'AVAILABLE'
	`
	var count int
	err := r.db.GetContext(ctx, &count, query, productID)
	return count, err
}

// GetReservedItemCount returns the count of reserved items for a product
func (r *Repository) GetReservedItemCount(ctx context.Context, productID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM inventory_items
		WHERE product_id = $1 AND status = 'RESERVED'
	`
	var count int
	err := r.db.GetContext(ctx, &count, query, productID)
	return count, err
}

// SearchProducts searches products by name, SKU, or barcode
func (r *Repository) SearchProducts(ctx context.Context, query string, limit int) ([]Product, error) {
	searchQuery := `
		SELECT id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		FROM products
		WHERE (name ILIKE $1 OR model ILIKE $1 OR sku ILIKE $1 OR barcode ILIKE $1)
		AND is_active = true
		AND deleted_at IS NULL
		ORDER BY name
		LIMIT $2
	`

	var products []Product
	err := r.db.SelectContext(ctx, &products, searchQuery, "%"+query+"%", limit)
	return products, err
}
