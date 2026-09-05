package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/partflow/smart-store/internal/acquisitions"
	"github.com/partflow/smart-store/internal/audit"
	"github.com/partflow/smart-store/internal/auth"
	"github.com/partflow/smart-store/internal/barcodes"
	"github.com/partflow/smart-store/internal/customers"
	"github.com/partflow/smart-store/internal/dashboard"
	"github.com/partflow/smart-store/internal/debts"
	"github.com/partflow/smart-store/internal/expenses"
	"github.com/partflow/smart-store/internal/inspections"
	"github.com/partflow/smart-store/internal/inventory"
	"github.com/partflow/smart-store/internal/ledgers"
	"github.com/partflow/smart-store/internal/notifications"
	"github.com/partflow/smart-store/internal/parttypes"
	"github.com/partflow/smart-store/internal/payments"
	"github.com/partflow/smart-store/internal/products"
	"github.com/partflow/smart-store/internal/purchases"
	"github.com/partflow/smart-store/internal/reports"
	"github.com/partflow/smart-store/internal/returns"
	"github.com/partflow/smart-store/internal/sales"
	"github.com/partflow/smart-store/internal/search"
	"github.com/partflow/smart-store/internal/settings"
	"github.com/partflow/smart-store/internal/supplierreturns"
	"github.com/partflow/smart-store/internal/suppliers"
	"github.com/partflow/smart-store/internal/sync"
	"github.com/partflow/smart-store/internal/users"
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
	notificationService := notifications.NewService(notificationRepo)
	partTypesService := parttypes.NewService(partTypesRepo)
	ledgerService := ledgers.NewService(db)
	acquisitionService := acquisitions.NewService(db)

	// Initialize all handlers
	authHandler := auth.NewHandler(authService, db)
	inventoryHandler := inventory.NewHandler(inventoryService, db)
	inventoryMainHandler := inventory.NewMainHandler(db)
	productHandler := products.NewHandler(productService)
	customerHandler := customers.NewHandler(customerService)
	salesHandler := sales.NewHandler(salesService)
	paymentHandler := payments.NewHandler(paymentService)
	supplierHandler := suppliers.NewHandler(supplierService)
	purchaseHandler := purchases.NewHandler(purchaseService)
	expenseHandler := expenses.NewHandler(expenseService)
	returnHandler := returns.NewHandler(returnService)
	inspectionHandler := inspections.NewHandler(inspectionService)
	notificationHandler := notifications.NewHandler(notificationService)
	cachedDashboardService := dashboard.NewCachedService(db)
	dashboardHandler := dashboard.NewHandler(cachedDashboardService)
	partTypesHandler := parttypes.NewHandler(partTypesService)
	settingsHandler := settings.NewHandler(db.DB)
	databaseHandler := settings.NewDatabaseHandler(db)
	localDatabaseHandler := settings.NewLocalDatabaseHandler()
	ledgerHandler := ledgers.NewHandler(ledgerService)
	acquisitionHandler := acquisitions.NewHandler(acquisitionService)
	debtsHandler := debts.NewHandler(db)

	// Aggregation handler (ARCHITECTURE-PRINCIPLES.md)
	aggregationHandler := NewAggregationHandler(db)
	syncHandler := sync.NewHandler(db)

	// SmartDelete handler (PRODUCT-PHILOSOPHY.md)
	purchaseSmartDeleteHandler := NewPurchaseSmartDeleteHandler(db)

	// Barcode handler
	barcodeRepo := barcodes.NewRepository(db)
	barcodeService := barcodes.NewService(barcodeRepo)
	barcodeHandler := barcodes.NewHandler(barcodeService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		public := v1.Group("")
		{
			public.POST("/auth/register", authHandler.Register)
			public.POST("/auth/login", authHandler.Login)
			public.POST("/auth/refresh", authHandler.RefreshToken)
			public.POST("/auth/cloud-session", authHandler.CloudSession)
		}

		// Protected routes (auth required)
		protected := v1.Group("")
		protected.Use(middleware.Auth())
		{
			// Dashboard routes
			protected.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
			protected.GET("/dashboard/low-stock-items", dashboardHandler.GetLowStockItems)
			protected.GET("/dashboard/overdue-debts", dashboardHandler.GetOverdueDebts)

			// Aggregation routes (ARCHITECTURE-PRINCIPLES.md)
			aggregations := protected.Group("/aggregations")
			{
				aggregations.GET("/daily-sales", aggregationHandler.GetDailySalesSummary)
				aggregations.GET("/monthly-sales", aggregationHandler.GetMonthlySalesSummary)
				aggregations.GET("/daily-inventory", aggregationHandler.GetDailyInventorySummary)
				aggregations.GET("/monthly-inventory", aggregationHandler.GetMonthlyInventorySummary)
				aggregations.GET("/daily-debt", aggregationHandler.GetDailyDebtSummary)
				aggregations.GET("/monthly-debt", aggregationHandler.GetMonthlyDebtSummary)
				aggregations.GET("/daily-profit", aggregationHandler.GetDailyProfitSummary)
				aggregations.GET("/monthly-profit", aggregationHandler.GetMonthlyProfitSummary)
				aggregations.GET("/status", aggregationHandler.GetAggregationStatus)
				aggregations.POST("/update", aggregationHandler.UpdateAggregations)
			}

			// Auth routes
			auth := protected.Group("/auth")
			{
				auth.POST("/logout", authHandler.Logout)
				auth.POST("/validate", authHandler.ValidateSubscription)
				auth.GET("/admin-check", middleware.Admin(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"is_admin": true}})
				})
			}

			// User routes (current user info)
			protected.GET("/users/me", authHandler.GetCurrentUser)

			// Inventory routes
			inventoryHandler.RegisterRoutes(protected)
			inventoryMainHandler.RegisterMainRoutes(protected)

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
				products.PATCH("/:id/min-stock", productHandler.UpdateMinimumStock)
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
				customers.GET("/:id/financial-timeline", customerHandler.GetFinancialTimeline)
				customers.POST("/:id/payments", customerHandler.AddPayment)
				customers.POST("/:id/debt-payments", customerHandler.AddPayment)
				customers.GET("/:id/debt-summary", customerHandler.GetCustomerDebtSummary)
				customers.PUT("/:id/credit-limit", customerHandler.UpdateCreditLimit)
				customers.GET("/overdue", customerHandler.GetOverdueCustomers)
				customers.POST("/:id/receipt", customerHandler.GeneratePaymentReceipt)
			}

			// Sales routes
			sales := protected.Group("/sales")
			{
				sales.POST("", salesHandler.CreateSale)
				sales.GET("/held", salesHandler.ListHeldSales)
				sales.POST("/held", salesHandler.HoldSale)
				sales.DELETE("/held/:id", salesHandler.DeleteHeldSale)
				sales.GET("/:id", salesHandler.GetSale)
				sales.GET("", salesHandler.ListSales)
				sales.POST("/:id/payment", salesHandler.UpdateSalePayment)
				sales.POST("/:id/cancel", salesHandler.CancelSale)
				sales.GET("/summary", salesHandler.GetSalesSummary)
				sales.GET("/top-products", salesHandler.GetTopSellingProducts)
			}

			// Financial transaction routes
			transactions := protected.Group("/transactions")
			{
				transactions.POST("", salesHandler.CreateTransaction)
				transactions.GET("/:id", salesHandler.GetTransaction)
				transactions.GET("", salesHandler.ListTransactions)
				transactions.GET("/accounts/:account/balance", salesHandler.GetAccountBalance)
			}

			// Profit routes
			profit := protected.Group("/profit")
			{
				profit.GET("/calculate", salesHandler.CalculateProfitForPeriod)
				profit.GET("/entries", salesHandler.GetProfitEntries)
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
				// SmartDelete - now returns SmartDeleteResult (PRODUCT-PHILOSOPHY.md)
				purchases.DELETE("/:id", purchaseSmartDeleteHandler.SmartDelete)
				purchases.GET("/:id/used-items", purchaseSmartDeleteHandler.GetUsedItemsInfo)
				purchases.POST("/:id/receive", purchaseHandler.ReceivePurchase)
				purchases.POST("/:id/cancel", purchaseHandler.CancelPurchase)
				purchases.POST("/:id/reverse", purchaseHandler.ReversePurchase)
				purchases.POST("/:id/payment", purchaseHandler.AddPayment)
				purchases.POST("/:id/items", purchaseHandler.AddPurchaseItem)
				purchases.PUT("/items/:item_id", purchaseHandler.UpdatePurchaseItem)
				purchases.DELETE("/items/:item_id", purchaseHandler.DeletePurchaseItem)
			}

			// Expenses routes
			expenses := protected.Group("/expenses")
			{
				expenses.POST("", expenseHandler.CreateExpense)
				expenses.GET("/categories", expenseHandler.ListExpenseCategories)
				expenses.POST("/categories", expenseHandler.CreateExpenseCategory)
				expenses.GET("/:id", expenseHandler.GetExpense)
				expenses.GET("", expenseHandler.ListExpenses)
				expenses.PUT("/:id", expenseHandler.UpdateExpense)
				expenses.DELETE("/:id", expenseHandler.DeleteExpense)
				expenses.POST("/:id/approve", expenseHandler.ApproveExpense)
				expenses.POST("/:id/reject", expenseHandler.RejectExpense)
				expenses.GET("/summary", expenseHandler.GetExpenseSummary)
			}
			supplierreturns.RegisterRoutes(protected, db)

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
				returns.POST("/:id/complete", returnHandler.CompleteReturn)
				returns.GET("/sale/:sale_id", returnHandler.GetReturnsBySale)
				returns.GET("/customer/:customer_id", returnHandler.GetReturnsByCustomer)
				returns.GET("/pending", returnHandler.GetPendingReturns)
				returns.GET("/statistics", returnHandler.GetReturnStatistics)
				returns.GET("/:id/with-items", returnHandler.GetReturnWithItems)
				returns.GET("/analysis/monthly", returnHandler.GetMonthlyReturnsAnalysis)
				returns.GET("/analysis/sales-returns", returnHandler.GetSalesReturnsAnalysis)
				returns.POST("/:id/items", returnHandler.AddReturnItem)
				returns.PUT("/:id/items/:item_id", returnHandler.UpdateReturnItem)
				returns.DELETE("/:id/items/:item_id", returnHandler.DeleteReturnItem)
				returns.POST("/items/:item_id/inspection", returnHandler.ProcessReturnItemInspection)
				returns.GET("/validate/:sale_item_id", returnHandler.ValidateReturnQuantity)
				returns.GET("/summary", returnHandler.GetReturnSummary)
				returns.POST("/:id/reverse", returnHandler.ReverseReturn)
			}

			// Inspections routes
			inspections := protected.Group("/inspections")
			{
				inspections.POST("", inspectionHandler.CreateInspection)
				inspections.GET("/summary", inspectionHandler.GetInspectionSummary)
				inspections.GET("/:id", inspectionHandler.GetInspection)
				inspections.GET("", inspectionHandler.ListInspections)
				inspections.PUT("/:id", inspectionHandler.UpdateInspection)
				inspections.DELETE("/:id", inspectionHandler.DeleteInspection)
				inspections.POST("/:id/pass", inspectionHandler.PassInspection)
				inspections.POST("/:id/fail", inspectionHandler.FailInspection)
			}

			// Reports routes
			reports.RegisterRoutes(protected, db)

			// Notifications routes
			notifications := protected.Group("/notifications")
			{
				notifications.POST("", notificationHandler.CreateNotification)
				notifications.GET("/:id", notificationHandler.GetNotification)
				notifications.GET("", notificationHandler.ListNotifications)
				notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
				notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
				notifications.DELETE("/:id", notificationHandler.DeleteNotification)
				notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
				notifications.GET("/preferences", notificationHandler.GetNotificationPreferences)
				notifications.PUT("/preferences", notificationHandler.UpdateNotificationPreferences)
			}

			// Barcodes routes
			barcodeHandler.RegisterRoutes(protected)

			// Search routes
			search.RegisterRoutes(protected, db)

			// Audit routes
			audit.RegisterRoutes(protected, db)

			// Ledger routes
			ledgers.RegisterRoutes(protected, ledgerHandler)

			// Acquisitions routes (USED-PARTS-ACQUISITION.md)
			acquisitions.RegisterRoutes(protected, acquisitionHandler)

			// Debts routes
			debtsHandler.RegisterRoutes(protected)

			// Sync routes
			sync := protected.Group("/sync")
			{
				// The current cloud schema is single-tenant. Never expose a global
				// operational snapshot to an ordinary subscriber until tenant
				// columns/database-per-store isolation is deployed.
				sync.GET("/initial-data", middleware.Admin(), syncHandler.GetInitialData)
			}

			// Settings routes
			settings := protected.Group("/settings")
			{
				settings.POST("/sync", middleware.Admin(), localDatabaseHandler.SyncCloudData)
				settings.POST("/sync/push", middleware.Admin(), localDatabaseHandler.SyncLocalDataToCloud)
				settings.GET("/sync/conflicts", localDatabaseHandler.GetSyncConflicts)
				settings.DELETE("/sync/conflicts", localDatabaseHandler.ClearSyncConflicts)
				settings.POST("/sync/conflicts/:id/resolve", localDatabaseHandler.ResolveSyncConflict)
				settings.GET("/operating-mode", localDatabaseHandler.GetOperatingMode)
				settings.PUT("/operating-mode", localDatabaseHandler.SetOperatingMode)
				settings.GET("/public", settingsHandler.GetPublicSettings)
				settings.GET("/:key", settingsHandler.GetSetting)
				settings.PUT("/:key", settingsHandler.UpdateSetting)
				settings.GET("/tax-rate", settingsHandler.GetTaxRate)
				settings.PUT("/tax-rate", settingsHandler.UpdateTaxRate)
				// Database reset and runtime migrations are destructive/operational
				// actions. Keep ordinary subscribers out even when authenticated.
				adminSettings := settings.Group("")
				adminSettings.Use(middleware.Admin())
				adminSettings.DELETE("/database", databaseHandler.DeleteAllData)
				adminSettings.POST("/migrate", databaseHandler.ApplyMigration)
			}
		}
	}
}
