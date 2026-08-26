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

	"github.com/partflow/smart-store/pkg/config"
	"github.com/partflow/smart-store/pkg/database"
	"github.com/partflow/smart-store/pkg/logger"
)

func main() {
	if _, err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	logConfig := logger.DefaultConfig()
	logConfig.Level = "info"
	logConfig.EnableConsole = true
	logConfig.EnableFile = true
	logConfig.EnableCaller = true
	logConfig.TimeFormat = time.RFC3339
	if err := logger.Initialize(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("Starting PartFlow Worker Service...", nil)

	if err := database.Initialize(); err != nil {
		logger.Fatal("Failed to initialize database", err, nil)
	}
	defer database.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startReservationExpirationWorker(ctx, database.GetDB())
	go startDebtScanWorker(ctx, database.GetDB())
	go startLowStockScanWorker(ctx, database.GetDB())
	go startDailyInsightsWorker(ctx, database.GetDB())

	logger.Info("Worker service started successfully", nil)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker service...", nil)

	cancel()
	time.Sleep(5 * time.Second)

	logger.Info("Worker service exited", nil)
}

func notifyAllUsers(ctx context.Context, db *sqlx.DB, notifType, title, message string, data string) {
	var users []uuid.UUID
	userQuery := `SELECT id FROM users WHERE is_active = true`
	if err := db.SelectContext(ctx, &users, userQuery); err != nil {
		logger.Error("Failed to fetch active users for notification", err, nil)
		return
	}

	notificationQuery := `
		INSERT INTO notifications (id, user_id, type, title, message,
			data, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

	for _, userID := range users {
		_, err := db.ExecContext(ctx, notificationQuery,
			uuid.New(), userID, notifType, title, message, data, false)
		if err != nil {
			logger.Error("Failed to create notification", err, nil)
		}
	}
}

func startReservationExpirationWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Reservation expiration worker stopped", nil)
			return
		case <-ticker.C:
			logger.Info("Checking for expired reservations...", nil)
			processExpiredReservations(ctx, db)
		}
	}
}

func processExpiredReservations(ctx context.Context, db *sqlx.DB) {
	query := `
		SELECT id, item_id
		FROM reservations
		WHERE status = 'active' AND expires_at < NOW()
	`

	var expiredReservations []struct {
		ID     uuid.UUID `db:"id"`
		ItemID uuid.UUID `db:"item_id"`
	}

	err := db.SelectContext(ctx, &expiredReservations, query)
	if err != nil {
		logger.Error("Failed to fetch expired reservations", err, nil)
		return
	}

	for _, reservation := range expiredReservations {
		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			logger.Error("Failed to begin transaction", err, nil)
			continue
		}

		updateQuery := `UPDATE reservations SET status = 'expired', updated_at = NOW() WHERE id = $1`
		_, err = tx.ExecContext(ctx, updateQuery, reservation.ID)
		if err != nil {
			tx.Rollback()
			logger.Error("Failed to update reservation status", err, nil)
			continue
		}

		updateItemQuery := `UPDATE inventory_items SET status = 'AVAILABLE', updated_at = NOW() WHERE id = $1`
		_, err = tx.ExecContext(ctx, updateItemQuery, reservation.ItemID)
		if err != nil {
			tx.Rollback()
			logger.Error("Failed to update item status", err, nil)
			continue
		}

		movementQuery := `
			INSERT INTO inventory_movements (id, item_id, movement_type,
				quantity, before_quantity, after_quantity, reference_type, reference_id,
				reason, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err = tx.ExecContext(ctx, movementQuery,
			uuid.New(), reservation.ItemID, "RELEASE",
			1, 0, 1, "reservation", reservation.ID, "Reservation expired", uuid.Nil, time.Now())
		if err != nil {
			tx.Rollback()
			logger.Error("Failed to create movement record", err, nil)
			continue
		}

		if err := tx.Commit(); err != nil {
			logger.Error("Failed to commit transaction", err, nil)
			continue
		}

		logger.Info(fmt.Sprintf("Processed expired reservation %s", reservation.ID), nil)
	}

	logger.Info(fmt.Sprintf("Processed %d expired reservations", len(expiredReservations)), nil)
}

func startDebtScanWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Debt scan worker stopped", nil)
			return
		case <-ticker.C:
			logger.Info("Scanning for overdue debts...", nil)
			processOverdueDebts(ctx, db)
		}
	}
}

