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
	"github.com/partflow/smart-store/internal/warranties"
	"github.com/partflow/smart-store/internal/inspections"
	"github.com/partflow/smart-store/internal/reports"
	"github.com/partflow/smart-store/internal/notifications"
	"github.com/partflow/smart-store/internal/dashboard"
	"github.com/partflow/smart-store/internal/permissions"
	"github.com/partflow/smart-store/internal/barcodes"
	"github.com/partflow/smart-store/internal/search"
	"github.com/partflow/smart-store/internal/organizations"
	"github.com/partflow/smart-store/internal/users"
	"github.com/partflow/smart-store/internal/roles"
	"github.com/partflow/smart-store/internal/audit"
	"github.com/partflow/smart-store/pkg/middleware"
)

// SetupRoutes يقوم بإعداد جميع المسارات بشكل مركزي
// مستوحى من نمط Fynexa المعماري لكنه متكيف مع احتياجات PartFlow
func SetupRoutes(router *gin.Engine, db *sqlx.DB, authService *auth.Service, permissionService *permissions.Service) {
	// Initialize all repositories
	productRepo := products.NewRepository(db)
	customerRepo := customers.NewRepository(db)
	salesRepo := sales.NewRepository(db)
	paymentRepo := payments.NewRepository(db)
	supplierRepo := suppliers.NewRepository(db)
	purchaseRepo := purchases.NewRepository(db)
	expenseRepo := expenses.NewRepository(db)
	returnRepo := returns.NewRepository(db)
	warrantyRepo := warranties.NewRepository(db)
	inspectionRepo := inspections.NewRepository(db)
	reportRepo := reports.NewRepository(db)
	notificationRepo := notifications.NewRepository(db)

	// Initialize all services
	inventoryService := inventory.NewService(inventory.NewRepository(db), db)
	productService := products.NewService(productRepo)
	customerService := customers.NewService(customerRepo)
	salesService := sales.NewService(salesRepo, db)
	paymentService := payments.NewService(paymentRepo)
	supplierService := suppliers.NewService(supplierRepo)
	purchaseService := purchases.NewService(purchaseRepo)
	expenseService := expenses.NewService(expenseRepo)
	returnService := returns.NewService(returnRepo)
	warrantyService := warranties.NewService(warrantyRepo)
	inspectionService := inspections.NewService(inspectionRepo)
	reportService := reports.NewService(reportRepo)
	notificationService := notifications.NewService(notificationRepo)
	dashboardService := dashboard.NewService(db)

	// Initialize all handlers
	authHandler := auth.NewHandler(authService, db)
	inventoryHandler := inventory.NewHandler(inventoryService)
	productHandler := products.NewHandler(productService)
	customerHandler := customers.NewHandler(customerService)
	salesHandler := sales.NewHandler(salesService)
	paymentHandler := payments.NewHandler(paymentService)
	supplierHandler := suppliers.NewHandler(supplierService)
	purchaseHandler := purchases.NewHandler(purchaseService)
	expenseHandler := expenses.NewHandler(expenseService)
	returnHandler := returns.NewHandler(returnService)
	warrantyHandler := warranties.NewHandler(warrantyService)
	inspectionHandler := inspections.NewHandler(inspectionService)
	reportHandler := reports.NewHandler(reportService)
	notificationHandler := notifications.NewHandler(notificationService)
	dashboardHandler := dashboard.NewHandler(dashboardService)

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

			// Organization routes
			organizations.RegisterRoutes(protected, db)

			// Users management routes
			users.RegisterRoutes(protected, db)

			// Roles routes
			roles.RegisterRoutes(protected, db)

			// Products routes
			categories := protected.Group("/categories")
			{
				categories.POST("", middleware.RequirePermission("products", "create"), productHandler.CreateCategory)
				categories.GET("/:id", middleware.RequirePermission("products", "read"), productHandler.GetCategory)
				categories.GET("", middleware.RequirePermission("products", "read"), productHandler.ListCategories)
				categories.PUT("/:id", middleware.RequirePermission("products", "update"), productHandler.UpdateCategory)
				categories.DELETE("/:id", middleware.RequirePermission("products", "delete"), productHandler.DeleteCategory)
			}

			brands := protected.Group("/brands")
			{
				brands.POST("", middleware.RequirePermission("products", "create"), productHandler.CreateBrand)
				brands.GET("/:id", middleware.RequirePermission("products", "read"), productHandler.GetBrand)
				brands.GET("", middleware.RequirePermission("products", "read"), productHandler.ListBrands)
				brands.PUT("/:id", middleware.RequirePermission("products", "update"), productHandler.UpdateBrand)
				brands.DELETE("/:id", middleware.RequirePermission("products", "delete"), productHandler.DeleteBrand)
			}

			products := protected.Group("/products")
			{
				products.POST("", middleware.RequirePermission("products", "create"), productHandler.CreateProduct)
				products.GET("/:id", middleware.RequirePermission("products", "read"), productHandler.GetProduct)
				products.GET("/barcode/:barcode", middleware.RequirePermission("products", "read"), productHandler.GetProductByBarcode)
				products.GET("", middleware.RequirePermission("products", "read"), productHandler.ListProducts)
				products.PUT("/:id", middleware.RequirePermission("products", "update"), productHandler.UpdateProduct)
				products.DELETE("/:id", middleware.RequirePermission("products", "delete"), productHandler.DeleteProduct)
				products.POST("/:id/archive", middleware.RequirePermission("products", "archive"), productHandler.ArchiveProduct)
				products.POST("/:id/barcode", middleware.RequirePermission("products", "update"), productHandler.GenerateBarcode)
				products.GET("/:id/stock", middleware.RequirePermission("products", "read"), productHandler.GetProductStock)
				products.GET("/search", middleware.RequirePermission("products", "read"), productHandler.SearchProducts)
			}

			// Customers routes
			customers := protected.Group("/customers")
			{
				customers.POST("", middleware.RequirePermission("customers", "create"), customerHandler.CreateCustomer)
				customers.GET("/:id", middleware.RequirePermission("customers", "read"), customerHandler.GetCustomer)
				customers.GET("", middleware.RequirePermission("customers", "read"), customerHandler.ListCustomers)
				customers.PUT("/:id", middleware.RequirePermission("customers", "update"), customerHandler.UpdateCustomer)
				customers.DELETE("/:id", middleware.RequirePermission("customers", "delete"), customerHandler.DeleteCustomer)
				customers.GET("/:id/ledger", middleware.RequirePermission("customers", "read"), customerHandler.GetCustomerLedger)
				customers.POST("/:id/payments", middleware.RequirePermission("debts", "manage"), customerHandler.AddPayment)
				customers.GET("/:id/debt-summary", middleware.RequirePermission("debts", "read"), customerHandler.GetCustomerDebtSummary)
				customers.PUT("/:id/credit-limit", middleware.RequirePermission("debts", "manage"), customerHandler.UpdateCreditLimit)
				customers.GET("/overdue", middleware.RequirePermission("debts", "read"), customerHandler.GetOverdueCustomers)
			}

			// Sales routes
			sales := protected.Group("/sales")
			{
				sales.POST("", middleware.RequirePermission("sales", "create"), salesHandler.CreateSale)
				sales.GET("/:id", middleware.RequirePermission("sales", "read"), salesHandler.GetSale)
				sales.GET("", middleware.RequirePermission("sales", "read"), salesHandler.ListSales)
				sales.POST("/:id/payment", middleware.RequirePermission("sales", "create"), salesHandler.UpdateSalePayment)
				sales.POST("/:id/cancel", middleware.RequirePermission("sales", "cancel"), salesHandler.CancelSale)
				sales.GET("/summary", middleware.RequirePermission("sales", "read"), salesHandler.GetSalesSummary)
				sales.GET("/top-products", middleware.RequirePermission("sales", "read"), salesHandler.GetTopSellingProducts)
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
				suppliers.POST("", middleware.RequirePermission("suppliers", "create"), supplierHandler.CreateSupplier)
				suppliers.GET("/:id", middleware.RequirePermission("suppliers", "read"), supplierHandler.GetSupplier)
				suppliers.GET("", middleware.RequirePermission("suppliers", "read"), supplierHandler.ListSuppliers)
				suppliers.PUT("/:id", middleware.RequirePermission("suppliers", "update"), supplierHandler.UpdateSupplier)
				suppliers.DELETE("/:id", middleware.RequirePermission("suppliers", "delete"), supplierHandler.DeleteSupplier)
				suppliers.GET("/:id/ledger", middleware.RequirePermission("suppliers", "read"), supplierHandler.GetSupplierLedger)
				suppliers.POST("/:id/payments", middleware.RequirePermission("debts", "manage"), supplierHandler.AddPayment)
				suppliers.GET("/:id/debt-summary", middleware.RequirePermission("debts", "read"), supplierHandler.GetSupplierDebtSummary)
				suppliers.PUT("/:id/credit-limit", middleware.RequirePermission("debts", "manage"), supplierHandler.UpdateCreditLimit)
				suppliers.GET("/overdue", middleware.RequirePermission("debts", "read"), supplierHandler.GetOverdueSuppliers)
			}

			// Purchases routes
			purchases := protected.Group("/purchases")
			{
				purchases.POST("", middleware.RequirePermission("purchases", "create"), purchaseHandler.CreatePurchase)
				purchases.GET("/:id", middleware.RequirePermission("purchases", "read"), purchaseHandler.GetPurchase)
				purchases.GET("", middleware.RequirePermission("purchases", "read"), purchaseHandler.ListPurchases)
				purchases.PUT("/:id", middleware.RequirePermission("purchases", "update"), purchaseHandler.UpdatePurchase)
				purchases.DELETE("/:id", middleware.RequirePermission("purchases", "update"), purchaseHandler.DeletePurchase)
				purchases.POST("/:id/receive", middleware.RequirePermission("purchases", "receive"), purchaseHandler.ReceivePurchase)
				purchases.POST("/:id/cancel", middleware.RequirePermission("purchases", "update"), purchaseHandler.CancelPurchase)
				purchases.POST("/:id/payment", middleware.RequirePermission("purchases", "update"), purchaseHandler.AddPayment)
				purchases.POST("/:id/items", middleware.RequirePermission("purchases", "create"), purchaseHandler.AddPurchaseItem)
				purchases.PUT("/items/:item_id", middleware.RequirePermission("purchases", "update"), purchaseHandler.UpdatePurchaseItem)
				purchases.DELETE("/items/:item_id", middleware.RequirePermission("purchases", "update"), purchaseHandler.DeletePurchaseItem)
			}

			// Expenses routes
			expenses := protected.Group("/expenses")
			{
				expenses.POST("", middleware.RequirePermission("expenses", "create"), expenseHandler.CreateExpense)
				expenses.GET("/:id", middleware.RequirePermission("expenses", "read"), expenseHandler.GetExpense)
				expenses.GET("", middleware.RequirePermission("expenses", "read"), expenseHandler.ListExpenses)
				expenses.PUT("/:id", middleware.RequirePermission("expenses", "update"), expenseHandler.UpdateExpense)
				expenses.DELETE("/:id", middleware.RequirePermission("expenses", "delete"), expenseHandler.DeleteExpense)
				expenses.POST("/:id/approve", middleware.RequirePermission("expenses", "update"), expenseHandler.ApproveExpense)
				expenses.POST("/:id/reject", middleware.RequirePermission("expenses", "update"), expenseHandler.RejectExpense)
				expenses.GET("/summary", middleware.RequirePermission("expenses", "read"), expenseHandler.GetExpenseSummary)
				expenses.GET("/categories", middleware.RequirePermission("expenses", "read"), expenseHandler.ListExpenseCategories)
			}

			// Returns routes
			returns := protected.Group("/returns")
			{
				returns.POST("", middleware.RequirePermission("returns", "create"), returnHandler.CreateReturn)
				returns.GET("/:id", middleware.RequirePermission("returns", "read"), returnHandler.GetReturn)
				returns.GET("", middleware.RequirePermission("returns", "read"), returnHandler.ListReturns)
				returns.PUT("/:id", middleware.RequirePermission("returns", "update"), returnHandler.UpdateReturn)
				returns.DELETE("/:id", middleware.RequirePermission("returns", "update"), returnHandler.DeleteReturn)
				returns.POST("/:id/approve", middleware.RequirePermission("returns", "approve"), returnHandler.ApproveReturn)
				returns.POST("/:id/reject", middleware.RequirePermission("returns", "approve"), returnHandler.RejectReturn)
				returns.POST("/:id/refund", middleware.RequirePermission("sales", "refund"), returnHandler.ProcessRefund)
				returns.POST("/:id/items", middleware.RequirePermission("returns", "create"), returnHandler.AddReturnItem)
				returns.PUT("/items/:item_id", middleware.RequirePermission("returns", "update"), returnHandler.UpdateReturnItem)
				returns.DELETE("/items/:item_id", middleware.RequirePermission("returns", "update"), returnHandler.DeleteReturnItem)
			}

			// Warranties routes
			warranties := protected.Group("/warranties")
			{
				warranties.POST("", middleware.RequirePermission("warranties", "read"), warrantyHandler.CreateWarranty)
				warranties.GET("/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.GetWarranty)
				warranties.GET("", middleware.RequirePermission("warranties", "read"), warrantyHandler.ListWarranties)
				warranties.PUT("/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.UpdateWarranty)
				warranties.DELETE("/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.DeleteWarranty)
				warranties.GET("/expiring-soon", middleware.RequirePermission("warranties", "read"), warrantyHandler.GetWarrantiesExpiringSoon)
				warranties.POST("/:id/claims", middleware.RequirePermission("warranties", "claim"), warrantyHandler.CreateWarrantyClaim)
				warranties.GET("/claims/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.GetWarrantyClaim)
				warranties.GET("/claims", middleware.RequirePermission("warranties", "read"), warrantyHandler.ListWarrantyClaims)
				warranties.PUT("/claims/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.UpdateWarrantyClaim)
				warranties.DELETE("/claims/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.DeleteWarrantyClaim)
				warranties.POST("/claims/:id/approve", middleware.RequirePermission("warranties", "read"), warrantyHandler.ApproveWarrantyClaim)
				warranties.POST("/claims/:id/reject", middleware.RequirePermission("warranties", "read"), warrantyHandler.RejectWarrantyClaim)
				warranties.POST("/claims/:id/complete", middleware.RequirePermission("warranties", "read"), warrantyHandler.CompleteWarrantyClaim)
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
				reports.POST("", middleware.RequirePermission("reports", "read"), reportHandler.GenerateReport)
				reports.GET("/:id", middleware.RequirePermission("reports", "read"), reportHandler.GetReport)
				reports.GET("", middleware.RequirePermission("reports", "read"), reportHandler.ListReports)
				reports.DELETE("/:id", middleware.RequirePermission("reports", "read"), reportHandler.DeleteReport)
			}

			// Register specific report routes
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
			}

			// Barcodes routes
			barcodes.RegisterRoutes(protected, db)

			// Search routes
			search.RegisterRoutes(protected, db)

			// Audit routes
			audit.RegisterRoutes(protected, db)
		}
	}
}
