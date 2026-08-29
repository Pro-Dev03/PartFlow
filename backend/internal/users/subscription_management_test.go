package users

import (
	"testing"
	"time"
)

func TestUserResponseIncludesSubscriptionDetails(t *testing.T) {
	now := time.Now()
	resp := UserResponse{
		Email:                 "owner@partflow.com",
		FirstName:             "Owner",
		LastName:              "User",
		SubscriptionStatus:    "active",
		SubscriptionExpiresAt: &now,
	}

	if resp.SubscriptionStatus != "active" {
		t.Fatal("subscription status should be present in user response")
	}
	if resp.SubscriptionExpiresAt == nil || resp.SubscriptionExpiresAt.UTC().Unix() != now.UTC().Unix() {
		t.Fatal("subscription expiry should be preserved in user response")
	}
}
