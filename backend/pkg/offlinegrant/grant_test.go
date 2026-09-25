package offlinegrant

import (
	"testing"
	"time"
)

func TestOfflineGrantIssueAndVerificationAreDisabled(t *testing.T) {
	grant, err := Issue("subscriber-id", nil, time.Now())
	if grant != "" || err != ErrDisabled {
		t.Fatalf("Issue() = (%q, %v), want empty grant and ErrDisabled", grant, err)
	}

	for name, verify := range map[string]func() (Claims, error){
		"normal verification":         func() (Claims, error) { return Verify("legacy-signed-grant", time.Now()) },
		"signature-only verification": func() (Claims, error) { return VerifyClaims("legacy-signed-grant") },
	} {
		t.Run(name, func(t *testing.T) {
			claims, err := verify()
			if claims != (Claims{}) || err != ErrDisabled {
				t.Fatalf("verification = (%+v, %v), want empty claims and ErrDisabled", claims, err)
			}
		})
	}
}
