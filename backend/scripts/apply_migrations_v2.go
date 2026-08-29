//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL environment variable is not set")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database successfully")

	// Create migrations table if not exists
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("❌ Failed to create migrations table: %v", err)
	}

	// Get migration files
	migrationsDir := "migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("❌ Failed to read migrations directory: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("❌ No migration files found")
	}

	// Sort files by numeric prefix (001, 002, etc.)
	sort.Slice(files, func(i, j int) bool {
		filenameI := filepath.Base(files[i])
		filenameJ := filepath.Base(files[j])
		return filenameI < filenameJ
	})

	fmt.Printf("📋 Found %d migration files\n", len(files))

	// Get applied migrations
	appliedMigrations := getAppliedMigrations(db)

	// Apply migrations with retry logic for dependencies
	var failedMigrations []string
	var migrationFiles []string
	
	for _, file := range files {
		filename := filepath.Base(file)
		
		// Skip if already applied
		if isApplied(filename, appliedMigrations) {
			fmt.Printf("⏭️  Skipping %s (already applied)\n", filename)
			continue
		}

		migrationFiles = append(migrationFiles, file)
	}

	// First pass: try to apply all migrations
	for _, file := range migrationFiles {
		filename := filepath.Base(file)
		
		fmt.Printf("🔄 Applying %s...\n", filename)
		
		if err := applyMigration(db, file, filename); err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				fmt.Printf("⚠️  Deferred %s (missing dependency)\n", filename)
				failedMigrations = append(failedMigrations, file)
			} else {
				log.Fatalf("❌ Failed to apply %s: %v", filename, err)
			}
		} else {
			fmt.Printf("✅ Applied %s successfully\n", filename)
		}
	}

	// Second pass: retry failed migrations (dependencies should now exist)
	maxRetries := 3
	for retry := 0; retry < maxRetries && len(failedMigrations) > 0; retry++ {
		if retry > 0 {
			fmt.Printf("\n🔄 Retry pass %d for %d deferred migrations...\n", retry+1, len(failedMigrations))
		}
		
		var stillFailed []string
		for _, file := range failedMigrations {
			filename := filepath.Base(file)
			
			fmt.Printf("🔄 Applying %s...\n", filename)
			
			if err := applyMigration(db, file, filename); err != nil {
				if strings.Contains(err.Error(), "does not exist") {
					fmt.Printf("⚠️  Still deferred %s (missing dependency)\n", filename)
					stillFailed = append(stillFailed, file)
				} else {
					log.Fatalf("❌ Failed to apply %s: %v", filename, err)
				}
			} else {
				fmt.Printf("✅ Applied %s successfully\n", filename)
			}
		}
		
		failedMigrations = stillFailed
	}

	if len(failedMigrations) > 0 {
		fmt.Printf("\n⚠️  Warning: %d migrations could not be applied due to missing dependencies:\n", len(failedMigrations))
		for _, file := range failedMigrations {
			fmt.Printf("   - %s\n", filepath.Base(file))
		}
		fmt.Println("💡 Hint: These may be old migrations that depend on tables removed in newer migrations.")
	}

	// Show final status
	showMigrationStatus(db)
	fmt.Println("\n🎉 All migrations completed successfully!")
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) map[string]bool {
	query := "SELECT version FROM schema_migrations"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("⚠️  Failed to get applied migrations: %v", err)
		return make(map[string]bool)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			continue
		}
		applied[version] = true
	}

	return applied
}

func isApplied(filename string, applied map[string]bool) bool {
	return applied[filename]
}

func applyMigration(db *sql.DB, filepath, filename string) error {
	// Read migration file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute entire file as single statement
	if _, err := tx.Exec(string(content)); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	if err := recordMigration(tx, filename); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func recordMigration(tx *sql.Tx, filename string) error {
	query := `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, NOW())`
	_, err := tx.Exec(query, filename)
	return err
}

func showMigrationStatus(db *sql.DB) {
	query := "SELECT version, applied_at FROM schema_migrations ORDER BY applied_at"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("⚠️  Failed to get migration status: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n📋 Current migration status:")
	fmt.Println("─" + strings.Repeat("─", 50))

	for rows.Next() {
		var version, appliedAt string
		if err := rows.Scan(&version, &appliedAt); err != nil {
			continue
		}
		fmt.Printf("   ✅ %s (applied: %s)\n", version, appliedAt)
	}

	fmt.Println("─" + strings.Repeat("─", 50))
}
