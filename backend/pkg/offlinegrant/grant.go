// Package offlinegrant is retained only to reject grants written by legacy
// builds. Local authorization now requires a live decision from the cloud;
// device wall-clock time and mutable local state are not authorization inputs.
package offlinegrant

import (
	"errors"
	"time"
)

var ErrDisabled = errors.New("offline authorization grants are disabled; cloud verification is required")

// Claims preserves the legacy shape for callers that must recognize old data.
// No local grant is accepted as authorization.
type Claims struct {
	Version int    `json:"v"`
	UserID  string `json:"sub"`
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

func Issue(string, *time.Time, time.Time) (string, error) {
	return "", ErrDisabled
}

func Verify(string, time.Time) (Claims, error) {
	return Claims{}, ErrDisabled
}

// VerifyClaims also rejects grants. Signature-only validation must not be
// mistaken for authorization by a future offline fallback.
func VerifyClaims(string) (Claims, error) {
	return Claims{}, ErrDisabled
}
