package search

import (
	"time"

	"github.com/google/uuid"
)

// SearchResult represents a generic search result
type SearchResult struct {
	Type       string                 `json:"type"`       // product, customer, supplier, sale, purchase, etc.
	ID         uuid.UUID              `json:"id"`
	Title      string                 `json:"title"`
	Subtitle   string                 `json:"subtitle"`
	Metadata   map[string]interface{} `json:"metadata"`
	Score      float64                `json:"score"`
	CreatedAt  time.Time              `json:"created_at"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string   `json:"query" binding:"required"`
	Types      []string `json:"types"`       // If empty, search all types
	Limit      int      `json:"limit" binding:"omitempty,min=1,max=100"`
	Offset     int      `json:"offset" binding:"omitempty,min=0"`
	Filters    map[string]interface{} `json:"filters"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Query       string          `json:"query"`
	Results     []SearchResult `json:"results"`
	Total       int            `json:"total"`
	Limit       int            `json:"limit"`
	Offset      int            `json:"offset"`
	HasMore     bool           `json:"has_more"`
}

// SearchStats represents search statistics
type SearchStats struct {
	TotalProducts   int `json:"total_products"`
	TotalCustomers  int `json:"total_customers"`
	TotalSuppliers  int `json:"total_suppliers"`
	TotalSales      int `json:"total_sales"`
	TotalPurchases  int `json:"total_purchases"`
}
