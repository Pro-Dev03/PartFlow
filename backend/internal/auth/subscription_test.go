package auth

import (
	"testing"
	"time"
)

func TestIsSubscriptionExpired(t *testing.T) {
	now := time.Now()
	service := &Service{}

	for _, status := range []string{"expired", "EXPIRED", " expired ", "cancelled", "CANCELED"} {
		if !service.IsSubscriptionExpired(status, &now) {
			t.Fatalf("status %q should be treated as expired", status)
		}
	}

	expiredAt := now.Add(-time.Hour)
	if !service.IsSubscriptionExpired("active", &expiredAt) {
		t.Fatal("past expiry date should be treated as expired")
	}

	futureAt := now.Add(time.Hour)
	if service.IsSubscriptionExpired("active", &futureAt) {
		t.Fatal("future expiry date should remain valid")
	}

	if service.IsSubscriptionExpired("active", nil) {
		t.Fatal("missing expiry date should not be considered expired for active plan")
	}

	if service.IsSubscriptionExpired("ACTIVE ", nil) {
		t.Fatal("active status with no expiry should remain valid")
	}
}
