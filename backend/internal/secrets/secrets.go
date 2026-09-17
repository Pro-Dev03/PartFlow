package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

const prefix = "enc:v1:"

func key() ([]byte, error) {
	value := strings.TrimSpace(os.Getenv("PARTFLOW_PAYMENT_ENCRYPTION_KEY"))
	if value == "" {
		return nil, fmt.Errorf("PARTFLOW_PAYMENT_ENCRYPTION_KEY is required for payment secrets")
	}
	digest := sha256.Sum256([]byte(value))
	return digest[:], nil
}

func Encrypt(value string) (string, error) {
	if value == "" || strings.HasPrefix(value, prefix) {
		return value, nil
	}
	keyBytes, err := key()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(value), nil)
	return prefix + base64.RawStdEncoding.EncodeToString(sealed), nil
}

func Decrypt(value string) (string, error) {
	if value == "" || !strings.HasPrefix(value, prefix) {
		return value, nil
	}
	keyBytes, err := key()
	if err != nil {
		return "", err
	}
	decoded, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(decoded) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid encrypted secret")
	}
	nonce, ciphertext := decoded[:gcm.NonceSize()], decoded[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
