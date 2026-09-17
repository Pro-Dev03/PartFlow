package assistant

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Reply(c *gin.Context) {
	var request ReplyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_ASSISTANT_REQUEST", "message": "أدخل رسالة صحيحة للمساعد."}})
		return
	}
	response, err := h.service.Reply(c.Request.Context(), request)
	if err != nil {
		log.Printf("assistant reply failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "ASSISTANT_UNAVAILABLE", "message": "تعذر جلب بيانات المتجر حاليًا. جرّب مرة ثانية."}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}
