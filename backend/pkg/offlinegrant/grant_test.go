package offlinegrant

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

func TestIssueAndVerifyGrantAndBoundItBySubscriptionExpiry(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(privateKeyEnv, base64.StdEncoding.EncodeToString(privateKey))
	t.Setenv(publicKeyEnv, base64.StdEncoding.EncodeToString(publicKey))

	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	subscriptionExpiry := now.Add(24 * time.Hour)
	grant, err := Issue("subscriber-id", &subscriptionExpiry, now)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Verify(grant, now.Add(23*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "subscriber-id" || claims.Expires != subscriptionExpiry.Unix() {
		t.Fatalf("claims=%+v, want user and subscription-capped expiry", claims)
	}
	if _, err := Verify(grant, subscriptionExpiry); err == nil {
		t.Fatal("grant should expire with the subscription")
	}
}

func TestVerifyRejectsEditedAndExpiredGrant(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(privateKeyEnv, base64.StdEncoding.EncodeToString(privateKey))
	t.Setenv(publicKeyEnv, base64.StdEncoding.EncodeToString(publicKey))
	now := time.Now().UTC().Truncate(time.Second)
	grant, err := Issue("subscriber-id", nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(grant+"x", now); err == nil {
		t.Fatal("edited grant must not verify")
	}
	if _, err := Verify(grant, now.Add(GracePeriod)); err == nil {
		t.Fatal("expired grant must not verify")
	}
}

func TestIssueRequiresConfiguredPrivateKey(t *testing.T) {
	t.Setenv(privateKeyEnv, "")
	_, err := Issue("subscriber-id", nil, time.Now())
	if err == nil || err != ErrUnavailable {
		t.Fatalf("Issue error=%v, want ErrUnavailable", err)
	}
}

func TestDecodeKeyAcceptsURLSafeEncoding(t *testing.T) {
	key := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(key)
	decoded, err := decodeKey(encoded)
	if err != nil || string(decoded) != string(key) {
		t.Fatalf("decoded key mismatch, err=%v", err)
	}
}
