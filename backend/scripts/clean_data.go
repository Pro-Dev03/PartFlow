package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("🧹 Cleaning database...")

	// Remove all users except the owner
	fmt.Println("🗑️  Cleaning users (keeping owner only)...")
	_, err = db.Exec("DELETE FROM users WHERE email != 'owner@partflow.com'")
	if err != nil {
		log.Printf("⚠️  Error cleaning users: %v", err)
	} else {
		fmt.Println("✅ Users cleaned")
	}

	// Clean all other tables
	tables := []string{
		"categories",
		"products",
		"customers",
		"suppliers",
		"inventory_items",
		"sales",
		"purchases",
		"expenses",
		"debts",
		"warranties",
		"settings",
		"notifications",
		"audit_logs",
		"return_items",
		"returns",
		"trade_ins",
		"inspections",
		"reports",
		"barcodes",
	}

	for _, table := range tables {
		fmt.Printf("🗑️  Cleaning %s...\n", table)
		query := fmt.Sprintf("DELETE FROM %s", table)
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("⚠️  Error cleaning %s: %v", table, err)
		} else {
			fmt.Printf("✅ %s cleaned\n", table)
		}
	}

	fmt.Println("\n🎉 Database cleaned successfully!")
	fmt.Println("📊 Remaining users: 1 (owner@partflow.com)")
}
