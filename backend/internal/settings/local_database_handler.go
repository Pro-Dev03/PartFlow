package settings

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/partflow/smart-store/internal/localdb"
)

type LocalDatabaseHandler struct{}

type OperatingModeRequest struct {
	Mode string `json:"mode"`
}

func NewLocalDatabaseHandler() *LocalDatabaseHandler {
	return &LocalDatabaseHandler{}
}

func (h *LocalDatabaseHandler) SetOperatingMode(c *gin.Context) {
	var request OperatingModeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تعذر قراءة وضع التشغيل", "details": err.Error()})
		return
	}
	if request.Mode != "offline" && request.Mode != "online" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "وضع التشغيل غير صالح", "details": "mode must be offline or online"})
		return
	}

	database, err := localdb.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "فشل تهيئة قاعدة البيانات المحلية",
			"details": err.Error(),
		})
		return
	}
	defer database.DB.Close()

	if err := localdb.SetMetadata(database.DB, "operating_mode", request.Mode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "فشل حفظ وضع التشغيل المحلي",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"mode":                 request.Mode,
			"local_database":       database.Path,
			"database_initialized": true,
		},
	})
}
