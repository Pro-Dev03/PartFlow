package reports

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/middleware"
)

// RegisterRoutes registers reports routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Reports routes with permission middleware
	reports := router.Group("/reports")
	{
		reports.GET("/sales", middleware.RequirePermission("reports", "read"), handler.GenerateSalesReport)
		reports.GET("/purchases", middleware.RequirePermission("reports", "read"), handler.GeneratePurchasesReport)
		reports.GET("/inventory", middleware.RequirePermission("reports", "read"), handler.GenerateInventoryReport)
		reports.GET("/expenses", middleware.RequirePermission("reports", "read"), handler.GenerateExpensesReport)
		reports.GET("/profit", middleware.RequirePermission("reports", "read"), handler.GenerateProfitsReport)
		reports.GET("/debts", middleware.RequirePermission("reports", "read"), handler.GenerateDebtsReport)
		reports.GET("/returns", middleware.RequirePermission("reports", "read"), handler.GenerateReturnsReport)
		reports.GET("/warranty", middleware.RequirePermission("reports", "read"), handler.GenerateWarrantyReport)
	}
}
