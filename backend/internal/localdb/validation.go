package localdb

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var sqliteHeader = []byte("SQLite format 3\x00")

// ValidateExistingFile verifies that a restore candidate is an intact PartFlow
// SQLite database without running migrations or modifying the file.
func ValidateExistingFile(path string) error {
	absPath, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return fmt.Errorf("resolve database path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("stat database: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() < int64(len(sqliteHeader)) {
		return fmt.Errorf("database file is empty or not a regular file")
	}
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("open database header: %w", err)
	}
	header := make([]byte, len(sqliteHeader))
	_, readErr := file.Read(header)
	closeErr := file.Close()
	if readErr != nil {
		return fmt.Errorf("read database header: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close database header: %w", closeErr)
	}
	if !bytes.Equal(header, sqliteHeader) {
		return fmt.Errorf("file is not a SQLite database")
	}

	db, err := sql.Open("sqlite", absPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA query_only = ON`); err != nil {
		return fmt.Errorf("make database validation read-only: %w", err)
	}

	var quickCheck string
	if err := db.QueryRow(`PRAGMA quick_check`).Scan(&quickCheck); err != nil {
		return fmt.Errorf("run SQLite integrity check: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(quickCheck), "ok") {
		return fmt.Errorf("SQLite integrity check failed: %s", quickCheck)
	}

	requiredColumns := map[string][]string{
		"users":           {"id"},
		"settings":        {"key", "value"},
		"products":        {"id", "sku", "name"},
		"inventory_items": {"id", "product_id", "item_code", "status"},
		"sales":           {"id", "total_amount"},
		"purchases":       {"id", "supplier_id", "total_amount"},
	}
	for table, required := range requiredColumns {
		var exists bool
		if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)`, table).Scan(&exists); err != nil {
			return fmt.Errorf("inspect PartFlow schema: %w", err)
		}
		if !exists {
			return fmt.Errorf("database is missing required PartFlow table %q", table)
		}
		rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
		if err != nil {
			return fmt.Errorf("inspect PartFlow table %q: %w", table, err)
		}
		columns := make(map[string]bool)
		for rows.Next() {
			var column string
			if err := rows.Scan(&column); err != nil {
				_ = rows.Close()
				return fmt.Errorf("read PartFlow table %q columns: %w", table, err)
			}
			columns[column] = true
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return fmt.Errorf("read PartFlow table %q columns: %w", table, err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close PartFlow table %q columns: %w", table, err)
		}
		for _, column := range required {
			if !columns[column] {
				return fmt.Errorf("PartFlow table %q is missing required column %q", table, column)
			}
		}
	}
	return nil
}

// CreateConsistentBackup creates a standalone SQLite snapshot with VACUUM INTO,
// including committed WAL data, without modifying the source. The destination must not exist.
func CreateConsistentBackup(sourcePath, destinationPath string) error {
	source, err := filepath.Abs(strings.TrimSpace(sourcePath))
	if err != nil {
		return fmt.Errorf("resolve source database path: %w", err)
	}
	destination, err := filepath.Abs(strings.TrimSpace(destinationPath))
	if err != nil {
		return fmt.Errorf("resolve backup path: %w", err)
	}
	if strings.EqualFold(source, destination) {
		return fmt.Errorf("backup destination must differ from the active database")
	}
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("backup destination already exists")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect backup destination: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}
	if err := ValidateExistingFile(source); err != nil {
		return fmt.Errorf("active database is not valid: %w", err)
	}

	db, err := sql.Open("sqlite", source+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return fmt.Errorf("open active database: %w", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("connect to active database: %w", err)
	}
	if _, err := db.Exec(`VACUUM INTO ?`, destination); err != nil {
		return fmt.Errorf("create SQLite snapshot: %w", err)
	}
	if err := ValidateExistingFile(destination); err != nil {
		_ = os.Remove(destination)
		return fmt.Errorf("verify backup snapshot: %w", err)
	}
	return nil
}
