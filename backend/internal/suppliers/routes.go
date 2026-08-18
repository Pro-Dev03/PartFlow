package suppliers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers suppliers routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Supplier routes
	suppliers := router.Group("/suppliers")
	{
		suppliers.POST("", handler.CreateSupplier)
		suppliers.GET("/:id", handler.GetSupplier)
		suppliers.GET("", handler.ListSuppliers)
		suppliers.PUT("/:id", handler.UpdateSupplier)
		suppliers.DELETE("/:id", handler.DeleteSupplier)
		suppliers.GET("/:id/ledger", handler.GetSupplierLedger)
		suppliers.POST("/:id/payments", handler.AddPayment)
		suppliers.GET("/:id/debt-summary", handler.GetSupplierDebtSummary)
		suppliers.PUT("/:id/credit-limit", handler.UpdateCreditLimit)
		suppliers.GET("/overdue", handler.GetOverdueSuppliers)
		
		// Debt management routes
		suppliers.POST("/:id/debts", handler.CreateDebtEntry)
		suppliers.GET("/:id/debts", handler.GetDebtEntries)
		suppliers.POST("/:id/debt-collections", handler.CreateDebtCollection)
		suppliers.GET("/:id/debt-collections", handler.GetDebtCollections)
		suppliers.POST("/:id/debt-payments", handler.ProcessDebtPayment)
	}

	// Organization-level debt collection routes
	router.GET("/supplier-debt-collections/pending", handler.GetPendingDebtCollections)
}
