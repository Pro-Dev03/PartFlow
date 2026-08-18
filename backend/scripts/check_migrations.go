package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Println("=== Checking Database Schema ===\n")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Check if schema_migrations table exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'schema_migrations'
		)
	`).Scan(&exists)
	if err != nil {
		log.Printf("Failed to check schema_migrations table: %v", err)
	} else if exists {
		fmt.Println("✅ schema_migrations table exists")
		
		// Get current migration version
		var version string
		err = db.QueryRow("SELECT version FROM schema_migrations LIMIT 1").Scan(&version)
		if err == nil {
			fmt.Printf("📌 Current migration version: %s\n", version)
		}
	} else {
		fmt.Println("❌ schema_migrations table does not exist")
		fmt.Println("⚠️  Migrations have not been tracked")
	}

	// Check all tables
	fmt.Println("\n📋 Current tables in database:")
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			tables = append(tables, tableName)
			fmt.Printf("   - %s\n", tableName)
		}
	}

	// Expected tables from migrations
	expectedTables := []string{
		"organizations", "roles", "users", "categories", "brands", "products",
		"inventory_items", "locations", "reservations", "customers", "sales",
		"sale_items", "suppliers", "purchases", "purchase_items", "expenses",
		"returns", "return_items", "warranties", "warranty_claims", "inspections",
		"reports", "notifications", "audit_logs", "permissions",
	}

	fmt.Println("\n🔍 Checking expected tables:")
	missingTables := []string{}
	for _, expected := range expectedTables {
		found := false
		for _, actual := range tables {
			if actual == expected {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("   ✅ %s\n", expected)
		} else {
			fmt.Printf("   ❌ %s (missing)\n", expected)
			missingTables = append(missingTables, expected)
		}
	}

	if len(missingTables) > 0 {
		fmt.Printf("\n⚠️  %d tables are missing. Need to run migrations.\n", len(missingTables))
	} else {
		fmt.Println("\n✅ All expected tables exist!")
	}

	// Check table structures
	fmt.Println("\n🔍 Checking table structures...")
	
	// Check organizations table
	fmt.Println("   organizations table:")
	checkTableStructure(db, "organizations", []string{"id", "name", "slug", "email", "created_at"})
	
	// Check users table
	fmt.Println("   users table:")
	checkTableStructure(db, "users", []string{"id", "email", "password_hash", "organization_id", "created_at"})
	
	// Check products table
	fmt.Println("   products table:")
	checkTableStructure(db, "products", []string{"id", "name", "sku", "barcode", "organization_id", "created_at"})
	
	// Check inventory_items table
	fmt.Println("   inventory_items table:")
	checkTableStructure(db, "inventory_items", []string{"id", "product_id", "barcode", "status", "organization_id", "created_at"})
}

func checkTableStructure(db *sql.DB, tableName string, expectedColumns []string) {
	rows, err := db.Query(`
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = $1 AND table_schema = 'public'
		ORDER BY ordinal_position
	`, tableName)
	if err != nil {
		fmt.Printf("      ❌ Failed to check columns: %v\n", err)
		return
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err == nil {
			columns = append(columns, colName)
		}
	}

	allFound := true
	for _, expected := range expectedColumns {
		found := false
		for _, actual := range columns {
			if actual == expected {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("      ✅ %s\n", expected)
		} else {
			fmt.Printf("      ❌ %s (missing)\n", expected)
			allFound = false
		}
	}

	if allFound {
		fmt.Printf("      ✅ All expected columns present\n")
	}
}
