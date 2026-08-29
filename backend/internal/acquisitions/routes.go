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
		acquisitions.GET("/aging", handler.GetUsedPartsAging)
		acquisitions.GET("/seller-balances", handler.GetSellerBalances)
		acquisitions.GET("/:id", handler.GetAcquisition)
		acquisitions.PUT("/:id/status", handler.UpdateAcquisitionStatus)
		acquisitions.POST("/:id/payments", handler.CreateSellerPayment)
	}

	// Item-related routes
	items := r.Group("/acquisitions/items")
	{
		items.POST("/:id/repair-cost", handler.AddRepairCost)
		// Acquisition item history is distinct from inventory movement history.
		items.GET("/:id/history", handler.GetItemHistory)
	}

	// Keep the original route as a compatibility alias for existing clients.
	legacyInventory := r.Group("/inventory")
	{
		legacyInventory.GET("/:id/history", handler.GetItemHistory)
	}
}
