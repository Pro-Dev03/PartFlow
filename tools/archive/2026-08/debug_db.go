package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "C:\\Users\\Administrator\\AppData\\Roaming\\PartFlow\\data\\partflow.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Check if users table exists
	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'")
	var tableName string
	err = row.Scan(&tableName)
	if err != nil {
		fmt.Println("Users table does not exist")
		return
	}

	// List all users
	rows, err := db.Query("SELECT id, email, first_name, last_name, is_active, password_hash FROM users LIMIT 10")
	if err != nil {
		log.Fatal("Failed to query users:", err)
	}
	defer rows.Close()

	fmt.Println("Users in database:")
	for rows.Next() {
		var id, email, firstName, lastName, passwordHash string
		var isActive bool
		err := rows.Scan(&id, &email, &firstName, &lastName, &isActive, &passwordHash)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		fmt.Printf("ID: %s, Email: %s, Name: %s %s, Active: %v, Has PW: %v\n", 
			id, email, firstName, lastName, isActive, len(passwordHash) > 0)
	}

	// Check tables
	tablesQuery := "SELECT name FROM sqlite_master WHERE type='table'"
	tableRows, err := db.Query(tablesQuery)
	if err != nil {
		log.Fatal("Failed to query tables:", err)
	}
	defer tableRows.Close()

	fmt.Println("\nTables in database:")
	for tableRows.Next() {
		var name string
		err := tableRows.Scan(&name)
		if err != nil {
			log.Fatal("Failed to scan table name:", err)
		}
		fmt.Println("  -", name)
	}
}
