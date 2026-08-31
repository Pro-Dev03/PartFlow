//go:build ignore

// seed_test_dataset resets the local operational dataset and creates a
// reproducible, realistic scenario for manual and automated testing.
// Authentication users and their subscription records are intentionally kept.
//
// Run from backend after stopping the local API:
//
//	go run ./scripts/seed_test_dataset.go --confirm
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/localdb"
)

type sqlExecer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

type customer struct {
	id, code, name string
}

type supplier struct {
	id, code, name string
}

type product struct {
	id, sku, name, categoryID, brandID string
	cost, price                        float64
	minStock                           int
}

type inventoryItem struct {
	id, productID, code, condition, status string
	cost, price                            float64
	createdAt, soldAt                      string
}

func main() {
	confirm := flag.Bool("confirm", false, "allow destructive local database reset")
	flag.Parse()
	if !*confirm {
		log.Fatal("refusing to reset local data; rerun with --confirm")
	}

	database, err := localdb.Open()
	if err != nil {
		log.Fatalf("open local database: %v", err)
	}
	defer database.DB.Close()

	if err := resetAndSeed(database.DB); err != nil {
		log.Fatal(err)
	}

	var customers, suppliers, products, inventory, sales, purchases, debts int
	for table, target := range map[string]*int{
		"customers":       &customers,
		"suppliers":       &suppliers,
		"products":        &products,
		"inventory_items": &inventory,
		"sales":           &sales,
		"purchases":       &purchases,
		"debts":           &debts,
	} {
		if err := database.DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(target); err != nil {
			log.Fatalf("count %s: %v", table, err)
		}
	}
	fmt.Printf("Test dataset ready: customers=%d suppliers=%d products=%d inventory_items=%d purchases=%d sales=%d debts=%d\n", customers, suppliers, products, inventory, purchases, sales, debts)
	fmt.Println("Authentication users and subscription records were preserved.")
}

