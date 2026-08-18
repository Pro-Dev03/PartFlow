package users

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	users := router.Group("/users")
	{
		users.POST("", middleware.RequirePermission("users", "create"), handler.CreateUser)
		users.GET("", middleware.RequirePermission("users", "read"), handler.ListUsers)
		users.GET("/:id", middleware.RequirePermission("users", "read"), handler.GetUser)
		users.PUT("/:id", middleware.RequirePermission("users", "update"), handler.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission("users", "delete"), handler.DeleteUser)
		users.POST("/:id/role", middleware.RequirePermission("users", "manage"), handler.AssignRole)
		users.DELETE("/:id/role", middleware.RequirePermission("users", "manage"), handler.RemoveRole)
		users.GET("/:id/roles", middleware.RequirePermission("users", "read"), handler.GetUserRoles)
		users.POST("/change-password", handler.ChangePassword)
	}
}