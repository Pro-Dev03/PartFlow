package notifications

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers notifications routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Notification routes
	notifications := router.Group("/notifications")
	{
		notifications.POST("", handler.CreateNotification)
		notifications.GET("/:id", handler.GetNotification)
		notifications.GET("", handler.ListNotifications)
		notifications.PUT("/:id/read", handler.MarkAsRead)
		notifications.PUT("/read-all", handler.MarkAllAsRead)
		notifications.DELETE("/:id", handler.DeleteNotification)
		notifications.GET("/unread-count", handler.GetUnreadCount)
	}
}
