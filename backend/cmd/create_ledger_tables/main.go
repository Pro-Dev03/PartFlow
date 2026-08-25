package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Direct database URL for Supabase
	databaseURL := "postgresql://postgres.vjaokclajtvpdjmdrybu:9IyxKnou9CUgrj5k@aws-0-eu-central-1.pooler.supabase.com:5432/postgres"

	// Connect to database
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Connected to database successfully")

	// Create customer_ledger table
	customerLedgerSQL := `
	CREATE TABLE IF NOT EXISTS customer_ledger (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
		type VARCHAR(20) NOT NULL CHECK (type IN ('debit', 'credit')),
		amount DECIMAL(10,2) NOT NULL,
		balance DECIMAL(10,2) NOT NULL,
		description TEXT,
		reference_id UUID,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_customer_ledger_customer ON customer_ledger(customer_id);
	CREATE INDEX IF NOT EXISTS idx_customer_ledger_type ON customer_ledger(type);
	CREATE INDEX IF NOT EXISTS idx_customer_ledger_created_at ON customer_ledger(created_at);
	`

	_, err = db.Exec(customerLedgerSQL)
	if err != nil {
		log.Fatalf("Failed to create customer_ledger table: %v", err)
	}
	fmt.Println("Successfully created customer_ledger table")

	// Create supplier_ledger table
	supplierLedgerSQL := `
	CREATE TABLE IF NOT EXISTS supplier_ledger (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
		type VARCHAR(20) NOT NULL CHECK (type IN ('debit', 'credit')),
		amount DECIMAL(10,2) NOT NULL,
		balance DECIMAL(10,2) NOT NULL,
		description TEXT,
		reference_id UUID,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_supplier_ledger_supplier ON supplier_ledger(supplier_id);
	CREATE INDEX IF NOT EXISTS idx_supplier_ledger_type ON supplier_ledger(type);
	CREATE INDEX IF NOT EXISTS idx_supplier_ledger_created_at ON supplier_ledger(created_at);
	`

	_, err = db.Exec(supplierLedgerSQL)
	if err != nil {
		log.Fatalf("Failed to create supplier_ledger table: %v", err)
	}
	fmt.Println("Successfully created supplier_ledger table")

	// Initialize ledger entries for existing customers with balances
	initLedgerSQL := `
	INSERT INTO customer_ledger (customer_id, type, amount, balance, description, created_at)
	SELECT 
		c.id,
		'debit',
		c.current_balance,
		c.current_balance,
		'Initial balance migration',
		c.created_at
	FROM customers c
	WHERE c.current_balance > 0
	AND NOT EXISTS (
		SELECT 1 FROM customer_ledger cl WHERE cl.customer_id = c.id
	);
	`

	result, err := db.Exec(initLedgerSQL)
	if err != nil {
		log.Fatalf("Failed to initialize ledger entries: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Initialized %d ledger entries for existing customers\n", rowsAffected)

	fmt.Println("All ledger tables created and initialized successfully")
}