package pagination

import (
	"math"
)

// Pagination represents pagination parameters
type Pagination struct {
	Page       int    `json:"page" form:"page"`
	PerPage    int    `json:"per_page" form:"per_page"`
	SortBy     string `json:"sort_by" form:"sort_by"`
	SortOrder  string `json:"sort_order" form:"sort_order"`
	Total      int64  `json:"total"`
	TotalPages int    `json:"total_pages"`
}

// CursorPagination represents cursor-based pagination
type CursorPagination struct {
	Limit      int    `json:"limit" form:"limit"`
	Cursor     string `json:"cursor" form:"cursor"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

// NewPagination creates a new pagination instance with defaults
func NewPagination(page, perPage int, sortBy, sortOrder string) *Pagination {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	return &Pagination{
		Page:      page,
		PerPage:   perPage,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}
}

// GetOffset calculates the offset for SQL queries
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// GetLimit returns the limit for SQL queries
func (p *Pagination) GetLimit() int {
	return p.PerPage
}

// CalculateTotalPages calculates total pages based on total records
func (p *Pagination) CalculateTotalPages(total int64) {
	p.Total = total
	p.TotalPages = int(math.Ceil(float64(total) / float64(p.PerPage)))
}

// GetOrderBy returns the ORDER BY clause for SQL
func (p *Pagination) GetOrderBy() string {
	return p.SortBy + " " + p.SortOrder
}

// NewCursorPagination creates a new cursor pagination instance
func NewCursorPagination(limit int, cursor string) *CursorPagination {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return &CursorPagination{
		Limit:  limit,
		Cursor: cursor,
	}
}
