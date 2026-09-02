package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/config"
	_ "modernc.org/sqlite"
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

	_ = os.Setenv("DATABASE_URL", databaseURL)

	var driver string
	var connectURL string
	if strings.HasPrefix(databaseURL, "sqlite://") {
		driver = "sqlite"
		connectURL = strings.TrimPrefix(databaseURL, "sqlite://")
		if connectURL == "" {
			return fmt.Errorf("sqlite database path is empty")
		}
		if !strings.Contains(connectURL, "?_pragma=") {
			connectURL += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"
		}
		log.Println("Using SQLite database for local mode:", connectURL)
	} else {
		driver = "pgx"
		connectURL = databaseURL
		if !strings.Contains(connectURL, "sslmode") {
			connectURL += "?sslmode=require"
		}
		log.Println("Using PostgreSQL database for cloud mode")
	}

	if driver == "pgx" {
		pgxConfig, parseErr := pgx.ParseConfig(connectURL)
		if parseErr != nil {
			return fmt.Errorf("failed to parse database URL: %w", parseErr)
		}
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		DB = sqlx.NewDb(stdlib.OpenDB(*pgxConfig), driver)
		err = DB.Ping()
	} else {
		DB, err = sqlx.Connect(driver, connectURL)
	}
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

	if err := ensureRequiredSchema(DB); err != nil {
		return fmt.Errorf("failed to ensure required schema: %w", err)
	}

	log.Println("Database connection established successfully using", driver)
	return nil
}

func ensureRequiredSchema(db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	if !strings.EqualFold(db.DriverName(), "sqlite") {
		if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS part_types (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				name_ar VARCHAR(255) NOT NULL,
				name_en VARCHAR(255) NOT NULL,
				icon VARCHAR(100),
				color VARCHAR(7),
				is_active BOOLEAN DEFAULT true,
				sort_order INTEGER DEFAULT 0,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			);
		`); err != nil {
			return err
		}

		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS held_sales (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				user_id UUID NOT NULL,
				items JSONB NOT NULL,
				created_at TIMESTAMPTZ DEFAULT NOW()
			);
		`); err != nil {
			return err
		}
		if ok, err := tableExists(db, "inventory_items"); err != nil {
			return err
		} else if ok {
			if _, err := db.Exec(`ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS part_type_id UUID REFERENCES part_types(id) ON DELETE SET NULL;`); err != nil {
				return err
			}
		}
		return nil
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS part_types (
			id TEXT PRIMARY KEY,
			name_ar TEXT NOT NULL,
			name_en TEXT NOT NULL,
			icon TEXT,
			color TEXT,
			is_active INTEGER DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			id TEXT PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			value TEXT,
			value_type TEXT DEFAULT 'string',
			category TEXT DEFAULT 'general',
			description TEXT,
			is_public INTEGER DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
		VALUES
			('setting-tax-rate', 'tax_rate', '0', 'number', 'financial', 'Tax rate', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			('setting-store-name', 'store_name', 'PartFlow Store', 'string', 'general', 'Store name', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			('setting-currency', 'currency', 'ILS', 'string', 'general', 'Currency', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS held_sales (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			items TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	// localdb.Open applies the complete local schema later, but this table is
	// required by startup checks that run before localdb.Open.
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS inventory_items (
			id TEXT PRIMARY KEY,
			product_id TEXT NOT NULL,
			item_code TEXT NOT NULL UNIQUE,
			barcode TEXT,
			serial_number TEXT,
			condition TEXT,
			grade TEXT,
			purchase_cost REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'AVAILABLE',
			location_id TEXT,
			supplier_id TEXT,
			purchase_date TEXT,
			sold_at TEXT,
			notes TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`); err != nil {
		return err
	}

	if ok, err := columnExists(db, "inventory_items", "part_type_id"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE inventory_items ADD COLUMN part_type_id TEXT;`); err != nil {
			return err
		}
	}

	return nil
}

func columnExists(db *sqlx.DB, tableName, columnName string) (bool, error) {
	if strings.EqualFold(db.DriverName(), "postgres") || strings.EqualFold(db.DriverName(), "pgx") {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = $1 AND column_name = $2
			)
		`
		if err := db.Get(&exists, query, tableName, columnName); err != nil {
			return false, err
		}
		return exists, nil
	}

	var exists bool
	query := `SELECT EXISTS (
		SELECT 1 FROM PRAGMA_TABLE_INFO(?) WHERE name = ?
	)`
	if err := db.Get(&exists, query, tableName, columnName); err != nil {
		return false, err
	}
	return exists, nil
}

func tableExists(db *sqlx.DB, tableName string) (bool, error) {
	if strings.EqualFold(db.DriverName(), "postgres") || strings.EqualFold(db.DriverName(), "pgx") {
		var exists bool
		query := `SELECT EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = $1
		)`
		if err := db.Get(&exists, query, tableName); err != nil {
			return false, err
		}
		return exists, nil
	}

	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)`
	if err := db.Get(&exists, query, tableName); err != nil {
		return false, err
	}
	return exists, nil
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
