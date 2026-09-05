package reports

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers reports routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Reports routes
	reports := router.Group("/reports")
	{
		reports.GET("/sales", handler.GenerateSalesReport)
		reports.GET("/net-sales", handler.GenerateNetSalesReport)
		reports.GET("/tax", handler.GenerateTaxReport)
		reports.GET("/purchases", handler.GeneratePurchasesReport)
		reports.GET("/inventory", handler.GenerateInventoryReport)
		reports.GET("/expenses", handler.GenerateExpensesReport)
		reports.GET("/profit", handler.GenerateProfitsReport)
		reports.GET("/debts", handler.GenerateDebtsReport)
		reports.GET("/returns", handler.GenerateReturnsReport)
		reports.GET("/products", handler.GenerateProductsReport)
		reports.GET("/suppliers", handler.GenerateSuppliersReport)
	}
}
