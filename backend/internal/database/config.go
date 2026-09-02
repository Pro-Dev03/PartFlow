package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns optimized default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxOpenConns:    25,              // Reduced for cloud database
		MaxIdleConns:    10,              // Reduced for cloud database
		ConnMaxLifetime: 5 * time.Minute, // Shorter lifetime for cloud
		ConnMaxIdleTime: 1 * time.Minute, // Shorter idle time for cloud
	}
}

// Connect creates a new database connection with optimized settings
func Connect(dbURL string) (*sqlx.DB, error) {
	config := DefaultConfig()

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
