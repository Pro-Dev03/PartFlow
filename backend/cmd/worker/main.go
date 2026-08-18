package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"smart-store/pkg/config"
	"smart-store/pkg/database"
	"smart-store/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger.Init()
	logger.Info().Msg("Starting PartFlow Worker Service...")

	// Initialize database
	if err := database.Initialize(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer database.Close()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background workers
	go startReservationExpirationWorker(ctx, database.GetDB().(*sqlx.DB))
	go startDebtScanWorker(ctx, database.GetDB().(*sqlx.DB))
	go startWarrantyExpirationWorker(ctx, database.GetDB().(*sqlx.DB))
	go startLowStockScanWorker(ctx, database.GetDB().(*sqlx.DB))
	go startDailyInsightsWorker(ctx, database.GetDB().(*sqlx.DB))

	logger.Info().Msg("Worker service started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down worker service...")

	// Cancel context to stop all workers
	cancel()

	// Give workers time to clean up
	time.Sleep(5 * time.Second)

	logger.Info().Msg("Worker service exited")
}

// startReservationExpirationWorker checks for expired reservations
func startReservationExpirationWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Reservation expiration worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Checking for expired reservations...")
			processExpiredReservations(ctx, db)
		}
	}
}

func processExpiredReservations(ctx context.Context, db *sqlx.DB) {
	// Find expired active reservations
	query := `
		SELECT id, item_id, organization_id 
		FROM reservations 
		WHERE status = 'active' AND expires_at < NOW()
	`
	
	var expiredReservations []struct {
		ID             uuid.UUID `db:"id"`
		ItemID         uuid.UUID `db:"item_id"`
		OrganizationID uuid.UUID `db:"organization_id"`
	}
	
	err := db.SelectContext(ctx, &expiredReservations, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch expired reservations")
		return
	}
	
	// Process each expired reservation
	for _, reservation := range expiredReservations {
		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to begin transaction")
			continue
		}
		
		// Update reservation status
		updateQuery := `UPDATE reservations SET status = 'expired', updated_at = NOW() WHERE id = $1`
		_, err = tx.ExecContext(ctx, updateQuery, reservation.ID)
		if err != nil {
			tx.Rollback()
			logger.Error().Err(err).Msg("Failed to update reservation status")
			continue
		}
		
		// Update item status back to available
		updateItemQuery := `UPDATE inventory_items SET status = 'AVAILABLE', updated_at = NOW() WHERE id = $1`
		_, err = tx.ExecContext(ctx, updateItemQuery, reservation.ItemID)
		if err != nil {
			tx.Rollback()
			logger.Error().Err(err).Msg("Failed to update item status")
			continue
		}
		
		// Create movement record
		movementQuery := `
			INSERT INTO inventory_movements (id, organization_id, item_id, movement_type, 
				quantity, before_quantity, after_quantity, reference_type, reference_id, 
				reason, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err = tx.ExecContext(ctx, movementQuery,
			uuid.New(), reservation.OrganizationID, reservation.ItemID, "RELEASE",
			1, 0, 1, "reservation", reservation.ID, "Reservation expired", uuid.Nil, time.Now())
		if err != nil {
			tx.Rollback()
			logger.Error().Err(err).Msg("Failed to create movement record")
			continue
		}
		
		if err := tx.Commit(); err != nil {
			logger.Error().Err(err).Msg("Failed to commit transaction")
			continue
		}
		
		logger.Info().Msgf("Processed expired reservation %s", reservation.ID)
	}
	
	logger.Info().Msgf("Processed %d expired reservations", len(expiredReservations))
}

// startDebtScanWorker scans for overdue debts
func startDebtScanWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Debt scan worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Scanning for overdue debts...")
			processOverdueDebts(ctx, db)
		}
	}
}

func processOverdueDebts(ctx context.Context, db *sqlx.DB) {
	// Find overdue debts
	query := `
		SELECT id, customer_id, organization_id, remaining_amount, due_date 
		FROM debts 
		WHERE status = 'pending' AND due_date < NOW()
	`
	
	var overdueDebts []struct {
		ID             uuid.UUID `db:"id"`
		CustomerID     uuid.UUID `db:"customer_id"`
		OrganizationID uuid.UUID `db:"organization_id"`
		RemainingAmount float64  `db:"remaining_amount"`
		DueDate        time.Time `db:"due_date"`
	}
	
	err := db.SelectContext(ctx, &overdueDebts, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch overdue debts")
		return
	}
	
	// Process each overdue debt
	for _, debt := range overdueDebts {
		// Update debt status to overdue
		updateQuery := `UPDATE debts SET status = 'overdue', updated_at = NOW() WHERE id = $1`
		_, err = db.ExecContext(ctx, updateQuery, debt.ID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to update debt status")
			continue
		}
		
		// Create notification for customer
		notificationQuery := `
			INSERT INTO notifications (id, organization_id, user_id, type, title, message, 
				data, is_read, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		`
		
		// Get organization users to notify
		var users []uuid.UUID
		userQuery := `SELECT id FROM users WHERE organization_id = $1 AND is_active = true`
		db.SelectContext(ctx, &users, userQuery, debt.OrganizationID)
		
		for _, userID := range users {
			_, err = db.ExecContext(ctx, notificationQuery,
				uuid.New(), debt.OrganizationID, userID, "debt_overdue",
				"Overdue Payment Alert",
				fmt.Sprintf("Customer has overdue payment of %.2f due on %s", debt.RemainingAmount, debt.DueDate.Format("2006-01-02")),
				fmt.Sprintf(`{"debt_id": "%s", "customer_id": "%s", "amount": %.2f}`, debt.ID, debt.CustomerID, debt.RemainingAmount),
				false)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create notification")
			}
		}
		
		logger.Info().Msgf("Processed overdue debt %s", debt.ID)
	}
	
	logger.Info().Msgf("Processed %d overdue debts", len(overdueDebts))
}

// startWarrantyExpirationWorker checks for expiring warranties
func startWarrantyExpirationWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Warranty expiration worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Checking for expiring warranties...")
		 processExpiringWarranties(ctx, db)
		}
	}
}

func processExpiringWarranties(ctx context.Context, db *sqlx.DB) {
	// Find warranties expiring in the next 30 days
	query := `
		SELECT id, organization_id, sale_id, product_id, expires_at 
		FROM warranties 
		WHERE is_active = true AND expires_at BETWEEN NOW() AND NOW() + INTERVAL '30 days'
	`
	
	var expiringWarranties []struct {
		ID             uuid.UUID `db:"id"`
		OrganizationID uuid.UUID `db:"organization_id"`
		SaleID         uuid.UUID `db:"sale_id"`
		ProductID      uuid.UUID `db:"product_id"`
		ExpiresAt      time.Time `db:"expires_at"`
	}
	
	err := db.SelectContext(ctx, &expiringWarranties, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch expiring warranties")
		return
	}
	
	// Process each expiring warranty
	for _, warranty := range expiringWarranties {
		// Get organization users to notify
		var users []uuid.UUID
		userQuery := `SELECT id FROM users WHERE organization_id = $1 AND is_active = true`
		db.SelectContext(ctx, &users, userQuery, warranty.OrganizationID)
		
		// Create notification for each user
		notificationQuery := `
			INSERT INTO notifications (id, organization_id, user_id, type, title, message, 
				data, is_read, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		`
		
		for _, userID := range users {
			_, err = db.ExecContext(ctx, notificationQuery,
				uuid.New(), warranty.OrganizationID, userID, "warranty_expiring",
				"Warranty Expiring Soon",
				fmt.Sprintf("Warranty expires on %s for product", warranty.ExpiresAt.Format("2006-01-02")),
				fmt.Sprintf(`{"warranty_id": "%s", "sale_id": "%s", "product_id": "%s", "expires_at": "%s"}`, 
					warranty.ID, warranty.SaleID, warranty.ProductID, warranty.ExpiresAt.Format(time.RFC3339)),
				false)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create notification")
			}
		}
		
		logger.Info().Msgf("Processed expiring warranty %s", warranty.ID)
	}
	
	logger.Info().Msgf("Processed %d expiring warranties", len(expiringWarranties))
}

// startLowStockScanWorker scans for low stock items
func startLowStockScanWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Low stock scan worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Scanning for low stock items...")
			processLowStockItems(ctx, db)
		}
	}
}

func processLowStockItems(ctx context.Context, db *sqlx.DB) {
	// Find products with low stock
	query := `
		SELECT p.id, p.name, p.organization_id, p.min_stock_level, 
		       COALESCE(SUM(ii.quantity), 0) as current_stock
		FROM products p
		LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE'
		WHERE p.is_active = true
		GROUP BY p.id, p.name, p.organization_id, p.min_stock_level
		HAVING COALESCE(SUM(ii.quantity), 0) <= p.min_stock_level
	`
	
	var lowStockItems []struct {
		ID             uuid.UUID `db:"id"`
		Name           string    `db:"name"`
		OrganizationID uuid.UUID `db:"organization_id"`
		MinStockLevel  int       `db:"min_stock_level"`
		CurrentStock   int       `db:"current_stock"`
	}
	
	err := db.SelectContext(ctx, &lowStockItems, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch low stock items")
		return
	}
	
	// Process each low stock item
	for _, item := range lowStockItems {
		// Get organization users to notify
		var users []uuid.UUID
		userQuery := `SELECT id FROM users WHERE organization_id = $1 AND is_active = true`
		db.SelectContext(ctx, &users, userQuery, item.OrganizationID)
		
		// Create notification for each user
		notificationQuery := `
			INSERT INTO notifications (id, organization_id, user_id, type, title, message, 
				data, is_read, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		`
		
		for _, userID := range users {
			_, err = db.ExecContext(ctx, notificationQuery,
				uuid.New(), item.OrganizationID, userID, "low_stock",
				"Low Stock Alert",
				fmt.Sprintf("Product '%s' is running low on stock (current: %d, minimum: %d)", 
					item.Name, item.CurrentStock, item.MinStockLevel),
				fmt.Sprintf(`{"product_id": "%s", "product_name": "%s", "current_stock": %d, "min_stock_level": %d}`, 
					item.ID, item.Name, item.CurrentStock, item.MinStockLevel),
				false)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create notification")
			}
		}
		
		logger.Info().Msgf("Processed low stock item %s", item.ID)
	}
	
	logger.Info().Msgf("Processed %d low stock items", len(lowStockItems))
}

// startDailyInsightsWorker generates daily insights
func startDailyInsightsWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Daily insights worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Generating daily insights...")
			generateDailyInsights(ctx, db)
		}
	}
}

func generateDailyInsights(ctx context.Context, db *sqlx.DB) {
	// Get all active organizations
	query := `SELECT id FROM organizations WHERE subscription_status = 'active'`
	
	var organizations []uuid.UUID
	err := db.SelectContext(ctx, &organizations, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch organizations")
		return
	}
	
	// Generate insights for each organization
	for _, orgID := range organizations {
		// Get today's sales summary
		today := time.Now().Format("2006-01-02")
		salesQuery := `
			SELECT 
				COUNT(*) as total_sales,
				COALESCE(SUM(total_amount), 0) as total_revenue,
				COALESCE(SUM(gross_profit), 0) as total_profit
			FROM sales 
			WHERE organization_id = $1 AND sale_date = $2 AND status = 'completed'
		`
		
		var salesSummary struct {
			TotalSales   int     `db:"total_sales"`
			TotalRevenue float64 `db:"total_revenue"`
			TotalProfit  float64 `db:"total_profit"`
		}
		
		err = db.GetContext(ctx, &salesSummary, salesQuery, orgID, today)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to fetch sales summary")
			continue
		}
		
		// Get low stock count
		lowStockQuery := `
			SELECT COUNT(DISTINCT p.id)
			FROM products p
			LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE'
			WHERE p.organization_id = $1 AND p.is_active = true
			GROUP BY p.id, p.min_stock_level
			HAVING COALESCE(SUM(ii.quantity), 0) <= p.min_stock_level
		`
		
		var lowStockCount int
		db.GetContext(ctx, &lowStockCount, lowStockQuery, orgID)
		
		// Get overdue debts count
		overdueQuery := `
			SELECT COUNT(*) FROM debts 
			WHERE organization_id = $1 AND status = 'overdue'
		`
		
		var overdueCount int
		db.GetContext(ctx, &overdueCount, overdueQuery, orgID)
		
		// Create notification with daily insights
		notificationQuery := `
			INSERT INTO notifications (id, organization_id, user_id, type, title, message, 
				data, is_read, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		`
		
		// Get organization users to notify
		var users []uuid.UUID
		userQuery := `SELECT id FROM users WHERE organization_id = $1 AND is_active = true`
		db.SelectContext(ctx, &users, userQuery, orgID)
		
		insightData := fmt.Sprintf(`{
			"date": "%s",
			"total_sales": %d,
			"total_revenue": %.2f,
			"total_profit": %.2f,
			"low_stock_count": %d,
			"overdue_debts_count": %d
		}`, today, salesSummary.TotalSales, salesSummary.TotalRevenue, salesSummary.TotalProfit, lowStockCount, overdueCount)
		
		for _, userID := range users {
			_, err = db.ExecContext(ctx, notificationQuery,
				uuid.New(), orgID, userID, "daily_insights",
				"Daily Business Insights",
				fmt.Sprintf("Today's summary: %d sales, %.2f revenue, %d low stock items, %d overdue debts", 
					salesSummary.TotalSales, salesSummary.TotalRevenue, lowStockCount, overdueCount),
				insightData,
				false)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create notification")
			}
		}
		
		logger.Info().Msgf("Generated daily insights for organization %s", orgID)
	}
	
	logger.Info().Msgf("Generated daily insights for %d organizations", len(organizations))
}
