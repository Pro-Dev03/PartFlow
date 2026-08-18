package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("=== PartFlow Database Connection Fix Tool ===\n")

	// Test network connectivity first
	fmt.Println("Step 1: Testing network connectivity...")
	testNetworkConnectivity()

	// Get the base database URL
	baseURL := os.Getenv("DATABASE_URL")
	if baseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Clean the URL
	baseURL = strings.TrimSuffix(baseURL, "?sslmode=require")
	baseURL = strings.TrimSuffix(baseURL, "?sslmode=disable")
	baseURL = strings.TrimSuffix(baseURL, "&sslmode=require")
	baseURL = strings.TrimSuffix(baseURL, "&sslmode=disable")

	fmt.Printf("\nStep 2: Testing different connection strategies...\n")
	fmt.Printf("Base URL: %s\n\n", baseURL)

	// Strategy 1: Try direct connection (non-pooler)
	fmt.Println("Strategy 1: Direct connection (non-pooler)")
	directURL := strings.Replace(baseURL, "pooler.supabase.com", "supabase.co", 1)
	directURL = strings.Replace(directURL, "aws-0-ap-south-1.", "db.", 1)
	if testConnection(directURL, "require", "Direct Connection") {
		updateEnvFile(directURL + "?sslmode=require")
		return
	}

	// Strategy 2: Try pooler with different SSL modes
	fmt.Println("\nStrategy 2: Pooler with SSL require")
	if testConnection(baseURL, "require", "Pooler SSL Require") {
		updateEnvFile(baseURL + "?sslmode=require")
		return
	}

	fmt.Println("\nStrategy 3: Pooler with SSL disable")
	if testConnection(baseURL, "disable", "Pooler SSL Disable") {
		updateEnvFile(baseURL + "?sslmode=disable")
		return
	}

	fmt.Println("\nStrategy 4: Pooler with SSL verify-full")
	if testConnection(baseURL, "verify-full", "Pooler SSL Verify-Full") {
		updateEnvFile(baseURL + "?sslmode=verify-full")
		return
	}

	// Strategy 5: Try Session pooler instead of Transaction pooler
	fmt.Println("\nStrategy 5: Session Pooler")
	sessionURL := strings.Replace(baseURL, "pooler.supabase.com:5432", "pooler.supabase.com:6543", 1)
	if testConnection(sessionURL, "require", "Session Pooler") {
		updateEnvFile(sessionURL + "?sslmode=require")
		return
	}

	// Strategy 6: Try with connection timeout and retry
	fmt.Println("\nStrategy 6: Connection with timeout and retry")
	timeoutURL := baseURL + "?connect_timeout=30&sslmode=require"
	if testConnectionWithRetry(timeoutURL, 5, 3*time.Second) {
		updateEnvFile(timeoutURL)
		return
	}

	fmt.Println("\n=== ALL STRATEGIES FAILED ===")
	fmt.Println("\nThe issue appears to be network-level or Supabase-side:")
	fmt.Println("1. Check if your Supabase project is active (not paused)")
	fmt.Println("2. Verify your IP is not blocked by Supabase")
	fmt.Println("3. Check if you're behind a corporate firewall/proxy")
	fmt.Println("4. Try connecting from a different network")
	fmt.Println("5. Contact Supabase support if the issue persists")
	fmt.Println("\nManual check:")
	fmt.Println("   - Go to Supabase Dashboard → Settings → Database")
	fmt.Println("   - Verify the connection string is correct")
	fmt.Println("   - Check Connection Pooling settings")
	fmt.Println("   - Review any IP restrictions")
}

func testNetworkConnectivity() {
	hosts := []string{
		"aws-0-ap-south-1.pooler.supabase.com",
		"db.auwushpeqokeglzklntv.supabase.co",
		"auwushpeqokeglzklntv.supabase.co",
	}

	for _, host := range hosts {
		fmt.Printf("  Testing %s... ", host)
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		resolver := net.Resolver{}
		_, err := resolver.LookupHost(ctx, host)
		if err != nil {
			fmt.Printf("❌ DNS lookup failed: %v\n", err)
		} else {
			fmt.Printf("✅ DNS OK\n")
			
			// Try TCP connection
			conn, err := net.DialTimeout("tcp", host+":5432", 5*time.Second)
			if err != nil {
				fmt.Printf("    TCP connection: ❌ %v\n", err)
			} else {
				fmt.Printf("    TCP connection: ✅ OK\n")
				conn.Close()
			}
		}
	}
}

func testConnection(databaseURL, sslMode, name string) bool {
	fmt.Printf("  Testing %s... ", name)
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", databaseURL+"?sslmode="+sslMode)
	if err != nil {
		fmt.Printf("❌ Open failed: %v\n", err)
		return false
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Try ping with retry
	for i := 1; i <= 3; i++ {
		err := db.PingContext(ctx)
		if err == nil {
			fmt.Printf("✅ Success!\n")
			
			// Verify with a query
			var version string
			err = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
			if err == nil {
				fmt.Printf("    Database query successful\n")
				return true
			}
		}
		
		if i < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Printf("❌ Failed: %v\n", err)
	return false
}

func testConnectionWithRetry(databaseURL string, maxAttempts int, delay time.Duration) bool {
	fmt.Printf("  Testing with %d attempts (delay %v)... ", maxAttempts, delay)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxAttempts)*delay+30*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		fmt.Printf("❌ Open failed: %v\n", err)
		return false
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	for i := 1; i <= maxAttempts; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("❌ Timeout\n")
			return false
		default:
			err := db.PingContext(ctx)
			if err == nil {
				fmt.Printf("✅ Success on attempt %d!\n", i)
				return true
			}
			fmt.Printf("    Attempt %d failed: %v\n", i, err)
			if i < maxAttempts {
				time.Sleep(delay)
			}
		}
	}

	return false
}

func updateEnvFile(newURL string) {
	fmt.Println("\n=== SUCCESS! ===")
	fmt.Printf("Working configuration found: %s\n", newURL)
	
	// Update the .env file
	envPath := "/home/dev-bit/project/PartFlow/backend/.env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		fmt.Printf("⚠️  Could not read .env file: %v\n", err)
		fmt.Printf("Please manually update DATABASE_URL to: %s\n", newURL)
		return
	}

	envContent := string(content)
	lines := strings.Split(envContent, "\n")
	
	for i, line := range lines {
		if strings.HasPrefix(line, "DATABASE_URL=") {
			lines[i] = "DATABASE_URL=" + newURL
			break
		}
	}

	newContent := strings.Join(lines, "\n")
	err = os.WriteFile(envPath, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("⚠️  Could not write .env file: %v\n", err)
		fmt.Printf("Please manually update DATABASE_URL to: %s\n", newURL)
		return
	}

	fmt.Println("✅ .env file updated successfully!")
	fmt.Println("You can now restart your application.")
}
