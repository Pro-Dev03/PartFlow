package sales

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestCreateCashSaleStoresAppliedPaymentAndChangeSeparatelySQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/cash-change.db")
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	productID, itemID := uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "CASH-CHANGE-001", "Cash Change Product", 80, 125, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, itemID, productID, "CASH-CHANGE-ITEM", "CASH-CHANGE-BC", "NEW", 80, 125, "AVAILABLE", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, uuid.New(), productID, 1, now, now); err != nil {
		t.Fatal(err)
	}

	method := "cash"
	sale, err := NewService(NewRepository(db), db).CreateSale(ctx, uuid.Nil, &CreateSaleRequest{
		Items:         []SaleItemRequest{{ProductID: productID, InventoryItemID: &itemID, Quantity: 1, UnitPrice: 125}},
		PaymentMethod: &method,
		PaymentAmount: 125,
		CashReceived:  150,
	})
	if err != nil {
		t.Fatalf("CreateSale failed: %v", err)
	}
	if sale.TotalAmount != 125 || sale.PaidAmount != 125 || sale.CashReceived != 150 || sale.ChangeAmount != 25 {
		t.Fatalf("sale amounts = total=%v paid=%v received=%v change=%v", sale.TotalAmount, sale.PaidAmount, sale.CashReceived, sale.ChangeAmount)
	}
	var storedPaid, storedReceived, storedChange, paymentAmount float64
	if err := db.QueryRow(`SELECT paid_amount, cash_received, change_amount FROM sales WHERE id = ?`, sale.ID).Scan(&storedPaid, &storedReceived, &storedChange); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&paymentAmount, `SELECT amount FROM payments WHERE sale_id = ?`, sale.ID); err != nil {
		t.Fatal(err)
	}
	if storedPaid != 125 || storedReceived != 150 || storedChange != 25 || paymentAmount != 125 {
		t.Fatalf("stored amounts = paid=%v received=%v change=%v payment=%v", storedPaid, storedReceived, storedChange, paymentAmount)
	}
	if sale.GrossProfit != 45 || sale.NetProfit != 45 {
		t.Fatalf("profit = gross=%v net=%v, want 45", sale.GrossProfit, sale.NetProfit)
	}
}
