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
		acquisitions.GET("/:id", handler.GetAcquisition)
		acquisitions.PUT("/:id/status", handler.UpdateAcquisitionStatus)
		acquisitions.POST("/:id/payments", handler.CreateSellerPayment)
		acquisitions.GET("/aging", handler.GetUsedPartsAging)
		acquisitions.GET("/seller-balances", handler.GetSellerBalances)
	}

	// Item-related routes
	items := r.Group("/acquisitions/items")
	{
		items.POST("/:id/repair-cost", handler.AddRepairCost)
	}

	// Inventory history routes
	inventory := r.Group("/inventory")
	{
		inventory.GET("/:id/history", handler.GetItemHistory)
	}
}