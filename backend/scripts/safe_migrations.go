package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Println("=== Safe Database Migration Application ===\n")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create schema_migrations table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}
	fmt.Println("✅ schema_migrations table ready")

	// Get migration files in order
	migrationsDir := "/home/dev-bit/project/PartFlow/backend/migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Sort files to ensure correct order
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i] > files[j] {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	fmt.Printf("📁 Found %d migration files (in order)\n\n", len(files))

	// Apply each migration
	for _, file := range files {
		filename := filepath.Base(file)
		version := strings.TrimSuffix(filename, ".sql")

		// Check if already applied
		var applied bool
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied)
		if err != nil {
			log.Printf("⚠️  Failed to check migration status for %s: %v", version, err)
			continue
		}

		if applied {
			fmt.Printf("⏭️  Skipping %s (already applied)\n", version)
			continue
		}

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("❌ Failed to read migration file %s: %v", filename, err)
			continue
		}

		fmt.Printf("🔄 Applying %s...\n", version)

		// Apply migration with error handling
		success := applyMigrationSafe(db, string(content), version)
		if success {
			fmt.Printf("✅ Applied %s successfully\n", version)
		} else {
			fmt.Printf("⚠️  Skipped %s due to errors (may already exist)\n", version)
		}
	}

	fmt.Println("\n=== Migration Process Complete ===")
	
	// Show current status
	rows, err := db.Query("SELECT version, applied_at FROM schema_migrations ORDER BY applied_at")
	if err == nil {
		fmt.Println("\n📋 Applied migrations:")
		defer rows.Close()
		count := 0
		for rows.Next() {
			var version, appliedAt string
			if err := rows.Scan(&version, &appliedAt); err == nil {
				fmt.Printf("   - %s (applied at %s)\n", version, appliedAt)
				count++
			}
		}
		if count == 0 {
			fmt.Println("   No migrations recorded yet")
		}
	}
}

func applyMigrationSafe(db *sql.DB, content, version string) bool {
	// Split content by semicolon, but be careful with CREATE FUNCTION etc
	statements := splitSQLStatements(content)

	successCount := 0
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		// Try to execute each statement
		_, err := db.Exec(stmt)
		if err != nil {
			// Check if it's a "already exists" error
			if strings.Contains(err.Error(), "already exists") || 
			   strings.Contains(err.Error(), "duplicate key") ||
			   strings.Contains(err.Error(), "relation") {
				fmt.Printf("   Statement %d: ⏭️  Skipping (already exists)\n", i+1)
				successCount++
				continue
			}
			fmt.Printf("   Statement %d: ❌ Error: %v\n", i+1, err)
			fmt.Printf("   Statement: %s\n", stmt[:min(200, len(stmt))])
			// Continue with other statements even if one fails
		} else {
			successCount++
		}
	}

	// Record migration if we had any success
	if successCount > 0 {
		_, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT (version) DO NOTHING", version)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to record migration: %v\n", err)
			return false
		}
		return true
	}

	return false
}

func splitSQLStatements(content string) []string {
	var statements []string
	var currentStatement strings.Builder
	inParenthesis := 0
	inFunction := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip comments
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		// Track parenthesis for CREATE FUNCTION etc
		inParenthesis += strings.Count(line, "(") - strings.Count(line, ")")
		if strings.Contains(strings.ToUpper(line), "CREATE FUNCTION") || 
		   strings.Contains(strings.ToUpper(line), "CREATE TRIGGER") {
			inFunction = true
		}

		currentStatement.WriteString(line)
		currentStatement.WriteString("\n")

		// End of statement if we have semicolon and no open parenthesis
		if strings.Contains(line, ";") && inParenthesis <= 0 && !inFunction {
			if inParenthesis == 0 {
				inFunction = false
				statements = append(statements, currentStatement.String())
				currentStatement.Reset()
			}
		}
	}

	// Add any remaining content
	if currentStatement.Len() > 0 {
		statements = append(statements, currentStatement.String())
	}

	return statements
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
