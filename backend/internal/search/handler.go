package search

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles HTTP requests for search
type Handler struct {
	service *Service
}

// NewHandler creates a new search handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Search handles global search
func (h *Handler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	results, err := h.service.Search(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, results, "Search completed successfully")
}

// SearchGET mirrors the JSON search endpoint for the desktop/web client's
// global search box, which issues a query-string request. Keeping both forms
// avoids a silent 404 when the user searches from the navigation bar.
func (h *Handler) SearchGET(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		query = strings.TrimSpace(c.Query("query"))
	}
	if query == "" {
		response.BadRequest(c, "query is required")
		return
	}
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	offset := 0
	if raw := c.Query("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			offset = parsed
		}
	}
	types := make([]string, 0)
	if raw := c.Query("types"); raw != "" {
		for _, item := range strings.Split(raw, ",") {
			if item = strings.TrimSpace(item); item != "" {
				types = append(types, item)
			}
		}
	}
	results, err := h.service.Search(c.Request.Context(), &SearchRequest{Query: query, Types: types, Limit: limit, Offset: offset})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, results, "Search completed successfully")
}

// GetSearchStats handles search statistics retrieval
func (h *Handler) GetSearchStats(c *gin.Context) {
	stats, err := h.service.GetSearchStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, stats, "Search statistics retrieved successfully")
}
