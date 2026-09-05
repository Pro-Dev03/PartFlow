package products

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles products data operations
type Repository struct {
	db *sqlx.DB
}

// scanBrandRow keeps SQLite TEXT timestamps compatible with the time.Time
// fields exposed by the API. PostgreSQL and SQLite both safely scan their
// UUID/timestamp values into strings here, after which they are parsed once.
func scanBrandRow(scanner interface{ Scan(...any) error }) (Brand, error) {
	var brand Brand
	var idRaw, createdRaw, updatedRaw string
	if err := scanner.Scan(&idRaw, &brand.Name, &brand.Description, &brand.LogoURL, &createdRaw, &updatedRaw); err != nil {
		return Brand{}, err
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return Brand{}, fmt.Errorf("parse brand id: %w", err)
	}
	brand.ID = id
	brand.CreatedAt, err = parseSQLiteTimestamp(createdRaw)
	if err != nil {
		return Brand{}, fmt.Errorf("parse brand created_at: %w", err)
	}
	brand.UpdatedAt, err = parseSQLiteTimestamp(updatedRaw)
	if err != nil {
		return Brand{}, fmt.Errorf("parse brand updated_at: %w", err)
	}
	return brand, nil
}

// NewRepository creates a new products repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func parseSQLiteTimestamp(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, nil
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999 -0700",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported SQLite timestamp format: %q", raw)
}

func scanProductRow(rows *sql.Rows) (Product, error) {
	var product Product
	var description, model, barcode, deletedAt, createdAt, updatedAt sql.NullString
	var categoryID, brandID, preferredSupplierID sql.NullString

	if err := rows.Scan(
		&product.ID,
		&categoryID,
		&brandID,
		&preferredSupplierID,
		&product.Name,
		&description,
		&model,
		&product.SKU,
		&barcode,
		&product.CostPrice,
		&product.SellingPrice,
		&product.TrackSerial,
		&product.TrackIndividual,
		&product.MinStockLevel,
		&product.WarrantyDays,
		&product.IsActive,
		&deletedAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Product{}, err
	}

	if categoryID.Valid {
		id, err := uuid.Parse(categoryID.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse category id: %w", err)
		}
		product.CategoryID = &id
	}
	if brandID.Valid {
		id, err := uuid.Parse(brandID.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse brand id: %w", err)
		}
		product.BrandID = &id
	}
	if preferredSupplierID.Valid {
		id, err := uuid.Parse(preferredSupplierID.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse preferred supplier id: %w", err)
		}
		product.PreferredSupplierID = &id
	}
	if description.Valid && strings.TrimSpace(description.String) != "" {
		v := description.String
		product.Description = &v
	}
	if model.Valid && strings.TrimSpace(model.String) != "" {
		v := model.String
		product.Model = &v
	}
	if barcode.Valid && strings.TrimSpace(barcode.String) != "" {
		product.Barcode = barcode.String
	}
	if deletedAt.Valid && strings.TrimSpace(deletedAt.String) != "" {
		parsed, err := parseSQLiteTimestamp(deletedAt.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse deleted_at: %w", err)
		}
		product.DeletedAt = &parsed
	}
	if createdAt.Valid && strings.TrimSpace(createdAt.String) != "" {
		parsed, err := parseSQLiteTimestamp(createdAt.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse created_at: %w", err)
		}
		product.CreatedAt = parsed
	}
	if updatedAt.Valid && strings.TrimSpace(updatedAt.String) != "" {
		parsed, err := parseSQLiteTimestamp(updatedAt.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse updated_at: %w", err)
		}
		product.UpdatedAt = parsed
	}
	return product, nil
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
	var description, parentID, icon, color, createdAt, updatedAt sql.NullString
	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &description, &parentID, &icon, &color,
		&category.IsActive, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	if description.Valid {
		category.Description = description.String
	}
	if parentID.Valid && strings.TrimSpace(parentID.String) != "" {
		parsed, parseErr := uuid.Parse(parentID.String)
		if parseErr != nil {
			return nil, fmt.Errorf("parse parent id: %w", parseErr)
		}
		category.ParentID = &parsed
	}
	if icon.Valid {
		category.Icon = &icon.String
	}
	if color.Valid {
		category.Color = &color.String
	}
	category.CreatedAt, err = parseSQLiteTimestamp(createdAt.String)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	category.UpdatedAt, err = parseSQLiteTimestamp(updatedAt.String)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	return &category, err
}

