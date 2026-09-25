package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCloudValidationPreservesRejectionCodes(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   string
	}{
		{http.StatusForbidden, `{"code":"SUBSCRIPTION_SUSPENDED"}`, "SUBSCRIPTION_SUSPENDED"},
		{http.StatusForbidden, `{}`, "PERMISSION_DENIED"},
		{http.StatusUnauthorized, `{"error":{"code":"INVALID_TOKEN"}}`, "INVALID_TOKEN"},
		{http.StatusServiceUnavailable, `{}`, "CLOUD_SERVICE_UNAVAILABLE"},
	}
	for _, tc := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		}))
		_, err := NewCloudAuthService(server.URL).ValidateCloudToken(context.Background(), "token")
		server.Close()
		var cloudErr *CloudValidationError
		if !errors.As(err, &cloudErr) || cloudErr.Status != tc.status || cloudErr.Code != tc.want {
			t.Fatalf("status=%d error=%v, want %s", tc.status, err, tc.want)
		}
	}
}
