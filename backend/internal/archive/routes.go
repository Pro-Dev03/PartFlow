package archive

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	handler := NewHandler(db)
	archive := router.Group("/archive")
	archive.GET("/events", middleware.RequirePermission("archive.read"), handler.ListEvents)
}
