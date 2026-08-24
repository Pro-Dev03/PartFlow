package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/pkg/response"
)

type Handler struct {
	service *CachedService
}

func NewHandler(service *CachedService) *Handler {
	return &Handler{service: service}
}

// GetDashboardStats handles dashboard statistics retrieval
func (h *Handler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, stats, "Dashboard statistics retrieved successfully")
}