package sync

import (
	"context"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
	"github.com/partflow/smart-store/pkg/logger"
)

const defaultOfflineSyncInterval = 30 * time.Second

func StartOfflineSyncWorker(ctx context.Context, postgresDB *sqlx.DB) {
	if postgresDB == nil {
		logger.Warn("Offline sync worker skipped because PostgreSQL is not initialized", nil)
		return
	}

	for {
		if err := processOfflineSyncCycle(postgresDB); err != nil {
			logger.Error("Offline sync cycle failed", err, nil)
		}

		select {
		case <-ctx.Done():
			logger.Info("Offline sync worker stopped", nil)
			return
		case <-time.After(defaultOfflineSyncInterval):
		}
	}
}

func processOfflineSyncCycle(postgresDB *sqlx.DB) error {
	sqliteDB, err := localdb.Open()
	if err != nil {
		return err
	}
	defer sqliteDB.DB.Close()

	mode, err := localdb.GetMetadata(sqliteDB.DB, "operating_mode")
	if err != nil {
		return err
	}
	if strings.TrimSpace(mode) == "" {
		mode = "online"
	}
	if mode != "online" {
		return nil
	}

	pendingCount, err := localdb.GetPendingSyncCount(sqliteDB.DB)
	if err != nil {
		return err
	}
	if pendingCount == 0 {
		return nil
	}

	result, err := ProcessPendingOfflineQueue(postgresDB, sqliteDB.DB, 50)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	if len(result.Errors) > 0 {
		logger.Warn("Some offline sync items failed", map[string]interface{}{
			"processed": result.Processed,
			"failed":    result.Failed,
			"errors":    result.Errors,
		})
		return nil
	}

	logger.Info("Offline sync queue processed successfully", map[string]interface{}{
		"processed": result.Processed,
		"failed":    result.Failed,
	})
	return nil
}
