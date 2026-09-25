package debts

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestListDebtsAppliesFiltersBeforePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, code TEXT, phone TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, invoice_number TEXT);
		CREATE TABLE debts (
			id TEXT PRIMARY KEY, customer_id TEXT, sale_id TEXT, amount REAL,
			remaining_amount REAL, due_date TEXT, status TEXT, created_at TEXT
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	type testCustomer struct {
		id     uuid.UUID
		name   string
		amount float64
		left   float64
		due    string
		status string
	}
	customers := []testCustomer{
		{id: uuid.New(), name: "Target One", amount: 200, left: 200, due: "2026-09-10", status: "overdue"},
		{id: uuid.New(), name: "Target Two", amount: 100, left: 50, due: "2026-09-24", status: "partial"},
		{id: uuid.New(), name: "Target Three", amount: 100, left: 100, due: "2026-09-24", status: "overdue"},
		{id: uuid.New(), name: "Paid Customer", amount: 40, left: 0, due: "2026-09-01", status: "paid"},
	}
	for _, customer := range customers {
		if _, err := db.Exec(`INSERT INTO customers (id, name, code, phone) VALUES (?, ?, ?, ?)`, customer.id.String(), customer.name, "C-"+customer.id.String()[:4], "0590000000"); err != nil {
			t.Fatal(err)
		}
		saleID := uuid.New()
		if _, err := db.Exec(`INSERT INTO sales (id, invoice_number) VALUES (?, ?)`, saleID.String(), "INV-"+customer.name); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO debts (id, customer_id, sale_id, amount, remaining_amount, due_date, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New().String(), customer.id.String(), saleID.String(), customer.amount, customer.left, customer.due, customer.status, "2026-09-20T10:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}

	handler := NewHandler(db)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/debts?page=1&per_page=1&search=Target+Three&search_type=name&status=overdue&tab=open&min_amount=80&max_amount=150&due_date_from=2026-09-24&due_date_to=2026-09-24", nil)
	handler.ListDebts(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Data []struct {
			CustomerName string  `json:"customer_name"`
			Amount       float64 `json:"amount"`
		} `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Meta.Total != 1 || len(response.Data) != 1 {
		t.Fatalf("filtered total/results = %d/%d, want 1/1", response.Meta.Total, len(response.Data))
	}
	if response.Data[0].CustomerName != "Target Three" || response.Data[0].Amount != 100 {
		t.Fatalf("filtered row = %+v, want Target Three with amount 100", response.Data[0])
	}

	recorder = httptest.NewRecorder()
	context, _ = gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/debts?page=1&per_page=10&tab=paid", nil)
	handler.ListDebts(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("paid-tab status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var paidResponse struct {
		Data []struct {
			CustomerName string `json:"customer_name"`
			Status       string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &paidResponse); err != nil {
		t.Fatal(err)
	}
	if len(paidResponse.Data) != 1 || paidResponse.Data[0].CustomerName != "Paid Customer" || paidResponse.Data[0].Status != "paid" {
		t.Fatalf("paid tab rows = %+v, want only Paid Customer", paidResponse.Data)
	}

	targetCustomerID := customers[2].id.String()
	recorder = httptest.NewRecorder()
	context, _ = gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/debts?page=1&per_page=1&customer_id="+targetCustomerID+"&tab=open", nil)
	handler.ListDebts(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("customer-filter status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var customerResponse struct {
		Data []struct {
			CustomerName string `json:"customer_name"`
		} `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &customerResponse); err != nil {
		t.Fatal(err)
	}
	if customerResponse.Meta.Total != 1 || len(customerResponse.Data) != 1 || customerResponse.Data[0].CustomerName != "Target Three" {
		t.Fatalf("customer filter response = %+v, want only Target Three", customerResponse)
	}
}
