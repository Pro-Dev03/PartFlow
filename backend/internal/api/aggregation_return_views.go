package api

import (
	"database/sql"
)

// Current local stores install accounting return union views in their schema
// migration. This compatibility alias is only used by legacy SQLite files and
// small report fixtures that still have the original return tables.
func ensureSQLiteAggregationReturnViews(db *sql.DB) {
	if db == nil {
		return
	}
	ensureSQLiteAggregationReturnView(db, "accounting_returns", "returns", "")
	var itemColumns map[string]bool
	itemColumns = make(map[string]bool)
	rows, err := db.Query(`PRAGMA table_info(return_items)`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cid, notNull, primaryKey int
			var name, dataType string
			var defaultValue sql.NullString
			if rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey) == nil {
				itemColumns[name] = true
			}
		}
	}
	extra := ""
	if itemColumns["quantity"] && !itemColumns["quantity_returned"] {
		extra = ", quantity AS quantity_returned"
	}
	ensureSQLiteAggregationReturnView(db, "accounting_return_items", "return_items", extra)
}

func ensureSQLiteAggregationReturnView(db *sql.DB, viewName, tableName, extraColumns string) {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='view' AND name=?`, viewName).Scan(&count); err != nil || count > 0 {
		return
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, tableName).Scan(&count); err != nil || count == 0 {
		return
	}
	_, _ = db.Exec(`CREATE VIEW ` + viewName + ` AS SELECT *` + extraColumns + ` FROM ` + tableName)
}
