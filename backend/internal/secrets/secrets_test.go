package secrets

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	t.Setenv("PARTFLOW_PAYMENT_ENCRYPTION_KEY", "test-key")
	ciphertext, err := Encrypt("secret-value")
	if err != nil {
		t.Fatal(err)
	}
	if ciphertext == "secret-value" || ciphertext[:7] != "enc:v1:" {
		t.Fatalf("unexpected ciphertext: %q", ciphertext)
	}
	plaintext, err := Decrypt(ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "secret-value" {
		t.Fatalf("plaintext = %q", plaintext)
	}
}
