package sales

import "testing"

func TestCalculateSaleAmountsAppliesDiscountBeforeTax(t *testing.T) {
	discount, tax, total, grossProfit, netProfit := calculateSaleAmounts(1000, 800, 10, 15, "percentage", 10)

	if discount != 100 {
		t.Fatalf("discount = %v, want 100", discount)
	}
	if tax != 90 {
		t.Fatalf("tax = %v, want 90", tax)
	}
	if total != 990 {
		t.Fatalf("total = %v, want 990", total)
	}
	if grossProfit != 100 || netProfit != 100 {
		t.Fatalf("profit = (%v, %v), want (100, 100)", grossProfit, netProfit)
	}
}

func TestCalculateSaleAmountsHonorsZeroDiscountLimit(t *testing.T) {
	discount, tax, total, _, _ := calculateSaleAmounts(1000, 800, 10, 0, "percentage", 10)

	if discount != 0 {
		t.Fatalf("discount = %v, want 0", discount)
	}
	if tax != 100 || total != 1100 {
		t.Fatalf("tax/total = (%v, %v), want (100, 1100)", tax, total)
	}
}
