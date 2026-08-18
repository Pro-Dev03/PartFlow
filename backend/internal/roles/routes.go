package roles

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	roles := router.Group("/roles")
	roles.Use(middleware.RequirePermission("users", "manage"))
	{
		roles.POST("", handler.CreateRole)
		roles.GET("", handler.ListRoles)
		roles.GET("/:id", handler.GetRole)
		roles.PUT("/:id", handler.UpdateRole)
		roles.DELETE("/:id", handler.DeleteRole)
		roles.POST("/initialize", handler.InitializeStandardRoles)
	}
}