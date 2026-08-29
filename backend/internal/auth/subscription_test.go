package auth

import (
	"testing"
	"time"
)

func TestIsSubscriptionExpired(t *testing.T) {
	now := time.Now()

	if !IsSubscriptionExpired("expired", &now) {
		t.Fatal("expired status should be treated as expired")
	}

	expiredAt := now.Add(-time.Hour)
	if !IsSubscriptionExpired("active", &expiredAt) {
		t.Fatal("past expiry date should be treated as expired")
	}

	futureAt := now.Add(time.Hour)
	if IsSubscriptionExpired("active", &futureAt) {
		t.Fatal("future expiry date should remain valid")
	}

	if IsSubscriptionExpired("active", nil) {
		t.Fatal("missing expiry date should not be considered expired for active plan")
	}
}
