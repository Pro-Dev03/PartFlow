package expenses

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers expenses routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Expense routes
	expenses := router.Group("/expenses")
	{
		expenses.POST("", handler.CreateExpense)
		expenses.GET("/:id", handler.GetExpense)
		expenses.GET("", handler.ListExpenses)
		expenses.PUT("/:id", handler.UpdateExpense)
		expenses.DELETE("/:id", handler.DeleteExpense)
		expenses.POST("/:id/approve", handler.ApproveExpense)
		expenses.POST("/:id/reject", handler.RejectExpense)
		expenses.GET("/summary", handler.GetExpenseSummary)
	}

	// Expense category routes
	categories := router.Group("/expense-categories")
	{
		categories.POST("", handler.CreateExpenseCategory)
		categories.GET("/:id", handler.GetExpenseCategory)
		categories.GET("", handler.ListExpenseCategories)
		categories.PUT("/:id", handler.UpdateExpenseCategory)
		categories.DELETE("/:id", handler.DeleteExpenseCategory)
	}
}
