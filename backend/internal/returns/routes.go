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
		returns.POST("/:id/items", handler.AddReturnItem)
		returns.PUT("/:id/items/:item_id", handler.UpdateReturnItem)
		returns.DELETE("/:id/items/:item_id", handler.DeleteReturnItem)
	}
}
