package customers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers customers routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Customer routes
	customers := router.Group("/customers")
	{
		customers.POST("", handler.CreateCustomer)
		customers.GET("/:id", handler.GetCustomer)
		customers.GET("", handler.ListCustomers)
		customers.PUT("/:id", handler.UpdateCustomer)
		customers.DELETE("/:id", handler.DeleteCustomer)
		customers.GET("/:id/ledger", handler.GetCustomerLedger)
		customers.POST("/:id/payments", handler.AddPayment)
		customers.GET("/:id/debt-summary", handler.GetCustomerDebtSummary)
		customers.PUT("/:id/credit-limit", handler.UpdateCreditLimit)
		customers.GET("/overdue", handler.GetOverdueCustomers)
		
		// Debt management routes
		customers.POST("/:id/debts", handler.CreateDebtEntry)
		customers.GET("/:id/debts", handler.GetDebtEntries)
		customers.POST("/:id/debt-collections", handler.CreateDebtCollection)
		customers.GET("/:id/debt-collections", handler.GetDebtCollections)
		customers.POST("/:id/debt-payments", handler.ProcessDebtPayment)
	}

	// Organization-level debt collection routes
	router.GET("/debt-collections/pending", handler.GetPendingDebtCollections)
}
