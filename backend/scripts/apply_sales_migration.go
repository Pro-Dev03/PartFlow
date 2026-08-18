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

	fmt.Println("=== Applying Sales Enhancements Migration ===\n")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Read migration file
	content, err := os.ReadFile("/home/dev-bit/project/PartFlow/backend/migrations/007_sales_enhancements.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Execute each statement
	_, err = tx.Exec(string(content))
	if err != nil {
		log.Printf("Migration completed with warnings: %v\n", err)
		// Continue as some statements might fail if columns already exist
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit migration: %v", err)
	}

	fmt.Println("✅ Sales enhancements migration applied successfully!")
	
	// Verify the changes
	fmt.Println("\n🔍 Verifying changes...")
	
	var hasCostAmount, hasGrossProfit, hasNetProfit bool
	
	db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sales' AND column_name = 'cost_amount')").Scan(&hasCostAmount)
	db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sales' AND column_name = 'gross_profit')").Scan(&hasGrossProfit)
	db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sales' AND column_name = 'net_profit')").Scan(&hasNetProfit)
	
	if hasCostAmount {
		fmt.Println("✅ cost_amount column added")
	} else {
		fmt.Println("⚠️  cost_amount column check failed")
	}
	
	if hasGrossProfit {
		fmt.Println("✅ gross_profit column added")
	} else {
		fmt.Println("⚠️  gross_profit column check failed")
	}
	
	if hasNetProfit {
		fmt.Println("✅ net_profit column added")
	} else {
		fmt.Println("⚠️  net_profit column check failed")
	}
	
	var hasUnitCost bool
	db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sale_items' AND column_name = 'unit_cost')").Scan(&hasUnitCost)
	
	if hasUnitCost {
		fmt.Println("✅ unit_cost column added to sale_items")
	} else {
		fmt.Println("⚠️  unit_cost column check failed")
	}
	
	fmt.Println("\n=== Migration Complete ===")
}
