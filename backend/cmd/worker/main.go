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

	dbutil "github.com/partflow/smart-store/internal/database"
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
	var users []string
	userQuery := `SELECT id FROM users WHERE is_active = true`
	if dbutil.IsSQLite(db) {
		userQuery = `SELECT id FROM users WHERE is_active = 1`
	}
	if err := db.SelectContext(ctx, &users, userQuery); err != nil {
		logger.Error("Failed to fetch active users for notification", err, nil)
		return
	}

	notificationQuery := `
		INSERT INTO notifications (id, user_id, type, title, message,
			data, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	if dbutil.IsSQLite(db) {
		notificationQuery = `INSERT INTO notifications (id, user_id, type, title, message, data, priority, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, 'medium', 'unread', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	}

	for _, userIDValue := range users {
		userID, parseErr := uuid.Parse(userIDValue)
		if parseErr != nil {
			logger.Error("Failed to parse active user ID", parseErr, nil)
			continue
		}
		var err error
		if dbutil.IsSQLite(db) {
			_, err = db.ExecContext(ctx, notificationQuery, uuid.New(), userID, notifType, title, message, data)
		} else {
			_, err = db.ExecContext(ctx, notificationQuery, uuid.New(), userID, notifType, title, message, data, false)
		}
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
		SELECT r.id, r.item_id, ii.product_id, r.user_id
		FROM reservations r
		JOIN inventory_items ii ON ii.id = r.item_id
		WHERE r.status = 'active' AND r.expires_at < NOW()
	`
	if dbutil.IsSQLite(db) {
		query = `SELECT r.id, r.item_id, ii.product_id, r.user_id FROM reservations r JOIN inventory_items ii ON ii.id = r.item_id WHERE r.status = 'active' AND r.expires_at < CURRENT_TIMESTAMP`
	}

	var expiredReservations []struct {
		ID        string `db:"id"`
		ItemID    string `db:"item_id"`
		ProductID string `db:"product_id"`
		CreatedBy string `db:"user_id"`
	}

	err := db.SelectContext(ctx, &expiredReservations, query)
	if err != nil {
		logger.Error("Failed to fetch expired reservations", err, nil)
		return
	}

	for _, reservation := range expiredReservations {
		reservationID, parseErr := uuid.Parse(reservation.ID)
		itemID, itemErr := uuid.Parse(reservation.ItemID)
		productID, productErr := uuid.Parse(reservation.ProductID)
		createdBy, userErr := uuid.Parse(reservation.CreatedBy)
		if parseErr != nil || itemErr != nil || productErr != nil || userErr != nil {
			logger.Error("Failed to parse reservation identifiers", fmt.Errorf("reservation=%s", reservation.ID), nil)
			continue
		}
		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			logger.Error("Failed to begin transaction", err, nil)
			continue
		}

		updateQuery := fmt.Sprintf(`UPDATE reservations SET status = 'expired', updated_at = %s WHERE id = $1`, dbutil.NowSQL(db))
		_, err = tx.ExecContext(ctx, updateQuery, reservationID)
		if err != nil {
			tx.Rollback()
			logger.Error("Failed to update reservation status", err, nil)
			continue
		}

		updateItemQuery := fmt.Sprintf(`UPDATE inventory_items SET status = 'AVAILABLE', updated_at = %s WHERE id = $1`, dbutil.NowSQL(db))
		_, err = tx.ExecContext(ctx, updateItemQuery, itemID)
		if err != nil {
			tx.Rollback()
			logger.Error("Failed to update item status", err, nil)
			continue
		}
		var beforeAvailable int
		if err = tx.GetContext(ctx, &beforeAvailable, `SELECT COUNT(*) FROM inventory_items WHERE product_id = $1 AND status = 'AVAILABLE'`, productID); err != nil {
			tx.Rollback()
			logger.Error("Failed to read available quantity", err, nil)
			continue
		}
		if _, err = tx.ExecContext(ctx, `UPDATE inventory SET reserved_quantity = CASE WHEN COALESCE(reserved_quantity, 0) > 0 THEN reserved_quantity - 1 ELSE 0 END WHERE product_id = $1`, productID); err != nil {
			tx.Rollback()
			logger.Error("Failed to update reserved quantity", err, nil)
			continue
		}

		movementQuery := `
			INSERT INTO inventory_movements (id, item_id, movement_type,
				quantity, before_quantity, after_quantity, reference_type, reference_id,
				reason, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err = tx.ExecContext(ctx, movementQuery,
			uuid.New(), itemID, "RELEASE",
			1, beforeAvailable, beforeAvailable+1, "reservation", reservationID, "Reservation expired", createdBy, time.Now())
		if err != nil {
			tx.Rollback()
			logger.Error("Failed to create movement record", err, nil)
			continue
		}

		if err := tx.Commit(); err != nil {
			logger.Error("Failed to commit transaction", err, nil)
			continue
		}

		logger.Info(fmt.Sprintf("Processed expired reservation %s", reservationID), nil)
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
	if dbutil.IsSQLite(db) {
		query = `SELECT id, customer_id, remaining_amount, due_date FROM debts WHERE status = 'pending' AND date(due_date) < date('now')`
	}

	var overdueDebts []struct {
		ID              string  `db:"id"`
		CustomerID      string  `db:"customer_id"`
		RemainingAmount float64 `db:"remaining_amount"`
		DueDate         string  `db:"due_date"`
	}

	err := db.SelectContext(ctx, &overdueDebts, query)
	if err != nil {
		logger.Error("Failed to fetch overdue debts", err, nil)
		return
	}

	for _, debt := range overdueDebts {
		debtID, parseErr := uuid.Parse(debt.ID)
		customerID, customerErr := uuid.Parse(debt.CustomerID)
		dueDate, dueErr := dbutil.ParseTimestamp(debt.DueDate)
		if parseErr != nil || customerErr != nil || dueErr != nil {
			logger.Error("Failed to parse overdue debt identifiers", fmt.Errorf("debt=%s", debt.ID), nil)
			continue
		}
		updateQuery := fmt.Sprintf(`UPDATE debts SET status = 'overdue', updated_at = %s WHERE id = $1`, dbutil.NowSQL(db))
		_, err = db.ExecContext(ctx, updateQuery, debtID)
		if err != nil {
			logger.Error("Failed to update debt status", err, nil)
			continue
		}

		notifyAllUsers(ctx, db, "debt_overdue",
			"Overdue Payment Alert",
			fmt.Sprintf("Customer has overdue payment of %.2f due on %s", debt.RemainingAmount, dueDate.Format("2006-01-02")),
			fmt.Sprintf(`{"debt_id": "%s", "customer_id": "%s", "amount": %.2f}`, debtID, customerID, debt.RemainingAmount))

		logger.Info(fmt.Sprintf("Processed overdue debt %s", debtID), nil)
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
		       COUNT(ii.id) as current_stock
		FROM products p
		LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE'
		WHERE p.is_active = true
		GROUP BY p.id, p.name, p.min_stock_level
		HAVING COUNT(ii.id) <= p.min_stock_level
	`
	if dbutil.IsSQLite(db) {
		query = `SELECT p.id, p.name, p.min_stock_level, COUNT(ii.id) AS current_stock FROM products p LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE' WHERE p.is_active = 1 GROUP BY p.id, p.name, p.min_stock_level HAVING COUNT(ii.id) <= p.min_stock_level`
	}

	var lowStockItems []struct {
		ID            string `db:"id"`
		Name          string `db:"name"`
		MinStockLevel int    `db:"min_stock_level"`
		CurrentStock  int    `db:"current_stock"`
	}

	err := db.SelectContext(ctx, &lowStockItems, query)
	if err != nil {
		logger.Error("Failed to fetch low stock items", err, nil)
		return
	}

	for _, item := range lowStockItems {
		productID, parseErr := uuid.Parse(item.ID)
		if parseErr != nil {
			logger.Error("Failed to parse low stock product ID", parseErr, nil)
			continue
		}
		notifyAllUsers(ctx, db, "low_stock",
			"Low Stock Alert",
			fmt.Sprintf("Product '%s' is running low on stock (current: %d, minimum: %d)",
				item.Name, item.CurrentStock, item.MinStockLevel),
			fmt.Sprintf(`{"product_id": "%s", "product_name": "%s", "current_stock": %d, "min_stock_level": %d}`,
				productID, item.Name, item.CurrentStock, item.MinStockLevel))

		logger.Info(fmt.Sprintf("Processed low stock item %s", productID), nil)
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
	if dbutil.IsSQLite(db) {
		salesQuery = `SELECT COUNT(*) AS total_sales, COALESCE(SUM(total_amount), 0) AS total_revenue, COALESCE(SUM(COALESCE(gross_profit, total_amount - cost_amount, 0)), 0) AS total_profit FROM sales WHERE date(COALESCE(sale_date, created_at)) = $1 AND lower(COALESCE(status, 'completed')) = 'completed'`
	}
	err := db.GetContext(ctx, &salesSummary, salesQuery, today)
	if err != nil {
		logger.Error("Failed to fetch sales summary", err, nil)
		return
	}

	var lowStockCount int
	lowStockQuery := `
		SELECT COUNT(*) FROM (
			SELECT p.id
			FROM products p
			LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE'
			WHERE p.is_active = true
			GROUP BY p.id, p.min_stock_level
			HAVING COUNT(ii.id) <= p.min_stock_level
		) low_stock
	`
	if dbutil.IsSQLite(db) {
		lowStockQuery = `SELECT COUNT(*) FROM (SELECT p.id FROM products p LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE' WHERE p.is_active = 1 GROUP BY p.id, p.min_stock_level HAVING COUNT(ii.id) <= p.min_stock_level) low_stock`
	}
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
