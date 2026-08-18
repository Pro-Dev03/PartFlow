package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Read DATABASE_URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Println("=== Verifying Real Database Connection ===")
	fmt.Printf("Database URL: %s\n\n", databaseURL)

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✅ Database connection successful!")

	// Get database version
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}
	fmt.Printf("📊 Database Version: %s\n", version[:50]+"...")

	// Get current database
	var dbName string
	err = db.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		log.Fatalf("Failed to get database name: %v", err)
	}
	fmt.Printf("🗄️  Current Database: %s\n", dbName)

	// Get current user
	var currentUser string
	err = db.QueryRow("SELECT current_user").Scan(&currentUser)
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}
	fmt.Printf("👤 Current User: %s\n", currentUser)

	// Check if tables exist
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err != nil {
		log.Fatalf("Failed to count tables: %v", err)
	}
	fmt.Printf("📋 Tables in public schema: %d\n", tableCount)

	// List some tables if they exist
	if tableCount > 0 {
		fmt.Println("\n📝 Table names:")
		rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' LIMIT 10")
		if err != nil {
			log.Printf("Failed to list tables: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err == nil {
					fmt.Printf("   - %s\n", tableName)
				}
			}
		}
	}

	// Get connection info
	var clientIP string
	err = db.QueryRow("SELECT inet_server_addr()").Scan(&clientIP)
	if err != nil {
		log.Printf("Failed to get server address: %v", err)
	} else {
		fmt.Printf("🌐 Server IP: %s\n", clientIP)
	}

	fmt.Println("\n=== ✅ VERIFICATION COMPLETE ===")
	fmt.Println("This is a REAL connection to your Supabase PostgreSQL database!")
	fmt.Println("You can now read and write actual data to your database.")
}
