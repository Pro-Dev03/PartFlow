package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetDashboardStats handles dashboard statistics retrieval
func (h *Handler) GetDashboardStats(c *gin.Context) {
	organizationID, exists := c.Get("organization_id")
	if !exists {
		response.BadRequest(c, "organization_id required")
		return
	}

	stats, err := h.service.GetDashboardStats(c.Request.Context(), organizationID.(uuid.UUID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, stats, "Dashboard statistics retrieved successfully")
}