package business

import "testing"

func TestIsDebtOverdueUsesAllOpenStatuses(t *testing.T) {
	for _, status := range []string{"pending", "partial", "overdue"} {
		if !IsDebtOverdue(100, true, status) {
			t.Fatalf("expected %q debt to be overdue", status)
		}
	}
	if IsDebtOverdue(100, true, "paid") {
		t.Fatal("paid debt must not be overdue")
	}
	if IsDebtOverdue(0, true, "pending") {
		t.Fatal("zero remaining debt must not be overdue")
	}
}
