package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type ConnectionTest struct {
	Name        string
	DatabaseURL string
	SSLMode     string
	Timeout     time.Duration
}

func main() {
	// Read the DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Println("=== Database Connection Diagnostic Tool ===")
	fmt.Printf("Base Database URL: %s\n\n", databaseURL)

	// Test different connection configurations
	tests := []ConnectionTest{
		{
			Name:        "Direct connection with SSL require",
			DatabaseURL: databaseURL + "?sslmode=require",
			SSLMode:     "require",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "Direct connection with SSL disable",
			DatabaseURL: databaseURL + "?sslmode=disable",
			SSLMode:     "disable",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "Direct connection with SSL allow",
			DatabaseURL: databaseURL + "?sslmode=allow",
			SSLMode:     "allow",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "Direct connection with SSL prefer",
			DatabaseURL: databaseURL + "?sslmode=prefer",
			SSLMode:     "prefer",
			Timeout:     10 * time.Second,
		},
		{
			Name:        "Connection with connect_timeout",
			DatabaseURL: databaseURL + "?sslmode=require&connect_timeout=10",
			SSLMode:     "require",
			Timeout:     15 * time.Second,
		},
		{
			Name:        "Connection with timeout and SSL disable",
			DatabaseURL: databaseURL + "?sslmode=disable&connect_timeout=10",
			SSLMode:     "disable",
			Timeout:     15 * time.Second,
		},
	}

	successfulConfig := ""
	var successfulDB *sql.DB

	for i, test := range tests {
		fmt.Printf("[%d/%d] Testing: %s\n", i+1, len(tests), test.Name)
		fmt.Printf("  URL: %s\n", test.DatabaseURL)

		ctx, cancel := context.WithTimeout(context.Background(), test.Timeout)
		defer cancel()

		db, err := sql.Open("postgres", test.DatabaseURL)
		if err != nil {
			fmt.Printf("  ❌ Failed to open connection: %v\n\n", err)
			continue
		}
		defer db.Close()

		// Configure connection pool
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(2)
		db.SetConnMaxLifetime(5 * time.Minute)

		// Test connection with retry logic
		var connErr error
		for attempt := 1; attempt <= 3; attempt++ {
			fmt.Printf("  Attempt %d/3... ", attempt)
			
			select {
			case <-ctx.Done():
				connErr = fmt.Errorf("timeout after %v", test.Timeout)
				fmt.Printf("❌ %v\n", connErr)
				break
			default:
				connErr = db.PingContext(ctx)
				if connErr == nil {
					fmt.Printf("✅ Success!\n")
					break
				}
				fmt.Printf("❌ %v\n", connErr)
				if attempt < 3 {
					time.Sleep(2 * time.Second)
				}
			}
		}

		if connErr == nil {
			fmt.Printf("  ✅ Connection successful!\n")
			
			// Test a simple query
			var result string
			err := db.QueryRowContext(ctx, "SELECT current_database()").Scan(&result)
			if err == nil {
				fmt.Printf("  ✅ Query successful! Database: %s\n", result)
			} else {
				fmt.Printf("  ⚠️  Query failed: %v\n", err)
			}

			successfulConfig = test.DatabaseURL
			successfulDB = db
			fmt.Printf("\n=== SUCCESSFUL CONFIGURATION FOUND ===\n")
			fmt.Printf("Name: %s\n", test.Name)
			fmt.Printf("Database URL: %s\n", test.DatabaseURL)
			fmt.Printf("SSL Mode: %s\n\n", test.SSLMode)
			break
		} else {
			fmt.Printf("  ❌ All attempts failed\n\n")
		}
	}

	if successfulConfig != "" {
		fmt.Println("=== RECOMMENDED ACTION ===")
		fmt.Printf("Update your .env file with this DATABASE_URL:\n")
		fmt.Printf("DATABASE_URL=%s\n", successfulConfig)
		
		// Keep the successful connection open for a moment to verify stability
		fmt.Println("\n=== Testing connection stability ===")
		for i := 1; i <= 5; i++ {
			time.Sleep(1 * time.Second)
			err := successfulDB.Ping()
			if err != nil {
				fmt.Printf("  Check %d/5: ❌ %v\n", i, err)
			} else {
				fmt.Printf("  Check %d/5: ✅ OK\n", i)
			}
		}
		
		successfulDB.Close()
		os.Exit(0)
	} else {
		fmt.Println("=== NO SUCCESSFUL CONFIGURATION FOUND ===")
		fmt.Println("\nPossible issues:")
		fmt.Println("1. Network connectivity issues (firewall, proxy)")
		fmt.Println("2. Supabase project is paused or suspended")
		fmt.Println("3. Incorrect credentials or database URL")
		fmt.Println("4. IP whitelist restrictions in Supabase")
		fmt.Println("5. Database server is not responding")
		fmt.Println("\nRecommended actions:")
		fmt.Println("1. Check Supabase dashboard to verify project status")
		fmt.Println("2. Verify the database URL in Supabase settings")
		fmt.Println("3. Check network connectivity: ping aws-0-ap-south-1.pooler.supabase.com")
		fmt.Println("4. Try using Session Pooler instead of Transaction Pooler")
		fmt.Println("5. Check if your IP is whitelisted in Supabase")
		os.Exit(1)
	}
}
