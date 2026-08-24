package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
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

	// Test organization ID
	orgID := uuid.MustParse("5794ced2-56fe-46a5-abeb-b09c8dcb89fd")

	fmt.Println("Testing Dashboard Queries...")
	fmt.Println("Organization ID:", orgID)
	fmt.Println()

	// Test 1: Total sales
	fmt.Println("1. Testing total sales query...")
	query := `SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE organization_id = $1 AND status = 'completed'`
	var totalSales float64
	err = db.GetContext(ctx, &totalSales, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Total Sales = %.2f\n", totalSales)
	}
	fmt.Println()

	// Test 2: Total purchases
	fmt.Println("2. Testing total purchases query...")
	query = `SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE organization_id = $1 AND status = 'received'`
	var totalPurchases float64
	err = db.GetContext(ctx, &totalPurchases, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Total Purchases = %.2f\n", totalPurchases)
	}
	fmt.Println()

	// Test 3: Total expenses
	fmt.Println("3. Testing total expenses query...")
	query = `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE organization_id = $1`
	var totalExpenses float64
	err = db.GetContext(ctx, &totalExpenses, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Total Expenses = %.2f\n", totalExpenses)
	}
	fmt.Println()

	// Test 4: Total products
	fmt.Println("4. Testing total products query...")
	query = `SELECT COUNT(*) FROM products WHERE organization_id = $1 AND is_active = true`
	var totalProducts int
	err = db.GetContext(ctx, &totalProducts, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Total Products = %d\n", totalProducts)
	}
	fmt.Println()

	// Test 5: Total customers
	fmt.Println("5. Testing total customers query...")
	query = `SELECT COUNT(*) FROM customers WHERE organization_id = $1 AND is_active = true`
	var totalCustomers int
	err = db.GetContext(ctx, &totalCustomers, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Total Customers = %d\n", totalCustomers)
	}
	fmt.Println()

	// Test 6: Total suppliers
	fmt.Println("6. Testing total suppliers query...")
	query = `SELECT COUNT(*) FROM suppliers WHERE organization_id = $1 AND is_active = true`
	var totalSuppliers int
	err = db.GetContext(ctx, &totalSuppliers, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Total Suppliers = %d\n", totalSuppliers)
	}
	fmt.Println()

	// Test 7: Pending orders
	fmt.Println("7. Testing pending orders query...")
	query = `SELECT COUNT(*) FROM sales WHERE organization_id = $1 AND status = 'pending'`
	var pendingOrders int
	err = db.GetContext(ctx, &pendingOrders, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Pending Orders = %d\n", pendingOrders)
	}
	fmt.Println()

	// Test 8: Low stock items
	fmt.Println("8. Testing low stock items query...")
	query = `
		SELECT COUNT(*) FROM products p
		WHERE p.organization_id = $1
		AND p.is_active = true
		AND p.min_stock_level > 0
		AND (SELECT COALESCE(SUM(quantity), 0) FROM inventory WHERE product_id = p.id) < p.min_stock_level
	`
	var lowStockItems int
	err = db.GetContext(ctx, &lowStockItems, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Low Stock Items = %d\n", lowStockItems)
	}
	fmt.Println()

	// Test 9: Overdue debts
	fmt.Println("9. Testing overdue debts query...")
	query = `
		SELECT COALESCE(SUM(current_balance), 0) FROM customers
		WHERE organization_id = $1
		AND current_balance > 0
		AND id IN (
			SELECT DISTINCT customer_id FROM customer_ledger
			WHERE transaction_type = 'SALE'
			AND created_at < NOW() - INTERVAL '30 days'
		)
	`
	var overdueDebts float64
	err = db.GetContext(ctx, &overdueDebts, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Overdue Debts = %.2f\n", overdueDebts)
	}
	fmt.Println()

	// Test 10: Pending returns
	fmt.Println("10. Testing pending returns query...")
	query = `SELECT COUNT(*) FROM returns WHERE organization_id = $1 AND status = 'pending'`
	var pendingReturns int
	err = db.GetContext(ctx, &pendingReturns, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Pending Returns = %d\n", pendingReturns)
	}
	fmt.Println()

	// Test 11: Pending claims
	fmt.Println("11. Testing pending claims query...")
	query = `SELECT COUNT(*) FROM warranties WHERE organization_id = $1 AND is_active = true AND expires_at < CURRENT_DATE + INTERVAL '30 days'`
	var pendingClaims int
	err = db.GetContext(ctx, &pendingClaims, query, orgID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("   ✅ SUCCESS: Pending Claims = %d\n", pendingClaims)
	}
	fmt.Println()

	fmt.Println("=== Test Complete ===")
}