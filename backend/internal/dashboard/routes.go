package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers dashboard routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	service := NewCachedService(db)
	handler := NewHandler(service)

	// Dashboard routes
	router.GET("/stats", handler.GetDashboardStats)
}
