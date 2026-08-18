package payments

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers payments routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Payment routes
	payments := router.Group("/payments")
	{
		payments.POST("", handler.CreatePayment)
		payments.GET("/:id", handler.GetPayment)
		payments.GET("", handler.ListPayments)
		payments.PUT("/:id", handler.UpdatePayment)
		payments.DELETE("/:id", handler.DeletePayment)
		payments.POST("/:id/complete", handler.CompletePayment)
		payments.POST("/:id/cancel", handler.CancelPayment)
		payments.GET("/summary", handler.GetPaymentSummary)
	}
}
