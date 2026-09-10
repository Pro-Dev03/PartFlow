package acquisitions

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers acquisition routes
func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	acquisitions := r.Group("/acquisitions")
	{
		acquisitions.POST("", handler.CreateAcquisition)
		acquisitions.GET("", handler.ListAcquisitions)
		acquisitions.GET("/seller-balances", handler.GetSellerBalances)
		acquisitions.POST("/seller-balances/:customer_id/payments", handler.CreateSellerBalancePayment)
		acquisitions.GET("/:id", handler.GetAcquisition)
		acquisitions.PUT("/:id/status", handler.UpdateAcquisitionStatus)
		acquisitions.POST("/:id/payments", handler.CreateSellerPayment)
	}

}
