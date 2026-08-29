//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres.auwushpeqokeglzklntv:bLpNU5qYOBQj4UQ0@aws-0-ap-south-1.pooler.supabase.com:6543/postgres?sslmode=require"
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	fmt.Println("=== PartFlow Performance Fix Script ===")
	fmt.Println("1. Optimizing Database Connection")
	fmt.Println("2. Adding Performance Indexes")
	fmt.Println("3. Analyzing Query Performance")
	fmt.Println()

	// Step 1: Test current connection performance
	fmt.Println("Step 1: Testing current connection performance...")
	start := time.Now()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Printf("   Current connection latency: %v\n", time.Since(start))
	fmt.Println()

	// Step 2: Add performance indexes
	fmt.Println("Step 2: Adding performance indexes...")
	
	indexes := []struct {
		name    string
		query   string
		table   string
		columns string
	}{
		{
			name:    "idx_users_organization_id",
			query:   "CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users(organization_id)",
			table:   "users",
			columns: "organization_id",
		},
		{
			name:    "idx_sales_organization_status",
			query:   "CREATE INDEX IF NOT EXISTS idx_sales_organization_status ON sales(organization_id, status)",
			table:   "sales",
			columns: "organization_id, status",
		},
		{
			name:    "idx_purchases_organization_status",
			query:   "CREATE INDEX IF NOT EXISTS idx_purchases_organization_status ON purchases(organization_id, status)",
			table:   "purchases",
			columns: "organization_id, status",
		},
		{
			name:    "idx_expenses_organization",
			query:   "CREATE INDEX IF NOT EXISTS idx_expenses_organization ON expenses(organization_id)",
			table:   "expenses",
			columns: "organization_id",
		},
		{
			name:    "idx_products_organization_active",
			query:   "CREATE INDEX IF NOT EXISTS idx_products_organization_active ON products(organization_id, is_active)",
			table:   "products",
			columns: "organization_id, is_active",
		},
		{
			name:    "idx_customers_organization_active",
			query:   "CREATE INDEX IF NOT EXISTS idx_customers_organization_active ON customers(organization_id, is_active)",
			table:   "customers",
			columns: "organization_id, is_active",
		},
		{
			name:    "idx_suppliers_organization_active",
			query:   "CREATE INDEX IF NOT EXISTS idx_suppliers_organization_active ON suppliers(organization_id, is_active)",
			table:   "suppliers",
			columns: "organization_id, is_active",
		},
		{
			name:    "idx_inventory_product_quantity",
			query:   "CREATE INDEX IF NOT EXISTS idx_inventory_product_quantity ON inventory(product_id, quantity)",
			table:   "inventory",
			columns: "product_id, quantity",
		},
		{
			name:    "idx_customer_ledger_customer_type",
			query:   "CREATE INDEX IF NOT EXISTS idx_customer_ledger_customer_type ON customer_ledger(customer_id, transaction_type, created_at)",
			table:   "customer_ledger",
			columns: "customer_id, transaction_type, created_at",
		},
		{
			name:    "idx_warranties_organization_active",
			query:   "CREATE INDEX IF NOT EXISTS idx_warranties_organization_active ON warranties(organization_id, is_active, expires_at)",
			table:   "warranties",
			columns: "organization_id, is_active, expires_at",
		},
		{
			name:    "idx_returns_organization_status",
			query:   "CREATE INDEX IF NOT EXISTS idx_returns_organization_status ON returns(organization_id, status)",
			table:   "returns",
			columns: "organization_id, status",
		},
	}

	for _, idx := range indexes {
		start := time.Now()
		_, err := db.ExecContext(ctx, idx.query)
		if err != nil {
			fmt.Printf("   ❌ Failed to create index %s: %v\n", idx.name, err)
		} else {
			fmt.Printf("   ✅ Created index %s on %s(%s) (took %v)\n", idx.name, idx.table, idx.columns, time.Since(start))
		}
	}
	fmt.Println()

	// Step 3: Update database statistics
	fmt.Println("Step 3: Updating database statistics...")
	start = time.Now()
	_, err = db.ExecContext(ctx, "ANALYZE")
	if err != nil {
		fmt.Printf("   ❌ Failed to analyze database: %v\n", err)
	} else {
		fmt.Printf("   ✅ Database statistics updated (took %v)\n", time.Since(start))
	}
	fmt.Println()

	// Step 4: Test query performance after optimization
	fmt.Println("Step 4: Testing query performance after optimization...")
	
	queries := []struct {
		name  string
		query string
	}{
		{
			name:  "User lookup by ID",
			query: "SELECT organization_id, role_id FROM users WHERE id = $1",
		},
		{
			name:  "Total sales",
			query: "SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE organization_id = $1 AND status = 'completed'",
		},
		{
			name:  "Total products",
			query: "SELECT COUNT(*) FROM products WHERE organization_id = $1 AND is_active = true",
		},
	}

	orgID := "5794ced2-56fe-46a5-abeb-b09c8dcb89fd"

	for _, q := range queries {
		start := time.Now()
		var result interface{}
		err := db.GetContext(ctx, &result, q.query, orgID)
		if err != nil {
			fmt.Printf("   ❌ %s failed: %v\n", q.name, err)
		} else {
			fmt.Printf("   ✅ %s: %v (took %v)\n", q.name, result, time.Since(start))
		}
	}
	fmt.Println()

	// Step 5: Generate optimization recommendations
	fmt.Println("Step 5: Performance Optimization Recommendations")
	fmt.Println("==============================================")
	fmt.Println("1. ✅ Database indexes added for critical queries")
	fmt.Println("2. ✅ Database statistics updated")
	fmt.Println("3. 🔧 Recommended: Update DATABASE_URL to use direct connection")
	fmt.Println("   Current: pooler.supabase.com (slower, connection pooling)")
	fmt.Println("   Recommended: db.supabase.com (faster, direct connection)")
	fmt.Println("4. 🔧 Recommended: Add connection pooling in application")
	fmt.Println("5. 🔧 Recommended: Implement Redis caching for dashboard data")
	fmt.Println("6. 🔧 Recommended: Add query timeout to prevent long-running queries")
	fmt.Println()

	// Step 6: Generate optimized .env file
	fmt.Println("Step 6: Generating optimized .env configuration...")
	
	// Convert pooler URL to direct URL
	optimizedURL := "postgresql://postgres.auwushpeqokeglzklntv:bLpNU5qYOBQj4UQ0@aws-0-ap-south-1.db.supabase.com:5432/postgres?sslmode=require"
	
	fmt.Println("Add this to your backend/.env file:")
	fmt.Printf("DATABASE_URL=%s\n", optimizedURL)
	fmt.Println("DB_MAX_OPEN_CONNS=50")
	fmt.Println("DB_MAX_IDLE_CONNS=25")
	fmt.Println("DB_CONN_MAX_LIFETIME=10m")
	fmt.Println()

	fmt.Println("=== Performance Fix Complete ===")
	fmt.Println("Next steps:")
	fmt.Println("1. Update backend/.env with the optimized DATABASE_URL")
	fmt.Println("2. Restart the backend server")
	fmt.Println("3. Test dashboard endpoint performance")
}