package returns

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers returns routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Return routes
	returns := router.Group("/returns")
	{
		returns.POST("", handler.CreateReturn)
		returns.GET("/:id", handler.GetReturn)
		returns.GET("", handler.ListReturns)
		returns.PUT("/:id", handler.UpdateReturn)
		returns.DELETE("/:id", handler.DeleteReturn)
		returns.POST("/:id/approve", handler.ApproveReturn)
		returns.POST("/:id/reject", handler.RejectReturn)
		returns.POST("/:id/refund", handler.ProcessRefund)
		returns.POST("/:id/complete", handler.CompleteReturn)
		returns.GET("/sale/:sale_id", handler.GetReturnsBySale)
		returns.GET("/customer/:customer_id", handler.GetReturnsByCustomer)
		returns.GET("/pending", handler.GetPendingReturns)
		returns.GET("/statistics", handler.GetReturnStatistics)
		returns.GET("/:id/with-items", handler.GetReturnWithItems)
		returns.GET("/analysis/monthly", handler.GetMonthlyReturnsAnalysis)
		returns.GET("/analysis/sales-returns", handler.GetSalesReturnsAnalysis)
		returns.POST("/:id/items", handler.AddReturnItem)
		returns.PUT("/:id/items/:item_id", handler.UpdateReturnItem)
		returns.DELETE("/:id/items/:item_id", handler.DeleteReturnItem)
		returns.POST("/items/:item_id/inspection", handler.ProcessReturnItemInspection)
		returns.GET("/validate/:sale_item_id", handler.ValidateReturnQuantity)
		returns.GET("/summary", handler.GetReturnSummary)
		returns.POST("/:id/reverse", handler.ReverseReturn)
	}
}
