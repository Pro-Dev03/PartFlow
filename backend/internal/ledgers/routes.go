package ledgers

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	// Inventory ledger routes
	r.GET("/inventory/:id/ledger", handler.GetInventoryLedger)
	r.GET("/inventory/:id/ledger/summary", handler.GetInventoryLedgerSummary)

	// Manual ledger entry (admin only)
	r.POST("/ledger", handler.CreateLedgerEntry)
}
