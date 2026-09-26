package settings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestSyncCloudDataMergesCloudSnapshotIntoSQLite(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "sync.db")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", localPath)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync/initial-data" {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer cloud-token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"customers": []map[string]any{{
					"id": "cust-1", "code": "C-001", "name": "Ahmad",
					"created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
				}},
				"products": []map[string]any{{
					"id": "prod-1", "sku": "SKU-1", "name": "Oil Filter", "barcode": "123",
					"created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
				}},
				"sales": []map[string]any{{
					"id": "sale-1", "invoice_number": "INV-1", "customer_id": "cust-1",
					"total_amount": 80, "paid_amount": 50, "remaining_amount": 30,
					"status": "completed", "created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
				}},
				"debts": []map[string]any{{
					"id": "debt-1", "customer_id": "cust-1", "sale_id": "sale-1", "amount": 30,
					"paid_amount": 0, "remaining_amount": 30, "status": "pending",
					"due_date": "2026-09-30", "created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
				}},
				"payments": []map[string]any{{
					"id": "payment-1", "reference_number": "PAY-1", "customer_id": "cust-1", "amount": 50,
					"payment_method": "cash", "payment_date": "2026-08-30", "created_at": "2026-08-30T00:00:00Z",
				}},
				"expenses": []map[string]any{{
					"id": "expense-1", "description": "Monthly rent", "amount": 100,
					"expense_date": "2026-08-30", "created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
				}},
				"customer_ledger": []map[string]any{{
					"id": "customer-ledger-1", "customer_id": "cust-1", "type": "debit", "amount": 30,
					"balance": 30, "description": "Invoice", "created_at": "2026-08-30T00:00:00Z",
				}},
			},
		})
	}))
	defer server.Close()
	t.Setenv("PARTFLOW_CLOUD_API_URL", server.URL)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request, err := http.NewRequest(http.MethodPost, "/settings/sync", nil)
	if err != nil {
		t.Fatalf("create sync request: %v", err)
	}
	ctx.Request = request
	ctx.Request.Header.Set("Authorization", "Bearer test-token")
	ctx.Request.Header.Set("X-PartFlow-Cloud-Token", "cloud-token")
	ctx.Request.Header.Set("X-PartFlow-Cloud-API-URL", "https://attacker.example/api/v1")
	NewLocalDatabaseHandler().SyncCloudData(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", recorder.Code, recorder.Body.String())
	}

	database, err := localdb.Open()
	if err != nil {
		t.Fatalf("reopen local database: %v", err)
	}
	defer database.DB.Close()
	var customerCount, productCount, saleCount, debtCount, paymentCount, expenseCount, ledgerCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM customers WHERE id = ?", "cust-1").Scan(&customerCount); err != nil {
		t.Fatalf("read synced customer: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM products WHERE id = ?", "prod-1").Scan(&productCount); err != nil {
		t.Fatalf("read synced product: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM sales WHERE id = ?", "sale-1").Scan(&saleCount); err != nil {
		t.Fatalf("read synced sale: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM debts WHERE id = ?", "debt-1").Scan(&debtCount); err != nil {
		t.Fatalf("read synced debt: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM payments WHERE id = ?", "payment-1").Scan(&paymentCount); err != nil {
		t.Fatalf("read synced payment: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM expenses WHERE id = ?", "expense-1").Scan(&expenseCount); err != nil {
		t.Fatalf("read synced expense: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM customer_ledger WHERE id = ?", "customer-ledger-1").Scan(&ledgerCount); err != nil {
		t.Fatalf("read synced customer ledger: %v", err)
	}
	if customerCount != 1 || productCount != 1 || saleCount != 1 || debtCount != 1 || paymentCount != 1 || expenseCount != 1 || ledgerCount != 1 {
		t.Fatalf("synced rows customer=%d product=%d sale=%d debt=%d payment=%d expense=%d ledger=%d, want all 1", customerCount, productCount, saleCount, debtCount, paymentCount, expenseCount, ledgerCount)
	}
	lastSync, err := localdb.GetMetadata(database.DB, "last_cloud_sync_at")
	if err != nil || lastSync == "" {
		t.Fatalf("last cloud sync metadata missing: value=%q err=%v", lastSync, err)
	}
}

func TestOwnerDatabaseSyncPushesNewerLocalRowsAndPullsCloudRows(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "owner-sync.db")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", localPath)
	localDB, err := localdb.Open()
	if err != nil {
		t.Fatalf("open local database: %v", err)
	}
	localTime := "2026-09-27T10:00:00Z"
	if _, err := localDB.DB.Exec(`
		INSERT INTO customers (id, code, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "local-customer", "C-LOCAL", "Local Customer", localTime, localTime); err != nil {
		localDB.DB.Close()
		t.Fatalf("insert local customer: %v", err)
	}
	if err := localDB.DB.Close(); err != nil {
		t.Fatalf("close local database: %v", err)
	}

	localCustomer := map[string]any(nil)
	initialDataRequests := 0
	pushRequests := 0
	cloudServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer cloud-token" {
			http.Error(w, "missing cloud authorization", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/sync/initial-data":
			initialDataRequests++
			customers := []map[string]any{{
				"id": "cloud-customer", "code": "C-CLOUD", "name": "Cloud Customer",
				"is_active": true, "created_at": "2026-09-27T09:00:00Z", "updated_at": "2026-09-27T09:00:00Z",
			}}
			if localCustomer != nil {
				customers = append(customers, localCustomer)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data":    map[string]any{"customers": customers},
			})
		case "/sync/push":
			pushRequests++
			var payload struct {
				Operations []struct {
					ID      string `json:"id"`
					Payload string `json:"payload"`
				} `json:"operations"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.Operations) != 1 {
				http.Error(w, "expected one local operation", http.StatusBadRequest)
				return
			}
			if err := json.Unmarshal([]byte(payload.Operations[0].Payload), &localCustomer); err != nil {
				http.Error(w, "invalid operation payload", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"data": map[string]any{
					"accepted_ids": []string{payload.Operations[0].ID},
					"processed":    1,
					"failed":       0,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer cloudServer.Close()
	t.Setenv("PARTFLOW_CLOUD_API_URL", cloudServer.URL)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/settings/database/sync", nil)
	ctx.Request.Header.Set("Authorization", "Bearer local-token")
	ctx.Request.Header.Set("X-PartFlow-Cloud-Token", "cloud-token")
	NewLocalDatabaseHandler().SyncLocalDataToCloud(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("sync status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if pushRequests != 1 || initialDataRequests != 2 {
		t.Fatalf("cloud requests: pushes=%d snapshots=%d; want one push and two snapshots", pushRequests, initialDataRequests)
	}

	resultDB, err := localdb.Open()
	if err != nil {
		t.Fatalf("reopen local database: %v", err)
	}
	defer resultDB.DB.Close()
	var localCount, cloudCount int
	if err := resultDB.DB.QueryRow(`SELECT COUNT(*) FROM customers WHERE id='local-customer'`).Scan(&localCount); err != nil {
		t.Fatalf("read synchronized local customer: %v", err)
	}
	if err := resultDB.DB.QueryRow(`SELECT COUNT(*) FROM customers WHERE id='cloud-customer'`).Scan(&cloudCount); err != nil {
		t.Fatalf("read synchronized cloud customer: %v", err)
	}
	if localCount != 1 || cloudCount != 1 {
		t.Fatalf("synchronized customer counts local=%d cloud=%d; want both rows present", localCount, cloudCount)
	}
}
