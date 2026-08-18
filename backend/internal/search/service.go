package search

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Service handles search business logic
type Service struct {
	db *sqlx.DB
}

// NewService creates a new search service
func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// Search performs a global search across all entities
func (s *Service) Search(ctx context.Context, organizationID uuid.UUID, req *SearchRequest) (*SearchResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	var results []SearchResult
	total := 0

	// If no types specified, search all
	types := req.Types
	if len(types) == 0 {
		types = []string{"products", "customers", "suppliers", "sales", "purchases"}
	}

	// Search each type
	for _, searchType := range types {
		var typeResults []SearchResult
		var typeTotal int

		switch searchType {
		case "products":
			typeResults, typeTotal = s.searchProducts(ctx, organizationID, req.Query, req.Limit, req.Offset)
		case "customers":
			typeResults, typeTotal = s.searchCustomers(ctx, organizationID, req.Query, req.Limit, req.Offset)
		case "suppliers":
			typeResults, typeTotal = s.searchSuppliers(ctx, organizationID, req.Query, req.Limit, req.Offset)
		case "sales":
			typeResults, typeTotal = s.searchSales(ctx, organizationID, req.Query, req.Limit, req.Offset)
		case "purchases":
			typeResults, typeTotal = s.searchPurchases(ctx, organizationID, req.Query, req.Limit, req.Offset)
		}

		results = append(results, typeResults...)
		total += typeTotal
	}

	return &SearchResponse{
		Query:   req.Query,
		Results: results,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: req.Offset + req.Limit < total,
	}, nil
}

// searchProducts searches for products
func (s *Service) searchProducts(ctx context.Context, organizationID uuid.UUID, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	queryStr := `
		SELECT id, name, sku, model, barcode, created_at
		FROM products
		WHERE organization_id = $1
		AND is_active = true
		AND (name ILIKE $2 OR sku ILIKE $2 OR model ILIKE $2 OR barcode ILIKE $2)
		ORDER BY name
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.QueryContext(ctx, queryStr, organizationID, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, sku, model, barcode string
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &sku, &model, &barcode, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"sku":     sku,
			"model":   model,
			"barcode": barcode,
		}

		results = append(results, SearchResult{
			Type:      "product",
			ID:        id,
			Title:     name,
			Subtitle:  fmt.Sprintf("SKU: %s", sku),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt,
		})
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM products
		WHERE organization_id = $1
		AND is_active = true
		AND (name ILIKE $2 OR sku ILIKE $2 OR model ILIKE $2 OR barcode ILIKE $2)
	`
	s.db.GetContext(ctx, &total, countQuery, organizationID, searchPattern)

	return results, total
}

// searchCustomers searches for customers
func (s *Service) searchCustomers(ctx context.Context, organizationID uuid.UUID, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	queryStr := `
		SELECT id, name, email, phone, created_at
		FROM customers
		WHERE organization_id = $1
		AND is_active = true
		AND (name ILIKE $2 OR email ILIKE $2 OR phone ILIKE $2)
		ORDER BY name
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.QueryContext(ctx, queryStr, organizationID, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, email, phone string
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &email, &phone, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"email": email,
			"phone": phone,
		}

		results = append(results, SearchResult{
			Type:      "customer",
			ID:        id,
			Title:     name,
			Subtitle:  fmt.Sprintf("Email: %s", email),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt,
		})
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM customers
		WHERE organization_id = $1
		AND is_active = true
		AND (name ILIKE $2 OR email ILIKE $2 OR phone ILIKE $2)
	`
	s.db.GetContext(ctx, &total, countQuery, organizationID, searchPattern)

	return results, total
}

