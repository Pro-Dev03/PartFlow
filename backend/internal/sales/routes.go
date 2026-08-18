package sales

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers sales routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Sale routes
	sales := router.Group("/sales")
	{
		sales.POST("", handler.CreateSale)
		sales.GET("/:id", handler.GetSale)
		sales.GET("", handler.ListSales)
		sales.POST("/:id/payment", handler.UpdateSalePayment)
		sales.POST("/:id/cancel", handler.CancelSale)
		sales.GET("/summary", handler.GetSalesSummary)
		sales.GET("/top-products", handler.GetTopSellingProducts)
	}

	// Transaction routes
	transactions := router.Group("/transactions")
	{
		transactions.POST("", handler.CreateTransaction)
		transactions.GET("/:id", handler.GetTransaction)
		transactions.GET("", handler.ListTransactions)
		transactions.GET("/accounts/:account/balance", handler.GetAccountBalance)
	}

	// Profit routes
	profit := router.Group("/profit")
	{
		profit.GET("/calculate", handler.CalculateProfitForPeriod)
		profit.GET("/entries", handler.GetProfitEntries)
	}
}
