package localdb

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestSalesCostAmountIsNullableForLegacyFallback(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "cost-schema.db"))
	database, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	var defaultValue sql.NullString
	var found bool
	rows, err := database.DB.Query(`PRAGMA table_info(sales)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultColumn sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultColumn, &primaryKey); err != nil {
			t.Fatal(err)
		}
		if name == "cost_amount" {
			found = true
			defaultValue = defaultColumn
			if notNull != 0 {
				t.Fatal("sales.cost_amount must remain nullable for legacy sales")
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("sales.cost_amount column is missing from local schema")
	}
	if defaultValue.Valid {
		t.Fatalf("sales.cost_amount default = %q, want NULL", defaultValue.String)
	}
}
