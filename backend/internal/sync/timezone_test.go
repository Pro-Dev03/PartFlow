package sync

import (
	"testing"
	"time"
)

func TestNormalizeCloudPayloadCanonicalizesUTCInstants(t *testing.T) {
	localOffset := time.FixedZone("legacy-local", 3*60*60)
	payload := map[string]any{
		"created_at": time.Date(2026, time.September, 16, 23, 59, 0, 0, localOffset),
		"updated_at": "2026-09-16 23:59:00 +0300 EEST",
		"sale_date":  "2026-09-16",
	}

	NormalizeCloudPayload("sales", payload)
	if got := payload["created_at"]; got != "2026-09-16T20:59:00Z" {
		t.Fatalf("created_at = %#v, want UTC instant 2026-09-16T20:59:00Z", got)
	}
	if got := payload["updated_at"]; got != "2026-09-16T20:59:00Z" {
		t.Fatalf("updated_at = %#v, want UTC instant 2026-09-16T20:59:00Z", got)
	}
	if got := payload["sale_date"]; got != "2026-09-16" {
		t.Fatalf("sale_date = %#v, want business date preserved", got)
	}
}

func TestNormalizeCloudPayloadTreatsNaiveTimestampAsUTC(t *testing.T) {
	payload := map[string]any{"updated_at": "2026-09-16 23:59:00"}
	NormalizeCloudPayload("products", payload)
	if got := payload["updated_at"]; got != "2026-09-16T23:59:00Z" {
		t.Fatalf("updated_at = %#v, want naive UTC timestamp", got)
	}
}

func TestNormalizeCloudPayloadUsesCloudPurchaseColumnsAndBooleans(t *testing.T) {
	item := map[string]any{"id": "item-id", "quantity": float64(2), "unit_cost": float64(10), "item_total": float64(20)}
	NormalizeCloudPayload("purchase_items", item)
	if item["unit_price"] != float64(10) || item["total_amount"] != float64(20) {
		t.Fatalf("purchase item aliases not mapped: %#v", item)
	}
	if _, present := item["unit_cost"]; present {
		t.Fatalf("obsolete unit_cost remains in Cloud payload: %#v", item)
	}
	if _, present := item["item_total"]; present {
		t.Fatalf("local-only item_total remains in Cloud payload: %#v", item)
	}
	product := map[string]any{"is_active": float64(1), "track_serial": float64(0)}
	NormalizeCloudPayload("products", product)
	if product["is_active"] != true || product["track_serial"] != false {
		t.Fatalf("SQLite booleans not normalized for Cloud: %#v", product)
	}
}
