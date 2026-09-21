package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/middleware"
)

// RegisterRoutes registers audit log routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	audit := router.Group("/audit")
	audit.Use(middleware.RequirePermission("audit.read"))
	{
		audit.POST("", handler.CreateAuditLog)
		audit.GET("/:id", handler.GetAuditLog)
		audit.GET("", handler.ListAuditLogs)
		audit.GET("/summary", handler.GetAuditLogSummary)
		audit.GET("/stats", handler.GetAuditStats)
		audit.GET("/export", middleware.RequirePermission("archive.export"), handler.ExportAuditLogs)
		audit.GET("/user/:user_id", handler.GetUserAuditLogs)
		audit.GET("/entity/:entity_id", handler.GetEntityAuditLogs)
	}
}
