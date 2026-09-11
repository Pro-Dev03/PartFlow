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
	if math.Abs(tax-90) > 0.000001 {
		t.Fatalf("tax = %v, want 90", tax)
	}
	if total != 990 {
		t.Fatalf("total = %v, want 990", total)
	}
	if math.Abs(grossProfit-100) > 0.000001 || math.Abs(netProfit-100) > 0.000001 {
		t.Fatalf("profit = (%v, %v), want (100, 100)", grossProfit, netProfit)
	}
}

func TestCalculateSaleAmountsHonorsZeroDiscountLimit(t *testing.T) {
	discount, tax, total, _, _ := calculateSaleAmounts(1000, 800, 10, 0, "percentage", 10)

	if discount != 0 {
		t.Fatalf("discount = %v, want 0", discount)
	}
	if math.Abs(tax-90.90909090909088) > 0.000001 || total != 1000 {
		t.Fatalf("tax/total = (%v, %v), want (%v, 1000)", tax, total, 90.90909090909088)
	}
}