func processOverdueDebts(ctx context.Context, db *sqlx.DB) {
	query := `
		SELECT id, customer_id, remaining_amount, due_date
		FROM debts
		WHERE status = 'pending' AND due_date < NOW()
	`

	var overdueDebts []struct {
		ID              uuid.UUID `db:"id"`
		CustomerID      uuid.UUID `db:"customer_id"`
		RemainingAmount float64   `db:"remaining_amount"`
		DueDate         time.Time `db:"due_date"`
	}

	err := db.SelectContext(ctx, &overdueDebts, query)
	if err != nil {
		logger.Error("Failed to fetch overdue debts", err, nil)
		return
	}

	for _, debt := range overdueDebts {
		updateQuery := `UPDATE debts SET status = 'overdue', updated_at = NOW() WHERE id = $1`
		_, err = db.ExecContext(ctx, updateQuery, debt.ID)
		if err != nil {
			logger.Error("Failed to update debt status", err, nil)
			continue
		}

		notifyAllUsers(ctx, db, "debt_overdue",
			"Overdue Payment Alert",
			fmt.Sprintf("Customer has overdue payment of %.2f due on %s", debt.RemainingAmount, debt.DueDate.Format("2006-01-02")),
			fmt.Sprintf(`{"debt_id": "%s", "customer_id": "%s", "amount": %.2f}`, debt.ID, debt.CustomerID, debt.RemainingAmount))

		logger.Info(fmt.Sprintf("Processed overdue debt %s", debt.ID), nil)
	}

	logger.Info(fmt.Sprintf("Processed %d overdue debts", len(overdueDebts)), nil)
}

func startLowStockScanWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Low stock scan worker stopped", nil)
			return
		case <-ticker.C:
			logger.Info("Scanning for low stock items...", nil)
			processLowStockItems(ctx, db)
		}
	}
}

func processLowStockItems(ctx context.Context, db *sqlx.DB) {
	query := `
		SELECT p.id, p.name, p.min_stock_level,
		       COALESCE(SUM(ii.quantity), 0) as current_stock
		FROM products p
		LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE'
		WHERE p.is_active = true
		GROUP BY p.id, p.name, p.min_stock_level
		HAVING COALESCE(SUM(ii.quantity), 0) <= p.min_stock_level
	`

	var lowStockItems []struct {
		ID            uuid.UUID `db:"id"`
		Name          string    `db:"name"`
		MinStockLevel int       `db:"min_stock_level"`
		CurrentStock  int       `db:"current_stock"`
	}

	err := db.SelectContext(ctx, &lowStockItems, query)
	if err != nil {
		logger.Error("Failed to fetch low stock items", err, nil)
		return
	}

	for _, item := range lowStockItems {
		notifyAllUsers(ctx, db, "low_stock",
			"Low Stock Alert",
			fmt.Sprintf("Product '%s' is running low on stock (current: %d, minimum: %d)",
				item.Name, item.CurrentStock, item.MinStockLevel),
			fmt.Sprintf(`{"product_id": "%s", "product_name": "%s", "current_stock": %d, "min_stock_level": %d}`,
				item.ID, item.Name, item.CurrentStock, item.MinStockLevel))

		logger.Info(fmt.Sprintf("Processed low stock item %s", item.ID), nil)
	}

	logger.Info(fmt.Sprintf("Processed %d low stock items", len(lowStockItems)), nil)
}

func startDailyInsightsWorker(ctx context.Context, db *sqlx.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Daily insights worker stopped", nil)
			return
		case <-ticker.C:
			logger.Info("Generating daily insights...", nil)
			generateDailyInsights(ctx, db)
		}
	}
}

func generateDailyInsights(ctx context.Context, db *sqlx.DB) {
	today := time.Now().Format("2006-01-02")

	var salesSummary struct {
		TotalSales   int     `db:"total_sales"`
		TotalRevenue float64 `db:"total_revenue"`
		TotalProfit  float64 `db:"total_profit"`
	}
	salesQuery := `
		SELECT
			COUNT(*) as total_sales,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(gross_profit), 0) as total_profit
		FROM sales
		WHERE sale_date = $1 AND status = 'completed'
	`
	err := db.GetContext(ctx, &salesSummary, salesQuery, today)
	if err != nil {
		logger.Error("Failed to fetch sales summary", err, nil)
		return
	}

	var lowStockCount int
	lowStockQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM products p
		LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE'
		WHERE p.is_active = true
		HAVING COALESCE(SUM(ii.quantity), 0) <= p.min_stock_level
	`
	err = db.GetContext(ctx, &lowStockCount, lowStockQuery)
	if err != nil {
		logger.Error("Failed to fetch low stock count for insights", err, nil)
		return
	}

	var overdueCount int
	overdueQuery := `SELECT COUNT(*) FROM debts WHERE status = 'overdue'`
	err = db.GetContext(ctx, &overdueCount, overdueQuery)
	if err != nil {
		logger.Error("Failed to fetch overdue debt count for insights", err, nil)
		return
	}

	insightData := fmt.Sprintf(`{
		"date": "%s",
		"total_sales": %d,
		"total_revenue": %.2f,
		"total_profit": %.2f,
		"low_stock_count": %d,
		"overdue_debts_count": %d
	}`, today, salesSummary.TotalSales, salesSummary.TotalRevenue, salesSummary.TotalProfit, lowStockCount, overdueCount)

	notifyAllUsers(ctx, db, "daily_insights",
		"Daily Business Insights",
		fmt.Sprintf("Today's summary: %d sales, %.2f revenue, %d low stock items, %d overdue debts",
			salesSummary.TotalSales, salesSummary.TotalRevenue, lowStockCount, overdueCount),
		insightData)

	logger.Info("Generated daily insights for system", nil)
}
