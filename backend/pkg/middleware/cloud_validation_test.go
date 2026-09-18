package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestValidateWithCloudDoesNotReuseSuccessfulValidation(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	var requests atomic.Int32
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"data":{"valid":true,"user":{"id":"%s","email":"subscriber@example.test"}}}`, userID)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer cloud.Close()

	t.Setenv("PARTFLOW_CLOUD_API_URL", cloud.URL)

	gotID, _, firstErr := validateWithCloud(t.Context(), "cloud-access-token")
	if firstErr != nil {
		t.Fatalf("first validation failed: %v", firstErr.err)
	}
	if gotID.String() != userID {
		t.Fatalf("first validation user id = %s, want %s", gotID, userID)
	}

	_, _, secondErr := validateWithCloud(t.Context(), "cloud-access-token")
	if secondErr == nil || secondErr.status != http.StatusUnauthorized {
		t.Fatalf("second validation error = %#v, want unauthorized", secondErr)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("cloud validation request count = %d, want 2", got)
	}
}
