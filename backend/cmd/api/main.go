package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

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
	"github.com/partflow/smart-store/internal/audit"
	"github.com/partflow/smart-store/internal/dashboard"
	"github.com/partflow/smart-store/internal/permissions"
	"github.com/partflow/smart-store/internal/barcodes"
	"github.com/partflow/smart-store/internal/search"
	"github.com/partflow/smart-store/internal/organizations"
	"github.com/partflow/smart-store/internal/users"
	"github.com/partflow/smart-store/internal/roles"

	"github.com/partflow/smart-store/pkg/config"
	"github.com/partflow/smart-store/pkg/database"
	"github.com/partflow/smart-store/pkg/errors"
	"github.com/partflow/smart-store/pkg/health"
	"github.com/partflow/smart-store/pkg/logger"
	"github.com/partflow/smart-store/pkg/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	logConfig := logger.DefaultConfig()
	logConfig.Level = cfg.LogLevel
	logConfig.EnableConsole = true
	logConfig.EnableFile = true
	logConfig.EnableCaller = true
	logConfig.TimeFormat = time.RFC3339
	if err := logger.Initialize(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logger.Info("Starting PartFlow API Server...")

	// Initialize database
	if err := database.Initialize(); err != nil {
		logger.Fatal("Failed to initialize database", err)
	}
	defer database.Close()

	// Set Gin mode
	gin.SetMode(cfg.ServerMode)

	// Create router
	router := gin.New()

	// Register middleware
	router.Use(middleware.CORS())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.ErrorLoggingMiddleware())
	router.Use(errors.ErrorHandlerMiddleware())
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityLoggingMiddleware())
	router.Use(middleware.PerformanceLoggingMiddleware())
	if cfg.RateLimitEnabled {
		router.Use(middleware.RateLimiter())
	}

	// Set JWT secret
	middleware.SetJWTSecret(cfg.JWTSecret)

	// Initialize services
	db := database.GetDB()

	// Permission service
	permissionService := permissions.NewService(db)
	middleware.SetPermissionService(permissionService)

	// Initialize standard permissions
	if err := permissionService.InitializeStandardPermissions(context.Background()); err != nil {
		logger.Warn("Failed to initialize standard permissions", err)
	}

	// Auth service
	authService, err := auth.NewService(db, cfg.JWTSecret, cfg.UseSupabaseAuth, cfg.SupabaseURL, cfg.SupabaseKey)
	if err != nil {
		logger.Fatal("Failed to initialize auth service", err)
	}

	// Health checker
	healthChecker := health.NewHealthChecker(db)

	// Auth handler
	authHandler := auth.NewHandler(authService, db)
	inventoryHandler := inventory.NewHandler(inventory.NewService(inventory.NewRepository(db), db))
	
	// Register health check routes
	router.GET("/health", healthChecker.Check)
	router.GET("/ready", healthChecker.Readiness)
	router.GET("/alive", healthChecker.Liveness)

	// Log application startup
	logger.LogSystemEvent("application_start", "api", map[string]interface{}{
		"server_mode": cfg.ServerMode,
		"port": cfg.ServerPort,
		"rate_limit_enabled": cfg.RateLimitEnabled,
	})

	// Register API routes
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
			dashboardService := dashboard.NewService(db)
			dashboardHandler := dashboard.NewHandler(dashboardService)
			protected.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)

			// Auth routes
			auth := protected.Group("/auth")
			{
				auth.POST("/logout", authHandler.Logout)
			}

			// User routes (current user info - separate from user management)
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
			productRepo := products.NewRepository(db)
			productService := products.NewService(productRepo)
			productHandler := products.NewHandler(productService)
			
			// Register products routes
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
			customerRepo := customers.NewRepository(db)
			customerService := customers.NewService(customerRepo)
			customerHandler := customers.NewHandler(customerService)
			
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
			salesRepo := sales.NewRepository(db)
			salesService := sales.NewService(salesRepo, db)
			salesHandler := sales.NewHandler(salesService)
			
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
			paymentRepo := payments.NewRepository(db)
			paymentService := payments.NewService(paymentRepo)
			paymentHandler := payments.NewHandler(paymentService)
			
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
			supplierRepo := suppliers.NewRepository(db)
			supplierService := suppliers.NewService(supplierRepo)
			supplierHandler := suppliers.NewHandler(supplierService)
			
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
			purchaseRepo := purchases.NewRepository(db)
			purchaseService := purchases.NewService(purchaseRepo)
			purchaseHandler := purchases.NewHandler(purchaseService)
			
			purchasesRoutes := protected.Group("/purchases")
			{
				purchasesRoutes.POST("", middleware.RequirePermission("purchases", "create"), purchaseHandler.CreatePurchase)
				purchasesRoutes.GET("/:id", middleware.RequirePermission("purchases", "read"), purchaseHandler.GetPurchase)
				purchasesRoutes.GET("", middleware.RequirePermission("purchases", "read"), purchaseHandler.ListPurchases)
				purchasesRoutes.PUT("/:id", middleware.RequirePermission("purchases", "update"), purchaseHandler.UpdatePurchase)
				purchasesRoutes.DELETE("/:id", middleware.RequirePermission("purchases", "update"), purchaseHandler.DeletePurchase)
				purchasesRoutes.POST("/:id/receive", middleware.RequirePermission("purchases", "receive"), purchaseHandler.ReceivePurchase)
				purchasesRoutes.POST("/:id/cancel", middleware.RequirePermission("purchases", "update"), purchaseHandler.CancelPurchase)
				purchasesRoutes.POST("/:id/payment", middleware.RequirePermission("purchases", "update"), purchaseHandler.AddPayment)
				purchasesRoutes.POST("/:id/items", middleware.RequirePermission("purchases", "create"), purchaseHandler.AddPurchaseItem)
				purchasesRoutes.PUT("/items/:item_id", middleware.RequirePermission("purchases", "update"), purchaseHandler.UpdatePurchaseItem)
				purchasesRoutes.DELETE("/items/:item_id", middleware.RequirePermission("purchases", "update"), purchaseHandler.DeletePurchaseItem)
			}

			// Expenses routes
			expenseRepo := expenses.NewRepository(db)
			expenseService := expenses.NewService(expenseRepo)
			expenseHandler := expenses.NewHandler(expenseService)
			
			expensesRoutes := protected.Group("/expenses")
			{
				expensesRoutes.POST("", middleware.RequirePermission("expenses", "create"), expenseHandler.CreateExpense)
				expensesRoutes.GET("/:id", middleware.RequirePermission("expenses", "read"), expenseHandler.GetExpense)
				expensesRoutes.GET("", middleware.RequirePermission("expenses", "read"), expenseHandler.ListExpenses)
				expensesRoutes.PUT("/:id", middleware.RequirePermission("expenses", "update"), expenseHandler.UpdateExpense)
				expensesRoutes.DELETE("/:id", middleware.RequirePermission("expenses", "delete"), expenseHandler.DeleteExpense)
				expensesRoutes.POST("/:id/approve", middleware.RequirePermission("expenses", "update"), expenseHandler.ApproveExpense)
				expensesRoutes.POST("/:id/reject", middleware.RequirePermission("expenses", "update"), expenseHandler.RejectExpense)
				expensesRoutes.GET("/summary", middleware.RequirePermission("expenses", "read"), expenseHandler.GetExpenseSummary)
				expensesRoutes.GET("/categories", middleware.RequirePermission("expenses", "read"), expenseHandler.ListExpenseCategories)
			}

			// Returns routes
			returnRepo := returns.NewRepository(db)
			returnService := returns.NewService(returnRepo)
			returnHandler := returns.NewHandler(returnService)
			
			returnsRoutes := protected.Group("/returns")
			{
				returnsRoutes.POST("", middleware.RequirePermission("returns", "create"), returnHandler.CreateReturn)
				returnsRoutes.GET("/:id", middleware.RequirePermission("returns", "read"), returnHandler.GetReturn)
				returnsRoutes.GET("", middleware.RequirePermission("returns", "read"), returnHandler.ListReturns)
				returnsRoutes.PUT("/:id", middleware.RequirePermission("returns", "update"), returnHandler.UpdateReturn)
				returnsRoutes.DELETE("/:id", middleware.RequirePermission("returns", "update"), returnHandler.DeleteReturn)
				returnsRoutes.POST("/:id/approve", middleware.RequirePermission("returns", "approve"), returnHandler.ApproveReturn)
				returnsRoutes.POST("/:id/reject", middleware.RequirePermission("returns", "approve"), returnHandler.RejectReturn)
				returnsRoutes.POST("/:id/refund", middleware.RequirePermission("sales", "refund"), returnHandler.ProcessRefund)
				returnsRoutes.POST("/:id/items", middleware.RequirePermission("returns", "create"), returnHandler.AddReturnItem)
				returnsRoutes.PUT("/items/:item_id", middleware.RequirePermission("returns", "update"), returnHandler.UpdateReturnItem)
				returnsRoutes.DELETE("/items/:item_id", middleware.RequirePermission("returns", "update"), returnHandler.DeleteReturnItem)
			}

			// Warranties routes
			warrantyRepo := warranties.NewRepository(db)
			warrantyService := warranties.NewService(warrantyRepo)
			warrantyHandler := warranties.NewHandler(warrantyService)
			
			warrantiesRoutes := protected.Group("/warranties")
			{
				warrantiesRoutes.POST("", middleware.RequirePermission("warranties", "read"), warrantyHandler.CreateWarranty)
				warrantiesRoutes.GET("/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.GetWarranty)
				warrantiesRoutes.GET("", middleware.RequirePermission("warranties", "read"), warrantyHandler.ListWarranties)
				warrantiesRoutes.PUT("/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.UpdateWarranty)
				warrantiesRoutes.DELETE("/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.DeleteWarranty)
				warrantiesRoutes.GET("/expiring-soon", middleware.RequirePermission("warranties", "read"), warrantyHandler.GetWarrantiesExpiringSoon)
				warrantiesRoutes.POST("/:id/claims", middleware.RequirePermission("warranties", "claim"), warrantyHandler.CreateWarrantyClaim)
				warrantiesRoutes.GET("/claims/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.GetWarrantyClaim)
				warrantiesRoutes.GET("/claims", middleware.RequirePermission("warranties", "read"), warrantyHandler.ListWarrantyClaims)
				warrantiesRoutes.PUT("/claims/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.UpdateWarrantyClaim)
				warrantiesRoutes.DELETE("/claims/:id", middleware.RequirePermission("warranties", "read"), warrantyHandler.DeleteWarrantyClaim)
				warrantiesRoutes.POST("/claims/:id/approve", middleware.RequirePermission("warranties", "read"), warrantyHandler.ApproveWarrantyClaim)
				warrantiesRoutes.POST("/claims/:id/reject", middleware.RequirePermission("warranties", "read"), warrantyHandler.RejectWarrantyClaim)
				warrantiesRoutes.POST("/claims/:id/complete", middleware.RequirePermission("warranties", "read"), warrantyHandler.CompleteWarrantyClaim)
			}

			// Inspections routes
			inspectionRepo := inspections.NewRepository(db)
			inspectionService := inspections.NewService(inspectionRepo)
			inspectionHandler := inspections.NewHandler(inspectionService)
			
			inspectionsRoutes := protected.Group("/inspections")
			{
				inspectionsRoutes.POST("", inspectionHandler.CreateInspection)
				inspectionsRoutes.GET("/:id", inspectionHandler.GetInspection)
				inspectionsRoutes.GET("", inspectionHandler.ListInspections)
				inspectionsRoutes.PUT("/:id", inspectionHandler.UpdateInspection)
				inspectionsRoutes.DELETE("/:id", inspectionHandler.DeleteInspection)
			}

			// Reports routes
			reportRepo := reports.NewRepository(db)
			reportService := reports.NewService(reportRepo)
			reportHandler := reports.NewHandler(reportService)

			reportsGroup := protected.Group("/reports")
			{
				reportsGroup.POST("", middleware.RequirePermission("reports", "read"), reportHandler.GenerateReport)
				reportsGroup.GET("/:id", middleware.RequirePermission("reports", "read"), reportHandler.GetReport)
				reportsGroup.GET("", middleware.RequirePermission("reports", "read"), reportHandler.ListReports)
				reportsGroup.DELETE("/:id", middleware.RequirePermission("reports", "read"), reportHandler.DeleteReport)
			}

			// Register specific report routes
			reports.RegisterRoutes(protected, db)

			// Notifications routes
			notificationRepo := notifications.NewRepository(db)
			notificationService := notifications.NewService(notificationRepo)
			notificationHandler := notifications.NewHandler(notificationService)
			
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

			// Audit routes
			audit.RegisterRoutes(protected, db)
			// Note: Audit routes are registered with their own permission middleware in the audit package

			// Barcodes routes
			barcodeRepo := barcodes.NewRepository(db)
			barcodeService := barcodes.NewService(barcodeRepo)
			barcodeHandler := barcodes.NewHandler(barcodeService)

			barcodeHandler.RegisterRoutes(protected)

			// Search routes
			search.RegisterRoutes(protected, db)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting...", map[string]interface{}{"port": cfg.ServerPort})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", err)
	}

	logger.Info("Server exited")
}
