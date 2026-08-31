package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := `C:\Users\Administrator\AppData\Roaming\PartFlow\data\partflow.db`
	fmt.Printf("Opening database: %s\n", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("✅ Connected to database\n")

	// Check for owner user
	fmt.Println("Checking for owner user...")
	var userID, email, firstName, lastName, passwordHash string
	var isActive int
	err = db.QueryRow(`
		SELECT id, email, first_name, last_name, password_hash, is_active 
		FROM users 
		WHERE email = 'owner@partflow.com'
	`).Scan(&userID, &email, &firstName, &lastName, &passwordHash, &isActive)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ No owner user found in database")
			// List all users
			fmt.Println("\nAll users in database:")
			rows, err := db.Query("SELECT id, email, first_name, last_name FROM users LIMIT 10")
			if err != nil {
				log.Fatal("Failed to query users:", err)
			}
			defer rows.Close()

			for rows.Next() {
				var id, email, fn, ln string
				if err := rows.Scan(&id, &email, &fn, &ln); err != nil {
					log.Fatal("Failed to scan row:", err)
				}
				fmt.Printf("  - %s (%s %s)\n", email, fn, ln)
			}
		} else {
			log.Fatal("Failed to query user:", err)
		}
		return
	}

	fmt.Println("✅ Owner user found:")
	fmt.Printf("  ID: %s\n", userID)
	fmt.Printf("  Email: %s\n", email)
	fmt.Printf("  Name: %s %s\n", firstName, lastName)
	fmt.Printf("  Active: %v\n", isActive == 1)
	fmt.Printf("  Password Hash Length: %d bytes\n", len(passwordHash))
	fmt.Printf("  Password Hash (first 30 chars): %s...\n", passwordHash[:30])

	// Try verifying password with bcrypt
	fmt.Println("\n🔍 Password verification would happen on login")
	fmt.Println("Expected password: Owner123456")
}