// ListCategories retrieves all categories
func (r *Repository) ListCategories(ctx context.Context) ([]Category, error) {
	query := `
		SELECT id, name, description, parent_id, icon, color, is_active, created_at, updated_at
		FROM categories
		ORDER BY name
	`
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		var description, parentID, icon, color, createdAt, updatedAt sql.NullString
		if err := rows.Scan(
			&category.ID, &category.Name, &description, &parentID, &icon, &color,
			&category.IsActive, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		if description.Valid {
			category.Description = description.String
		}
		if parentID.Valid && strings.TrimSpace(parentID.String) != "" {
			parsed, parseErr := uuid.Parse(parentID.String)
			if parseErr != nil {
				return nil, fmt.Errorf("parse parent id: %w", parseErr)
			}
			category.ParentID = &parsed
		}
		if icon.Valid {
			category.Icon = &icon.String
		}
		if color.Valid {
			category.Color = &color.String
		}
		category.CreatedAt, err = parseSQLiteTimestamp(createdAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		category.UpdatedAt, err = parseSQLiteTimestamp(updatedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
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
func (r *Repository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
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
		SELECT id, name, COALESCE(description, '') AS description, COALESCE(logo_url, '') AS logo_url, created_at, updated_at
		FROM brands
		WHERE id = $1
	`
	row := r.db.QueryRowxContext(ctx, query, id)
	brand, err := scanBrandRow(row)
	if err == sql.ErrNoRows {
		return nil, ErrBrandNotFound
	}
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

// ListBrands retrieves all brands
func (r *Repository) ListBrands(ctx context.Context) ([]Brand, error) {
	query := `
		SELECT id, name, COALESCE(description, '') AS description, COALESCE(logo_url, '') AS logo_url, created_at, updated_at
		FROM brands
		ORDER BY name
	`
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	brands := make([]Brand, 0)
	for rows.Next() {
		brand, scanErr := scanBrandRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		brands = append(brands, brand)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return brands, nil
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
func (r *Repository) DeleteBrand(ctx context.Context, id uuid.UUID) error {
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
func (r *Repository) GetProductByBarcode(ctx context.Context, barcode string) (*Product, error) {
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
	// Build base query with current quantity from inventory
	baseQuery := `
		SELECT p.id, p.category_id, p.brand_id, p.preferred_supplier_id, p.name, p.description, p.model, p.sku, p.barcode, p.cost_price, p.selling_price, p.track_serial, p.track_individual, p.min_stock_level, p.warranty_days, p.is_active, p.deleted_at, p.created_at, p.updated_at
		FROM products p
		WHERE p.deleted_at IS NULL
	`
	countQuery := `SELECT COUNT(DISTINCT p.id) FROM products p WHERE p.deleted_at IS NULL`

	args := []interface{}{}
	argCount := 0
	likeOperator := "ILIKE"
	if r.db.DriverName() == "sqlite" {
		likeOperator = "LIKE"
	}

	// Add filters
	if req.CategoryID != nil {
		argCount++
		paramNum := argCount
		baseQuery += ` AND category_id = $` + fmt.Sprint(paramNum)
		countQuery += ` AND category_id = $` + fmt.Sprint(paramNum)
		args = append(args, req.CategoryID)
	}

	if req.BrandID != nil {
		argCount++
		paramNum := argCount
		baseQuery += ` AND brand_id = $` + fmt.Sprint(paramNum)
		countQuery += ` AND brand_id = $` + fmt.Sprint(paramNum)
		args = append(args, req.BrandID)
	}

	if req.Search != "" {
		argCount++
		searchPattern := "%" + req.Search + "%"
		baseQuery += ` AND (name ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR model ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR sku ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR description ` + likeOperator + ` $` + fmt.Sprint(argCount) + `)`
		countQuery += ` AND (name ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR model ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR sku ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR description ` + likeOperator + ` $` + fmt.Sprint(argCount) + `)`
		args = append(args, searchPattern)
	}

	if req.TrackSerial != nil {
		argCount++
		paramNum := argCount
		baseQuery += ` AND track_serial = $` + fmt.Sprint(paramNum)
		countQuery += ` AND track_serial = $` + fmt.Sprint(paramNum)
		args = append(args, req.TrackSerial)
	}

	if req.TrackIndividual != nil {
		argCount++
		paramNum := argCount
		baseQuery += ` AND track_individual = $` + fmt.Sprint(paramNum)
		countQuery += ` AND track_individual = $` + fmt.Sprint(paramNum)
		args = append(args, req.TrackIndividual)
	}

	// Advanced filtering: low stock only
	if req.LowStockOnly != nil && *req.LowStockOnly {
		baseQuery += ` AND (
			SELECT COALESCE(COUNT(*), 0)
			FROM inventory_items ii
			WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND ii.condition <> 'USED'
		) < p.min_stock_level
		AND p.min_stock_level > 0`
		countQuery += ` AND (
			SELECT COALESCE(COUNT(*), 0)
			FROM inventory_items ii
			WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE' AND ii.condition <> 'USED'
		) < p.min_stock_level
		AND p.min_stock_level > 0`
	}

	// Advanced filtering: in stock only
	if req.InStockOnly != nil && *req.InStockOnly {
		baseQuery += ` AND (
			SELECT COALESCE(COUNT(*), 0)
			FROM inventory_items ii
			WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE'
		) > 0`
		countQuery += ` AND (
			SELECT COALESCE(COUNT(*), 0)
			FROM inventory_items ii
			WHERE ii.product_id = p.id AND ii.status = 'AVAILABLE'
		) > 0`
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
	paramNum := argCount + 1
	baseQuery += ` ORDER BY ` + sortBy + ` ` + sortOrder + ` LIMIT $` + fmt.Sprint(paramNum) + ` OFFSET $` + fmt.Sprint(paramNum+1)
	args = append(args, req.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]Product, 0, req.PerPage)
	for rows.Next() {
		product, err := scanProductRow(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
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

func (r *Repository) UpdateMinimumStock(ctx context.Context, id uuid.UUID, minStockLevel int) error {
	result, err := r.db.ExecContext(ctx, `UPDATE products SET min_stock_level = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`, minStockLevel, time.Now(), id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
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
	likeOperator := "ILIKE"
	if r.db.DriverName() == "sqlite" {
		likeOperator = "LIKE"
	}
	searchQuery := `
		SELECT id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		FROM products
		WHERE (name ` + likeOperator + ` $1 OR model ` + likeOperator + ` $1 OR sku ` + likeOperator + ` $1 OR barcode ` + likeOperator + ` $1)
		AND is_active = true
		AND deleted_at IS NULL
		ORDER BY name
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	products := make([]Product, 0)
	for rows.Next() {
		product, scanErr := scanProductRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}
