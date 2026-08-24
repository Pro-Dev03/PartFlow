package search

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers search routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	// Search routes
	router.POST("/search", handler.Search)
	router.GET("/search/stats", handler.GetSearchStats)
}
