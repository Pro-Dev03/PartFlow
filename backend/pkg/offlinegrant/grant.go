package offlinegrant

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	// GracePeriod bounds local operation when the cloud authority is unreachable.
	GracePeriod   = 72 * time.Hour
	privateKeyEnv = "PARTFLOW_OFFLINE_GRANT_PRIVATE_KEY"
	publicKeyEnv  = "PARTFLOW_OFFLINE_GRANT_PUBLIC_KEY"
)

var (
	ErrUnavailable = errors.New("offline grant signing key is not configured")
	ErrInvalid     = errors.New("offline grant is invalid or expired")
)

// Claims is a cloud-signed, time-bounded authorization for the local API.
// The local copy can be edited, but its signature and signed expiration cannot.
type Claims struct {
	Version int    `json:"v"`
	UserID  string `json:"sub"`
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

func Issue(userID string, subscriptionExpiresAt *time.Time, now time.Time) (string, error) {
	key, err := readPrivateKey()
	if err != nil {
		return "", err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fmt.Errorf("offline grant user id is required")
	}
	now = now.UTC().Truncate(time.Second)
	expiresAt := now.Add(GracePeriod)
	if subscriptionExpiresAt != nil && subscriptionExpiresAt.Before(expiresAt) {
		expiresAt = subscriptionExpiresAt.UTC().Truncate(time.Second)
	}
	if !expiresAt.After(now) {
		return "", ErrInvalid
	}
	claims := Claims{Version: 1, UserID: userID, Issued: now.Unix(), Expires: expiresAt.Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	signature := ed25519.Sign(key, []byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func Verify(token string, now time.Time) (Claims, error) {
	var claims Claims
	key, err := readPublicKey()
	if err != nil {
		return claims, err
	}
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return claims, ErrInvalid
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !ed25519.Verify(key, []byte(parts[0]), signature) {
		return claims, ErrInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || json.Unmarshal(payload, &claims) != nil {
		return claims, ErrInvalid
	}
	if claims.Version != 1 || strings.TrimSpace(claims.UserID) == "" || claims.Issued <= 0 || claims.Expires <= claims.Issued {
		return Claims{}, ErrInvalid
	}
	if now.UTC().Unix() >= claims.Expires || now.UTC().Unix() < claims.Issued {
		return Claims{}, ErrInvalid
	}
	return claims, nil
}

func readPrivateKey() (ed25519.PrivateKey, error) {
	encoded := strings.TrimSpace(os.Getenv(privateKeyEnv))
	if encoded == "" {
		return nil, ErrUnavailable
	}
	data, err := decodeKey(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", privateKeyEnv, err)
	}
	switch len(data) {
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(data), nil
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(data), nil
	default:
		return nil, fmt.Errorf("%s must contain a base64 encoded Ed25519 private key", privateKeyEnv)
	}
}

func readPublicKey() (ed25519.PublicKey, error) {
	encoded := strings.TrimSpace(os.Getenv(publicKeyEnv))
	if encoded == "" {
		return nil, ErrUnavailable
	}
	data, err := decodeKey(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", publicKeyEnv, err)
	}
	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("%s must contain a base64 encoded Ed25519 public key", publicKeyEnv)
	}
	return ed25519.PublicKey(data), nil
}

func decodeKey(value string) ([]byte, error) {
	for _, encoding := range []*base64.Encoding{base64.RawStdEncoding, base64.StdEncoding, base64.RawURLEncoding, base64.URLEncoding} {
		if decoded, err := encoding.DecodeString(value); err == nil {
			return decoded, nil
		}
	}
	return nil, errors.New("key is not valid base64")
}

// Fingerprint returns a non-secret identifier suitable for deployment logs.
func Fingerprint(publicKey []byte) string {
	digest := sha256.Sum256(publicKey)
	return base64.RawURLEncoding.EncodeToString(digest[:8])
}
