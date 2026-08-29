package users

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	users := router.Group("/users")
	{
		users.POST("", handler.CreateUser)
		users.GET("", handler.ListUsers)
		users.GET("/subscription-summary", handler.GetSubscriptionSummary)
		users.GET("/subscriptions", handler.ListSubscriptionAccounts)
		users.POST("/change-password", handler.ChangePassword)
		users.GET("/:id", handler.GetUser)
		users.PUT("/:id", handler.UpdateUser)
		users.PUT("/:id/subscription", handler.UpdateSubscriptionStatus)
		users.POST("/:id/subscription/renew", handler.RenewSubscriptionByDays)
		users.DELETE("/:id", handler.DeleteUser)
	}
}
