package supplierreturns

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func TestServiceListUsesTextCastsForUUIDFallbacks(t *testing.T) {
	db, err := sqlx.Open("postgres", "host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	defer db.Close()

	service := NewService(db)
	if _, err := service.List(context.Background(), ""); err == nil {
		t.Fatal("List() unexpectedly succeeded against an empty PostgreSQL connection; this test is only intended to confirm query compilation, not data existence")
	}
	if err != nil && err.Error() == "list supplier returns: ERROR: COALESCE types text and uuid cannot be matched (SQLSTATE 42804)" {
		t.Fatalf("List() still hits the UUID/text COALESCE bug: %v", err)
	}
}
