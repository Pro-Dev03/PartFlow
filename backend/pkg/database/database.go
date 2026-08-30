package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/partflow/smart-store/pkg/config"
)

var DB *sqlx.DB

// Initialize initializes the database connection
func Initialize() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	databaseURL := cfg.DatabaseURL
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// Add SSL mode to connection string for Supabase (only if not already specified)
	if !strings.Contains(databaseURL, "sslmode") {
		databaseURL += "?sslmode=require"
	}

	_ = os.Setenv("DATABASE_URL", databaseURL)

	DB, err = sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(1 * time.Minute)

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *sqlx.DB {
	return DB
}

// Health checks the database connection
func Health() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return DB.Ping()
}
