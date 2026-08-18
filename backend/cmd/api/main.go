package main

import (
	"log"
	"smart-store/internal/auth"
	"smart-store/internal/organizations"
	"smart-store/internal/users"
	"smart-store/internal/roles"
	"smart-store/internal/products"
	"smart-store/internal/inventory"
	"smart-store/internal/sales"
	"smart-store/internal/payments"
	"smart-store/internal/customers"
	"smart-store/internal/debts"
	"smart-store/internal/purchases"
	"smart-store/internal/suppliers"
	"smart-store/internal/expenses"
	"smart-store/internal/returns"
	"smart-store/internal/warranties"
	"smart-store/internal/inspections"
	"smart-store/internal/reports"
	"smart-store/internal/notifications"
	"smart-store/internal/automation"
	"smart-store/internal/audit"
	"smart-store/pkg/logger"
	"smart-store/pkg/middleware"
)

func main() {
	// Initialize logger
	logger.Init()

	// Initialize database connection
	// db := database.Initialize()

	// Initialize router
	// router := gin.Default()

	// Register middleware
	// router.Use(middleware.CORS())
	// router.Use(middleware.Logger())
	// router.Use(middleware.Recovery())

	// Register routes
	// auth.RegisterRoutes(router, db)
	// organizations.RegisterRoutes(router, db)
	// users.RegisterRoutes(router, db)
	// roles.RegisterRoutes(router, db)
	// products.RegisterRoutes(router, db)
	// inventory.RegisterRoutes(router, db)
	// sales.RegisterRoutes(router, db)
	// payments.RegisterRoutes(router, db)
	// customers.RegisterRoutes(router, db)
	// debts.RegisterRoutes(router, db)
	// purchases.RegisterRoutes(router, db)
	// suppliers.RegisterRoutes(router, db)
	// expenses.RegisterRoutes(router, db)
	// returns.RegisterRoutes(router, db)
	// warranties.RegisterRoutes(router, db)
	// inspections.RegisterRoutes(router, db)
	// reports.RegisterRoutes(router, db)
	// notifications.RegisterRoutes(router, db)
	// automation.RegisterRoutes(router, db)
	// audit.RegisterRoutes(router, db)

	log.Println("Smart Store API Server starting...")
	// router.Run(":8080")
}
