package sales

import (
	"math"
	"testing"
)

func TestCalculateSaleAmountsAppliesDiscountBeforeTax(t *testing.T) {
	discount, tax, total, grossProfit, netProfit := calculateSaleAmounts(1100, 800, 10, 15, "percentage", 10)

	if discount != 110 {
		t.Fatalf("discount = %v, want 110", discount)
	}
	if math.Abs(tax-99) > 0.000001 {
		t.Fatalf("tax = %v, want 99", tax)
	}
	if total != 1089 {
		t.Fatalf("total = %v, want 1089", total)
	}
	if math.Abs(grossProfit-190) > 0.000001 || math.Abs(netProfit-190) > 0.000001 {
		t.Fatalf("profit = (%v, %v), want (190, 190)", grossProfit, netProfit)
	}
}

func TestCalculateSaleAmountsHonorsZeroDiscountLimit(t *testing.T) {
	discount, tax, total, _, _ := calculateSaleAmounts(1000, 800, 10, 0, "percentage", 10)

	if discount != 0 {
		t.Fatalf("discount = %v, want 0", discount)
	}
	if math.Abs(tax-100) > 0.000001 || total != 1100 {
		t.Fatalf("tax/total = (%v, %v), want (100, 1100)", tax, total)
	}
}

func TestCalculateSaleAmountsExemptKeepsEnteredPrice(t *testing.T) {
	_, tax, total, _, _ := calculateSaleAmounts(150, 100, 0, 15, "", 0)
	if tax != 0 || total != 150 {
		t.Fatalf("tax/total = (%v, %v), want (0, 150)", tax, total)
	}
}
