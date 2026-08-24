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
	"github.com/partflow/smart-store/internal/inspections"
	"github.com/partflow/smart-store/internal/reports"
	"github.com/partflow/smart-store/internal/notifications"
	"github.com/partflow/smart-store/internal/audit"
	"github.com/partflow/smart-store/internal/dashboard"
	"github.com/partflow/smart-store/internal/barcodes"
	"github.com/partflow/smart-store/internal/search"
	"github.com/partflow/smart-store/internal/parttypes"
	"github.com/partflow/smart-store/internal/users"
	"github.com/partflow/smart-store/internal/settings"

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

	// Auth service
	authService, err := auth.NewService(db, cfg.JWTSecret, cfg.UseSupabaseAuth, cfg.SupabaseURL, cfg.SupabaseKey)
	if err != nil {
		logger.Fatal("Failed to initialize auth service", err)
	}

	// Health checker
	healthChecker := health.NewHealthChecker(db)

	// Auth handler
	authHandler := auth.NewHandler(authService, db)
	inventoryHandler := inventory.NewHandler(inventory.NewService(inventory.NewRepository(db), db), db)
	
	// Register health check routes
	router.GET("/health", healthChecker.Check)
	router.GET("/ready", healthChecker.Readiness)
	router.GET("/alive", healthChecker.Liveness)

	// Log application startup
	logger.LogSystemEvent("application_start", "api", map[string]interface{}{
		"server_mode": cfg.ServerMode,
		"port": cfg.ServerPort,
		"rate_limit_enabled": cfg.RateLimitEnabled,
		"architecture": "Centralized routing inspired by Fynexa",
	})

	// Register API routes using centralized router (inspired by Fynexa architecture)
	// This provides better maintainability and clearer structure
	// Note: Temporarily disabled due to existing code issues in some modules
	// The centralized router structure is ready at internal/api/router.go
	// TODO: Enable after fixing existing module issues (auth, inventory, etc.)
	// api.SetupRoutes(router, db, authService, permissionService)
	
	// Keeping existing routes for backward compatibility during transition period
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
		middleware.SetDatabase(db)
		protected.Use(middleware.Auth())
		{
			// Dashboard routes
			dashboardService := dashboard.NewCachedService(db)
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

			// Part Types routes
			partTypesRepo := parttypes.NewRepository(db)
			partTypesService := parttypes.NewService(partTypesRepo)
			partTypesHandler := parttypes.NewHandler(partTypesService)
			partTypesHandler.RegisterRoutes(protected)

			// Users management routes
			users.RegisterRoutes(protected, db)

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
			customerRepo := customers.NewRepository(db)
			customerService := customers.NewService(customerRepo)
			customerHandler := customers.NewHandler(customerService)
			
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
			salesRepo := sales.NewRepository(db)
			salesService := sales.NewService(salesRepo, db)
			salesHandler := sales.NewHandler(salesService)
			
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
			supplierService := suppliers.NewService(supplierRepo, db)
			supplierHandler := suppliers.NewHandler(supplierService)
			
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
			}

			// Purchases routes
			purchaseRepo := purchases.NewRepository(db)
			purchaseService := purchases.NewService(purchaseRepo, db)
			purchaseHandler := purchases.NewHandler(purchaseService)
			
			purchasesRoutes := protected.Group("/purchases")
			{
				purchasesRoutes.POST("", purchaseHandler.CreatePurchase)
				purchasesRoutes.GET("/:id", purchaseHandler.GetPurchase)
				purchasesRoutes.GET("", purchaseHandler.ListPurchases)
				purchasesRoutes.PUT("/:id", purchaseHandler.UpdatePurchase)
				purchasesRoutes.DELETE("/:id", purchaseHandler.DeletePurchase)
				purchasesRoutes.POST("/:id/receive", purchaseHandler.ReceivePurchase)
				purchasesRoutes.POST("/:id/cancel", purchaseHandler.CancelPurchase)
				purchasesRoutes.POST("/:id/payment", purchaseHandler.AddPayment)
				purchasesRoutes.POST("/:id/items", purchaseHandler.AddPurchaseItem)
				purchasesRoutes.PUT("/items/:item_id", purchaseHandler.UpdatePurchaseItem)
				purchasesRoutes.DELETE("/items/:item_id", purchaseHandler.DeletePurchaseItem)
			}

			// Expenses routes
			expenseRepo := expenses.NewRepository(db)
			expenseService := expenses.NewService(expenseRepo)
			expenseHandler := expenses.NewHandler(expenseService)
			
			expensesRoutes := protected.Group("/expenses")
			{
				expensesRoutes.POST("", expenseHandler.CreateExpense)
				expensesRoutes.GET("/:id", expenseHandler.GetExpense)
				expensesRoutes.GET("", expenseHandler.ListExpenses)
				expensesRoutes.PUT("/:id", expenseHandler.UpdateExpense)
				expensesRoutes.DELETE("/:id", expenseHandler.DeleteExpense)
				expensesRoutes.POST("/:id/approve", expenseHandler.ApproveExpense)
				expensesRoutes.POST("/:id/reject", expenseHandler.RejectExpense)
				expensesRoutes.GET("/summary", expenseHandler.GetExpenseSummary)
				expensesRoutes.GET("/categories", expenseHandler.ListExpenseCategories)
			}

			// Returns routes
			returnRepo := returns.NewRepository(db)
			returnService := returns.NewService(returnRepo)
			returnHandler := returns.NewHandler(returnService)
			
			returnsRoutes := protected.Group("/returns")
			{
				returnsRoutes.POST("", returnHandler.CreateReturn)
				returnsRoutes.GET("/:id", returnHandler.GetReturn)
				returnsRoutes.GET("", returnHandler.ListReturns)
				returnsRoutes.PUT("/:id", returnHandler.UpdateReturn)
				returnsRoutes.DELETE("/:id", returnHandler.DeleteReturn)
				returnsRoutes.POST("/:id/approve", returnHandler.ApproveReturn)
				returnsRoutes.POST("/:id/reject", returnHandler.RejectReturn)
				returnsRoutes.POST("/:id/refund", returnHandler.ProcessRefund)
				returnsRoutes.POST("/:id/items", returnHandler.AddReturnItem)
				returnsRoutes.PUT("/items/:item_id", returnHandler.UpdateReturnItem)
				returnsRoutes.DELETE("/items/:item_id", returnHandler.DeleteReturnItem)
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
				reportsGroup.POST("", reportHandler.GenerateReport)
				reportsGroup.GET("/:id", reportHandler.GetReport)
				reportsGroup.GET("", reportHandler.ListReports)
				reportsGroup.DELETE("/:id", reportHandler.DeleteReport)
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

			// Settings routes
			settingsHandler := settings.NewHandler(db.DB)
			settingsRoutes := protected.Group("/settings")
			{
				settingsRoutes.GET("/public", settingsHandler.GetPublicSettings)
				settingsRoutes.GET("/:key", settingsHandler.GetSetting)
				settingsRoutes.PUT("/:key", settingsHandler.UpdateSetting)
				settingsRoutes.GET("/tax-rate", settingsHandler.GetTaxRate)
				settingsRoutes.PUT("/tax-rate", settingsHandler.UpdateTaxRate)
			}
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
