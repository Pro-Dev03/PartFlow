package organizations

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	orgs := router.Group("/organizations")
	orgs.Use(middleware.RequirePermission("settings", "manage"))
	{
		orgs.POST("", handler.CreateOrganization)
		orgs.GET("", handler.ListOrganizations)
		orgs.GET("/:id", handler.GetOrganization)
		orgs.PUT("/:id", handler.UpdateOrganization)
		orgs.DELETE("/:id", handler.DeleteOrganization)
	}
}