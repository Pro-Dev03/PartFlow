package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Msg("Starting PartFlow Worker Service...")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background workers
	go startReservationExpirationWorker(ctx, logger)
	go startDebtScanWorker(ctx, logger)
	go startWarrantyExpirationWorker(ctx, logger)
	go startLowStockScanWorker(ctx, logger)
	go startDailyInsightsWorker(ctx, logger)

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
func startReservationExpirationWorker(ctx context.Context, logger zerolog.Logger) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Reservation expiration worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Checking for expired reservations...")
			// TODO: Implement reservation expiration logic
		}
	}
}

// startDebtScanWorker scans for overdue debts
func startDebtScanWorker(ctx context.Context, logger zerolog.Logger) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Debt scan worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Scanning for overdue debts...")
			// TODO: Implement debt scan logic
		}
	}
}

// startWarrantyExpirationWorker checks for expiring warranties
func startWarrantyExpirationWorker(ctx context.Context, logger zerolog.Logger) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Warranty expiration worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Checking for expiring warranties...")
			// TODO: Implement warranty expiration logic
		}
	}
}

// startLowStockScanWorker scans for low stock items
func startLowStockScanWorker(ctx context.Context, logger zerolog.Logger) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Low stock scan worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Scanning for low stock items...")
			// TODO: Implement low stock scan logic
		}
	}
}

// startDailyInsightsWorker generates daily insights
func startDailyInsightsWorker(ctx context.Context, logger zerolog.Logger) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Daily insights worker stopped")
			return
		case <-ticker.C:
			logger.Info().Msg("Generating daily insights...")
			// TODO: Implement daily insights generation
		}
	}
}
