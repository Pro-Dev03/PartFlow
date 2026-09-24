package dashboard

import "database/sql"

// ensureSQLiteAccountingReturnViews keeps legacy SQLite files usable before
// their next schema migration. Current databases have permanent union views
// that include detached accounting effects; this fallback only aliases the
// operational tables when those views are absent.
func ensureSQLiteAccountingReturnViews(db *sql.DB) {
	if db == nil {
		return
	}
	ensureSQLiteAccountingReturnView(db, "accounting_returns", "returns")
	ensureSQLiteAccountingReturnView(db, "accounting_return_items", "return_items")
}

func ensureSQLiteAccountingReturnView(db *sql.DB, viewName, tableName string) {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='view' AND name=?`, viewName).Scan(&count); err != nil || count > 0 {
		return
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, tableName).Scan(&count); err != nil || count == 0 {
		return
	}
	_, _ = db.Exec(`CREATE VIEW ` + viewName + ` AS SELECT * FROM ` + tableName)
}
