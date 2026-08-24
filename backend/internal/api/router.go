package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/partflow/smart-store/internal/auth"
	"github.com/partflow/smart-store/internal/inventory"
	"github.com/partflow/smart-store/internal/products"
	"github.com/partflow/smart-store/internal/customers"
	"github.com/partflow/smart-store/internal/sales"
	"github.com/partflow/smart-store/internal/payments"
	"github.com/partflow/smart-store/internal/suppliers"
	"github.com/partflow/smart-store/internal/purchases"
	"github.com/partflow/smart-store/internal/expenses"
	"github.com/partflow/smart-store/internal/returns"
	"github.com/partflow/smart-store/internal/inspections"
	"github.com/partflow/smart-store/internal/reports"
	"github.com/partflow/smart-store/internal/notifications"
	"github.com/partflow/smart-store/internal/dashboard"
	"github.com/partflow/smart-store/internal/search"
	"github.com/partflow/smart-store/internal/users"
	"github.com/partflow/smart-store/internal/audit"
	"github.com/partflow/smart-store/internal/parttypes"
	"github.com/partflow/smart-store/internal/settings"
	"github.com/partflow/smart-store/pkg/middleware"
)

// SetupRoutes يقوم بإعداد جميع المسارات بشكل مركزي
// مستوحى من نمط Fynexa المعماري لكنه متكيف مع احتياجات PartFlow
func SetupRoutes(router *gin.Engine, db *sqlx.DB, authService *auth.Service) {
	// Initialize all repositories
	productRepo := products.NewRepository(db)
	customerRepo := customers.NewRepository(db)
	salesRepo := sales.NewRepository(db)
	paymentRepo := payments.NewRepository(db)
	supplierRepo := suppliers.NewRepository(db)
	purchaseRepo := purchases.NewRepository(db)
	expenseRepo := expenses.NewRepository(db)
	returnRepo := returns.NewRepository(db)
	inspectionRepo := inspections.NewRepository(db)
	reportRepo := reports.NewRepository(db)
	notificationRepo := notifications.NewRepository(db)
	partTypesRepo := parttypes.NewRepository(db)

	// Initialize all services
	inventoryService := inventory.NewService(inventory.NewRepository(db), db)
	productService := products.NewService(productRepo)
	customerService := customers.NewService(customerRepo)
	salesService := sales.NewService(salesRepo, db)
	paymentService := payments.NewService(paymentRepo)
	supplierService := suppliers.NewService(supplierRepo, db)
	purchaseService := purchases.NewService(purchaseRepo, db)
	expenseService := expenses.NewService(expenseRepo)
	returnService := returns.NewService(returnRepo)
	inspectionService := inspections.NewService(inspectionRepo)
	reportService := reports.NewService(reportRepo)
	notificationService := notifications.NewService(notificationRepo)
	dashboardService := dashboard.NewService(db)
	partTypesService := parttypes.NewService(partTypesRepo)

	// Initialize all handlers
	authHandler := auth.NewHandler(authService, db)
	inventoryHandler := inventory.NewHandler(inventoryService, db)
	productHandler := products.NewHandler(productService)
	customerHandler := customers.NewHandler(customerService)
	salesHandler := sales.NewHandler(salesService)
	paymentHandler := payments.NewHandler(paymentService)
	supplierHandler := suppliers.NewHandler(supplierService, db)
	purchaseHandler := purchases.NewHandler(purchaseService)
	expenseHandler := expenses.NewHandler(expenseRepo)
	returnHandler := returns.NewHandler(returnService)
	inspectionHandler := inspections.NewHandler(inspectionService)
	reportHandler := reports.NewHandler(reportRepo)
	notificationHandler := notifications.NewHandler(notificationRepo)
	dashboardHandler := dashboard.NewHandler(db)
	partTypesHandler := parttypes.NewHandler(partTypesService)
	settingsHandler := settings.NewHandler(db.DB)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		public := v1.Group("")
		{
			public.POST("/auth/register", authHandler.Register)
			public.POST("/auth/login", authHandler.Login)
			public.POST("/auth/refresh", authHandler.RefreshToken)
		}

		// Protected routes (auth required)
		protected := v1.Group("")
		protected.Use(middleware.Auth())
		{
			// Dashboard routes
			protected.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)

			// Auth routes
			auth := protected.Group("/auth")
			{
				auth.POST("/logout", authHandler.Logout)
			}

			// User routes (current user info)
			protected.GET("/users/me", authHandler.GetCurrentUser)

			// Inventory routes
			inventoryHandler.RegisterRoutes(protected)

			// Part Types routes
			partTypesHandler.RegisterRoutes(protected)

			// Users management routes
			users.RegisterRoutes(protected, db)

			// Products routes
			categories := protected.Group("/categories")
			{
				categories.POST("", productHandler.CreateCategory)
				categories.GET("/:id", productHandler.GetCategory)
				categories.GET("", productHandler.ListCategories)
				categories.PUT("/:id", productHandler.UpdateCategory)
				categories.DELETE("/:id", productHandler.DeleteCategory)
			}

			brands := protected.Group("/brands")
			{
				brands.POST("", productHandler.CreateBrand)
				brands.GET("/:id", productHandler.GetBrand)
				brands.GET("", productHandler.ListBrands)
				brands.PUT("/:id", productHandler.UpdateBrand)
				brands.DELETE("/:id", productHandler.DeleteBrand)
			}

			products := protected.Group("/products")
			{
				products.POST("", productHandler.CreateProduct)
				products.GET("/:id", productHandler.GetProduct)
				products.GET("/barcode/:barcode", productHandler.GetProductByBarcode)
				products.GET("", productHandler.ListProducts)
				products.PUT("/:id", productHandler.UpdateProduct)
				products.DELETE("/:id", productHandler.DeleteProduct)
				products.POST("/:id/archive", productHandler.ArchiveProduct)
				products.POST("/:id/barcode", productHandler.GenerateBarcode)
				products.GET("/:id/stock", productHandler.GetProductStock)
				products.GET("/search", productHandler.SearchProducts)
			}

			// Customers routes
			customers := protected.Group("/customers")
			{
				customers.POST("", customerHandler.CreateCustomer)
				customers.GET("/:id", customerHandler.GetCustomer)
				customers.GET("", customerHandler.ListCustomers)
				customers.PUT("/:id", customerHandler.UpdateCustomer)
				customers.DELETE("/:id", customerHandler.DeleteCustomer)
				customers.GET("/:id/ledger", customerHandler.GetCustomerLedger)
				customers.POST("/:id/payments", customerHandler.AddPayment)
				customers.GET("/:id/debt-summary", customerHandler.GetCustomerDebtSummary)
				customers.PUT("/:id/credit-limit", customerHandler.UpdateCreditLimit)
				customers.GET("/overdue", customerHandler.GetOverdueCustomers)
			}

			// Sales routes
			sales := protected.Group("/sales")
			{
				sales.POST("", salesHandler.CreateSale)
				sales.GET("/:id", salesHandler.GetSale)
				sales.GET("", salesHandler.ListSales)
				sales.POST("/:id/payment", salesHandler.UpdateSalePayment)
				sales.POST("/:id/cancel", salesHandler.CancelSale)
				sales.GET("/summary", salesHandler.GetSalesSummary)
				sales.GET("/top-products", salesHandler.GetTopSellingProducts)
			}

			// Payments routes
			payments := protected.Group("/payments")
			{
				payments.POST("", paymentHandler.CreatePayment)
				payments.GET("/:id", paymentHandler.GetPayment)
				payments.GET("", paymentHandler.ListPayments)
				payments.PUT("/:id", paymentHandler.UpdatePayment)
				payments.DELETE("/:id", paymentHandler.DeletePayment)
				payments.POST("/:id/complete", paymentHandler.CompletePayment)
				payments.POST("/:id/cancel", paymentHandler.CancelPayment)
				payments.GET("/summary", paymentHandler.GetPaymentSummary)
			}

			// Suppliers routes
			suppliers := protected.Group("/suppliers")
			{
				suppliers.POST("", supplierHandler.CreateSupplier)
				suppliers.GET("/:id", supplierHandler.GetSupplier)
				suppliers.GET("", supplierHandler.ListSuppliers)
				suppliers.PUT("/:id", supplierHandler.UpdateSupplier)
				suppliers.DELETE("/:id", supplierHandler.DeleteSupplier)
				suppliers.GET("/:id/ledger", supplierHandler.GetSupplierLedger)
				suppliers.POST("/:id/payments", supplierHandler.AddPayment)
				suppliers.GET("/:id/debt-summary", supplierHandler.GetSupplierDebtSummary)
				suppliers.PUT("/:id/credit-limit", supplierHandler.UpdateCreditLimit)
				suppliers.GET("/overdue", supplierHandler.GetOverdueSuppliers)
				suppliers.GET("/:id/inventory", supplierHandler.GetSupplierInventory)
			}

			// Purchases routes
			purchases := protected.Group("/purchases")
			{
				purchases.POST("", purchaseHandler.CreatePurchase)
				purchases.GET("/:id", purchaseHandler.GetPurchase)
				purchases.GET("", purchaseHandler.ListPurchases)
				purchases.PUT("/:id", purchaseHandler.UpdatePurchase)
				purchases.DELETE("/:id", purchaseHandler.DeletePurchase)
				purchases.POST("/:id/receive", purchaseHandler.ReceivePurchase)
				purchases.POST("/:id/cancel", purchaseHandler.CancelPurchase)
				purchases.POST("/:id/payment", purchaseHandler.AddPayment)
				purchases.POST("/:id/items", purchaseHandler.AddPurchaseItem)
				purchases.PUT("/items/:item_id", purchaseHandler.UpdatePurchaseItem)
				purchases.DELETE("/items/:item_id", purchaseHandler.DeletePurchaseItem)
			}

			// Expenses routes
			expenses := protected.Group("/expenses")
			{
				expenses.POST("", expenseHandler.CreateExpense)
				expenses.GET("/:id", expenseHandler.GetExpense)
				expenses.GET("", expenseHandler.ListExpenses)
				expenses.PUT("/:id", expenseHandler.UpdateExpense)
				expenses.DELETE("/:id", expenseHandler.DeleteExpense)
				expenses.POST("/:id/approve", expenseHandler.ApproveExpense)
				expenses.POST("/:id/reject", expenseHandler.RejectExpense)
				expenses.GET("/summary", expenseHandler.GetExpenseSummary)
				expenses.GET("/categories", expenseHandler.ListExpenseCategories)
			}

			// Returns routes
			returns := protected.Group("/returns")
			{
				returns.POST("", returnHandler.CreateReturn)
				returns.GET("/:id", returnHandler.GetReturn)
				returns.GET("", returnHandler.ListReturns)
				returns.PUT("/:id", returnHandler.UpdateReturn)
				returns.DELETE("/:id", returnHandler.DeleteReturn)
				returns.POST("/:id/approve", returnHandler.ApproveReturn)
				returns.POST("/:id/reject", returnHandler.RejectReturn)
				returns.POST("/:id/refund", returnHandler.ProcessRefund)
				returns.POST("/:id/items", returnHandler.AddReturnItem)
				returns.PUT("/items/:item_id", returnHandler.UpdateReturnItem)
				returns.DELETE("/items/:item_id", returnHandler.DeleteReturnItem)
			}

			// Inspections routes
			inspections := protected.Group("/inspections")
			{
				inspections.POST("", inspectionHandler.CreateInspection)
				inspections.GET("/:id", inspectionHandler.GetInspection)
				inspections.GET("", inspectionHandler.ListInspections)
				inspections.PUT("/:id", inspectionHandler.UpdateInspection)
				inspections.DELETE("/:id", inspectionHandler.DeleteInspection)
			}

			// Reports routes
			reports := protected.Group("/reports")
			{
				reports.POST("", reportHandler.GenerateReport)
				reports.GET("/:id", reportHandler.GetReport)
				reports.GET("", reportHandler.ListReports)
				reports.DELETE("/:id", reportHandler.DeleteReport)
			}

			// Register specific report routes
			// reports.RegisterRoutes(protected, db)

			// Notifications routes
			notifications := protected.Group("/notifications")
			{
				notifications.POST("", notificationHandler.CreateNotification)
				notifications.GET("/:id", notificationHandler.GetNotification)
				notifications.GET("", notificationHandler.ListNotifications)
				notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
				notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
				notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			}

			// Barcodes routes
			// barcodes.RegisterRoutes(protected, db)

			// Search routes
			search.RegisterRoutes(protected, db)

			// Audit routes
			audit.RegisterRoutes(protected, db)

			// Settings routes
			settings := protected.Group("/settings")
			{
				settings.GET("/public", settingsHandler.GetPublicSettings)
				settings.GET("/:key", settingsHandler.GetSetting)
				settings.PUT("/:key", settingsHandler.UpdateSetting)
				settings.GET("/tax-rate", settingsHandler.GetTaxRate)
				settings.PUT("/tax-rate", settingsHandler.UpdateTaxRate)
			}
		}
	}
}
