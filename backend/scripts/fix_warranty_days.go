//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	godotenv.Load()

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

	// Add warranty_days column to products table
	fmt.Println("🔄 Adding warranty_days column to products table...")
	_, err = db.Exec(`
		ALTER TABLE products 
		ADD COLUMN IF NOT EXISTS warranty_days INTEGER DEFAULT 0
	`)
	if err != nil {
		log.Fatalf("❌ Failed to add warranty_days to products: %v", err)
	}
	fmt.Println("✅ warranty_days column added to products table")

	// Add warranty_days column to part_types table
	fmt.Println("🔄 Adding warranty_days column to part_types table...")
	_, err = db.Exec(`
		ALTER TABLE part_types 
		ADD COLUMN IF NOT EXISTS warranty_days INTEGER DEFAULT 0
	`)
	if err != nil {
		log.Fatalf("❌ Failed to add warranty_days to part_types: %v", err)
	}
	fmt.Println("✅ warranty_days column added to part_types table")

	fmt.Println("\n🎉 Database schema updated successfully!")
}