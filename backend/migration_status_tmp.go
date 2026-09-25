package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil || db.Ping() != nil {
		fmt.Println("database-unavailable")
		os.Exit(2)
	}
	defer db.Close()
	var ledgerExists bool
	if err := db.QueryRow(`SELECT to_regclass('public.schema_migrations') IS NOT NULL`).Scan(&ledgerExists); err != nil || !ledgerExists {
		fmt.Println("migration-ledger-missing; stop-before-applying")
		os.Exit(3)
	}
	rows, err := db.Query(`SELECT version FROM public.schema_migrations ORDER BY version`)
	if err != nil {
		fmt.Println("migration-ledger-unreadable")
		os.Exit(4)
	}
	defer rows.Close()
	applied := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			fmt.Println("migration-ledger-unreadable")
			os.Exit(4)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		fmt.Println("migration-ledger-unreadable")
		os.Exit(4)
	}
	files, err := filepath.Glob(filepath.Join("migrations", "*.sql"))
	if err != nil {
		fmt.Println("migration-files-unreadable")
		os.Exit(5)
	}
	sort.Strings(files)
	sort.SliceStable(files, func(i, j int) bool {
		left, right := filepath.Base(files[i]), filepath.Base(files[j])
		if strings.HasPrefix(left, "001_") && strings.HasPrefix(right, "001_") {
			if strings.Contains(left, "_initial_") {
				return true
			}
			if strings.Contains(right, "_initial_") {
				return false
			}
		}
		return left < right
	})
	pending := 0
	for _, file := range files {
		filename := filepath.Base(file)
		version := strings.TrimSuffix(filename, ".sql")
		if version == "080_tenant_isolation" {
			continue
		}
		if applied[version] || applied[filename] {
			continue
		}
		pending++
		fmt.Println(filename)
	}
	fmt.Printf("pending-count=%d\n", pending)
	for _, table := range []string{
		"warranties", "warranty_claims", "roles", "role_permissions", "permissions",
		"organizations", "returns", "return_items", "inspections", "inspection_items",
		"item_repair_costs", "item_history",
	} {
		var exists bool
		if err := db.QueryRow(`SELECT to_regclass($1) IS NOT NULL`, "public."+table).Scan(&exists); err != nil {
			fmt.Printf("table-check-failed=%s\n", table)
			continue
		}
		if !exists {
			fmt.Printf("table=%s exists=false\n", table)
			continue
		}
		var rows int64
		if err := db.QueryRow(`SELECT COUNT(*) FROM public.` + table).Scan(&rows); err != nil {
			fmt.Printf("table=%s exists=true row-count=unavailable\n", table)
			continue
		}
		fmt.Printf("table=%s exists=true row-count=%d\n", table, rows)
	}
	for _, name := range []string{
		"accounting_returns", "accounting_return_items", "returns_summary", "sales_returns_analysis",
		"return_effects", "return_effect_items", "customer_debts", "supplier_debts",
	} {
		var objectType sql.NullString
		if err := db.QueryRow(`SELECT c.relkind::text FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relname=$1`, name).Scan(&objectType); err == sql.ErrNoRows {
			fmt.Printf("object=%s exists=false\n", name)
		} else if err != nil {
			fmt.Printf("object-check-failed=%s\n", name)
		} else {
			fmt.Printf("object=%s relkind=%s\n", name, objectType.String)
		}
	}
	for _, table := range []string{"returns", "return_items"} {
		rows, err := db.Query(`SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 ORDER BY ordinal_position`, table)
		if err != nil {
			fmt.Printf("columns-unavailable=%s\n", table)
			continue
		}
		columns := []string{}
		for rows.Next() {
			var column string
			if rows.Scan(&column) == nil {
				columns = append(columns, column)
			}
		}
		_ = rows.Close()
		fmt.Printf("columns=%s:%s\n", table, strings.Join(columns, ","))
	}
}
