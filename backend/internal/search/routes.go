package search

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/middleware"
)

// RegisterRoutes registers search routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	// Search routes with permission middleware
	router.POST("/search", middleware.RequirePermission("products", "read"), handler.Search)
	router.GET("/search/stats", middleware.RequirePermission("products", "read"), handler.GetSearchStats)
}