// searchSuppliers searches for suppliers
func (s *Service) searchSuppliers(ctx context.Context, organizationID uuid.UUID, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	queryStr := `
		SELECT id, name, email, phone, created_at
		FROM suppliers
		WHERE organization_id = $1
		AND is_active = true
		AND (name ILIKE $2 OR email ILIKE $2 OR phone ILIKE $2)
		ORDER BY name
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.QueryContext(ctx, queryStr, organizationID, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, email, phone string
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &email, &phone, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"email": email,
			"phone": phone,
		}

		results = append(results, SearchResult{
			Type:      "supplier",
			ID:        id,
			Title:     name,
			Subtitle:  fmt.Sprintf("Email: %s", email),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt,
		})
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM suppliers
		WHERE organization_id = $1
		AND is_active = true
		AND (name ILIKE $2 OR email ILIKE $2 OR phone ILIKE $2)
	`
	s.db.GetContext(ctx, &total, countQuery, organizationID, searchPattern)

	return results, total
}

// searchSales searches for sales
func (s *Service) searchSales(ctx context.Context, organizationID uuid.UUID, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	queryStr := `
		SELECT id, invoice_number, total_amount, sale_date, created_at
		FROM sales
		WHERE organization_id = $1
		AND (invoice_number ILIKE $2)
		ORDER BY sale_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.QueryContext(ctx, queryStr, organizationID, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var invoiceNumber string
		var totalAmount float64
		var saleDate, createdAt time.Time

		if err := rows.Scan(&id, &invoiceNumber, &totalAmount, &saleDate, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"total_amount": totalAmount,
			"sale_date":    saleDate,
		}

		results = append(results, SearchResult{
			Type:      "sale",
			ID:        id,
			Title:     invoiceNumber,
			Subtitle:  fmt.Sprintf("Amount: %.2f", totalAmount),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt,
		})
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM sales
		WHERE organization_id = $1
		AND invoice_number ILIKE $2
	`
	s.db.GetContext(ctx, &total, countQuery, organizationID, searchPattern)

	return results, total
}

// searchPurchases searches for purchases
func (s *Service) searchPurchases(ctx context.Context, organizationID uuid.UUID, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	queryStr := `
		SELECT id, invoice_number, total_amount, purchase_date, created_at
		FROM purchases
		WHERE organization_id = $1
		AND (invoice_number ILIKE $2)
		ORDER BY purchase_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.QueryContext(ctx, queryStr, organizationID, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var invoiceNumber string
		var totalAmount float64
		var purchaseDate, createdAt time.Time

		if err := rows.Scan(&id, &invoiceNumber, &totalAmount, &purchaseDate, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"total_amount":    totalAmount,
			"purchase_date":   purchaseDate,
		}

		results = append(results, SearchResult{
			Type:      "purchase",
			ID:        id,
			Title:     invoiceNumber,
			Subtitle:  fmt.Sprintf("Amount: %.2f", totalAmount),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt,
		})
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM purchases
		WHERE organization_id = $1
		AND invoice_number ILIKE $2
	`
	s.db.GetContext(ctx, &total, countQuery, organizationID, searchPattern)

	return results, total
}

// GetSearchStats retrieves search statistics
func (s *Service) GetSearchStats(ctx context.Context, organizationID uuid.UUID) (*SearchStats, error) {
	stats := &SearchStats{}

	// Get total products
	query := `SELECT COUNT(*) FROM products WHERE organization_id = $1 AND is_active = true`
	err := s.db.GetContext(ctx, &stats.TotalProducts, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total products: %w", err)
	}

	// Get total customers
	query = `SELECT COUNT(*) FROM customers WHERE organization_id = $1 AND is_active = true`
	err = s.db.GetContext(ctx, &stats.TotalCustomers, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total customers: %w", err)
	}

	// Get total suppliers
	query = `SELECT COUNT(*) FROM suppliers WHERE organization_id = $1 AND is_active = true`
	err = s.db.GetContext(ctx, &stats.TotalSuppliers, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total suppliers: %w", err)
	}

	// Get total sales
	query = `SELECT COUNT(*) FROM sales WHERE organization_id = $1`
	err = s.db.GetContext(ctx, &stats.TotalSales, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sales: %w", err)
	}

	// Get total purchases
	query = `SELECT COUNT(*) FROM purchases WHERE organization_id = $1`
	err = s.db.GetContext(ctx, &stats.TotalPurchases, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total purchases: %w", err)
	}

	return stats, nil
}
