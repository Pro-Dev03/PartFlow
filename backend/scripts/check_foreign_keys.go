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
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check foreign key constraints
	fmt.Println("Checking foreign key constraints on sales table...")
	
	rows, err := db.Query(`
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
		WHERE tc.table_name = 'sales'
			AND tc.constraint_type = 'FOREIGN KEY'
	`)
	if err != nil {
		log.Fatalf("Error checking constraints: %v", err)
	}
	defer rows.Close()

	fmt.Println("Foreign Key Constraints on sales table:")
	for rows.Next() {
		var constraintName, columnName, foreignTable, foreignColumn string
		if err := rows.Scan(&constraintName, &columnName, &foreignTable, &foreignColumn); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("  %s: %s -> %s.%s\n", constraintName, columnName, foreignTable, foreignColumn)
	}

	// Check if user exists
	fmt.Println("\nChecking if user exists in users table...")
	var userExists bool
	err = db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM users WHERE id = '6deb0af1-136e-4aab-a46d-bb4799627321')
	`).Scan(&userExists)
	if err != nil {
		log.Printf("Error checking user: %v", err)
	} else {
		fmt.Printf("User 6deb0af1-136e-4aab-a46d-bb4799627321 exists: %v\n", userExists)
	}

	// Check all users
	fmt.Println("\nAll users in database:")
	userRows, err := db.Query(`SELECT id, email, first_name, last_name FROM users LIMIT 5`)
	if err != nil {
		log.Printf("Error getting users: %v", err)
	} else {
		defer userRows.Close()
		for userRows.Next() {
			var id, email, firstName, lastName string
			if err := userRows.Scan(&id, &email, &firstName, &lastName); err != nil {
				log.Printf("Error scanning user: %v", err)
				continue
			}
			fmt.Printf("  %s: %s (%s %s)\n", id, email, firstName, lastName)
		}
	}
}
