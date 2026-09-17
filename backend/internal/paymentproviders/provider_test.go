package paymentproviders

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCardcomAdapterCreateVerifyAndRefund(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		paths = append(paths, request.URL.Path)
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/LowProfile/Create":
			_ = json.NewEncoder(writer).Encode(map[string]any{"ResponseCode": 0, "LowProfileId": "lp-1", "Url": "https://cardcom.test/pay/lp-1"})
		case "/LowProfile/GetLpResult":
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"ResponseCode":    0,
				"LowProfileId":    "lp-1",
				"TranzactionInfo": map[string]any{"ResponseCode": 0, "Amount": 12.5, "TranzactionId": 9876},
			})
		case "/Transactions/RefundByTransactionId":
			_ = json.NewEncoder(writer).Encode(map[string]any{"ResponseCode": 0, "NewTranzactionId": 9999})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	adapter := NewCardcomAdapter(Config{
		Provider:    ProviderCardcom,
		BaseURL:     server.URL,
		TerminalID:  "1000",
		APIKey:      "api-user",
		APISecret:   "api-secret",
		HTTPTimeout: time.Second,
	})
	created, err := adapter.CreatePayment(context.Background(), PaymentRequest{
		PaymentID: uuid.New(), OrderID: uuid.New(), Amount: 1250, Currency: "ILS", Description: "Test sale", IdempotencyKey: "sale-1",
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if created.Status != StatusPending || created.ProviderPaymentID != "lp-1" || created.CheckoutURL == "" {
		t.Fatalf("unexpected create result: %#v", created)
	}

	verified, err := adapter.VerifyPayment(context.Background(), StatusRequest{ProviderPaymentID: "lp-1"})
	if err != nil {
		t.Fatalf("VerifyPayment failed: %v", err)
	}
	if verified.Status != StatusPaid || verified.ProviderTransactionID != "9876" || verified.Amount != 1250 {
		t.Fatalf("unexpected verify result: %#v", verified)
	}

	refunded, err := adapter.RefundPayment(context.Background(), RefundRequest{ProviderTransactionID: verified.ProviderTransactionID, Amount: 500, Currency: "ILS", Reason: "return-1"})
	if err != nil {
		t.Fatalf("RefundPayment failed: %v", err)
	}
	if refunded.Status != StatusPartiallyRefunded || refunded.ProviderRefundID != "9999" {
		t.Fatalf("unexpected refund result: %#v", refunded)
	}
	if len(paths) != 3 {
		t.Fatalf("provider request count = %d, want 3", len(paths))
	}
}

func TestCanTransitionProtectsPaymentLifecycle(t *testing.T) {
	valid := [][2]string{{"", StatusPending}, {StatusPending, StatusProcessing}, {StatusProcessing, StatusPaid}, {StatusPaid, StatusPartiallyRefunded}, {StatusPartiallyRefunded, StatusRefunded}}
	for _, transition := range valid {
		if !CanTransition(transition[0], transition[1]) {
			t.Fatalf("expected valid transition %q -> %q", transition[0], transition[1])
		}
	}
	invalid := [][2]string{{StatusPaid, StatusFailed}, {StatusFailed, StatusPaid}, {StatusRefunded, StatusPaid}, {StatusCancelled, StatusProcessing}}
	for _, transition := range invalid {
		if CanTransition(transition[0], transition[1]) {
			t.Fatalf("expected invalid transition %q -> %q", transition[0], transition[1])
		}
	}
}

func TestCardcomAdapterProviderFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "provider unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	adapter := NewCardcomAdapter(Config{BaseURL: server.URL, TerminalID: "1000", APIKey: "api", APISecret: "secret"})
	if _, err := adapter.CreatePayment(context.Background(), PaymentRequest{OrderID: uuid.New(), Amount: 100, Currency: "ILS"}); err == nil {
		t.Fatal("expected provider failure")
	}
}

func TestCardcomAdapterTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(100 * time.Millisecond)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter := NewCardcomAdapter(Config{BaseURL: server.URL, TerminalID: "1000", APIKey: "api", APISecret: "secret", HTTPTimeout: 10 * time.Millisecond})
	if _, err := adapter.CreatePayment(context.Background(), PaymentRequest{OrderID: uuid.New(), Amount: 100, Currency: "ILS"}); err == nil {
		t.Fatal("expected provider timeout")
	}
}

func TestRegistrySeparatesImplementedAndPlannedProviders(t *testing.T) {
	registry := NewRegistry()
	adapter, err := registry.Build(ProviderCardcom, Config{TerminalID: "1000", APIKey: "api", APISecret: "secret"})
	if err != nil || adapter.Name() != ProviderCardcom {
		t.Fatalf("Cardcom registry entry = %v, %v", adapter, err)
	}
	if _, err := registry.Build(ProviderStripe, Config{}); err == nil {
		t.Fatal("expected unimplemented Stripe adapter to be rejected")
	}
	if len(registry.Descriptors()) != 6 {
		t.Fatalf("provider descriptor count = %d, want 6", len(registry.Descriptors()))
	}
}
