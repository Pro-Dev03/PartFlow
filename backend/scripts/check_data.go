//go:build ignore

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

	fmt.Println("🔍 Checking database content...")

	tables := []string{
		"users",
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
	}

	for _, table := range tables {
		count := getCount(db, table)
		fmt.Printf("📊 %s: %d records\n", table, count)
	}
}

func getCount(db *sql.DB, table string) int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