func resetAndSeed(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disable foreign keys for reset: %w", err)
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin reset transaction: %w", err)
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		_, _ = db.Exec("PRAGMA foreign_keys = ON")
		return cause
	}

	// Operational data only. Users, local sessions, settings and subscription
	// data are deliberately not in this list.
	for _, table := range []string{
		"acquisition_items", "acquisitions", "audit_logs", "barcodes", "customer_debts", "customer_ledger", "customer_payments",
		"daily_debt_summary", "daily_inventory_summary", "daily_profit_summary", "daily_sales_summary",
		"debt_collections", "debts", "expense_categories", "expenses", "held_sales",
		"inspection_items", "inspections", "inventory", "inventory_items", "inventory_movements",
		"item_history", "item_repair_costs", "item_specification_values", "ledger_entries",
		"locations", "monthly_debt_summary", "monthly_inventory_summary", "monthly_profit_summary",
		"monthly_sales_summary", "notification_preferences", "notifications", "part_specifications",
		"payments", "products", "purchase_items", "purchases", "reports", "reservations",
		"return_items", "returns", "sale_items", "sales", "seller_payments", "supplier_ledger", "supplier_payments",
		"supplier_return_items", "supplier_returns", "sync_conflicts", "sync_queue", "trade_ins",
		"type_specifications", "part_types", "brands", "categories", "customers", "suppliers",
		"warranty_claims", "password_reset_tokens", "refresh_tokens",
	} {
		if _, err := tx.Exec("DELETE FROM \"" + table + "\""); err != nil {
			return rollback(fmt.Errorf("clear %s: %w", table, err))
		}
	}

	now := time.Now()
	stamp := func(t time.Time) string { return t.Format(time.RFC3339) }
	date := func(t time.Time) string { return t.Format("2006-01-02") }
	today := date(now)
	yesterday := date(now.AddDate(0, 0, -1))
	tenDaysAgo := date(now.AddDate(0, 0, -10))
	fourteenDaysAhead := date(now.AddDate(0, 0, 14))
	sevenDaysAhead := date(now.AddDate(0, 0, 7))
	var ownerID string
	if err := tx.QueryRow("SELECT id FROM users ORDER BY created_at LIMIT 1").Scan(&ownerID); err != nil || ownerID == "" {
		return rollback(fmt.Errorf("find a preserved user for test records: %w", err))
	}

	// Keep local operation mode explicit so a normal browser refresh does not
	// replace this test dataset with a cloud snapshot.
	if _, err := tx.Exec("INSERT OR REPLACE INTO local_metadata (key, value, updated_at) VALUES (?, ?, ?)", "operating_mode", "offline", stamp(now)); err != nil {
		return rollback(fmt.Errorf("set local operating mode: %w", err))
	}
	if _, err := tx.Exec("DELETE FROM local_metadata WHERE key IN ('last_cloud_sync_at', 'cloud_snapshot_version')"); err != nil {
		return rollback(fmt.Errorf("clear cloud sync metadata: %w", err))
	}

	categoryNew := uuid.NewString()
	categoryUsed := uuid.NewString()
	categoryElectrical := uuid.NewString()
	categoryBody := uuid.NewString()
	for _, row := range []struct{ id, name, description string }{
		{categoryNew, "New Computer Components", "Factory-sealed PC components"},
		{categoryUsed, "Used Computer Components", "Inspected used components and trade-ins"},
		{categoryElectrical, "Memory and Storage", "RAM, SSD and other data components"},
		{categoryBody, "Peripherals", "Displays, cases and computer accessories"},
	} {
		if _, err := tx.Exec(`INSERT INTO categories (id, name, description, icon, color, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?)`, row.id, row.name, row.description, "package", "#2563eb", stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert category %s: %w", row.name, err))
		}
	}

	brandID := uuid.NewString()
	if _, err := tx.Exec(`INSERT INTO brands (id, name, description, logo_url, created_at, updated_at) VALUES (?, ?, ?, '', ?, ?)`, brandID, "TechSource Components", "Computer hardware distributor", stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert test brand: %w", err))
	}

	partTypeIDs := make([]string, 0, 4)
	for i, row := range []struct{ id, ar, en string }{
		{uuid.NewString(), "قطع جديدة", "New computer parts"},
		{uuid.NewString(), "قطع مستعملة", "Used computer parts"},
		{uuid.NewString(), "ذاكرة وتخزين", "Memory and storage"},
		{uuid.NewString(), "ملحقات الحاسوب", "Computer peripherals"},
	} {
		partTypeIDs = append(partTypeIDs, row.id)
		if _, err := tx.Exec(`INSERT INTO part_types (id, name_ar, name_en, icon, color, is_active, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?)`, row.id, row.ar, row.en, "package", "#2563eb", i+1, stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert part type %s: %w", row.en, err))
		}
	}

	locationID := uuid.NewString()
	if _, err := tx.Exec(`INSERT INTO locations (id, name, type, description, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, 1, ?, ?)`, locationID, "Main Test Warehouse", "warehouse", "Synthetic test location", stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert location: %w", err))
	}
	expenseCategoryID := uuid.NewString()
	if _, err := tx.Exec(`INSERT INTO expense_categories (id, name, description, color, icon, budget, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`, expenseCategoryID, "Operations", "Test operating expenses", "#f59e0b", "wallet", 5000, stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert expense category: %w", err))
	}

	customers := []customer{
		{uuid.NewString(), "C-001", "Ahmad Mansour"},
		{uuid.NewString(), "C-002", "Maya Khalil"},
		{uuid.NewString(), "C-003", "Omar Saleh"},
		{uuid.NewString(), "C-004", "Rana Haddad"},
	}
	for i, row := range customers {
		note := "Retail customer"
		if i >= 2 {
			note = "Customer and used-parts seller"
		}
		if _, err := tx.Exec(`INSERT INTO customers (id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, 1, ?, ?)`, row.id, row.code, row.name, fmt.Sprintf("%s@example.test", row.code), fmt.Sprintf("+97059000000%d", i+1), fmt.Sprintf("Test street %d", i+1), "Gaza", "Palestine", 1500, note, stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert customer %s: %w", row.name, err))
		}
	}

	suppliers := []supplier{
		{uuid.NewString(), "SUP-001", "TechCore Wholesale"},
		{uuid.NewString(), "SUP-002", "Digital Gear Distribution"},
	}
	for i, row := range suppliers {
		if _, err := tx.Exec(`INSERT INTO suppliers (id, code, name, email, phone, address, city, country, credit_limit, current_balance, notes, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, 1, ?, ?)`, row.id, row.code, row.name, fmt.Sprintf("%s@example.test", row.code), fmt.Sprintf("+97059100000%d", i+1), fmt.Sprintf("Supplier road %d", i+1), "Gaza", "Palestine", 10000, "Synthetic supplier for testing", stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert supplier %s: %w", row.name, err))
		}
	}

	products := []product{
		{uuid.NewString(), "CPU-NEW-001", "Intel Core i5-14600K Processor", categoryNew, brandID, 900, 1250, 1},
		{uuid.NewString(), "GPU-NEW-001", "NVIDIA GeForce RTX 4070 SUPER 12GB", categoryNew, brandID, 2200, 2850, 2},
		{uuid.NewString(), "RAM-NEW-001", "Corsair Vengeance 32GB DDR5 Kit", categoryElectrical, brandID, 360, 520, 2},
		{uuid.NewString(), "SSD-NEW-001", "Samsung 990 PRO NVMe SSD 1TB", categoryElectrical, brandID, 300, 450, 3},
		{uuid.NewString(), "GPU-USED-001", "Used ASUS RTX 3060 12GB", categoryUsed, brandID, 700, 980, 1},
		{uuid.NewString(), "LAP-USED-001", "Refurbished Lenovo ThinkPad T14", categoryBody, brandID, 900, 1350, 1},
	}
	for _, row := range products {
		if _, err := tx.Exec(`INSERT INTO products (id, sku, name, description, category_id, brand_id, purchase_price, cost_price, selling_price, currency, min_stock_level, max_stock_level, warranty_days, track_serial, track_individual, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'ILS', ?, ?, ?, 0, 1, 1, ?, ?)`, row.id, row.sku, row.name, "Synthetic product for end-to-end testing", row.categoryID, row.brandID, row.cost, row.cost, row.price, row.minStock, row.minStock+10, 30, stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert product %s: %w", row.name, err))
		}
	}

	productBySKU := make(map[string]product, len(products))
	for _, row := range products {
		productBySKU[row.sku] = row
	}
	// Populate searchable metadata used by product, barcode and supplier views.
	for i, row := range products {
		barcode := fmt.Sprintf("8901000000%02d", i+1)
		model := []string{"BX8071514600K", "RTX4070S-12G", "CMK32GX5M2B5600C36", "MZ-V9P1T0BW", "DUAL-RTX3060-12G", "T14-GEN2-I5"}[i]
		supplierID := suppliers[i%len(suppliers)].id
		if _, err := tx.Exec(`UPDATE products SET model = ?, barcode = ?, preferred_supplier_id = ? WHERE id = ?`, model, barcode, supplierID, row.id); err != nil {
			return rollback(fmt.Errorf("update product metadata %s: %w", row.sku, err))
		}
	}
	items := []inventoryItem{
		{uuid.NewString(), productBySKU["CPU-NEW-001"].id, "IT-CPU-001", "NEW", "SOLD", 900, 1250, stamp(now.Add(-2 * time.Hour)), stamp(now.Add(-90 * time.Minute))},
		{uuid.NewString(), productBySKU["CPU-NEW-001"].id, "IT-CPU-002", "NEW", "AVAILABLE", 900, 1250, stamp(now.Add(-2 * time.Hour)), ""},
		{uuid.NewString(), productBySKU["CPU-NEW-001"].id, "IT-CPU-003", "NEW", "SOLD", 900, 1250, stamp(now.AddDate(0, 0, -1)), stamp(now.AddDate(0, 0, -1))},
		{uuid.NewString(), productBySKU["CPU-NEW-001"].id, "IT-CPU-004", "NEW", "SOLD", 900, 1250, stamp(now.AddDate(0, 0, -10)), stamp(now.AddDate(0, 0, -10))},
		{uuid.NewString(), productBySKU["GPU-NEW-001"].id, "IT-GPU-001", "NEW", "SOLD", 2200, 2850, stamp(now.Add(-2 * time.Hour)), stamp(now.Add(-80 * time.Minute))},
		{uuid.NewString(), productBySKU["GPU-NEW-001"].id, "IT-GPU-002", "NEW", "AVAILABLE", 2200, 2850, stamp(now.Add(-2 * time.Hour)), ""},
		{uuid.NewString(), productBySKU["RAM-NEW-001"].id, "IT-RAM-001", "NEW", "SOLD", 360, 520, stamp(now.Add(-2 * time.Hour)), stamp(now.Add(-80 * time.Minute))},
		{uuid.NewString(), productBySKU["RAM-NEW-001"].id, "IT-RAM-002", "NEW", "AVAILABLE", 360, 520, stamp(now.Add(-2 * time.Hour)), ""},
		{uuid.NewString(), productBySKU["RAM-NEW-001"].id, "IT-RAM-003", "NEW", "AVAILABLE", 360, 520, stamp(now.Add(-2 * time.Hour)), ""},
		{uuid.NewString(), productBySKU["SSD-NEW-001"].id, "IT-SSD-001", "NEW", "SOLD", 300, 450, stamp(now.Add(-2 * time.Hour)), stamp(now.Add(-80 * time.Minute))},
		{uuid.NewString(), productBySKU["SSD-NEW-001"].id, "IT-SSD-002", "NEW", "AVAILABLE", 300, 450, stamp(now.Add(-2 * time.Hour)), ""},
		{uuid.NewString(), productBySKU["SSD-NEW-001"].id, "IT-SSD-003", "NEW", "AVAILABLE", 300, 450, stamp(now.Add(-2 * time.Hour)), ""},
		{uuid.NewString(), productBySKU["GPU-USED-001"].id, "IT-UGPU-001", "USED", "SOLD", 700, 980, stamp(now.Add(-3 * time.Hour)), stamp(now.Add(-70 * time.Minute))},
		{uuid.NewString(), productBySKU["GPU-USED-001"].id, "IT-UGPU-002", "USED", "AVAILABLE", 700, 980, stamp(now.Add(-3 * time.Hour)), ""},
		{uuid.NewString(), productBySKU["LAP-USED-001"].id, "IT-LAP-001", "USED", "AVAILABLE", 900, 1350, stamp(now.AddDate(0, 0, -1)), ""},
		{uuid.NewString(), productBySKU["LAP-USED-001"].id, "IT-LAP-002", "USED", "IN_REPAIR", 850, 1350, stamp(now.AddDate(0, 0, -2)), ""},
	}
	for _, row := range items {
		if _, err := tx.Exec(`INSERT INTO inventory_items (id, product_id, item_code, condition, purchase_cost, selling_price, status, location_id, purchase_date, sold_at, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.id, row.productID, row.code, row.condition, row.cost, row.price, row.status, locationID, row.createdAt, nullable(row.soldAt), "Synthetic item for testing", row.createdAt, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert inventory item %s: %w", row.code, err))
		}
	}

	itemByCode := make(map[string]inventoryItem, len(items))
	for _, row := range items {
		itemByCode[row.code] = row
	}

	// Purchases: received, partially paid, and pending supplier balances.
	purchase1 := uuid.NewString()
	purchase2 := uuid.NewString()
	purchase3 := uuid.NewString()
	purchases := []struct {
		id, number, supplierID, created, purchaseDate, status string
		total, paid, remaining                                float64
	}{
		{purchase1, "PO-TEST-001", suppliers[0].id, stamp(now.Add(-4 * time.Hour)), today, "received", 4000, 2500, 1500},
		{purchase2, "PO-TEST-002", suppliers[1].id, stamp(now.AddDate(0, 0, -1)), yesterday, "received", 2700, 2700, 0},
		{purchase3, "PO-TEST-003", suppliers[0].id, stamp(now.AddDate(0, 0, -10)), tenDaysAgo, "pending", 850, 0, 850},
	}
	for _, row := range purchases {
		if _, err := tx.Exec(`INSERT INTO purchases (id, purchase_number, invoice_number, supplier_id, total_amount, tax_amount, discount_amount, paid_amount, remaining_amount, status, notes, purchase_date, expected_delivery_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?, ?, ?, ?)`, row.id, row.number, "INV-"+row.number, row.supplierID, row.total, row.paid, row.remaining, row.status, "Synthetic purchase for testing", row.purchaseDate, sevenDaysAhead, row.created, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert purchase %s: %w", row.number, err))
		}
	}
	for _, row := range []struct {
		id, purchaseID, productID string
		qty                       int
		unit, total               float64
		created                   string
	}{
		{uuid.NewString(), purchase1, productBySKU["CPU-NEW-001"].id, 2, 900, 1800, stamp(now.Add(-4 * time.Hour))},
		{uuid.NewString(), purchase1, productBySKU["GPU-NEW-001"].id, 1, 2200, 2200, stamp(now.Add(-4 * time.Hour))},
		{uuid.NewString(), purchase2, productBySKU["RAM-NEW-001"].id, 5, 360, 1800, stamp(now.AddDate(0, 0, -1))},
		{uuid.NewString(), purchase2, productBySKU["SSD-NEW-001"].id, 3, 300, 900, stamp(now.AddDate(0, 0, -1))},
		{uuid.NewString(), purchase3, productBySKU["LAP-USED-001"].id, 1, 850, 850, stamp(now.AddDate(0, 0, -10))},
	} {
		if _, err := tx.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, row.id, row.purchaseID, row.productID, row.qty, row.unit, row.total, row.created); err != nil {
			return rollback(fmt.Errorf("insert purchase item: %w", err))
		}
	}

	// Used-parts acquisitions from two customers, including inspection and
	// trade-in records.
	acq1 := uuid.NewString()
	acq2 := uuid.NewString()
	acq3 := uuid.NewString()
	for _, row := range []struct {
		id, number, customerID, created, acquisitionDate, status, paymentStatus string
		total, paid                                                             float64
	}{
		{acq1, "ACQ-TEST-001", customers[2].id, stamp(now.Add(-3 * time.Hour)), today, "received", "unpaid", 700, 0},
		{acq2, "ACQ-TEST-002", customers[3].id, stamp(now.AddDate(0, 0, -1)), yesterday, "completed", "paid", 900, 900},
		{acq3, "ACQ-TEST-003", customers[3].id, stamp(now.AddDate(0, 0, -2)), date(now.AddDate(0, 0, -2)), "pending", "unpaid", 850, 0},
	} {
		if _, err := tx.Exec(`INSERT INTO acquisitions (id, type, acquisition_date, customer_id, total_cost, paid_amount, payment_status, status, notes, created_at, updated_at) VALUES (?, 'CUSTOMER', ?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.id, row.acquisitionDate, row.customerID, row.total, row.paid, row.paymentStatus, row.status, "Customer used-part purchase", row.created, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert acquisition %s: %w", row.number, err))
		}
	}
	for _, row := range []struct {
		id, acquisitionID, productID, inventoryID, code, condition, grade, inspectionStatus, itemStatus string
		unit, total                                                                                     float64
		created                                                                                         string
	}{
		{uuid.NewString(), acq1, productBySKU["GPU-USED-001"].id, itemByCode["IT-UGPU-001"].id, "IT-UGPU-001", "USED", "B", "PASSED", "RECEIVED", 700, 700, stamp(now.Add(-3 * time.Hour))},
		{uuid.NewString(), acq2, productBySKU["LAP-USED-001"].id, itemByCode["IT-LAP-001"].id, "IT-LAP-001", "USED", "A", "PASSED", "RECEIVED", 900, 900, stamp(now.AddDate(0, 0, -1))},
		{uuid.NewString(), acq3, productBySKU["LAP-USED-001"].id, itemByCode["IT-LAP-002"].id, "IT-LAP-002", "USED", "C", "PENDING", "RECEIVED", 850, 850, stamp(now.AddDate(0, 0, -2))},
	} {
		if _, err := tx.Exec(`INSERT INTO acquisition_items (id, acquisition_id, product_id, inventory_item_id, item_code, condition, grade, inspection_status, item_status, unit_cost, total_cost, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.id, row.acquisitionID, row.productID, row.inventoryID, row.code, row.condition, row.grade, row.inspectionStatus, row.itemStatus, row.unit, row.total, row.created, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert acquisition item %s: %w", row.code, err))
		}
	}
	for _, row := range []struct {
		id, productID, inventoryID, inspectorID, inspectionDate, result, condition, grade, notes, acquisitionItemID string
	}{
		{uuid.NewString(), productBySKU["GPU-USED-001"].id, itemByCode["IT-UGPU-001"].id, "", today, "PASSED", "USED", "B", "GPU stress-tested for 30 minutes", ""},
		{uuid.NewString(), productBySKU["LAP-USED-001"].id, itemByCode["IT-LAP-001"].id, "", yesterday, "PASSED", "USED", "A", "Laptop diagnostics passed", ""},
	} {
		if _, err := tx.Exec(`INSERT INTO inspections (id, product_id, inventory_item_id, inspector_id, inspection_date, result, condition, grade, notes, acquisition_item_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?)`, row.id, row.productID, row.inventoryID, ownerID, row.inspectionDate, row.result, row.condition, row.grade, row.notes, row.acquisitionItemID, stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert inspection: %w", err))
		}
	}
	for _, row := range []struct {
		id, customerID, inventoryID, purchaseDate, notes string
		price                                            float64
	}{
		{uuid.NewString(), customers[2].id, itemByCode["IT-UGPU-001"].id, today, "Trade-in from Omar Saleh", 700},
		{uuid.NewString(), customers[3].id, itemByCode["IT-LAP-001"].id, yesterday, "Trade-in from Rana Haddad", 900},
	} {
		if _, err := tx.Exec(`INSERT INTO trade_ins (id, customer_id, inventory_item_id, purchase_price, purchase_date, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, row.id, row.customerID, row.inventoryID, row.price, row.purchaseDate, row.notes, stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert trade-in: %w", err))
		}
	}

	// Sales cover cash, card, customer debt, a previous-day sale and an overdue
	// debt. The dates are intentional so dashboard day comparisons are visible.
	sale1 := uuid.NewString()
	sale2 := uuid.NewString()
	sale3 := uuid.NewString()
	sale4 := uuid.NewString()
	sale5 := uuid.NewString()
	for _, row := range []struct {
		id, number, invoice, customerID, created, saleDate, method, paymentStatus, status string
		subtotal, total, paid, remaining, cost, gross, net                                float64
	}{
		{sale1, "S-TEST-001", "INV-S-001", customers[0].id, stamp(now.Add(-90 * time.Minute)), stamp(now.Add(-90 * time.Minute)), "cash", "paid", "completed", 1250, 1250, 1250, 0, 900, 350, 350},
		{sale2, "S-TEST-002", "INV-S-002", customers[1].id, stamp(now.Add(-70 * time.Minute)), stamp(now.Add(-70 * time.Minute)), "card", "paid", "completed", 980, 980, 980, 0, 700, 280, 280},
		{sale3, "S-TEST-003", "INV-S-003", customers[2].id, stamp(now.Add(-80 * time.Minute)), stamp(now.Add(-80 * time.Minute)), "debt", "partial", "completed", 3370, 3370, 1000, 2370, 2560, 810, 810},
		{sale4, "S-TEST-004", "INV-S-004", customers[3].id, stamp(now.AddDate(0, 0, -1)), stamp(now.AddDate(0, 0, -1)), "cash", "paid", "completed", 1250, 1250, 1250, 0, 900, 350, 350},
		{sale5, "S-TEST-005", "INV-S-005", customers[3].id, stamp(now.AddDate(0, 0, -10)), stamp(now.AddDate(0, 0, -10)), "debt", "unpaid", "completed", 450, 450, 0, 450, 300, 150, 150},
	} {
		if _, err := tx.Exec(`INSERT INTO sales (id, sale_number, invoice_number, customer_id, total_amount, tax_amount, discount_amount, paid_amount, remaining_amount, payment_method, status, notes, created_at, updated_at, sale_date, subtotal, cost_amount, gross_profit, net_profit, payment_status) VALUES (?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.id, row.number, row.invoice, row.customerID, row.total, row.paid, row.remaining, row.method, row.status, "Synthetic sale for testing", row.created, stamp(now), row.saleDate, row.subtotal, row.cost, row.gross, row.net, row.paymentStatus); err != nil {
			return rollback(fmt.Errorf("insert sale %s: %w", row.number, err))
		}
	}
	sale2ItemID := uuid.NewString()
	sale4ItemID := uuid.NewString()
	sale5ItemID := uuid.NewString()
	for _, row := range []struct {
		id, saleID, itemID, productID string
		qty                           int
		unit, total, cost             float64
		created                       string
	}{
		{uuid.NewString(), sale1, itemByCode["IT-CPU-001"].id, productBySKU["CPU-NEW-001"].id, 1, 1250, 1250, 900, stamp(now.Add(-90 * time.Minute))},
		{sale2ItemID, sale2, itemByCode["IT-UGPU-001"].id, productBySKU["GPU-USED-001"].id, 1, 980, 980, 700, stamp(now.Add(-70 * time.Minute))},
		{uuid.NewString(), sale3, itemByCode["IT-GPU-001"].id, productBySKU["GPU-NEW-001"].id, 1, 2850, 2850, 2200, stamp(now.Add(-80 * time.Minute))},
		{uuid.NewString(), sale3, itemByCode["IT-RAM-001"].id, productBySKU["RAM-NEW-001"].id, 1, 520, 520, 360, stamp(now.Add(-80 * time.Minute))},
		{sale4ItemID, sale4, itemByCode["IT-CPU-003"].id, productBySKU["CPU-NEW-001"].id, 1, 1250, 1250, 900, stamp(now.AddDate(0, 0, -1))},
		{sale5ItemID, sale5, itemByCode["IT-SSD-001"].id, productBySKU["SSD-NEW-001"].id, 1, 450, 450, 300, stamp(now.AddDate(0, 0, -10))},
	} {
		if _, err := tx.Exec(`INSERT INTO sale_items (id, sale_id, inventory_item_id, product_id, quantity, unit_price, item_total, created_at, total_amount, unit_cost) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.id, row.saleID, row.itemID, row.productID, row.qty, row.unit, row.total, row.created, row.total, row.cost); err != nil {
			return rollback(fmt.Errorf("insert sale item: %w", err))
		}
	}

	// Customer debts and supplier balances/payments.
	debt1 := uuid.NewString()
	debt2 := uuid.NewString()
	for _, row := range []struct {
		id, customerID, saleID  string
		amount, paid, remaining float64
		due, status, created    string
	}{
		{debt1, customers[2].id, sale3, 3370, 1000, 2370, fourteenDaysAhead, "pending", today},
		{debt2, customers[3].id, sale5, 450, 0, 450, date(now.AddDate(0, 0, -2)), "overdue", tenDaysAgo},
	} {
		if _, err := tx.Exec(`INSERT INTO debts (id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.id, row.customerID, row.saleID, row.amount, row.paid, row.remaining, row.due, row.status, "Synthetic debt for testing", row.created, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert debt: %w", err))
		}
	}
	if _, err := tx.Exec(`UPDATE customers SET current_balance = CASE WHEN id = ? THEN 2370 WHEN id = ? THEN 450 ELSE 0 END, updated_at = ?`, customers[2].id, customers[3].id, stamp(now)); err != nil {
		return rollback(fmt.Errorf("update customer balances: %w", err))
	}
	for _, row := range []struct {
		id, customerID, typ, transactionType, referenceID, description string
		amount, balance                                                float64
		created                                                        string
	}{
		{uuid.NewString(), customers[2].id, "debit", "sale", sale3, "Invoice S-TEST-003", 3370, 3370, stamp(now.Add(-80 * time.Minute))},
		{uuid.NewString(), customers[2].id, "credit", "payment", "", "Partial payment", 1000, 2370, stamp(now.Add(-60 * time.Minute))},
		{uuid.NewString(), customers[3].id, "debit", "sale", sale5, "Overdue invoice S-TEST-005", 450, 450, stamp(now.AddDate(0, 0, -10))},
	} {
		if _, err := tx.Exec(`INSERT INTO customer_ledger (id, customer_id, type, transaction_type, amount, balance, description, reference_id, reference_type, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'sale', ?)`, row.id, row.customerID, row.typ, row.transactionType, row.amount, row.balance, row.description, nullable(row.referenceID), row.created); err != nil {
			return rollback(fmt.Errorf("insert customer ledger: %w", err))
		}
	}
	paymentCustomer := uuid.NewString()
	paymentSupplier := uuid.NewString()
	if _, err := tx.Exec(`INSERT INTO payments (id, transaction_number, customer_id, amount, payment_method, reference, notes, created_at, payment_date, sale_id, payment_status) VALUES (?, ?, ?, 1000, 'cash', ?, ?, ?, ?, ?, 'completed')`, paymentCustomer, "PAY-TEST-C-001", customers[2].id, "S-TEST-003", "Customer partial payment", stamp(now.Add(-60*time.Minute)), today, sale3); err != nil {
		return rollback(fmt.Errorf("insert customer payment: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO payments (id, transaction_number, supplier_id, amount, payment_method, reference, notes, created_at, payment_date, purchase_id, payment_status) VALUES (?, ?, ?, 2500, 'bank_transfer', ?, ?, ?, ?, ?, 'completed')`, paymentSupplier, "PAY-TEST-S-001", suppliers[0].id, "PO-TEST-001", "Supplier partial payment", stamp(now.Add(-3*time.Hour)), today, purchase1); err != nil {
		return rollback(fmt.Errorf("insert supplier payment: %w", err))
	}
	if _, err := tx.Exec(`UPDATE suppliers SET current_balance = CASE WHEN id = ? THEN 2350 ELSE 0 END, updated_at = ?`, suppliers[0].id, stamp(now)); err != nil {
		return rollback(fmt.Errorf("update supplier balances: %w", err))
	}
	for _, row := range []struct {
		id, supplierID, typ, transactionType, referenceID, description string
		amount, balance                                                float64
	}{
		{uuid.NewString(), suppliers[0].id, "debit", "purchase", purchase1, "Purchase PO-TEST-001", 4000, 4000},
		{uuid.NewString(), suppliers[0].id, "credit", "payment", purchase1, "Supplier payment", 2500, 1500},
		{uuid.NewString(), suppliers[0].id, "debit", "purchase", purchase3, "Pending purchase PO-TEST-003", 850, 2350},
	} {
		if _, err := tx.Exec(`INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, reference_type, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'purchase', ?)`, row.id, row.supplierID, row.typ, row.transactionType, row.amount, row.balance, row.description, row.referenceID, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert supplier ledger: %w", err))
		}
	}
	if _, err := tx.Exec(`INSERT INTO supplier_payments (id, supplier_id, amount, payment_date, method, reference, notes, created_at) VALUES (?, ?, 2500, ?, 'bank_transfer', ?, ?, ?)`, uuid.NewString(), suppliers[0].id, today, "PO-TEST-001", "Partial supplier payment", stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert supplier payment ledger: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO seller_payments (id, acquisition_id, customer_id, amount, payment_method, payment_date, notes, created_at) VALUES (?, ?, ?, 900, 'cash', ?, ?, ?)`, uuid.NewString(), acq2, customers[3].id, yesterday, "Paid used laptop trade-in", stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert seller payment: %w", err))
	}

	// Expenses make the net-profit calculation visibly different from gross
	// profit while keeping the values small and easy to verify by hand.
	for _, row := range []struct {
		id, title, expenseDate, status string
		amount                         float64
	}{
		{uuid.NewString(), "Delivery fuel", today, "approved", 35},
		{uuid.NewString(), "Packing supplies", yesterday, "paid", 20},
	} {
		if _, err := tx.Exec(`INSERT INTO expenses (id, title, category_id, amount, currency, description, expense_date, payment_method, created_at, updated_at, status, category) VALUES (?, ?, ?, ?, 'ILS', ?, ?, 'cash', ?, ?, ?, 'Operations')`, row.id, row.title, expenseCategoryID, row.amount, "Synthetic expense for testing", row.expenseDate, stamp(now), stamp(now), row.status); err != nil {
			return rollback(fmt.Errorf("insert expense %s: %w", row.title, err))
		}
	}

	// One pending customer return and a few movement/audit records exercise the
	// returns, history and activity screens.
	returnID := uuid.NewString()
	returnItemID := uuid.NewString()
	if _, err := tx.Exec(`INSERT INTO returns (id, return_number, customer_id, sale_id, total_refund_amount, refund_status, status, reason, notes, return_date, created_at, updated_at, return_type, refund_method) VALUES (?, ?, ?, ?, 1250, 'pending', 'pending', 'CUSTOMER_REQUEST', 'Pending exchange for CPU upgrade', ?, ?, ?, 'PARTIAL', 'CASH')`, returnID, "RET-TEST-001", customers[3].id, sale4, yesterday, stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert return: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO return_items (id, return_id, product_id, quantity, unit_price, total_refund_amount, sale_item_id, inventory_item_id, quantity_returned, original_quantity, reason, resolution, inventory_status, created_at, updated_at) VALUES (?, ?, ?, 1, 1250, 1250, ?, ?, 1, 1, 'Exchange requested', 'pending', 'SOLD', ?, ?)`, returnItemID, returnID, productBySKU["CPU-NEW-001"].id, sale4ItemID, itemByCode["IT-CPU-003"].id, stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert return item: %w", err))
	}
	completedReturnID := uuid.NewString()
	completedReturnItemID := uuid.NewString()
	if _, err := tx.Exec(`INSERT INTO returns (id, return_number, customer_id, sale_id, total_refund_amount, refund_status, status, reason, notes, return_date, created_at, updated_at, return_type, refund_method) VALUES (?, ?, ?, ?, 980, 'completed', 'COMPLETED', 'DEFECTIVE', 'Refund issued after diagnostics failure', ?, ?, ?, 'FULL', 'CARD')`, completedReturnID, "RET-TEST-002", customers[1].id, sale2, today, stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert completed return: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO return_items (id, return_id, product_id, quantity, unit_price, total_refund_amount, sale_item_id, inventory_item_id, quantity_returned, original_quantity, reason, resolution, inventory_status, created_at, updated_at) VALUES (?, ?, ?, 1, 980, 980, ?, ?, 1, 1, 'Defective fan', 'refund', 'IN_REPAIR', ?, ?)`, completedReturnItemID, completedReturnID, productBySKU["GPU-USED-001"].id, sale2ItemID, itemByCode["IT-UGPU-001"].id, stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert completed return item: %w", err))
	}
	if _, err := tx.Exec(`UPDATE inventory_items SET status = 'IN_REPAIR', sold_at = NULL, notes = 'Returned defective; awaiting repair', updated_at = ? WHERE id = ?`, stamp(now), itemByCode["IT-UGPU-001"].id); err != nil {
		return rollback(fmt.Errorf("update returned inventory item: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_at, is_reversed) VALUES (?, ?, ?, 'RETURN', 1, 0, 1, 'return', ?, 'Defective GPU returned by customer', ?, 0)`, uuid.NewString(), itemByCode["IT-UGPU-001"].id, productBySKU["GPU-USED-001"].id, completedReturnID, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert return inventory movement: %w", err))
	}
	for _, row := range []struct {
		id, itemID, productID, movementType, refType, refID, reason, created string
		qty, before, after                                                   int
	}{
		{uuid.NewString(), itemByCode["IT-CPU-001"].id, productBySKU["CPU-NEW-001"].id, "SALE", "sale", sale1, "CPU sold to Ahmad", stamp(now.Add(-90 * time.Minute)), 1, 1, 0},
		{uuid.NewString(), itemByCode["IT-UGPU-001"].id, productBySKU["GPU-USED-001"].id, "SALE", "sale", sale2, "Used GPU sold to Maya", stamp(now.Add(-70 * time.Minute)), 1, 1, 0},
		{uuid.NewString(), itemByCode["IT-LAP-002"].id, productBySKU["LAP-USED-001"].id, "DAMAGE", "inspection", acq3, "Laptop needs screen repair", stamp(now.AddDate(0, 0, -2)), 1, 0, 1},
	} {
		if _, err := tx.Exec(`INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_at, is_reversed) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`, row.id, row.itemID, row.productID, row.movementType, row.qty, row.before, row.after, row.refType, row.refID, row.reason, row.created); err != nil {
			return rollback(fmt.Errorf("insert inventory movement: %w", err))
		}
	}

	// Product-level inventory snapshots are maintained for screens that use the
	// aggregate table instead of individual inventory_items.
	for _, row := range []struct {
		productID          string
		quantity, reserved int
	}{
		{productBySKU["CPU-NEW-001"].id, 1, 0},
		{productBySKU["GPU-NEW-001"].id, 1, 0},
		{productBySKU["RAM-NEW-001"].id, 2, 0},
		{productBySKU["SSD-NEW-001"].id, 2, 0},
		{productBySKU["GPU-USED-001"].id, 1, 0},
		{productBySKU["LAP-USED-001"].id, 1, 0},
	} {
		if _, err := tx.Exec(`INSERT INTO inventory (id, product_id, quantity, reserved_quantity, location, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, uuid.NewString(), row.productID, row.quantity, row.reserved, "Main Test Warehouse", stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert inventory aggregate: %w", err))
		}
	}

	// Dashboard summaries are seeded for today/yesterday so aggregation and
	// report screens have immediate data; all source rows above remain the
	// authoritative records.
	if _, err := tx.Exec(`INSERT INTO daily_sales_summary (date, total_sales, total_revenue, total_profit, total_customers, average_order_value, total_items_sold, cash_sales, card_sales, debt_sales, updated_at) VALUES (?, 3, 5600, 1405, 3, 1866.6667, 4, 1250, 980, 3370, ?)`, today, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert today sales summary: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO daily_sales_summary (date, total_sales, total_revenue, total_profit, total_customers, average_order_value, total_items_sold, cash_sales, card_sales, debt_sales, updated_at) VALUES (?, 1, 1250, 330, 1, 1250, 1, 1250, 0, 0, ?)`, yesterday, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert yesterday sales summary: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO daily_profit_summary (date, gross_profit, net_profit, total_revenue, total_cost, profit_margin, updated_at) VALUES (?, 1440, 1405, 5600, 4160, 25.7143, ?)`, today, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert today profit summary: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO daily_profit_summary (date, gross_profit, net_profit, total_revenue, total_cost, profit_margin, updated_at) VALUES (?, 350, 330, 1250, 900, 26.4, ?)`, yesterday, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert yesterday profit summary: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO daily_debt_summary (date, total_debt, new_debt, payments_received, overdue_debt, overdue_count, paid_debt, updated_at) VALUES (?, 2820, 2370, 1000, 450, 1, 0, ?)`, today, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert today debt summary: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO daily_inventory_summary (date, total_items, total_value, low_stock_count, out_of_stock_count, new_items_added, items_sold, items_returned, items_damaged, updated_at) VALUES (?, 16, 17020, 2, 0, 16, 6, 1, 1, ?)`, today, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert today inventory summary: %w", err))
	}

	month := now.Format("2006-01")
	if _, err := tx.Exec(`INSERT INTO monthly_sales_summary (year, month, total_sales, total_revenue, total_profit, total_customers, average_order_value, total_items_sold, cash_sales, card_sales, debt_sales, updated_at) VALUES (?, ?, 5, 7300, 1940, 4, 1460, 6, 2500, 980, 3820, ?)`, now.Year(), int(now.Month()), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert monthly sales summary %s: %w", month, err))
	}
	if _, err := tx.Exec(`INSERT INTO monthly_profit_summary (year, month, gross_profit, net_profit, total_revenue, total_cost, profit_margin, updated_at) VALUES (?, ?, 1940, 1885, 7300, 5360, 25.8904, ?)`, now.Year(), int(now.Month()), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert monthly profit summary: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO monthly_debt_summary (year, month, total_debt, new_debt, payments_received, overdue_debt, overdue_count, paid_debt, updated_at) VALUES (?, ?, 2820, 2820, 1000, 450, 1, 0, ?)`, now.Year(), int(now.Month()), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert monthly debt summary: %w", err))
	}

	// A couple of notifications exercise the dashboard notification list.
	if ownerID != "" {
		for _, row := range []struct{ typ, title, message, priority string }{
			{"LOW_STOCK", "Low stock alert", "RTX 4070 SUPER and used ThinkPad stock are near the reorder level.", "high"},
			{"DEBT_OVERDUE", "Overdue customer balance", "Rana Haddad has an overdue balance of ILS 450.", "normal"},
		} {
			if _, err := tx.Exec(`INSERT INTO notifications (id, user_id, type, title, message, priority, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'unread', ?, ?)`, uuid.NewString(), ownerID, row.typ, row.title, row.message, row.priority, stamp(now), stamp(now)); err != nil {
				return rollback(fmt.Errorf("insert notification: %w", err))
			}
		}
	}

	// Supplemental records exercise every non-CRUD screen: barcodes, serials,
	// reservations, specifications, inspections, warranties, returns, ledgers,
	// reports and the held-sale workflow. All references point to rows created
	// in this same transaction so the dataset remains internally consistent.
	for _, row := range []struct{ itemCode, barcode, serial string }{
		{"IT-CPU-001", "89010000000101", "SN-CPU-2026-001"},
		{"IT-GPU-002", "89010000000202", "SN-GPU-2026-002"},
		{"IT-UGPU-002", "89010000000502", "SN-UGPU-2026-002"},
		{"IT-LAP-001", "89010000000601", "SN-LAP-2026-001"},
	} {
		if _, err := tx.Exec(`UPDATE inventory_items SET barcode = ?, serial_number = ? WHERE item_code = ?`, row.barcode, row.serial, row.itemCode); err != nil {
			return rollback(fmt.Errorf("set serial metadata %s: %w", row.itemCode, err))
		}
	}
	for i, row := range products {
		if _, err := tx.Exec(`INSERT INTO barcodes (id, code, product_id, type, is_active, generated_at, created_at, updated_at) VALUES (?, ?, ?, 'PRODUCT', 1, ?, ?, ?)`, uuid.NewString(), fmt.Sprintf("8901000000%02d", i+1), row.id, stamp(now), stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert product barcode %s: %w", row.sku, err))
		}
	}
	for _, row := range []struct{ itemCode, barcode, typ string }{
		{"IT-CPU-001", "89010000000101", "SERIAL"},
		{"IT-GPU-002", "89010000000202", "SERIAL"},
		{"IT-UGPU-002", "89010000000502", "SERIAL"},
		{"IT-LAP-001", "89010000000601", "SERIAL"},
	} {
		var itemID, productID string
		if err := tx.QueryRow(`SELECT id, product_id FROM inventory_items WHERE item_code = ?`, row.itemCode).Scan(&itemID, &productID); err != nil {
			return rollback(fmt.Errorf("find barcode item %s: %w", row.itemCode, err))
		}
		if _, err := tx.Exec(`INSERT INTO barcodes (id, code, product_id, inventory_item_id, type, is_active, generated_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?)`, uuid.NewString(), row.barcode, productID, itemID, row.typ, stamp(now), stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert item barcode %s: %w", row.itemCode, err))
		}
	}

	specCapacity := uuid.NewString()
	specVram := uuid.NewString()
	specHealth := uuid.NewString()
	for _, row := range []struct {
		id, ar, en, dataType, options string
		required                      bool
	}{
		{specCapacity, "السعة", "Capacity", "number", `["8","16","32","64"]`, true},
		{specVram, "ذاكرة الرسوميات", "VRAM (GB)", "number", `["8","12","16","24"]`, false},
		{specHealth, "حالة القطعة", "Condition notes", "text", `[]`, false},
	} {
		if _, err := tx.Exec(`INSERT INTO part_specifications (id, name_ar, name_en, data_type, options, is_required, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, row.id, row.ar, row.en, row.dataType, row.options, row.required, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert specification %s: %w", row.en, err))
		}
	}
	for _, row := range []struct {
		partTypeID, specID string
		order              int
	}{
		{partTypeIDs[0], specCapacity, 1}, {partTypeIDs[0], specVram, 2},
		{partTypeIDs[1], specHealth, 1}, {partTypeIDs[1], specCapacity, 2},
	} {
		if _, err := tx.Exec(`INSERT INTO type_specifications (id, part_type_id, specification_id, sort_order, created_at) VALUES (?, ?, ?, ?, ?)`, uuid.NewString(), row.partTypeID, row.specID, row.order, stamp(now)); err != nil {
			return rollback(fmt.Errorf("link specification: %w", err))
		}
	}
	for _, row := range []struct {
		itemCode, specID, textValue string
		numberValue                 float64
	}{
		{"IT-GPU-002", specVram, "", 12},
		{"IT-RAM-002", specCapacity, "", 32},
		{"IT-UGPU-002", specVram, "", 12},
		{"IT-LAP-001", specHealth, "Battery health 92%", 0},
	} {
		var itemID string
		if err := tx.QueryRow(`SELECT id FROM inventory_items WHERE item_code = ?`, row.itemCode).Scan(&itemID); err != nil {
			return rollback(fmt.Errorf("find specification item %s: %w", row.itemCode, err))
		}
		if _, err := tx.Exec(`INSERT INTO item_specification_values (id, inventory_item_id, specification_id, value_text, value_number, created_at, updated_at) VALUES (?, ?, ?, ?, NULLIF(?, 0), ?, ?)`, uuid.NewString(), itemID, row.specID, nullable(row.textValue), row.numberValue, stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert item specification %s: %w", row.itemCode, err))
		}
	}

	var reservedItemID string
	if err := tx.QueryRow(`SELECT id FROM inventory_items WHERE item_code = 'IT-GPU-002'`).Scan(&reservedItemID); err != nil {
		return rollback(fmt.Errorf("find reserved item: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO reservations (id, item_id, customer_id, user_id, reserved_at, expires_at, status, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?, ?)`, uuid.NewString(), reservedItemID, customers[0].id, ownerID, stamp(now.Add(-30*time.Minute)), stamp(now.AddDate(0, 0, 2)), "GPU reserved for Ahmad's workstation build", stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert active reservation: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO reservations (id, item_id, customer_id, user_id, reserved_at, expires_at, status, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'expired', ?, ?, ?)`, uuid.NewString(), reservedItemID, customers[1].id, ownerID, stamp(now.AddDate(0, 0, -3)), stamp(now.AddDate(0, 0, -1)), "Expired reservation for testing", stamp(now.AddDate(0, 0, -3)), stamp(now.AddDate(0, 0, -3))); err != nil {
		return rollback(fmt.Errorf("insert expired reservation: %w", err))
	}

	// Inspection checkpoints, repair history and warranty claims make used
	// parts and after-sales flows visible in their dedicated screens.
	var gpuInspectionID, laptopInspectionID string
	if err := tx.QueryRow(`SELECT id FROM inspections WHERE product_id = ? ORDER BY inspection_date DESC LIMIT 1`, productBySKU["GPU-USED-001"].id).Scan(&gpuInspectionID); err != nil {
		return rollback(fmt.Errorf("find GPU inspection: %w", err))
	}
	if err := tx.QueryRow(`SELECT id FROM inspections WHERE product_id = ? ORDER BY inspection_date DESC LIMIT 1`, productBySKU["LAP-USED-001"].id).Scan(&laptopInspectionID); err != nil {
		return rollback(fmt.Errorf("find laptop inspection: %w", err))
	}
	for _, row := range []struct{ inspectionID, itemCode, checkpoint, status, notes string }{
		{gpuInspectionID, "IT-UGPU-001", "GPU stress test", "passed", "30-minute FurMark test completed"},
		{gpuInspectionID, "IT-UGPU-001", "Ports and fans", "passed", "All outputs and fans working"},
		{laptopInspectionID, "IT-LAP-001", "Battery health", "passed", "Battery capacity above 90%"},
		{laptopInspectionID, "IT-LAP-001", "Keyboard and display", "passed", "No dead pixels or stuck keys"},
	} {
		var itemID string
		if err := tx.QueryRow(`SELECT id FROM inventory_items WHERE item_code = ?`, row.itemCode).Scan(&itemID); err != nil {
			return rollback(fmt.Errorf("find checkpoint item %s: %w", row.itemCode, err))
		}
		if _, err := tx.Exec(`INSERT INTO inspection_items (id, inspection_id, item_id, checkpoint_name, status, notes, images, created_at) VALUES (?, ?, ?, ?, ?, ?, '[]', ?)`, uuid.NewString(), row.inspectionID, itemID, row.checkpoint, row.status, row.notes, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert inspection checkpoint: %w", err))
		}
	}
	var repairItemID, repairAcquisitionID string
	if err := tx.QueryRow(`SELECT id FROM inventory_items WHERE item_code = 'IT-LAP-002'`).Scan(&repairItemID); err != nil {
		return rollback(fmt.Errorf("find repair item: %w", err))
	}
	if err := tx.QueryRow(`SELECT id FROM acquisition_items WHERE item_code = 'IT-LAP-002'`).Scan(&repairAcquisitionID); err != nil {
		return rollback(fmt.Errorf("find repair acquisition item: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO item_repair_costs (id, inventory_item_id, acquisition_item_id, repair_date, repair_type, cost, description, performed_by, created_at) VALUES (?, ?, ?, ?, 'screen_replacement', 180, 'Replace cracked LCD panel', 'Tech bench', ?)`, uuid.NewString(), repairItemID, repairAcquisitionID, yesterday, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert repair cost: %w", err))
	}
	for _, row := range []struct{ itemCode, eventType, eventDate, refType, refID, description string }{
		{"IT-UGPU-001", "acquired", today, "acquisition", acq1, "GPU purchased from Omar Saleh"},
		{"IT-UGPU-001", "inspected", today, "inspection", gpuInspectionID, "GPU passed diagnostics"},
		{"IT-UGPU-001", "sold", today, "sale", sale2, "Sold to Maya Khalil"},
		{"IT-LAP-002", "repair_started", date(now.AddDate(0, 0, -2)), "acquisition", acq3, "Screen repair pending"},
	} {
		var itemID string
		if err := tx.QueryRow(`SELECT id FROM inventory_items WHERE item_code = ?`, row.itemCode).Scan(&itemID); err != nil {
			return rollback(fmt.Errorf("find history item %s: %w", row.itemCode, err))
		}
		if _, err := tx.Exec(`INSERT INTO item_history (id, inventory_item_id, event_type, event_date, reference_type, reference_id, description, metadata, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, '{}', ?, ?)`, uuid.NewString(), itemID, row.eventType, row.eventDate, row.refType, row.refID, row.description, ownerID, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert item history: %w", err))
		}
	}
	for _, row := range []struct{ id, productID, customerID, serial, issue, status string }{
		{uuid.NewString(), productBySKU["CPU-NEW-001"].id, customers[0].id, "SN-CPU-2026-001", "CPU intermittently throttles under load", "pending"},
		{uuid.NewString(), productBySKU["LAP-USED-001"].id, customers[3].id, "SN-LAP-2026-001", "Keyboard replacement completed", "completed"},
	} {
		if _, err := tx.Exec(`INSERT INTO warranty_claims (id, claim_number, customer_id, product_id, serial_number, issue_description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, row.id, "WAR-TEST-"+row.id[:8], row.customerID, row.productID, row.serial, row.issue, row.status, stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert warranty claim: %w", err))
		}
	}

	// Customer/supplier collection tables back the debt detail screens in
	// addition to the primary debts/payments tables.
	if _, err := tx.Exec(`INSERT INTO customer_debts (id, customer_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at) VALUES (?, ?, 2370, ?, 'sale', ?, 0, 1000, ?)`, uuid.NewString(), customers[2].id, sale3, fourteenDaysAhead, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert customer debt detail: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO customer_debts (id, customer_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at) VALUES (?, ?, 450, ?, 'sale', ?, 0, 0, ?)`, uuid.NewString(), customers[3].id, sale5, date(now.AddDate(0, 0, -2)), stamp(now.AddDate(0, 0, -10))); err != nil {
		return rollback(fmt.Errorf("insert overdue debt detail: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO customer_payments (id, customer_id, amount, payment_date, method, reference, notes, created_at) VALUES (?, ?, 1000, ?, 'cash', ?, 'Partial payment for graphics workstation', ?)`, uuid.NewString(), customers[2].id, today, "S-TEST-003", stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert customer payment detail: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO debt_collections (id, customer_id, type, status, notes, scheduled_date, created_at) VALUES (?, ?, 'call', 'pending', 'Call customer about overdue SSD invoice', ?, ?)`, uuid.NewString(), customers[3].id, today, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert debt collection: %w", err))
	}

	var purchaseItemID string
	if err := tx.QueryRow(`SELECT id FROM purchase_items WHERE purchase_id = ? ORDER BY created_at LIMIT 1`, purchase3).Scan(&purchaseItemID); err != nil {
		return rollback(fmt.Errorf("find supplier return item: %w", err))
	}
	supplierReturnID := uuid.NewString()
	if _, err := tx.Exec(`INSERT INTO supplier_returns (id, purchase_id, supplier_id, return_number, status, reason, refund_amount, notes, created_by, created_at, updated_at) VALUES (?, ?, ?, 'SRET-TEST-001', 'PENDING', 'Damaged laptop panel', 850, 'Awaiting supplier credit note', ?, ?, ?)`, supplierReturnID, purchase3, suppliers[0].id, ownerID, stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert supplier return: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO supplier_return_items (id, supplier_return_id, purchase_item_id, product_id, quantity, unit_cost, created_at) VALUES (?, ?, ?, ?, 1, 850, ?)`, uuid.NewString(), supplierReturnID, purchaseItemID, productBySKU["LAP-USED-001"].id, stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert supplier return item: %w", err))
	}

	// Saved settings and a held cart ensure settings, POS and notification
	// components have meaningful local state immediately after login.
	for _, row := range []struct{ key, value, valueType, category, description string }{
		{"store_name", "PartFlow Computer Store", "string", "general", "Store display name"},
		{"currency", "ILS", "string", "general", "Default currency"},
		{"tax_rate", "0", "number", "financial", "Tax rate for test data"},
		{"default_profit_margin", "25", "number", "financial", "Default margin percentage"},
	} {
		if _, err := tx.Exec(`INSERT OR REPLACE INTO settings (id, key, value, value_type, category, description, is_public, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`, "setting-"+row.key, row.key, row.value, row.valueType, row.category, row.description, stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert setting %s: %w", row.key, err))
		}
	}
	if _, err := tx.Exec(`INSERT OR REPLACE INTO notification_preferences (id, user_id, email_enabled, push_enabled, low_stock, debt_overdue, return_requests, expense_approval, sales_updates, created_at, updated_at) VALUES (?, ?, 1, 1, 1, 1, 1, 1, 1, ?, ?)`, uuid.NewString(), ownerID, stamp(now), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert notification preferences: %w", err))
	}
	if _, err := tx.Exec(`INSERT INTO held_sales (id, user_id, items, created_at) VALUES (?, ?, ?, ?)`, uuid.NewString(), ownerID, fmt.Sprintf(`{"customer_id":"%s","items":[{"product_id":"%s","quantity":1}]}`, customers[0].id, productBySKU["GPU-NEW-001"].id), stamp(now)); err != nil {
		return rollback(fmt.Errorf("insert held sale: %w", err))
	}

	// Persist generated report snapshots for report history and audit views.
	for _, row := range []struct{ typ, title, description, data string }{
		{"sales", "August sales overview", "Sales by payment method and day", fmt.Sprintf(`{"total_revenue":7300,"total_sales":5,"currency":"ILS","generated_for":"%s"}`, month)},
		{"inventory", "Current inventory valuation", "Available, sold and repair stock", `{"available_items":8,"inventory_value":8370}`},
		{"debts", "Open customer balances", "Pending and overdue debts", `{"outstanding":2820,"overdue":450}`},
		{"profit", "Monthly profit statement", "Gross margin less operating expenses", `{"gross_profit":1940,"net_profit":1885}`},
	} {
		if _, err := tx.Exec(`INSERT INTO reports (id, type, title, description, parameters, data, status, generated_by, generated_at, created_at, updated_at) VALUES (?, ?, ?, ?, '{}', ?, 'completed', ?, ?, ?, ?)`, uuid.NewString(), row.typ, row.title, row.description, row.data, ownerID, stamp(now), stamp(now), stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert report %s: %w", row.typ, err))
		}
	}
	for _, row := range []struct{ action, entityType, entityID, description string }{
		{"CREATE", "sale", sale1, "Created cash sale S-TEST-001"},
		{"CREATE", "sale", sale3, "Created partial-credit sale S-TEST-003"},
		{"CREATE", "purchase", purchase1, "Received purchase PO-TEST-001"},
		{"UPDATE", "inventory_item", repairItemID, "Moved laptop to repair"},
	} {
		if _, err := tx.Exec(`INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, description, status, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?, 'success', '{}', ?)`, uuid.NewString(), ownerID, row.action, row.entityType, row.entityID, row.description, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert audit log: %w", err))
		}
	}
	for _, row := range []struct {
		ledgerType, entityID, transactionType, refID, refType, description string
		amount, balance, previous                                          float64
	}{
		{"customer", customers[2].id, "sale", sale3, "sale", "Credit sale S-TEST-003", 3370, 3370, 0},
		{"customer", customers[2].id, "payment", paymentCustomer, "payment", "Cash payment received", -1000, 2370, 3370},
		{"supplier", suppliers[0].id, "purchase", purchase1, "purchase", "Purchase PO-TEST-001", 4000, 4000, 0},
		{"supplier", suppliers[0].id, "payment", paymentSupplier, "payment", "Supplier payment sent", -2500, 1500, 4000},
	} {
		if _, err := tx.Exec(`INSERT INTO ledger_entries (id, ledger_type, entity_id, transaction_type, reference_id, reference_type, amount, balance, previous_balance, description, metadata, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, ?)`, uuid.NewString(), row.ledgerType, row.entityID, row.transactionType, row.refID, row.refType, row.amount, row.balance, row.previous, row.description, ownerID, stamp(now)); err != nil {
			return rollback(fmt.Errorf("insert ledger entry: %w", err))
		}
	}

	if err := tx.Commit(); err != nil {
		_, _ = db.Exec("PRAGMA foreign_keys = ON")
		return fmt.Errorf("commit test dataset: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("restore foreign keys: %w", err)
	}
	return nil
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
