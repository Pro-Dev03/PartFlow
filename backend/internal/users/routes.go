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
		users.POST("/change-password", handler.ChangePassword)
	}

	// Account and subscription management must never be available to a regular
	// subscriber. The outer router already applies JWT/cloud authentication.
	adminUsers := router.Group("/users")
	adminUsers.Use(middleware.Admin())
	{
		adminUsers.POST("", handler.CreateUser)
		adminUsers.GET("", handler.ListUsers)
		adminUsers.GET("/subscription-summary", handler.GetSubscriptionSummary)
		adminUsers.GET("/subscriptions", handler.ListSubscriptionAccounts)
		adminUsers.GET("/:id", handler.GetUser)
		adminUsers.PUT("/:id", handler.UpdateUser)
		adminUsers.PUT("/:id/subscription", handler.UpdateSubscriptionStatus)
		adminUsers.POST("/:id/subscription/renew", handler.RenewSubscriptionByDays)
		adminUsers.DELETE("/:id", handler.DeleteUser)
	}
}
