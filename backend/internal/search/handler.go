package search

import (
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

// GetSearchStats handles search statistics retrieval
func (h *Handler) GetSearchStats(c *gin.Context) {
	stats, err := h.service.GetSearchStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, stats, "Search statistics retrieved successfully")
}
