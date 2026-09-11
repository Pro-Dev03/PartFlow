package dashboard

import "testing"

func TestInventoryStatusPresentationSeparatesReservedAndReturned(t *testing.T) {
	reservedName, _, reservedHealth := inventoryStatusPresentation("RESERVED")
	returnedName, _, returnedHealth := inventoryStatusPresentation("RETURNED")

	if reservedName != "محجوز" || reservedHealth != "low" {
		t.Fatalf("reserved presentation = (%q, %q), want (محجوز, low)", reservedName, reservedHealth)
	}
	if returnedName != "مرتجع" || returnedHealth != "low" {
		t.Fatalf("returned presentation = (%q, %q), want (مرتجع, low)", returnedName, returnedHealth)
	}
}

func TestInventoryStatusPresentationShowsReversedSeparately(t *testing.T) {
	name, _, health := inventoryStatusPresentation("REVERSED")
	if name != "شراء ملغى" || health != "low" {
		t.Fatalf("reversed presentation = (%q, %q), want (شراء ملغى, low)", name, health)
	}
}
